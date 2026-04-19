package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/ilyinon/go-musthave-metrics/internal/repository/mem"
)

func ExampleUpdateJSONHandler() {
	store := mem.New()
	h := NewUpdateJSON(store, nil)

	val := 42.0
	body, _ := json.Marshal(model.Metrics{
		ID:    "cpu",
		MType: model.MetricGauge,
		Value: &val,
	})

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	// Output: 200
}
