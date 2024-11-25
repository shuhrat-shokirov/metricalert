package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"metricalert/internal/server/core/model"
)

type Client interface {
	SendMetrics(metrics []model.Metric) error
}

type handler struct {
	addr string
}

type metrics struct {
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
}

func NewClient(addr string) Client {
	return &handler{addr: addr}
}

func (c *handler) SendMetrics(list []model.Metric) error {
	url := fmt.Sprintf("http://%s/updates/", c.addr)

	request := make([]metrics, 0, len(list))
	for _, metric := range list {
		var m = metrics{
			ID:    metric.Name,
			MType: metric.Type,
		}

		switch metric.Type {
		case "counter":
			v, ok := metric.Value.(int64)
			if !ok {
				return fmt.Errorf("invalid counter value type, type: %T, value: %v", metric.Value, metric.Value)
			}
			m.Delta = &v
		case "gauge":
			v, ok := metric.Value.(float64)
			if !ok {
				return fmt.Errorf("invalid gauge value type, type: %T, value: %v", metric.Value, metric.Value)
			}
			m.Value = &v
		}

		request = append(request, m)
	}

	byteData, err := compress(request)
	if err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(byteData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	const timeout = 5 * time.Second

	client := &http.Client{Timeout: timeout}

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
