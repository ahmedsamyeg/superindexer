# Usenet Indexer Makefile
# Provides common development workflows with testcontainers support

.PHONY: all build test test-integration lint vet fmt run run-dev docker-up docker-down docker-clean help

# Default target
all: build

# ─── Build ───────────────────────────────────────────────────────────────────

build: ## Build all binaries
	@echo "==> Building API, worker, and migrate binaries..."
	@cd backend && go build -o ../build/indexer-api ./cmd/api
	@cd backend && go build -o ../build/indexer-worker ./cmd/worker
	@cd backend && go build -o ../build/indexer-migrate ./cmd/migrate
	@echo "==> Done"

build-api: ## Build only the API binary
	@cd backend && go build -o ../build/indexer-api ./cmd/api

build-worker: ## Build only the worker binary
	@cd backend && go build -o ../build/indexer-worker ./cmd/worker

build-migrate: ## Build only the migrate binary
	@cd backend && go build -o ../build/indexer-migrate ./cmd/migrate

# ─── Test ────────────────────────────────────────────────────────────────────

test: ## Run all tests (unit + integration)
	@echo "==> Running all tests..."
	@cd backend && go test -v -cover ./...
	@echo "==> Done"

test-unit: ## Run only unit tests
	@echo "==> Running unit tests..."
	@cd backend && go test -v -cover ./... -short
	@echo "==> Done"

test-integration: ## Run integration tests with testcontainers
	@echo "==> Running integration tests with testcontainers..."
	@cd backend && go test -v -cover ./internal/database/integration/...
	@echo "==> Done"

test-migrations: ## Run migration reversibility tests
	@echo "==> Running migration reversibility tests..."
	@cd backend && go test -v ./internal/database/migrations/...
	@echo "==> Done"

test-all: test test-integration test-migrations ## Run all tests

# ─── Lint ────────────────────────────────────────────────────────────────────

lint: ## Run golint
	@echo "==> Running golint..."
	@cd backend && go vet ./...
	@echo "==> Done"

fmt: ## Format Go code
	@echo "==> Formatting Go code..."
	@cd backend && gofmt -w .
	@echo "==> Done"

# ─── Run ─────────────────────────────────────────────────────────────────────

run: ## Run the API server
	@cd backend && go run ./cmd/api

run-dev: ## Run all services via Docker Compose
	@echo "==> Starting development environment..."
	@cd deployments/docker && docker-compose up -d
	@echo "==> Development environment started"
	@echo "    API:      http://localhost:8080"
	@echo "    Frontend: http://localhost:3000"
	@echo "    Prometheus: http://localhost:9090"
	@echo "    Grafana:  http://localhost:3001"

run-migrate: ## Run database migrations
	@echo "==> Running migrations..."
	@cd backend && go run ./cmd/migrate up
	@echo "==> Migrations complete"

run-migrate-down: ## Rollback last migration
	@echo "==> Rolling back last migration..."
	@cd backend && go run ./cmd/migrate down 1
	@echo "==> Rollback complete"

# ─── Docker ──────────────────────────────────────────────────────────────────

docker-up: ## Start all services via Docker Compose
	@cd deployments/docker && docker-compose up -d --build
	@echo "==> Services started"

docker-down: ## Stop all services
	@cd deployments/docker && docker-compose down
	@echo "==> Services stopped"

docker-clean: ## Stop and remove all containers, volumes, images
	@cd deployments/docker && docker-compose down -v --rmi all
	@echo "==> Cleaned up"

# ─── Help ────────────────────────────────────────────────────────────────────

help: ## Show this help
	@echo "Usenet Indexer - Development Makefile"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
