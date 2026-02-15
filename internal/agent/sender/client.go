package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

type Client struct {
	baseURL string
	client  *resty.Client
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: resty.New().
			SetTimeout(3*time.Second).
			SetHeader("Content-Type", "text/plain"),
	}
}

func (c *Client) Gauge(name string, value float64) error {
	val := strconv.FormatFloat(value, 'g', -1, 64)
	uri := fmt.Sprintf(
		"%s/update/gauge/%s/%s",
		c.baseURL,
		url.PathEscape(name),
		val,
	)
	return c.post(uri)
}

func (c *Client) Counter(name string, value int64) error {
	uri := fmt.Sprintf(
		"%s/update/counter/%s/%d",
		c.baseURL,
		url.PathEscape(name),
		value,
	)
	return c.post(uri)
}

func (c *Client) post(uri string) error {
	resp, err := c.client.R().Post(uri)
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

func (c *Client) Batch(metrics []model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	raw, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(raw); err != nil {
		return err
	}
	_ = gz.Close()

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(&buf).
		Post(c.baseURL + "/updates/")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf(
			"batch post failed: status=%d body=%s",
			resp.StatusCode(),
			resp.String(),
		)
	}

	log.Printf("[DEBUG] POST %s/updates/ -> %s", c.baseURL, resp.Status())
	return nil
}
