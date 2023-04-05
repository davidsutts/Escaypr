package main

import (
	"html/template"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl = template.Must(template.ParseFiles("static/html/login.html"))

	var loginData = struct{ Title string }{Title: "Login"}

	err := tmpl.Execute(w, loginData)
	if err != nil {
		http.Error(w, "failed to write template", 500)
	}
}