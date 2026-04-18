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
	r := New(store, nil)

	server := httptest.NewServer(r)
	defer server.Close()

	// POST
	postResp, err := http.Post(server.URL+"/update/gauge/Load/12.5", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	postResp.Body.Close()

	// GET
	resp, err := http.Get(server.URL + "/value/gauge/Load")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "12.5" {
		t.Fatalf("expected 12.5, got %s", body)
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	store := mem.New()
	r := New(store, nil)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, _ := http.Get(server.URL + "/update/gauge/test/1")
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
}
