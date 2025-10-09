package utils

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type FileMetadata struct {
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	PasswordHash []byte     `json:"password_hash,omitempty"`
}

func ParseExpiration(d string) (*time.Time, error) {
	if d == "" {
		return nil, nil
	}
	duration, err := time.ParseDuration(d)
	if err != nil {
		return nil, err
	}
	t := time.Now().Add(duration)
	return &t, nil
}

func ParsePassword(d string) ([]byte, error) {
	if d == "" {
		return nil, nil
	}
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(d), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return bcryptHash, nil
}

func JsonPathFromFilePath(filePath string) string {
	dir, file := filepath.Split(filePath)
	jsonDir := filepath.Join(dir, "json")
	jsonFile := file + ".json"
	return filepath.Join(jsonDir, jsonFile)
}

func SaveMetadata(path string, expiry *time.Time, passwordHash []byte) error {
	metadata := FileMetadata{
		ExpiresAt:    expiry,
		PasswordHash: passwordHash,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		panic(err)
	}

	jsonPath := JsonPathFromFilePath(path)

	err = os.MkdirAll(filepath.Dir(jsonPath), 0755)
	if err != nil {
		slog.Error("failed to create metadata directory", "error", err, "dir", filepath.Dir(jsonPath))
		return err
	}

	err = os.WriteFile(jsonPath, data, 0644)
	if err != nil {
		slog.Error("failed to write metadata file", "error", err, "path", jsonPath)
		return err
	}

	return nil
}

func ReadMetadata(filePath string) (*FileMetadata, error) {
	jsonPath := JsonPathFromFilePath(filePath)

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return &FileMetadata{}, nil
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		slog.Error("failed to read metadata file", "path", jsonPath, "error", err)
		return nil, err
	}

	var meta FileMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		slog.Error("failed to unmarshal metadata", "path", jsonPath, "error", err)
		return nil, err
	}

	return &meta, nil
}
