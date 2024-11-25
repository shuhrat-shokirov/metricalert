package store

import (
	"context"
	"fmt"

	"metricalert/internal/server/infra/store/db"
	"metricalert/internal/server/infra/store/file"
	"metricalert/internal/server/infra/store/memory"
)

type Store interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateGauges(ctx context.Context, gauges map[string]float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	UpdateCounters(ctx context.Context, counters map[string]int64) error
	GetGaugeList(context.Context) (map[string]float64, error)
	GetCounterList(context.Context) (map[string]int64, error)
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	Close() error
	Ping(ctx context.Context) error
}

func NewStore(conf Config) (Store, error) {
	switch {
	case conf.DB != nil:
		store, err := db.New(conf.DB.DSN)
		if err != nil {
			return nil, fmt.Errorf("can't create db store: %w", err)
		}

		return store, nil
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
