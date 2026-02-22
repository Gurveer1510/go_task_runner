package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewPool(databaseUrl string) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, databaseUrl)
	if err != nil {
		log.Fatalf("Unable to create the connection pool: %v", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to ping the database: %v", err)
	}

	log.Println("Database connected successfully")

	return pool
}