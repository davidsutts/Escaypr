package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"

	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/argon2"
)

type params struct {
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl = template.Must(template.ParseFiles("static/html/login.html"))

	log.Println(r.URL.Path)

	var loginData = struct{ Title string }{Title: "Login"}

	err := tmpl.Execute(w, loginData)
	if err != nil {
		http.Error(w, "failed to write template", 500)
	}

}

func loginFormHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	// Get context.
	ctx := r.Context()

	// Get form inputs.
	username := r.FormValue("username")
	password := r.FormValue("password")

	if hash := validateLogin(username, password, ctx); hash != "" {
		writeAuthCookie(w, hash)
		w.WriteHeader(http.StatusOK)
		log.Printf("Successful login attempt for %s", username)
		return
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid Username or password"))
		log.Printf("Failed Login attempt")
		return
	}

}

func validateLogin(uname, pword string, ctx context.Context) (sessionHash string) {

	// Get password hash from db.
	var dbPwordHash string
	err := db.QueryRowContext(
		ctx,
		"SELECT pword FROM Users WHERE uname = @uname;",
		sql.Named("uname", uname),
	).Scan(&dbPwordHash)
	if err != nil {
		log.Println("failed to get value:", err)
		return ""
	}

	// Check if password is correct.
	match, err := comparePasswordAndHash(pword, dbPwordHash)
	if err != nil {
		log.Println("Failed to compare password: %w", err)
		return ""
	}

	// Return session hash, or fail.
	if match {
		return generateSessionHash(dbPwordHash)
	} else {
		return ""
	}

}

func generateHash(password string, p params) (encodedHash string, err error) {
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

func comparePasswordAndHash(password, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password hash.
	p, salt, hash, err := decodeHash(encodedHash)
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

func generateSalt(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil

}

func decodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
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

	p = &params{}
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
