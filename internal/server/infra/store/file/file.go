package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"metricalert/internal/server/infra/store/memory"
)

type Store struct {
	*memory.Store
	file   *os.File
	mu     *sync.Mutex
	ticker *time.Ticker
}

func NewStore(conf *Config) (*Store, error) {
	const perm = 0o666
	file, err := os.OpenFile(conf.FilePath, os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		return nil, fmt.Errorf("can't open file: %w", err)
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("can't read file: %w", err)
	}

	s := &Store{
		Store:  memory.NewStore(conf.MemoryStore),
		file:   file,
		mu:     &sync.Mutex{},
		ticker: time.NewTicker(time.Duration(conf.StoreInterval) * time.Second),
	}

	go func() {
		for range s.ticker.C {
			if err := s.saveToFile(context.Background()); err != nil {
				log.Printf("can't save to file: %v", err)
			}
		}
	}()

	if len(bytes) == 0 {
		return s, nil
	}

	var metrics metric
	err = json.Unmarshal(bytes, &metrics)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal data: %w", err)
	}

	s.RestoreGauges(metrics.Gauges)
	s.RestoreCounters(metrics.Counters)

	return s, nil
}

type metric struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func (s *Store) UpdateGauge(ctx context.Context, name string, value float64) error {
	err := s.Store.UpdateGauge(ctx, name, value)
	if err != nil {
		return fmt.Errorf("can't update gauge: %w", err)
	}

	return nil
}

func (s *Store) UpdateGauges(ctx context.Context, gauges map[string]float64) error {
	err := s.Store.UpdateGauges(ctx, gauges)
	if err != nil {
		return fmt.Errorf("can't update gauges: %w", err)
	}

	return nil
}

func (s *Store) saveToFile(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	gaugeList, err := s.GetGaugeList(ctx)
	if err != nil {
		return fmt.Errorf("can't get gauge list: %w", err)
	}

	counterList, err := s.GetCounterList(ctx)
	if err != nil {
		return fmt.Errorf("can't get counter list: %w", err)
	}

	metrics := metric{
		Gauges:   gaugeList,
		Counters: counterList,
	}

	bytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("can't marshal data: %w", err)
	}

	_, err = s.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("can't seek file: %w", err)
	}

	_, err = s.file.Write(bytes)
	if err != nil {
		return fmt.Errorf("can't write to file: %w", err)
	}

	return nil
}

func (s *Store) UpdateCounter(ctx context.Context, name string, value int64) error {
	err := s.Store.UpdateCounter(ctx, name, value)
	if err != nil {
		return fmt.Errorf("can't update counter: %w", err)
	}

	return nil
}

func (s *Store) UpdateCounters(ctx context.Context, counters map[string]int64) error {
	err := s.Store.UpdateCounters(ctx, counters)
	if err != nil {
		return fmt.Errorf("can't update counters: %w", err)
	}

	return nil
}

func (s *Store) GetGauge(ctx context.Context, name string) (float64, error) {
	value, err := s.Store.GetGauge(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("can't get gauge: %w", err)
	}

	return value, nil
}

func (s *Store) GetCounter(ctx context.Context, name string) (int64, error) {
	value, err := s.Store.GetCounter(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("can't get counter: %w", err)
	}

	return value, nil
}

func (s *Store) Close() error {
	s.ticker.Stop()

	err := s.saveToFile(context.Background())
	if err != nil {
		return fmt.Errorf("can't save to file: %w", err)
	}

	err = s.file.Close()
	if err != nil {
		return fmt.Errorf("can't close file: %w", err)
	}

	return nil
}

func (s *Store) GetGaugeList(ctx context.Context) (map[string]float64, error) {
	gaugeList, err := s.Store.GetGaugeList(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't get gauge list: %w", err)
	}

	return gaugeList, nil
}

func (s *Store) GetCounterList(ctx context.Context) (map[string]int64, error) {
	counterList, err := s.Store.GetCounterList(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't get counter list: %w", err)
	}

	return counterList, nil
}

func (s *Store) Ping(_ context.Context) error {
	return errors.New("not implemented")
}
