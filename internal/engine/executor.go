package engine

import (
	"context"
	"fmt"

	"github.com/go-task-runner/internal/models"
)

type Executor interface {
	Execute(ctx context.Context, job *models.Job) error
}


type JobHandler func(ctx context.Context, payload[]byte) error

type DefaultExecutor struct {
	handlers map[string]JobHandler
}

func NewDefaultExecutor() *DefaultExecutor {
	return &DefaultExecutor{
		handlers: make(map[string]JobHandler),
	}
}

func (e *DefaultExecutor) Register(jobType string, handler JobHandler) {
	e.handlers[jobType] = handler
}

func (e *DefaultExecutor) Execute(ctx context.Context, job *models.Job) error {
	handler, ok := e.handlers[job.Type]
	if !ok {
		return fmt.Errorf("no handler registered fo job type: %s", job.Type)
	}

	return handler(ctx, job.Payload)
}

