package models

import (
	"testing"
)

func TestMetrics_IsGauge(t *testing.T) {
	m := Metrics{MType: Gauge}
	if !m.IsGauge() {
		t.Fatal("expected IsGauge to return true")
	}
}
