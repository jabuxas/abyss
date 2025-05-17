package main

import (
	"bytes"
	"embed"
	"encoding/xml"
	"fmt"
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

//go:embed templates/fileDisplay.html
var filesTemplate embed.FS

//go:embed templates/colorscheme.xml
var colorschemeXML embed.FS

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
		lexer := getLexer(file, fileContent)

		iterator, err := lexer.Tokenise(nil, string(fileContent))
		if err != nil {
			slog.Warn("Chroma tokenizing error: " + err.Error())
		}

		style := LoadStyle(app)
		if style == nil {
			style = styles.Fallback
		}

		formatter := html.New(html.WithLineNumbers(true), html.TabWidth(4), html.WithClasses(false), html.WithLinkableLineNumbers(true, ""))

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

func LoadStyle(app *Application) *chroma.Style {
	var data []byte
	var err error

	if _, statErr := os.Stat("templates/colorscheme.xml"); statErr == nil {
		data, err = os.ReadFile("templates/colorscheme.xml")
	} else {
		data, err = colorschemeXML.ReadFile("templates/colorscheme.xml")
	}

	if err != nil {
		slog.Warn("failed to read style XML: " + err.Error())
		return styles.Fallback
	}

	var styleXML struct {
		Entries []struct {
			Token string `xml:"type,attr"`
			Value string `xml:"style,attr"`
		} `xml:"entry"`
	}

	if err := xml.Unmarshal(data, &styleXML); err != nil {
		slog.Warn("invalid XML: " + err.Error())
		return styles.Fallback
	}

	builder := chroma.NewStyleBuilder("custom")
	for _, e := range styleXML.Entries {
		if e.Token == "" {
			slog.Warn("Empty token found in style XML, skipping")
			continue
		}

		token, err := chroma.TokenTypeString(e.Token)
		if err != nil {
			slog.Warn(fmt.Sprintf("Invalid token type '%s': %s", e.Token, err.Error()))
			continue
		}

		styleEntry, err := chroma.ParseStyleEntry(e.Value)
		if err != nil {
			slog.Warn(fmt.Sprintf("Invalid style value '%s' for token '%s': %s", e.Value, e.Token, err.Error()))
			continue
		}

		builder.AddEntry(token, styleEntry)
	}

	style, err := builder.Build()
	if err != nil {
		slog.Warn("Style build failed: " + err.Error())
		return styles.Fallback
	}

	return style
}

func getLexer(file string, fileContent []byte) chroma.Lexer {
	ext := strings.ToLower(filepath.Ext(file))

	var lexer chroma.Lexer
	if ext == ".txt" {
		lexer = lexers.Analyse(string(fileContent))
	} else {
		lexer = lexers.Match(file)
		if lexer == nil {
			lexer = lexers.Analyse(string(fileContent))
		}
	}

	if lexer == nil {
		lexer = lexers.Fallback
	}

	return lexer
}

func getType(file string) string {
	extension := strings.ToLower(filepath.Ext(file))

	if fileType, exists := extensions[extension]; exists {
		return fileType
	}
	return "unknown"
}
