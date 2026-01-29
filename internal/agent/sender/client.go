package sender

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
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
