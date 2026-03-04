package engine

import (
	"time"

	"github.com/go-task-runner/internal/repository"
)

type Engine struct {
	repository  repository.JobRepositoryInterface
	executor    Executor
	concurrency int
	baseDelay   time.Duration
}

func New(repo *repository.JobRepo, executor Executor, concurrency int, baseDelay time.Duration) *Engine {
	return &Engine{
		repository:  repo,
		executor:    executor,
		concurrency: concurrency,
		baseDelay:   baseDelay,
	}
}
