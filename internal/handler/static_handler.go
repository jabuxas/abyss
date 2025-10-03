package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/jabuxas/abyss/assets"
	"github.com/jabuxas/abyss/internal/app"
	"github.com/jabuxas/abyss/internal/util"
)

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(h.App.Config.FilesDir); errors.Is(err, os.ErrNotExist) {
		if mkdirErr := os.MkdirAll(h.App.Config.FilesDir, 0750); mkdirErr != nil {
			h.App.Logger.Error("Failed to create storage directory", "dir", h.App.Config.FilesDir, "err", mkdirErr)
			http.Error(w, "Error creating storage directory", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		h.App.Logger.Error("Failed to stat storage directory", "dir", h.App.Config.FilesDir, "err", err)
		http.Error(w, "Error accessing storage directory", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		h.uploadDelegator(w, r)
		return
	}

	requestedFile := filepath.Base(r.URL.Path)
	isStaticAssetRequest := requestedFile == "index.html" || requestedFile == "style.css" || requestedFile == "" || requestedFile == "."

	if !isStaticAssetRequest {
		fullPath := filepath.Join(h.App.Config.FilesDir, requestedFile)
		cleanPath := filepath.Clean(fullPath)
		if !strings.HasPrefix(cleanPath, filepath.Clean(h.App.Config.FilesDir)+string(os.PathSeparator)) && cleanPath != filepath.Clean(h.App.Config.FilesDir) {
			h.App.Logger.Warn("Attempt to access path outside filesDir", "requested", r.URL.Path, "resolved", fullPath)
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		if fileInfo, err := os.Stat(fullPath); err == nil && !fileInfo.IsDir() {
			h.displayFile(w, r, requestedFile)
			return
		} else if err != nil && !os.IsNotExist(err) {
			h.App.Logger.Error("Error checking for uploaded file", "path", fullPath, "err", err)
		}
	}

	// Serve static files (index.html, style.css)
	staticDirOnDisk := "./assets/static"
	var fileSystem http.FileSystem

	if _, err := os.Stat(staticDirOnDisk); err == nil {
		h.App.Logger.Debug("Serving static files from disk", "directory", staticDirOnDisk)
		fileSystem = http.Dir(staticDirOnDisk)
	} else {
		h.App.Logger.Debug("Serving static files from embedded FS")
		subFS, err := fs.Sub(assets.StaticFS, "static")
		if err != nil {
			h.App.Logger.Error("Failed to create sub FS for embedded static files", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fileSystem = http.FS(subFS)
	}

	serveFile := r.URL.Path
	if serveFile == "/" || serveFile == "" {
		serveFile = "index.html"
	} else {
		serveFile = strings.TrimPrefix(serveFile, "/")
	}

	f, err := fileSystem.Open(serveFile)
	if err != nil {
		if os.IsNotExist(err) && !isStaticAssetRequest {
			http.NotFound(w, r)
			return
		} else if os.IsNotExist(err) && isStaticAssetRequest {
			http.NotFound(w, r)
			return
		}
		// Other error
		h.App.Logger.Error("Error opening static file", "path", serveFile, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	f.Close()

	http.FileServer(fileSystem).ServeHTTP(w, r)
}

func (h *Handler) LastUploadedHandler(w http.ResponseWriter, r *http.Request) {
	lastFile := h.App.GetLastUploadedFile()
	if lastFile == "" {
		http.Error(w, "No files uploaded yet in this session", http.StatusNotFound)
		return
	}
	h.displayFile(w, r, filepath.Base(lastFile))
}

var extensionsMap = map[string]string{
	".mp4": "video", ".mkv": "video", ".webm": "video",

	".pdf": "pdf",

	".png": "image", ".jpg": "image", ".jpeg": "image", ".webp": "image",

	".mp3": "audio", ".aac": "audio", ".wav": "audio", ".flac": "audio", ".ogg": "audio",

	".sh": "text", ".bash": "text", ".zsh": "text", ".bat": "text", ".cmd": "text", ".ps1": "text",
	".ini": "text", ".cfg": "text", ".conf": "text", ".toml": "text", ".yml": "text", ".yaml": "text",
	".c": "text", ".cpp": "text", ".h": "text", ".go": "text", ".py": "text", ".js": "text",
	".ts": "text", ".html": "text", ".htm": "text", ".xml": "text", ".css": "text", ".java": "text",
	".rs": "text", ".rb": "text", ".php": "text", ".pl": "text", ".sql": "text", ".md": "text",
	".log": "text", ".txt": "text", ".csv": "text", ".json": "text", ".env": "text", ".sum": "text",
	".gitignore": "text", ".dockerfile": "text", ".Makefile": "text", ".rst": "text", ".el": "text", ".fish": "text",
}

func (h *Handler) getFileType(filename string) string {
	extension := strings.ToLower(filepath.Ext(filename))
	if fileType, exists := extensionsMap[extension]; exists {
		return fileType
	}
	return "unknown"
}

func (h *Handler) getLexer(filename string, fileContent []byte) chroma.Lexer {
	ext := strings.ToLower(filepath.Ext(filename))
	var lexer chroma.Lexer
	if ext == ".txt" {
		lexer = lexers.Analyse(string(fileContent))
	} else {
		lexer = lexers.Match(filename)
		if lexer == nil {
			lexer = lexers.Analyse(string(fileContent))
		}
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	return lexer
}

func (h *Handler) displayFile(w http.ResponseWriter, r *http.Request, filename string) {
	tmplName := "fileDisplay.html"
	tmpl, ok := h.App.Templates[tmplName]
	if !ok {
		h.App.Logger.Error("Template not found in cache", "name", tmplName)
		http.Error(w, "Internal server error (template not found)", http.StatusInternalServerError)
		return
	}

	actualFilePath := filepath.Join(h.App.Config.FilesDir, filename)
	fileStat, err := os.Stat(actualFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			h.App.Logger.Warn("File not found for display", "path", actualFilePath)
			http.NotFound(w, r)
		} else {
			h.App.Logger.Error("Error stating file for display", "path", actualFilePath, "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	jsonPath := util.JsonPathFromFilePath(actualFilePath)
	_, err = os.Stat(jsonPath)
	if err != nil {
		h.App.Logger.Info("File does not have metadata", "path", actualFilePath)
	} else {
		file, err := os.ReadFile(jsonPath)
		if err != nil {
			h.App.Logger.Error("Failed to open metadata file", "path", jsonPath, "error", err)
		}

		var meta app.PasteMetadata
		if err := json.Unmarshal(file, &meta); err != nil {
			h.App.Logger.Error("Failed to unmarshal metadata", "error", err)
		}

		if meta.ExpiresAt != nil && time.Now().After(*meta.ExpiresAt) {
			os.Remove(actualFilePath)
			os.Remove(jsonPath)
		}
	}

	fileType := h.getFileType(filename)
	var highlightedContent template.HTML

	if fileType == "text" {
		fileContent, readErr := os.ReadFile(actualFilePath)
		if readErr != nil {
			h.App.Logger.Error("Failed to read text file for highlighting", "path", actualFilePath, "error", readErr)
			http.Error(w, "Could not read file content", http.StatusInternalServerError)
			return
		}

		lexer := h.getLexer(filename, fileContent)
		iterator, tokeniseErr := lexer.Tokenise(nil, string(fileContent))
		if tokeniseErr != nil {
			h.App.Logger.Warn("Chroma tokenizing error", "file", filename, "error", tokeniseErr, "lexer", lexer.Config().Name)
			// Corrected: Use template.HTMLEscapeString
			escapedString := template.HTMLEscapeString(string(fileContent))
			highlightedContent = template.HTML("<pre>" + escapedString + "</pre>")
		} else {
			style := h.App.ChromaStyle
			if style == nil {
				h.App.Logger.Warn("Chroma style not loaded from app, using fallback.")
				style = styles.Fallback
			}
			formatter := chromahtml.New(chromahtml.WithLineNumbers(true), chromahtml.TabWidth(4), chromahtml.WithClasses(false), chromahtml.WithLinkableLineNumbers(true, ""))
			var buf bytes.Buffer
			if formatErr := formatter.Format(&buf, style, iterator); formatErr != nil {
				h.App.Logger.Warn("Chroma formatting error", "file", filename, "error", formatErr)
				// Corrected: Use template.HTMLEscapeString
				escapedString := template.HTMLEscapeString(string(fileContent))
				highlightedContent = template.HTML("<pre>" + escapedString + "</pre>")
			} else {
				highlightedContent = template.HTML(buf.String())
			}
		}
	}

	rawAccessPath := ""
	if h.App.Config.AbyssURL != "" {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		} else if strings.HasPrefix(r.Host, "localhost") && r.Header.Get("X-Forwarded-Proto") == "" {
			scheme = "http"
		}

		hostAndPort := h.App.Config.AbyssURL
		if !strings.HasPrefix(hostAndPort, "http://") && !strings.HasPrefix(hostAndPort, "https://") {
			hostAndPort = scheme + "://" + hostAndPort
		}
		baseRawURL := strings.TrimSuffix(hostAndPort, "/")
		rawAccessPath = baseRawURL + "/raw/" + filename
	} else {
		rawAccessPath = "/raw/" + filename
	}

	data := TemplateData{
		PageTitle: "View File: " + filename,
		URL:       h.App.Config.AbyssURL,
		SingleFile: &FileInfo{
			Name:         filename,
			Path:         rawAccessPath,
			Type:         fileType,
			Content:      highlightedContent,
			TimeUploaded: fileStat.ModTime().UTC().Format("2006-01-02 15:04:05 UTC"),
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		h.App.Logger.Error("Failed to execute fileDisplay template", "error", err)
	}
}
