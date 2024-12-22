package application

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"metricalert/internal/server/core/model"
)

type Client interface {
	SendMetrics(metrics []model.Metric) error
}

type Collector interface {
	CollectMetrics() []model.Metric
	CollectMemoryMetrics() []model.Metric
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

type Config struct {
	PoolInterval   time.Duration
	ReportInterval time.Duration
	RateLimit      int64
}

func (a *Agent) Start(conf Config) {
	ticker := time.NewTicker(conf.ReportInterval)
	defer ticker.Stop()

	poll := time.NewTicker(conf.PoolInterval)
	defer poll.Stop()

	metricsChan := make(chan []model.Metric, conf.RateLimit)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	for i := 0; i < int(conf.RateLimit); i++ {
		wg.Add(1)
		go a.worker(ctx, &wg, metricsChan)
	}

	go func(collector Collector, ctx context.Context, a chan<- []model.Metric) {
		for {
			select {
			case <-poll.C:
				metrics := collector.CollectMemoryMetrics()
				a <- metrics
			case <-ticker.C:
				metrics := collector.CollectMemoryMetrics()
				a <- metrics
			case <-ctx.Done():
				return
			}
		}
	}(a.collector, ctx, metricsChan)

	for {
		select {
		case <-poll.C:
			metrics := a.collector.CollectMetrics()
			metricsChan <- metrics
		case <-ticker.C:
			metrics := a.collector.CollectMetrics()
			metricsChan <- metrics

			// Сброс счетчиков каждые reportInterval
			a.collector.ResetCounters()
		case <-ctx.Done():
			close(metricsChan)
			wg.Wait()
			return
		}
	}
}

func (a *Agent) worker(ctx context.Context, wg *sync.WaitGroup, metricsChan <-chan []model.Metric) {
	defer wg.Done()

	for {
		select {
		case metrics, ok := <-metricsChan:
			if !ok {
				return
			}

			if err := a.client.SendMetrics(metrics); err != nil {
				zap.L().Error("can't send metrics", zap.Error(err))
				continue
			}
		case <-ctx.Done():
			return
		}
	}
}
