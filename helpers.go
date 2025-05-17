package main

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
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
	Content       template.HTML
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
	}
	if err := validateToken(receivedKey, key); err != nil {
		slog.Warn("token validation failed", "error", err)
		return false
	}
	return true
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

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("invalid token")
	}

	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			return fmt.Errorf("token expired")
		}
	}

	return nil
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

	hashed := strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
	filename := fmt.Sprintf("%s%s", hashed, extension)
	if fullHash {
		return filename, nil
	} else {
		return fmt.Sprintf("%s%s", hashed[:5], extension), nil
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
			usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(app.auth.username)) == 1
			passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(app.auth.password)) == 1

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	})
}

func ResponseURLHandler(r *http.Request, w http.ResponseWriter, url, filename string) {
	protocol := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		protocol = "https"
	}
	pasteURL := fmt.Sprintf("%s://%s/%s\n", protocol, url, filename)

	w.Header().Set("Location", pasteURL)
	w.WriteHeader(http.StatusSeeOther)

	fmt.Fprintf(w, "%s", pasteURL)
}

const (
	green  = "\033[32m"
	blue   = "\033[34m"
	yellow = "\033[33m"
	red    = "\033[31m"
	reset  = "\033[0m"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func LogHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// dump request early
		x, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		fn(sr, r) // use wrapped writer to capture status

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

		fmt.Printf("%s%s%-6s%s %s => %s(%s) status=%d\n",
			reset, methodColor, r.Method, reset, r.URL.Path, green, duration, sr.status,
		)

		slog.Debug("Request Details",
			"method", r.Method,
			"url", r.URL.String(),
			"headers", r.Header,
			"body", string(x),
			"status", sr.status,
			"duration", duration.String(),
		)
	}
}
