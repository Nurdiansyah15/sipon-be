.PHONY: dev-up dev-all-up dev-down run build migrate-up migrate-down migrate-fresh migrate-version migrate-force migrate-create seed-all seed-role seed-user tidy test lint

COMPOSE_FILE := docker-compose.dev.yml

# Development
dev-up:
	docker compose -f $(COMPOSE_FILE) up -d postgres redis

dev-all-up:
	docker compose -f $(COMPOSE_FILE) up -d --build

dev-down:
	docker compose -f $(COMPOSE_FILE) down

run:
	docker compose -f $(COMPOSE_FILE) up --build app

# Build
build:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate
	go build -o bin/seeder ./cmd/seeder

tidy:
	go mod tidy

# Database Migration
migrate-up:
	docker compose -f $(COMPOSE_FILE) --profile tooling run --rm migrate up

migrate-down:
	docker compose -f $(COMPOSE_FILE) --profile tooling run --rm migrate down

migrate-fresh:
	docker compose -f $(COMPOSE_FILE) --profile tooling run --rm migrate fresh

migrate-version:
	docker compose -f $(COMPOSE_FILE) --profile tooling run --rm migrate version

migrate-force:
	@if [ -z "$(VERSION)" ]; then echo "VERSION=... diperlukan"; exit 1; fi
	docker compose -f $(COMPOSE_FILE) --profile tooling run --rm migrate force $(VERSION)

migrate-create:
	@if [ -z "$(NAME)" ]; then echo "NAME=... diperlukan"; exit 1; fi
	docker compose -f $(COMPOSE_FILE) --profile tooling run --rm devtools sh -c "migrate create -ext sql -dir migrations -seq $(NAME)"

# Seeding
seed-all:
	docker compose -f $(COMPOSE_FILE) --profile tooling run --rm seeder all

seed-role:
	docker compose -f $(COMPOSE_FILE) --profile tooling run --rm seeder role

seed-user:
	docker compose -f $(COMPOSE_FILE) --profile tooling run --rm seeder user

# Testing
test:
	go test ./...

test-unit:
	go test ./internal/...

test-integration:
	go test -tags=integration ./...

# Linting
lint:
	golangci-lint run ./...
