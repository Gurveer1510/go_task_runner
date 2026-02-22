package repository

import (
	"context"

	"github.com/go-task-runner/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO jobs (
			id, type, payload, status, retry_count, max_retries, next_run_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		job.ID,
		job.Type,
		job.Payload,
		job.Status,
		job.RetryCount,
		job.MaxRetries,
		job.NextRunAt,
	)

	return err
}
