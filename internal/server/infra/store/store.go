package store

import (
	"fmt"

	"metricalert/internal/server/infra/store/memory"
)

type Store interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGaugeList() map[string]float64
	GetCounterList() map[string]int64
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)

	RestoreGauges(gauges map[string]float64)
	RestoreCounters(counters map[string]int64)
}

func NewStore(conf Config) (Store, error) {
	switch {
	case conf.Memory != nil:
		return memory.NewMemStorage(conf.Memory), nil
	default:
		return nil, fmt.Errorf("unknown store type, config: %+v", conf)
	}
}
