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
- 🏢 **Ownership & Control** - Legal owners, beneficial owners, controllers (v1.1)
- ✅ **Advanced Validation** - Ownership percentages, duplicates, structural checks
- 🔄 **Amendment System** - Incremental case evolution through lifecycle phases
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

#### Apply Amendments

```bash
# Add policy discovery
./bin/kycctl amend CASE-NAME --step=policy-discovery

# Add ownership structure
./bin/kycctl amend CASE-NAME --step=ownership-discovery

# Finalize case
./bin/kycctl amend CASE-NAME --step=approve
```

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

- `DISCOVER-POLICIES` - Policy discovery phase
- `SOLICIT-DOCUMENTS` - Document solicitation phase
- `EXTRACT-DATA` - Data extraction
- `BUILD-OWNERSHIP-TREE` - Ownership structure building
- `VERIFY-OWNERSHIP` - Ownership verification
- `ASSESS-RISK` - Risk assessment phase
- `REGULATOR-NOTIFY` - Regulatory notification

### Valid Token States

- `pending` - Initial state
- `approved` - Case approved
- `declined` - Case declined
- `review` - Under review

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
│   ├── storage/          # PostgreSQL layer
│   │   └── postgres.go
│   └── amend/            # Amendment system
│       ├── amend.go      # Core amendment engine
│       ├── mutations.go  # Predefined mutations
│       └── transitions.go # Lifecycle phases
├── Makefile              # Build automation
├── verify.sh             # Verification script
├── sample_case.dsl       # Example DSL file
├── ownership_case.dsl    # Ownership example
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

### Amendment Commands

```bash
# Policy discovery phase
kycctl amend CASE-NAME --step=policy-discovery

# Document solicitation phase  
kycctl amend CASE-NAME --step=document-solicitation

# Ownership & control phase
kycctl amend CASE-NAME --step=ownership-discovery

# Risk assessment phase
kycctl amend CASE-NAME --step=risk-assessment

# Finalization
kycctl amend CASE-NAME --step=approve
kycctl amend CASE-NAME --step=decline
kycctl amend CASE-NAME --step=review
```

### Amendment Commands

```bash
# Evolve cases through lifecycle phases
kycctl amend <case> --step=policy-discovery
kycctl amend <case> --step=document-solicitation
kycctl amend <case> --step=ownership-discovery
kycctl amend <case> --step=risk-assessment
kycctl amend <case> --step=approve
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

**`kyc_case_amendments`** - Amendment audit trail
- `id`, `case_name`, `step`, `change_type`, `diff`, `created_at`

**`kyc_case_amendments`** - Amendment audit trail
- `id`, `case_name`, `step`, `change_type`, `diff`, `created_at`

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

### Ownership Validation Tests

- **Structural checks** - At least one owner or controller required
- **Percentage validation** - Legal ownership must sum to 100% ± 0.5%
- **Duplicate detection** - No duplicate owners, beneficial owners, or controllers
- **Controller requirements** - Multiple owners require at least one controller

## Code Quality

- ✅ **go vet**: Clean
- ✅ **golangci-lint**: Clean
- ✅ **gofmt**: All code formatted
- ✅ **errcheck**: All errors handled
- ✅ **Test Coverage**: Core parser and ownership validation covered
- ✅ **Ownership Validation**: Structural and semantic checks

## Audit Trail

Every case mutation produces:
- New row in `kyc_case_versions` with complete DSL snapshot
- New row in `kyc_case_amendments` with step and change type
- Immutable SHA-256 hash fingerprint
- Automatic version numbering per case

You can prove:
- **Determinism**: Identical DSL → identical hash
- **Provenance**: Each version linked to timestamp and step
- **Replayability**: Rebuild state from any historical snapshot
- **Compliance**: Complete audit trail for regulatory requirements

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

## Ownership & Control (Grammar v1.1)

The system supports comprehensive ownership and control tracking:

### Ownership Types

- **Legal Ownership** - Registered shareholders (must sum to 100%)
- **Beneficial Ownership** - Economic interest or voting rights
- **Controllers** - Persons with significant control or influence

### Validation Rules

1. At least one owner or controller required
2. Legal ownership percentages must sum to 100% ± 0.5%
3. No duplicate entities allowed
4. Multiple owners require at least one controller

### Example

```lisp
(ownership-structure
  (owner BLACKROCK-PLC 100)
  (beneficial-owner LARRY-FINK 35)
  (controller JANE-DOE "Senior Managing Official")
  (controller JOHN-SMITH "Director, Risk Oversight"))
```

See `OWNERSHIP_CONTROL.md` for complete documentation.

## Amendment System

Cases evolve through defined lifecycle phases:

1. **Case Creation** - Initial setup
2. **Policy Discovery** - Auto-inject policies
3. **Document Solicitation** - Add obligations
4. **Ownership &

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

For more details, see:
- `CLAUDE.md` - Project guidance
- `REFACTORING_SUMMARY.md` - CLI refactoring details
- `AMENDMENT_SYSTEM.md` - Amendment lifecycle
- `OWNERSHIP_CONTROL.md` - Ownership & control system
- `SESSION_SUMMARY.md` - Current system state