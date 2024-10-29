package application

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Repo interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
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
		return fmt.Errorf("not found, error: %w", errors.New("empty metric name"))
	}

	switch metricType(metricTypeName) {
	case gaugeType:
		return a.updateGaugeType(metricName, value)
	case counterType:
		return a.updateCounterType(metricName, value)
	default:
		return fmt.Errorf("bad request, error: %w", errors.New("unknown metric type"))
	}
}

func (a *Application) updateGaugeType(metricName, metricValue string) error {
	value, err := strconv.Atoi(metricValue)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}

	if err = a.repo.UpdateCounter(metricName, int64(value)); err != nil {
		return fmt.Errorf("can't update counter: %w", err)
	}

	return nil
}

func (a *Application) updateCounterType(metricName, metricValue string) error {
	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}

	if err = a.repo.UpdateGauge(metricName, value); err != nil {
		return fmt.Errorf("can't update gauge: %w", err)
	}

	return nil
}
