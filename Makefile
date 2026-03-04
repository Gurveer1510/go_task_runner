.PHONY: api worker db-up migrate test

api:
	go run ./cmd/api

worker:
	go run ./cmd/worker

db-up:
	docker-compose up -d

test:
	go test ./...
