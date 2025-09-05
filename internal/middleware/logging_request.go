package middleware

import (
	"bytes"
	"encoding/json"
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
		slog.Info("Received request", "method", r.Method, "path", r.URL.Path)

		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("Error reading request body", "error", err)
			} else {
				// try compact JSON
				var compactBuf bytes.Buffer
				if json.Compact(&compactBuf, bodyBytes) == nil {
					slog.Info("Request body", "body", compactBuf.String())
				} else {
					// fallback: raw body
					slog.Info("Request body (raw)", "body", string(bodyBytes))
				}

				// restore body for the next handler
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		next.ServeHTTP(w, r)

		slog.Info("Request finished", "duration", time.Since(start))
	})
}
