# KYC-DSL

A Domain-Specific Language (DSL) processor for Know Your Customer (KYC) compliance cases with versioning, validation, and audit trail capabilities.

## Overview

KYC-DSL is a Go-based system that parses S-expression formatted DSL files containing KYC case definitions, validates them against grammar and semantic rules, and persists them to PostgreSQL with full version tracking and SHA-256 content hashing.

## Features

- 🔍 **S-Expression Parser** - Custom tokenizer with quoted string support
- 📋 **Grammar Versioning** - EBNF grammar stored and tracked in database
- ✅ **Multi-Layer Validation** - Structural and semantic checks
- 🔄 **Round-Trip Serialization** - Parse → Bind → Serialize → Parse
- 🔒 **Content Integrity** - SHA-256 hashing for audit trail
- 📜 **Version Control** - Automatic versioning of case snapshots
- 🗄️ **PostgreSQL Storage** - Persistent storage with complete history
- 🧪 **Comprehensive Testing** - Unit tests for all core components
- ⚡ **Green Tea GC** - Built with `GOEXPERIMENT=greenteagc` for enhanced performance

## Quick Start

### Prerequisites

- Go 1.25+ (with greenteagc support)
- PostgreSQL 12+
- golangci-lint (optional, for linting)

### Installation

```bash
# Clone the repository
git clone https://github.com/adamtc007/KYC-DSL.git
cd KYC-DSL

# Install dependencies
make deps

# Build the binary
make build
```

### Database Setup

Set environment variables for PostgreSQL connection:

```bash
export PGHOST=localhost
export PGPORT=5432
export PGUSER=your_user
export PGPASSWORD=your_password  # optional
export PGDATABASE=kyc_dsl
```

Create the database:

```bash
createdb kyc_dsl
```

Tables are automatically created on first connection.

### Usage

#### Store Grammar Definition

```bash
./bin/kycctl grammar
```

This stores the EBNF grammar in the database for validation purposes.

#### Process a DSL File

```bash
./bin/kycctl sample_case.dsl
```

This will:
1. Parse the DSL file
2. Bind to typed models
3. Validate against grammar and semantics
4. Serialize back to DSL text
5. Store with automatic versioning

#### Get Help

```bash
./bin/kycctl help
```

## DSL Syntax

KYC cases are defined using S-expressions:

```lisp
(kyc-case CASE-NAME
  (nature-purpose
    (nature "Description of the nature")
    (purpose "Description of the purpose")
  )
  (client-business-unit UNIT-NAME)
  (function FUNCTION-NAME)
  (policy POLICY-CODE)
  (obligation OBLIGATION-CODE)
  (kyc-token "status")
)
```

### Valid Function Names

- `DISCOVER-POLICIES`
- `SOLICIT-DOCUMENTS`
- `EXTRACT-DATA`
- `VERIFY-OWNERSHIP`
- `ASSESS-RISK`
- `REGULATOR-NOTIFY`

### Valid Token States

- `pending`
- `approved`
- `declined`
- `review`

## Project Structure

```
KYC-DSL/
├── cmd/kycctl/           # CLI entry point (11 lines)
│   └── main.go
├── internal/
│   ├── cli/              # CLI business logic (128 lines)
│   │   └── cli.go
│   ├── parser/           # DSL parsing and manipulation
│   │   ├── parser.go     # Tokenizer and AST parser
│   │   ├── parser_test.go # Comprehensive tests
│   │   ├── binder.go     # AST → Model binding
│   │   ├── serializer.go # Model → DSL serialization
│   │   ├── validator.go  # Grammar + semantic validation
│   │   └── grammar.go    # EBNF grammar definition
│   ├── model/            # Data models
│   │   └── model.go
│   ├── engine/           # Execution engine
│   │   └── engine.go
│   └── storage/          # PostgreSQL layer
│       └── postgres.go
├── Makefile              # Build automation
├── verify.sh             # Verification script
├── sample_case.dsl       # Example DSL file
└── CLAUDE.md             # Project documentation
```

## Development

### Build Commands

```bash
make build              # Build with greenteagc
make run                # Build and run sample case
make test               # Run all tests
make test-parser        # Run parser tests only
make test-verbose       # Verbose test output
make clean              # Remove build artifacts
make deps               # Download dependencies
make fmt                # Format code
make lint               # Run golangci-lint
make verify             # Run comprehensive checks
```

### Running Tests

```bash
# All tests
make test

# Parser tests only
make test-parser

# With verbose output
make test-verbose
```

### Verification

Run comprehensive checks including build, tests, linting, and security:

```bash
make verify
```

Or directly:

```bash
./verify.sh
```

This checks:
- ✅ Go installation
- ✅ Module verification
- ✅ Build success
- ✅ Code formatting
- ✅ Tests
- ✅ Linting
- ✅ Security (hardcoded credentials)
- ✅ Binary functionality

## Architecture

### Parser Pipeline

```
DSL File → Tokenize → Parse → AST → Bind → Model
                                              ↓
                                         Validate
                                              ↓
                                         Serialize
                                              ↓
                                       Store (Versioned)
```

### Database Schema

**`kyc_cases`** - Base case records
- `id`, `name`, `version`, `status`, `last_updated`

**`kyc_case_versions`** - Version history with snapshots
- `id`, `case_name`, `version`, `dsl_snapshot`, `hash`, `created_at`

**`kyc_grammar`** - Grammar definitions
- `id`, `name`, `version`, `ebnf`, `created_at`

**`kyc_policies`** - Policy registry
- `id`, `code`, `description`, `created_at`

## Testing

The project includes comprehensive tests:

- **TestTokenize** - 8 tokenization scenarios
- **TestParse** - AST parsing validation
- **TestBind** - Model binding verification
- **TestSerializeCases** - Serialization output
- **TestRoundTrip** - Complete cycle validation
- **TestTrimQuotes** - Edge case handling
- **TestParseMultipleCases** - Multi-case files

All tests pass with `GOEXPERIMENT=greenteagc`.

## Code Quality

- ✅ **go vet**: Clean
- ✅ **golangci-lint**: Clean
- ✅ **gofmt**: All code formatted
- ✅ **errcheck**: All errors handled
- ✅ **Test Coverage**: Core parser covered

## Audit Trail

Every case mutation produces:
- New row in `kyc_case_versions`
- Immutable SHA-256 hash fingerprint
- Complete serialized DSL text
- Automatic version numbering

You can prove:
- **Determinism**: Identical DSL → identical hash
- **Provenance**: Each version linked to time
- **Replayability**: Rebuild state from any snapshot

## Example Queries

```sql
-- Get all versions of a case
SELECT version, hash, created_at 
FROM kyc_case_versions 
WHERE case_name = 'AVIVA-EU-EQUITY-FUND' 
ORDER BY version;

-- Get specific version snapshot
SELECT dsl_snapshot 
FROM kyc_case_versions 
WHERE case_name = 'AVIVA-EU-EQUITY-FUND' 
AND version = 3;

-- Check for duplicate content
SELECT hash, COUNT(*) 
FROM kyc_case_versions 
GROUP BY hash 
HAVING COUNT(*) > 1;
```

## Refactoring History

The CLI was refactored from a monolithic 85-line `main.go` into:
- **11-line entry point** (`cmd/kycctl/main.go`)
- **128-line CLI library** (`internal/cli/cli.go`)

This provides:
- Clean separation of concerns
- Testable business logic
- Reusable components
- Better maintainability

See `REFACTORING_SUMMARY.md` for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Run `make verify` to ensure quality
5. Submit a pull request

## License

[Your License Here]

## Authors

- Adam TC

## Acknowledgments

Built with:
- [participle/v2](https://github.com/alecthomas/participle) - Parser generation
- [sqlx](https://github.com/jmoiron/sqlx) - PostgreSQL extensions
- [pq](https://github.com/lib/pq) - PostgreSQL driver

---

For more details, see `CLAUDE.md` and `REFACTORING_SUMMARY.md`.