package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-task-runner/internal/models"
	"github.com/go-task-runner/internal/repository"
	"github.com/google/uuid"
)

type Handler struct{
	repo *repository.JobRepository
}

func NewHandler(repo *repository.JobRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) CreateJob(rw http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "invalid payload", http.StatusBadRequest)
		return
	}

	if req.NextRunAt != nil && req.NextRunAt.Before(time.Now()) {
		http.Error(rw, "Invalid schedule time for the task", http.StatusBadRequest)
		return
	}

	id := uuid.New()

	now := time.Now()

	var nextRun time.Time

	if req.NextRunAt != nil {
		nextRun = *req.NextRunAt
	} else {
		nextRun = now
	}

	job := &models.Job {
		ID: id,
		Type: req.Type,
		Payload: req.Payload,
		Status: models.StatusPending,
		RetryCount: 0,
		MaxRetries: req.MaxRetries,
		NextRunAt: nextRun,
	}

	err := h.repo.Create(context.Background(), job)
	if err != nil {
		http.Error(rw, "failed to create job [Error from REPO]", http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	resp := CreateJobResponse{ID: id.String()}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(resp)
}