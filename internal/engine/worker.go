package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

func (e *Engine) Start(ctx context.Context) {
	for i := 0 ; i < e.concurrency; i++ {
		go e.workerLoop(ctx, i)
	}
}


func (e *Engine) workerLoop(ctx context.Context, id int) {
	workerID := fmt.Sprintf("worker-%d",id)

	for {
		select {
			case <-ctx.Done():
				return
			default:
		}

		job, err := e.repository.ClaimJob(ctx, workerID)
		if err != nil {
			if err == pgx.ErrNoRows {
				time.Sleep(300*time.Millisecond)
				continue
			}
			time.Sleep(time.Second)
			continue
		}

		err = e.executor.Execute(ctx, job)
		if err != nil {
			if job.RetryCount < job.MaxRetries {
				e.repository.MarkFailed(ctx, job.ID.String())
			} else {
				e.repository.MarkFailed(ctx, job.ID.String())
			}
		} else {
			e.repository.MarkCompleted(ctx, job.ID.String())
		}
	}
}