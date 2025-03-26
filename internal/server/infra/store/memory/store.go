// Package memory Реализация хранилища метрик в памяти.
//
// Содержит реализацию интерфейса Store из core/repositories.
//
// Для хранения метрик используются два словаря: gauges и counters.
// Для обеспечения потокобезопасности используются мьютексы.
package memory

import (
	"context"
	"sync"

	"metricalert/internal/server/core/repositories"
)

// Store структура хранит метрики в памяти.
type Store struct {
	gauges    map[string]float64
	counters  map[string]int64
	gaugesM   *sync.Mutex
	countersM *sync.Mutex
}

// NewStore создает новый экземпляр Store.
func NewStore(config *Config) *Store {
	return &Store{
		gauges:    make(map[string]float64),
		counters:  make(map[string]int64),
		gaugesM:   &sync.Mutex{},
		countersM: &sync.Mutex{},
	}
}

// UpdateGauge обновляет значение метрики типа gauge.
func (s *Store) UpdateGauge(_ context.Context, name string, value float64) error {
	s.gaugesM.Lock()
	defer s.gaugesM.Unlock()

	s.gauges[name] = value

	return nil
}

// UpdateGauges обновляет значения метрик типа gauge.
func (s *Store) UpdateGauges(_ context.Context, gauges map[string]float64) error {
	s.gaugesM.Lock()
	defer s.gaugesM.Unlock()

	for name, value := range gauges {
		s.gauges[name] = value
	}

	return nil
}

// UpdateCounter обновляет значение метрики типа counter.
func (s *Store) UpdateCounter(_ context.Context, name string, value int64) error {
	s.countersM.Lock()
	defer s.countersM.Unlock()

	s.counters[name] += value

	return nil
}

// UpdateCounters обновляет значения метрик типа counter.
func (s *Store) UpdateCounters(_ context.Context, counters map[string]int64) error {
	s.countersM.Lock()
	defer s.countersM.Unlock()

	for name, value := range counters {
		s.counters[name] += value
	}

	return nil
}

// GetGauge возвращает значение метрики типа gauge.
func (s *Store) GetGauge(_ context.Context, name string) (float64, error) {
	val, ok := s.gauges[name]
	if !ok {
		return 0, repositories.ErrNotFound
	}

	return val, nil
}

// GetCounter возвращает значение метрики типа counter.
func (s *Store) GetCounter(_ context.Context, name string) (int64, error) {
	val, ok := s.counters[name]
	if !ok {
		return 0, repositories.ErrNotFound
	}
	return val, nil
}

// GetGaugeList возвращает список метрик типа gauge.
func (s *Store) GetGaugeList(_ context.Context) (map[string]float64, error) {
	return s.gauges, nil
}

// GetCounterList возвращает список метрик типа counter.
func (s *Store) GetCounterList(_ context.Context) (map[string]int64, error) {
	return s.counters, nil
}

// RestoreGauges восстанавливает значения метрик типа gauge.
func (s *Store) RestoreGauges(gauges map[string]float64) {
	s.gaugesM.Lock()
	defer s.gaugesM.Unlock()

	s.gauges = gauges
}

// RestoreCounters восстанавливает значения метрик типа counter.
func (s *Store) RestoreCounters(counters map[string]int64) {
	s.countersM.Lock()
	defer s.countersM.Unlock()

	s.counters = counters
}

// Close закрывает хранилище.
func (s *Store) Close() error {
	return nil
}

// Ping проверяет доступность хранилища.
func (s *Store) Ping(_ context.Context) error {
	return nil
}

func (s *Store) Sync(_ context.Context) {
}
