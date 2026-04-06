package engine

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-task-runner/internal/models"
)

func TestDefaultExecutorExecutesRegisteredHandler(t *testing.T) {
	t.Parallel()

	executor := NewDefaultExecutor()
	called := false
	executor.Register("email", func(ctx context.Context, payload []byte) error {
		called = true
		if string(payload) != `{"ok":true}` {
			t.Fatalf("expected payload to be forwarded, got %s", string(payload))
		}
		return nil
	})

	err := executor.Execute(context.Background(), &models.Job{
		Type:    "email",
		Payload: json.RawMessage(`{"ok":true}`),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected registered handler to be called")
	}
}

func TestDefaultExecutorReturnsErrorForUnknownJobType(t *testing.T) {
	t.Parallel()

	executor := NewDefaultExecutor()

	err := executor.Execute(context.Background(), &models.Job{Type: "email"})
	if err == nil {
		t.Fatal("expected missing handler error, got nil")
	}
	if !strings.Contains(err.Error(), "no handler registered") {
		t.Fatalf("expected missing handler error, got %q", err.Error())
	}
}
