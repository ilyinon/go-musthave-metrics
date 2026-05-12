package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"crypto/rsa"

	"github.com/ilyinon/go-musthave-metrics/internal/agent"
	"github.com/ilyinon/go-musthave-metrics/internal/agent/sender"
	"github.com/ilyinon/go-musthave-metrics/internal/buildinfo"
	"github.com/ilyinon/go-musthave-metrics/internal/config"
	"github.com/ilyinon/go-musthave-metrics/internal/crypto"
)

// AgentConfigFile описывает формат JSON конфигурации для агента
type AgentConfigFile struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

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

	// JSON config file support
	var configFile string
	flag.StringVar(&configFile, "c", "", "path to JSON config file")
	flag.StringVar(&configFile, "config", "", "path to JSON config file")

	// читаем config до flag.Parse, чтобы применить только если значения дефолтные
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

		var cfg AgentConfigFile
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("failed to parse config file: %v", err)
		}

		// применяем значения из JSON только если флаги/ENV не заданы
		if cfg.Address != "" && addr.String() == "http://localhost:8080" {
			if err := addr.Set(cfg.Address); err != nil {
				log.Println("invalid address in config file:", err)
			}
		}
		if cfg.PollInterval != "" && poll == config.SecondsDuration(2*time.Second) {
			if d, err := time.ParseDuration(cfg.PollInterval); err == nil {
				poll = config.SecondsDuration(d)
			} else {
				log.Println("invalid poll_interval in config file:", err)
			}
		}
		if cfg.ReportInterval != "" && report == config.SecondsDuration(10*time.Second) {
			if d, err := time.ParseDuration(cfg.ReportInterval); err == nil {
				report = config.SecondsDuration(d)
			} else {
				log.Println("invalid report_interval in config file:", err)
			}
		}
		if cfg.CryptoKey != "" && cryptoKeyPath == "" {
			cryptoKeyPath = cfg.CryptoKey
		}
	}

	// парсим флаги, чтобы они имели самый высокий приоритет
	flag.Parse()

	// fallback на переменные окружения, если флаги и JSON не задали
	if key == "" {
		key = os.Getenv("KEY")
	}

	if cryptoKeyPath == "" {
		cryptoKeyPath = os.Getenv("CRYPTO_KEY")
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

	app.Run()
}
