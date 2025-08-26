package middleware

import (
	"net/http"
	"time"

	"github.com/malakagl/kart-challenge/pkg/log"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		logger := log.WithCtx(r.Context())
		logger.Info().Msgf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		clientIP := r.RemoteAddr
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			clientIP = ip
		}
		logger.Info().Msgf("Request %s %s from %s -> %d processed in %s",
			r.Method, r.URL.Path, clientIP, rw.statusCode, duration)
	})
}
