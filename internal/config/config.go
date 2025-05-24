package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AuthUsername string
	AuthPassword string
	AbyssURL     string
	UploadKey    string
	FilesDir     string
	Port         string
	ShouldAuth   bool
	Debug        bool
	JWTSecretKey []byte // For JWT, same as UploadKey but for simplicity here

	ServerIdleTimeout  time.Duration
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found or error loading it, relying on environment variables.")
	}

	cfg := &Config{
		FilesDir:           "./files",
		Port:               "3235",
		ShouldAuth:         true,
		Debug:              false,
		ServerIdleTimeout:  10 * time.Second,
		ServerReadTimeout:  10 * time.Second,
		ServerWriteTimeout: 60 * time.Second,
	}

	if val := os.Getenv("AUTH_USERNAME"); val != "" {
		cfg.AuthUsername = val
	} else {
		slog.Error("AUTH_USERNAME environment variable is required but not set.")
		return nil, fmt.Errorf("AUTH_USERNAME environment variable is required")
	}

	if val := os.Getenv("AUTH_PASSWORD"); val != "" {
		cfg.AuthPassword = val
	} else {
		slog.Error("AUTH_PASSWORD environment variable is required but not set.")
		return nil, fmt.Errorf("AUTH_PASSWORD environment variable is required")
	}

	if val := os.Getenv("ABYSS_URL"); val != "" {
		cfg.AbyssURL = val
	} else {
		cfg.AbyssURL = "localhost:" + cfg.Port
		slog.Warn("ABYSS_URL not set, defaulting to " + cfg.AbyssURL)
	}

	if val := os.Getenv("UPLOAD_KEY"); val != "" {
		cfg.UploadKey = val
		cfg.JWTSecretKey = []byte(val)
	} else {
		slog.Warn("UPLOAD_KEY not set. File uploads via API key will be insecure.")
	}

	if val := os.Getenv("ABYSS_FILEDIR"); val != "" {
		cfg.FilesDir = val
	}

	if val := os.Getenv("ABYSS_PORT"); val != "" {
		cfg.Port = val
		if cfg.AbyssURL == "localhost:3235" && cfg.Port != "3235" {
			cfg.AbyssURL = "localhost:" + cfg.Port
		}
	}

	if val := os.Getenv("SHOULD_AUTH"); val != "" {
		cfg.ShouldAuth = (val == "yes" || val == "true")
	}

	if val := os.Getenv("DEBUG"); val == "1" || val == "true" {
		cfg.Debug = true
	}

	// could be parsing other types if needed in future
	// if val := os.Getenv("SERVER_IDLE_TIMEOUT_SECONDS"); val != "" {
	//  seconds, err := strconv.Atoi(val)
	//  if err == nil {
	//      cfg.ServerIdleTimeout = time.Duration(seconds) * time.Second
	//  }
	// }

	return cfg, nil
}
