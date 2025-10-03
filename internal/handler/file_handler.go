package handler

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jabuxas/abyss/internal/middleware"
	"github.com/jabuxas/abyss/internal/util"
)

// uploadDelegator decides which upload handler to call based on Content-Type.
func (h *Handler) uploadDelegator(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		middleware.BasicAuth(h.App, h.formUploadHandler)(w, r)
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		h.curlUploadHandler(w, r)
	} else {
		h.App.Logger.Warn("Upload attempt with unsupported content type", "contentType", contentType)
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
	}
}

// formUploadHandler handles uploads from the HTML form.
func (h *Handler) formUploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.App.Logger.Error("Failed to parse form data", "error", err)
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "No content provided", http.StatusBadRequest)
		return
	}

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	contentBytes := []byte(normalized)

	useFullHash := len(r.Form["secret"]) > 0

	filename, err := util.HashFile(bytes.NewReader(contentBytes), ".txt", useFullHash)
	if err != nil {
		h.App.Logger.Error("Error hashing form content", "error", err)
		http.Error(w, "Error processing file", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(h.App.Config.FilesDir, filename)
	if err := util.SaveFile(filePath, bytes.NewReader(contentBytes)); err != nil {
		h.App.Logger.Error("Error saving form uploaded file", "file", filePath, "error", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	h.App.SetLastUploadedFile(filePath)

	redirectURL := "/" + filename
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// curlUploadHandler handles uploads via curl (multipart/form-data).
func (h *Handler) curlUploadHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckUploadAuth(r, h.App) {
		h.App.Logger.Warn("Unauthorized curl upload attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		h.App.Logger.Warn("Failed to retrieve form file (curl)", "error", err)
		http.Error(w, "Error retrieving the file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		h.App.Logger.Error("Failed to buffer uploaded file (curl)", "error", err)
		http.Error(w, "Error processing file", http.StatusInternalServerError)
		return
	}
	fileBytes := buf.Bytes()

	useFullHash := len(r.Form["secret"]) > 0
	filename, err := util.HashFile(bytes.NewReader(fileBytes), filepath.Ext(handler.Filename), useFullHash)
	if err != nil {
		h.App.Logger.Error("Failed to hash buffered file (curl)", "error", err)
		http.Error(w, "Error processing file", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(h.App.Config.FilesDir, filename)
	if err := util.SaveFile(filePath, bytes.NewReader(fileBytes)); err != nil {
		h.App.Logger.Error("Failed to save curl uploaded file", "file", filePath, "error", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	h.App.SetLastUploadedFile(filePath)
	util.ResponseURLHandler(r, w, h.App.Config.AbyssURL, filename)
}

// ServeRawFileHandler serves the raw content of a file.
func (h *Handler) ServeRawFileHandler(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	if filename == "" || filename == "." || filename == ".." {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(h.App.Config.FilesDir, filename)

	cleanBasePath := filepath.Clean(h.App.Config.FilesDir)
	cleanFullPath := filepath.Clean(fullPath)

	if !strings.HasPrefix(cleanFullPath, cleanBasePath+string(os.PathSeparator)) && cleanFullPath != cleanBasePath {
		h.App.Logger.Warn("Attempt to access raw file outside filesDir", "requested", r.URL.Path, "resolved", fullPath)
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
		} else {
			h.App.Logger.Error("Error accessing raw file", "path", fullPath, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if fileInfo.IsDir() {
		http.Error(w, "Cannot serve directories", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, fullPath)
}

// ListAllFilesHandler lists all files in the specified directory (relative to filesDir).
func (h *Handler) ListAllFilesHandler(w http.ResponseWriter, r *http.Request) {
	requestedDir := filepath.Join(h.App.Config.FilesDir, r.URL.Path)

	cleanBasePath := filepath.Clean(h.App.Config.FilesDir)
	cleanRequestedDir := filepath.Clean(requestedDir)
	if !strings.HasPrefix(cleanRequestedDir, cleanBasePath) {
		h.App.Logger.Warn("Attempt to list directory outside filesDir", "requested", r.URL.Path)
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	dirEntries, err := os.ReadDir(cleanRequestedDir)
	if err != nil {
		h.App.Logger.Error("Failed to read directory for listing", "dir", cleanRequestedDir, "err", err)
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var fileInfos []FileInfo
	for _, entry := range dirEntries {
		info, err := entry.Info()
		if err != nil {
			h.App.Logger.Error("Failed to get info for directory entry", "entry", entry.Name(), "err", err)
			continue
		}

		linkPath := filepath.Join(r.URL.Path, entry.Name())

		fileInfos = append(fileInfos, FileInfo{
			Name:          entry.Name(),
			Path:          linkPath,
			Size:          info.Size(),
			FormattedSize: util.FormatFileSize(info.Size()),
			TimeUploaded:  info.ModTime().UTC().Format("2006-01-02 15:04:05 UTC"),
			Type:          h.getFileType(entry.Name()),
		})
	}

	tmplName := "dirlist.html"

	// cache
	tmpl, ok := h.App.Templates[tmplName]
	if !ok {
		h.App.Logger.Error("Template not found in cache", "name", tmplName)
		http.Error(w, "Internal server error (template not found)", http.StatusInternalServerError)
		return
	}

	templateData := TemplateData{
		Files:      fileInfos,
		URL:        h.App.Config.AbyssURL,
		PageTitle:  "Directory Listing: " + r.URL.Path,
		ActivePath: r.URL.Path,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, templateData); err != nil {
		h.App.Logger.Error("Failed to execute dirlist template", "err", err)
	}
}
