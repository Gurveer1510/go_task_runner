package engine

import (
	"sync"
	"time"

	"github.com/go-task-runner/internal/repository"
)

type Engine struct {
	Wg          *sync.WaitGroup
	repository  repository.JobRepositoryInterface
	executor    Executor
	concurrency int
	baseDelay   time.Duration
}

func New(repo repository.JobRepositoryInterface, executor Executor, concurrency int, baseDelay time.Duration) *Engine {
	return &Engine{
		Wg:          &sync.WaitGroup{},
		repository:  repo,
		executor:    executor,
		concurrency: concurrency,
		baseDelay:   baseDelay,
	}
}
