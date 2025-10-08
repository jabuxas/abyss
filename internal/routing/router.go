package routing

import (
	"log"

	"github.com/gin-gonic/gin"
)

var cfg *Config

func GetRouter() gin.Engine {
	var err error

	cfg, err = newConfig()
	if err != nil {
		log.Println("failed to load config from environment variables or .env file, did you run generate_config.sh?")
		log.Panic("error loading config:", err)
	}

	r := gin.Default()

	r.LoadHTMLGlob("assets/templates/*")
	r.Static("/static", "./assets/static")

	r.GET("/", indexHandler)
	r.GET("/:file", serveFileHandler)
	r.GET("/raw/:file", serveRawFileHandler)
	r.POST("/upload", uploadFileHandler)

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		cfg.AuthUsername: cfg.AuthPassword,
	}))

	authorized.GET("/token", generateJWTToken)
	authorized.GET("/all", listFilesHandler)

	return *r
}
