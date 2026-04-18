package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/crypto"
)

func HashVerifier(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			received := r.Header.Get("HashSHA256")
			if received == "" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()

			expected := crypto.HashSHA256(body, key)
			if !hmacEqual(received, expected) {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}

func hmacEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var res byte
	for i := 0; i < len(a); i++ {
		res |= a[i] ^ b[i]
	}
	return res == 0
}
