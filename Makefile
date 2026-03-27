APP_NAME := diet-api
MIGRATIONS_DIR := migrations

.PHONY: run build test fmt tidy migrate-up migrate-down

run:
	go run ./cmd/api

build:
	go build -o bin/$(APP_NAME) ./cmd/api

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
