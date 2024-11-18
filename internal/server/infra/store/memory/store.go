package memory

import (
	"sync"

	"metricalert/internal/server/core/repositories"
)

type MemStorage struct {
	gauges    map[string]float64
	counters  map[string]int64
	gaugesM   *sync.Mutex
	countersM *sync.Mutex
}

func NewMemStorage(config *Config) *MemStorage {
	return &MemStorage{
		gauges:    make(map[string]float64),
		counters:  make(map[string]int64),
		gaugesM:   &sync.Mutex{},
		countersM: &sync.Mutex{},
	}
}

func (s *MemStorage) UpdateGauge(name string, value float64) error {
	s.gaugesM.Lock()
	defer s.gaugesM.Unlock()

	s.gauges[name] = value

	return nil
}

func (s *MemStorage) UpdateCounter(name string, value int64) error {
	s.countersM.Lock()
	defer s.countersM.Unlock()

	s.counters[name] += value

	return nil
}

func (s *MemStorage) GetGauge(name string) (float64, error) {
	val, ok := s.gauges[name]
	if !ok {
		return 0, repositories.ErrNotFound
	}

	return val, nil
}

func (s *MemStorage) GetCounter(name string) (int64, error) {
	val, ok := s.counters[name]
	if !ok {
		return 0, repositories.ErrNotFound
	}
	return val, nil
}

func (s *MemStorage) GetGaugeList() map[string]float64 {
	return s.gauges
}

func (s *MemStorage) GetCounterList() map[string]int64 {
	return s.counters
}

func (s *MemStorage) RestoreGauges(gauges map[string]float64) {
	s.gaugesM.Lock()
	defer s.gaugesM.Unlock()

	s.gauges = gauges
}

func (s *MemStorage) RestoreCounters(counters map[string]int64) {
	s.countersM.Lock()
	defer s.countersM.Unlock()

	s.counters = counters
}
