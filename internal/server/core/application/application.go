package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
	RestoreGauges(gauges map[string]float64)
	RestoreCounters(counters map[string]int64)
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

		return a.repo.UpdateGauge(metricName, metricValue)
	case counterType:
		metricValue, ok := value.(int64)
		if !ok {
			return fmt.Errorf("can't parse counter value, type: %T, value: %v, error: %w", value, value, ErrBadRequest)
		}

		return a.repo.UpdateCounter(metricName, metricValue)
	default:
		return fmt.Errorf("unknown metric type, value: %s, error: %w", metricTypeName, ErrBadRequest)
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
		return "", fmt.Errorf("unknown metric type, value: %s, error: %w", metricType, ErrBadRequest)
	}
}

func (a *Application) GetMetrics() []model.MetricData {
	gaugeList := a.repo.GetGaugeList()

	var metrics []model.MetricData
	for name, value := range gaugeList {
		metrics = append(metrics, model.MetricData{
			Name:  name,
			Value: strconv.FormatFloat(value, 'g', -1, 64),
		})
	}

	return metrics
}

func (a *Application) SaveMetricsToFile(filePath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	gaugeList := a.repo.GetGaugeList()
	counterList := a.repo.GetCounterList()

	metrics := metric{
		Gauges:   gaugeList,
		Counters: counterList,
	}

	bytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = os.WriteFile(filePath, bytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

type metric struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func (a *Application) LoadMetricsFromFile(filePath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to read file: %w", err)
	}

	var metrics metric
	err = json.Unmarshal(bytes, &metrics)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	a.repo.RestoreGauges(metrics.Gauges)
	a.repo.RestoreCounters(metrics.Counters)

	return nil
}
