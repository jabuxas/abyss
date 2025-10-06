package routing

import (
	"github.com/gin-gonic/gin"
)

func GetRouter() gin.Engine {
	r := gin.Default()

	r.Static("/static", "./assets/static")

	r.GET("/", IndexHandler)
	r.GET("/raw/:file", ServeRawFileHandler)
	r.POST("/upload", UploadFileHandler)

	return *r
}
