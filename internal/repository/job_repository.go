package repository

import (
	"context"

	"github.com/go-task-runner/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type JobRepositoryInterface interface {
	Create(ctx context.Context, job *models.Job) (string, error)
	MarkFailed(ctx context.Context, jobID string) error
	MarkCompleted(ctx context.Context, jobID string) error
	MarkRetrying(ctx context.Context, jobID string) error
	ClaimJob(ctx context.Context, workerID string) (*models.Job, error)
}

type JobRepo struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepo {
	return &JobRepo{db: db}
}

func (r *JobRepo) Create(ctx context.Context, job *models.Job) (string, error) {
	query := `
		INSERT INTO jobs (
			id, type, payload, status, retry_count, max_retries, next_run_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id
	`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, query,
		job.ID,
		job.Type,
		job.Payload,
		job.Status,
		job.RetryCount,
		job.MaxRetries,
		job.NextRunAt,
	).Scan(&id)

	return id.String(), err
}

func (r *JobRepo) MarkFailed(ctx context.Context, jobID string) error {
	query := `
		UPDATE jobs
		SET status = 'failed'
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, jobID)
	return err
}

func (r *JobRepo) MarkCompleted(ctx context.Context, jobID string) error {
	query := `
		UPDATE jobs
		SET status = 'completed'
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, jobID)
	return err
}

func (r *JobRepo) MarkRetrying(ctx context.Context, jobID string) error {
	query := `
		UPDATE jobs
		SET status = 'retrying',
			retry_count = retry_count + 1,
			next_run_at = NOW() + INTERVAL '1 second' * (retry_count + 1),
			locked_by = NULL,
			locked_at = NULL,
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, jobID)
	return err
}

func (r *JobRepo) ClaimJob(ctx context.Context, workerID string) (*models.Job, error) {
	query := `
	UPDATE jobs
	SET status = 'processing',
	    locked_by = $1,
	    locked_at = NOW(),
	    updated_at = NOW()
	WHERE id = (
	    SELECT id
	    FROM jobs
	    WHERE status IN ('pending', 'retrying')
	      AND next_run_at <= NOW()
	    ORDER BY created_at
	    FOR UPDATE SKIP LOCKED
	    LIMIT 1
	)
	RETURNING id, type, payload, status, retry_count, max_retries, next_run_at, locked_by, locked_at, created_at, updated_at;
	`

	var job models.Job

	err := r.db.QueryRow(ctx, query, workerID).Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.RetryCount,
		&job.MaxRetries,
		&job.NextRunAt,
		&job.LockedBy,
		&job.LockedAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &job, nil
}
