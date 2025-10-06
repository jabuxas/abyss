package routing

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jabuxas/abyss/internal/utils"
)

func IndexHandler(c *gin.Context) {
	c.File("assets/static/index.html")
}

func ServeFileHandler(c *gin.Context) {
	filename := c.Param("file")
	filePath := "./files/new/" + filename

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		c.String(http.StatusNotFound, "file not found")
		return
	}

	fileType := utils.DetectFileType(filename)

	if fileType == "unknown" {
		c.Redirect(http.StatusSeeOther, "/raw/"+filename)
	}

	fileData := FileData{
		Name:         filename,
		Path:         "/raw/" + filename,
		Extension:    fileType,
		UploadedDate: fileInfo.ModTime().Format("2001-01-01 00:00:00"),
	}

	if fileType == "text" {
		content, err := os.ReadFile(filePath)
		if err == nil {
			fileData.Content = string(content)
		}
	}

	c.HTML(http.StatusOK, "fileDisplay.html", gin.H{
		"data": fileData,
	})
}

func ServeRawFileHandler(c *gin.Context) {
	file := c.Param("file")
	log.Println("Serving file:", file)
	c.File("./files/new/" + file)
}

func UploadFileHandler(c *gin.Context) {
	if !IsAuthorized(c) {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}
	file, _ := c.FormFile("file")
	c.SaveUploadedFile(file, "./files/new/"+file.Filename)
	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
}

func ListFilesHandler(c *gin.Context) {
	dirPath := "./files/new"

	dirEntries, err := os.ReadDir(dirPath)
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
			UploadedDate:  info.ModTime().UTC().Format("2006-01-02 15:04:05 UTC"),
			FormattedSize: utils.FormatFileSize(info.Size()),
			Extension:     filepath.Ext(entry.Name()),
		})
	}

	c.HTML(http.StatusOK, "fileList.html", gin.H{
		"Files": fileInfos,
		"URL":   "localhost",
	})
}
