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
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var tmpl *template.Template

var (
	db         *gorm.DB
	host       = os.Getenv("DB_HOST")
	port, _    = strconv.Atoi(os.Getenv("DB_PORT"))
	user       = os.Getenv("DB_USER")
	sapassword = os.Getenv("DB_PWORD")
	database   = os.Getenv("DB")
)

func main() {
	// Connect to database.
	var err error
	db, err = dbConnect()
	if err != nil {
		log.Fatalln(err)
	}

	mux := http.NewServeMux()

	// Redirect file requests to static dir.
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))

	// Assign function handlers to each page.
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/login/", loginHandler)
	mux.HandleFunc("/logout/", logoutHandler)
	mux.HandleFunc("/login/form", loginFormHandler)
	mux.HandleFunc("/signup/form", signupFormHandler)
	mux.HandleFunc("/favicon.ico", faviconHandler)

	// Create a HTTP server.
	log.Println("listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// dbConnect initialises a connection with the database and returns a reference to a gorm.DB.
func dbConnect() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Australia/Sydney", host, user, sapassword, database, port)
	for i := 0; i < 5; i++ {
		gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Printf("connected to database: %s", database)
			return gormDB, err
		}
		log.Printf("failed to connect to %s: attempt %d", database, i+1)
	}
	return new(gorm.DB), fmt.Errorf("exceeded max retries and couldn't connect to database")

}

// indexHandler handles requests to the index (home) page.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	valid, username := validateCookie(r)
	if !valid {
    w.Header().Set("Location", "/login")
    w.WriteHeader(http.StatusSeeOther)
    return
  }

	tmpl = template.Must(template.ParseFiles("static/html/index.html"))

	log.Println(r.URL.Path)

	var indexData = struct {
		Title string
		Uname string
	}{Title: "Escapyr", Uname: username}

	err := tmpl.Execute(w, indexData)
	if err != nil {
		http.Error(w, "failed to write template", 500)
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/images/favicon.ico")
}
