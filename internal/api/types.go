package api

import (
	"encoding/json"
	"time"
)

type CreateJobRequest struct {
	Type       string          `json:"type,omitempty" validate:"required"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	MaxRetries int             `json:"max_retries,omitempty" validate:"gte=0,lte=10"`
	NextRunAt  *time.Time      `json:"next_run_at,omitempty"`
}

type CreateJobResponse struct {
	ID string `json:"id,omitempty"`
}
