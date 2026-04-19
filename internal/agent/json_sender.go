package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

var httpClient = &http.Client{
	Timeout: 3 * time.Second,
}

// sendJSON sends a single metric to the server in JSON format using gzip compression.
func sendJSON(serverURL string, m model.Metrics) error {
	raw, err := json.Marshal(m)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err = gz.Write(raw)
	if err := gz.Close(); err != nil {
		return err
	}
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		serverURL+"/update",
		&buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}
