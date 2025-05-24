package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/jabuxas/abyss/internal/app"
	"github.com/jabuxas/abyss/internal/config"
	"github.com/jabuxas/abyss/internal/handler"
)

func main() {
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
