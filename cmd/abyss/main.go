package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/jabuxas/abyss/internal/app"
	"github.com/jabuxas/abyss/internal/config"
	"github.com/jabuxas/abyss/internal/handler"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	versionFlag := flag.Bool("version", false, "print version information and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Printf("abyss\n")
		fmt.Printf(" version:   %s\n", version)
		fmt.Printf(" git commit: %s\n", commit)
		fmt.Printf(" built on:   %s\n", date)
		fmt.Printf(" built by:   %s\n", builtBy)
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration during startup", "error", err)
		os.Exit(1)
	}

	var logLevel slog.Level
	if cfg.Debug {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	application, err := app.NewApplication(cfg)
	if err != nil {
		slog.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}
	application.Logger = logger

	httpHandler := handler.NewHandler(application)

	mux := http.NewServeMux()
	httpHandler.SetupRoutes(mux)

	listenAddr := cfg.Port
	if len(listenAddr) > 0 && listenAddr[0] != ':' {
		listenAddr = ":" + listenAddr
	} else if len(listenAddr) == 0 {
		listenAddr = ":3235"
	}

	srv := &http.Server{
		Addr:         listenAddr,
		Handler:      mux,
		IdleTimeout:  cfg.ServerIdleTimeout,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
	}

	slog.Info("Starting server", "address", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
