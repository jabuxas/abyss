package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
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

	var tmpl *template.Template

	if _, err := os.Stat("./templates/dirlist.html"); err == nil {
		tmpl = template.Must(template.ParseFiles("templates/dirlist.html"))
	} else {
		tmpl = template.Must(template.ParseFS(treeTemplate, "templates/dirlist.html"))
	}
	templateData := TemplateData{
		Files: fileInfos,
		URL:   app.url,
	}
	if err := tmpl.Execute(w, templateData); err != nil {
		slog.Warn(err.Error())
	}
}

func (app *Application) fileHandler(w http.ResponseWriter, r *http.Request) {
	path := fmt.Sprintf("%s", filepath.Base(r.URL.Path))
	realPath := filepath.Join(app.filesDir, path)

	if !filepath.IsLocal(realPath) {
		http.Error(w, "Wrong url", http.StatusBadRequest)
		return
	}

	if fileInfo, err := os.Stat(realPath); err == nil && !fileInfo.IsDir() {
		http.ServeFile(w, r, realPath)
		return
	}
}

func (app *Application) indexHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(app.filesDir); err != nil {
		if err := os.Mkdir(app.filesDir, 0750); err != nil {
			http.Error(w, "Error creating storage directory", http.StatusInternalServerError)
		}
	}

	if r.Method == http.MethodPost {
		app.uploadHandler(w, r)
		return
	}

	name := filepath.Base(r.URL.Path)
	realPath := filepath.Join(app.filesDir, name)

	if !filepath.IsLocal(realPath) || strings.Contains(r.URL.Path, filepath.Clean(app.filesDir)) {
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
			BasicAuth(app.formHandler, app)(w, r)
		} else {
			app.formHandler(w, r)
		}
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

	full := true
	if len(r.Form["secret"]) == 0 {
		full = false
	}
	filename := app.publicURL(file, ".txt", full)

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

	ResponseURLHandler(r, w, app.url, filename)
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

	full := true
	if len(r.Form["secret"]) == 0 {
		full = false
	}
	filename := app.publicURL(file, filepath.Ext(handler.Filename), full)

	// reopen the file for copying, as the hash process consumed the file reader
	file, _, err = r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err = SaveFile(app.lastUploadedFile, file); err != nil {
		fmt.Fprintf(w, "Error parsing file: %s", err.Error())
	}

	ResponseURLHandler(r, w, app.url, filename)
}

func (app *Application) publicURL(file io.Reader, extension string, full bool) string {
	filename, _ := HashFile(file, extension, full)

	filepath := filepath.Join(app.filesDir, filename)

	app.lastUploadedFile = filepath

	return filename
}

func (app *Application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	})

	tokenString, err := token.SignedString([]byte(app.key))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", tokenString)
}
