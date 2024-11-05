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
		return a.updateGaugeType(metricName, value)
	case counterType:
		return a.updateCounterType(metricName, value)
	default:
		return fmt.Errorf("unknown metric type, error: %w", ErrBadRequest)
	}
}

func (a *Application) updateGaugeType(metricName, metricValue string) error {

	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		return fmt.Errorf("can't parse value: %w", errors.Join(err, ErrBadRequest))
	}

	if err = a.repo.UpdateGauge(metricName, value); err != nil {
		return fmt.Errorf("can't update counter: %w", err)
	}

	return nil
}

func (a *Application) updateCounterType(metricName, metricValue string) error {

	value, err := strconv.Atoi(metricValue)
	if err != nil {
		return fmt.Errorf("can't parse value: %w", errors.Join(err, ErrBadRequest))
	}

	if err = a.repo.UpdateCounter(metricName, int64(value)); err != nil {
		return fmt.Errorf("can't update gauge: %w", err)
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
			return "", err
		}

		return fmt.Sprintf("%g", gauge), nil
	case "counter":
		counter, err := a.repo.GetCounter(metricName)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return "", fmt.Errorf("metric not found: %w", ErrNotFound)
			}
			return "", err
		}

		return fmt.Sprintf("%d", counter), nil
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
