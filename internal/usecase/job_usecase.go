package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-task-runner/internal/jobs"
	"github.com/go-task-runner/internal/logger"
	"github.com/go-task-runner/internal/models"
	"github.com/go-task-runner/internal/repository"
	"github.com/google/uuid"
)

// ValidationError represents a client-side input error (HTTP 400).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type JobUsecaseInterface interface {
	CreateJob(ctx context.Context, job *models.Job) (string, error)
}

type jobUsecase struct {
	jobRepo   repository.JobRepositoryInterface
	validator *validator.Validate
}

func NewJobUsecase(jobRepo repository.JobRepositoryInterface, v *validator.Validate) JobUsecaseInterface {
	return &jobUsecase{
		jobRepo:   jobRepo,
		validator: v,
	}
}

func (j *jobUsecase) CreateJob(ctx context.Context, job *models.Job) (string, error) {
	if job.NextRunAt != nil && job.NextRunAt.Before(time.Now()) {
		return "", &ValidationError{Message: "invalid schedule time: must be in the future"}
	}

	if err := jobs.ValidateJob(job, j.validator); err != nil {
		return "", &ValidationError{Message: err.Error()}
	}

	now := time.Now()
	newJob := &models.Job{
		ID:         uuid.New(),
		Type:       job.Type,
		Payload:    job.Payload,
		Status:     models.StatusPending,
		RetryCount: 0,
		MaxRetries: job.MaxRetries,
		NextRunAt:  job.NextRunAt,
	}

	if newJob.NextRunAt == nil {
		newJob.NextRunAt = &now
	}

	id, err := j.jobRepo.Create(ctx, newJob)
	if err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}

	logger.Log.Info("job created", "job_id", id, "job_type", job.Type)
	return id, nil
}
