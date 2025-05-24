// internal/middleware/middleware.go
package middleware

import (
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/jabuxas/abyss/internal/app"
)

// statusRecorder captures the HTTP status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

const (
	ColorBlack   = "\033[30m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"

	ColorBrightBlack   = "\033[90m"
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"

	ColorReset = "\033[0m"
	ColorBold  = "\033[1m"
)

// getClientIP extracts the client IP address from the request.
// It considers X-Forwarded-For and X-Real-IP headers for proxies.
func getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		clientIP := strings.TrimSpace(ips[0])
		if clientIP != "" {
			return clientIP
		}
	}

	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Fallback to RemoteAddr if no proxy headers are found.
	remoteAddr := r.RemoteAddr
	if strings.Contains(remoteAddr, ":") {
		host, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			return host
		}
	}
	return remoteAddr
}

// LogHandler logs request details.
func LogHandler(logger *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		clientIP := getClientIP(r)

		// Dump request body only in debug
		var requestDump []byte
		var dumpErr error
		isDebugging := logger.Enabled(r.Context(), slog.LevelDebug)

		if isDebugging {
			requestDump, dumpErr = httputil.DumpRequest(r, true)
			if dumpErr != nil {
				logger.Error("Failed to dump request", "error", dumpErr, "client_ip", clientIP)
			}
		}

		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(sr, r)

		duration := time.Since(startTime)
		statusCode := sr.status

		statusColor := ColorGreen
		if statusCode >= 500 {
			statusColor = ColorBrightRed
		} else if statusCode >= 400 {
			statusColor = ColorRed
		} else if statusCode >= 300 {
			statusColor = ColorYellow
		}

		methodColor := ColorCyan
		switch r.Method {
		case http.MethodGet:
			methodColor = ColorBlue
		case http.MethodPost:
			methodColor = ColorGreen
		case http.MethodPut:
			methodColor = ColorYellow
		case http.MethodDelete:
			methodColor = ColorRed
		case http.MethodPatch:
			methodColor = ColorMagenta
		case http.MethodHead:
			methodColor = ColorBrightBlue
		case http.MethodOptions:
			methodColor = ColorBrightBlack
		}

		durationStr := fmt.Sprintf("%.2fms", float64(duration.Nanoseconds())/1e6)
		if duration >= time.Second {
			durationStr = fmt.Sprintf("%.2fs", duration.Seconds())
		}

		// [TIMESTAMP] IP | STATUS METHOD PATH (DURATION)
		fmt.Printf("[%s%s%s] %s%-15s%s | %s%3d%s | %s%s%-7s%s | %s%s%s (%s%s%s)\n",
			ColorBrightBlack, startTime.Format("2006-01-02 15:04:05.000"), ColorReset, // timestamp
			ColorBrightBlue, clientIP, ColorReset, // client IP
			statusColor, statusCode, ColorReset, // status
			ColorBold, methodColor, r.Method, ColorReset, // method
			ColorWhite, r.URL.RequestURI(), ColorReset, // path
			ColorMagenta, durationStr, ColorReset, // duration
		)

		logLevel := slog.LevelInfo
		if statusCode >= 500 {
			logLevel = slog.LevelError
		} else if statusCode >= 400 {
			logLevel = slog.LevelWarn
		}

		logAttrs := []slog.Attr{
			slog.Time("time", startTime),
			slog.String("client_ip", clientIP),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.Int("status", statusCode),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("duration_ns", duration.Nanoseconds()),
			slog.String("duration_pretty", durationStr),
			slog.String("proto", r.Proto),
			slog.Int64("content_length", r.ContentLength),
		}

		if r.Referer() != "" {
			logAttrs = append(logAttrs, slog.String("referer", r.Referer()))
		}

		if isDebugging && requestDump != nil {
			logAttrs = append(logAttrs, slog.String("debug_request_dump_preview", string(requestDump[:min(256, len(requestDump))])+"..."))
		}

		logger.LogAttrs(r.Context(), logLevel, "HTTP request handled", logAttrs...)
	}
}

func BasicAuth(application *app.Application, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			expectedUsername := application.Config.AuthUsername
			expectedPassword := application.Config.AuthPassword
			usernameMatch := (subtle.ConstantTimeCompare([]byte(username), []byte(expectedUsername)) == 1)
			passwordMatch := (subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
