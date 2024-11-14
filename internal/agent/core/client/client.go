package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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

type metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewClient(addr string) Client {
	return &handler{addr: addr}
}

func (c *handler) SendMetric(metricName, metricType string, value any) error {

	url := fmt.Sprintf("%s/update/", c.addr)

	var metric = metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case "counter":
		v, ok := value.(int64)
		if !ok {
			return fmt.Errorf("invalid value type")
		}
		metric.Delta = &v
	case "gauge":
		v, ok := value.(float64)
		if !ok {
			return fmt.Errorf("invalid value type")
		}
		metric.Value = &v
	}

	byteData, err := compress(metric)
	if err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(byteData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

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

func compress(data any) ([]byte, error) {
	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)
	err := json.NewEncoder(gz).Encode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	err = gz.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}
