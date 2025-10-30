# Makefile for KYC-DSL
# Builds with greenteagc garbage collector experiment

.PHONY: build build-server run run-server test clean lint fmt deps verify

# Build variables
GOEXPERIMENT := greenteagc
BUILD_DIR := bin
BINARY := kycctl
SERVER_BINARY := kycserver
CMD_DIR := ./cmd/kycctl
SERVER_CMD_DIR := ./cmd/kycserver

# Default target - build both CLI and server
all: build build-server

# Build CLI
build: $(BUILD_DIR)/$(BINARY)

# Build the main binary with greenteagc
$(BUILD_DIR)/$(BINARY):
	@echo "Building $(BINARY) with GOEXPERIMENT=$(GOEXPERIMENT)..."
	@mkdir -p $(BUILD_DIR)
	GOEXPERIMENT=$(GOEXPERIMENT) go build -o $(BUILD_DIR)/$(BINARY) $(CMD_DIR)

# Build the server binary with greenteagc
build-server: $(BUILD_DIR)/$(SERVER_BINARY)

$(BUILD_DIR)/$(SERVER_BINARY):
	@echo "Building $(SERVER_BINARY) with GOEXPERIMENT=$(GOEXPERIMENT)..."
	@mkdir -p $(BUILD_DIR)
	GOEXPERIMENT=$(GOEXPERIMENT) go build -o $(BUILD_DIR)/$(SERVER_BINARY) $(SERVER_CMD_DIR)

# Run with sample case
run: build
	@echo "Running $(BINARY) with sample case..."
	./$(BUILD_DIR)/$(BINARY) sample_case.dsl

# Run with custom DSL file
run-file: build
	@if [ -z "$(FILE)" ]; then echo "Usage: make run-file FILE=<dsl-file>"; exit 1; fi
	./$(BUILD_DIR)/$(BINARY) $(FILE)

# Run the RAG API server
run-server: build-server
	@echo "Starting RAG API server..."
	@echo "Make sure OPENAI_API_KEY is set and database is running"
	./$(BUILD_DIR)/$(SERVER_BINARY)

# Run all tests with greenteagc (exclude examples)
test:
	@echo "Running tests with GOEXPERIMENT=$(GOEXPERIMENT)..."
	GOEXPERIMENT=$(GOEXPERIMENT) go test ./internal/... ./cmd/...

# Run tests with verbose output (exclude examples)
test-verbose:
	@echo "Running tests with verbose output..."
	GOEXPERIMENT=$(GOEXPERIMENT) go test -v ./internal/... ./cmd/...

# Run parser tests specifically
test-parser:
	@echo "Running parser tests..."
	GOEXPERIMENT=$(GOEXPERIMENT) go test -v ./internal/parser

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)

# Download dependencies
deps:
	go mod download
	go mod tidy

# Install binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY) to GOPATH/bin..."
	GOEXPERIMENT=$(GOEXPERIMENT) go install $(CMD_DIR)

# Show build info
info:
	@echo "Go version: $(shell go version)"
	@echo "GOEXPERIMENT: $(GOEXPERIMENT)"
	@echo "Build directory: $(BUILD_DIR)"
	@echo "CLI binary: $(BINARY)"
	@echo "Server binary: $(SERVER_BINARY)"

# Run comprehensive verification checks
verify:
	@./verify.sh
