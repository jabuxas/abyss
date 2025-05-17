package main

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type FileInfo struct {
	Name          string
	Path          string
	Size          int64
	FormattedSize string
	Type          string
	Content       string
	TimeUploaded  string
}

type TemplateData struct {
	Files []FileInfo
	URL   string
}

func CheckAuth(r *http.Request, key string) bool {
	receivedKey := r.Header.Get("X-Auth")
	if receivedKey == key {
		return true
	} else if err := validateToken(receivedKey, key); err == nil {
		return true
	}
	return false
}

func validateToken(tokenString, key string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(key), nil
	})
	if err != nil {
		return err
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return nil
	} else {
		return fmt.Errorf("invalid token")
	}
}

func FormatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
}

func HashFile(file io.Reader, extension string, fullHash bool) (string, error) {
	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	sha1Hash := strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
	filename := fmt.Sprintf("%s%s", sha1Hash, extension)
	if fullHash {
		return filename, nil
	} else {
		return fmt.Sprintf("%s%s", sha1Hash[:5], extension), nil
	}
}

func SaveFile(name string, file io.Reader) error {
	dst, err := os.Create(name)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return err
	}
	return nil
}

func BasicAuth(next http.HandlerFunc, app *Application) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			// hash password received
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))

			// hash our password
			expectedUsernameHash := sha256.Sum256([]byte(app.auth.username))
			expectedPasswordHash := sha256.Sum256([]byte(app.auth.password))

			// compare hashes
			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func ResponseURLHandler(r *http.Request, w http.ResponseWriter, url, filename string) {
	protocol := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		protocol = "https"
	}
	pasteURL := fmt.Sprintf("%s://%s/%s\n", protocol, url, filename)

	w.Header().Set("Location", pasteURL)

	fmt.Fprintf(w, "%s", pasteURL)
}

const (
	green  = "\033[32m"
	blue   = "\033[34m"
	yellow = "\033[33m"
	red    = "\033[31m"
	reset  = "\033[0m"
)

func LogHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		x, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		fn(w, r)

		duration := time.Since(start)

		methodColor := map[string]string{
			"GET":    blue,
			"POST":   green,
			"PUT":    yellow,
			"DELETE": red,
		}[r.Method]
		if methodColor == "" {
			methodColor = reset
		}

		fmt.Printf("%s%s%-6s%s %s => %s(%s)\n",
			reset, methodColor, r.Method, reset, r.URL.Path, green, duration,
		)

		slog.Debug("Request Details",
			"method", r.Method,
			"url", r.URL.String(),
			"headers", r.Header,
			"body", string(x),
		)
	}
}
