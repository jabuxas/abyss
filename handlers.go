package main

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Application struct {
	auth struct {
		username string
		password string
	}
	url              string
	key              string
	filesDir         string
	port             string
	authText         string
	lastUploadedFile string
}

func (app *Application) fileListingHandler(w http.ResponseWriter, r *http.Request) {
	dir := app.filesDir + r.URL.Path

	files, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var fileInfos []FileInfo
	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fileInfos = append(fileInfos, FileInfo{
			Name:          file.Name(),
			Path:          filepath.Join(r.URL.Path, file.Name()),
			Size:          info.Size(),
			FormattedSize: FormatFileSize(info.Size()),
			TimeUploaded: info.ModTime().
				UTC().
				Format("2006-01-02 15:04:05 UTC"),
		})
	}

	tmpl := template.Must(template.ParseFiles("templates/dirlist.html"))
	templateData := TemplateData{
		Files: fileInfos,
		URL:   app.url,
	}
	if err := tmpl.Execute(w, templateData); err != nil {
		slog.Warn(err.Error())
	}
}

func (app *Application) fileHandler(w http.ResponseWriter, r *http.Request) {
	path := fmt.Sprintf(".%s", filepath.Clean(r.URL.Path))

	if !filepath.IsLocal(path) {
		http.Error(w, "Wrong url", http.StatusBadRequest)
		return
	}

	if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
		http.ServeFile(w, r, path)
		return
	}
}

func (app *Application) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		app.uploadHandler(w, r)
		return
	}

	name := filepath.Clean(r.URL.Path)
	path := filepath.Join(app.filesDir, name)

	if !filepath.IsLocal(path) {
		http.Error(w, "Wrong url", http.StatusBadRequest)
		return
	}

	if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
		DisplayFile(app, "/"+path, w)
		return
	}

	http.StripPrefix("/", http.FileServer(http.Dir("./static"))).ServeHTTP(w, r)
}

func (app *Application) lastUploadedHandler(w http.ResponseWriter, r *http.Request) {
	if app.lastUploadedFile == "" {
		http.Error(w, "No new files uploaded yet", http.StatusNotFound)
		return
	}
	DisplayFile(app, "/"+app.lastUploadedFile, w)
}

func (app *Application) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(app.filesDir); err != nil {
		if err := os.Mkdir(app.filesDir, 0750); err != nil {
			http.Error(w, "Error creating storage directory", http.StatusInternalServerError)
		}
	}
	if contentType := r.Header.Get("Content-Type"); contentType == "application/x-www-form-urlencoded" {
		app.formHandler(w, r)
	} else if strings.Split(contentType, ";")[0] == "multipart/form-data" {
		app.curlHandler(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusUnauthorized)
	}
}

func (app *Application) formHandler(w http.ResponseWriter, r *http.Request) {
	content := r.FormValue("content")

	if err := os.WriteFile("/tmp/file.txt", []byte(content), 0666); err != nil {
		http.Error(w, "Couldn't parse content body", http.StatusNoContent)
	}

	file, err := os.Open("/tmp/file.txt")
	if err != nil {
		http.Error(w, "Couldn't find file", http.StatusNotFound)
	}
	defer file.Close()

	filename := app.publicURL(file, ".txt")

	// reopening file because hash consumes it
	file, err = os.Open("/tmp/file.txt")
	if err != nil {
		http.Error(w, "Couldn't find file", http.StatusNotFound)
	}
	defer file.Close()

	err = SaveFile(app.lastUploadedFile, file)
	if err != nil {
		fmt.Fprintf(w, "Error parsing file: %s", err.Error())
	}

	fmt.Fprintf(w, "%s", fmt.Sprintf("http://%s/%s\n", app.url, filename))
}

func (app *Application) curlHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Method not allowed", http.StatusUnauthorized)
		return
	}

	if !CheckAuth(r, app.key) {
		http.Error(w, "You're not authorized.", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		slog.Warn(err.Error())
		return
	}
	defer file.Close()

	filename := app.publicURL(file, filepath.Ext(handler.Filename))

	// reopen the file for copying, as the hash process consumed the file reader
	file, _, err = r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = SaveFile(app.lastUploadedFile, file)
	if err != nil {
		fmt.Fprintf(w, "Error parsing file: %s", err.Error())
	}

	fmt.Fprintf(w, "%s", fmt.Sprintf("http://%s/%s\n", app.url, filename))
}

func (app *Application) publicURL(file io.Reader, extension string) string {
	filename, _ := HashFile(file, extension)

	filepath := filepath.Join(app.filesDir, filename)

	app.lastUploadedFile = filepath

	return filename
}
