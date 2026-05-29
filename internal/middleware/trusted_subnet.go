package middleware

import (
	"net"
	"net/http"
	"strings"
)

const realIPHeader = "X-Real-IP"

// TrustedSubnet restricts metric update requests to the configured trusted subnet.
func TrustedSubnet(subnet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if subnet == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isMetricUpdateRequest(r) {
				next.ServeHTTP(w, r)
				return
			}

			ip := net.ParseIP(strings.TrimSpace(r.Header.Get(realIPHeader)))
			if ip == nil || !subnet.Contains(ip) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isMetricUpdateRequest(r *http.Request) bool {
	if r.Method != http.MethodPost {
		return false
	}

	path := r.URL.Path
	return path == "/update" ||
		path == "/update/" ||
		strings.HasPrefix(path, "/update/") ||
		path == "/updates/" ||
		strings.HasPrefix(path, "/updates/")
}
