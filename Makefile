APP_NAME=iacctl

.PHONY: help build run-api run-worker run-cli tidy fmt test up down

help:
	@echo "Usage:"
	@echo "  make build        # Build all Go binaries"
	@echo "  make run-api      # Run API server"
	@echo "  make run-worker   # Run background worker"
	@echo "  make run-cli      # Run CLI"
	@echo "  make tidy         # Run go mod tidy"
	@echo "  make fmt          # Format code"
	@echo "  make test         # Run tests"
	@echo "  make up           # Start docker-compose dev env"
	@echo "  make down         # Stop docker-compose"

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker
	go build -o bin/cli ./cmd/cli

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

run-cli:
	go run ./cmd/cli

tidy:
	go mod tidy

fmt:
	go fmt ./...

test:
	go test ./...

up:
	docker-compose -f deploy/docker-compose.yml up -d

down:
	docker-compose -f deploy/docker-compose.yml down
