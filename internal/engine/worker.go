package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/go-task-runner/internal/logger"
	"github.com/jackc/pgx/v4"
)

func (e *Engine) Start(ctx context.Context) {
	for i := 0; i < e.concurrency; i++ {
		go e.workerLoop(ctx, i)
	}
}

func (e *Engine) workerLoop(ctx context.Context, id int) {
	workerID := fmt.Sprintf("worker-%d", id)

	logger.Log.Info("worker started", "worker_id", workerID)

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("worker shutting down", "worker_id", workerID)
			return
		default:
		}

		job, err := e.repository.ClaimJob(ctx, workerID)
		if err != nil {
			if err == pgx.ErrNoRows {
				time.Sleep(300 * time.Millisecond)
				continue
			}
			logger.Log.Error("failed to claim job", "worker_id", workerID, "error", err)
			time.Sleep(time.Second)
			continue
		}

		logger.Log.Info("job claimed", "worker_id", workerID, "job_id", job.ID.String(), "job_type", job.Type)

		err = e.executor.Execute(ctx, job)
		if err != nil {
			logger.Log.Error("job execution failed", "worker_id", workerID, "job_id", job.ID.String(), "error", err)
			if job.RetryCount < job.MaxRetries {
				if markErr := e.repository.MarkRetrying(ctx, job.ID.String()); markErr != nil {
					logger.Log.Error("failed to mark job as retrying", "job_id", job.ID.String(), "error", markErr)
				}
			} else {
				if markErr := e.repository.MarkFailed(ctx, job.ID.String()); markErr != nil {
					logger.Log.Error("failed to mark job as failed", "job_id", job.ID.String(), "error", markErr)
				}
			}
		} else {
			logger.Log.Info("job completed", "worker_id", workerID, "job_id", job.ID.String())
			if markErr := e.repository.MarkCompleted(ctx, job.ID.String()); markErr != nil {
				logger.Log.Error("failed to mark job as completed", "job_id", job.ID.String(), "error", markErr)
			}
		}
	}
}
