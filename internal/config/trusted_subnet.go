package config

import (
	"net"
	"strings"
)

// ParseTrustedSubnet parses a CIDR subnet. Empty values disable subnet checks.
func ParseTrustedSubnet(value string) (*net.IPNet, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	_, subnet, err := net.ParseCIDR(value)
	if err != nil {
		return nil, err
	}

	return subnet, nil
}
