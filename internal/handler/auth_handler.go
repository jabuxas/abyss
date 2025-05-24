package handler

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jabuxas/abyss/internal/app"
)

// CreateJWTHandler generates and returns a JWT token.
func (h *Handler) CreateJWTHandler(w http.ResponseWriter, r *http.Request) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(2 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(h.App.JWTSecret)
	if err != nil {
		h.App.Logger.Error("Failed to sign JWT token", "error", err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, tokenString)
}

// CheckUploadAuth validates the X-Auth header (API key or JWT).
func CheckUploadAuth(r *http.Request, appInstance *app.Application) bool {
	receivedKey := r.Header.Get("X-Auth")
	if receivedKey == "" {
		appInstance.Logger.Warn("X-Auth header missing")
		return false
	}

	if subtle.ConstantTimeCompare([]byte(receivedKey), []byte(appInstance.Config.UploadKey)) == 1 {
		return true
	}

	if err := ValidateJWT(receivedKey, appInstance.JWTSecret); err == nil {
		return true
	} else {
		appInstance.Logger.Warn("JWT validation failed", "error", err, "token_prefix", receivedKey[:min(10, len(receivedKey))]+"...")
		return false
	}
}

// ValidateJWT parses and validates a JWT token string.
func ValidateJWT(tokenString string, secretKey []byte) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return fmt.Errorf("JWT parsing error: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("invalid JWT token or claims")
	}

	if expFloat, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(expFloat), 0)
		if time.Now().After(expTime) {
			return fmt.Errorf("JWT token has expired")
		}
	} else {
		return fmt.Errorf("JWT 'exp' claim missing or not a number")
	}

	return nil
}
