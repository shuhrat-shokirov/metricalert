package memory

import (
	"fmt"
	"sync"

	"metricalert/internal/server/core/repositories"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

func NewMemStorage(config *Config) *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) UpdateGauge(name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = value

	return nil
}

func (s *MemStorage) UpdateCounter(name string, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name] += value

	return nil
}

func (s *MemStorage) GetGauge(name string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.gauges[name]
	if !ok {
		return 0, repositories.ErrNotFound
	}

	return val, nil
}

func (s *MemStorage) GetCounter(name string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.counters[name]
	if !ok {
		return 0, repositories.ErrNotFound
	}
	return val, nil
}

func (s *MemStorage) GetGaugeList() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range s.gauges {
		result[k] = fmt.Sprintf("%f", v)
	}

	return result
}
