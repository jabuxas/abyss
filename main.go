package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	port     = ":8080"
	filesDir = "./files"
)

var url string = os.Getenv("URL")

func main() {
	http.HandleFunc("/upload", uploadHandler)
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(filesDir))))
	log.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	file, _, err := r.FormFile("file")
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
