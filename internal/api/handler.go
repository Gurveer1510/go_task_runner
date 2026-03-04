package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-task-runner/internal/models"
	"github.com/go-task-runner/internal/usecase"
)

type Handler struct {
	usecase usecase.JobUsecaseInterface
}

func NewHandler(usecase usecase.JobUsecaseInterface) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) CreateJob(rw http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "invalid payload", http.StatusBadRequest)
		return
	}

	job := &models.Job{
		Type:       req.Type,
		Payload:    req.Payload,
		MaxRetries: req.MaxRetries,
		NextRunAt:  req.NextRunAt,
	}

	id, err := h.usecase.CreateJob(r.Context(), job)
	if err != nil {
		errResp := ErrorResponse{Error: err.Error()}
		http.Error(rw, errResp.Error, http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	log.Printf("in the handler job with ID: %+v", job)

	resp := CreateJobResponse{ID: id}
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(resp)
}
