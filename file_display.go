package main

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var extensions = map[string]string{
	".mp4": "video", ".mkv": "video", ".webm": "video",

	".pdf": "pdf",

	".png": "image", ".jpg": "image", ".jpeg": "image", ".webp": "image",

	".mp3": "audio", ".aac": "audio", ".wav": "audio", ".flac": "audio", ".ogg": "audio",

	".sh": "text", ".bash": "text", ".zsh": "text",
	".bat": "text", ".cmd": "text", ".ps1": "text",
	".ini": "text", ".cfg": "text", ".conf": "text",
	".toml": "text", ".yml": "text", ".yaml": "text",
	".c": "text", ".cpp": "text", ".h": "text",
	".go": "text", ".py": "text", ".js": "text",
	".ts": "text", ".html": "text", ".htm": "text",
	".xml": "text", ".css": "text", ".java": "text",
	".rs": "text", ".rb": "text", ".php": "text",
	".pl": "text", ".sql": "text", ".md": "text",
	".log": "text", ".txt": "text", ".csv": "text",
	".json": "text", ".env": "text", ".sum": "text",
	".gitignore": "text", ".dockerfile": "text", ".Makefile": "text",
	".rst": "text", ".el": "text", ".fish": "text",
}

//go:embed templates/fileDisplay.html
var filesTemplate embed.FS

func DisplayFile(app *Application, file string, w http.ResponseWriter) {
	var tmpl *template.Template

	if _, err := os.Stat("./templates/dirlist.html"); err == nil {
		tmpl = template.Must(template.ParseFiles("templates/fileDisplay.html"))
	} else {
		tmpl = template.Must(template.ParseFS(filesTemplate, "templates/fileDisplay.html"))
	}

	realPath := filepath.Join(app.filesDir, filepath.Base(file))

	fileStat, _ := os.Stat(realPath)
	fileContent, _ := os.ReadFile(realPath)

	var highlighted bytes.Buffer
	if getType(file) == "text" {
		lexer := lexers.Analyse(string(fileContent))
		if lexer == nil {
			lexer = lexers.Fallback
		}

		iterator, err := lexer.Tokenise(nil, string(fileContent))
		if err != nil {
			slog.Warn("Chroma tokenizing error: " + err.Error())
		}

		style := styles.Get(app.colorscheme)
		if style == nil {
			style = styles.Fallback
		}

		builder := style.Builder()
		builder.AddEntry(chroma.Background, chroma.MustParseStyleEntry("#2e2e2e"))
		style, _ = builder.Build()

		formatter := html.New(html.WithLineNumbers(true), html.WithClasses(false))

		if err := formatter.Format(&highlighted, style, iterator); err != nil {
			slog.Warn("Chroma formatting error: " + err.Error())
		}
	}

	fileInfo := FileInfo{
		Name:         file,
		Path:         filepath.Join(app.url, file),
		Type:         getType(file),
		Content:      template.HTML(highlighted.String()),
		TimeUploaded: fileStat.ModTime().UTC().Format("2006-01-02 15:04:05 UTC"),
	}

	if err := tmpl.Execute(w, fileInfo); err != nil {
		slog.Warn(err.Error())
	}
}

func getType(file string) string {
	extension := strings.ToLower(filepath.Ext(file))

	if fileType, exists := extensions[extension]; exists {
		return fileType
	}
	return "unknown"
}
