.PHONY: all build test lint generate clean dev proto

BINARY_NAME := barq-slides-server
BIN_DIR     := bin
CMD_DIR     := cmd/server

# Go toolchain
GO      := go
GOFLAGS :=

# Buf (protobuf)
BUF := buf

# Default target
all: lint test build

## build: compile the server binary
build:
	@echo "==> Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./$(CMD_DIR)/...

## run / dev: run server with live-reload via 'air' if available, else go run
dev:
	@if command -v air > /dev/null 2>&1; then \
		air; \
	else \
		$(GO) run ./$(CMD_DIR)/...; \
	fi

## test: run all unit and integration tests
test:
	@echo "==> Running tests..."
	$(GO) test ./... -v -race -count=1

## test-cover: run tests with coverage report
test-cover:
	@echo "==> Running tests with coverage..."
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: run golangci-lint
lint:
	@echo "==> Linting..."
	golangci-lint run ./...

## proto: generate Go stubs from proto definitions
proto:
	@echo "==> Generating proto stubs..."
	cd .. && $(BUF) generate

## generate: run go generate across all packages
generate:
	@echo "==> Running go generate..."
	$(GO) generate ./...

## tidy: tidy go modules
tidy:
	$(GO) mod tidy

## clean: remove build artifacts
clean:
	@echo "==> Cleaning..."
	rm -rf $(BIN_DIR) coverage.out coverage.html tmp/

## help: print this help
help:
	@echo "Available targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
