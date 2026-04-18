package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/ilyinon/go-musthave-metrics/internal/crypto"
	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

var retryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

type Client struct {
	baseURL string
	client  *resty.Client
	key     string
}

func New(baseURL string, key string) *Client {
	return &Client{
		baseURL: baseURL,
		key:     key,
		client: resty.New().
			SetTimeout(3 * time.Second),
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
	return c.postWithRetry(uri)
}

func (c *Client) Counter(name string, value int64) error {
	uri := fmt.Sprintf(
		"%s/update/counter/%s/%d",
		c.baseURL,
		url.PathEscape(name),
		value,
	)
	return c.postWithRetry(uri)
}

func (c *Client) postWithRetry(uri string) error {
	var lastErr error

	for i := 0; i <= len(retryDelays); i++ {
		resp, err := c.client.R().Post(uri)
		if err == nil && !resp.IsError() {
			log.Printf("[DEBUG] POST %s -> %s", uri, resp.Status())
			return nil
		}

		if err != nil {
			lastErr = err
			if !isRetriableHTTPError(err) || i == len(retryDelays) {
				break
			}
		} else {
			lastErr = fmt.Errorf("http status %d", resp.StatusCode())
			break
		}

		time.Sleep(retryDelays[i])
	}

	return lastErr
}

func (c *Client) Batch(metrics []model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	var lastErr error

	for i := 0; i <= len(retryDelays); i++ {
		lastErr = c.batchOnce(metrics)
		if lastErr == nil {
			return nil
		}

		if !isRetriableHTTPError(lastErr) || i == len(retryDelays) {
			break
		}

		time.Sleep(retryDelays[i])
	}

	return lastErr
}

func (c *Client) batchOnce(metrics []model.Metrics) error {
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

	req := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(&buf)

	if c.key != "" {
		hash := crypto.HashSHA256(buf.Bytes(), c.key)
		req.SetHeader("HashSHA256", hash)
	}

	resp, err := req.Post(c.baseURL + "/updates/")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("batch http status %d", resp.StatusCode())
	}

	log.Printf("[DEBUG] POST %s/updates/ -> %s", c.baseURL, resp.Status())
	return nil
}

func isRetriableHTTPError(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout()
	}

	return true
}
