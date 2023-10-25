package main

import (
	"crypto/sha256"
	"html/template"
	"log"
	"net/http"

	"gorm.io/gorm"
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
		Version string
	}{Title: "Login", Signup: false, Version: version}

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
	encHash := encodeCookieHash(uaVals.ID, h.Sum(nil))

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

	// Check if username or email are taken.
	user := Users{Uname: username, PwordHash: encHash, Email: email}
	result := db.First(&user, "uname = ? OR email = ?", username, email)
	if result.Error != gorm.ErrRecordNotFound {
		if (result.Error != nil) {
			log.Println("failed finding existing users:", result.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			log.Println("failed signup attempt: duplicate key")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("duplicate key err"))
			return
		}
	}

	// Create a new user.
	result = db.Create(&user)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not create new user"))
		return
	}
	uid := user.Id

	// Create new cookie and log user in.
	err = writeAuthCookie(w, uid, username, generateSessionHash(encHash))
	if err != nil {
		log.Println("couldn't write cookie:", err)
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Successful signup for %s", username)

}
