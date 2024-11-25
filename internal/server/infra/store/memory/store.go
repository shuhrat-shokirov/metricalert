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

func (s *Store) UpdateGauge(_ context.Context, name string, value float64) error {
	s.gaugesM.Lock()
	defer s.gaugesM.Unlock()

	s.gauges[name] = value

	return nil
}

func (s *Store) UpdateGauges(_ context.Context, gauges map[string]float64) error {
	s.gaugesM.Lock()
	defer s.gaugesM.Unlock()

	for name, value := range gauges {
		s.gauges[name] = value
	}

	return nil
}

func (s *Store) UpdateCounter(_ context.Context, name string, value int64) error {
	s.countersM.Lock()
	defer s.countersM.Unlock()

	s.counters[name] += value

	return nil
}

func (s *Store) UpdateCounters(_ context.Context, counters map[string]int64) error {
	s.countersM.Lock()
	defer s.countersM.Unlock()

	for name, value := range counters {
		s.counters[name] += value
	}

	return nil
}

func (s *Store) GetGauge(_ context.Context, name string) (float64, error) {
	val, ok := s.gauges[name]
	if !ok {
		return 0, repositories.ErrNotFound
	}

	return val, nil
}

func (s *Store) GetCounter(_ context.Context, name string) (int64, error) {
	val, ok := s.counters[name]
	if !ok {
		return 0, repositories.ErrNotFound
	}
	return val, nil
}

func (s *Store) GetGaugeList(_ context.Context) (map[string]float64, error) {
	return s.gauges, nil
}

func (s *Store) GetCounterList(_ context.Context) (map[string]int64, error) {
	return s.counters, nil
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
