package config

import (
	"fmt"
	"net"
	"strconv"
)

type AgentAddress struct {
	Host string
	Port int
}

func (a *AgentAddress) String() string {
	return fmt.Sprintf("http://%s:%d", a.Host, a.Port)
}

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
