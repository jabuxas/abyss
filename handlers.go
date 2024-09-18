package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Application struct {
	auth struct {
		username string
		password string
	}
	url      string
	key      string
	filesDir string
	port     string
}

func (app *Application) treeHandler(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/tree/", http.FileServer(http.Dir(app.filesDir))).ServeHTTP(w, r)
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
		http.ServeFile(w, r, path)
	} else {
		http.StripPrefix("/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)
	}
}

func (app *Application) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Method not allowed", http.StatusUnauthorized)
		return
	}

	if !app.checkAuth(r) {
		http.Error(w, "You're not authorized.", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if _, err := os.Stat(app.filesDir); err != nil {
		if err := os.Mkdir(app.filesDir, 0750); err != nil {
			http.Error(w, "Error creating storage directory", http.StatusInternalServerError)
		}
	}

	time := int64(float64(time.Now().Unix()) * 2.71828) // euler :)

	filename := fmt.Sprintf("%d%s", time, filepath.Ext(handler.Filename))

	filepath := fmt.Sprintf("%s/%s", app.filesDir, filename)

	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error creating file\n", http.StatusInternalServerError)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error copying the file", http.StatusInternalServerError)
	}

	if app.url == "" {
		fmt.Fprintf(w, "http://localhost%s/%s\n", app.port, filename)
	} else {
		fmt.Fprintf(w, "http://%s/%s\n", app.url, filename)
	}
}

func (app *Application) checkAuth(r *http.Request) bool {
	return r.Header.Get("X-Auth") == string(app.key)
}

func (app *Application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			// hash password received
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))

			// hash our password
			expectedUsernameHash := sha256.Sum256([]byte(app.auth.username))
			expectedPasswordHash := sha256.Sum256([]byte(app.auth.password))

			// compare hashes
			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
