package store

import (
	"fmt"

	"metricalert/internal/server/infra/store/file"
	"metricalert/internal/server/infra/store/memory"
)

type Store interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGaugeList() map[string]float64
	GetCounterList() map[string]int64
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	Close() error
}

func NewStore(conf Config) (Store, error) {
	switch {
	case conf.File != nil:
		if !conf.File.Restore {
			return memory.NewStore(conf.File.MemoryStore), nil
		}

		store, err := file.NewStore(conf.File)
		if err != nil {
			return nil, fmt.Errorf("can't create file store: %w", err)
		}

		return store, nil
	default:
		return nil, fmt.Errorf("unknown store type, config: %+v", conf)
	}
}
