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

	var loginData = struct{ Title string }{Title: "Login"}

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
		w.Write([]byte("Invalid Username or password"))
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
	_, err = db.Exec(
		"DELETE FROM Cookies WHERE sessionHash = @sessionHash",
		sql.Named("sessionHash", encHash),
	)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("logged out")
	w.WriteHeader(200)

}
