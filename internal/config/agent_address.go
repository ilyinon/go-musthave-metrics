package config

import (
	"fmt"
	"net"
	"strconv"
)

// AgentAddress represents a server address with host and port.
type AgentAddress struct {
	Host string
	Port int
}

// String returns the full HTTP address in "http://host:port" format.
func (a *AgentAddress) String() string {
	return fmt.Sprintf("http://%s:%d", a.Host, a.Port)
}

// Set parses a "host:port" string and updates the address fields.
func (a *AgentAddress) Set(value string) error {
	host, portStr, err := net.SplitHostPort(value)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}
	a.Host = host
	a.Port = port
	if a.Host == "" {
		a.Host = "localhost"
	}
	return nil
}
