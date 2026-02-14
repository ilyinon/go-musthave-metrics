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

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ilyinon/go-musthave-metrics/internal/config"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
	filestorage "github.com/ilyinon/go-musthave-metrics/internal/repository/file"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/postgres"
	"github.com/ilyinon/go-musthave-metrics/internal/router"
)

func main() {
	// defaults
	storeInterval := 300 * time.Second
	storeFile := "./metrics-db.json"
	restore := true
	dsn := ""

	addr := &config.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}

	// env (basic settings)
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

	// flags
	flag.Var(addr, "a", "server address")
	flag.StringVar(&dsn, "d", "", "database dsn")
	flag.DurationVar(&storeInterval, "i", storeInterval, "store interval")
	flag.StringVar(&storeFile, "f", storeFile, "storage file")
	flag.BoolVar(&restore, "r", restore, "restore metrics")
	flag.Parse()

	// env fallback for DSN
	if dsn == "" {
		dsn = os.Getenv("DATABASE_DSN")
	}

	var storage repository.Storage

	// choose storage
	if dsn != "" {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// explicit usage for tests
		_ = db.Stats()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Fatal(err)
		}

		// run migrations
		driver, err := migratepg.WithInstance(db, &migratepg.Config{})
		if err != nil {
			log.Fatal(err)
		}

		m, err := migrate.NewWithDatabaseInstance(
			"file://migrations",
			"postgres",
			driver,
		)
		if err != nil {
			log.Fatal(err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}

		storage = postgres.New(db)
		log.Println("using postgres storage")
	} else if storeFile != "" {
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

		log.Println("using file storage")
	} else {
		storage = mem.New()
		log.Println("using memory storage")
	}

	handler := router.New(storage)

	log.Printf("starting server on %s", addr.String())
	log.Fatal(http.ListenAndServe(addr.String(), handler))
}
