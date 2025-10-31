# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

KYC-DSL is a **dual-codebase system** with both Go and Rust implementations sharing Protocol Buffer definitions. The system processes Domain Specific Language (DSL) files for Know Your Customer (KYC) compliance cases, with PostgreSQL persistence, full version control, and semantic search capabilities.

**Current Version**: 1.5  
**Architecture**: Go (primary) + Rust (high-performance alternative) + Shared gRPC/Protobuf API  
**Key Features**: DSL parsing, regulatory ontology, ownership tracking, amendments, RAG vector search, feedback loop learning

## Quick Architecture

```
KYC-DSL/
├── Go Stack (Primary - Production Ready)
│   ├── cmd/kycctl/          CLI tool
│   ├── cmd/kycserver/       REST API (port 8080)
│   ├── cmd/server/          gRPC server (port 50051)
│   ├── cmd/dataserver/      Data Service gRPC (port 50070) ★ NEW
│   └── internal/            Parser, storage, ontology, RAG
│       └── dataservice/     Centralized database layer ★ NEW
│
├── Rust Stack (High-Performance Alternative)
│   ├── kyc_dsl_core/        Core engine library
│   └── kyc_dsl_service/     gRPC service (port 50060)
│
├── Shared Layer
│   ├── api/proto/           Protobuf definitions
│   │   ├── dsl_service.proto
│   │   ├── kyc_case.proto
│   │   ├── rag_service.proto
│   │   └── cbu_graph.proto
│   └── proto_shared/        Shared Go/Rust protos ★ NEW
│       └── data_service.proto   Dictionary + Case services
│
└── Database
    └── PostgreSQL            Version control, ontology, embeddings
```
</thinking>

**Data Service:** Centralized gRPC service (port 50070) that owns all PostgreSQL 
connections and exposes Dictionary (attributes/documents) and Case (version control) 
APIs. Used by Go CLI, Rust DSL engine, and UI clients.

## Common Development Commands

### Go Stack (Primary)

**Build & Run:**
```bash
make build              # Build kycctl CLI
make run                # Run with sample case
make run-file FILE=x.dsl # Run specific DSL file
make install            # Install to GOPATH/bin
```

**Testing:**
```bash
make test               # All tests
make test-verbose       # Verbose output
make test-parser        # Parser tests only
./scripts/test_semantic_search.sh  # RAG tests
./scripts/test_feedback.sh         # Feedback system tests
```

**Maintenance:**
```bash
make deps               # Update dependencies
make fmt                # Format code
make lint               # Run linter
make clean              # Clean artifacts
```

**Data Service (Database Layer):**
```bash
make init-dataserver    # Initialize database schema
make build-dataserver   # Build Data Service
make run-dataserver     # Run Data Service (port 50070)
make proto-data         # Regenerate data service protos
./scripts/test_data_service.sh  # Integration tests
```

### Rust Stack (Alternative Engine)

**Build & Run:**
```bash
make rust-build         # Build Rust workspace
make rust-service       # Run gRPC service (port 50060)
make rust-test          # Run Rust tests
```

**Direct Commands:**
```bash
cd rust
cargo build             # Build workspace
cargo test              # Run tests
cargo run -p kyc_dsl_service  # Start service
./verify.sh             # Verify installation
```

### Database

**Setup:**
```bash
make docker-up          # Start PostgreSQL (if using Docker)
./scripts/init_ontology.sh  # Initialize schema + seed data
```

**Environment Variables:**
- `PGHOST` (default: localhost)
- `PGPORT` (default: 5432)
- `PGUSER` (default: current user)
- `PGDATABASE` (default: kyc_dsl)
- `OPENAI_API_KEY` (required for RAG features)

## Go Stack Architecture

### Core Components

**CLI & Servers:**
- `cmd/kycctl/` - Command-line interface
- `cmd/kycserver/` - REST API server (port 8080)
- `cmd/server/` - gRPC server (port 50051)

**Internal Packages:**
- `internal/cli/` - CLI command routing
- `internal/parser/` - S-expression DSL parser with validation
- `internal/engine/` - Case execution engine
- `internal/storage/` - PostgreSQL layer (sqlx)
- `internal/model/` - Data models (KycCase, attributes, documents)
- `internal/ontology/` - Regulatory ontology + metadata repository
- `internal/amend/` - Amendment system with predefined mutations
- `internal/rag/` - OpenAI embeddings & vector search
- `internal/lineage/` - Attribute derivation engine
- `internal/api/` - HTTP API handlers

### Data Flow

```
DSL File → Parse → Bind → Validate → Execute → Persist
              ↓      ↓       ↓         ↓        ↓
            AST   Model   Grammar   Engine   PostgreSQL
                         Ontology            (versioned)
```

### Key Operations

1. **Parse**: `parser.ParseFile()` → S-expression AST
2. **Bind**: AST → `model.KycCase` (typed structure)
3. **Validate**: Grammar + semantics + ontology + ownership rules
4. **Execute**: `engine.RunCase()` processes the case
5. **Persist**: `storage.InsertCase()` with SHA-256 versioning
6. **Amend**: `amend.ApplyAmendment()` incremental changes

## Rust Stack Architecture

### Core Components

**kyc_dsl_core (Library):**
- `parser.rs` - nom-based S-expression parser
- `compiler.rs` - AST → instruction compilation
- `executor.rs` - Stateful execution engine
- Pure Rust, type-safe, no unsafe code

**kyc_dsl_service (gRPC Server):**
- Wraps `kyc_dsl_core` library
- Implements `api/proto/dsl_service.proto`
- Compatible with Go gRPC clients
- Runs on port **50060**

### Rust API

```rust
// Compile DSL source to execution plan
pub fn compile_dsl(src: &str) -> Result<String, DslError>

// Execute compiled plan
pub fn execute_plan(plan_json: &str) -> Result<String, DslError>
```

## Shared Protocol Buffers

Both Go and Rust implement the same gRPC services defined in `api/proto/`:

**Services:**
- `DslService` - DSL parsing, validation, execution
- `RagService` - Semantic search and feedback
- `CbuGraphService` - Client Business Unit graph operations

**Message Types:**
- `KycCase` - Core case structure
- `ExecuteRequest/Response` - Function execution
- `SearchRequest/Response` - Vector search
- `FeedbackRequest` - Learning feedback

## CLI Commands Reference

### Grammar & Ontology

```bash
./kycctl grammar        # Store grammar definition
./kycctl ontology       # Display ontology structure
```

### Process Cases

```bash
./kycctl sample_case.dsl           # Process sample
./kycctl ontology_example.dsl      # Ontology-aware case
./kycctl ownership_case.dsl        # Ownership structures
```

### Amendments (Incremental Updates)

```bash
./kycctl amend CASE-NAME --step=policy-discovery
./kycctl amend CASE-NAME --step=document-solicitation
./kycctl amend CASE-NAME --step=document-discovery    # Auto-populate from ontology
./kycctl amend CASE-NAME --step=ownership-discovery
./kycctl amend CASE-NAME --step=risk-assessment
./kycctl amend CASE-NAME --step=approve
```

### RAG & Semantic Search

```bash
# Setup
export OPENAI_API_KEY="sk-..."
./kycctl seed-metadata              # Generate embeddings

# Search
./kycctl search-metadata "tax compliance"  # Semantic search
./kycctl similar-attributes UBO_NAME       # Find related attributes
./kycctl text-search "ownership"           # Keyword search
./kycctl metadata-stats                    # Repository statistics
```

## DSL Format

S-expression syntax with ontology-aware extensions:

```lisp
(kyc-case CASE-NAME
  (nature-purpose
    (nature "...")
    (purpose "..."))
  (client-business-unit CBU-NAME)
  (policy POLICY-CODE)
  (function ACTION)
  (obligation OBLIGATION-CODE)
  
  ; Ownership & Control
  (ownership-structure
    (entity ENTITY-NAME)
    (owner NAME PERCENT%)
    (beneficial-owner NAME PERCENT%)
    (controller NAME "ROLE"))
  
  ; Regulatory Ontology
  (data-dictionary
    (attribute ATTR-CODE
      (primary-source (document DOC-CODE))
      (secondary-source (document DOC-CODE))))
  
  (document-requirements
    (jurisdiction JURISDICTION)
    (required (document CODE "NAME")))
  
  (kyc-token "status")
)
```

## Database Schema

**PostgreSQL Database**: `kyc_dsl`  
**Tables**: 21 core tables + 3 views

**Key Tables:**
- `kyc_cases`, `case_versions`, `case_amendments` - Version control
- `kyc_regulations`, `kyc_documents`, `kyc_attributes` - Ontology (8 regs, 27 docs, 36 attrs)
- `kyc_attr_doc_links`, `kyc_doc_reg_links` - Relationships (51+ mappings)
- `kyc_attribute_metadata` - Embeddings with pgvector
- `rag_feedback` - Learning feedback system

**Extensions:**
- `pgvector` - 1536-dimensional embeddings (OpenAI text-embedding-3-large)

## Testing

### Go Tests

```bash
make test-parser        # Parser + validation
make test               # Full suite
make lint               # Code quality
```

### Rust Tests

```bash
cd rust
cargo test              # All Rust tests
cargo clippy            # Linter
./verify.sh             # Integration check
```

### Integration Tests

```bash
./scripts/test_semantic_search.sh
./scripts/test_feedback.sh
./test_ontology_validation.sh
```

## Key Files

**Documentation:**
- `README.md` - Project overview
- `RUST_QUICKSTART.md` - Rust 5-minute guide
- `RUST_MIGRATION_REPORT.md` - Architecture details
- `REGULATORY_ONTOLOGY.md` - Ontology documentation
- `RAG_VECTOR_SEARCH.md` - Semantic search guide
- `RAG_FEEDBACK.md` - Feedback loop system
- `AMENDMENT_SYSTEM.md` - Amendment workflows
- `OWNERSHIP_CONTROL.md` - Ownership validation

**Configuration:**
- `Makefile` - Build targets for both Go and Rust
- `go.mod` - Go dependencies
- `rust/Cargo.toml` - Rust workspace

**Examples:**
- `sample_case.dsl` - Basic case
- `ontology_example.dsl` - Full ontology-aware example
- `ownership_case.dsl` - Ownership structures
- `derived_attributes_example.dsl` - Attribute lineage

## Dependencies

**Go:**
- `github.com/alecthomas/participle/v2` - Parser generation
- `github.com/jmoiron/sqlx` - PostgreSQL extensions
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/sashabaranov/go-openai` - OpenAI embeddings
- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers

**Rust:**
- `nom` - Parser combinators
- `serde` - Serialization
- `tonic` - gRPC framework
- `tokio` - Async runtime
- `prost` - Protocol Buffers

## Current Features

**DSL Processing:**
- S-expression parsing with validation
- Grammar-based syntax checking
- Ontology reference validation
- Ownership structure validation (sum rules, controllers)

**Regulatory Ontology:**
- 8 regulations (FATCA, CRS, AMLD5/6, MAS626, etc.)
- 27 document types
- 36 attributes
- 51+ attribute-document mappings
- 18+ document-regulation links

**Version Control:**
- SHA-256 content hashing
- Full case history tracking
- Incremental amendment system
- Rollback capability

**RAG & Semantic Search:**
- OpenAI embeddings (text-embedding-3-large)
- pgvector similarity search
- Feedback loop learning
- Multi-agent feedback support

**APIs:**
- REST API (port 8080)
- gRPC Go service (port 50051)
- gRPC Rust service (port 50060)
- Shared protobuf definitions

## Port Allocation

- **8080** - Go REST API (`cmd/kycserver`)
- **50051** - Go gRPC service (`cmd/server`)
- **50060** - Rust gRPC service (`kyc_dsl_service`)
- **50070** - Data Service gRPC (`cmd/dataserver`) ★ NEW
- **5432** - PostgreSQL database

---

**Last Updated**: 2024  
**Version**: 1.5  
**Architecture**: Dual Go/Rust with Shared Protobuf