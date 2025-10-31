# KYC-DSL

**Know Your Customer Domain-Specific Language for Financial Compliance**

A high-performance DSL system for processing regulatory KYC cases with semantic search, version control, and ontology-aware validation.

## Architecture

**Rust-Powered Computation + Go Data Layer**

```
┌─────────────────────────────────────┐
│     CLI / REST API / Clients        │
└─────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│   Rust DSL Service (Port 50060)    │
│   • Parse DSL                       │
│   • Validate                        │
│   • Execute                         │
│   • Amend                           │
└─────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│   Go Data Service (Port 50070)     │
│   • PostgreSQL Access              │
│   • Ontology Repository            │
│   • RAG Vector Search              │
└─────────────────────────────────────┘
                 │
                 ▼
         ┌──────────────┐
         │  PostgreSQL  │
         │  + pgvector  │
         └──────────────┘
```

**Key Principle:** Rust owns computation, Go owns data.

## Quick Start

### Prerequisites

- Go 1.21+
- Rust 1.70+
- PostgreSQL 14+ with pgvector extension
- OpenAI API key (for RAG features)

### 1. Setup Database

```bash
# Start PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE kyc_dsl;

# Initialize schema and ontology
./scripts/init_ontology.sh
```

### 2. Start Services

```bash
# Terminal 1: Start Rust DSL Service
cd rust
cargo run -p kyc_dsl_service
# Listening on [::1]:50060

# Terminal 2: Start Go Data Service (optional)
go run cmd/dataserver/main.go
# Listening on localhost:50070

# Terminal 3: Use CLI
make build
./kycctl sample_case.dsl
```

### 3. Process DSL Files

```bash
# Parse and store case
./kycctl sample_case.dsl

# Validate case
./kycctl validate CASE-NAME

# Apply amendments
./kycctl amend CASE-NAME --step=policy-discovery
./kycctl amend CASE-NAME --step=document-discovery
./kycctl amend CASE-NAME --step=ownership-discovery
```

## DSL Format

S-expression based syntax:

```lisp
(kyc-case AVIVA-EU-EQUITY-FUND
  (nature-purpose
    (nature "Institutional Investment Fund")
    (purpose "Cross-border equity investment"))
  
  (client-business-unit AVIVA-INVESTORS)
  (policy AML-POLICY-001)
  (function ONBOARDING)
  (obligation CDD-STANDARD)
  
  (ownership-structure
    (entity AVIVA-INVESTORS-LTD)
    (beneficial-owner AVIVA-PLC 100.0%)
    (controller BOARD-OF-DIRECTORS "Governance"))
  
  (data-dictionary
    (attribute UBO_NAME
      (primary-source (document PASSPORT))
      (secondary-source (document UTILITY_BILL))))
  
  (document-requirements
    (jurisdiction UK)
    (required (document INCORPORATION_CERT "Certificate of Incorporation"))
    (required (document MEMORANDUM "Memorandum of Association")))
  
  (kyc-token "pending-review"))
```

## Core Features

### DSL Processing (Rust)
- **Parse**: nom-based S-expression parser
- **Validate**: Grammar + semantics + ontology checks
- **Execute**: Stateful case execution engine
- **Serialize**: Round-trip DSL generation

### Regulatory Ontology (Go + PostgreSQL)
- 8 regulations (FATCA, CRS, AMLD5/6, MAS626, etc.)
- 27 document types
- 36 attributes
- 50+ attribute-document mappings
- Jurisdiction-aware document requirements

### Version Control (PostgreSQL)
- SHA-256 content hashing
- Full case history tracking
- Incremental amendments
- Rollback capability

### RAG & Semantic Search (Go + OpenAI + pgvector)
- OpenAI embeddings (text-embedding-3-large, 1536d)
- Vector similarity search
- Feedback loop learning
- Multi-agent feedback support

## CLI Commands

### Core Operations
```bash
# Store grammar definition
./kycctl grammar

# Display ontology
./kycctl ontology

# Process DSL file
./kycctl <file>.dsl

# Validate case
./kycctl validate <case-name>
```

### Amendments
```bash
./kycctl amend <case> --step=policy-discovery
./kycctl amend <case> --step=document-solicitation
./kycctl amend <case> --step=document-discovery
./kycctl amend <case> --step=ownership-discovery
./kycctl amend <case> --step=risk-assessment
./kycctl amend <case> --step=approve
./kycctl amend <case> --step=decline
```

### RAG & Search
```bash
# Seed metadata with embeddings
./kycctl seed-metadata

# Semantic search
./kycctl search-metadata "tax residency"

# Find similar attributes
./kycctl similar-attributes UBO_NAME

# Keyword search
./kycctl text-search "ownership"

# Statistics
./kycctl metadata-stats
```

## Project Structure

```
KYC-DSL/
├── cmd/
│   ├── kycctl/          CLI tool
│   ├── dataserver/      Data Service gRPC server (port 50070)
│   └── kycserver/       REST API (port 8080)
│
├── internal/
│   ├── rustclient/      Rust gRPC client wrapper
│   ├── cli/             CLI command handlers
│   ├── amend/           Amendment system
│   ├── storage/         PostgreSQL operations
│   ├── ontology/        Regulatory ontology repository
│   ├── rag/             RAG & vector search
│   ├── dataservice/     Data service implementation
│   └── model/           Data models
│
├── rust/
│   ├── kyc_dsl_core/    Core DSL engine library
│   │   ├── parser.rs    nom-based S-expression parser
│   │   ├── compiler.rs  AST → instruction compiler
│   │   └── executor.rs  Stateful execution engine
│   └── kyc_dsl_service/ gRPC service (port 50060)
│
├── api/proto/           Protocol Buffer definitions
│   ├── dsl_service.proto
│   ├── kyc_case.proto
│   ├── rag_service.proto
│   └── cbu_graph.proto
│
├── proto_shared/        Shared Go/Rust protos
│   └── data_service.proto
│
└── scripts/             Utility scripts
```

## Environment Variables

```bash
# Rust DSL service
export RUST_DSL_SERVICE_ADDR="localhost:50060"  # Default

# PostgreSQL
export PGHOST="localhost"
export PGPORT="5432"
export PGUSER="postgres"
export PGDATABASE="kyc_dsl"

# OpenAI (for RAG)
export OPENAI_API_KEY="sk-..."
```

## Development

### Build
```bash
# Go CLI and services
make build

# Rust DSL service
cd rust && cargo build --release
```

### Test
```bash
# Go tests
make test

# Rust tests
cd rust && cargo test

# Integration tests
./scripts/test_semantic_search.sh
./scripts/test_feedback.sh
./test_ontology_validation.sh
```

### Clean
```bash
make clean
```

## Service Ports

| Port  | Service           | Purpose                      |
|-------|-------------------|------------------------------|
| 50060 | Rust DSL Service  | Parse, validate, execute DSL |
| 50070 | Go Data Service   | Database access, ontology    |
| 8080  | REST API          | HTTP gateway (optional)      |
| 5432  | PostgreSQL        | Database                     |

## API

### gRPC Services

**Rust DSL Service (port 50060):**
- `Parse` - Parse DSL text to structured format
- `Validate` - Validate DSL case
- `Execute` - Execute function on case
- `Amend` - Apply amendment
- `Serialize` - Convert case to DSL
- `GetGrammar` - Retrieve EBNF grammar
- `ListAmendments` - Available amendment types

**Go Data Service (port 50070):**
- `DictionaryService` - Attributes and documents
- `CaseService` - Version control operations
- `OntologyService` - Regulatory ontology queries

**Go RAG Service:**
- `AttributeSearch` - Semantic vector search
- `SimilarAttributes` - Find similar attributes
- `TextSearch` - Keyword search
- `SubmitFeedback` - Learning feedback
- `GetMetadataStats` - Repository statistics

## Database Schema

**PostgreSQL Database:** `kyc_dsl`  
**Extensions:** `pgvector`

**Key Tables:**
- `kyc_cases`, `case_versions`, `case_amendments` - Version control
- `kyc_regulations`, `kyc_documents`, `kyc_attributes` - Ontology
- `kyc_attr_doc_links`, `kyc_doc_reg_links` - Relationships
- `kyc_attribute_metadata` - Embeddings (1536d vectors)
- `rag_feedback` - Learning feedback

## Performance

**Rust DSL Parser:**
- Parse time: ~20-30ms (complex cases)
- Memory: ~20MB process RSS
- Throughput: ~500 cases/second

**RAG Search:**
- Vector similarity: <50ms (10 results)
- Keyword search: <10ms
- Index size: 36 attributes, 27 documents

## Documentation

- `README.md` - This file (overview + quick start)
- `CLAUDE.md` - AI assistant context
- `RUST_QUICKSTART.md` - Rust service 5-minute guide
- `DATA_SERVICE_GUIDE.md` - Data service documentation
- `REGULATORY_ONTOLOGY.md` - Ontology structure
- `RAG_VECTOR_SEARCH.md` - Semantic search guide
- `RAG_FEEDBACK.md` - Feedback loop system

## Examples

See DSL examples in project root:
- `sample_case.dsl` - Basic case structure
- `ontology_example.dsl` - Full ontology-aware example
- `ownership_case.dsl` - Complex ownership structures
- `derived_attributes_example.dsl` - Attribute lineage

## Technology Stack

**Computation Layer (Rust):**
- nom (parser combinators)
- tonic (gRPC)
- tokio (async runtime)
- serde (serialization)
- prost (Protocol Buffers)

**Data Layer (Go):**
- pgx/v5 (PostgreSQL driver)
- sqlx (SQL extensions)
- go-openai (OpenAI embeddings)
- grpc (gRPC framework)
- protobuf (Protocol Buffers)

**Database:**
- PostgreSQL 14+
- pgvector extension

## License

Proprietary - Internal use only

## Version

**Current:** 1.5  
**Architecture:** Rust DSL Service + Go Data Layer  
**Status:** Production Ready