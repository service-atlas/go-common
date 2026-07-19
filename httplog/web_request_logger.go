package httplog

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// requestKey is the slog attribute key used for the request id in logs.
const requestKey = "request_id"

// loggerKey is an unexported type used as the context key for *slog.Logger.
type loggerKey struct{}

// requestIDKey is an unexported type used as the context key for the request id value.
type requestIDKey struct{}

// WebRequestLogger returns a middleware that logs each request in structured (JSON) form using slog.
func WebRequestLogger(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := r.Header.Get("Request-Id")
		if id == "" {
			id = uuid.NewString()
		}
		logger := slog.Default().With(slog.String(requestKey, id))
		ctx := context.WithValue(r.Context(), loggerKey{}, logger)
		ctx = context.WithValue(ctx, requestIDKey{}, id)
		// Use a wrapper to get the status code
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r.WithContext(ctx))

		duration := time.Since(start)

		logger.Info("WEB_REQUEST",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", ww.status),
			slog.String("remote", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("duration_ms", duration.Milliseconds()),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// LoggerFromContext returns the *slog.Logger stored in context, or slog.Default() if missing.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

// GetRequestIdFromContext retrieves the request id stored in the context.
// It returns an empty string if the value is missing or not a string.
func GetRequestIdFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}

// GetRequestId retrieves the request id stored in the request's context.
// It returns an empty string if the value is missing or not a string.
func GetRequestId(r *http.Request) string {
	if r == nil || r.Context() == nil {
		return ""
	}
	if v, ok := r.Context().Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}
