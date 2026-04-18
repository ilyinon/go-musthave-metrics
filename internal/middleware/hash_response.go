package middleware

import (
	"bytes"
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/crypto"
)

type hashResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (w *hashResponseWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

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
			}

			next.ServeHTTP(hw, r)

			hash := crypto.HashSHA256(buf.Bytes(), key)
			w.Header().Set("HashSHA256", hash)
		})
	}
}
