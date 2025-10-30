# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

KYC-DSL is a Go-based domain-specific language (DSL) processor for Know Your Customer (KYC) compliance cases. The system parses DSL files containing KYC case definitions and persists them to a PostgreSQL database with full version control and amendment tracking.

**Version**: 1.5  
**Key Features**: DSL parsing, regulatory ontology, ownership tracking, incremental amendments, version control, RAG & vector search, feedback loop learning

## Common Development Commands

### Building and Running
- `make build` - Build the main CLI tool with greenteagc GC experiment
- `make run` - Build and run with the sample case
- `make run-file FILE=<dsl-file>` - Build and run with a specific DSL file
- `make install` - Install binary to GOPATH/bin
- `make info` - Show build configuration

### Testing
- `make test` - Run all tests with greenteagc experiment
- `make test-verbose` - Run all tests with verbose output
- `make test-parser` - Run parser tests specifically
- `./scripts/test_semantic_search.sh` - Test RAG & vector search functionality
- `./scripts/test_feedback.sh` - Test RAG feedback loop system

### Dependencies and Maintenance
- `make deps` - Download and tidy dependencies
- `make fmt` - Format all Go code
- `make lint` - Run golangci-lint
- `make clean` - Remove build artifacts

## Architecture

### Core Components

**`cmd/kycctl/`** - Main CLI application entry point (delegates to internal/cli)

**`internal/`** - Private application packages:
- `cli/` - CLI routing and command handlers
- `parser/` - DSL parsing, binding, serialization, validation
- `engine/` - Execution engine that processes parsed cases
- `storage/` - PostgreSQL database layer using sqlx
- `model/` - Data models (KycCase, AttributeSource, DocumentRequirement, AttributeMetadata, etc.)
- `ontology/` - Regulatory data ontology (regulations, documents, attributes) + metadata repository + feedback repository
- `amend/` - Amendment system with predefined mutations
- `token/` - KYC token management
- `rag/` - RAG embeddings and vector search (OpenAI integration)
- `lineage/` - Attribute lineage and derivation engine
- `api/` - HTTP API handlers for RAG search and feedback endpoints

### Data Flow

1. **Parse**: CLI reads DSL file → `parser.ParseFile()` → AST
2. **Bind**: AST → `parser.Bind()` → typed `model.KycCase`
3. **Validate**: `parser.ValidateDSL()` checks grammar + semantics + ownership rules
4. **Execute**: `engine.RunCase()` processes the case
5. **Persist**: `storage.InsertCase()` saves to PostgreSQL with SHA-256 hash
6. **Amend**: `amend.ApplyAmendment()` applies incremental changes with versioning

### DSL Format (v1.2)

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
  
  ; v1.2: Ownership & Control
  (ownership-structure
    (entity ENTITY-NAME)
    (owner NAME PERCENT%)
    (beneficial-owner NAME PERCENT%)
    (controller NAME "ROLE"))
  
  ; v1.2: Regulatory Ontology
  (data-dictionary
    (attribute ATTR-CODE
      (primary-source (document DOC-CODE))
      (secondary-source (document DOC-CODE))
      (tertiary-source "TEXT")))
  
  (document-requirements
    (jurisdiction JURISDICTION)
    (required
      (document CODE "NAME")))
  
  (kyc-token "status")
)
```

## Database Setup

### Environment Variables
- `PGHOST` (default: localhost)
- `PGPORT` (default: 5432)
- `PGUSER` (default: current user)
- `PGPASSWORD` (optional)
- `PGDATABASE` (default: kyc_dsl)

### Initialize Ontology
```bash
./scripts/init_ontology.sh
```

This creates:
- `kyc_cases`, `case_versions`, `case_amendments`, `grammar_versions`, `policy_registry`
- `kyc_regulations`, `kyc_documents`, `kyc_attributes`, `kyc_attr_doc_links`, `kyc_doc_reg_links`

## CLI Commands & Call Trees

### 1. Grammar Command
```bash
./kycctl grammar
```

**Call Tree:**
```
cli.Run(["grammar"])
└── cli.RunGrammarCommand()
    ├── storage.ConnectPostgres()
    ├── parser.CurrentGrammarEBNF()
    └── storage.InsertGrammar()
```

**Purpose**: Store current DSL grammar definition in database for validation reference.

---

### 2. Ontology Command
```bash
./kycctl ontology
```

**Call Tree:**
```
cli.Run(["ontology"])
└── cli.RunOntologyCommand()
    ├── storage.ConnectPostgres()
    ├── ontology.NewRepository()
    └── repo.DebugPrintOntologySummary()
        ├── repo.ListRegulations()
        └── repo.ListDocumentsByRegulation()
```

**Purpose**: Display regulatory data ontology structure (regulations → documents).

---

### 3. Process DSL File
```bash
./kycctl sample_case.dsl
./kycctl ontology_example.dsl
```

**Call Tree:**
```
cli.Run(["sample_case.dsl"])
└── cli.RunProcessCommand("sample_case.dsl")
    ├── parser.ParseFile()           # Read & tokenize DSL
    ├── parser.Bind()                # AST → model.KycCase
    ├── storage.ConnectPostgres()
    ├── storage.GetGrammar()
    ├── parser.ValidateDSL()         # Grammar + semantics + ownership
    ├── parser.SerializeCases()      # Model → DSL text
    ├── cli.displayCaseInfo()
    └── engine.NewExecutor().RunCase()
        └── storage.InsertCase()     # Persist with hash
```

**Purpose**: Parse, validate, and persist a DSL case to database.

---

### 4. Amendment Commands
```bash
./kycctl amend CASE-NAME --step=STEP-NAME
```

**Available Steps:**
- `policy-discovery` - Add policy discovery function and policies
- `document-solicitation` - Add document solicitation and obligations
- `document-discovery` - Auto-populate documents from ontology (ontology-aware)
- `ownership-discovery` - Add ownership structure and control hierarchy
- `risk-assessment` - Add risk assessment function
- `regulator-notify` - Add regulator notification
- `approve` - Finalize case as approved
- `decline` - Finalize case as declined
- `review` - Set case to review status

**Call Tree (Standard Amendment):**
```
cli.Run(["amend", "CASE-NAME", "--step=policy-discovery"])
└── cli.RunAmendCommand("CASE-NAME", "policy-discovery")
    ├── storage.ConnectPostgres()
    ├── amend.AddPolicyDiscovery        # Mutation function
    └── amend.ApplyAmendment()
        ├── storage.GetLatestCase()     # Load current version
        ├── parser.Bind()               # DSL → model
        ├── mutation(kycCase)           # Apply changes
        ├── parser.SerializeCases()     # Model → DSL
        ├── storage.InsertCase()        # New version
        └── storage.InsertAmendment()   # Record amendment
```

**Call Tree (Ontology-Aware Amendment):**
```
cli.Run(["amend", "CASE-NAME", "--step=document-discovery"])
└── cli.RunAmendCommand("CASE-NAME", "document-discovery")
    ├── storage.ConnectPostgres()
    ├── ontology.NewRepository()
    └── amend.ApplyAmendment()
        ├── storage.GetLatestCase()
        ├── parser.Bind()
        ├── amend.AddDocumentDiscovery(case, repo)
        │   ├── repo.ListDocumentsByRegulation("AMLD5")
        │   ├── repo.GetDocumentSources("UBO_NAME")
        │   └── Populates DataDictionary & DocumentRequirements
        ├── parser.SerializeCases()
        ├── storage.InsertCase()
        └── storage.InsertAmendment()
```

**Purpose**: Apply incremental changes to existing cases with full version control.

---

### 5. Seed Metadata with Embeddings
```bash
./kycctl seed-metadata
```

**Call Tree:**
```
cli.Run(["seed-metadata"])
└── cli.RunSeedMetadataCommand()
    ├── storage.ConnectPostgres()
    ├── ontology.NewMetadataRepo()
    ├── rag.NewEmbedder()
    └── For each sample attribute:
        ├── embedder.GenerateEmbedding()    # OpenAI API call
        └── repo.UpsertMetadata()           # Store with embedding
```

**Purpose**: Generate vector embeddings for all attributes using OpenAI's text-embedding-3-large model.

---

### 6. Semantic Search
```bash
./kycctl search-metadata "tax reporting requirements"
./kycctl search-metadata "beneficial ownership" --limit=5
```

**Call Tree:**
```
cli.Run(["search-metadata", "query"])
└── cli.RunSearchMetadataCommand("query", limit)
    ├── storage.ConnectPostgres()
    ├── ontology.NewMetadataRepo()
    ├── rag.NewEmbedder()
    ├── embedder.GenerateEmbeddingFromText()  # Query embedding
    └── repo.SearchByVector()                 # Vector similarity search
```

**Purpose**: Perform semantic search on attribute metadata using vector embeddings.

---

### 7. Find Similar Attributes
```bash
./kycctl similar-attributes UBO_NAME
./kycctl similar-attributes SANCTIONS_SCREENING_STATUS --limit=5
```

**Call Tree:**
```
cli.Run(["similar-attributes", "ATTR_CODE"])
└── cli.RunSimilarAttributesCommand("ATTR_CODE", limit)
    ├── storage.ConnectPostgres()
    ├── ontology.NewMetadataRepo()
    ├── repo.GetMetadata()                    # Get source attribute
    └── repo.FindSimilarAttributes()          # Vector similarity
```

**Purpose**: Find attributes semantically related to a given attribute code.

---

### 8. Text Search
```bash
./kycctl text-search "ownership"
./kycctl text-search "PEP"
```

**Call Tree:**
```
cli.Run(["text-search", "term"])
└── cli.RunTextSearchCommand("term")
    ├── storage.ConnectPostgres()
    ├── ontology.NewMetadataRepo()
    └── repo.SearchByText()                   # Traditional text search
```

**Purpose**: Search attributes by keyword, synonym, or business context (no embedding required).

---

### 9. Metadata Statistics
```bash
./kycctl metadata-stats
```

**Call Tree:**
```
cli.Run(["metadata-stats"])
└── cli.RunMetadataStatsCommand()
    ├── storage.ConnectPostgres()
    ├── ontology.NewMetadataRepo()
    └── repo.GetMetadataStats()
        ├── repo.CountMetadata()
        ├── repo.CountEmbeddings()
        └── SQL: Risk distribution query
```

**Purpose**: Display repository health, embedding coverage, and risk distribution.

---

## Test Invocations

### Parser Tests
```bash
make test-parser
# or
GOEXPERIMENT=greenteagc go test ./internal/parser -v
```

**Test Coverage:**
- `TestParseSimpleCase` - Basic DSL parsing
- `TestParseMultipleCases` - Multiple cases in one file
- `TestBindCase` - AST to model binding
- `TestSerializeCase` - Model to DSL round-trip
- `TestRoundTrip` - Full parse → bind → serialize → parse cycle
- `TestOwnershipValidation` - Ownership sum and controller rules
- `TestQuotedStrings` - Quoted string handling

### Full Test Suite
```bash
make test
# or
GOEXPERIMENT=greenteagc go test ./...
```

### Linting
```bash
make lint
# or
golangci-lint run
```

---

## Sample Test Cases

### Basic Case
```bash
./kycctl sample_case.dsl
```

### Ownership Cases
```bash
./kycctl ownership_case.dsl
./kycctl test_ownership.dsl
./kycctl test_valid_multi_owner.dsl
```

### Ontology-Aware Case
```bash
./kycctl ontology_example.dsl
```

### Amendment Workflow
```bash
# 1. Process initial case
./kycctl sample_case.dsl

# 2. Add policies
./kycctl amend AVIVA-EU-EQUITY-FUND --step=policy-discovery

# 3. Auto-discover documents (uses ontology)
./kycctl amend AVIVA-EU-EQUITY-FUND --step=document-discovery

# 4. Add ownership
./kycctl amend AVIVA-EU-EQUITY-FUND --step=ownership-discovery

# 5. Approve
./kycctl amend AVIVA-EU-EQUITY-FUND --step=approve
```

---

## Key Files

### Documentation
- `README.md` - Project overview and getting started
- `REGULATORY_ONTOLOGY.md` - Comprehensive ontology documentation
- `AMENDMENT_SYSTEM.md` - Amendment system details
- `OWNERSHIP_CONTROL.md` - Ownership structure documentation
- `RAG_VECTOR_SEARCH.md` - Complete RAG & vector search documentation
- `RAG_QUICKSTART.md` - Quick start guide for semantic search
- `LINEAGE_EVALUATOR.md` - Attribute lineage and derivation
- `VALIDATION_AUDIT.md` - Validation and audit trail system

### Migrations & Seeds
- `internal/storage/migrations/001_regulatory_ontology.sql` - Ontology schema
- `internal/ontology/seeds/ontology_seed.sql` - Regulations, documents, attributes

### Example DSL Files
- `sample_case.dsl` - Basic case example
- `ontology_example.dsl` - Full ontology-aware example
- `ownership_case.dsl` - Ownership structure example

---

## Dependencies

- `github.com/alecthomas/participle/v2` - Grammar-based parser generation
- `github.com/jmoiron/sqlx` - PostgreSQL extensions for database/sql
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/sashabaranov/go-openai` - OpenAI API client for embeddings
- `github.com/expr-lang/expr` - Expression language for lineage rules

---

## Regulatory Ontology (v1.2)

The ontology provides semantic grounding for compliance requirements:

**Regulations**: FATCA, CRS, AMLD5, AMLD6, MAS626, HKMAAML, UKMLR2017, BSAAML

**Documents**: 30+ types (W-8BEN, Certificates, UBO Declarations, etc.)

**Attributes**: 30+ data points (Tax Residency, UBO Info, Entity Details, etc.)

**Relationships**: 60+ attribute-document mappings with source tiers

See `REGULATORY_ONTOLOGY.md` for complete details.

---

## Version History

- **v1.0**: Initial DSL with parsing, validation, storage
- **v1.1**: Added ownership structures, control hierarchy, amendments
- **v1.2**: Added regulatory data ontology, data dictionary, document requirements
- **v1.3**: Added lineage engine, attribute derivation, validation audit trail
- **v1.4**: Added RAG & vector search with OpenAI embeddings

---

## RAG & Vector Search (v1.4)

The system now includes semantic search capabilities over the regulatory ontology:

### Key Features
- **Semantic Search**: Find attributes by meaning using OpenAI embeddings
- **Vector Similarity**: Discover related attributes automatically
- **Agent Integration**: Power AI agents with regulatory context
- **Synonym Resolution**: Map natural language to formal codes
- **Embedding Storage**: 1536-dimensional vectors in PostgreSQL with pgvector

### Quick Setup
```bash
# Install pgvector extension
brew install pgvector  # macOS
sudo apt install postgresql-15-pgvector  # Ubuntu

# Enable in database
psql -d kyc_dsl -c "CREATE EXTENSION vector;"

# Set OpenAI API key
export OPENAI_API_KEY="sk-..."

# Seed embeddings
./kycctl seed-metadata

# Try semantic search
./kycctl search-metadata "tax compliance requirements"
./kycctl similar-attributes UBO_NAME
```

### Architecture
```
Query → OpenAI Embedding → Vector Search (pgvector) → Ranked Results
```

### Use Cases
1. **AI Agent Context**: "Find attributes for EU fund KYC" → AMLD5 attributes
2. **Explainability**: "Why need UBO?" → AMLD5 Article 3, FATF Rec 24
3. **Synonym Resolution**: "Company Name" → REGISTERED_NAME
4. **Risk Prioritization**: Find CRITICAL attributes for enhanced due diligence

### Database Schema
```sql
CREATE TABLE kyc_attribute_metadata (
    attribute_code TEXT PRIMARY KEY,
    synonyms TEXT[],
    business_context TEXT,
    regulatory_citations TEXT[],
    embedding vector(1536),  -- OpenAI text-embedding-3-large
    ...
);

CREATE INDEX ON kyc_attribute_metadata 
    USING ivfflat (embedding vector_cosine_ops);
```

### Example Results
```
Query: "tax reporting requirements"

Results:
1. TAX_RESIDENCY_COUNTRY (similarity: 0.87)
   - Citations: FATCA §1471(b)(1)(D), CRS
   - Risk: HIGH
   
2. FATCA_STATUS (similarity: 0.85)
   - Citations: FATCA §1471-1474
   - Risk: HIGH
```

See [RAG_VECTOR_SEARCH.md](RAG_VECTOR_SEARCH.md) for complete documentation.

---

---

## RAG Feedback Loop (v1.5)

The system now includes a self-correcting feedback mechanism that continuously improves search relevance:

### Key Features
- **Self-Learning**: Automatically adjusts relevance scores based on user and AI agent feedback
- **Multi-Agent Support**: Accepts feedback from humans, AI agents, and automated systems
- **Confidence Weighting**: Scales impact based on feedback confidence (0.0-1.0)
- **Real-Time Updates**: Database triggers apply changes immediately
- **Analytics Dashboard**: Track sentiment trends and learning progress

### Quick Setup
```bash
# Apply migration
./scripts/migrate_feedback.sh

# Start API server
go run cmd/kycserver/main.go

# Submit test feedback
curl -X POST http://localhost:8080/rag/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "query_text": "beneficial owner name",
    "attribute_code": "UBO_NAME",
    "feedback": "positive",
    "confidence": 0.9,
    "agent_type": "human"
  }'

# View analytics
curl http://localhost:8080/rag/feedback/analytics
```

### API Endpoints
- `POST /rag/feedback` - Submit feedback on search results
- `GET /rag/feedback/recent` - Get recent feedback entries
- `GET /rag/feedback/analytics` - Get feedback analytics and trends
- `GET /rag/feedback/attribute/{code}` - Get feedback for specific attribute
- `GET /rag/feedback/summary` - Get aggregated summary

### Database Schema
```sql
-- Feedback table with trigger-based learning
CREATE TABLE rag_feedback (
    id SERIAL PRIMARY KEY,
    query_text TEXT NOT NULL,
    attribute_code TEXT,
    document_code TEXT,
    regulation_code TEXT,
    feedback feedback_sentiment,  -- positive/negative/neutral
    confidence FLOAT DEFAULT 1.0,
    agent_name TEXT,
    agent_type TEXT,              -- human/ai/automated
    created_at TIMESTAMP DEFAULT NOW()
);

-- Automatic relevance score adjustment
CREATE TRIGGER trig_feedback_relevance
AFTER INSERT ON rag_feedback
FOR EACH ROW
EXECUTE FUNCTION update_relevance();
```

### Learning Mechanism
```
Feedback → Trigger → Score Adjustment
  positive: relevance_score + (0.05 × confidence)
  negative: relevance_score - (0.05 × confidence)
  neutral:  no change
```

### Use Cases
1. **AI Agent Improvement**: AI agents submit feedback to improve future searches
2. **Human Validation**: Compliance officers validate search results
3. **A/B Testing**: Automated systems test different relevance strategies
4. **Quality Monitoring**: Track search effectiveness over time

See [RAG_FEEDBACK.md](RAG_FEEDBACK.md) for complete documentation.

---

**Last Updated**: 2024  
**Version**: 1.5  
**Maintainer**: See repository metadata
</text>
