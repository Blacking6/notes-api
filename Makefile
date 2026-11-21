.PHONY: help run build test test-unit test-integration migrate-up migrate-down docker-up docker-down lint tidy

APP_NAME=notes-api
DOCKER_COMPOSE=docker compose

help:
	@echo "Available targets:"
	@echo "  run                 - Run the application locally"
	@echo "  build               - Build binary"
	@echo "  test                - Run all tests"
	@echo "  test-unit           - Run unit tests"
	@echo "  test-integration    - Run integration tests"
	@echo "  migrate-up          - Apply migrations"
	@echo "  migrate-down        - Rollback last migration"
	@echo "  docker-up           - Start all services with Docker Compose"
	@echo "  docker-down         - Stop all services"
	@echo "  lint                - Run golangci-lint"
	@echo "  tidy                - go mod tidy"

run:
	go run ./cmd/api

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/$(APP_NAME) ./cmd/api

test: test-unit test-integration

test-unit:
	go test -v -race -count=1 ./internal/...

test-integration:
	go test -v -race -count=1 -tags=integration ./tests/...

migrate-up:
	migrate -path migrations -database "$${DATABASE_URL:-postgres://notes:notes@localhost:5432/notes?sslmode=disable}" up

migrate-down:
	migrate -path migrations -database "$${DATABASE_URL:-postgres://notes:notes@localhost:5432/notes?sslmode=disable}" down 1

docker-up:
	$(DOCKER_COMPOSE) up --build -d

docker-down:
	$(DOCKER_COMPOSE) down -v

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
