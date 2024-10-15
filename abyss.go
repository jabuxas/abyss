package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	app := new(Application)

	err := godotenv.Load()
	if err != nil {
		slog.Warn("no .env file detected, getting env from running process")
	}

	app.auth.username = os.Getenv("AUTH_USERNAME")
	app.auth.password = os.Getenv("AUTH_PASSWORD")
	app.url = os.Getenv("ABYSS_URL")
	app.key = os.Getenv("UPLOAD_KEY")
	app.filesDir = os.Getenv("ABYSS_FILEDIR")
	app.port = os.Getenv("ABYSS_PORT")

	auth := os.Getenv("SHOULD_AUTH")

	if app.auth.username == "" {
		log.Fatal("basic auth username must be provided")
	}

	if app.auth.password == "" {
		log.Fatal("basic auth password must be provided")
	}

	if app.key == "" {
		slog.Warn("no upload key detected")
	}

	if app.filesDir == "" {
		slog.Warn("file dir is not set, running on default ./files")
		app.filesDir = "./files"
	}

	if app.port == "" {
		slog.Info("running on default port")
		app.port = ":3235"
	} else {
		slog.Info("running on modified port")
		app.port = ":" + app.port
	}

	if app.url == "" {
		slog.Warn("no root url detected, defaulting to localhost.")
		app.url = "localhost" + app.port
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.indexHandler)
	mux.Handle(
		"/tree/",
		http.StripPrefix(
			"/tree",
			app.basicAuth(app.fileListingHandler),
		),
	)
	mux.HandleFunc("/last", app.lastUploadedHandler)
	if auth == "yes" {
		mux.HandleFunc("/upload", app.basicAuth(app.uploadHandler))
		slog.Warn("text uploading through the browser will be restricted")
	} else {
		mux.HandleFunc("/upload", app.uploadHandler)
		slog.Warn("text uploading through the browser will NOT be restricted")
	}

	srv := &http.Server{
		Addr:         app.port,
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
