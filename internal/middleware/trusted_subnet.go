package middleware

import (
	"net"
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/realip"
)

// TrustedSubnet restricts requests to the configured trusted subnet.
func TrustedSubnet(subnet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if subnet == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := realip.CheckTrustedSubnet(subnet, r.Header.Get(realip.Header)); err != nil {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
