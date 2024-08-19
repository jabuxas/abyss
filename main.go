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
	url      = "localhost"
	port     = ":8080"
	imageDir = "./images"
)

func main() {
	http.HandleFunc("/upload", uploadHandler)
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(imageDir))))
	log.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if _, err := os.Stat(imageDir); err != nil {
		if err := os.Mkdir(imageDir, 0750); err != nil {
			log.Fatalf("Couldn't create dir %s\n", imageDir)
		}
	}

	time := time.Now().Unix() * 3
	filename := fmt.Sprintf("%s/%d", imageDir, time)

	dst, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Couldn't create file %s\n", filename)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error copying the file", http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "http://%s%s/%d\n", url, port, time)
}
