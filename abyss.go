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
	// load env from .env
	err := godotenv.Load()
	if err != nil {
		slog.Warn("no .env file detected, getting env from running process")
	}

	if os.Getenv("DEBUG") == "1" {
		// start logging
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))

		slog.SetDefault(logger)
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

	app.initApplication()

	mux := http.NewServeMux()

	app.setupHandlersOnMux(mux)

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
