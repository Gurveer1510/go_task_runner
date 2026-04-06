package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-task-runner/internal/logger"
	"github.com/go-task-runner/internal/models"
)

type jobRepoStub struct {
	createFn func(ctx context.Context, job *models.Job) (string, error)
}

func (s *jobRepoStub) Create(ctx context.Context, job *models.Job) (string, error) {
	if s.createFn != nil {
		return s.createFn(ctx, job)
	}
	return "", nil
}

func (s *jobRepoStub) MarkFailed(context.Context, string, string) error {
	return nil
}

func (s *jobRepoStub) MarkCompleted(context.Context, string, string) error {
	return nil
}

func (s *jobRepoStub) MarkRetrying(context.Context, string, string, time.Duration) error {
	return nil
}

func (s *jobRepoStub) ClaimJob(context.Context, string, time.Duration) (*models.Job, error) {
	return nil, nil
}

func (s *jobRepoStub) RefreshLock(context.Context, string, string) error {
	return nil
}

func TestCreateJobRejectsPastSchedule(t *testing.T) {
	logger.Init("development")

	usecase := NewJobUsecase(&jobRepoStub{}, validator.New())
	past := time.Now().Add(-time.Minute)

	_, err := usecase.CreateJob(context.Background(), &models.Job{
		Type:       "email",
		Payload:    json.RawMessage(`{"to":"alice@example.com","subject":"Hello","body":"Hello world"}`),
		MaxRetries: 1,
		NextRunAt:  &past,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestCreateJobRejectsInvalidPayload(t *testing.T) {
	logger.Init("development")

	usecase := NewJobUsecase(&jobRepoStub{}, validator.New())

	_, err := usecase.CreateJob(context.Background(), &models.Job{
		Type:       "email",
		Payload:    json.RawMessage(`{"to":"bad-email","subject":"Hello","body":"Hello world"}`),
		MaxRetries: 1,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if !strings.Contains(validationErr.Error(), "invalid email address") {
		t.Fatalf("expected invalid email validation message, got %q", validationErr.Error())
	}
}

func TestCreateJobDefaultsJobFieldsBeforePersisting(t *testing.T) {
	logger.Init("development")

	var createdJob *models.Job
	repo := &jobRepoStub{
		createFn: func(ctx context.Context, job *models.Job) (string, error) {
			jobCopy := *job
			createdJob = &jobCopy
			return "job-123", nil
		},
	}
	usecase := NewJobUsecase(repo, validator.New())

	before := time.Now()
	id, err := usecase.CreateJob(context.Background(), &models.Job{
		Type:       "email",
		Payload:    json.RawMessage(`{"to":"alice@example.com","subject":"Hello","body":"Hello world"}`),
		MaxRetries: 3,
	})
	after := time.Now()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "job-123" {
		t.Fatalf("expected id job-123, got %q", id)
	}
	if createdJob == nil {
		t.Fatal("expected repository Create to be called")
	}
	if createdJob.ID.String() == "" {
		t.Fatal("expected generated job ID")
	}
	if createdJob.Status != models.StatusPending {
		t.Fatalf("expected pending status, got %q", createdJob.Status)
	}
	if createdJob.RetryCount != 0 {
		t.Fatalf("expected retry count 0, got %d", createdJob.RetryCount)
	}
	if createdJob.NextRunAt == nil {
		t.Fatal("expected next_run_at to be set")
	}
	if createdJob.NextRunAt.Before(before.Add(-time.Second)) || createdJob.NextRunAt.After(after.Add(time.Second)) {
		t.Fatalf("expected next_run_at near now, got %v", createdJob.NextRunAt)
	}
}

func TestCreateJobWrapsRepositoryErrors(t *testing.T) {
	logger.Init("development")

	repo := &jobRepoStub{
		createFn: func(ctx context.Context, job *models.Job) (string, error) {
			return "", errors.New("insert failed")
		},
	}
	usecase := NewJobUsecase(repo, validator.New())

	_, err := usecase.CreateJob(context.Background(), &models.Job{
		Type:       "email",
		Payload:    json.RawMessage(`{"to":"alice@example.com","subject":"Hello","body":"Hello world"}`),
		MaxRetries: 1,
	})
	if err == nil {
		t.Fatal("expected repository error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create job") {
		t.Fatalf("expected wrapped repository error, got %q", err.Error())
	}
}
