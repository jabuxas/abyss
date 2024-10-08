package main

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	lastUploadedFile string
}

type FileInfo struct {
	Name          string
	Path          string
	Size          int64
	FormattedSize string
}

type TemplateData struct {
	Files []FileInfo
	URL   string
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
			FormattedSize: formatFileSize(info.Size()),
		})
	}

	tmpl := template.Must(template.ParseFiles("templates/dirlist.html"))
	templateData := TemplateData{
		Files: fileInfos,
		URL:   app.url,
	}
	if err := tmpl.Execute(w, templateData); err != nil {
		slog.Warn(error.Error(err))
	}
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
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
		ext := filepath.Ext(path)

		textExtensions := map[string]bool{
			".sh": true, ".bash": true, ".zsh": true,
			".bat": true, ".cmd": true, ".ps1": true,
			".ini": true, ".cfg": true, ".conf": true,
			".toml": true, ".yml": true, ".yaml": true,
			".c": true, ".cpp": true, ".h": true,
			".go": true, ".py": true, ".js": true,
			".ts": true, ".html": true, ".htm": true,
			".xml": true, ".css": true, ".java": true,
			".rs": true, ".rb": true, ".php": true,
			".pl": true, ".sql": true, ".md": true,
			".log": true, ".txt": true, ".csv": true,
			".json": true, ".env": true, ".sum": true,
			".gitignore": true, ".dockerfile": true, ".Makefile": true,
			".rst": true,
		}

		videoExtensions := map[string]bool{
			".mkv": true,
		}

		if textExtensions[ext] {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}

		if videoExtensions[ext] {
			w.Header().Set("Content-Type", "video/mp4")
		}

		http.ServeFile(w, r, path)
		return
	}

	http.StripPrefix("/", http.FileServer(http.Dir("./static"))).ServeHTTP(w, r)
}

func (app *Application) lastHandler(w http.ResponseWriter, r *http.Request) {
	if app.lastUploadedFile == "" {
		http.Error(w, "No new files uploaded yet", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, app.lastUploadedFile)
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

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		http.Error(w, "Error hashing file content", http.StatusInternalServerError)
		return
	}

	sha1Hash := hex.EncodeToString(hasher.Sum(nil))[:8]

	filename := fmt.Sprintf("%s%s", sha1Hash, filepath.Ext(handler.Filename))

	filepath := filepath.Join(app.filesDir, filename)

	// reopen the file for copying, as the hash process consumed the file reader
	file, _, err = r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error creating file\n", http.StatusInternalServerError)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error copying the file", http.StatusInternalServerError)
	}

	app.lastUploadedFile = filepath

	fmt.Fprintf(w, "http://%s/%s\n", app.url, filename)
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
