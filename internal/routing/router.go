package routing

import (
	"github.com/gin-gonic/gin"
)

var KEY = "1"

func GetRouter() gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("assets/templates/*")
	r.Static("/static", "./assets/static")

	r.GET("/", IndexHandler)
	r.GET("/:file", ServeFileHandler)
	r.GET("/raw/:file", ServeRawFileHandler)
	r.POST("/upload", UploadFileHandler)

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo": "bar",
	}))
	authorized.GET("/token", GenerateJWTToken)
	authorized.GET("/list", ListFilesHandler)

	return *r
}
