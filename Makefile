APP_NAME=iacctl

.PHONY: help build run-api run-worker run-cli tidy fmt test up up-dev down down-dev logs logs-dev ps ps-dev

help:
	@echo "Usage:"
	@echo "  make build        # Build all Go binaries"
	@echo "  make run-api      # Run API server"
	@echo "  make run-worker   # Run background worker"
	@echo "  make run-cli      # Run CLI"
	@echo "  make tidy         # Run go mod tidy"
	@echo "  make fmt          # Format code"
	@echo "  make test         # Run tests"
	@echo "  make up           # Start production docker-compose"
	@echo "  make up-dev       # Start development docker-compose"
	@echo "  make down         # Stop production docker-compose"
	@echo "  make down-dev     # Stop development docker-compose"
	@echo "  make logs         # View production logs"
	@echo "  make logs-dev     # View development logs"
	@echo "  make ps           # Show production containers"
	@echo "  make ps-dev       # Show development containers"

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

up-dev:
	docker-compose -f deploy/docker-compose.dev.yml up -d

down:
	docker-compose -f deploy/docker-compose.yml down

down-dev:
	docker-compose -f deploy/docker-compose.dev.yml down

logs:
	docker-compose -f deploy/docker-compose.yml logs -f

logs-dev:
	docker-compose -f deploy/docker-compose.dev.yml logs -f

ps:
	docker-compose -f deploy/docker-compose.yml ps

ps-dev:
	docker-compose -f deploy/docker-compose.dev.yml ps
