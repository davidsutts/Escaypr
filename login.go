package main

import (
	"crypto/sha256"
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

// loginHandler handles requests to the login page.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl = template.Must(template.ParseFiles("static/html/login.html"))

	log.Println(r.URL.Path)

	valid, _ := validateCookie(r)
	if valid {
		h := http.RedirectHandler("/", http.StatusSeeOther)
		h.ServeHTTP(w, r)
		return
	}

	var loginData = struct {
		Title  string
		Signup bool
	}{Title: "Login", Signup: false}

	err := tmpl.Execute(w, loginData)
	if err != nil {
		http.Error(w, "failed to write template", 500)
	}
}

// loginFormHandler handles form requests sent to /login/form. This is used
// to validate logins.
func loginFormHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	// Get context.
	ctx := r.Context()

	// Get form inputs.
	username := r.FormValue("username")
	password := r.FormValue("password")

	if uid, hash := validateLogin(username, password, ctx); hash != "" {
		err := writeAuthCookie(w, uid, username, hash)
		if err != nil {
			log.Printf("couldn't write cookie: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		log.Printf("Successful login attempt for %s", username)
		return
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Invalid Username or Password")
		w.Write([]byte("Invalid Username or Password"))
		log.Printf("Failed Login attempt")
		return
	}
}

// logoutFormHandler handles requests to the /logout page.
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	log.Println("MADE IT")

	ck, err := r.Cookie("userAuth")
	if err != nil {
		log.Println("userAuth cookie couldn't be parsed:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get cookie sessionID
	uaVals, err := decodeCookie(ck)
	if err != nil {
		log.Println("couldn't decode cookie:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Hash the hash to get the db hash.
	h := sha256.New()
	h.Write(uaVals.SessionHash)
	encHash := encodeCookieHash(uaVals.UserID, h.Sum(nil))

	// Delete the session from the dB.
	cookie := Cookies{}
	result := db.Delete(&cookie, "cookie_hash = ?", encHash)
	if result.Error != nil {
		log.Println(result.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("logged out")
	w.WriteHeader(200)

}

// signupFormHandler handles form requests to the /signup/form URL.
func signupFormHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get form inputs.
	username := r.FormValue("username")
	email := r.FormValue("email")
	pword := r.FormValue("password")

	// Create new user.
	encHash, err := generateArgon2Hash(pword, defaultArgonParams)
	if err != nil {
		log.Println("couldn't hash password:", err)
	}
	log.Println(len(encHash))
	row := db.QueryRow(
		"IF EXISTS (SELECT * FROM Users WHERE email = @email) BEGIN RAISERROR('Duplicate email', 16, 1) RETURN END "+
			"IF EXISTS (SELECT * FROM Users WHERE uname = @uname) BEGIN RAISERROR('Duplicate uname', 16, 1) RETURN END "+
			"ELSE INSERT INTO Users (email, uname, pword) "+
			"VALUES (@email,@uname,@pword) "+
			"SELECT userID FROM Users WHERE email=@email AND uname=@uname",
		sql.Named("uname", username),
		sql.Named("email", email),
		sql.Named("pword", encHash),
	)

	// Get the userID and error.
	var uid int
	err = row.Scan(&uid)

	if err != nil {
		log.Println("couldn't create user:", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	// Create new cookie and log user in.
	err = writeAuthCookie(w, uid, username, generateSessionHash(encHash))
	if err != nil {
		log.Println("couldn't write cookie:", err)
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Successful signup for %s", username)

}
