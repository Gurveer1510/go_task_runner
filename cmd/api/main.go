package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-task-runner/internal/api"
	"github.com/go-task-runner/internal/config"
	"github.com/go-task-runner/internal/db"
	"github.com/go-task-runner/internal/logger"
	"github.com/go-task-runner/internal/repository"
	"github.com/go-task-runner/internal/usecase"
)

func main() {
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

	jobRepo := repository.NewJobRepository(pool)
	v := validator.New()
	jobUsecase := usecase.NewJobUsecase(jobRepo, v)
	handler := api.NewHandler(jobUsecase)

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, handler)

	server := http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Log.Info("api server started", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	logger.Log.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Error("graceful shutdown failed", "error", err)
		if err := server.Close(); err != nil {
			logger.Log.Error("force close failed", "error", err)
		}
	}

	logger.Log.Info("server exited")
}
