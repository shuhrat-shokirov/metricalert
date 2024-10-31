package store

import (
	"errors"

	"metricalert/internal/infra/store/memory"
)

type Store interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGaugeList() map[string]string
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
}

func NewStore(conf Config) (Store, error) {
	switch {
	case conf.Memory != nil:
		return memory.NewMemStorage(conf.Memory), nil
	default:
		return nil, errors.New("unknown store type")
	}
}
