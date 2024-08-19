package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	port        = ":8080"
	filesDir    = "./files"
	maxFileSize = 20 * 1024 * 1024 // 20MiB
)

var url string = os.Getenv("URL")

func main() {
	http.HandleFunc("/upload", uploadHandler)
	http.Handle("/tree/", http.StripPrefix("/tree", http.FileServer(http.Dir(filesDir))))
	http.HandleFunc("/", fileHandler)
	log.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
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

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	if !checkAuth(w, r) {
		http.Error(w, "You're not authorized.", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file or file size exceeds limit", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if _, err := os.Stat(filesDir); err != nil {
		if err := os.Mkdir(filesDir, 0750); err != nil {
			http.Error(w, "Error creating storage directory", http.StatusInternalServerError)
		}
	}

	time := int64(float64(time.Now().Unix()) * 2.71828) // euler :)

	filename := fmt.Sprintf("%s/%d", filesDir, time)

	dst, err := os.Create(filename)
	if err != nil {
		http.Error(w, "Error creating file\n", http.StatusInternalServerError)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error copying the file", http.StatusInternalServerError)
	}

	if url == "" {
		fmt.Fprintf(w, "http://localhost%s/%d\n", port, time)
	} else {
		fmt.Fprintf(w, "http://%s/%d\n", url, time)
	}
}

func checkAuth(w http.ResponseWriter, r *http.Request) bool {
	authKey, _ := os.ReadFile(".key")
	return r.Header.Get("X-Auth")+"\n" == string(authKey)
}
