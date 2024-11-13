package services

import (
	"math/rand"
	"runtime"
	"sync/atomic"

	"metricalert/internal/server/core/model"
)

type Collector struct {
	pollCount int64
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) CollectMetrics() []model.Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Увеличиваем PollCount на 1 при каждом опросе метрик
	atomic.AddInt64(&c.pollCount, 1)

	// Собираем метрики
	metrics := make([]model.Metric, 0, 24)

	metrics = append(metrics, model.Metric{Name: "Alloc", Value: float64(memStats.Alloc), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "BuckHashSys", Value: float64(memStats.BuckHashSys), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "Frees", Value: float64(memStats.Frees), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "GCCPUFraction", Value: memStats.GCCPUFraction, Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "HeapAlloc", Value: float64(memStats.HeapAlloc), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "HeapIdle", Value: float64(memStats.HeapIdle), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "HeapInuse", Value: float64(memStats.HeapInuse), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "HeapObjects", Value: float64(memStats.HeapObjects), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "HeapReleased", Value: float64(memStats.HeapReleased), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "HeapSys", Value: float64(memStats.HeapSys), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "LastGC", Value: float64(memStats.LastGC), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "Lookups", Value: float64(memStats.Lookups), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "Mallocs", Value: float64(memStats.Mallocs), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "MCacheInuse", Value: float64(memStats.MCacheInuse), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "MCacheSys", Value: float64(memStats.MCacheSys), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "MSpanInuse", Value: float64(memStats.MSpanInuse), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "MSpanSys", Value: float64(memStats.MSpanSys), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "NextGC", Value: float64(memStats.NextGC), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "NumGC", Value: float64(memStats.NumGC), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "PauseTotalNs", Value: float64(memStats.PauseTotalNs), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "StackInuse", Value: float64(memStats.StackInuse), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "StackSys", Value: float64(memStats.StackSys), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "Sys", Value: float64(memStats.Sys), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "TotalAlloc", Value: float64(memStats.TotalAlloc), Type: "gauge"})
	metrics = append(metrics, model.Metric{Name: "PollCount", Value: atomic.LoadInt64(&c.pollCount), Type: "counter"})
	metrics = append(metrics, model.Metric{Name: "RandomValue", Value: rand.Float64() * 100, Type: "gauge"})

	return metrics
}

// ResetCounters сбрасывает счетчики
func (c *Collector) ResetCounters() {
	atomic.StoreInt64(&c.pollCount, 0)
}
