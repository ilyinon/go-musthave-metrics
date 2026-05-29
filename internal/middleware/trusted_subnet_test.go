package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustedSubnetNilAllowsMetricUpdateWithoutHeader(t *testing.T) {
	handler := TrustedSubnet(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/1", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestTrustedSubnetRejectsMissingRealIP(t *testing.T) {
	handler := TrustedSubnet(mustTrustedSubnet(t, "192.168.1.0/24"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/1", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func mustTrustedSubnet(t *testing.T, cidr string) *net.IPNet {
	t.Helper()

	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatal(err)
	}

	return subnet
}
