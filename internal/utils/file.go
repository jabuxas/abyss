package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var customStyle *chroma.Style

func DetectFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return "image"
		}
	}

	videoExts := []string{".mp4", ".webm", ".ogg", ".mov", ".avi", ".mkv"}
	for _, vidExt := range videoExts {
		if ext == vidExt {
			return "video"
		}
	}

	audioExts := []string{".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a"}
	for _, audExt := range audioExts {
		if ext == audExt {
			return "audio"
		}
	}

	if ext == ".pdf" {
		return "pdf"
	}

	textExts := []string{".txt", ".md", ".log", ".json", ".xml", ".html", ".css", ".js", ".go", ".py", ".java", ".c", ".cpp", ".h", ".rb", ".rs", ".sh", ".yml", ".yaml", ".ini", ".cfg", ".toml", ".csv", ".tsv", ".tex", ".el", ".php", ".rtf", ".srt", ".sub", ".vtt", ".sql", ".conf", ".bat", ".ps1", ".jsx", ".tsx", ".vue", ".scss", ".sass", ".less", ".pl", ".swift", ".kt", ".kts", ".groovy", ".r", ".lua", ".dockerfile", ".tf", ".diff", ".patch", ".asciidoc", ".rst", ".m", ".mm", ".f", ".f90", ".asm", ".vb", ".org"}
	for _, txtExt := range textExts {
		if ext == txtExt {
			return "text"
		}
	}

	return "unknown"
}

func FormatFileSize(size int64) string {
	const (
		KB = 1 << (10 * 1)
		MB = 1 << (10 * 2)
		GB = 1 << (10 * 3)
		TB = 1 << (10 * 4)
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func HashedName(filename string) string {
	var hash int
	for _, char := range time.Now().String() {
		hash = (hash << 3) - hash + int(char)
	}
	return fmt.Sprint(strings.ToUpper(fmt.Sprintf("%x", hash)[0:5]), filepath.Ext(filename))
}

func init() {
	customStyle = loadCustomStyle("assets/templates/colorscheme.xml")
}

func loadCustomStyle(path string) *chroma.Style {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to load custom style: %v, using fallback", err)
		return styles.Get("monokai")
	}

	style, err := chroma.NewXMLStyle(strings.NewReader(string(data)))
	if err != nil {
		log.Printf("failed to parse custom style: %v, using fallback", err)
		return styles.Get("monokai")
	}

	return style
}

func detectLexer(filename, content string) chroma.Lexer {
	ext := filepath.Ext(filename)
	if ext != "" {
		lexer := lexers.Match(filename)
		if lexer != nil {
			return lexer
		}
	}

	if content != "" {
		lexer := lexers.Analyse(content)
		if lexer != nil {
			return lexer
		}
	}

	return lexers.Fallback
}

func HighlightCode(code, filename string) (template.HTML, error) {
	lexer := detectLexer(filename, code)

	lexer = chroma.Coalesce(lexer)

	formatter := html.New(
		html.WithClasses(false),
		html.Standalone(false),
		html.WithLineNumbers(true),
		html.WithLinkableLineNumbers(true, ""),
		html.WrapLongLines(false),
	)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, customStyle, iterator)
	if err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}
