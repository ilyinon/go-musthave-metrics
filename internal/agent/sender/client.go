package sender

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
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

// Client sends metrics to the server over HTTP.
type Client struct {
	baseURL   string
	client    *resty.Client
	key       string
	publicKey *rsa.PublicKey
	realIP    string
}

// New creates a new Client with the given base URL and optional signing key.
func New(baseURL string, key string) *Client {
	return &Client{
		baseURL: baseURL,
		key:     key,
		client:  resty.New().SetTimeout(3 * time.Second),
		realIP:  localIPForURL(baseURL),
	}
}

// SetPublicKey sets the RSA public key for encryption.
func (c *Client) SetPublicKey(pub *rsa.PublicKey) {
	c.publicKey = pub
}

// SetRealIP overrides the IP sent in the X-Real-IP header.
func (c *Client) SetRealIP(ip string) {
	c.realIP = ip
}

// Gauge sends a gauge metric to the server.
func (c *Client) Gauge(name string, value float64) error {
	if c.publicKey != nil {
		return c.metricJSONWithRetry(model.Metrics{
			ID:    name,
			MType: model.MetricGauge,
			Value: &value,
		})
	}

	val := strconv.FormatFloat(value, 'g', -1, 64)
	uri := fmt.Sprintf("%s/update/gauge/%s/%s", c.baseURL, url.PathEscape(name), val)
	return c.postWithRetry(uri)
}

// Counter sends a counter metric to the server.
func (c *Client) Counter(name string, value int64) error {
	if c.publicKey != nil {
		return c.metricJSONWithRetry(model.Metrics{
			ID:    name,
			MType: model.MetricCounter,
			Delta: &value,
		})
	}

	uri := fmt.Sprintf("%s/update/counter/%s/%d", c.baseURL, url.PathEscape(name), value)
	return c.postWithRetry(uri)
}

func (c *Client) postWithRetry(uri string) error {
	var lastErr error

	for i := 0; i <= len(retryDelays); i++ {
		raw := []byte{}

		req := c.newRequest()

		if c.publicKey != nil {
			var err error
			raw, err = crypto.EncryptHybridRSA(c.publicKey, raw)
			if err != nil {
				return err
			}
			req.SetBody(raw)
		}

		resp, err := req.Post(uri)
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

// Batch sends multiple metrics in a single request.
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
	_, err = gz.Write(raw)
	if err != nil {
		return err
	}
	_ = gz.Close()

	resp, err := c.postPayload(c.baseURL+"/updates/", buf.Bytes(), "application/json", "gzip")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("batch http status %d", resp.StatusCode())
	}

	log.Printf("[DEBUG] POST %s/updates/ -> %s", c.baseURL, resp.Status())
	return nil
}

func (c *Client) metricJSONWithRetry(metric model.Metrics) error {
	var lastErr error

	for i := 0; i <= len(retryDelays); i++ {
		lastErr = c.metricJSONOnce(metric)
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

func (c *Client) metricJSONOnce(metric model.Metrics) error {
	payload, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	resp, err := c.postPayload(c.baseURL+"/update", payload, "application/json", "")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("metric http status %d", resp.StatusCode())
	}

	log.Printf("[DEBUG] POST %s/update -> %s", c.baseURL, resp.Status())
	return nil
}

func (c *Client) postPayload(uri string, payload []byte, contentType string, contentEncoding string) (*resty.Response, error) {
	body := payload

	req := c.newRequest()
	if contentType != "" {
		req.SetHeader("Content-Type", contentType)
	}
	if contentEncoding != "" {
		req.SetHeader("Content-Encoding", contentEncoding)
	}

	if c.key != "" {
		hash := crypto.HashSHA256(payload, c.key)
		req.SetHeader("HashSHA256", hash)
	}

	if c.publicKey != nil {
		encrypted, err := crypto.EncryptHybridRSA(c.publicKey, payload)
		if err != nil {
			return nil, err
		}
		body = encrypted
	}

	return req.SetBody(body).Post(uri)
}

func (c *Client) newRequest() *resty.Request {
	req := c.client.R()
	if c.realIP != "" {
		req.SetHeader("X-Real-IP", c.realIP)
	}
	return req
}

func isRetriableHTTPError(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout()
	}

	return true
}

func localIPForURL(rawURL string) string {
	if host, port := targetHostPort(rawURL); host != "" {
		if ip := outboundIP(net.JoinHostPort(host, port)); ip != "" {
			return ip
		}
	}

	if ip := firstInterfaceIP(); ip != "" {
		return ip
	}

	return "127.0.0.1"
}

func targetHostPort(rawAddress string) (string, string) {
	if u, err := url.Parse(rawAddress); err == nil && u.Hostname() != "" {
		port := u.Port()
		if port == "" {
			port = "80"
		}
		return u.Hostname(), port
	}

	host, port, err := net.SplitHostPort(rawAddress)
	if err == nil {
		if host == "" {
			host = "localhost"
		}
		return host, port
	}

	if rawAddress != "" {
		return rawAddress, "80"
	}

	return "", ""
}

func outboundIP(address string) string {
	conn, err := net.DialTimeout("udp", address, time.Second)
	if err != nil {
		return ""
	}
	defer conn.Close()

	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || addr.IP == nil {
		return ""
	}

	return addr.IP.String()
}

func firstInterfaceIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				return ip4.String()
			}
			return ip.String()
		}
	}

	return ""
}
