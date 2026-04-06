package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-task-runner/internal/models"
)

func TestRegisterRoutesWiresCreateJobHandler(t *testing.T) {
	mux := http.NewServeMux()
	handler := NewHandler(&usecaseStub{
		createJobFn: func(ctx context.Context, job *models.Job) (string, error) {
			return "job-456", nil
		},
	})
	RegisterRoutes(mux, handler)

	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"type":"email","payload":{"to":"alice@example.com","subject":"Hello","body":"Hello world"},"max_retries":1}`))
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}
