package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-task-runner/internal/jobs"
	"github.com/go-task-runner/internal/models"
	"github.com/go-task-runner/internal/repository"
	"github.com/google/uuid"
)

type Handler struct {
	repo      repository.JobRepositoryInterface
	validator *validator.Validate
}

func NewHandler(repo repository.JobRepositoryInterface, v *validator.Validate) *Handler {
	return &Handler{repo: repo, validator: v}
}

func (h *Handler) CreateJob(rw http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	now := time.Now()
	if req.NextRunAt != nil && req.NextRunAt.Before(now) {
		http.Error(rw, "Invalid schedule time for the task", http.StatusBadRequest)
		return
	}

	switch req.Type {
	case "email":
		var emailPayload jobs.EmailPayload

		if err := json.Unmarshal(req.Payload, &emailPayload); err != nil {
			http.Error(rw, "invalid email payload format", http.StatusBadRequest)
			return
		}

		if err := h.validator.Struct(emailPayload); err != nil {
			validationErrors := err.(validator.ValidationErrors)

			for _, e := range validationErrors {
				switch e.Field() {
				case "To":
					http.Error(rw, "invalid email address", http.StatusBadRequest)
					return
				case "Subject":
					http.Error(rw, "subject must be between 3 and 200 characters", http.StatusBadRequest)
					return
				case "Body":
					http.Error(rw, "body must be at least 5 characters", http.StatusBadRequest)
					return
				}
			}
		}
	default:
		http.Error(rw, "unsupported job type", http.StatusBadRequest)
		return
	}

	id := uuid.New()

	var nextRun time.Time

	if req.NextRunAt != nil {
		nextRun = *req.NextRunAt
	} else {
		nextRun = now
	}

	job := &models.Job{
		ID:         id,
		Type:       req.Type,
		Payload:    req.Payload,
		Status:     models.StatusPending,
		RetryCount: 0,
		MaxRetries: req.MaxRetries,
		NextRunAt:  nextRun,
	}

	err := h.repo.Create(r.Context(), job)
	if err != nil {
		http.Error(rw, "failed to create job [Error from REPO]", http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	log.Printf("in the handler job with ID: %+v", job)

	resp := CreateJobResponse{ID: id.String()}
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(resp)
}
