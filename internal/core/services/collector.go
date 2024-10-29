package services

import (
	"math/rand"
	"runtime"
	"sync/atomic"

	"metricalert/internal/core/model"
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

	return []model.Metric{
		{"Alloc", float64(memStats.Alloc), "gauge"},
		{"BuckHashSys", float64(memStats.BuckHashSys), "gauge"},
		{"Frees", float64(memStats.Frees), "gauge"},
		{"GCCPUFraction", float64(memStats.GCCPUFraction), "gauge"},
		{"HeapAlloc", float64(memStats.HeapAlloc), "gauge"},
		{"HeapIdle", float64(memStats.HeapIdle), "gauge"},
		{"HeapInuse", float64(memStats.HeapInuse), "gauge"},
		{"HeapObjects", float64(memStats.HeapObjects), "gauge"},
		{"HeapReleased", float64(memStats.HeapReleased), "gauge"},
		{"HeapSys", float64(memStats.HeapSys), "gauge"},
		{"LastGC", float64(memStats.LastGC), "gauge"},
		{"Lookups", float64(memStats.Lookups), "gauge"},
		{"Mallocs", float64(memStats.Mallocs), "gauge"},
		{"MCacheInuse", float64(memStats.MCacheInuse), "gauge"},
		{"MCacheSys", float64(memStats.MCacheSys), "gauge"},
		{"MSpanInuse", float64(memStats.MSpanInuse), "gauge"},
		{"MSpanSys", float64(memStats.MSpanSys), "gauge"},
		{"NextGC", float64(memStats.NextGC), "gauge"},
		{"NumGC", float64(memStats.NumGC), "gauge"},
		{"PauseTotalNs", float64(memStats.PauseTotalNs), "gauge"},
		{"StackInuse", float64(memStats.StackInuse), "gauge"},
		{"StackSys", float64(memStats.StackSys), "gauge"},
		{"Sys", float64(memStats.Sys), "gauge"},
		{"TotalAlloc", float64(memStats.TotalAlloc), "gauge"},
		{"PollCount", float64(atomic.LoadInt64(&c.pollCount)), "counter"}, // PollCount
		{"RandomValue", rand.Float64() * 100, "gauge"},                    // RandomValue
	}
}
