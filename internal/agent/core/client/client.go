package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"metricalert/internal/server/core/model"
)

type Client interface {
	SendMetrics(metrics []model.Metric) error
}

type handler struct {
	addr      string
	hashKey   string
	publicKey *rsa.PublicKey
}

type metrics struct {
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
}

func NewClient(addr, hashKey, cryptoKey string) Client {
	h := &handler{
		addr:    addr,
		hashKey: hashKey,
	}

	if cryptoKey != "" {
		pubKey, err := loadPublicKey(cryptoKey)
		if err != nil {
			zap.L().Fatal("can't load public key", zap.Error(err))
		}

		h.publicKey = pubKey
	}

	return h
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("ошибка декодирования PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("ключ не является RSA публичным ключом")
	}

	return publicKey, nil
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

	if c.publicKey != nil {
		encrypted, err := c.rsaEncrypt(byteData)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}

		byteData = encrypted
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(byteData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if c.hashKey != "" {
		req.Header.Set("HashSHA256", hashRequest(byteData, c.hashKey))
	}

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
			newErr := resp.Body.Close()
			if newErr != nil {
				zap.L().Error("can't close response body", zap.Error(newErr))
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

// получаем данные и шифрируем их по публику
func (c *handler) rsaEncrypt(data []byte) ([]byte, error) {
	newData := make([]byte, base64.StdEncoding.EncodedLen(len(data)))

	base64.StdEncoding.Encode(newData, data)

	encryptPKCS1v15, err := rsa.EncryptPKCS1v15(rand.Reader, c.publicKey, newData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return encryptPKCS1v15, nil
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

func hashRequest(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	dst := h.Sum(nil)
	return hex.EncodeToString(dst)
}
