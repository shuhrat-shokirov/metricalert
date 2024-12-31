package services

import (
	"math/rand"
	"runtime"
	"sync/atomic"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

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

	const metricCount = 29

	// Собираем метрики
	metrics := make([]model.Metric, 0, metricCount)

	metrics = append(metrics,
		model.Metric{Name: "Alloc", Value: float64(memStats.Alloc), Type: "gauge"},
		model.Metric{Name: "BuckHashSys", Value: float64(memStats.BuckHashSys), Type: "gauge"},
		model.Metric{Name: "Frees", Value: float64(memStats.Frees), Type: "gauge"},
		model.Metric{Name: "GCCPUFraction", Value: memStats.GCCPUFraction, Type: "gauge"},
		model.Metric{Name: "HeapAlloc", Value: float64(memStats.HeapAlloc), Type: "gauge"},
		model.Metric{Name: "HeapIdle", Value: float64(memStats.HeapIdle), Type: "gauge"},
		model.Metric{Name: "HeapInuse", Value: float64(memStats.HeapInuse), Type: "gauge"},
		model.Metric{Name: "HeapObjects", Value: float64(memStats.HeapObjects), Type: "gauge"},
		model.Metric{Name: "HeapReleased", Value: float64(memStats.HeapReleased), Type: "gauge"},
		model.Metric{Name: "HeapSys", Value: float64(memStats.HeapSys), Type: "gauge"},
		model.Metric{Name: "LastGC", Value: float64(memStats.LastGC), Type: "gauge"},
		model.Metric{Name: "Lookups", Value: float64(memStats.Lookups), Type: "gauge"},
		model.Metric{Name: "Mallocs", Value: float64(memStats.Mallocs), Type: "gauge"},
		model.Metric{Name: "MCacheInuse", Value: float64(memStats.MCacheInuse), Type: "gauge"},
		model.Metric{Name: "MCacheSys", Value: float64(memStats.MCacheSys), Type: "gauge"},
		model.Metric{Name: "MSpanInuse", Value: float64(memStats.MSpanInuse), Type: "gauge"},
		model.Metric{Name: "MSpanSys", Value: float64(memStats.MSpanSys), Type: "gauge"},
		model.Metric{Name: "NextGC", Value: float64(memStats.NextGC), Type: "gauge"},
		model.Metric{Name: "NumGC", Value: float64(memStats.NumGC), Type: "gauge"},
		model.Metric{Name: "PauseTotalNs", Value: float64(memStats.PauseTotalNs), Type: "gauge"},
		model.Metric{Name: "StackInuse", Value: float64(memStats.StackInuse), Type: "gauge"},
		model.Metric{Name: "StackSys", Value: float64(memStats.StackSys), Type: "gauge"},
		model.Metric{Name: "Sys", Value: float64(memStats.Sys), Type: "gauge"},
		model.Metric{Name: "TotalAlloc", Value: float64(memStats.TotalAlloc), Type: "gauge"},
		model.Metric{Name: "PollCount", Value: atomic.LoadInt64(&c.pollCount), Type: "counter"},
		model.Metric{Name: "RandomValue", Value: rand.Float64(), Type: "gauge"},
		model.Metric{Name: "GCSys", Value: float64(memStats.GCSys), Type: "gauge"},
		model.Metric{Name: "NumForcedGC", Value: float64(memStats.NumForcedGC), Type: "gauge"},
		model.Metric{Name: "OtherSys", Value: float64(memStats.OtherSys), Type: "gauge"})

	return metrics
}

func (c *Collector) CollectMemoryMetrics() []model.Metric {
	const metricCount = 3

	metrics := make([]model.Metric, 0, metricCount)

	vmem, err := mem.VirtualMemory()
	if err == nil {
		metrics = append(metrics,
			model.Metric{Name: "TotalMemory", Value: float64(vmem.Total), Type: "gauge"},
			model.Metric{Name: "FreeMemory", Value: float64(vmem.Used), Type: "gauge"},
		)
	}

	cpuPercent, err := cpu.Percent(0, false)
	if err == nil {
		metrics = append(metrics,
			model.Metric{Name: "CPUutilization1", Value: cpuPercent[0], Type: "gauge"},
		)
	}

	return metrics
}

func (c *Collector) ResetCounters() {
	atomic.StoreInt64(&c.pollCount, 0)
}
