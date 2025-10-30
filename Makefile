# Makefile for KYC-DSL
# Builds with greenteagc garbage collector experiment

.PHONY: build build-server run run-server test clean lint fmt deps verify proto gateway run-grpc

# Build variables
GOEXPERIMENT := greenteagc
BUILD_DIR := bin
BINARY := kycctl
SERVER_BINARY := kycserver
GRPC_SERVER_BINARY := grpcserver
CMD_DIR := ./cmd/kycctl
SERVER_CMD_DIR := ./cmd/kycserver
GRPC_SERVER_DIR := ./cmd/server
PROTO_DIR := api/proto
PB_DIR := api/pb

# Default target - build all binaries
all: build build-server build-grpc

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

# Build the gRPC server binary
build-grpc: $(BUILD_DIR)/$(GRPC_SERVER_BINARY)

$(BUILD_DIR)/$(GRPC_SERVER_BINARY):
	@echo "Building $(GRPC_SERVER_BINARY) with GOEXPERIMENT=$(GOEXPERIMENT)..."
	@mkdir -p $(BUILD_DIR)
	GOEXPERIMENT=$(GOEXPERIMENT) go build -o $(BUILD_DIR)/$(GRPC_SERVER_BINARY) $(GRPC_SERVER_DIR)

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

# Run the gRPC server
run-grpc: build-grpc
	@echo "Starting gRPC server (port 50051) with REST gateway (port 8080)..."
	@echo "Make sure OPENAI_API_KEY is set and database is running"
	./$(BUILD_DIR)/$(GRPC_SERVER_BINARY)

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

# Generate protobuf code
proto:
	@echo "Generating protobuf Go code..."
	@mkdir -p $(PB_DIR)
	protoc --go_out=$(PB_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PB_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
	@echo "✓ Proto files generated in $(PB_DIR)"

# Generate gRPC gateway (optional - requires grpc-gateway)
gateway:
	@echo "Generating gRPC gateway code..."
	@mkdir -p $(PB_DIR)
	protoc --grpc-gateway_out=$(PB_DIR) --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=$(PB_DIR) \
		$(PROTO_DIR)/*.proto
	@echo "✓ Gateway files generated in $(PB_DIR)"

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
