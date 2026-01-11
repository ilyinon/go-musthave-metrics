package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
)

func TestRouter_FullFlow(t *testing.T) {
	store := mem.New()
	r := New(store)

	server := httptest.NewServer(r)
	defer server.Close()

	http.Post(server.URL+"/update/gauge/Load/12.5", "text/plain", nil)

	resp, _ := http.Get(server.URL + "/value/gauge/Load")
	body, _ := io.ReadAll(resp.Body)

	defer resp.Body.Close()

	if string(body) != "12.5" {
		t.Fatalf("expected 12.5, got %s", body)
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	store := mem.New()
	r := New(store)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, _ := http.Get(server.URL + "/update/gauge/test/1")
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
}
