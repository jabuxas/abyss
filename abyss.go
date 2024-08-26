package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

const (
	filesDir = "./files"
	port     = ":8080"
)

func main() {
	app := new(Application)

	app.auth.username = os.Getenv("AUTH_USERNAME")
	app.auth.password = os.Getenv("AUTH_PASSWORD")
	app.url = os.Getenv("URL")

	if app.auth.username == "" {
		log.Fatal("basic auth username must be provided")
	}

	if app.auth.password == "" {
		log.Fatal("basic auth password must be provided")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.fileHandler)
	mux.HandleFunc(
		"/tree/",
		app.basicAuth(app.treeHandler),
	)

	srv := &http.Server{
		Addr:         port,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	log.Printf("starting server on %s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
