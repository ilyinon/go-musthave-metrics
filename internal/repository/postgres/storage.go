package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

var retryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

type Storage struct {
	db *sql.DB
}

func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func isRetriablePGError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}
	return false
}

func (s *Storage) UpdateGauge(name string, value float64) {
	query := `
		INSERT INTO gauges (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = EXCLUDED.value
	`

	var err error
	for i := 0; i <= len(retryDelays); i++ {
		_, err = s.db.Exec(query, name, value)
		if err == nil {
			return
		}
		if !isRetriablePGError(err) || i == len(retryDelays) {
			return
		}
		time.Sleep(retryDelays[i])
	}
}

func (s *Storage) GetGauge(name string) (float64, bool) {
	var v float64
	err := s.db.QueryRow(`SELECT value FROM gauges WHERE name = $1`, name).Scan(&v)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (s *Storage) GetAllGauges() map[string]float64 {
	rows, err := s.db.Query(`SELECT name, value FROM gauges`)
	if err != nil {
		return map[string]float64{}
	}
	defer rows.Close()

	res := make(map[string]float64)
	for rows.Next() {
		var k string
		var v float64
		_ = rows.Scan(&k, &v)
		res[k] = v
	}
	if rows.Err() != nil {
		return map[string]float64{}
	}
	return res
}

func (s *Storage) ListGauges() map[string]float64 {
	return s.GetAllGauges()
}

func (s *Storage) UpdateCounter(name string, delta int64) {
	query := `
		INSERT INTO counters (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = counters.value + EXCLUDED.value
	`

	var err error
	for i := 0; i <= len(retryDelays); i++ {
		_, err = s.db.Exec(query, name, delta)
		if err == nil {
			return
		}
		if !isRetriablePGError(err) || i == len(retryDelays) {
			return
		}
		time.Sleep(retryDelays[i])
	}
}

func (s *Storage) GetCounter(name string) (int64, bool) {
	var v int64
	err := s.db.QueryRow(`SELECT value FROM counters WHERE name = $1`, name).Scan(&v)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (s *Storage) GetAllCounters() map[string]int64 {
	rows, err := s.db.Query(`SELECT name, value FROM counters`)
	if err != nil {
		return map[string]int64{}
	}
	defer rows.Close()

	res := make(map[string]int64)
	for rows.Next() {
		var k string
		var v int64
		_ = rows.Scan(&k, &v)
		res[k] = v
	}
	if rows.Err() != nil {
		return map[string]int64{}
	}
	return res
}

func (s *Storage) ListCounters() map[string]int64 {
	return s.GetAllCounters()
}
