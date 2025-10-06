package routing

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWTToken(c *gin.Context) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(2 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(KEY))
	if err != nil {
		c.String(http.StatusInternalServerError, "could not generate token")
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("%s", tokenString))
}

func IsAuthorized(c *gin.Context) bool {
	authHeader := c.GetHeader("X-Auth")
	if authHeader == "" {
		return false
	}

	if subtle.ConstantTimeCompare([]byte(authHeader), []byte(KEY)) == 1 {
		return true
	}

	token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(KEY), nil
	})

	if err != nil {
		return false
	}

	return token.Valid
}
