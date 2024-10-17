package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var extensions = map[string]string{
	".mp4": "video", ".mkv": "video", ".webm": "video",

	".pdf": "pdf",

	".png": "image", ".jpg": "image", ".jpeg": "image", ".webp": "image",

	".sh": "text", ".bash": "text", ".zsh": "text",
	".bat": "text", ".cmd": "text", ".ps1": "text",
	".ini": "text", ".cfg": "text", ".conf": "text",
	".toml": "text", ".yml": "text", ".yaml": "text",
	".c": "text", ".cpp": "text", ".h": "text",
	".go": "text", ".py": "text", ".js": "text",
	".ts": "text", ".html": "text", ".htm": "text",
	".xml": "text", ".css": "text", ".java": "text",
	".rs": "text", ".rb": "text", ".php": "text",
	".pl": "text", ".sql": "text", ".md": "text",
	".log": "text", ".txt": "text", ".csv": "text",
	".json": "text", ".env": "text", ".sum": "text",
	".gitignore": "text", ".dockerfile": "text", ".Makefile": "text",
	".rst": "text",
}

func DisplayFile(app *Application, file string, w http.ResponseWriter) {
	tmpl := template.Must(template.ParseFiles("templates/files.html"))

	fileContent, _ := os.ReadFile("." + file)

	fileInfo := FileInfo{
		Name:    file,
		Path:    filepath.Join(app.url, file),
		Type:    getType(file),
		Content: string(fileContent),
	}

	if err := tmpl.Execute(w, fileInfo); err != nil {
		slog.Warn(err.Error())
	}
}

func getType(file string) string {
	extension := strings.ToLower(filepath.Ext(file))

	if fileType, exists := extensions[extension]; exists {
		return fileType
	}
	return "text"
}