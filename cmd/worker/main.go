package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-task-runner/internal/config"
	"github.com/go-task-runner/internal/db"
	"github.com/go-task-runner/internal/engine"
	"github.com/go-task-runner/internal/logger"
	"github.com/go-task-runner/internal/repository"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logger.Init(cfg.AppEnv)

	pool, err := db.NewPool(cfg.DBUrl)
	if err != nil {
		logger.Log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewJobRepository(pool)
	executor := engine.NewDefaultExecutor()

	executor.Register("email", func(ctx context.Context, payload []byte) error {
		// Placeholder — will be replaced with jobs.EmailHandler
		fmt.Println("sending email:", string(payload))
		return nil
	})

	eng := engine.New(repo, executor, cfg.Concurrency, cfg.BaseDelay)

	logger.Log.Info("worker started", "concurrency", cfg.Concurrency, "base_delay", cfg.BaseDelay)

	eng.Start(ctx)

	<-ctx.Done()
	eng.Wg.Wait()
	logger.Log.Info("worker shutdown complete")
}
