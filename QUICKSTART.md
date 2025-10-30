# KYC-DSL Quick Start Guide

## Installation & Setup

### 1. Prerequisites
- Go 1.21+
- PostgreSQL 14+
- golangci-lint (optional, for development)

### 2. Clone & Build
```bash
git clone <repository-url>
cd KYC-DSL
make build
```

### 3. Database Setup

Set environment variables:
```bash
export PGHOST=localhost
export PGPORT=5432
export PGUSER=myuser
export PGDATABASE=kyc_dsl
export PGPASSWORD=mypassword  # optional
```

Create database:
```bash
createdb kyc_dsl
```

### 4. Initialize Ontology
```bash
./scripts/init_ontology.sh
```

This creates all tables and loads seed data for:
- 8 regulations (FATCA, AMLD5, MAS626, etc.)
- 30+ document types
- 30+ compliance attributes
- 60+ attribute-document mappings

---

## Basic Usage

### View Ontology
```bash
./bin/kycctl ontology
```

Expected output:
```
=== Regulatory Data Ontology Summary ===

📘 FATCA — Foreign Account Tax Compliance Act (US)
   📄 W8BEN — IRS Form W-8BEN
   📄 W8BENE — IRS Form W-8BEN-E
   📄 W9 — IRS Form W-9
📘 AMLD5 — 5th EU Anti-Money Laundering Directive (EU)
   📄 CERT-INC — Certificate of Incorporation
   📄 UBO-DECL — Ultimate Beneficial Owner Declaration
   ...
```

### Process a DSL Case
```bash
./bin/kycctl sample_case.dsl
```

This will:
1. Parse the DSL file
2. Validate grammar and semantics
3. Persist to database with version 1
4. Display case summary

### Try the Ontology-Aware Example
```bash
./bin/kycctl ontology_example.dsl
```

This demonstrates:
- Data dictionary with primary/secondary/tertiary sources
- Document requirements by jurisdiction
- Full ownership structure
- Integration with ontology

---

## Amendment Workflow

### Complete Lifecycle Example

```bash
# 1. Create initial case
./bin/kycctl sample_case.dsl

# 2. Discover applicable policies
./bin/kycctl amend AVIVA-EU-EQUITY-FUND --step=policy-discovery

# 3. Auto-populate documents from ontology (NEW in v1.2!)
./bin/kycctl amend AVIVA-EU-EQUITY-FUND --step=document-discovery

# 4. Add ownership structure
./bin/kycctl amend AVIVA-EU-EQUITY-FUND --step=ownership-discovery

# 5. Assess risk
./bin/kycctl amend AVIVA-EU-EQUITY-FUND --step=risk-assessment

# 6. Approve case
./bin/kycctl amend AVIVA-EU-EQUITY-FUND --step=approve
```

Each amendment:
- Creates a new version in the database
- Records the amendment in audit trail
- Preserves full history with SHA-256 hashing

---

## Creating Your First Case

### 1. Simple Case (basic.dsl)
```lisp
(kyc-case MY-FIRST-CASE
  (nature-purpose
    (nature "Test Case")
    (purpose "Learning KYC-DSL"))
  
  (client-business-unit OPERATIONS)
  
  (policy KYCPOL-GLOBAL-2025)
  
  (function VALIDATE-IDENTITY)
  
  (kyc-token "pending")
)
```

Process it:
```bash
./bin/kycctl basic.dsl
```

### 2. Ontology-Aware Case (my-case.dsl)
```lisp
(kyc-case MY-ONTOLOGY-CASE
  (nature-purpose
    (nature "Fund KYC")
    (purpose "EU Investment Fund Onboarding"))
  
  (client-business-unit FUND-SERVICES)
  
  (data-dictionary
    (attribute REGISTERED_NAME
      (primary-source (document CERT-INC)))
    (attribute UBO_NAME
      (primary-source (document UBO-DECL))
      (secondary-source (document SHARE-REGISTER)))
  )
  
  (document-requirements
    (jurisdiction EU)
    (required
      (document CERT-INC "Certificate of Incorporation")
      (document UBO-DECL "Ultimate Beneficial Owner Declaration")
    ))
  
  (kyc-token "pending")
)
```

Process it:
```bash
./bin/kycctl my-case.dsl
```

---

## Querying the Ontology

### From Go Code
```go
import "github.com/adamtc007/KYC-DSL/internal/ontology"

db, _ := storage.ConnectPostgres()
repo := ontology.NewRepository(db)

// Get all regulations
regs, _ := repo.ListRegulations()

// Find documents for a regulation
docs, _ := repo.ListDocumentsByRegulation("AMLD5")

// Find what proves an attribute
sources, _ := repo.GetDocumentSources("UBO_NAME")

// Find what a document proves
attrs, _ := repo.GetAttributesForDocument("CERT-INC")
```

### From SQL
```sql
-- All EU documents
SELECT d.code, d.name 
FROM kyc_documents d 
WHERE d.jurisdiction = 'EU';

-- What proves UBO_NAME?
SELECT d.code, d.name, l.source_tier
FROM kyc_attr_doc_links l
JOIN kyc_documents d ON d.code = l.document_code
WHERE l.attribute_code = 'UBO_NAME'
ORDER BY 
  CASE l.source_tier 
    WHEN 'Primary' THEN 1 
    WHEN 'Secondary' THEN 2 
    WHEN 'Tertiary' THEN 3 
  END;

-- Personal data attributes (GDPR)
SELECT code, name, domain 
FROM kyc_attributes 
WHERE is_personal_data = TRUE;
```

---

## Testing

### Run All Tests
```bash
make test
```

### Run Specific Tests
```bash
make test-parser
```

### Test Coverage
```bash
go test -cover ./...
```

### Lint Code
```bash
make lint
```

---

## Common Tasks

### Add a New Regulation
```sql
INSERT INTO kyc_regulations (code, name, jurisdiction, authority, description, effective_from)
VALUES ('FINMA', 'Swiss FINMA AML Rules', 'CH', 'FINMA', 'Swiss AML requirements', '2020-01-01');
```

### Add a New Document Type
```sql
INSERT INTO kyc_documents (code, name, domain, jurisdiction, regulation_code, source_type, validity_years, description)
VALUES ('CH-CERT', 'Swiss Company Certificate', 'Entity', 'CH', 'FINMA', 'Official', 0, 'Swiss entity registration');
```

### Map Attribute to Document
```sql
INSERT INTO kyc_attr_doc_links (attribute_code, document_code, source_tier, regulation_code)
VALUES ('REGISTERED_NAME', 'CH-CERT', 'Primary', 'FINMA');
```

### View Case History
```sql
-- All versions of a case
SELECT name, version, status, last_updated, dsl_hash
FROM kyc_cases
WHERE name = 'AVIVA-EU-EQUITY-FUND'
ORDER BY version DESC;

-- Amendment trail
SELECT amendment_type, created_at
FROM case_amendments
WHERE case_name = 'AVIVA-EU-EQUITY-FUND'
ORDER BY created_at;
```

---

## Troubleshooting

### Database Connection Issues
```bash
# Test connection
psql -h $PGHOST -p $PGPORT -U $PGUSER -d $PGDATABASE -c '\q'

# Check environment variables
env | grep PG
```

### Parsing Errors
- Check DSL syntax (S-expressions must be balanced)
- Ensure quoted strings use double quotes: `"text"`
- Validate ownership percentages sum to 100%
- Ensure at least one controller exists in ownership structures

### Linting Errors
```bash
# Auto-fix formatting
make fmt

# View specific issues
golangci-lint run --verbose
```

---

## Next Steps

### Learn More
- **DSL Syntax**: See `sample_case.dsl` and `ontology_example.dsl`
- **Ontology Details**: Read `REGULATORY_ONTOLOGY.md`
- **Amendment System**: Read `AMENDMENT_SYSTEM.md`
- **Ownership Rules**: Read `OWNERSHIP_CONTROL.md`

### Extend the System
1. Add new regulations for your jurisdiction
2. Define jurisdiction-specific document types
3. Create custom amendment steps
4. Build frontend integrations with the repository API

### Get Help
- Check call trees in `CLAUDE.md`
- Review test cases in `internal/parser/parser_test.go`
- Examine seed data in `internal/ontology/seeds/ontology_seed.sql`

---

**Version**: 1.2  
**Status**: Production Ready  
**Coverage**: US, EU, UK, Singapore, Hong Kong