package routing

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gin-gonic/gin"
	"github.com/jabuxas/abyss/internal/utils"
)

var customStyle *chroma.Style

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

func highlightCode(code, filename string) (template.HTML, error) {
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

func indexHandler(c *gin.Context) {
	c.File("assets/static/index.html")
}

func serveFileHandler(c *gin.Context) {
	filename := c.Param("file")
	filePath := filepath.Join(cfg.FilesDir, filename)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		c.String(http.StatusNotFound, "file not found")
		return
	}

	fileType := utils.DetectFileType(filename)
	if fileType == "unknown" {
		c.Redirect(http.StatusSeeOther, "/raw/"+filename)
		return
	}

	fileData := FileData{
		Name:       filename,
		Path:       "/raw/" + filename,
		Extension:  fileType,
		ModTimeStr: fileInfo.ModTime().Format("2006-01-02 15:04:05"),
	}

	if fileType == "text" {
		content, err := os.ReadFile(filePath)
		if err == nil {
			fileData.Content = string(content)

			highlighted, err := highlightCode(fileData.Content, filename)
			if err != nil {
				log.Printf("failed to highlight code: %v", err)
				fileData.Content = string(content)
			} else {
				fileData.HighlightedContent = highlighted
			}
		}
	}

	c.HTML(http.StatusOK, "fileDisplay.html", gin.H{
		"data": fileData,
	})
}

func serveRawFileHandler(c *gin.Context) {
	file := c.Param("file")
	log.Println("Serving file:", file)
	c.File(filepath.Join(cfg.FilesDir, file))
}

func uploadFileHandler(c *gin.Context) {
	if !isAuthorized(c) {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}
	file, _ := c.FormFile("file")
	fileName := utils.HashedName(file.Filename)
	savePath := filepath.Join(cfg.FilesDir, fileName)
	c.SaveUploadedFile(file, savePath)

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	fullURL := fmt.Sprintf("%s://%s/%s", scheme, c.Request.Host, fileName)

	c.String(http.StatusOK, fullURL+"\n")
}

func listFilesHandler(c *gin.Context) {
	dirEntries, err := os.ReadDir(cfg.FilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			c.String(http.StatusNotFound, "directory not found")
			return
		}
		c.String(http.StatusInternalServerError, "failed to read directory")
		return
	}

	var fileInfos []FileData
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			log.Println("failed to get file info for", entry.Name(), ":", err)
			continue
		}

		linkPath := filepath.Join("/" + entry.Name())
		fileInfos = append(fileInfos, FileData{
			Name:          entry.Name(),
			Path:          linkPath,
			ModTimeStr:    info.ModTime().UTC().Format("2006-01-02 15:04:05 UTC"),
			ModTime:       info.ModTime(),
			FormattedSize: utils.FormatFileSize(info.Size()),
			Extension:     filepath.Ext(entry.Name()),
		})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].ModTime.After(fileInfos[j].ModTime)
	})

	c.HTML(http.StatusOK, "fileList.html", gin.H{
		"Files": fileInfos,
		"URL":   cfg.AbyssURL,
	})
}
