package routing

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func generateJWTToken(c *gin.Context) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(2 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(cfg.UploadKey))
	if err != nil {
		c.String(http.StatusInternalServerError, "could not generate token")
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("%s", tokenString))
}

func isAuthorized(c *gin.Context) bool {
	auth := c.GetHeader("X-Auth")
	if auth == "" {
		auth = c.PostForm("auth")
	}
	if auth == "" {
		auth = c.Query("auth")
	}

	if auth == "" {
		return false
	}

	if subtle.ConstantTimeCompare([]byte(auth), []byte(cfg.UploadKey)) == 1 {
		return true
	}

	token, err := jwt.Parse(auth, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.UploadKey), nil
	})

	if err != nil {
		return false
	}

	return token.Valid
}
