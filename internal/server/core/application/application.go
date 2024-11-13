package application

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"metricalert/internal/server/core/model"
	"metricalert/internal/server/core/repositories"
)

type Repo interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGaugeList() map[string]string
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
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

func (a *Application) UpdateMetric(metricName, metricTypeName, value string) error {
	if strings.TrimSpace(metricName) == "" {
		return fmt.Errorf("empty metric name, error: %w", ErrNotFound)
	}

	switch metricType(metricTypeName) {
	case gaugeType:
		metricValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("can't parse gauge value: %w", errors.Join(err, ErrBadRequest))
		}

		return a.repo.UpdateGauge(metricName, metricValue)
	case counterType:
		metricValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("can't parse counter value: %w", errors.Join(err, ErrBadRequest))
		}

		return a.repo.UpdateCounter(metricName, int64(metricValue))
	default:
		return fmt.Errorf("unknown metric type, error: %w", ErrBadRequest)
	}
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
			return "", err
		}

		return strconv.FormatFloat(gauge, 'g', -1, 64), nil
	case "counter":
		counter, err := a.repo.GetCounter(metricName)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return "", fmt.Errorf("metric not found: %w", ErrNotFound)
			}
			return "", err
		}

		return strconv.Itoa(int(counter)), nil
	default:
		return "", fmt.Errorf("unknown metric type: %w", ErrBadRequest)
	}
}

func (a *Application) GetMetrics() []model.MetricData {
	gaugeList := a.repo.GetGaugeList()

	var metrics []model.MetricData
	for name, value := range gaugeList {
		metrics = append(metrics, model.MetricData{
			Name:  name,
			Value: value,
		})
	}

	return metrics
}
