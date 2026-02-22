package api

import (
	"encoding/json"
	"time"
)

type CreateJobRequest struct {
	Type       string          `json:"type,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	MaxRetries int             `json:"max_retries,omitempty"`
	NextRunAt  *time.Time      `json:"next_run_at,omitempty"`
}

type CreateJobResponse struct {
	ID string `json:"id,omitempty"`
}

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
