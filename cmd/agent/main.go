package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/agent"
	"github.com/ilyinon/go-musthave-metrics/internal/agent/sender"
	"github.com/ilyinon/go-musthave-metrics/internal/config"
)

func main() {
	// ===== defaults =====
	addr := &config.AgentAddress{
		Host: "localhost",
		Port: 8080,
	}
	poll := config.SecondsDuration(2 * time.Second)
	report := config.SecondsDuration(10 * time.Second)
	var key string

	// ===== env overrides =====
	if v, ok := os.LookupEnv("ADDRESS"); ok {
		_ = addr.Set(v)
	}

	if v, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		if d, err := time.ParseDuration(v); err == nil {
			poll = config.SecondsDuration(d)
		}
	}

	if v, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		if d, err := time.ParseDuration(v); err == nil {
			report = config.SecondsDuration(d)
		}
	}

	// ===== flags override env =====
	flag.Var(addr, "a", "server address host:port")
	flag.Var(&poll, "p", "poll interval")
	flag.Var(&report, "r", "report interval")
	flag.StringVar(&key, "k", "", "signing key")
	flag.Parse()

	if key == "" {
		key = os.Getenv("KEY")
	}

	client := sender.New(addr.String(), key)
	app := agent.New(client, addr.String(), time.Duration(poll), time.Duration(report))

	log.Printf(
		"agent started, server=%s poll=%s report=%s",
		addr.String(),
		time.Duration(poll),
		time.Duration(report),
	)

	app.Run()
}
