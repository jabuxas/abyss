package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
)

var extensions = map[string]string{
	".sh":   "text",
	".mp4":  "video",
	".pdf":  "pdf",
	".txt":  "text",
	".png":  "image",
	".jpg":  "image",
	".json": "text",
}

func displayFile(app *Application, file string, w http.ResponseWriter) {
	tmpl := template.Must(template.ParseFiles("templates/files.html"))

	fileInfo := FileInfo{
		Name: file,
		Path: app.url,
		Type: getType(file),
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
