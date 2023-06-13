package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

// generateSessionHash generates the session hash from an encoded string
// containing the argon2 hashed password.
func generateSessionHash(encPwordHash string) string {
	h := sha256.New()
	h.Write([]byte(encPwordHash))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// writeCookie creates an auth cookie and writes it to the http response.
func writeAuthCookie(w http.ResponseWriter, hash string) {
	// Create cookie.
	cookie := http.Cookie{
		Name:     "userAuth",
		Value:    hash,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	// Write cookie to response.
	http.SetCookie(w, &cookie)

}
