package collector

import "testing"

func TestRuntimeMetricsCollected(t *testing.T) {
	m := Runtime()

	if len(m) == 0 {
		t.Fatal("runtime metrics map is empty")
	}

	required := []string{
		"Alloc",
		"TotalAlloc",
		"Sys",
		"HeapAlloc",
		"HeapSys",
		"HeapIdle",
		"HeapInuse",
		"HeapReleased",
		"HeapObjects",
		"RandomValue",
	}

	for _, k := range required {
		if _, ok := m[k]; !ok {
			t.Errorf("metric %q missing", k)
		}
	}
}
