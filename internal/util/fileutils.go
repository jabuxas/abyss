package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// FormatFileSize converts bytes to a human-readable string.
func FormatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
}

// HashFile generates an MD5 hash for the given file content.
// fullHash determines if the full hash or a shortened version is used.
func HashFile(file io.Reader, extension string, fullHash bool) (string, error) {
	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to copy file content to hasher: %w", err)
	}

	hashed := strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
	filename := hashed + extension
	if fullHash {
		return filename, nil
	}
	return hashed[:5] + extension, nil
}

// SaveFile saves the content from the reader to the specified path.
func SaveFile(path string, file io.Reader) error {
	dst, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", path, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return fmt.Errorf("failed to copy content to destination file %s: %w", path, err)
	}
	return nil
}
