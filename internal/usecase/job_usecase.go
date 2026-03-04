package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-task-runner/internal/jobs"
	"github.com/go-task-runner/internal/models"
	"github.com/go-task-runner/internal/repository"
	"github.com/google/uuid"
)

type JobUsecaseInterface interface {
	CreateJob(ctx context.Context, job *models.Job) (string, error)
}

type jobUsecase struct {
	jobRepo repository.JobRepositoryInterface

	validator *validator.Validate
}

func NewJobUsecase(jobRepo repository.JobRepositoryInterface, v *validator.Validate) JobUsecaseInterface {
	return &jobUsecase{
		jobRepo:   jobRepo,
		validator: v,
	}
}

func (j *jobUsecase) CreateJob(ctx context.Context, job *models.Job) (string, error) {
	log.Printf("in the usecase job with ID: %+v", job)

	if job.NextRunAt != nil && job.NextRunAt.Before(time.Now()) {
		return "", errors.New("Invalid schedule time for the task")
	}

	if err := jobs.ValidateJob(job, j.validator); err != nil {
		return "", err
	}

	id := uuid.New()
	job.ID = id

	var nextRun time.Time
	if job.NextRunAt != nil {
		nextRun = *job.NextRunAt
	} else {
		nextRun = time.Now()
	}
	job.NextRunAt = &nextRun

	newJob := &models.Job{
		ID:         id,
		Type:       job.Type,
		Payload:    job.Payload,
		Status:     models.StatusPending,
		RetryCount: 0,
		MaxRetries: job.MaxRetries,
		NextRunAt:  &nextRun,
	}

	return j.jobRepo.Create(ctx, newJob)
}
