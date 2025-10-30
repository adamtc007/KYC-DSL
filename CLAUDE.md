# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

KYC-DSL is a Go-based domain-specific language (DSL) processor for Know Your Customer (KYC) compliance cases. The system parses DSL files containing KYC case definitions and persists them to a PostgreSQL database.

## Common Development Commands

### Building and Running
- `make build` - Build the main CLI tool with greenteagc GC experiment
- `make run` - Build and run with the sample case
- `make run-file FILE=<dsl-file>` - Build and run with a specific DSL file
- `make install` - Install binary to GOPATH/bin
- `make info` - Show build configuration

**Legacy commands (use Makefile instead):**
- `go build ./cmd/kycctl` - Build without greenteagc experiment
- `GOEXPERIMENT=greenteagc go build ./cmd/kycctl` - Manual build with experiment

### Testing
- `make test` - Run all tests with greenteagc experiment
- `make test-verbose` - Run all tests with verbose output
- `make test-parser` - Run parser tests specifically

**Legacy commands:**
- `go test ./...` - Run all tests without experiment
- `GOEXPERIMENT=greenteagc go test ./...` - Manual test with experiment

### Dependencies and Maintenance
- `make deps` - Download and tidy dependencies
- `make fmt` - Format all Go code
- `make lint` - Run golangci-lint
- `make clean` - Remove build artifacts

**Legacy commands:**
- `go mod tidy` - Clean up and sync dependencies
- `go mod download` - Download dependencies

## Architecture

### Core Components

The project follows standard Go project layout:

**`cmd/kycctl/`** - Main CLI application entry point that orchestrates the parsing and execution flow

**`internal/`** - Private application packages:
- `parser/` - DSL parsing using github.com/alecthomas/participle/v2
- `engine/` - Execution engine that processes parsed cases
- `storage/` - PostgreSQL database layer using sqlx
- `model/` - Data models and types (CaseStatus enum, KycCase struct)

**`pkg/`** - Public packages (currently contains placeholder directories for dictionary and policyindex)

### Data Flow

1. CLI tool reads DSL file via `parser.ParseFile()`
2. Parser converts DSL syntax to AST using participle
3. Engine executor processes the case via `RunCase()`
4. Storage layer persists to PostgreSQL via `InsertCase()`

### DSL Format

The DSL uses S-expression-like syntax:
```
(kyc-case CASE-NAME
  (nature-purpose ...)
  (client-business-unit ...)
  (function ...)
  (policy ...)
  (obligation ...)
  (kyc-token "status")
)
```

## Database Setup

The application requires PostgreSQL with the following environment variables:
- `PGHOST` (default: localhost)
- `PGPORT` (default: 5432)
- `PGUSER` (default: current user)
- `PGPASSWORD` (optional)
- `PGDATABASE` (default: kyc_dsl)

The storage layer automatically creates the `kyc_cases` table on first connection.

## Dependencies

- `github.com/alecthomas/participle/v2` - Grammar-based parser generation
- `github.com/jmoiron/sqlx` - PostgreSQL extensions for database/sql
- `github.com/lib/pq` - PostgreSQL driver