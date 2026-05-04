// Package e2e contains end-to-end tests that exercise the HTTP API as a black
// box. The test client mirrors what the iOS Ktor client will do.
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/paulsizon/notify/server/internal/api"
)

type Client struct {
	BaseURL  string
	DeviceID string
	HTTP     *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL, HTTP: &http.Client{Timeout: 120 * time.Second}}
}

func (c *Client) RegisterDevice(token string) (string, error) {
	var out api.RegisterDeviceResponse
	if err := c.do("POST", "/v1/devices", api.RegisterDeviceRequest{APNsToken: token}, &out); err != nil {
		return "", err
	}
	c.DeviceID = out.DeviceID
	return out.DeviceID, nil
}

func (c *Client) CreateSubscription(query, typ string, cadence int) (api.SubscriptionDTO, error) {
	var out api.SubscriptionDTO
	err := c.do("POST", "/v1/subscriptions",
		api.CreateSubscriptionRequest{Query: query, Type: typ, CadenceSeconds: cadence}, &out)
	return out, err
}

func (c *Client) ListSubscriptions() ([]api.SubscriptionDTO, error) {
	var out []api.SubscriptionDTO
	err := c.do("GET", "/v1/subscriptions", nil, &out)
	return out, err
}

func (c *Client) RunSubscription(id string) (api.RunResponse, error) {
	var out api.RunResponse
	err := c.do("POST", "/v1/subscriptions/"+id+"/run", nil, &out)
	return out, err
}

func (c *Client) ListSignals(subID string) ([]api.SignalDTO, error) {
	var out []api.SignalDTO
	err := c.do("GET", "/v1/subscriptions/"+subID+"/signals", nil, &out)
	return out, err
}

func (c *Client) DeleteSubscription(id string) error {
	return c.do("DELETE", "/v1/subscriptions/"+id, nil, nil)
}

func (c *Client) do(method, path string, body, out any) error {
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.DeviceID != "" {
		req.Header.Set("X-Device-Id", c.DeviceID)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s: %d %s", method, path, resp.StatusCode, string(b))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
