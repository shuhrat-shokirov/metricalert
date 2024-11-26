//nolint:nolintlint,dupl,gocritic,goconst
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

	operation := func() error {
		_, err := s.pool.Exec(ctx, query, name, value)
		if err != nil {
			return fmt.Errorf("can't exec: %w", err)
		}

		return nil
	}

	return retry(operation)
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

	operation := func() error {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("can't exec: %w", err)
		}

		return nil
	}

	return retry(operation)
}

func (s *Store) UpdateCounter(ctx context.Context, name string, value int64) error {
	query := `
		INSERT INTO counter_metrics (name, value)
		VALUES ($1, $2)
		ON CONFLICT on constraint counter_metrics_name_key DO 
		    UPDATE SET value = counter_metrics.value + $2, updated_at = now();`

	operation := func() error {
		_, err := s.pool.Exec(ctx, query, name, value)
		if err != nil {
			return fmt.Errorf("can't exec: %w", err)
		}

		return nil
	}

	return retry(operation)
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

	operation := func() error {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("can't exec: %w", err)
		}

		return nil
	}

	return retry(operation)
}

func (s *Store) GetGauge(ctx context.Context, name string) (float64, error) {
	query := `
		SELECT value
		FROM gauge_metrics
		WHERE name = $1;`

	var value float64
	row := s.pool.QueryRow(ctx, query, name)

	operation := func() error {
		return row.Scan(&value)
	}

	return value, retry(operation)
}

func (s *Store) GetCounter(ctx context.Context, name string) (int64, error) {
	query := `
		SELECT value
		FROM counter_metrics
		WHERE name = $1;`

	var value int64
	row := s.pool.QueryRow(ctx, query, name)

	operation := func() error {
		return row.Scan(&value)
	}

	return value, retry(operation)
}

func (s *Store) Close() error {
	s.pool.Close()
	return nil
}

func (s *Store) GetGaugeList(ctx context.Context) (map[string]float64, error) {
	query := `
		SELECT name, value
		FROM gauge_metrics;`

	result := make(map[string]float64)

	operation := func() error {
		rows, err := s.pool.Query(ctx, query)
		if err != nil {
			return fmt.Errorf("can't query: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			var value float64
			if err = rows.Scan(&name, &value); err != nil {
				return fmt.Errorf("can't scan: %w", err)
			}
			result[name] = value
		}

		return nil
	}

	return result, retry(operation)
}

func (s *Store) GetCounterList(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT name, value
		FROM counter_metrics;`

	result := make(map[string]int64)

	operation := func() error {
		rows, err := s.pool.Query(ctx, query)
		if err != nil {
			return fmt.Errorf("can't query: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			var value int64
			if err = rows.Scan(&name, &value); err != nil {
				return fmt.Errorf("can't scan: %w", err)
			}
			result[name] = value
		}

		return nil
	}

	return result, retry(operation)
}

func (s *Store) Ping(ctx context.Context) error {
	err := s.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("can't ping: %w", err)
	}

	return nil
}

func isRetrievableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException, pgerrcode.ConnectionDoesNotExist, pgerrcode.ConnectionFailure:
			return true
		}
	}
	return false
}

func retry(operation func() error) error {
	const maxRetries = 3
	retryIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if err := operation(); err != nil {
			if isRetrievableError(err) {
				lastErr = err
				if i < maxRetries {
					time.Sleep(retryIntervals[i])
					continue
				}
			}

			if errors.Is(err, pgx.ErrNoRows) {
				return repositories.ErrNotFound
			}

			return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
		}
		return nil
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}
