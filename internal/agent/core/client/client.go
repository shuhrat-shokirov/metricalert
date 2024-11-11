package client

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

type Client interface {
	SendMetric(metricName, metricType string, value interface{}) error
}

type handler struct {
	addr string
}

func NewClient(addr string) Client {
	return &handler{addr: addr}
}

func (c *handler) SendMetric(metricName, metricType string, value interface{}) error {

	url := fmt.Sprintf("%s/update/%s/%s/%v", c.addr, metricType, metricName, value)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metric: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send metric: status code %d", resp.StatusCode)
	}

	return nil
}
