package pg

import (
	"chatx-01-backend/pkg/errs"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func WrapRepoError(op string, err error) error {
	if isNotFound(err) {
		return errs.Wrap(op, errs.ErrNotFound)
	}
	if isConflict(err) {
		return errs.Wrap(op, errs.ErrAlreadyExists)
	}
	return errs.Wrap(op, err)
}

func isConflict(err error) bool {
	const pgConflictCode = "23505"

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgConflictCode
	}
	return false
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
