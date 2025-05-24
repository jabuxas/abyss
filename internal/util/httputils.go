package util

import (
	"fmt"
	"net/http"
)

// ResponseURLHandler sends a redirect response with the paste URL.
func ResponseURLHandler(r *http.Request, w http.ResponseWriter, baseURL, filename string) {
	protocol := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		protocol = "https"
	}
	pasteURL := fmt.Sprintf("%s://%s/%s", protocol, baseURL, filename)

	w.Header().Set("Location", pasteURL)
	w.WriteHeader(http.StatusSeeOther)
	fmt.Fprint(w, pasteURL)
}
