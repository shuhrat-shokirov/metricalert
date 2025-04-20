//nolint:wrapcheck,nolintlint,gocritic,errcheck,dupl,forcetypeassert
package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"metricalert/internal/server/core/model"
	"metricalert/internal/server/core/repositories"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) UpdateGauge(ctx context.Context, name string, value float64) error {
	args := m.Called(ctx, name, value)
	return args.Error(0)
}

func (m *mockRepo) UpdateGauges(ctx context.Context, gauges map[string]float64) error {
	args := m.Called(ctx, gauges)
	return args.Error(0)
}

func (m *mockRepo) UpdateCounter(ctx context.Context, name string, value int64) error {
	args := m.Called(ctx, name, value)
	return args.Error(0)
}

func (m *mockRepo) UpdateCounters(ctx context.Context, counters map[string]int64) error {
	args := m.Called(ctx, counters)
	return args.Error(0)
}

func (m *mockRepo) GetGaugeList(ctx context.Context) (map[string]float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *mockRepo) GetCounterList(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *mockRepo) GetGauge(ctx context.Context, name string) (float64, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(float64), args.Error(1)
}

func (m *mockRepo) GetCounter(ctx context.Context, name string) (int64, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRepo) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockRepo) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockRepo) Sync(ctx context.Context) {
	args := m.Called(ctx)
	_ = args.Error(0)
}

func TestApplication_NewApplication(t *testing.T) {
	repo := new(mockRepo)
	app := NewApplication(repo)

	assert.NotNil(t, app)
	assert.Equal(t, repo, app.repo)
}

func TestApplication_UpdateMetric(t *testing.T) {
	t.Run("id is empty", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		err := app.UpdateMetric(context.Background(), model.MetricRequest{})
		assert.Error(t, err)
		assert.Equal(t, "empty metric name, error: not found", err.Error())
	})

	t.Run("unknown metric type", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		err := app.UpdateMetric(context.Background(), model.MetricRequest{ID: "test", MType: "unknown"})
		assert.Error(t, err)
		assert.Equal(t, "unknown metric type, value: unknown, error: bad request", err.Error())
	})

	t.Run("counter metric, delta is nil", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		err := app.UpdateMetric(context.Background(), model.MetricRequest{ID: "test", MType: "counter"})
		assert.Error(t, err)
		assert.Equal(t, "delta is nil on counter metric, error: bad request", err.Error())
	})

	t.Run("gauge metric, value is nil", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		err := app.UpdateMetric(context.Background(), model.MetricRequest{ID: "test", MType: "gauge"})
		assert.Error(t, err)
		assert.Equal(t, "value is nil on gauge metric, error: bad request", err.Error())
	})

	t.Run("update counter", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("UpdateCounter", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := app.UpdateMetric(context.Background(), model.MetricRequest{ID: "test", MType: "counter", Delta: new(int64)})
		assert.NoError(t, err)
	})

	t.Run("update counter, with err", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("UpdateCounter", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		err := app.UpdateMetric(context.Background(), model.MetricRequest{ID: "test", MType: "counter", Delta: new(int64)})
		assert.Errorf(t, err, "failed to update counter: %v", assert.AnError)
	})

	t.Run("update gauge", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("UpdateGauge", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := app.UpdateMetric(context.Background(), model.MetricRequest{ID: "test", MType: "gauge", Value: new(float64)})
		assert.NoError(t, err)
	})

	t.Run("update gauge, with err", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("UpdateGauge", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		err := app.UpdateMetric(context.Background(), model.MetricRequest{ID: "test", MType: "gauge", Value: new(float64)})
		assert.Errorf(t, err, "failed to update gauge: %v", assert.AnError)
	})
}

func TestApplication_GetMetric(t *testing.T) {
	t.Run("unknown metric type", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		_, err := app.GetMetric(context.Background(), "test", "unknown")
		assert.Error(t, err)
		assert.Equal(t, "unknown metric type, value: unknown, error: bad request", err.Error())
	})

	t.Run("get counter", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("GetCounter", mock.Anything, mock.Anything).Return(int64(0), nil)

		value, err := app.GetMetric(context.Background(), "test", "counter")
		assert.NoError(t, err)
		assert.Equal(t, "0", value)
	})

	t.Run("get counter, with err", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("GetCounter", mock.Anything, mock.Anything).Return(int64(0), assert.AnError)

		_, err := app.GetMetric(context.Background(), "test", "counter")
		assert.Error(t, err)
	})

	t.Run("get counter, with err not found", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("GetCounter", mock.Anything, mock.Anything).Return(int64(0), repositories.ErrNotFound)

		_, err := app.GetMetric(context.Background(), "test", "counter")
		require.Error(t, err)
		assert.Equal(t, "metric not found: not found", err.Error())
	})

	t.Run("get gauge", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("GetGauge", mock.Anything, mock.Anything).Return(1.2, nil)

		value, err := app.GetMetric(context.Background(), "test", "gauge")
		assert.NoError(t, err)
		assert.Equal(t, "1.2", value)
	})

	t.Run("get gauge, with err", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("GetGauge", mock.Anything, mock.Anything).Return(0.0, assert.AnError)

		_, err := app.GetMetric(context.Background(), "test", "gauge")
		assert.Error(t, err)
	})

	t.Run("get gauge, with err not found", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("GetGauge", mock.Anything, mock.Anything).Return(0.0, repositories.ErrNotFound)

		_, err := app.GetMetric(context.Background(), "test", "gauge")
		require.Error(t, err)
		assert.Equal(t, "metric not found: not found", err.Error())
	})
}

func TestApplication_GetMetrics(t *testing.T) {
	t.Run("with err", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("GetGaugeList", mock.Anything).Return(map[string]float64{}, assert.AnError)

		_, err := app.GetMetrics(context.Background())
		assert.Error(t, err)
	})

	t.Run("without err", func(t *testing.T) {
		repo := new(mockRepo)
		app := NewApplication(repo)

		repo.On("GetGaugeList", mock.Anything).Return(map[string]float64{"test": 1.2}, nil)

		metrics, err := app.GetMetrics(context.Background())
		assert.NoError(t, err)
		assert.Len(t, metrics, 1)
		assert.Equal(t, "test", metrics[0].Name)
		assert.Equal(t, "1.2", metrics[0].Value)
	})
}
