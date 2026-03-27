APP_NAME := diet-api

.PHONY: run tidy fmt test

run:
	go run ./cmd/api

tidy:
	go mod tidy

fmt:
	go fmt ./...

test:
	go test ./...
