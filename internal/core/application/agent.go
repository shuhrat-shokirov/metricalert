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
	ResetCounters()
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
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	poll := time.NewTicker(pollInterval)
	defer poll.Stop()

	var metrics []model.Metric

	for {
		select {
		case <-poll.C:
			metrics = a.collector.CollectMetrics()
		case <-ticker.C:
			// Отправка метрик на сервер каждые reportInterval
			for _, metric := range metrics {
				err := a.client.SendMetric(metric.Name, metric.Type, metric.Value)
				if err != nil {
					log.Println("Error sending metric:", err)
				}
			}

			// Сброс счетчиков каждые reportInterval
			a.collector.ResetCounters()
		}
	}
}
