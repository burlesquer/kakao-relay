.PHONY: help up down docker-up docker-down docker-logs docker-clean db-shell db-migrate db-reset dev build

.DEFAULT_GOAL := help

# Colors
BLUE := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RESET := \033[0m

DATABASE_URL ?= postgresql://postgres:postgres@localhost:5433/kakao_relay

##@ Help

help: ## Show this help message
	@echo "$(GREEN)카카오톡 채널 릴레이 서버$(RESET) - Available commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(BLUE)%-15s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Docker Commands

up: docker-up ## Start all services (alias)

down: docker-down ## Stop all services (alias)

docker-up: ## Start PostgreSQL and Redis containers
	docker compose up -d --force-recreate
	@echo "$(GREEN)Services started. Waiting for health checks...$(RESET)"
	@sleep 2
	@docker compose ps

docker-down: ## Stop Docker containers
	docker compose down

docker-logs: ## Follow Docker logs
	docker compose logs -f

docker-clean: ## Stop and remove Docker volumes
	docker compose down -v
	@echo "$(YELLOW)Docker volumes removed$(RESET)"

##@ Database Commands

db-shell: ## Open PostgreSQL shell
	docker compose exec postgres psql -U $${POSTGRES_USER:-postgres} -d $${POSTGRES_DB:-kakao_relay}

db-migrate: ## Run database migrations
	@for f in drizzle/migrations/*.sql; do \
		echo "Applying $$f"; \
		psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$$f"; \
	done

db-reset: ## Reset database (drop and recreate)
	docker compose exec postgres psql -U $${POSTGRES_USER:-postgres} -c "DROP DATABASE IF EXISTS $${POSTGRES_DB:-kakao_relay};"
	docker compose exec postgres psql -U $${POSTGRES_USER:-postgres} -c "CREATE DATABASE $${POSTGRES_DB:-kakao_relay};"
	@echo "$(GREEN)Database reset. Run 'make db-migrate' to apply migrations.$(RESET)"

##@ Development Commands

dev: ## Start development server with hot reload
	go run ./cmd/server

build: ## Build for production
	go build ./cmd/server

##@ Quick Start

setup: docker-up db-migrate ## Start PostgreSQL and run migrations
	@echo "$(GREEN)Setup complete! Run 'make dev' to start the server.$(RESET)"

clean: docker-clean ## Clean up all Docker resources
	@echo "$(GREEN)Cleanup complete!$(RESET)"
