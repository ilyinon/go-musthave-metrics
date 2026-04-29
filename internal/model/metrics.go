package model

const (
	MetricGauge   = "gauge"
	MetricCounter = "counter"
)

type RuntimeMetrics map[string]float64
type CustomMetrics map[string]int64

// Metrics represents a metric with its type and value.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"` // gauge | counter
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
