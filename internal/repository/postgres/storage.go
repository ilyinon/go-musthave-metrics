package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

var retryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

// Storage is a PostgreSQL implementation of repository.Storage.
type Storage struct {
	db *sql.DB
}

// New creates a new PostgreSQL storage.
func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}

// Ping checks database connectivity.
func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

/*
	Postgres retry helpers
*/

func isRetriablePGError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}
	return false
}

func retryPG(fn func() error) error {
	return retry.Do(
		fn,
		retry.Attempts(uint(len(retryDelays))),
		retry.DelayType(func(n uint, _ error, _ *retry.Config) time.Duration {
			if int(n) >= len(retryDelays) {
				return retryDelays[len(retryDelays)-1]
			}
			return retryDelays[n]
		}),
		retry.RetryIf(isRetriablePGError),
		retry.LastErrorOnly(true),
	)
}

/*
	Gauges
*/

func (s *Storage) UpdateGauge(ctx context.Context, name string, value float64) {
	const query = `
		INSERT INTO gauges (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = EXCLUDED.value
	`

	if err := retryPG(func() error {
		_, err := s.db.ExecContext(ctx, query, name, value)
		return err
	}); err != nil {
		log.Printf("postgres: update gauge failed (%s): %v", name, err)
	}
}

func (s *Storage) GetGauge(ctx context.Context, name string) (float64, bool) {
	var v float64
	err := s.db.QueryRowContext(
		ctx,
		`SELECT value FROM gauges WHERE name = $1`,
		name,
	).Scan(&v)

	if err != nil {
		return 0, false
	}
	return v, true
}

func (s *Storage) GetAllGauges(ctx context.Context) map[string]float64 {
	rows, err := s.db.QueryContext(ctx, `SELECT name, value FROM gauges`)
	if err != nil {
		log.Printf("postgres: get all gauges query failed: %v", err)
		return map[string]float64{}
	}
	defer rows.Close()

	res := make(map[string]float64)
	for rows.Next() {
		var k string
		var v float64
		if err := rows.Scan(&k, &v); err != nil {
			log.Printf("postgres: failed to scan gauge row: %v", err)
			return map[string]float64{}
		}
		res[k] = v
	}
	if err := rows.Err(); err != nil {
		log.Printf("postgres: rows error (gauges): %v", err)
		return map[string]float64{}
	}
	return res
}

func (s *Storage) ListGauges(ctx context.Context) map[string]float64 {
	return s.GetAllGauges(ctx)
}

/*
	Counters
*/

func (s *Storage) UpdateCounter(ctx context.Context, name string, delta int64) {
	const query = `
		INSERT INTO counters (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = counters.value + EXCLUDED.value
	`

	if err := retryPG(func() error {
		_, err := s.db.ExecContext(ctx, query, name, delta)
		return err
	}); err != nil {
		log.Printf("postgres: update counter failed (%s): %v", name, err)
	}
}

func (s *Storage) GetCounter(ctx context.Context, name string) (int64, bool) {
	var v int64
	err := s.db.QueryRowContext(
		ctx,
		`SELECT value FROM counters WHERE name = $1`,
		name,
	).Scan(&v)

	if err != nil {
		return 0, false
	}
	return v, true
}

func (s *Storage) GetAllCounters(ctx context.Context) map[string]int64 {
	rows, err := s.db.QueryContext(ctx, `SELECT name, value FROM counters`)
	if err != nil {
		log.Printf("postgres: get all counters query failed: %v", err)
		return map[string]int64{}
	}
	defer rows.Close()

	res := make(map[string]int64)
	for rows.Next() {
		var k string
		var v int64
		if err := rows.Scan(&k, &v); err != nil {
			log.Printf("postgres: failed to scan counter row: %v", err)
			return map[string]int64{}
		}
		res[k] = v
	}
	if err := rows.Err(); err != nil {
		log.Printf("postgres: rows error (counters): %v", err)
		return map[string]int64{}
	}
	return res
}

func (s *Storage) ListCounters(ctx context.Context) map[string]int64 {
	return s.GetAllCounters(ctx)
}
