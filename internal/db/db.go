package db

import (
	"context"
	"fmt"
	"time"

	"github.com/go-task-runner/internal/logger"
	"github.com/jackc/pgx/v4/pgxpool"
)

func NewPool(databaseUrl string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Log.Info("database connected successfully")

	return pool, nil
}
