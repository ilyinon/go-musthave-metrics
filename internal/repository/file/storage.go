package filestorage

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

type Storage struct {
	mem  repository.Storage
	path string
	mu   sync.Mutex
}

func New(mem repository.Storage, path string) *Storage {
	return &Storage{
		mem:  mem,
		path: path,
	}
}

// Save сохраняет все метрики в файл
func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var data []model.Metrics

	for k, v := range s.mem.GetAllGauges() {
		val := v
		data = append(data, model.Metrics{
			ID: k, MType: "gauge", Value: &val,
		})
	}

	for k, v := range s.mem.GetAllCounters() {
		d := v
		data = append(data, model.Metrics{
			ID: k, MType: "counter", Delta: &d,
		})
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, b, 0644)
}

// Restore загружает метрики из файла при старте
func (s *Storage) Restore() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var data []model.Metrics
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	for _, m := range data {
		switch m.MType {
		case "gauge":
			if m.Value != nil {
				s.mem.UpdateGauge(m.ID, *m.Value)
			}
		case "counter":
			if m.Delta != nil {
				s.mem.UpdateCounter(m.ID, *m.Delta)
			}
		}
	}

	return nil
}
