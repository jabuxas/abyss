package main

import (
	"embed"
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

	".mp3": "audio", ".aac": "audio", ".wav": "audio", ".flac": "audio", ".ogg": "audio",

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
	".rst": "text", ".el": "text", ".fish": "text",
}

//go:embed templates/files.html
var filesTemplate embed.FS

func DisplayFile(app *Application, file string, w http.ResponseWriter) {
	var tmpl *template.Template

	if _, err := os.Stat("./templates/dirlist.html"); err == nil {
		tmpl = template.Must(template.ParseFiles("templates/files.html"))
	} else {
		tmpl = template.Must(template.ParseFS(filesTemplate, "templates/files.html"))
	}

	realPath := filepath.Join(app.filesDir, filepath.Base(file))

	fileStat, _ := os.Stat("./" + realPath)
	fileContent, _ := os.ReadFile("./" + realPath)

	fileInfo := FileInfo{
		Name:    file,
		Path:    filepath.Join(app.url, file),
		Type:    getType(file),
		Content: string(fileContent),
		TimeUploaded: fileStat.ModTime().
			UTC().
			Format("2006-01-02 15:04:05 UTC"),
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
	return "unknown"
}
