package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

// generateSessionHash generates the session hash from an encoded string
// containing the argon2 hashed password.
func generateSessionHash(encPwordHash string) string {
	h := sha256.New()
	h.Write([]byte(encPwordHash))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// writeCookie creates an auth cookie and writes it to the http response
// and creates a new session in the database.
func writeAuthCookie(w http.ResponseWriter, uid int, hash string) error {
	// Create cookie.
	cookie := http.Cookie{
		Name:     "userAuth",
		Value:    fmt.Sprintf("%d:%s", uid, hash),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	// Write cookie to response.
	http.SetCookie(w, &cookie)

	// Rehash the session hash for the database.
	h := sha256.New()
	h.Write([]byte(hash))

	// Encode the user id and hash.
	sessionHash := encodeCookieHash(uid, h.Sum(nil))
	log.Println(sessionHash)
	log.Println(len(sessionHash))

	// Write session to database.
	_, err := db.Exec(
		"INSERT INTO Cookies (userID, sessionHash) VALUES (@uid, @sessionHash)",
		sql.Named("uid", uid),
		sql.Named("sessionHash", sessionHash),
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
