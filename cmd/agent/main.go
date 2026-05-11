package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/agent"
	"github.com/ilyinon/go-musthave-metrics/internal/agent/sender"
	"github.com/ilyinon/go-musthave-metrics/internal/buildinfo"
	"github.com/ilyinon/go-musthave-metrics/internal/config"
	"github.com/ilyinon/go-musthave-metrics/internal/crypto"
	"crypto/rsa"
)

// main configures the agent from environment variables and flags,
// initializes the sender and starts metric collection and reporting.
func main() {

	buildinfo.Print()

	// default configuration values
	addr := &config.AgentAddress{
		Host: "localhost",
		Port: 8080,
	}
	poll := config.SecondsDuration(2 * time.Second)
	report := config.SecondsDuration(10 * time.Second)
	var key string
	var cryptoKeyPath string

	rateLimit := 1
	var publicKey *rsa.PublicKey

	// environment variables override defaults
	if v, ok := os.LookupEnv("ADDRESS"); ok {
		if err := addr.Set(v); err != nil {
			log.Println("invalid ADDRESS:", err)
		}
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

	if v, ok := os.LookupEnv("RATE_LIMIT"); ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rateLimit = n
		}
	}

	// command-line flags override environment variables
	flag.Var(addr, "a", "server address host:port")
	flag.Var(&poll, "p", "poll interval")
	flag.Var(&report, "r", "report interval")
	flag.StringVar(&key, "k", "", "signing key")
	flag.IntVar(&rateLimit, "l", rateLimit, "rate limit")
	flag.StringVar(&cryptoKeyPath, "crypto-key", "", "path to public key file for encryption")
	flag.Parse()

	if key == "" {
		key = os.Getenv("KEY")
	}

	if cryptoKeyPath == "" {
		cryptoKeyPath = os.Getenv("CRYPTO_KEY")
	}

	if cryptoKeyPath != "" {
		pubData, err := os.ReadFile(cryptoKeyPath)
		if err != nil {
			log.Fatalf("failed to read public key: %v", err)
		}
		pub, err := crypto.ParseRSAPublicKey(pubData)
		if err != nil {
			log.Fatalf("failed to parse public key: %v", err)
		}
		publicKey = pub
	}

	client := sender.New(addr.String(), key)
	client.SetPublicKey(publicKey)

	app := agent.New(
		client,
		addr.String(),
		time.Duration(poll),
		time.Duration(report),
		rateLimit,
	)

	log.Printf(
		"agent started, server=%s poll=%s report=%s rateLimit=%d",
		addr.String(),
		time.Duration(poll),
		time.Duration(report),
		rateLimit,
	)

	app.Run()
}
