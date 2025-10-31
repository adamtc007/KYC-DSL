# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

KYC-DSL is a **Domain Specific Language (DSL) processing system** for Know Your Customer (KYC) compliance cases. The system uses Rust for high-performance DSL parsing and execution, with Go providing data access, ontology management, and RAG capabilities.

**Current Version**: 2.0  
**Architecture**: Rust (computation) + Go (data) + Shared gRPC/Protobuf API  
**Key Features**: DSL parsing, regulatory ontology, ownership tracking, amendments, RAG vector search, feedback loop learning

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    CLI (kycctl)                         │
└─────────────────────────────────────────────────────────┘
                           │
                           │ gRPC
                           ▼
┌─────────────────────────────────────────────────────────┐
│          Rust DSL Service (Port 50060)                  │
│  - Parse DSL (nom-based S-expression parser)            │
│  - Validate DSL (grammar + semantics)                   │
│  - Execute Functions                                    │
│  - Serialize Cases                                      │
│  - Apply Amendments                                     │
└─────────────────────────────────────────────────────────┘
                           │
                           │ (persistence)
                           ▼
┌─────────────────────────────────────────────────────────┐
│          Go Data Service (Port 50070)                   │
│  - PostgreSQL Access (pgx/sqlx)                         │
│  - Dictionary Service (attributes/documents)            │
│  - Case Version Control (SHA-256 hashing)              │
│  - Ontology Repository (regulations/mappings)          │
│  - RAG/Vector Search (OpenAI + pgvector)               │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │  PostgreSQL  │
                    │  + pgvector  │
                    └──────────────┘
```

**Key Principle:** Rust owns computation, Go owns data.

## Project Structure

```
KYC-DSL/
├── Rust Stack (Computation Layer)
│   ├── kyc_dsl_core/        Core DSL engine library
│   │   ├── parser.rs        nom-based S-expression parser
│   │   ├── compiler.rs      AST → instruction compilation
│   │   └── executor.rs      Execution engine
│   └── kyc_dsl_service/     gRPC service (port 50060)
│       └── main.rs          DslService implementation
│
├── Go Stack (Data Layer)
│   ├── cmd/
│   │   ├── kycctl/          CLI tool (uses Rust service)
│   │   ├── kycserver/       REST API server (port 8080)
│   │   └── dataserver/      Data Service gRPC (port 50070)
│   └── internal/
│       ├── rustclient/      Rust gRPC client wrapper ★
│       ├── cli/             CLI command handlers
│       ├── amend/           Amendment system
│       ├── storage/         PostgreSQL operations (pgx/sqlx)
│       ├── dataservice/     Data service implementation
│       ├── ontology/        Ontology repository
│       ├── rag/             OpenAI embeddings + vector search
│       ├── model/           Data models
│       └── api/             REST API handlers
│
├── Shared Layer
│   ├── api/proto/           Protobuf definitions
│   │   ├── dsl_service.proto      Rust DSL service API
│   │   ├── kyc_case.proto         Case data structures
│   │   ├── rag_service.proto      RAG/search API
│   │   └── cbu_graph.proto        Business unit graphs
│   └── proto_shared/        Shared Go/Rust protos
│       └── data_service.proto     Dictionary + Case services
│
└── Database
    └── PostgreSQL            21 tables + pgvector extension
```

## Service Ports

| Port  | Service            | Purpose                      |
|-------|--------------------|------------------------------|
| 50060 | Rust DSL Service   | Parse, validate, execute DSL |
| 50070 | Go Data Service    | Database access, ontology    |
| 8080  | REST API           | HTTP gateway (optional)      |

## Common Development Commands

### Build & Run

```bash
# Build CLI
make build              # Creates ./kycctl binary
make install            # Install to GOPATH/bin

# Run Rust DSL Service
cd rust
cargo run -p kyc_dsl_service
# Listening on [::1]:50060

# Run Data Service
make run-dataserver     # Port 50070

# Process DSL files
./kycctl sample_case.dsl
./kycctl ontology_example.dsl
```

### Testing

```bash
# Run all tests
make test
make test-verbose

# Run specific test scripts
./scripts/test_semantic_search.sh
./scripts/test_feedback.sh
./test_ontology_validation.sh
```

### Maintenance

```bash
make deps               # Update dependencies
make fmt                # Format code
make lint               # Run linter
make clean              # Clean artifacts
go mod tidy             # Clean Go dependencies
```

### Database

```bash
# Initialize database
./scripts/init_ontology.sh

# Environment variables
export PGHOST=localhost
export PGPORT=5432
export PGUSER=youruser
export PGDATABASE=kyc_dsl
export OPENAI_API_KEY=sk-...  # Required for RAG
```

## CLI Commands

### DSL Processing

```bash
# Store grammar definition
./kycctl grammar

# Process DSL files
./kycctl sample_case.dsl
./kycctl ontology_example.dsl
./kycctl ownership_case.dsl

# Validate existing case
./kycctl validate CASE-NAME
```

### Amendments (Incremental Updates)

```bash
./kycctl amend CASE-NAME --step=policy-discovery
./kycctl amend CASE-NAME --step=document-solicitation
./kycctl amend CASE-NAME --step=document-discovery
./kycctl amend CASE-NAME --step=ownership-discovery
./kycctl amend CASE-NAME --step=risk-assessment
./kycctl amend CASE-NAME --step=approve
```

### RAG & Semantic Search

```bash
# Setup
export OPENAI_API_KEY="sk-..."
./kycctl seed-metadata

# Search
./kycctl search-metadata "tax compliance"
./kycctl similar-attributes UBO_NAME
./kycctl text-search "ownership"
./kycctl metadata-stats
```

### Ontology

```bash
./kycctl ontology           # Display ontology structure
```

## DSL Format

S-expression syntax with ontology-aware extensions:

```lisp
(kyc-case CASE-NAME
  (nature-purpose
    (nature "Investment Fund")
    (purpose "Global Equity Strategy"))
  
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
  
  (kyc-token "approved")
)
```

## Rust DSL Service API

The Rust service (port 50060) implements these gRPC RPCs:

```protobuf
service DslService {
  rpc Execute(ExecuteRequest) returns (ExecuteResponse);
  rpc Validate(ValidateRequest) returns (ValidationResult);
  rpc Parse(ParseRequest) returns (ParseResponse);
  rpc Serialize(SerializeRequest) returns (SerializeResponse);
  rpc Amend(AmendRequest) returns (AmendResponse);
  rpc ListAmendments(ListAmendmentsRequest) returns (ListAmendmentsResponse);
  rpc GetGrammar(GetGrammarRequest) returns (GrammarResponse);
}
```

### Using Rust Service from Go

```go
import "github.com/adamtc007/KYC-DSL/internal/rustclient"

// Connect to Rust DSL service
client, err := rustclient.NewDslClient("localhost:50060")
defer client.Close()

// Parse DSL
parseResp, err := client.ParseDSL(dslText)

// Validate
valResult, err := client.ValidateDSL(dslText)

// Execute
execResp, err := client.ExecuteCase(caseID, "process")

// Amend
amendResp, err := client.AmendCase(caseName, "policy-discovery")
```

## Database Schema

**PostgreSQL Database**: `kyc_dsl`  
**Tables**: 21 core tables + 3 views

**Key Tables:**
- `kyc_cases`, `case_versions`, `case_amendments` - Version control (