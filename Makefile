# Makefile for KYC-DSL
# Builds with greenteagc garbage collector experiment

.PHONY: build build-server build-client build-dataserver run run-server run-client run-dataserver test clean lint fmt deps verify proto proto-data gateway run-grpc init-dataserver rust-build rust-test run-rust rust-clean rust-fmt rust-lint rust-clippy rust-verify lint-all fmt-all

# Build variables
GOEXPERIMENT := greenteagc
BUILD_DIR := bin
BINARY := kycctl
SERVER_BINARY := kycserver
GRPC_SERVER_BINARY := grpcserver
DATA_SERVER_BINARY := dataserver
CLIENT_BINARY := kycclient
CMD_DIR := ./cmd/kycctl
SERVER_CMD_DIR := ./cmd/kycserver
GRPC_SERVER_DIR := ./cmd/server
DATA_SERVER_DIR := ./cmd/dataserver
CLIENT_CMD_DIR := ./cmd/client
PROTO_DIR := api/proto
PROTO_SHARED_DIR := proto_shared
PB_DIR := api/pb

# Default target - build all binaries
all: build build-server build-grpc build-dataserver build-client

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

# Build the Gio client binary
build-client: $(BUILD_DIR)/$(CLIENT_BINARY)

$(BUILD_DIR)/$(CLIENT_BINARY):
	@echo "Building $(CLIENT_BINARY) with GOEXPERIMENT=$(GOEXPERIMENT)..."
	@mkdir -p $(BUILD_DIR)
	GOEXPERIMENT=$(GOEXPERIMENT) go build -o $(BUILD_DIR)/$(CLIENT_BINARY) $(CLIENT_CMD_DIR)

# Build the Data Service gRPC server binary
build-dataserver: $(BUILD_DIR)/$(DATA_SERVER_BINARY)

$(BUILD_DIR)/$(DATA_SERVER_BINARY):
	@echo "Building $(DATA_SERVER_BINARY) with GOEXPERIMENT=$(GOEXPERIMENT)..."
	@mkdir -p $(BUILD_DIR)
	GOEXPERIMENT=$(GOEXPERIMENT) go build -o $(BUILD_DIR)/$(DATA_SERVER_BINARY) $(DATA_SERVER_DIR)

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

# Run the Gio client
run-client: build-client
	@echo "Starting Gio CBU Graph Viewer..."
	@echo "Make sure gRPC server is running on localhost:50051"
	@echo "Set GRPC_SERVER and CBU_ID env vars to customize"
	./$(BUILD_DIR)/$(CLIENT_BINARY)

# Run the Data Service gRPC server
run-dataserver: build-dataserver
	@echo "Starting Data Service gRPC server (port 50070)..."
	@echo "Make sure database is running and initialized"
	@echo "Run 'make init-dataserver' to initialize the database schema"
	./$(BUILD_DIR)/$(DATA_SERVER_BINARY)

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

# Generate Data Service protobuf code
proto-data:
	@echo "Generating Data Service protobuf Go code..."
	@mkdir -p $(PB_DIR)/kycdata
	protoc --go_out=. --go-grpc_out=. \
		--go_opt=module=github.com/adamtc007/KYC-DSL \
		--go-grpc_opt=module=github.com/adamtc007/KYC-DSL \
		$(PROTO_SHARED_DIR)/data_service.proto
	@echo "✓ Data Service proto files generated in $(PB_DIR)/kycdata"

# Initialize Data Service database schema
init-dataserver:
	@echo "Initializing Data Service database schema..."
	@chmod +x scripts/init_data_service.sh
	@./scripts/init_data_service.sh

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
	@echo "Data Server binary: $(DATA_SERVER_BINARY)"
	@echo "gRPC Server binary: $(GRPC_SERVER_BINARY)"
	@echo "Client binary: $(CLIENT_BINARY)"

# Run comprehensive verification checks
verify:
	@./verify.sh

# Rust targets
# ============

# Build the Rust workspace (core library + gRPC service)
rust-build:
	@echo "Building Rust workspace..."
	cd rust && cargo build --release

# Run Rust tests
rust-test:
	@echo "Running Rust tests..."
	cd rust && cargo test

# Run the Rust gRPC service (listens on port 50060)
run-rust:
	@echo "Starting Rust DSL gRPC service on port 50060..."
	@echo "Make sure protobuf definitions are up to date"
	cd rust/kyc_dsl_service && cargo run

# Format Rust code
rust-fmt:
	@echo "Formatting Rust code..."
	cd rust && cargo fmt

# Run Rust linter (clippy)
rust-lint: rust-clippy

rust-clippy:
	@echo "Running Rust clippy linter..."
	cd rust && cargo clippy -- -D warnings

# Clean Rust build artifacts
rust-clean:
	@echo "Cleaning Rust build artifacts..."
	cd rust && cargo clean

# Run Rust verification script
rust-verify:
	@echo "Running Rust verification checks..."
	@chmod +x rust/verify.sh
	cd rust && ./verify.sh

# Build everything (Go + Rust)
all-with-rust: all rust-build
	@echo "✓ All components built (Go + Rust)"

# Combined targets (Go + Rust)
# ============================

# Format all code (Go + Rust)
fmt-all: fmt rust-fmt
	@echo "✓ All code formatted (Go + Rust)"

# Lint all code (Go + Rust)
lint-all: lint rust-clippy
	@echo "✓ All linters passed (Go + Rust)"

# Verify everything (Go + Rust)
verify-all: verify rust-verify
	@echo "✓ Full verification complete (Go + Rust)"
