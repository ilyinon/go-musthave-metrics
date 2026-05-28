package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/agent"
	"github.com/ilyinon/go-musthave-metrics/internal/agent/sender"
	"github.com/ilyinon/go-musthave-metrics/internal/buildinfo"
	"github.com/ilyinon/go-musthave-metrics/internal/config"
	"github.com/ilyinon/go-musthave-metrics/internal/crypto"
)

// main configures the agent from environment variables, flags, and optional JSON config file
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
	configured := make(map[string]bool)

	// environment variables override defaults
	if v, ok := os.LookupEnv("ADDRESS"); ok {
		configured["address"] = true
		if err := addr.Set(v); err != nil {
			log.Println("invalid ADDRESS:", err)
		}
	}

	if v, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		configured["poll"] = true
		if err := poll.Set(v); err != nil {
			log.Println("invalid POLL_INTERVAL:", err)
		}
	}

	if v, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		configured["report"] = true
		if err := report.Set(v); err != nil {
			log.Println("invalid REPORT_INTERVAL:", err)
		}
	}

	if v, ok := os.LookupEnv("RATE_LIMIT"); ok {
		configured["rateLimit"] = true
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rateLimit = n
		} else if err != nil {
			log.Println("invalid RATE_LIMIT:", err)
		}
	}
	if v, ok := os.LookupEnv("KEY"); ok {
		configured["key"] = true
		key = v
	}
	if v, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		configured["cryptoKey"] = true
		cryptoKeyPath = v
	}

	// command-line flags override environment variables
	flag.Var(addr, "a", "server address host:port")
	flag.Var(&poll, "p", "poll interval")
	flag.Var(&report, "r", "report interval")
	flag.StringVar(&key, "k", key, "signing key")
	flag.IntVar(&rateLimit, "l", rateLimit, "rate limit")
	flag.StringVar(&cryptoKeyPath, "crypto-key", cryptoKeyPath, "path to public key file for encryption")

	// JSON config file support
	var configFile string
	flag.StringVar(&configFile, "c", "", "path to JSON config file")
	flag.StringVar(&configFile, "config", "", "path to JSON config file")

	// парсим флаги до чтения config, чтобы -c/-config реально задавали путь к файлу
	flag.Parse()
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			configured["address"] = true
		case "p":
			configured["poll"] = true
		case "r":
			configured["report"] = true
		case "k":
			configured["key"] = true
		case "l":
			configured["rateLimit"] = true
		case "crypto-key":
			configured["cryptoKey"] = true
		}
	})

	if configFile == "" {
		if env := os.Getenv("CONFIG"); env != "" {
			configFile = env
		}
	}

	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatalf("failed to read config file: %v", err)
		}

		var cfg config.AgentConfigFile
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("failed to parse config file: %v", err)
		}

		// применяем значения из JSON только если флаги/ENV не заданы
		if cfg.Address != "" && !configured["address"] {
			if err := addr.Set(cfg.Address); err != nil {
				log.Println("invalid address in config file:", err)
			}
		}
		if cfg.PollInterval != "" && !configured["poll"] {
			if err := poll.Set(cfg.PollInterval); err != nil {
				log.Println("invalid poll_interval in config file:", err)
			}
		}
		if cfg.ReportInterval != "" && !configured["report"] {
			if err := report.Set(cfg.ReportInterval); err != nil {
				log.Println("invalid report_interval in config file:", err)
			}
		}
		if cfg.CryptoKey != "" && !configured["cryptoKey"] {
			cryptoKeyPath = cfg.CryptoKey
		}
	}

	// загружаем публичный ключ
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

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	app.Run(ctx)
	log.Println("agent stopped")
}
