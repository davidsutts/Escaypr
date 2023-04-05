package main

import (
	"html/template"
	"log"
	"net/http"
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

	username := r.FormValue("username")
	password := r.FormValue("password")

	log.Println("Username:", username, "Password:", password)
}
