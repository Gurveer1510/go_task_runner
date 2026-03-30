package api

import (
	"encoding/json"
	"time"
)

type CreateJobRequest struct {
	Type       string          `json:"type,omitempty" validate:"required"`
	Payload    json.RawMessage `json:"payload,omitempty" validate:"required"`
	MaxRetries int             `json:"max_retries,omitempty" validate:"gte=0,lte=10,required"`
	NextRunAt  *time.Time      `json:"next_run_at,omitempty"`
}

type CreateJobResponse struct {
	ID string `json:"id,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}