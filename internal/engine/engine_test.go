package engine

import (
	"context"
	"testing"
	"time"

	"github.com/go-task-runner/internal/models"
)

type noopRepository struct{}

func (noopRepository) Create(context.Context, *models.Job) (string, error) {
	return "", nil
}

func (noopRepository) MarkFailed(context.Context, string, string) error {
	return nil
}

func (noopRepository) MarkCompleted(context.Context, string, string) error {
	return nil
}

func (noopRepository) MarkRetrying(context.Context, string, string, time.Duration) error {
	return nil
}

func (noopRepository) ClaimJob(context.Context, string, time.Duration) (*models.Job, error) {
	return nil, nil
}

func (noopRepository) RefreshLock(context.Context, string, string) error {
	return nil
}

type executorFunc func(ctx context.Context, job *models.Job) error

func (f executorFunc) Execute(ctx context.Context, job *models.Job) error {
	return f(ctx, job)
}

func TestNewDefaultsLeaseTiming(t *testing.T) {
	t.Parallel()

	engine := New(noopRepository{}, executorFunc(func(ctx context.Context, job *models.Job) error {
		return nil
	}), 1, time.Second, 0)

	if engine.leaseTime != defaultLeaseTime {
		t.Fatalf("expected default lease time %v, got %v", defaultLeaseTime, engine.leaseTime)
	}
	if engine.refreshTick != defaultLeaseTime/2 {
		t.Fatalf("expected refresh tick %v, got %v", defaultLeaseTime/2, engine.refreshTick)
	}
}
