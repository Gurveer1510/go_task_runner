package api

import (
	"encoding/json"
	"errors"
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
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		errorResp := ErrorResponse{Error: "invalid payload", Message: "failed"}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorResp)
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
		var valErr *usecase.ValidationError
		if errors.As(err, &valErr) {
			rw.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(rw).Encode(ErrorResponse{Error: valErr.Error(), Message: "failed"})
			return
		}
	}
	if err != nil {
		errResp := ErrorResponse{Error: err.Error(), Message: "failed"}
		http.Error(rw, errResp.Error, http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	log.Printf("in the handler job with ID: %+v", job)

	resp := CreateJobResponse{ID: id}
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(resp)
}
