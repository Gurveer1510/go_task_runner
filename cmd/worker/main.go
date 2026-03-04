package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-task-runner/internal/config"
	"github.com/go-task-runner/internal/db"
	"github.com/go-task-runner/internal/engine"
	"github.com/go-task-runner/internal/repository"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, _ := config.LoadConfig()

	pool := db.NewPool(cfg.DBUrl)
	defer pool.Close()

	repo := repository.NewJobRepository(pool)
	executor := engine.NewDefaultExecutor()

	executor.Register("email", func(ctx context.Context, payload []byte) error {
		// Simulate processing an email job
		fmt.Println("sending email:", string(payload))
		return nil
	})

	engine := engine.New(repo, executor, 5, 2*time.Second)

	engine.Start(ctx)

	<-ctx.Done()
}
