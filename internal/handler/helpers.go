package handler

import (
	"net"
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/realip"
)

// extractIP returns the client IP address from the request.
// It prefers the real IP header and falls back to RemoteAddr.
func extractIP(r *http.Request) string {
	ip := r.Header.Get(realip.Header)
	if ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
