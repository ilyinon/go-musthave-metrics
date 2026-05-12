package middleware

import (
	"bytes"
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/crypto"
)

type hashResponseWriter struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
	wroteHead  bool
}

func (w *hashResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHead {
		w.WriteHeader(http.StatusOK)
	}

	return w.buf.Write(b)
}

func (w *hashResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHead {
		return
	}

	w.statusCode = statusCode
	w.wroteHead = true
}

// HashSigner is a middleware that calculates HMAC-SHA256 hash
// of the response body and adds it to the HashSHA256 header.
func HashSigner(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			buf := &bytes.Buffer{}
			hw := &hashResponseWriter{
				ResponseWriter: w,
				buf:            buf,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(hw, r)

			hash := crypto.HashSHA256(buf.Bytes(), key)
			w.Header().Set("HashSHA256", hash)
			w.WriteHeader(hw.statusCode)
			_, _ = w.Write(buf.Bytes())
		})
	}
}
