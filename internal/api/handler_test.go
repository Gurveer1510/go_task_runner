package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-task-runner/internal/models"
	"github.com/go-task-runner/internal/usecase"
)

type usecaseStub struct {
	createJobFn func(ctx context.Context, job *models.Job) (string, error)
}

func (s *usecaseStub) CreateJob(ctx context.Context, job *models.Job) (string, error) {
	if s.createJobFn != nil {
		return s.createJobFn(ctx, job)
	}
	return "", nil
}

func TestCreateJobReturnsBadRequestForInvalidPayload(t *testing.T) {
	handler := NewHandler(&usecaseStub{})
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"type":"email","payload":{},"max_retries":1,"extra":true}`))
	recorder := httptest.NewRecorder()

	handler.CreateJob(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("expected JSON error response, got %v", err)
	}
	if response.Error != "invalid payload" {
		t.Fatalf("expected invalid payload error, got %q", response.Error)
	}
}

func TestCreateJobReturnsValidationErrors(t *testing.T) {
	handler := NewHandler(&usecaseStub{
		createJobFn: func(ctx context.Context, job *models.Job) (string, error) {
			return "", &usecase.ValidationError{Message: "invalid email address"}
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"type":"email","payload":{"to":"bad-email"},"max_retries":1}`))
	recorder := httptest.NewRecorder()

	handler.CreateJob(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("expected JSON validation response, got %v", err)
	}
	if response.Error != "invalid email address" {
		t.Fatalf("expected validation error, got %q", response.Error)
	}
}

func TestCreateJobReturnsInternalServerError(t *testing.T) {
	handler := NewHandler(&usecaseStub{
		createJobFn: func(ctx context.Context, job *models.Job) (string, error) {
			return "", errors.New("database unavailable")
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"type":"email","payload":{"to":"alice@example.com","subject":"Hello","body":"Hello world"},"max_retries":1}`))
	recorder := httptest.NewRecorder()

	handler.CreateJob(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "database unavailable") {
		t.Fatalf("expected response body to contain internal error, got %q", recorder.Body.String())
	}
}

func TestCreateJobReturnsCreatedJobID(t *testing.T) {
	var receivedJob *models.Job
	handler := NewHandler(&usecaseStub{
		createJobFn: func(ctx context.Context, job *models.Job) (string, error) {
			jobCopy := *job
			receivedJob = &jobCopy
			return "job-123", nil
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"type":"email","payload":{"to":"alice@example.com","subject":"Hello","body":"Hello world"},"max_retries":2}`))
	recorder := httptest.NewRecorder()

	handler.CreateJob(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response CreateJobResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("expected JSON success response, got %v", err)
	}
	if response.ID != "job-123" {
		t.Fatalf("expected response id job-123, got %q", response.ID)
	}
	if receivedJob == nil {
		t.Fatal("expected usecase to receive a job")
	}
	if receivedJob.Type != "email" {
		t.Fatalf("expected email job type, got %q", receivedJob.Type)
	}
	if receivedJob.MaxRetries != 2 {
		t.Fatalf("expected max retries 2, got %d", receivedJob.MaxRetries)
	}
}
