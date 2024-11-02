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
	err := godotenv.Load()
	if err != nil {
		slog.Warn("no .env file detected, getting env from running process")
	}

	app := &Application{
		auth: struct {
			username string
			password string
		}{
			username: os.Getenv("AUTH_USERNAME"),
			password: os.Getenv("AUTH_PASSWORD"),
		},
		url:        os.Getenv("ABYSS_URL"),
		key:        os.Getenv("UPLOAD_KEY"),
		filesDir:   os.Getenv("ABYSS_FILEDIR"),
		port:       os.Getenv("ABYSS_PORT"),
		authUpload: os.Getenv("SHOULD_AUTH"),
	}

	parseEnv(app)

	mux := http.NewServeMux()

	setupHandlers(mux, app)

	srv := &http.Server{
		Addr:         app.port,
		Handler:      mux,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	log.Printf("starting server on %s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func parseEnv(app *Application) {
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
}

func setupHandlers(mux *http.ServeMux, app *Application) {
	mux.HandleFunc("/", app.indexHandler)

	mux.Handle(
		"/tree/",
		http.StripPrefix(
			"/tree",
			BasicAuth(app.fileListingHandler, app),
		),
	)

	mux.HandleFunc("/last", app.lastUploadedHandler)

	mux.HandleFunc("/token", BasicAuth(app.createTokenHandler, app))

	mux.HandleFunc("/files/", app.fileHandler)
}
