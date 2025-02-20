package rest

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/core/model"
)

// MockServerService is a mock implementation of the ServerService interface
type MockServerService struct {
	mock.Mock
}

func (m *MockServerService) UpdateMetric(ctx context.Context, request model.MetricRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockServerService) UpdateMetrics(ctx context.Context, request []model.MetricRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockServerService) GetMetric(ctx context.Context, metricName, metricType string) (string, error) {
	args := m.Called(ctx, metricName, metricType)
	return args.String(0), args.Error(1)
}

func (m *MockServerService) GetMetrics(ctx context.Context) ([]model.MetricData, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.MetricData), args.Error(1)
}

func (m *MockServerService) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewServerAPI(t *testing.T) {
	mockServerService := new(MockServerService)
	logger := zap.NewNop().Sugar()

	conf := Config{
		Server:  mockServerService,
		Logger:  *logger,
		HashKey: "test-hash-key",
		Port:    8080,
	}

	api := NewServerAPI(conf)

	assert.NotNil(t, api)
	assert.NotNil(t, api.srv)
	assert.Equal(t, ":8080", api.srv.Addr)
}

func TestServerAPI_UpdateMetric(t *testing.T) {
	t.Run("metric type empty", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		h.update(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("metric type invalid", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		c.Params = append(c.Params, gin.Param{Key: "type", Value: "invalid"})

		h.update(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("metric type valid, metric value invalid", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		c.Params = append(c.Params, gin.Param{Key: "type", Value: "gauge"})

		h.update(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})

	t.Run("err from store", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		c.Params = append(c.Params, gin.Param{Key: "type", Value: "gauge"})
		c.Params = append(c.Params, gin.Param{Key: "name", Value: "test"})
		c.Params = append(c.Params, gin.Param{Key: "value", Value: "1"})

		mockServerService.On("UpdateMetric", mock.Anything, mock.Anything).
			Return(errors.New("store error"))

		h.update(c)

		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})

	t.Run("bad request", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		c.Params = append(c.Params, gin.Param{Key: "type", Value: "gauge"})
		c.Params = append(c.Params, gin.Param{Key: "name", Value: "test"})
		c.Params = append(c.Params, gin.Param{Key: "value", Value: "1"})

		mockServerService.On("UpdateMetric", mock.Anything, mock.Anything).
			Return(application.ErrBadRequest)

		h.update(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		c.Params = append(c.Params, gin.Param{Key: "type", Value: "gauge"})
		c.Params = append(c.Params, gin.Param{Key: "name", Value: "test"})
		c.Params = append(c.Params, gin.Param{Key: "value", Value: "1"})

		mockServerService.On("UpdateMetric", mock.Anything, mock.Anything).
			Return(application.ErrNotFound)

		h.update(c)

		assert.Equal(t, http.StatusNotFound, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})

	t.Run("gauge success", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		c.Params = append(c.Params, gin.Param{Key: "type", Value: "gauge"})
		c.Params = append(c.Params, gin.Param{Key: "name", Value: "test"})
		c.Params = append(c.Params, gin.Param{Key: "value", Value: "1.1"})

		mockServerService.On("UpdateMetric", mock.Anything, mock.Anything).
			Return(nil)

		h.update(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})

	t.Run("counter success", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		c.Params = append(c.Params, gin.Param{Key: "type", Value: "counter"})
		c.Params = append(c.Params, gin.Param{Key: "name", Value: "test"})
		c.Params = append(c.Params, gin.Param{Key: "value", Value: "1"})

		mockServerService.On("UpdateMetric", mock.Anything, mock.Anything).
			Return(nil)

		h.update(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})

	t.Run("counter filed type", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		c.Params = append(c.Params, gin.Param{Key: "type", Value: "counter"})
		c.Params = append(c.Params, gin.Param{Key: "name", Value: "test"})
		c.Params = append(c.Params, gin.Param{Key: "value", Value: "1.1"})

		h.update(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})
}

func TestServerAPI_UpdateWithBody(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)
		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{}`)),
		}

		mockServerService.On("UpdateMetric", mock.Anything, mock.Anything).
			Return(errors.New("store error"))

		h.updateWithBody(c)

		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	})

	t.Run("bad request", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)
		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{}`)),
		}

		mockServerService.On("UpdateMetric", mock.Anything, mock.Anything).
			Return(application.ErrBadRequest)

		h.updateWithBody(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("bad request", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{"id": 1}`)),
		}

		h.updateWithBody(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("err not found", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)
		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{}`)),
		}

		mockServerService.On("UpdateMetric", mock.Anything, mock.Anything).
			Return(application.ErrNotFound)

		h.updateWithBody(c)

		assert.Equal(t, http.StatusNotFound, c.Writer.Status())
	})

	t.Run("success", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)
		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{}`)),
		}

		mockServerService.On("UpdateMetric", mock.Anything, mock.Anything).
			Return(nil)

		h.updateWithBody(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})
}

func TestServerAPI_GetMetric(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		c.Params = append(c.Params, gin.Param{Key: "name", Value: "test"})

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("", errors.New("store error"))

		h.get(c)

		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	})

	t.Run("not found", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("", application.ErrNotFound)

		h.get(c)

		assert.Equal(t, http.StatusNotFound, c.Writer.Status())
	})

	t.Run("bad request", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(nil)

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("", application.ErrBadRequest)

		h.get(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("success", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Params = append(c.Params, gin.Param{Key: "name", Value: "test"})

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("test", nil)

		h.get(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "test", recorder.Body.String())

		mockServerService.AssertExpectations(t)
	})
}

func TestServerAPI_GetMetricValue(t *testing.T) {
	t.Run("internal error", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{}`)),
		}

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("", errors.New("store error"))

		h.getMetricValue(c)

		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	})

	t.Run("Request is incorrect", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		h.getMetricValue(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("not found", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{}`)),
		}

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("", application.ErrNotFound)

		h.getMetricValue(c)

		assert.Equal(t, http.StatusNotFound, c.Writer.Status())
	})

	t.Run("bad request", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{}`)),
		}

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("", application.ErrBadRequest)

		h.getMetricValue(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("success gauge", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{"type":"gauge"}`)),
		}

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("10", nil)

		h.getMetricValue(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, `{"value":10,"id":"","type":"gauge"}`, recorder.Body.String())

		mockServerService.AssertExpectations(t)
	})

	t.Run("incorrect gauge value", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{"type":"gauge"}`)),
		}

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("str", nil)

		h.getMetricValue(c)

		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})

	t.Run("incorrect counter value", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{"type":"counter"}`)),
		}

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("str", nil)

		h.getMetricValue(c)

		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})

	t.Run("success counter", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{"type":"counter"}`)),
		}

		mockServerService.On("GetMetric", mock.Anything, mock.Anything, mock.Anything).
			Return("10", nil)

		h.getMetricValue(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, `{"delta":10,"id":"","type":"counter"}`, recorder.Body.String())

		mockServerService.AssertExpectations(t)
	})
}
