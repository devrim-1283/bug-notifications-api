package db

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_create_bug_reports.sql
var migrationSQL string

// Connect creates a new PostgreSQL connection pool.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	config.MaxConns = 20
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// Migrate runs all embedded SQL migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	slog.Info("running database migrations...")
	if _, err := pool.Exec(ctx, migrationSQL); err != nil {
		return fmt.Errorf("run migration: %w", err)
	}
	slog.Info("database migrations completed")
	return nil
}
