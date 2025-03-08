//nolint:wrapcheck,nolintlint,gocritic,errcheck,dupl,forcetypeassert
package rest

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/core/model"
)

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

func TestServerAPI_GetMetrics(t *testing.T) {
	t.Run("internal error", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server:  mockServerService,
			logger:  *logger,
			hashKey: "test-hash-key",
		}
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		mockServerService.On("GetMetrics", mock.Anything).
			Return([]model.MetricData{}, errors.New("store error"))

		h.metrics(c)

		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
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

		mockServerService.On("GetMetrics", mock.Anything).
			Return([]model.MetricData{{Name: "test"}}, nil)

		h.metrics(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})
}

func TestServerAPI_ResponseGzipMiddleware(t *testing.T) {
	t.Run("success with no gzip", func(t *testing.T) {
		h := handler{
			hashKey: "test-hash-key",
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Writer.Header().Set("Content-Type", "application/json")
		_, _ = c.Writer.Write([]byte(`{"type":"gauge"}`))

		h.responseGzipMiddleware()

		assert.Equal(t, "application/json", c.Writer.Header().Get("Content-Type"))
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, `{"type":"gauge"}`, recorder.Body.String())
	})

	t.Run("success with gzip", func(t *testing.T) {
		h := handler{
			hashKey: "test-hash-key",
		}

		// Создаём тестовый HTTP-запрос с Accept-Encoding: gzip
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.Nil(t, err)

		req.Header.Set("Accept-Encoding", "gzip")

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = req
		c.Writer.Header().Set("Content-Type", "application/json")
		_, err = c.Writer.Write([]byte(`{"type":"gauge"}`))
		require.Nil(t, err)

		// Применяем middleware
		middleware := h.responseGzipMiddleware()
		middleware(c)

		// Проверяем заголовки
		assert.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))
		assert.Equal(t, "gzip", c.Writer.Header().Get("Accept-Encoding"))
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestServerAPI_MwDecompress(t *testing.T) {
	t.Run("without content type", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		h := handler{
			hashKey: "test-hash-key",
		}

		c.Request = &http.Request{
			Header: http.Header{},
			Body:   io.NopCloser(bytes.NewBufferString(`{}`)),
		}

		c.Request.Header.Set("Application-Type", "application/json")

		middleware := h.mwDecompress()
		middleware(c)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("parsing error with content type", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		h := handler{
			hashKey: "test-hash-key",
			logger:  *zap.NewNop().Sugar(),
		}

		c.Request = &http.Request{
			Header: http.Header{
				"Content-Encoding": []string{"gzip"},
			},
			Body: io.NopCloser(bytes.NewBufferString(``)),
		}

		middleware := h.mwDecompress()
		middleware(c)

		require.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	})
}

func TestServerAPI_MwLog(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		h := handler{
			hashKey: "test-hash-key",
			logger:  *zap.NewNop().Sugar(),
		}

		c.Request = &http.Request{
			Header: http.Header{
				"Content-Encoding": []string{"gzip"},
			},
		}

		h.mwLog()(c)

		c.Writer.WriteHeader(http.StatusOK)
	})
}

func Test_gzipResponseWriter_Write(t *testing.T) {
	t.Run("success json writer", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Header("Content-Type", "application/json")

		w := &gzipResponseWriter{
			Writer:         c.Writer,
			ResponseWriter: c.Writer,
		}

		_, err := w.Write([]byte(`{"type":"gauge"}`))
		require.Nil(t, err)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, `{"type":"gauge"}`, recorder.Body.String())
	})

	t.Run("success gzip writer", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Header("Content-Encoding", "gzip")

		w := &gzipResponseWriter{
			Writer:         c.Writer,
			ResponseWriter: c.Writer,
		}

		_, err := w.Write([]byte(`{"type":"gauge"}`))
		require.Nil(t, err)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, `{"type":"gauge"}`, recorder.Body.String())
		assert.Equal(t, `gzip`, recorder.Header().Get("Content-Encoding"))
	})
}

func TestServerAPI_Ping(t *testing.T) {
	t.Run("internal error", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server: mockServerService,
			logger: *logger,
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)

		mockServerService.On("Ping", mock.Anything).
			Return(errors.New("store error"))

		h.dbPing(c)

		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	})

	t.Run("success", func(t *testing.T) {
		mockServerService := new(MockServerService)
		logger := zap.NewNop().Sugar()

		h := handler{
			server: mockServerService,
			logger: *logger,
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)

		mockServerService.On("Ping", mock.Anything).
			Return(nil)

		h.dbPing(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())

		mockServerService.AssertExpectations(t)
	})
}

func TestServerAPI_Running(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := NewServerAPI(Config{
			Server:  new(MockServerService),
			Logger:  *zap.NewNop().Sugar(),
			HashKey: "test-hash-key",
			Port:    8080,
		})

		go func() {
			_ = api.Run()
		}()

		time.Sleep(50 * time.Millisecond)
		err := api.Run()
		assert.Error(t, err)

		err = api.srv.Shutdown(context.Background())

		assert.NoError(t, err)
	})
}
