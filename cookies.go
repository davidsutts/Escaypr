package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const expLength = 240 * time.Hour // Auth cookies last 10 days.

// generateSessionHash generates the session hash from an encoded string
// containing the argon2 hashed password.
func generateSessionHash(encPwordHash string) string {
	h := sha256.New()
	h.Write([]byte(encPwordHash))
	return fmt.Sprintf("%x", h.Sum(nil))
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

// encodeCookieHash encodes the user id and the hash to create a string
// that looks like "$uid=<uid>$<hash>".
func encodeCookieHash(uid int, hash []byte) (encCookieHash string) {
	return fmt.Sprintf("$uid=%d$%x", uid, hash)
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
