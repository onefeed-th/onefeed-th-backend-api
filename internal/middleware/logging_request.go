package middleware

import (
	"bytes"
	"io"
	"log"
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
		log.Printf("Received request: %s %s\n", r.Method, r.URL.Path)

		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading request body: %v\n", err)
			} else {
				log.Printf("Request body: %s\n", bodyBytes)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		next.ServeHTTP(w, r)

		log.Printf("Request duration: %v\n", time.Since(start))
	})
}
