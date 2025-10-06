package routing

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func IndexHandler(c *gin.Context) {
	c.File("assets/static/index.html")
}

func ServeRawFileHandler(c *gin.Context) {
	file := c.Param("file")
	log.Println("Serving file:", file)
	c.File("./files/new/" + file)
}

func UploadFileHandler(c *gin.Context) {
	file, _ := c.FormFile("file")
	c.SaveUploadedFile(file, "./files/new/"+file.Filename)
	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
}
