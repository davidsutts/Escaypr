package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// argon2params is used to pass the argon 2 arguments easily between functions.
type argon2params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

// expLength is the cookie expiry time used for auth cookies.
const expLength = 240 * time.Hour // Auth cookies last 10 days.

// validateLogin is used to check whether a username and password pair are a valid pair that
// correspond to a user in the database.
func validateLogin(uname, pword string, ctx context.Context) (uid int, sessionHash string) {

	// Get password hash from db.
	var (
		dbPwordHash string
	)

	err := db.QueryRowContext(
		ctx,
		"SELECT userID, pword FROM Users WHERE uname = @uname;",
		sql.Named("uname", uname),
	).Scan(&uid, &dbPwordHash)
	if err != nil {
		log.Println("failed to get value:", err)
		return -1, ""
	}

	// Check if password is correct.
	match, err := comparePasswordAndHash(pword, dbPwordHash)
	if err != nil {
		log.Println("Failed to compare password: %w", err)
		return -1, ""
	}

	// Return session hash, or fail.
	if match {
		return uid, generateSessionHash(dbPwordHash)
	} else {
		return -1, ""
	}
}

// writeCookie creates an auth cookie and writes it to the http response
// and creates a new session in the database.
func writeAuthCookie(w http.ResponseWriter, uid int, userHash string) error {
	// Salt the hash for unique client sessionHash.
	decUserHash, err := hex.DecodeString(userHash)
	if err != nil {
		return fmt.Errorf("unable to decode userHash: %w", err)
	}
	hash, err := addSalt(decUserHash)
	if err != nil {
		return fmt.Errorf("unable to add salt: %w", err)
	}

	// Create cookie.
	cookie := http.Cookie{
		Name:     "userAuth",
		Value:    fmt.Sprintf("%d:%x", uid, hash),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	// Write cookie to response.
	http.SetCookie(w, &cookie)

	// Rehash the session hash for the database.
	h := sha256.New()
	h.Write(hash)

	// Encode the user id and hash.
	sessionHash := encodeCookieHash(uid, h.Sum(nil))
	expTime := time.Now().UTC().Add(expLength)

	// Write session to database.
	_, err = db.Exec(
		"INSERT INTO Cookies (userID, sessionHash, expiryTime) VALUES (@uid, @sessionHash, @expTime)",
		sql.Named("uid", uid),
		sql.Named("sessionHash", sessionHash),
		sql.Named("expTime", expTime),
	)
	if err != nil {
		return fmt.Errorf("unable to write session to db: %w", err)
	}

	return nil
}

// validateCookie should be called whenever a request contains a cookie
// to validate whether it is a valid cookie associated with a login.
func validateCookie(r *http.Request) (valid bool) {
	ck, err := r.Cookie("userAuth")
	if err != nil {
		return false
	}
	// Get the user id and sessionhash from the cookie.
	s := strings.Split(ck.Value, ":")
	if len(s) != 2 {
		return false
	}
	uid, err := strconv.Atoi(s[0])
	if err != nil {
		return false
	}
	strHash := s[1]
	hash, err := hex.DecodeString(strHash)
	if err != nil || len(hash) != 32 {
		return false
	}

	// Encode the hash.
	h := sha256.New()
	h.Write([]byte(hash))
	encHash := encodeCookieHash(uid, h.Sum(nil))
	// Query the database for the hash.
	var (
		dbUid   int
		expTime time.Time
	)
	row := db.QueryRow(
		"SELECT userID, expiryTime FROM Cookies WHERE sessionHash = @encHash",
		sql.Named("encHash", encHash),
	)
	err = row.Scan(&dbUid, &expTime)
	if err != nil {
		log.Println(err)
		return false
	}

	// Check uid lines up with cookie.
	if dbUid != uid {
		return false
	}

	// Check if expiry time was in the past.
	if time.Since(expTime).Seconds() > 0 {
		_, err := db.Exec(
			"DELETE FROM Cookies WHERE sessionHash = @encHash",
			sql.Named("encHash", encHash),
		)
		if err != nil {
			log.Println(err)
		}
		log.Printf("userID %d logged out: expired cookie", uid)
		return false
	}

	// Extend expiry time.
	formExpTime := time.Now().UTC().Add(expLength).Format("2006-01-02 15:04:05")
	_, err = db.Exec(
		"UPDATE Cookies SET expiryTime = @expTime WHERE sessionHash = @encHash",
		sql.Named("expTime", formExpTime),
		sql.Named("encHash", encHash),
	)
	if err != nil {
		log.Println("couldn't update expiryTime: %w", err)
	}

	return true
}

// generateSessionHash generates the session hash from an encoded string
// containing the argon2 hashed password.
func generateSessionHash(encPwordHash string) string {
	h := sha256.New()
	h.Write([]byte(encPwordHash))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// encodeCookieHash encodes the user id and the hash to create a string
// that looks like "$uid=<uid>$<hash>".
func encodeCookieHash(uid int, hash []byte) (encCookieHash string) {
	return fmt.Sprintf("$uid=%d$%x", uid, hash)
}

// addSalt generates a random salt, and adds it to the referenced
// hash, and returns the salt.
func addSalt(hash []byte) (hashOut []byte, err error) {
	const saltLength = 8
	salt, err := generateSalt(saltLength)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate salt: %w", err)
	}
	for i := range salt {
		hash[i] += salt[i]
	}
	return hash, nil
}

// generateArgon2Hash takes a user password, and some defined parameters for hashing
// and returns an argon2 algorithm hash of the password with an additional salt.
//
// learn more: https://en.wikipedia.org/wiki/Argon2
// Using the argon2 go library.
func generateArgon2Hash(password string, p argon2params) (encodedHash string, err error) {
	// Generate salt.
	salt, err := generateSalt(p.saltLength)
	if err != nil {
		return "", fmt.Errorf("could not generate salt: %w", err)
	}

	// Generate hash.
	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Encode salt and hash.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// comparePasswordAndHash takes in an encoded hash string containing an argon 2 hash, and a
// salt, and runs a constant time comparison to determine whether the password is the correct
// input.
func comparePasswordAndHash(password, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password hash.
	p, salt, hash, err := decodeArgon2Hash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// generateSalt takes a salt length and generates a cryptographically secure byte slice
// to be used as a salt when hashing.
func generateSalt(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// decodeArgon2Hash takes an encoded string with $ seperators and returns the parameters
// used to achieve the hash.
func decodeArgon2Hash(encodedHash string) (p *argon2params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &argon2params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}
