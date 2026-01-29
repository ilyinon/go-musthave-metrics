package mem

import "testing"

func TestStorage_UpdateAndGetGauge(t *testing.T) {
	store := New()

	store.UpdateGauge("cpu", 1.23)
	val, ok := store.GetGauge("cpu")
	if !ok || val != 1.23 {
		t.Fatalf("expected 1.23, got %v", val)
	}
}

func TestStorage_UpdateAndGetCounter(t *testing.T) {
	store := New()

	store.UpdateCounter("requests", 5)
	store.UpdateCounter("requests", 3)

	val, _ := store.GetCounter("requests")
	if val != 8 {
		t.Fatalf("expected 8, got %d", val)
	}
}

func TestStorage_List(t *testing.T) {
	store := New()
	store.UpdateGauge("cpu", 1.1)
	store.UpdateCounter("req", 10)

	if len(store.ListGauges()) != 1 {
		t.Fatal("expected 1 gauge")
	}
	if len(store.ListCounters()) != 1 {
		t.Fatal("expected 1 counter")
	}
}
