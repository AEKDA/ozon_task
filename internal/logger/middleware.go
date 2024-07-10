package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Middleware(log *Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		lrw := &loggingResponseWriter{w, http.StatusOK}
		next.ServeHTTP(lrw, r)

		duration := time.Since(startTime)
		log.Info("request",
			zap.String("method", r.Method), zap.String("url_path", r.URL.Path), zap.Duration("exec time", duration), zap.Int("status", lrw.statusCode))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
