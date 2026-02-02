package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

type Client struct {
	baseURL string
	client  *resty.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: resty.New().
			SetTimeout(3*time.Second).
			SetHeader("Content-Type", "application/json"),
	}
}

func (c *Client) Gauge(name string, value float64) error {
	m := model.Metrics{
		ID:    name,
		MType: "gauge",
		Value: &value,
	}
	return c.postJSON(m)
}

func (c *Client) Counter(name string, value int64) error {
	m := model.Metrics{
		ID:    name,
		MType: "counter",
		Delta: &value,
	}
	return c.postJSON(m)
}

func (c *Client) postJSON(m model.Metrics) error {
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%s/update", c.baseURL)

	resp, err := c.client.R().
		SetBody(bytes.NewReader(body)).
		Post(uri)

	if err != nil {
		return fmt.Errorf("post %s failed: %w", uri, err)
	}

	if resp.IsError() {
		return fmt.Errorf(
			"post %s failed: status=%d body=%s",
			uri,
			resp.StatusCode(),
			resp.String(),
		)
	}

	log.Printf("[DEBUG] POST %s -> %s", uri, resp.Status())
	return nil
}
