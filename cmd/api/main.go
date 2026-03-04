package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-task-runner/internal/api"
	"github.com/go-task-runner/internal/config"
	"github.com/go-task-runner/internal/db"
	"github.com/go-task-runner/internal/repository"
	"github.com/go-task-runner/internal/usecase"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	pool := db.NewPool(cfg.DBUrl)
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
		log.Println("Server running on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not start server: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		if err := server.Close(); err != nil {
			log.Fatalf("force closed failed: %v", err)
		}
	}

	log.Println("Server exited.")

}
