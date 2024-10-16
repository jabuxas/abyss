package main

import (
	"html/template"
	"log/slog"
	"net/http"
)

func (app *Application) displayFile(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/files.html"))

	fileName := "19ad500a.pdf"

	file := FileInfo{
		Name: fileName,
		Path: app.url,
		Type: "pdf",
	}

	if err := tmpl.Execute(w, file); err != nil {
		slog.Warn(err.Error())
	}
}
