package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type ServerAddress struct {
	Host string
	Port int
}

func (a *ServerAddress) String() string {
	host := a.Host
	if host == "" {
		host = "localhost"
	}
	port := a.Port
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf("%s:%d", host, port)
}

func (a *ServerAddress) Set(value string) error {
	if !strings.Contains(value, ":") {
		value += ":8080"
	}
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
	return nil
}
