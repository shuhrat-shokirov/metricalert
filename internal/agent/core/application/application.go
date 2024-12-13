package application

import (
	"time"

	"go.uber.org/zap"

	"metricalert/internal/server/core/model"
)

type Client interface {
	SendMetrics(metrics []model.Metric) error
}

type Collector interface {
	CollectMetrics() []model.Metric
	ResetCounters()
}

type Agent struct {
	client    Client
	collector Collector
}

func NewApplication(client Client, collector Collector) *Agent {
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
			err := a.client.SendMetrics(metrics)
			if err != nil {
				zap.L().Error("can't send metrics", zap.Error(err))
				continue
			}

			// Сброс счетчиков каждые reportInterval
			a.collector.ResetCounters()
		}
	}
}
