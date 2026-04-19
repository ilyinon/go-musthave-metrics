package handler

import (
	"net"
	"net/http"
)

// extractIP returns the client IP address from the request.
// It prefers the X-Real-IP header and falls back to RemoteAddr.
func extractIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
