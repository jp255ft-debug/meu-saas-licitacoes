package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// wrappedWriter wraps http.ResponseWriter to capture the status code.
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs structured request information using slog.
// Requires RequestIDMiddleware to be applied first.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		reqID := GetRequestID(r.Context())

		slog.InfoContext(r.Context(), "request",
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.Int("status", wrapped.statusCode),
			slog.Duration("duration", duration),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}