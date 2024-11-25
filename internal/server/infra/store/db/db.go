package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(dsn string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("can't create pool: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) UpdateGauge(name string, value float64) error {
	return nil
}

func (s *Store) UpdateCounter(name string, value int64) error {
	return nil
}

func (s *Store) GetGauge(name string) (float64, error) {
	return 0, nil
}

func (s *Store) GetCounter(name string) (int64, error) {
	return 0, nil
}

func (s *Store) Close() error {
	s.pool.Close()
	return nil
}

func (s *Store) GetGaugeList() map[string]float64 {
	return nil
}

func (s *Store) GetCounterList() map[string]int64 {
	return nil
}

func (s *Store) Ping(ctx context.Context) error {
	err := s.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("can't ping: %w", err)
	}

	return nil
}
