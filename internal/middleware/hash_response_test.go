package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilyinon/go-musthave-metrics/internal/crypto"
)

func TestHashSignerSetsHeaderBeforeResponseIsWritten(t *testing.T) {
	const key = "test-key"

	handler := HashSigner(key)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "hello")
	}

	wantHash := crypto.HashSHA256([]byte("hello"), key)
	if got := rec.Header().Get("HashSHA256"); got != wantHash {
		t.Fatalf("HashSHA256 = %q, want %q", got, wantHash)
	}
}
