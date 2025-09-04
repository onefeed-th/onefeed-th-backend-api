package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		slog.Info("Received request: %s %s\n", r.Method, r.URL.Path)

		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("Error reading request body", "error", err)
			} else {
				slog.Info("Request body", "body", string(bodyBytes))
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		next.ServeHTTP(w, r)

		slog.Info("Request finished", "duration", time.Since(start))
	})
}
