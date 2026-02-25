package mem

import (
	"context"
	"testing"
)

func TestStorage_UpdateAndGetGauge(t *testing.T) {
	store := New()
	ctx := context.Background()

	store.UpdateGauge(ctx, "cpu", 1.23)
	val, ok := store.GetGauge(ctx, "cpu")
	if !ok || val != 1.23 {
		t.Fatalf("expected 1.23, got %v", val)
	}
}

func TestStorage_UpdateAndGetCounter(t *testing.T) {
	store := New()
	ctx := context.Background()

	store.UpdateCounter(ctx, "requests", 5)
	store.UpdateCounter(ctx, "requests", 3)

	val, _ := store.GetCounter(ctx, "requests")
	if val != 8 {
		t.Fatalf("expected 8, got %d", val)
	}
}

func TestStorage_List(t *testing.T) {
	store := New()
	ctx := context.Background()

	store.UpdateGauge(ctx, "cpu", 1.1)
	store.UpdateCounter(ctx, "req", 10)

	if len(store.ListGauges(ctx)) != 1 {
		t.Fatal("expected 1 gauge")
	}
	if len(store.ListCounters(ctx)) != 1 {
		t.Fatal("expected 1 counter")
	}
}
