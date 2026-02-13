package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ilyinon/go-musthave-metrics/internal/config"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
	filestorage "github.com/ilyinon/go-musthave-metrics/internal/repository/file"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/postgres"
	"github.com/ilyinon/go-musthave-metrics/internal/router"
)

func main() {
	storeInterval := 300 * time.Second
	storeFile := "./metrics-db.json"
	restore := true
	dsn := ""

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
	flag.StringVar(&dsn, "d", "", "database dsn")
	flag.DurationVar(&storeInterval, "i", storeInterval, "store interval")
	flag.StringVar(&storeFile, "f", storeFile, "storage file")
	flag.BoolVar(&restore, "r", restore, "restore metrics")
	flag.Parse()

	if dsn == "" {
		if v, ok := os.LookupEnv("DATABASE_DSN"); ok {
			dsn = v
		}
	}

	var storage repository.Storage

	if dsn != "" {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			log.Fatal(err)
		}
		_ = db.Stats()
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Fatal(err)
		}

		storage = postgres.New(db)
		log.Println("using postgres storage")
	} else {
		memStorage := mem.New()
		storage = memStorage

		fs := filestorage.New(memStorage, storeFile)

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

		log.Println("using memory/file storage")
	}

	handler := router.New(storage)

	log.Printf("starting server on %s", addr.String())
	log.Fatal(http.ListenAndServe(addr.String(), handler))
}
