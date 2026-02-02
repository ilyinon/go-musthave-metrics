package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/ilyinon/go-musthave-metrics/internal/config"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
	"github.com/ilyinon/go-musthave-metrics/internal/router"
)

func main() {
	// ===== default =====
	addr := &config.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}

	// ===== env override =====
	if v, ok := os.LookupEnv("ADDRESS"); ok {
		_ = addr.Set(v)
	}

	// ===== flag override env =====
	flag.Var(addr, "a", "server address")
	flag.Parse()

	storage := mem.New()
	handler := router.New(storage)

	log.Printf("starting server on %s", addr.String())
	log.Fatal(http.ListenAndServe(addr.String(), handler))
}
