package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	filesDir = "./files"
	port     = ":8080"
)

type application struct {
	auth struct {
		username string
		password string
	}
	url string
}

func main() {
	app := new(application)

	app.auth.username = os.Getenv("AUTH_USERNAME")
	app.auth.password = os.Getenv("AUTH_PASSWORD")
	app.url = os.Getenv("URL")

	if app.auth.username == "" {
		log.Fatal("basic auth username must be provided")
	}

	if app.auth.password == "" {
		log.Fatal("basic auth password must be provided")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.fileHandler)
	mux.HandleFunc(
		"/tree/",
		app.basicAuth(app.treeHandler),
	)
	mux.HandleFunc("/upload", app.uploadHandler)

	srv := &http.Server{
		Addr:         port,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	log.Printf("starting server on %s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (app *application) treeHandler(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/tree/", http.FileServer(http.Dir(filesDir))).ServeHTTP(w, r)
}

func (app *application) fileHandler(w http.ResponseWriter, r *http.Request) {
	name := filepath.Clean(r.URL.Path)
	path := filepath.Join(filesDir, name)

	if !filepath.IsLocal(path) {
		http.Error(w, "Wrong url", http.StatusBadRequest)
		return
	}

	if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
		http.ServeFile(w, r, path)
	} else {
		http.NotFound(w, r)
	}
}

func (app *application) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	if !checkAuth(w, r) {
		http.Error(w, "You're not authorized.", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if _, err := os.Stat(filesDir); err != nil {
		if err := os.Mkdir(filesDir, 0750); err != nil {
			http.Error(w, "Error creating storage directory", http.StatusInternalServerError)
		}
	}

	time := int64(float64(time.Now().Unix()) * 2.71828) // euler :)

	filename := fmt.Sprintf("%d%s", time, filepath.Ext(handler.Filename))

	filepath := fmt.Sprintf("%s/%s", filesDir, filename)

	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error creating file\n", http.StatusInternalServerError)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error copying the file", http.StatusInternalServerError)
	}

	if app.url == "" {
		fmt.Fprintf(w, "http://localhost%s/%s\n", port, filename)
	} else {
		fmt.Fprintf(w, "http://%s/%s\n", app.url, filename)
	}
}

func checkAuth(w http.ResponseWriter, r *http.Request) bool {
	authKey, err := os.ReadFile(".key")
	if err != nil {
		http.Error(w, "Couldn't find your .key", http.StatusNotFound)
	}
	return r.Header.Get("X-Auth")+"\n" == string(authKey)
}

func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
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
