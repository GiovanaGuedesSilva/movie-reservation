# ── Variables ─────────────────────────────────────────────────────────────────
BINARY     = bin/api
MAIN       = ./cmd/api
MODULE     = github.com/movie-reservation/api

MIGRATE    = migrate
MIGRATE_DB = "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"
MIGRATIONS = ./migrations

# Load .env for local commands (won't override already-set env vars)
-include .env
export

# ── Development ───────────────────────────────────────────────────────────────

.PHONY: run
run: ## Start the API server
	go run $(MAIN)/main.go

.PHONY: build
build: ## Compile the binary to ./bin/api
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY) $(MAIN)

.PHONY: install
install: ## Download Go module dependencies
	go mod download

.PHONY: tidy
tidy: ## Tidy and verify go.mod / go.sum
	go mod tidy
	go mod verify

# ── Testing ───────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run all tests
	go test -v -race ./...

.PHONY: test-coverage
test-coverage: ## Run tests and generate an HTML coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ── Database ──────────────────────────────────────────────────────────────────

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	$(MIGRATE) -path $(MIGRATIONS) -database $(MIGRATE_DB) up

.PHONY: migrate-down
migrate-down: ## Revert the last migration
	$(MIGRATE) -path $(MIGRATIONS) -database $(MIGRATE_DB) down 1

.PHONY: migrate-drop
migrate-drop: ## Drop everything (DANGER: irreversible)
	$(MIGRATE) -path $(MIGRATIONS) -database $(MIGRATE_DB) drop -f

.PHONY: migrate-create
migrate-create: ## Create a new migration: make migrate-create NAME=create_users
	@test -n "$(NAME)" || (echo "Usage: make migrate-create NAME=migration_name" && exit 1)
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS) -seq $(NAME)

.PHONY: seed
seed: ## Populate the database with initial data
	go run ./scripts/seed.go

# ── Docker ────────────────────────────────────────────────────────────────────

.PHONY: docker-up
docker-up: ## Build images and start all containers
	docker compose up --build -d
	@echo "API running at http://localhost:$(APP_PORT)"

.PHONY: docker-down
docker-down: ## Stop and remove all containers
	docker compose down

.PHONY: docker-logs
docker-logs: ## Tail logs from all containers
	docker compose logs -f

.PHONY: docker-db
docker-db: ## Start only the database container
	docker compose up -d postgres

# ── Code Quality ──────────────────────────────────────────────────────────────

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format all Go source files
	gofmt -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

# ── Utilities ─────────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/ coverage.out coverage.html

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
