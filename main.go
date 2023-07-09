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
	"context"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/microsoft/go-mssqldb"
)

var tmpl *template.Template

var (
	db         *sql.DB
	server     = "localhost"
	port       = 1433
	user       = "sa"
	sapassword string
	database   = "escaypr"
)

func main() {
	// Parse command-line flags.
	flag.StringVar(&sapassword, "dbpword", "", "DB SA Password")
	flag.Parse()

	// Create context.
	ctx := context.Background()

	// Connect to database.
	db = dbConnect(ctx)

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

// dbConnect initialises a connection with the database and returns a reference to a sql.DB.
func dbConnect(ctx context.Context) *sql.DB {
	// Create connection string.
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;", server, user, sapassword, port, database)

	// Connect to server.
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}

	// Test server connection.
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Connected!\n")

	return db
}

// indexHandler handles requests to the index (home) page.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	valid, username := validateCookie(r)
	if !valid {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
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
