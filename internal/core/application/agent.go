package application

import (
	"log"
	"time"

	"metricalert/internal/core/model"
)

type Client interface {
	SendMetric(name string, metricType string, value interface{}) error
}

type Collector interface {
	CollectMetrics() []model.Metric
}

type Agent struct {
	client    Client
	collector Collector
}

func NewAgent(client Client, collector Collector) *Agent {
	return &Agent{
		client:    client,
		collector: collector,
	}
}

func (a *Agent) Start(pollInterval, reportInterval time.Duration) {
	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	var metrics []model.Metric

	for {
		select {
		case <-pollTicker.C:
			metrics = a.collector.CollectMetrics()

		case <-reportTicker.C:
			for _, metric := range metrics {
				err := a.client.SendMetric(metric.Name, metric.Type, metric.Value)
				if err != nil {
					log.Println("Error sending metric:", err)
				}
			}
		}
	}
}
