package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-task-runner/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type fakeStore struct {
	execFn     func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	queryRowFn func(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func (s *fakeStore) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if s.execFn != nil {
		return s.execFn(ctx, sql, args...)
	}
	return pgconn.CommandTag([]byte("UPDATE 1")), nil
}

func (s *fakeStore) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if s.queryRowFn != nil {
		return s.queryRowFn(ctx, sql, args...)
	}
	return fakeRow{}
}

type fakeRow struct {
	scanFn func(dest ...interface{}) error
}

func (r fakeRow) Scan(dest ...interface{}) error {
	if r.scanFn != nil {
		return r.scanFn(dest...)
	}
	return nil
}

func TestCreateReturnsInsertedID(t *testing.T) {
	expectedID := uuid.New()
	repo := newJobRepositoryForTest(&fakeStore{
		queryRowFn: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return fakeRow{
				scanFn: func(dest ...interface{}) error {
					*(dest[0].(*uuid.UUID)) = expectedID
					return nil
				},
			}
		},
	})

	id, err := repo.Create(context.Background(), &models.Job{
		ID:         expectedID,
		Type:       "email",
		Payload:    json.RawMessage(`{"ok":true}`),
		Status:     models.StatusPending,
		RetryCount: 0,
		MaxRetries: 1,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != expectedID.String() {
		t.Fatalf("expected id %s, got %s", expectedID.String(), id)
	}
}

func TestMarkCompletedReturnsErrJobLockLostWhenNoRowsAffected(t *testing.T) {
	repo := newJobRepositoryForTest(&fakeStore{
		execFn: func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
			return pgconn.CommandTag([]byte("UPDATE 0")), nil
		},
	})

	err := repo.MarkCompleted(context.Background(), uuid.NewString(), "worker-1")
	if err != ErrJobLockLost {
		t.Fatalf("expected ErrJobLockLost, got %v", err)
	}
}

func TestMarkRetryingUsesDelayMillisecondsAndWorkerID(t *testing.T) {
	delay := 3 * time.Second
	repo := newJobRepositoryForTest(&fakeStore{
		execFn: func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
			if len(args) != 3 {
				t.Fatalf("expected 3 args, got %d", len(args))
			}
			if args[1] != delay.Milliseconds() {
				t.Fatalf("expected delay milliseconds %d, got %v", delay.Milliseconds(), args[1])
			}
			if args[2] != "worker-1" {
				t.Fatalf("expected worker id worker-1, got %v", args[2])
			}
			return pgconn.CommandTag([]byte("UPDATE 1")), nil
		},
	})

	err := repo.MarkRetrying(context.Background(), uuid.NewString(), "worker-1", delay)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestClaimJobUsesLeaseDurationAndScansRow(t *testing.T) {
	expectedID := uuid.New()
	nextRunAt := time.Now().Add(time.Minute)
	lockedAt := time.Now()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()
	lockedBy := "worker-1"
	lease := 45 * time.Second
	repo := newJobRepositoryForTest(&fakeStore{
		queryRowFn: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			if len(args) != 2 {
				t.Fatalf("expected 2 args, got %d", len(args))
			}
			if args[0] != "worker-2" {
				t.Fatalf("expected worker id worker-2, got %v", args[0])
			}
			if args[1] != lease.Milliseconds() {
				t.Fatalf("expected lease milliseconds %d, got %v", lease.Milliseconds(), args[1])
			}

			return fakeRow{
				scanFn: func(dest ...interface{}) error {
					*(dest[0].(*uuid.UUID)) = expectedID
					*(dest[1].(*string)) = "email"
					*(dest[2].(*json.RawMessage)) = json.RawMessage(`{"ok":true}`)
					*(dest[3].(*models.JobStatus)) = models.StatusProcessing
					*(dest[4].(*int)) = 1
					*(dest[5].(*int)) = 3
					*(dest[6].(**time.Time)) = &nextRunAt
					*(dest[7].(**string)) = &lockedBy
					*(dest[8].(**time.Time)) = &lockedAt
					*(dest[9].(*time.Time)) = createdAt
					*(dest[10].(*time.Time)) = updatedAt
					return nil
				},
			}
		},
	})

	job, err := repo.ClaimJob(context.Background(), "worker-2", lease)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if job.ID != expectedID {
		t.Fatalf("expected job id %s, got %s", expectedID, job.ID)
	}
	if job.LockedBy == nil || *job.LockedBy != lockedBy {
		t.Fatalf("expected locked_by %s, got %v", lockedBy, job.LockedBy)
	}
	if job.NextRunAt == nil || !job.NextRunAt.Equal(nextRunAt) {
		t.Fatalf("expected next_run_at %v, got %v", nextRunAt, job.NextRunAt)
	}
}

func TestRefreshLockReturnsErrJobLockLostWhenNoRowsAffected(t *testing.T) {
	repo := newJobRepositoryForTest(&fakeStore{
		execFn: func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
			return pgconn.CommandTag([]byte("UPDATE 0")), nil
		},
	})

	err := repo.RefreshLock(context.Background(), uuid.NewString(), "worker-1")
	if err != ErrJobLockLost {
		t.Fatalf("expected ErrJobLockLost, got %v", err)
	}
}
