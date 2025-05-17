package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	authUpload       string
	lastUploadedFile string
}

//go:embed static/**
var static embed.FS

//go:embed templates/dirlist.html
var treeTemplate embed.FS

func (app *Application) listAllFilesHandler(w http.ResponseWriter, r *http.Request) {
	dir := app.filesDir + r.URL.Path

	files, err := os.ReadDir(dir)
	if err != nil {
		slog.Error("failed to read directory", "dir", dir, "err", err)
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var fileInfos []FileInfo
	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			slog.Error("failed to stat file", "path", filePath, "err", err)
			http.Error(w, "Failed to access file", http.StatusInternalServerError)
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

	// use embedded case error
	tmpl, err := template.ParseFiles("templates/dirlist.html")
	if err != nil {
		tmpl = template.Must(template.ParseFS(treeTemplate, "templates/dirlist.html"))
	}

	templateData := TemplateData{
		Files: fileInfos,
		URL:   app.url,
	}

	if err := tmpl.Execute(w, templateData); err != nil {
		slog.Error("failed to execute template", "err", err, "data", templateData)
		http.Error(w, "Template rendering failed", http.StatusInternalServerError)
	}
}

func (app *Application) serveRawFileHandler(w http.ResponseWriter, r *http.Request) {
	path := fmt.Sprintf("%s", filepath.Base(r.URL.Path))
	realPath := filepath.Join(app.filesDir, path)

	if !filepath.IsLocal(realPath) {
		slog.Warn("non-local path detected", "path", realPath)
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if fileInfo, err := os.Stat(realPath); err == nil && !fileInfo.IsDir() {
		http.ServeFile(w, r, realPath)
		return
	}
}

func (app *Application) indexHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(app.filesDir); errors.Is(err, os.ErrNotExist) {
		if mkdirErr := os.Mkdir(app.filesDir, 0750); mkdirErr != nil {
			slog.Error("failed to create storage directory", "dir", app.filesDir, "err", mkdirErr)
			http.Error(w, "Error creating storage directory", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		slog.Error("failed to stat storage directory", "dir", app.filesDir, "err", err)
		http.Error(w, "Error accessing storage directory", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		app.uploadHandler(w, r)
		return
	}

	name := filepath.Base(r.URL.Path)
	realPath := filepath.Join(app.filesDir, name)

	if !filepath.IsLocal(realPath) || strings.Contains(r.URL.Path, filepath.Clean(app.filesDir)) {
		slog.Error("invalid file path detected", "url_path", r.URL.Path, "clean_name", name)
		http.Error(w, "Wrong url", http.StatusBadRequest)
		return
	}

	if fileInfo, err := os.Stat(realPath); err == nil && !fileInfo.IsDir() {
		DisplayFile(app, filepath.Join("/raw", name), w)
		return
	}

	if _, err := os.Stat("./static"); err == nil {
		http.StripPrefix("/", http.FileServer(http.Dir("./static"))).ServeHTTP(w, r)
	} else {
		fs, _ := fs.Sub(static, "static")
		http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP(w, r)
	}
}

func (app *Application) lastUploadedHandler(w http.ResponseWriter, r *http.Request) {
	if app.lastUploadedFile == "" {
		http.Error(w, "No new files uploaded yet", http.StatusNotFound)
		return
	}
	DisplayFile(app, filepath.Join("/raw", filepath.Base(app.lastUploadedFile)), w)
}

func (app *Application) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType == "application/x-www-form-urlencoded" {
		if app.authUpload == "yes" {
			BasicAuth(app.formUploadHandler, app)(w, r)
		} else {
			app.formUploadHandler(w, r)
		}
	} else if strings.Split(contentType, ";")[0] == "multipart/form-data" {
		app.curlHandler(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusUnauthorized)
	}
}

func (app *Application) formUploadHandler(w http.ResponseWriter, r *http.Request) {
	content := r.FormValue("content")

	normalized := strings.ReplaceAll(content, "\r\n", "\n")

	if err := os.WriteFile("/tmp/file.txt", []byte(normalized), 0666); err != nil {
		slog.Error("failed to write file", "error", err)
		http.Error(w, "Couldn't parse content body", http.StatusNoContent)
	}

	file, err := os.Open("/tmp/file.txt")
	if err != nil {
		slog.Warn("file not found", "path", "/tmp/file.txt", "error", err)
		http.Error(w, "Couldn't find file", http.StatusNotFound)
	}
	defer file.Close()

	full := true
	if len(r.Form["secret"]) == 0 {
		full = false
	}

	filename, _ := HashFile(file, ".txt", full)

	// set as lastUploadedFile
	filepath := filepath.Join(app.filesDir, filename)
	app.lastUploadedFile = filepath

	// reopening file because hash consumes it
	file, err = os.Open("/tmp/file.txt")
	if err != nil {
		slog.Warn("file not found", "path", "/tmp/file.txt", "error", err)
		http.Error(w, "Couldn't find file", http.StatusNotFound)
	}
	defer file.Close()

	err = SaveFile(app.lastUploadedFile, file)
	if err != nil {
		slog.Error("error saving file", "file", app.lastUploadedFile, "error", err)
		fmt.Fprintf(w, "Error parsing file: %s", err.Error())
		return
	}

	http.Redirect(w, r, fmt.Sprintf("http://%s/%s", app.url, filename), http.StatusSeeOther)
}

func (app *Application) curlHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		slog.Warn("invalid path accessed", "path", r.URL.Path)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !CheckAuth(r, app.key) {
		slog.Warn("unauthorized access attempt")
		http.Error(w, "You're not authorized.", http.StatusUnauthorized)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		slog.Warn("failed to retrieve form file", "error", err)
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	full := true
	if len(r.Form["secret"]) == 0 {
		full = false
	}

	filename, _ := HashFile(file, filepath.Ext(handler.Filename), full)

	// set as lastUploadedFile
	filepath := filepath.Join(app.filesDir, filename)
	app.lastUploadedFile = filepath

	// reopen the file for copying, as the hash process consumed the file reader
	file, _, err = r.FormFile("file")
	if err != nil {
		slog.Warn("failed to retrieve form file", "error", err)
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err = SaveFile(app.lastUploadedFile, file); err != nil {
		slog.Error("failed to save file", "file", app.lastUploadedFile, "error", err)
		fmt.Fprintf(w, "Error parsing file: %s", err.Error())
		return
	}

	ResponseURLHandler(r, w, app.url, filename)
}

func (app *Application) createJWTHandler(w http.ResponseWriter, r *http.Request) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(2 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(app.key))
	if err != nil {
		slog.Error("failed to sign JWT token", "error", err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", tokenString)
}
