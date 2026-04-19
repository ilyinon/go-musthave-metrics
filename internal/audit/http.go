package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPSink sends audit events to a remote HTTP endpoint.ы
type HTTPSink struct {
	url    string
	client *http.Client
}

// NewHTTPSink creates a new HTTPSink.
func NewHTTPSink(url string) *HTTPSink {
	return &HTTPSink{
		url: url,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (h *HTTPSink) Send(e Event) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, h.url, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	return nil
}
