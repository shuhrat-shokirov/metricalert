package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

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

	var (
		client = &http.Client{Timeout: timeout}
		resp   *http.Response
	)

	err = retry(func() error {
		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send metric: %w", err)
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				zap.L().Error("can't close response body", zap.Error(err))
			}
		}()

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to send metric: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send metric: status code %d", resp.StatusCode)
	}

	return nil
}

func retry(operation func() error) error {
	const maxRetries = 3
	retryIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		err := operation()
		if err == nil {
			return nil
		}

		if !isRetrievableError(err) {
			return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
		}

		lastErr = err
		if i < maxRetries {
			time.Sleep(retryIntervals[i])
			continue
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

func isRetrievableError(err error) bool {
	if errors.Is(err, http.ErrHandlerTimeout) {
		return true
	}

	if errors.Is(err, http.ErrServerClosed) {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if netErr.Op == "dial" {
			return true
		}
	}

	var httpErr net.Error
	if errors.As(err, &httpErr) && httpErr.Timeout() {
		return true
	}

	return false
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
