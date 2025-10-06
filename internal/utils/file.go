package utils

import (
	"path/filepath"
	"strings"
)

func DetectFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return "image"
		}
	}

	videoExts := []string{".mp4", ".webm", ".ogg", ".mov", ".avi", ".mkv"}
	for _, vidExt := range videoExts {
		if ext == vidExt {
			return "video"
		}
	}

	audioExts := []string{".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a"}
	for _, audExt := range audioExts {
		if ext == audExt {
			return "audio"
		}
	}

	if ext == ".pdf" {
		return "pdf"
	}

	textExts := []string{".txt", ".md", ".log", ".json", ".xml", ".html", ".css", ".js", ".go", ".py", ".java", ".c", ".cpp", ".h"}
	for _, txtExt := range textExts {
		if ext == txtExt {
			return "text"
		}
	}

	return "unknown"
}
