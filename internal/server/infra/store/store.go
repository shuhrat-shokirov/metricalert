package store

import (
	"fmt"

	"metricalert/internal/server/infra/store/memory"
)

type Store interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGaugeList() map[string]string
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
}

func NewStore(conf Config) (Store, error) {
	switch {
	case conf.Memory != nil:
		return memory.NewMemStorage(conf.Memory), nil
	default:
		return nil, fmt.Errorf("unknown store type")
	}
}
