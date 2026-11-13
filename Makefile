SHELL := /bin/sh

APP_NAME := flight-api
BIN_DIR := bin
CMD_PATH := ./internal/cmd/rest
GO ?= go

.PHONY: help run build test vet fmt tidy clean openapi

help:
	@printf "Flight Search API\n\n"
	@printf "Usage:\n"
	@printf "  make run       Start the API server\n"
	@printf "  make build     Build the API binary into $(BIN_DIR)/$(APP_NAME)\n"
	@printf "  make test      Run all Go tests\n"
	@printf "  make vet       Run go vet\n"
	@printf "  make fmt       Format all Go files\n"
	@printf "  make tidy      Run go mod tidy\n"
	@printf "  make clean     Remove build artifacts\n"
	@printf "  make openapi   Show the OpenAPI specification path\n"

run:
	$(GO) run $(CMD_PATH)

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)

openapi:
	@printf "OpenAPI specification: docs/openapi.yaml\n"
