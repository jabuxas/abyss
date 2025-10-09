package routing

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jabuxas/abyss/internal/utils"
)

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

			highlighted, err := utils.HighlightCode(fileData.Content, filename)
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

	var expiry *time.Time
	var passwordHash []byte
	var err error

	if len(c.Request.Form["expiration"]) > 0 {
		expiry, err = utils.ParseExpiration(c.Request.FormValue("expiration"))
		if err != nil {
			// idk yet
		}
	}

	if len(c.Request.Form["password"]) > 0 {
		passwordHash, err = utils.ParsePassword(c.Request.FormValue("password"))
		if err != nil {
			// idk yet
		}
	}

	err = utils.SaveMetadata(savePath, expiry, passwordHash)
	if err != nil {
		// idk
	}

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
