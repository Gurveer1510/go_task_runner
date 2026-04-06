package jobs

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/go-task-runner/internal/models"
)

func TestValidateJob(t *testing.T) {
	t.Parallel()

	validator := validator.New()

	tests := []struct {
		name    string
		job     *models.Job
		wantErr string
	}{
		{
			name: "valid email job",
			job: &models.Job{
				Type:    "email",
				Payload: json.RawMessage(`{"to":"alice@example.com","subject":"Hello","body":"Hello world"}`),
			},
		},
		{
			name: "invalid payload format",
			job: &models.Job{
				Type:    "email",
				Payload: json.RawMessage(`{"to":`),
			},
			wantErr: "invalid email payload format",
		},
		{
			name: "invalid email address",
			job: &models.Job{
				Type:    "email",
				Payload: json.RawMessage(`{"to":"not-an-email","subject":"Hello","body":"Hello world"}`),
			},
			wantErr: "invalid email address",
		},
		{
			name: "unsupported job type",
			job: &models.Job{
				Type:    "sms",
				Payload: json.RawMessage(`{}`),
			},
			wantErr: "unsupported job type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJob(tt.job, validator)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}
