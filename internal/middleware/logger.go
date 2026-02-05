package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type loggerKey struct{}

func NewSlogLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := middleware.GetReqID(r.Context())

			requestLogger := logger.With(
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)

			ctx := context.WithValue(r.Context(), loggerKey{}, requestLogger)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMinor)
			startTime := time.Now()
			requestLogger.Info("Request started")
			next.ServeHTTP(ww, r.WithContext(ctx))
			duration := time.Since(startTime)
			requestLogger.Info("Request completed",
				slog.Int("status", ww.Status()),
				slog.Int("bytes_written", ww.BytesWritten()),
				slog.Float64("duration", float64(duration.Milliseconds())))
		})
	}
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
