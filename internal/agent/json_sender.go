package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

var httpClient = &http.Client{
	Timeout: 3 * time.Second,
}

func sendJSON(serverURL string, m model.Metrics) error {
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		serverURL+"/update",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}
