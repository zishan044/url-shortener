.PHONY: help build run test test-coverage test-short lint clean migrate-up migrate-down migrate-create docker-up docker-down swagger fmt vet install-tools

# Variables
GO := go
GOFLAGS := -v
BINARY_NAME := api
BINARY_PATH := bin/$(BINARY_NAME)
MAIN_PACKAGE := ./cmd/api
MIGRATION_PATH := internal/database/migrations
DB_URL ?= postgres://postgres:postgres@localhost:5432/url_shortener?sslmode=disable

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install required tools
	@echo "Installing tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang-migrate/migrate/cmd/migrate@latest
	go install github.com/swaggo/swag/cmd/swag@v1.16.6

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_PATH) $(MAIN_PACKAGE)
	@echo "Build complete: $(BINARY_PATH)"

run: build ## Build and run the application
	@echo "Starting application..."
	$(BINARY_PATH)

run-dev: ## Run the application in development mode with hot reload (requires air)
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-short: ## Run tests with short timeout
	@echo "Running tests (short)..."
	$(GO) test -short ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	$(GO) test -v -race ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./... --timeout=5m

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...
	goimports -w .

vet: ## Run vet
	@echo "Running go vet..."
	$(GO) vet ./...

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	$(GO) run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go
	@echo "Swagger docs generated"

clean: ## Clean build artifacts
	@echo "Cleaning up..."
	rm -f $(BINARY_PATH)
	rm -f coverage.out coverage.html
	$(GO) clean

migrate: migrate-up ## Alias for migrate-up

migrate-up: ## Run migrations up
	@echo "Running migrations up..."
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" up
	@echo "Migrations completed"

migrate-down: ## Run migrations down
	@echo "Running migrations down..."
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" down
	@echo "Migrations reversed"

migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(NAME)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t url-shortener:latest .

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	docker-compose up -d
	@echo "Containers started"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down
	@echo "Containers stopped"

docker-logs: ## Show Docker container logs
	docker-compose logs -f

all: install-tools fmt vet lint test build ## Run all checks and build

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

check: fmt vet lint ## Run all checks (fmt, vet, lint)

.DEFAULT_GOAL := help
