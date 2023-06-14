package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	hash := s[1]
	if len([]byte(hash)) != 64 {
		return false
	}

	// Encode the hash.
	h := sha256.New()
	h.Write([]byte(hash))
	encHash := encodeCookieHash(uid, h.Sum(nil))

	// Query the database for the hash.
	var dbUid int
	row := db.QueryRow(
		"SELECT userID FROM Cookies WHERE sessionHash = @encHash",
		sql.Named("encHash", encHash),
	)
	err = row.Scan(&dbUid)
	if err != nil {
		return false
	}
	if dbUid != uid {
		return false
	}

	return true

}
