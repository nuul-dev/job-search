-include .env
export DATABASE_URL

# Pinned to the last goose release compatible with Go 1.24 (v3.27.0+ requires Go 1.25).
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.26.0

.DEFAULT_GOAL := help

.PHONY: ui
ui: ## Open the local vacancy inbox at http://127.0.0.1:8080
	cd backend/server && go run ./cmd

.PHONY: help db-up db-down db-reset migrate-up migrate-status

help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

db-up: ## Start postgres and adminer containers
	docker compose up -d postgres adminer

db-down: ## Stop and remove containers
	docker compose down

db-reset: ## Destroy volumes and restart postgres and adminer
	docker compose down -v && docker compose up -d postgres adminer

migrate-up: ## Apply all pending migrations
	$(GOOSE) -dir db/migrations postgres "$(DATABASE_URL)" up

migrate-status: ## Show migration status
	$(GOOSE) -dir db/migrations postgres "$(DATABASE_URL)" status
