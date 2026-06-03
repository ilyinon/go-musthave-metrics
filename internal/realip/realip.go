package realip

import (
	"errors"
	"net"
	"strings"
)

const (
	Header      = "X-Real-IP"
	MetadataKey = "x-real-ip"
)

var (
	ErrMissing   = errors.New("missing real ip")
	ErrUntrusted = errors.New("real ip is not trusted")
)

// CheckTrustedSubnet verifies that realIP belongs to subnet.
func CheckTrustedSubnet(subnet *net.IPNet, realIP string) error {
	if subnet == nil {
		return nil
	}

	realIP = strings.TrimSpace(realIP)
	if realIP == "" {
		return ErrMissing
	}

	ip := net.ParseIP(realIP)
	if ip == nil || !subnet.Contains(ip) {
		return ErrUntrusted
	}

	return nil
}
