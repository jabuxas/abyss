package utils

import (
	"fmt"
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

	textExts := []string{".txt", ".md", ".log", ".json", ".xml", ".html", ".css", ".js", ".go", ".py", ".java", ".c", ".cpp", ".h", ".rb", ".rs", ".sh", ".yml", ".yaml", ".ini", ".cfg", ".toml", ".csv", ".tsv", ".tex", ".el", ".php", ".rtf", ".srt", ".sub", ".vtt", ".sql", ".conf", ".bat", ".ps1", ".jsx", ".tsx", ".vue", ".scss", ".sass", ".less", ".pl", ".swift", ".kt", ".kts", ".groovy", ".r", ".lua", ".dockerfile", ".tf", ".diff", ".patch", ".asciidoc", ".rst", ".m", ".mm", ".f", ".f90", ".asm", ".vb", ".org"}
	for _, txtExt := range textExts {
		if ext == txtExt {
			return "text"
		}
	}

	return "unknown"
}

func FormatFileSize(size int64) string {
	const (
		KB = 1 << (10 * 1)
		MB = 1 << (10 * 2)
		GB = 1 << (10 * 3)
		TB = 1 << (10 * 4)
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}
