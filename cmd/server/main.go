package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/config"
	filestorage "github.com/ilyinon/go-musthave-metrics/internal/repository/file"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
	"github.com/ilyinon/go-musthave-metrics/internal/router"
)

func main() {
	storeInterval := 300 * time.Second
	storeFile := "./metrics-db.json"
	restore := true

	addr := &config.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}

	if v, ok := os.LookupEnv("ADDRESS"); ok {
		addr.Set(v)
	}

	if v, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		if sec, err := strconv.Atoi(v); err == nil {
			storeInterval = time.Duration(sec) * time.Second
		}
	}
	if v, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		storeFile = v
	}
	if v, ok := os.LookupEnv("RESTORE"); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			restore = b
		}
	}

	flag.Var(addr, "a", "server address")
	flag.DurationVar(&storeInterval, "i", storeInterval, "store interval")
	flag.StringVar(&storeFile, "f", storeFile, "storage file")
	flag.BoolVar(&restore, "r", restore, "restore metrics")
	flag.Parse()

	storage := mem.New()
	fs := filestorage.New(storage, storeFile)

	if restore {
		_ = fs.Restore()
	}

	if storeInterval > 0 {
		go func() {
			t := time.NewTicker(storeInterval)
			defer t.Stop()
			for range t.C {
				_ = fs.Save()
			}
		}()
	}

	handler := router.New(storage)

	log.Printf("starting server on %s", addr.String())
	log.Fatal(http.ListenAndServe(addr.String(), handler))
}
