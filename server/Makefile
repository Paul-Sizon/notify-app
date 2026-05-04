SHELL := /bin/bash
ENV := set -a; . ./.env; set +a

.PHONY: up down migrate migrate-test test test-integration test-e2e run tidy

up:
	docker compose up -d postgres
	@until docker compose exec -T postgres pg_isready -U notify >/dev/null 2>&1; do sleep 1; done
	@echo "postgres ready"

down:
	docker compose down

migrate:
	@$(ENV); goose -dir migrations postgres "$$DATABASE_URL" up

migrate-test:
	@$(ENV); goose -dir migrations postgres "$$TEST_DATABASE_URL" up

test:
	go test ./...

test-integration:
	@$(ENV); go test -tags=integration -count=1 ./...

test-e2e:
	@$(ENV); go test -tags=e2e -count=1 -timeout 5m ./e2e/...

test-all:
	@$(ENV); go test -count=1 ./...
	@$(ENV); go test -tags=integration -count=1 ./...
	@$(ENV); go test -tags=e2e -count=1 -timeout 5m ./e2e/...

run:
	@$(ENV); go run ./cmd/server

tidy:
	go mod tidy
