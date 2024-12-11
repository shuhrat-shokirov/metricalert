package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"metricalert/internal/server/core/model"
	"metricalert/internal/server/core/repositories"
)

type Repo interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateGauges(ctx context.Context, gauges map[string]float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	UpdateCounters(ctx context.Context, counters map[string]int64) error
	GetGaugeList(ctx context.Context) (map[string]float64, error)
	GetCounterList(ctx context.Context) (map[string]int64, error)
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	Close() error
	Ping(ctx context.Context) error
}

type Application struct {
	repo Repo
}

func NewApplication(repo Repo) *Application {
	return &Application{
		repo: repo,
	}
}

type metricType string

const (
	gaugeType   metricType = "gauge"
	counterType metricType = "counter"
)

func (a *Application) UpdateMetric(ctx context.Context, metric model.MetricRequest) error {
	if strings.TrimSpace(metric.ID) == "" {
		return fmt.Errorf("empty metric name, error: %w", ErrNotFound)
	}

	switch metricType(metric.MType) {
	case counterType:
		if metric.Delta == nil {
			return fmt.Errorf("delta is nil on counter metric, error: %w", ErrBadRequest)
		}

		if err := a.repo.UpdateCounter(ctx, metric.ID, *metric.Delta); err != nil {
			return fmt.Errorf("failed to update counter: %w", err)
		}
	case gaugeType:
		if metric.Value == nil {
			return fmt.Errorf("value is nil on gauge metric, error: %w", ErrBadRequest)
		}

		if err := a.repo.UpdateGauge(ctx, metric.ID, *metric.Value); err != nil {
			return fmt.Errorf("failed to update gauge: %w", err)
		}
	default:
		return fmt.Errorf("unknown metric type, value: %s, error: %w", metric.MType, ErrBadRequest)
	}

	return nil
}

var (
	ErrBadRequest = errors.New("bad request")
	ErrNotFound   = errors.New("not found")
)

func (a *Application) GetMetric(ctx context.Context, metricName, metricType string) (string, error) {
	switch metricType {
	case "gauge":
		gauge, err := a.repo.GetGauge(ctx, metricName)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return "", fmt.Errorf("metric not found: %w", ErrNotFound)
			}
			return "", fmt.Errorf("failed to get gauge: %w", err)
		}

		return strconv.FormatFloat(gauge, 'g', -1, 64), nil
	case "counter":
		counter, err := a.repo.GetCounter(ctx, metricName)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return "", fmt.Errorf("metric not found: %w", ErrNotFound)
			}
			return "", fmt.Errorf("failed to get counter: %w", err)
		}

		return strconv.Itoa(int(counter)), nil
	default:
		return "", fmt.Errorf("unknown metric type, value: %s, error: %w", metricType, ErrBadRequest)
	}
}

func (a *Application) GetMetrics(ctx context.Context) ([]model.MetricData, error) {
	gaugeList, err := a.repo.GetGaugeList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gauge list: %w", err)
	}

	var metrics = make([]model.MetricData, 0, len(gaugeList))
	for name, value := range gaugeList {
		metrics = append(metrics, model.MetricData{
			Name:  name,
			Value: strconv.FormatFloat(value, 'g', -1, 64),
		})
	}

	return metrics, nil
}

func (a *Application) Ping(ctx context.Context) error {
	err := a.repo.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}

	return nil
}

func (a *Application) UpdateMetrics(ctx context.Context, metrics []model.MetricRequest) error {
	var (
		gaugeMetricList   = map[string]float64{}
		counterMetricList = map[string]int64{}
	)

	for _, r := range metrics {
		switch metricType(r.MType) {
		case counterType:
			if r.Delta == nil {
				zap.L().Warn("delta is nil on counter metric", zap.String("id", r.ID))
				continue
			}

			counterMetricList[r.ID] += *r.Delta
		case gaugeType:
			if r.Value == nil {
				zap.L().Warn("value is nil on gauge metric", zap.String("id", r.ID))
				continue
			}

			gaugeMetricList[r.ID] = *r.Value

		default:
			zap.L().Warn("unknown metric type", zap.String("type", r.MType))
			continue
		}
	}

	if len(gaugeMetricList) > 0 {
		if err := a.repo.UpdateGauges(ctx, gaugeMetricList); err != nil {
			return fmt.Errorf("failed to update gauges: %w", err)
		}
	}

	if len(counterMetricList) > 0 {
		if err := a.repo.UpdateCounters(ctx, counterMetricList); err != nil {
			return fmt.Errorf("failed to update counters: %w", err)
		}
	}

	return nil
}
