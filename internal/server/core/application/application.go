package application

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"metricalert/internal/server/core/model"
	"metricalert/internal/server/core/repositories"
)

type Repo interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGaugeList() map[string]float64
	GetCounterList() map[string]int64
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	Close() error
}

type Application struct {
	repo Repo
	mu   *sync.Mutex
}

func NewApplication(repo Repo) *Application {
	return &Application{
		repo: repo,
		mu:   &sync.Mutex{},
	}
}

type metricType string

const (
	gaugeType   metricType = "gauge"
	counterType metricType = "counter"
)

func (a *Application) UpdateMetric(metricName, metricTypeName string, value any) error {
	if strings.TrimSpace(metricName) == "" {
		return fmt.Errorf("empty metric name, error: %w", ErrNotFound)
	}

	switch metricType(metricTypeName) {
	case gaugeType:
		metricValue, ok := value.(float64)
		if !ok {
			return fmt.Errorf("can't parse gauge value, type: %T, value: %v, error: %w", value, value, ErrBadRequest)
		}

		if err := a.repo.UpdateGauge(metricName, metricValue); err != nil {
			return fmt.Errorf("failed to update gauge: %w", err)
		}
	case counterType:
		metricValue, ok := value.(int64)
		if !ok {
			return fmt.Errorf("can't parse counter value, type: %T, value: %v, error: %w", value, value, ErrBadRequest)
		}

		if err := a.repo.UpdateCounter(metricName, metricValue); err != nil {
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

func (a *Application) GetMetric(metricName, metricType string) (string, error) {
	switch metricType {
	case "gauge":
		gauge, err := a.repo.GetGauge(metricName)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return "", fmt.Errorf("metric not found: %w", ErrNotFound)
			}
			return "", fmt.Errorf("failed to get gauge: %w", err)
		}

		return strconv.FormatFloat(gauge, 'g', -1, 64), nil
	case "counter":
		counter, err := a.repo.GetCounter(metricName)
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

func (a *Application) GetMetrics() []model.MetricData {
	gaugeList := a.repo.GetGaugeList()

	var metrics = make([]model.MetricData, 0, len(gaugeList))
	for name, value := range gaugeList {
		metrics = append(metrics, model.MetricData{
			Name:  name,
			Value: strconv.FormatFloat(value, 'g', -1, 64),
		})
	}

	return metrics
}
