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
	client        Client
	collector     Collector
	memoryMetrics []model.Metric
	memoryMutex   sync.Mutex
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

	for range conf.RateLimit {
		wg.Add(1)
		go a.worker(ctx, &wg, metricsChan)
	}

	go func(collector Collector, ctx context.Context) {
		for {
			select {
			case <-poll.C:
				memoryMetrics := collector.CollectMemoryMetrics()

				a.memoryMutex.Lock()
				a.memoryMetrics = memoryMetrics
				a.memoryMutex.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}(a.collector, ctx)

	var metrics []model.Metric
	for {
		select {
		case <-poll.C:
			metrics = a.collector.CollectMetrics()
		case <-ticker.C:
			a.memoryMutex.Lock()
			if len(a.memoryMetrics) > 0 {
				metrics = append(metrics, a.memoryMetrics...)
			}
			a.memoryMutex.Unlock()

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
