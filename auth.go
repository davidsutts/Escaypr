package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
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
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// Users mirrors the structure of the table in the database.
type Users struct {
	Id        int
	Uname     string
	PwordHash string
	Email     string
}

// Cookies mirrors the structure of the table in the database.
type Cookies struct {
	ID         int
	User       string
	CookieHash string
	ExpiryTime time.Time
}

// userAuthVals contains all the data which is stored in the encoded value of the
// userAuth cookies.
type userAuthVals struct {
	ID          int
	Username    string
	SessionHash []byte
}

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

var defaultArgonParams = argon2params{Memory: 64 * 1024, Iterations: 3, Parallelism: 4, SaltLength: 16, KeyLength: 32}

// expLength is the cookie expiry time used for auth cookies.
const expLength = 240 * time.Hour // Auth cookies last 10 days.

// validateLogin is used to check whether a username and password pair are a valid pair that
// correspond to a user in the database.
func validateLogin(uname, pword string, ctx context.Context) (uid int, sessionHash string) {
	user := Users{}
	result := db.First(&user, "uname = ?", uname)
	if result.Error != nil {
		log.Println("failed to get user record:", result.Error)
		return -1, ""
	}

	// Check if password is correct.
	match, err := comparePasswordAndHash(pword, user.PwordHash)
	if err != nil {
		log.Println("Failed to compare password: %w", err)
		return -1, ""
	}

	// Return session hash, or fail.
	if match {
		return uid, generateSessionHash(user.PwordHash)
	} else {
		return -1, ""
	}
}

// writeCookie creates an auth cookie and writes it to the http response
// and creates a new session in the database.
func writeAuthCookie(w http.ResponseWriter, uid int, username, userHash string) error {
	// Salt the hash for unique client sessionHash.
	decUserHash, err := hex.DecodeString(userHash)
	if err != nil {
		return fmt.Errorf("unable to decode userHash: %w", err)
	}
	hash, err := addSalt(decUserHash)
	if err != nil {
		return fmt.Errorf("unable to add salt: %w", err)
	}

	// Set maximum expiry date of 400 days
	expTime := time.Now().UTC().Add(400 * 24 * time.Hour)

	// Create cookie.
	cookie := http.Cookie{
		Name:     "userAuth",
		Value:    fmt.Sprintf("%d:%s:%x", uid, username, hash),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  expTime,
	}

	// Write cookie to response.
	http.SetCookie(w, &cookie)

	// Rehash the session hash for the database.
	h := sha256.New()
	h.Write(hash)

	// Encode the user id and hash.
	sessionHash := encodeCookieHash(uid, h.Sum(nil))

	// Write session to database.
	sessionCookie := Cookies{User: username, CookieHash: sessionHash, ExpiryTime: expTime}
	result := db.Create(&sessionCookie)
	if result.Error != nil {
		return fmt.Errorf("unable to write session to db: %w", err)
	}

	return nil
}

// validateCookie should be called whenever a request contains a cookie
// to validate whether it is a valid cookie associated with a login.
func validateCookie(r *http.Request) (valid bool, username string) {
	ck, err := r.Cookie("userAuth")
	if err != nil {
		return false, ""
	}
	uaVals, err := decodeCookie(ck)
	if err != nil {
		return false, ""
	}

	// Encode the hash.
	h := sha256.New()
	h.Write([]byte(uaVals.SessionHash))
	encHash := encodeCookieHash(uaVals.ID, h.Sum(nil))

	// Query the database for the hash.
	cookie := Cookies{}
	result := db.Where("cookie_hash = ?", encHash).First(&cookie)
	if result.Error != nil {
		log.Println(err)
		return false, ""
	}

	// Check uid lines up with cookie.
	if cookie.User != uaVals.Username {
		return false, ""
	}

	return true, uaVals.Username
}

// decodeCookie takes an encoded cookie string and returns the userID,
// username and hash. Returns an error if these values cannot be parsed.
func decodeCookie(ck *http.Cookie) (*userAuthVals, error) {
	// Get the userID, username, sessionhash from the cookie.
	s := strings.Split(ck.Value, ":")
	if len(s) != 3 {
		return nil, errors.New("invalid cookie length")
	}
	uid, err := strconv.Atoi(s[0])
	if err != nil {
		return nil, fmt.Errorf("couldn't parse uid: %w", err)
	}
	uname := s[1]
	if uname == "" {
		return nil, errors.New("no username")
	}
	strHash := s[2]
	hash, err := hex.DecodeString(strHash)
	if err != nil || len(hash) != 32 {
		return nil, fmt.Errorf("unable to decode string to hex: %w", err)
	}

	return &userAuthVals{ID: uid, Username: uname, SessionHash: hash}, nil

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
//
// This hash has a length of 97 characters
func generateArgon2Hash(password string, p argon2params) (encodedHash string, err error) {
	// Generate salt.
	salt, err := generateSalt(p.SaltLength)
	if err != nil {
		return "", fmt.Errorf("could not generate salt: %w", err)
	}

	// Generate hash.
	hash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	// Encode salt and hash.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.Memory, p.Iterations, p.Parallelism, b64Salt, b64Hash)

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
	otherHash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

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
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
