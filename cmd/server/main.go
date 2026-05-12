package main

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/audit"
	"github.com/ilyinon/go-musthave-metrics/internal/buildinfo"
	"github.com/ilyinon/go-musthave-metrics/internal/config"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
	filestorage "github.com/ilyinon/go-musthave-metrics/internal/repository/file"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/postgres"
	"github.com/ilyinon/go-musthave-metrics/internal/router"

	appmw "github.com/ilyinon/go-musthave-metrics/internal/middleware"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	_ "net/http/pprof"

	"github.com/ilyinon/go-musthave-metrics/internal/crypto"
)

type ServerConfigFile struct {
	Address       string `json:"address"`
	Restore       *bool  `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
}

// main configures application components, initializes storage,
// sets up audit sinks and starts the HTTP server.
func main() {
	storeInterval := 300 * time.Second
	storeFile := "./metrics-db.json"
	restore := true
	dsn := ""

	buildinfo.Print()

	var key string
	var cryptoKeyPath string

	var auditFile string
	var auditURL string

	pprofServer := &http.Server{Addr: "localhost:6060"}
	go func() {
		if err := pprofServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	// новый флаг для приватного ключа
	flag.StringVar(&cryptoKeyPath, "crypto-key", "", "path to private key for decryption")

	var configFile string
	flag.StringVar(&configFile, "c", "", "path to JSON config file")
	flag.StringVar(&configFile, "config", "", "path to JSON config file")

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

		var cfg ServerConfigFile
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("failed to parse config file: %v", err)
		}

		// Применяем только если флаг/ENV не установлен
		if cfg.Address != "" && addr.String() == "localhost:8080" {
			addr.Set(cfg.Address)
		}
		if cfg.StoreFile != "" && storeFile == "./metrics-db.json" {
			storeFile = cfg.StoreFile
		}
		if cfg.StoreInterval != "" && storeInterval == 300*time.Second {
			if d, err := time.ParseDuration(cfg.StoreInterval); err == nil {
				storeInterval = d
			}
		}
		if cfg.Restore != nil && !restore {
			restore = *cfg.Restore
		}
		if cfg.DatabaseDSN != "" && dsn == "" {
			dsn = cfg.DatabaseDSN
		}
		if cfg.CryptoKey != "" && cryptoKeyPath == "" {
			cryptoKeyPath = cfg.CryptoKey
		}
	}

	flag.Parse()

	if dsn == "" {
		dsn = os.Getenv("DATABASE_DSN")
	}
	if key == "" {
		key = os.Getenv("KEY")
	}
	if cryptoKeyPath == "" {
		cryptoKeyPath = os.Getenv("CRYPTO_KEY")
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

	// load private key for decryption
	var privateKey *rsa.PrivateKey
	if cryptoKeyPath != "" {
		privData, err := os.ReadFile(cryptoKeyPath)
		if err != nil {
			log.Fatalf("failed to read private key: %v", err)
		}
		privateKey, err = crypto.ParseRSAPrivateKey(privData)
		if err != nil {
			log.Fatalf("failed to parse private key: %v", err)
		}
	}

	// initialize storage backend
	var storage repository.Storage
	var fileStorage *filestorage.Storage
	var stopPeriodicSave func()

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
		fileStorage = fs

		if restore {
			if err := fs.Restore(); err != nil {
				log.Println("restore error:", err)
			}
		}

		if storeInterval > 0 {
			saveCtx, stopSave := context.WithCancel(context.Background())
			saveDone := make(chan struct{})
			stopPeriodicSave = func() {
				stopSave()
				<-saveDone
			}

			go func() {
				defer close(saveDone)

				t := time.NewTicker(storeInterval)
				defer t.Stop()

				for {
					select {
					case <-t.C:
						if err := fs.Save(); err != nil {
							log.Println("save error:", err)
						}
					case <-saveCtx.Done():
						return
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
	if privateKey != nil {
		handler = appmw.DecryptRSA(privateKey)(handler)
	}

	server := &http.Server{
		Addr:    addr.String(),
		Handler: handler,
	}
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("starting server on %s", addr.String())
		serverErr <- server.ListenAndServe()
	}()

	shutdownCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}

	case <-shutdownCtx.Done():
		log.Println("shutdown signal received")
		stop()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}

		if err := pprofServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("pprof shutdown error: %v", err)
		}

		if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
		}

		if stopPeriodicSave != nil {
			stopPeriodicSave()
		}

		if fileStorage != nil {
			if err := fileStorage.Save(); err != nil {
				log.Printf("final save error: %v", err)
			}
		}

		log.Println("server stopped")
	}
}
