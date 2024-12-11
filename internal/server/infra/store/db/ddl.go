package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	gaugeTable := `
    CREATE TABLE IF NOT EXISTS gauge_metrics (
        id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        name TEXT NOT NULL,
        value DOUBLE PRECISION NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
        CONSTRAINT gauge_metrics_name_key UNIQUE (name)
    );`

	counterTable := `
    CREATE TABLE IF NOT EXISTS counter_metrics (
        id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        name TEXT NOT NULL,
        value BIGINT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
        CONSTRAINT counter_metrics_name_key UNIQUE (name)
    );`

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("err starting transaction: %w", err)
	}

	if _, err := tx.Exec(ctx, gaugeTable); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("err creating gauge_metrics table: %w", err)
	}

	if _, err := tx.Exec(ctx, counterTable); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("err creating counter_metrics table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("err committing transaction: %w", err)
	}

	return nil
}
