package main

import (
	"flag"
	"log"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/agent"
	"github.com/ilyinon/go-musthave-metrics/internal/agent/sender"
	"github.com/ilyinon/go-musthave-metrics/internal/config"
)

func main() {
	addr := &config.AgentAddress{
		Host: "localhost",
		Port: 8080,
	}

	var poll = config.SecondsDuration(2 * time.Second)
	var report = config.SecondsDuration(10 * time.Second)

	flag.Var(addr, "a", "server address host:port")
	flag.Var(&poll, "p", "poll interval")
	flag.Var(&report, "r", "report interval")
	flag.Parse()

	client := sender.New(addr.String())
	app := agent.New(client, time.Duration(poll), time.Duration(report))

	log.Printf("agent started, server=%s", addr.String())
	app.Run()
}
