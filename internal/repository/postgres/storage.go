package postgres

import (
	"context"
	"database/sql"
)

type Storage struct {
	db *sql.DB
}

func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Storage) UpdateGauge(name string, value float64) {
	_, _ = s.db.Exec(`
		INSERT INTO gauges (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = EXCLUDED.value
	`, name, value)
}

func (s *Storage) GetGauge(name string) (float64, bool) {
	var v float64
	err := s.db.QueryRow(
		`SELECT value FROM gauges WHERE name = $1`, name,
	).Scan(&v)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (s *Storage) ListGauges() map[string]float64 {
	return map[string]float64{}
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
	return res
}

func (s *Storage) UpdateCounter(name string, delta int64) {
	_, _ = s.db.Exec(`
		INSERT INTO counters (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = counters.value + EXCLUDED.value
	`, name, delta)
}

func (s *Storage) GetCounter(name string) (int64, bool) {
	var v int64
	err := s.db.QueryRow(
		`SELECT value FROM counters WHERE name = $1`, name,
	).Scan(&v)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (s *Storage) ListCounters() map[string]int64 {
	return map[string]int64{}
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
	return res
}
