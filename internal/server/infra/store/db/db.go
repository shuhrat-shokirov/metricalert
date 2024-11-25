//nolint:nolintlint,dupl,gocritic
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"metricalert/internal/server/core/repositories"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(dsn string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("can't create pool: %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("can't ping: %w", err)
	}

	if err = createTables(context.Background(), pool); err != nil {
		return nil, fmt.Errorf("can't create tables: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) UpdateGauge(ctx context.Context, name string, value float64) error {
	query := `
		INSERT INTO gauge_metrics (name, value)
		VALUES ($1, $2)
		ON CONFLICT on constraint gauge_metrics_name_key DO 
		    UPDATE SET value = $2, updated_at = now();`

	_, err := s.pool.Exec(ctx, query, name, value)
	if err != nil {
		return fmt.Errorf("can't exec query: %w", err)
	}

	return nil
}

func (s *Store) UpdateGauges(ctx context.Context, gauges map[string]float64) error {
	query := `
		INSERT INTO gauge_metrics (name, value)
		VALUES ($1, $2)
		ON CONFLICT on constraint gauge_metrics_name_key DO 
		    UPDATE SET value = $2, updated_at = now();`

	batch := &pgx.Batch{}

	for name, value := range gauges {
		batch.Queue(query, name, value)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer func() {
		if err := br.Close(); err != nil {
			fmt.Println("can't close batch: %w", err)
		}
	}()

	_, err := br.Exec()
	if err != nil {
		return fmt.Errorf("can't exec batch: %w", err)
	}

	return nil
}

func (s *Store) UpdateCounter(ctx context.Context, name string, value int64) error {
	query := `
		INSERT INTO counter_metrics (name, value)
		VALUES ($1, $2)
		ON CONFLICT on constraint counter_metrics_name_key DO 
		    UPDATE SET value = counter_metrics.value + $2, updated_at = now();`

	_, err := s.pool.Exec(ctx, query, name, value)
	if err != nil {
		return fmt.Errorf("can't exec query: %w", err)
	}

	return nil
}

func (s *Store) UpdateCounters(ctx context.Context, counters map[string]int64) error {
	query := `
		INSERT INTO counter_metrics (name, value)
		VALUES ($1, $2)
		ON CONFLICT on constraint counter_metrics_name_key DO 
		    UPDATE SET value = counter_metrics.value + $2, updated_at = now();`

	batch := &pgx.Batch{}

	for name, value := range counters {
		batch.Queue(query, name, value)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer func() {
		if err := br.Close(); err != nil {
			fmt.Println("can't close batch: %w", err)
		}
	}()

	_, err := br.Exec()
	if err != nil {
		return fmt.Errorf("can't exec batch: %w", err)
	}

	return nil
}

func (s *Store) GetGauge(ctx context.Context, name string) (float64, error) {
	query := `
		SELECT value
		FROM gauge_metrics
		WHERE name = $1;`

	var value float64
	err := s.pool.QueryRow(ctx, query, name).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, repositories.ErrNotFound
		}

		return 0, fmt.Errorf("can't query row: %w", err)
	}

	return value, nil
}

func (s *Store) GetCounter(ctx context.Context, name string) (int64, error) {
	query := `
		SELECT value
		FROM counter_metrics
		WHERE name = $1;`

	var value int64
	err := s.pool.QueryRow(ctx, query, name).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, repositories.ErrNotFound
		}

		return 0, fmt.Errorf("can't query row: %w", err)
	}

	return value, nil
}

func (s *Store) Close() error {
	s.pool.Close()
	return nil
}

func (s *Store) GetGaugeList(ctx context.Context) (map[string]float64, error) {
	query := `
		SELECT name, value
		FROM gauge_metrics;`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("can't query: %w", err)
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var name string
		var value float64
		if err = rows.Scan(&name, &value); err != nil {
			return nil, fmt.Errorf("can't scan: %w", err)
		}
		result[name] = value
	}

	return result, nil
}

func (s *Store) GetCounterList(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT name, value
		FROM counter_metrics;`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("can't query: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var name string
		var value int64
		if err = rows.Scan(&name, &value); err != nil {
			return nil, fmt.Errorf("can't scan: %w", err)
		}
		result[name] = value
	}

	return result, nil
}

func (s *Store) Ping(ctx context.Context) error {
	err := s.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("can't ping: %w", err)
	}

	return nil
}
