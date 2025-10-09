package utils

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func CleanupExpiredFiles(filesDir string) {
	jsonDir := filepath.Join(filesDir, "json")

	entries, err := os.ReadDir(jsonDir)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Error("failed to read metadata directory for cleanup", "path", jsonDir, "error", err)
		}
		return
	}

	slog.Info("running background cleanup for expired files...")
	deletedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		jsonPath := filepath.Join(jsonDir, entry.Name())

		file, err := os.ReadFile(jsonPath)
		if err != nil {
			slog.Error("failed to read metadata file during cleanup", "path", jsonPath, "error", err)
			continue
		}

		var meta FileMetadata
		if err := json.Unmarshal(file, &meta); err != nil {
			slog.Error("Failed to unmarshal metadata during cleanup", "path", jsonPath, "error", err)
			continue
		}

		if meta.ExpiresAt != nil && time.Now().After(*meta.ExpiresAt) {
			originalFilename := entry.Name()[:len(entry.Name())-len(".json")]
			actualFilePath := filepath.Join(filesDir, originalFilename)

			slog.Info("file expired, removing.", "path", actualFilePath)

			if err := os.Remove(actualFilePath); err != nil && !os.IsNotExist(err) {
				slog.Error("failed to remove expired file", "path", actualFilePath, "error", err)
			}
			if err := os.Remove(jsonPath); err != nil && !os.IsNotExist(err) {
				slog.Error("failed to remove expired metadata file", "path", jsonPath, "error", err)
			}
			deletedCount++
		}
	}
	slog.Info("cleanup finished.", "deleted_files", deletedCount)
}
