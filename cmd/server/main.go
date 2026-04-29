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

	_ "net/http/pprof"

	"github.com/ilyinon/go-musthave-metrics/internal/audit"
	"github.com/ilyinon/go-musthave-metrics/internal/config"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
	filestorage "github.com/ilyinon/go-musthave-metrics/internal/repository/file"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/postgres"
	"github.com/ilyinon/go-musthave-metrics/internal/router"
)

// main configures application components, initializes storage,
// sets up audit sinks and starts the HTTP server.
func main() {
	storeInterval := 300 * time.Second
	storeFile := "./metrics-db.json"
	restore := true
	dsn := ""
	var key string

	var auditFile string
	var auditURL string

	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Println("pprof server error:", err)
		}
	}()

	addr := &config.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}

	// read configuration from environment variables
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
	if v, ok := os.LookupEnv("AUDIT_FILE"); ok {
		auditFile = v
	}
	if v, ok := os.LookupEnv("AUDIT_URL"); ok {
		auditURL = v
	}

	// parse command-line flags
	flag.Var(addr, "a", "server address")
	flag.StringVar(&dsn, "d", "", "database dsn")
	flag.DurationVar(&storeInterval, "i", storeInterval, "store interval")
	flag.StringVar(&storeFile, "f", storeFile, "storage file")
	flag.BoolVar(&restore, "r", restore, "restore metrics")
	flag.StringVar(&key, "k", "", "signing key")
	flag.StringVar(&auditFile, "audit-file", auditFile, "audit file path")
	flag.StringVar(&auditURL, "audit-url", auditURL, "audit url")
	flag.Parse()

	if dsn == "" {
		dsn = os.Getenv("DATABASE_DSN")
	}
	if key == "" {
		key = os.Getenv("KEY")
	}

	// initialize audit sinks
	var auditor *audit.Auditor
	var sinks []audit.Sink

	if auditFile != "" {
		sink, err := audit.NewFileSink(auditFile)
		if err != nil {
			log.Fatalf("failed to init audit file sink: %v", err)
		}
		sinks = append(sinks, sink)
	}

	if auditURL != "" {
		sinks = append(sinks, audit.NewHTTPSink(auditURL))
	}

	if len(sinks) > 0 {
		auditor = audit.New(sinks...)
	}

	// initialize storage backend
	var storage repository.Storage

	if dsn != "" {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		_ = db.Stats()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			log.Fatal(err)
		}

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

		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}

		storage = postgres.New(db)
		log.Println("using postgres storage")

	} else if storeFile != "" {
		memStorage := mem.New()
		storage = memStorage

		fs := filestorage.New(memStorage, storeFile)

		if restore {
			if err := fs.Restore(); err != nil {
				log.Println("restore error:", err)
			}
		}

		if storeInterval > 0 {
			go func() {
				t := time.NewTicker(storeInterval)
				defer t.Stop()
				for range t.C {
					if err := fs.Save(); err != nil {
						log.Println("save error:", err)
					}
				}
			}()
		}

		log.Println("using file storage")

	} else {
		storage = mem.New()
		log.Println("using memory storage")
	}

	// start HTTP server
	handler := router.New(storage, key, auditor)

	log.Printf("starting server on %s", addr.String())
	log.Fatal(http.ListenAndServe(addr.String(), handler))
}