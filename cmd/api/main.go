package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/go-task-runner/internal/api"
	"github.com/go-task-runner/internal/config"
	"github.com/go-task-runner/internal/db"
	"github.com/go-task-runner/internal/repository"
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
	handler := api.NewHandler(jobRepo, v)

	http.HandleFunc("/jobs", handler.CreateJob)

	log.Println("API running on", cfg.Port)

	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%v",cfg.Port), nil))
}
