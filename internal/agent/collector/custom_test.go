package collector

import (
	"sync/atomic"
	"testing"
)

func TestCustomMetricsCounter(t *testing.T) {
	atomic.StoreInt64(&pollCount, 0)

	first := Custom()
	second := Custom()
	third := Custom()

	if first["PollCount"] != 1 {
		t.Errorf("first = %d, want 1", first["PollCount"])
	}
	if second["PollCount"] != 2 {
		t.Errorf("second = %d, want 2", second["PollCount"])
	}
	if third["PollCount"] != 3 {
		t.Errorf("third = %d, want 3", third["PollCount"])
	}
}
