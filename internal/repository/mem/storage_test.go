package mem

import (
	"context"
	"testing"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
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

func TestStorage_UpdateBatchRejectsAtomically(t *testing.T) {
	store := New()
	ctx := context.Background()

	value := 12.5
	err := store.UpdateBatch(ctx, []model.Metrics{
		{ID: "Load", MType: model.MetricGauge, Value: &value},
		{ID: "Broken", MType: model.MetricGauge},
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if _, ok := store.GetGauge(ctx, "Load"); ok {
		t.Fatal("valid metric from rejected batch was stored")
	}
}
