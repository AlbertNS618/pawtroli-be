package middleware

import (
	"net/http"
	"time"

	"pawtroli-be/internal/logger"
)

// ResponseWriter wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log incoming request
		logger.LogInfof("Incoming request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Log the completed request
		duration := time.Since(start)
		logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, wrapped.statusCode, duration)
	})
}
