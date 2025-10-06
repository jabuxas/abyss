package routing

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
		name:          filename,
		path:          "/raw/" + filename,
		extension:     fileType,
		uploaded_date: fileInfo.ModTime().Format("2001-01-01 00:00:00"),
	}

	if fileType == "text" {
		content, err := os.ReadFile(filePath)
		if err == nil {
			fileData.content = string(content)
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
