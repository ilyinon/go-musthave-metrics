package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/config"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
	"github.com/ilyinon/go-musthave-metrics/internal/router"
)

func main() {
	addr := &config.ServerAddress{}
	flag.Var(addr, "a", "server address")
	flag.Parse()

	storage := mem.New()
	handler := router.New(storage)

	log.Printf("starting server on %s", addr.String())
	log.Fatal(http.ListenAndServe(addr.String(), handler))
}
