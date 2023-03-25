/*
DESCRIPTION:
main.go handles the setup and initialisation of the website,
as welll as handles the home/index page requests.

AUTHORS:
David Sutton <dsutton1202@gmail.com>

LICENSE:
The following code is licensed under the MIT license.
See LICENSE for more.

COPYRIGHT 2023 - DAVID SUTTON
*/

package main

import (
	"html/template"
	"log"
	"net/http"
)

var tmpl *template.Template

func main() {

	mux := http.NewServeMux()

	// Redirect file requests to static dir.
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/static"))))

	// Assign function handlers to each page.
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/favicon.ico", faviconHandler)

	// Create a HTTP server.
	log.Println("listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl = template.Must(template.ParseFiles("static/html/index.html"))

	var indexData = struct{ Title string }{Title: "Excapyr"}

	err := tmpl.Execute(w, indexData)
	if err != nil {
		http.Error(w, "failed to write template", 500)
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/images/favicon.ico")
}
