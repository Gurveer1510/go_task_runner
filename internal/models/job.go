package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
	StatusRetrying   JobStatus = "retrying"
)

type Job struct {
	ID         uuid.UUID
	Type       string
	Payload    json.RawMessage
	Status     JobStatus
	RetryCount int
	MaxRetries int
	NextRunAt  *time.Time
	LockedBy   *string
	LockedAt   *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
