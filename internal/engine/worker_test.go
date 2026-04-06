package engine

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-task-runner/internal/logger"
	"github.com/go-task-runner/internal/models"
	"github.com/google/uuid"
)

type workerRepoStub struct {
	claimJobFn      func(ctx context.Context, workerID string, staleAfter time.Duration) (*models.Job, error)
	markRetryingFn  func(ctx context.Context, jobID string, workerID string, delay time.Duration) error
	markFailedFn    func(ctx context.Context, jobID string, workerID string) error
	markCompletedFn func(ctx context.Context, jobID string, workerID string) error
	refreshLockFn   func(ctx context.Context, jobID string, workerID string) error
}

func (s *workerRepoStub) Create(context.Context, *models.Job) (string, error) {
	return "", nil
}

func (s *workerRepoStub) MarkFailed(ctx context.Context, jobID string, workerID string) error {
	if s.markFailedFn != nil {
		return s.markFailedFn(ctx, jobID, workerID)
	}
	return nil
}

func (s *workerRepoStub) MarkCompleted(ctx context.Context, jobID string, workerID string) error {
	if s.markCompletedFn != nil {
		return s.markCompletedFn(ctx, jobID, workerID)
	}
	return nil
}

func (s *workerRepoStub) MarkRetrying(ctx context.Context, jobID string, workerID string, delay time.Duration) error {
	if s.markRetryingFn != nil {
		return s.markRetryingFn(ctx, jobID, workerID, delay)
	}
	return nil
}

func (s *workerRepoStub) ClaimJob(ctx context.Context, workerID string, staleAfter time.Duration) (*models.Job, error) {
	if s.claimJobFn != nil {
		return s.claimJobFn(ctx, workerID, staleAfter)
	}
	return nil, nil
}

func (s *workerRepoStub) RefreshLock(ctx context.Context, jobID string, workerID string) error {
	if s.refreshLockFn != nil {
		return s.refreshLockFn(ctx, jobID, workerID)
	}
	return nil
}

func TestRunJobRecoversFromHandlerPanic(t *testing.T) {
	logger.Init("development")

	engine := New(&workerRepoStub{}, executorFunc(func(ctx context.Context, job *models.Job) error {
		panic("boom")
	}), 1, 10*time.Millisecond, time.Hour)

	err := engine.runJob(context.Background(), "worker-1", &models.Job{
		ID:   uuid.New(),
		Type: "email",
	})
	if err == nil {
		t.Fatal("expected panic to be converted into an error")
	}
	if !strings.Contains(err.Error(), "job handler panicked") {
		t.Fatalf("expected panic error, got %q", err.Error())
	}
}

func TestWorkerMarksCompletedUsingBackgroundContextAfterCancellation(t *testing.T) {
	logger.Init("development")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	completed := make(chan struct{}, 1)
	job := &models.Job{
		ID:   uuid.New(),
		Type: "email",
	}
	repo := &workerRepoStub{
		claimJobFn: func(ctx context.Context, workerID string, staleAfter time.Duration) (*models.Job, error) {
			return job, nil
		},
		markCompletedFn: func(ctx context.Context, jobID string, workerID string) error {
			if ctx.Err() != nil {
				t.Fatal("expected background context for completion update")
			}
			if jobID != job.ID.String() {
				t.Fatalf("expected job id %s, got %s", job.ID.String(), jobID)
			}
			if workerID != "worker-0" {
				t.Fatalf("expected worker-0, got %s", workerID)
			}
			completed <- struct{}{}
			return nil
		},
	}
	engine := New(repo, executorFunc(func(ctx context.Context, job *models.Job) error {
		cancel()
		return nil
	}), 1, 5*time.Millisecond, time.Hour)

	engine.Start(ctx)

	select {
	case <-completed:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for completion update")
	}

	waitDone := make(chan struct{})
	go func() {
		engine.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker shutdown")
	}
}

func TestWorkerMarksRetryingOnExecutionError(t *testing.T) {
	logger.Init("development")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	retrying := make(chan struct{}, 1)
	job := &models.Job{
		ID:         uuid.New(),
		Type:       "email",
		RetryCount: 0,
		MaxRetries: 2,
	}
	baseDelay := 20 * time.Millisecond
	repo := &workerRepoStub{
		claimJobFn: func(ctx context.Context, workerID string, staleAfter time.Duration) (*models.Job, error) {
			return job, nil
		},
		markRetryingFn: func(ctx context.Context, jobID string, workerID string, delay time.Duration) error {
			if ctx.Err() != nil {
				t.Fatal("expected background context for retry update")
			}
			if workerID != "worker-0" {
				t.Fatalf("expected worker-0, got %s", workerID)
			}
			if delay < baseDelay {
				t.Fatalf("expected retry delay >= %v, got %v", baseDelay, delay)
			}
			retrying <- struct{}{}
			cancel()
			return nil
		},
	}
	engine := New(repo, executorFunc(func(ctx context.Context, job *models.Job) error {
		return errors.New("handler failed")
	}), 1, baseDelay, time.Hour)

	engine.Start(ctx)

	select {
	case <-retrying:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for retry update")
	}

	waitDone := make(chan struct{})
	go func() {
		engine.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker shutdown")
	}
}
