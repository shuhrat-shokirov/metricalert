package memory

import (
	"context"
	"errors"
	"sync"

	"metricalert/internal/server/core/repositories"
)

type Store struct {
	gauges    map[string]float64
	counters  map[string]int64
	gaugesM   *sync.Mutex
	countersM *sync.Mutex
}

func NewStore(config *Config) *Store {
	return &Store{
		gauges:    make(map[string]float64),
		counters:  make(map[string]int64),
		gaugesM:   &sync.Mutex{},
		countersM: &sync.Mutex{},
	}
}

func (s *Store) UpdateGauge(name string, value float64) error {
	s.gaugesM.Lock()
	defer s.gaugesM.Unlock()

	s.gauges[name] = value

	return nil
}

func (s *Store) UpdateCounter(name string, value int64) error {
	s.countersM.Lock()
	defer s.countersM.Unlock()

	s.counters[name] += value

	return nil
}

func (s *Store) GetGauge(name string) (float64, error) {
	val, ok := s.gauges[name]
	if !ok {
		return 0, repositories.ErrNotFound
	}

	return val, nil
}

func (s *Store) GetCounter(name string) (int64, error) {
	val, ok := s.counters[name]
	if !ok {
		return 0, repositories.ErrNotFound
	}
	return val, nil
}

func (s *Store) GetGaugeList() map[string]float64 {
	return s.gauges
}

func (s *Store) GetCounterList() map[string]int64 {
	return s.counters
}

func (s *Store) RestoreGauges(gauges map[string]float64) {
	s.gaugesM.Lock()
	defer s.gaugesM.Unlock()

	s.gauges = gauges
}

func (s *Store) RestoreCounters(counters map[string]int64) {
	s.countersM.Lock()
	defer s.countersM.Unlock()

	s.counters = counters
}

func (s *Store) Close() error {
	return nil
}

func (s *Store) Ping(_ context.Context) error {
	return errors.New("not implemented")
}
