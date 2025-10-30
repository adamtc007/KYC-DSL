# Regulatory Data Ontology

## Overview

The Regulatory Data Ontology is a formal semantic model that defines all compliance concepts, relationships, documents, and evidence that underpin KYC/AML obligations across jurisdictions. It transforms the KYC-DSL from a simple case processing language into a **deterministic, explainable, and globally extensible compliance system**.

### What It Provides

| Component | Description |
|-----------|-------------|
| **Regulations** | Formal definitions of laws and regulatory frameworks (FATCA, AMLD5, MAS 626, etc.) |
| **Documents** | Evidence types that prove compliance attributes (W-8BEN, certificates, declarations) |
| **Attributes** | Data points required for compliance (tax residency, UBO info, entity details) |
| **Relationships** | Semantic links between attributes, documents, and regulations |
| **Jurisdictions** | Geographic and regulatory scopes (US, EU, UK, SG, HK, etc.) |

## Architecture

### Database Schema

The ontology is implemented as a relational schema with five core tables:

```
kyc_regulations          - Regulatory frameworks and laws
kyc_documents           - Evidence types and proof documents
kyc_attributes          - Required compliance data points
kyc_attr_doc_links      - Maps attributes → documents (with source tiers)
kyc_doc_reg_links       - Maps documents → regulations
```

### Source Tiers

Documents are classified by evidence strength:

- **Primary**: Official, authoritative sources (e.g., Certificate of Incorporation)
- **Secondary**: Corroborating evidence (e.g., Share Register confirms UBO Declaration)
- **Tertiary**: Operational validation (e.g., internal verification notes)

### Coverage

#### Regulations Included

| Code | Name | Jurisdiction | Authority |
|------|------|--------------|-----------|
| `FATCA` | Foreign Account Tax Compliance Act | US | IRS |
| `CRS` | Common Reporting Standard | GLOBAL | OECD |
| `AMLD5` | 5th EU Anti-Money Laundering Directive | EU | European Commission |
| `AMLD6` | 6th EU Anti-Money Laundering Directive | EU | European Commission |
| `MAS626` | MAS Notice 626 | SG | Monetary Authority of Singapore |
| `HKMAAML` | HKMA AML Guideline | HK | Hong Kong Monetary Authority |
| `UKMLR2017` | UK Money Laundering Regulations 2017 | UK | HM Treasury |
| `BSAAML` | Bank Secrecy Act / AML | US | FinCEN |

#### Document Types

**Tax Documents**: W-8BEN, W-8BEN-E, W-9, CRS Self-Certification

**Entity Documents**: Certificate of Incorporation, Business Registration, Articles of Association

**Ownership Documents**: UBO Declaration, Share Register, PSC Register, Ownership Charts

**Identity Documents**: Passport, National ID, Driver's License

**Financial Documents**: Audited Financials, Tax Returns, Source of Wealth Letters

**Control Documents**: Board Resolutions, Power of Attorney, Director Lists

## DSL Integration

### Version 1.2 Grammar Extensions

The KYC-DSL now supports two ontology-aware constructs:

#### 1. Data Dictionary

Declares attribute sources with tiered evidence:

```lisp
(data-dictionary
  (attribute REGISTERED_NAME
    (primary-source (document CERT-INC))
    (tertiary-source "Ops Validation"))
  (attribute UBO_NAME
    (primary-source (document UBO-DECL))
    (secondary-source (document SHARE-REGISTER)))
)
```

#### 2. Document Requirements

Lists required documents by jurisdiction:

```lisp
(document-requirements
  (jurisdiction EU)
  (required
    (document CERT-INC "Certificate of Incorporation")
    (document UBO-DECL "Ultimate Beneficial Owner Declaration")
    (document W8BENE "IRS Form W-8BEN-E")
  ))
```

### Example: Ontology-Aware Case

See `ontology_example.dsl` for a complete example that demonstrates:
- Data dictionary with multi-tier sources
- Document requirements across multiple jurisdictions
- Integration with ownership structures
- Full lifecycle functions

## Setup

### 1. Initialize Database

Run the initialization script to create schema and load seed data:

```bash
./scripts/init_ontology.sh
```

This will:
- Create ontology tables (`kyc_regulations`, `kyc_documents`, `kyc_attributes`, etc.)
- Load seed data for US, EU, and APAC regulations
- Create indexes for efficient queries
- Display summary statistics

### 2. Verify Installation

View the ontology structure:

```bash
./kycctl ontology
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

## Usage

### CLI Commands

#### View Ontology

```bash
./kycctl ontology
```

Displays all regulations and their associated documents.

#### Process Ontology-Aware DSL

```bash
./kycctl ontology_example.dsl
```

Parses and persists a case with data dictionary and document requirements.

#### Auto-Populate from Ontology

```bash
./kycctl amend BLACKROCK-GLOBAL-EQUITY-FUND --step=document-discovery
```

Automatically discovers and adds document requirements based on jurisdiction and regulation.

### Programmatic Access

#### Repository Methods

```go
import "github.com/adamtc007/KYC-DSL/internal/ontology"

repo := ontology.NewRepository(db)

// List all regulations
regs, _ := repo.ListRegulations()

// Get documents for a specific regulation
docs, _ := repo.ListDocumentsByRegulation("AMLD5")

// Find which documents prove an attribute
sources, _ := repo.GetDocumentSources("UBO_NAME")

// Get attributes required by a document
attrs, _ := repo.GetAttributesForDocument("CERT-INC")
```

## Amendment System Integration

### Document Discovery Amendment

The `document-discovery` amendment step uses the ontology to automatically populate document requirements:

```bash
./kycctl amend CASE-NAME --step=document-discovery
```

This:
1. Queries the ontology database for applicable regulations
2. Retrieves required documents for the jurisdiction
3. Maps attributes to their evidence sources
4. Adds `(data-dictionary ...)` and `(document-requirements ...)` sections to the case
5. Persists the updated case with full version control

### Workflow Example

```bash
# 1. Create initial case
./kycctl initial_case.dsl

# 2. Discover policies
./kycctl amend MY-CASE --step=policy-discovery

# 3. Auto-populate documents from ontology
./kycctl amend MY-CASE --step=document-discovery

# 4. Add ownership structure
./kycctl amend MY-CASE --step=ownership-discovery

# 5. Complete case
./kycctl amend MY-CASE --step=approve
```

## Benefits

### 1. Deterministic Compliance

Every document requirement is traceable to:
- A specific regulation
- A jurisdiction
- A set of attributes it evidences
- Its tier of proof (primary/secondary/tertiary)

### 2. Explainability

You can answer questions like:
- "Why do we need this document?" → Links to regulation
- "What does this document prove?" → Links to attributes
- "What are alternative sources?" → Shows secondary/tertiary tiers

### 3. Global Extensibility

Adding new jurisdictions is simple:
- Insert regulation records
- Add document types
- Map attribute-document relationships
- DSL cases automatically inherit new requirements

### 4. Audit Trail

The ontology creates a compliance audit trail:
- Every case references specific regulations
- Document requirements are version-controlled
- Changes are tracked through amendments
- Full provenance from regulation → document → attribute

## Data Model

### Regulation

```go
type Regulation struct {
    Code          string    // e.g., "FATCA", "AMLD5"
    Name          string    // Full regulation name
    Jurisdiction  string    // "US", "EU", "GLOBAL", etc.
    Authority     string    // Regulatory authority
    Description   string
    EffectiveFrom time.Time // When regulation came into force
    EffectiveTo   *time.Time // Optional sunset date
}
```

### Document

```go
type Document struct {
    Code           string // e.g., "W8BEN", "CERT-INC"
    Name           string // Full document name
    Domain         string // "Tax", "Entity", "Ownership", etc.
    Jurisdiction   string
    RegulationCode string // Links to regulation
    SourceType     string // "Official", "Client", "Operational"
    ValidityYears  int    // Document expiration (0 = never)
    Description    string
}
```

### Attribute

```go
type Attribute struct {
    Code         string // e.g., "UBO_NAME", "TAX_RESIDENCY_COUNTRY"
    Name         string // Human-readable name
    Domain       string // "Entity", "Tax", "Ownership", etc.
    Description  string
    RiskCategory string // "LOW", "MEDIUM", "HIGH"
    IsPersonal   bool   // GDPR/privacy flag
}
```

### AttributeDocumentLink

```go
type AttributeDocumentLink struct {
    AttributeCode  string // Which attribute
    DocumentCode   string // Which document proves it
    SourceTier     string // "Primary", "Secondary", "Tertiary"
    IsMandatory    bool
    Jurisdiction   string
    RegulationCode string
    Notes          string
}
```

## Extensibility

### Adding New Regulations

```sql
INSERT INTO kyc_regulations (code, name, jurisdiction, authority, description, effective_from)
VALUES ('CASL', 'Canadian AML Legislation', 'CA', 'FINTRAC', 'Canadian AML requirements', '2021-06-01');
```

### Adding New Documents

```sql
INSERT INTO kyc_documents (code, name, domain, jurisdiction, regulation_code, source_type, validity_years, description)
VALUES ('CA-INC-CERT', 'Canadian Certificate of Incorporation', 'Entity', 'CA', 'CASL', 'Official', 0, 'Canadian entity formation document');
```

### Mapping Attributes to Documents

```sql
INSERT INTO kyc_attr_doc_links (attribute_code, document_code, source_tier, regulation_code)
VALUES ('REGISTERED_NAME', 'CA-INC-CERT', 'Primary', 'CASL');
```

## Query Examples

### Find All Documents for a Jurisdiction

```sql
SELECT d.code, d.name, r.name as regulation
FROM kyc_documents d
JOIN kyc_regulations r ON r.code = d.regulation_code
WHERE d.jurisdiction = 'EU'
ORDER BY r.code, d.code;
```

### Find Attributes Requiring Personal Data

```sql
SELECT a.code, a.name, a.domain
FROM kyc_attributes a
WHERE a.is_personal_data = TRUE
ORDER BY a.risk_category DESC;
```

### Find Alternative Evidence Sources

```sql
SELECT l.source_tier, d.code, d.name
FROM kyc_attr_doc_links l
JOIN kyc_documents d ON d.code = l.document_code
WHERE l.attribute_code = 'UBO_NAME'
ORDER BY 
  CASE l.source_tier 
    WHEN 'Primary' THEN 1 
    WHEN 'Secondary' THEN 2 
    WHEN 'Tertiary' THEN 3 
  END;
```

## Future Enhancements

### Planned Features

1. **Dynamic Policy Derivation**: Automatically generate `(policy ...)` declarations based on client profile
2. **Risk Scoring Integration**: Map ontology attributes to risk assessment models
3. **Document Validation Rules**: Define validation logic in the ontology (e.g., "W-8BEN expires in 3 years")
4. **Multi-Lingual Support**: Add translations for document names and descriptions
5. **Ontology Versioning**: Track changes to regulations and requirements over time
6. **GraphQL API**: Expose ontology via GraphQL for frontend integrations
7. **Conflict Detection**: Identify when jurisdictions have competing requirements

### Research Directions

- **Semantic Reasoning**: Use ontology for automated compliance gap analysis
- **Natural Language Generation**: Generate human-readable compliance explanations
- **Regulatory Change Detection**: Monitor and alert on regulatory updates
- **Cross-Jurisdiction Mapping**: Identify equivalent regulations across borders

## References

### Regulations

- [FATCA (IRS)](https://www.irs.gov/businesses/corporations/foreign-account-tax-compliance-act-fatca)
- [OECD CRS](https://www.oecd.org/tax/automatic-exchange/common-reporting-standard/)
- [EU AMLD5](https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32018L0843)
- [MAS Notice 626](https://www.mas.gov.sg/regulation/notices/notice-626)
- [HKMA AML Guidelines](https://www.hkma.gov.hk/eng/regulatory-resources/regulatory-guides/anti-money-laundering-and-counter-financing-of-terrorism/)

### Implementation

- Database: PostgreSQL with `sqlx`
- Parser: `participle/v2` for DSL grammar
- Go modules: See `go.mod` for dependencies

## Support

For questions or issues with the Regulatory Data Ontology:

1. Check the seed data: `internal/ontology/seeds/ontology_seed.sql`
2. Review the schema: `internal/storage/migrations/001_regulatory_ontology.sql`
3. Examine the example: `ontology_example.dsl`
4. Run diagnostics: `./kycctl ontology`

---

**Version**: 1.2  
**Last Updated**: 2024  
**Status**: Production-ready for US, EU, UK, Singapore, and Hong Kong jurisdictions