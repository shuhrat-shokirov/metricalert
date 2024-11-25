package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	gaugeTable := `
    CREATE TABLE IF NOT EXISTS gauge_metrics (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        value DOUBLE PRECISION NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
        CONSTRAINT gauge_metrics_name_key UNIQUE (name)
    );`

	counterTable := `
    CREATE TABLE IF NOT EXISTS counter_metrics (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        value BIGINT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
        CONSTRAINT counter_metrics_name_key UNIQUE (name)
    );`

	if _, err := pool.Exec(ctx, gaugeTable); err != nil {
		return fmt.Errorf("err creating gauge_metrics table: %w", err)
	}

	if _, err := pool.Exec(ctx, counterTable); err != nil {
		return fmt.Errorf("err creating counter_metrics table: %w", err)
	}

	return nil
}
