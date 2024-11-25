package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

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

func (a *Application) UpdateMetric(ctx context.Context, metricName, metricTypeName string, value any) error {
	if strings.TrimSpace(metricName) == "" {
		return fmt.Errorf("empty metric name, error: %w", ErrNotFound)
	}

	switch metricType(metricTypeName) {
	case gaugeType:
		metricValue, ok := value.(float64)
		if !ok {
			return fmt.Errorf("can't parse gauge value, type: %T, value: %v, error: %w", value, value, ErrBadRequest)
		}

		if err := a.repo.UpdateGauge(ctx, metricName, metricValue); err != nil {
			return fmt.Errorf("failed to update gauge: %w", err)
		}
	case counterType:
		metricValue, ok := value.(int64)
		if !ok {
			return fmt.Errorf("can't parse counter value, type: %T, value: %v, error: %w", value, value, ErrBadRequest)
		}

		if err := a.repo.UpdateCounter(ctx, metricName, metricValue); err != nil {
			return fmt.Errorf("failed to update counter: %w", err)
		}
	default:
		return fmt.Errorf("unknown metric type, value: %s, error: %w", metricTypeName, ErrBadRequest)
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

func (a *Application) UpdateMetrics(ctx context.Context,
	gaugeList map[string]float64, counterList map[string]int64) error {
	if len(gaugeList) > 0 {
		if err := a.repo.UpdateGauges(ctx, gaugeList); err != nil {
			return fmt.Errorf("failed to update gauges: %w", err)
		}
	}

	if len(counterList) > 0 {
		if err := a.repo.UpdateCounters(ctx, counterList); err != nil {
			return fmt.Errorf("failed to update counters: %w", err)
		}
	}

	return nil
}
