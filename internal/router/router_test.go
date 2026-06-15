package router

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilyinon/go-musthave-metrics/internal/realip"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
)

func TestRouter_FullFlow(t *testing.T) {
	store := mem.New()
	r := New(store, "", nil, nil)

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
	r := New(store, "", nil, nil)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, _ := http.Get(server.URL + "/update/gauge/test/1")
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
}

func TestRouter_TrustedSubnetAllowsMetricUpdate(t *testing.T) {
	store := mem.New()
	r := New(store, "", nil, mustSubnet(t, "192.168.1.0/24"))

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Load/12.5", nil)
	req.Header.Set(realip.Header, "192.168.1.10")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRouter_TrustedSubnetRejectsMetricUpdate(t *testing.T) {
	store := mem.New()
	r := New(store, "", nil, mustSubnet(t, "192.168.1.0/24"))

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Load/12.5", nil)
	req.Header.Set(realip.Header, "10.0.0.1")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if _, ok := store.GetGauge(req.Context(), "Load"); ok {
		t.Fatal("metric was updated from untrusted IP")
	}
}

func TestRouter_TrustedSubnetDoesNotAffectValueRoutes(t *testing.T) {
	store := mem.New()
	store.UpdateGauge(context.Background(), "Load", 12.5)
	r := New(store, "", nil, mustSubnet(t, "192.168.1.0/24"))

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/Load", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "12.5" {
		t.Fatalf("expected 12.5, got %s", body)
	}
}

func mustSubnet(t *testing.T, cidr string) *net.IPNet {
	t.Helper()

	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatal(err)
	}

	return subnet
}
