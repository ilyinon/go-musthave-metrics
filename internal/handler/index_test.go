package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
)

func TestIndexHandler(t *testing.T) {
	ctx := context.Background()

	store := mem.New()
	store.UpdateGauge(ctx, "g1", 0.5)
	store.UpdateCounter(ctx, "c1", 7)

	h := NewIndex(store)
	server := httptest.NewServer(h)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	if !strings.Contains(content, "Gauge g1") {
		t.Fatal("gauge not found in response")
	}
	if !strings.Contains(content, "Counter c1") {
		t.Fatal("counter not found in response")
	}
}
