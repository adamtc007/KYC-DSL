# Derived Attributes & Data Lineage

## Overview

The Derived Attributes system introduces a **public/private attribute classification** with explicit **data lineage tracking**. This enables transparent, auditable computation of risk flags, scores, and status indicators from observable source data.

### Key Concept

| Attribute Class | Description | Example |
|----------------|-------------|---------|
| **Public** | Observable data from documents or client-supplied facts | `RegisteredName`, `TaxResidencyCountry`, `UBO_NAME` |
| **Private** | Derived/computed values via transformation or rule from public attributes | `HighRiskJurisdictionFlag`, `EntityActiveStatus`, `UBOConcentrationScore` |

Every private attribute has **lineage** — explicit references to the public attributes and rules that generate it.

## Why This Matters

### 1. Data Lineage
Every private/system value is traceable to public, document-sourced data. No "black box" computations.

### 2. Configurability
Derived attributes are defined by declarative rules, not hard-coded logic. Rules can be changed without code changes.

### 3. Auditability
Auditors can reconstruct how a flag was computed: "Flag X is TRUE because attributes A and B matched rule R."

### 4. Explainability
AI agents can reason: "HighRiskJurisdictionFlag is TRUE because TaxResidencyCountry='IR' matches rule '(in [...IR...])'"

### 5. Cross-Jurisdiction Tuning
Same derivation pattern can have multiple implementations per jurisdiction (different rule sets).

## Database Schema

### Attribute Classification

```sql
ALTER TABLE kyc_attributes
ADD COLUMN attribute_class TEXT
    CHECK (attribute_class IN ('Public', 'Private'))
    DEFAULT 'Public';
```

### Derivation Tracking

```sql
CREATE TABLE kyc_attribute_derivations (
    id SERIAL PRIMARY KEY,
    derived_attribute_code TEXT REFERENCES kyc_attributes(code),
    source_attribute_code TEXT REFERENCES kyc_attributes(code),
    rule_expression TEXT NOT NULL,
    jurisdiction TEXT,
    regulation_code TEXT REFERENCES kyc_regulations(code),
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Example Data

| derived_attribute_code | source_attribute_code | rule_expression |
|------------------------|----------------------|-----------------|
| `HIGH_RISK_JURISDICTION_FLAG` | `TAX_RESIDENCY_COUNTRY` | `(if (in TaxResidencyCountry ['IR' 'KP']) true false)` |
| `ENTITY_ACTIVE_STATUS` | `REGISTERED_NAME` | `(registry-active? RegisteredName IncorporationJurisdiction)` |
| `UBO_CONCENTRATION_SCORE` | `UBO_PERCENT` | `(max UBO_PERCENT)` |

### Lineage View

```sql
CREATE VIEW attribute_lineage AS
SELECT
    a.code as derived_attribute,
    a.name as derived_attribute_name,
    d.source_attribute_code as source_attribute,
    sa.name as source_attribute_name,
    d.rule_expression,
    d.jurisdiction,
    d.regulation_code
FROM kyc_attributes a
JOIN kyc_attribute_derivations d ON d.derived_attribute_code = a.code
JOIN kyc_attributes sa ON sa.code = d.source_attribute_code
WHERE a.attribute_class = 'Private'
ORDER BY a.code, d.source_attribute_code;
```

## DSL Syntax

### Grammar (v1.3)

```ebnf
DerivedAttributes  = "(" "derived-attributes" { DerivedAttrDef } ")" ;
DerivedAttrDef     = "(" "attribute" Identifier
                        (sources SourceList)
                        (rule RuleExpr)
                        [jurisdiction Identifier]
                        [regulation Identifier]
                     ")" ;
SourceList         = "(" { Identifier } ")" ;
RuleExpr           = QuotedString ;
```

### Example DSL Section

```lisp
(derived-attributes
  ; Risk flag: High-risk jurisdiction
  (attribute HIGH_RISK_JURISDICTION_FLAG
    (sources (TAX_RESIDENCY_COUNTRY))
    (rule "(if (in TAX_RESIDENCY_COUNTRY ['IR' 'KP' 'SY']) true false)")
    (jurisdiction GLOBAL)
    (regulation AMLD5)
  )

  ; Ownership concentration metric
  (attribute UBO_CONCENTRATION_SCORE
    (sources (UBO_PERCENT))
    (rule "(max UBO_PERCENT)")
    (jurisdiction GLOBAL)
    (regulation AMLD5)
  )

  ; Entity age calculation
  (attribute ENTITY_AGE_YEARS
    (sources (INCORPORATION_DATE))
    (rule "(- (year (now)) (year INCORPORATION_DATE))")
    (jurisdiction GLOBAL)
  )
)
```

## Rule Expression Language

### Boolean Rules

Check conditions and return true/false:

```lisp
; Single condition
(if (in TAX_RESIDENCY_COUNTRY ['IR' 'KP' 'SY']) true false)

; Multiple conditions
(if (or (= PEP_STATUS true) (= SANCTIONS_STATUS true)) true false)

; Comparison
(if (> UBO_PERCENT 75) true false)
```

### Numeric Rules

Compute numeric values:

```lisp
; Maximum value
(max UBO_PERCENT)

; Sum
(sum UBO_PERCENT)

; Arithmetic
(- (year (now)) (year INCORPORATION_DATE))

; Conditional scoring
(+ (if HIGH_RISK_JURISDICTION_FLAG 30 0)
   (if PEP_EXPOSURE_FLAG 40 0))
```

### String Rules

Transform or lookup strings:

```lisp
; Registry lookup (pseudo-code)
(registry-active? REGISTERED_NAME INCORPORATION_JURISDICTION)

; String operations
(uppercase REGISTERED_NAME)
(concat FIRST_NAME " " LAST_NAME)
```

### Case-Based Rules

Map values to scores:

```lisp
(case TAX_RESIDENCY_COUNTRY
  (['IR' 'KP' 'SY'] 100)
  (['AF' 'YE' 'MM'] 90)
  (['RU' 'BY'] 80)
  (['CN' 'HK'] 60)
  (['US' 'GB' 'SG'] 20)
  (['CH' 'DE' 'FR'] 10)
  (else 50))
```

## Complete Example

```lisp
(kyc-case BLACKROCK-GLOBAL-EQUITY-FUND
  (nature-purpose
    (nature "Institutional Fund Management")
    (purpose "EU Equity Fund KYC with Risk Assessment"))

  (client-business-unit FUND-SERVICES-EU)

  (policy KYCPOL-EU-2025)

  ; Public attributes with their document sources
  (data-dictionary
    (attribute REGISTERED_NAME
      (primary-source (document CERT-INC)))
    (attribute TAX_RESIDENCY_COUNTRY
      (primary-source (document W8BENE)))
    (attribute UBO_NAME
      (primary-source (document UBO-DECL)))
    (attribute UBO_PERCENT
      (primary-source (document UBO-DECL)))
    (attribute PEP_STATUS
      (primary-source (document UBO-DECL)))
  )

  ; Private (derived) attributes with lineage and rules
  (derived-attributes
    (attribute HIGH_RISK_JURISDICTION_FLAG
      (sources (TAX_RESIDENCY_COUNTRY))
      (rule "(if (in TAX_RESIDENCY_COUNTRY ['IR' 'KP' 'SY']) true false)")
      (jurisdiction GLOBAL)
      (regulation AMLD5)
    )

    (attribute PEP_EXPOSURE_FLAG
      (sources (PEP_STATUS))
      (rule "(if (= PEP_STATUS true) true false)")
      (jurisdiction GLOBAL)
      (regulation AMLD5)
    )

    (attribute UBO_CONCENTRATION_SCORE
      (sources (UBO_PERCENT))
      (rule "(max UBO_PERCENT)")
      (jurisdiction GLOBAL)
      (regulation AMLD5)
    )
  )

  (kyc-token "pending")
)
```

## Validation

The system validates:

### 1. Attribute Existence
- Derived attribute must exist in `kyc_attributes` table
- Derived attribute must have `attribute_class='Private'`

### 2. Source Validation
- All source attributes must exist in `kyc_attributes` table
- All source attributes must have `attribute_class='Public'`
- At least one source attribute required

### 3. Rule Expression
- Rule expression cannot be empty
- Rule syntax should be valid (future: parse validation)

### 4. Regulation References
- If specified, regulation code must exist in `kyc_regulations`

**Error Examples:**

```
❌ derived attribute 'HIGH_RISK_JURISDICTION_FLAG' not found in ontology
❌ derived attribute 'ENTITY_AGE_YEARS' must have attribute_class='Private', got 'Public'
❌ derived attribute 'UBO_CONCENTRATION_SCORE' references unknown source 'FAKE_ATTRIBUTE'
❌ derived attribute 'RISK_FLAG' source 'DERIVED_OTHER' must be Public, got 'Private'
❌ derived attribute 'STATUS_FLAG' missing rule expression
```

## Seeded Private Attributes

The system includes pre-seeded private attributes:

### Risk Flags
- `HIGH_RISK_JURISDICTION_FLAG` - Client operates in high-risk jurisdiction
- `SANCTIONED_COUNTRY_FLAG` - Exposure to sanctioned countries
- `PEP_EXPOSURE_FLAG` - Beneficial owner is PEP
- `COMPLEX_STRUCTURE_FLAG` - Ownership structure with >5 layers

### Computed Scores
- `UBO_CONCENTRATION_SCORE` - Percentage held by largest UBO
- `ENTITY_AGE_YEARS` - Years since incorporation
- `JURISDICTION_RISK_SCORE` - Numeric risk score for country (0-100)
- `OVERALL_RISK_RATING` - Composite risk from all factors

### Status Indicators
- `ENTITY_ACTIVE_STATUS` - Active in business registry
- `DOCUMENT_COMPLETENESS_FLAG` - All required documents present
- `DATA_QUALITY_SCORE` - Percentage of required attributes populated

## Querying Lineage

### Find All Derived Attributes

```sql
SELECT code, name, domain
FROM kyc_attributes
WHERE attribute_class = 'Private'
ORDER BY domain, code;
```

### Trace Derivation Sources

```sql
SELECT
    derived_attribute,
    derived_attribute_name,
    source_attribute,
    source_attribute_name,
    rule_expression
FROM attribute_lineage
WHERE derived_attribute = 'HIGH_RISK_JURISDICTION_FLAG';
```

### Find Attributes Using a Public Source

```sql
SELECT derived_attribute_code, rule_expression
FROM kyc_attribute_derivations
WHERE source_attribute_code = 'TAX_RESIDENCY_COUNTRY';
```

### Multi-Source Derivations

```sql
SELECT
    derived_attribute_code,
    COUNT(*) as source_count,
    STRING_AGG(source_attribute_code, ', ') as sources
FROM kyc_attribute_derivations
GROUP BY derived_attribute_code
HAVING COUNT(*) > 1;
```

## Go API

### Repository Methods

```go
import "github.com/adamtc007/KYC-DSL/internal/ontology"

repo := ontology.NewRepository(db)

// List public attributes
publicAttrs, _ := repo.ListPublicAttributes()

// List private (derived) attributes
privateAttrs, _ := repo.ListPrivateAttributes()

// Get derivation rules for a private attribute
derivations, _ := repo.GetAttributeDerivations("HIGH_RISK_JURISDICTION_FLAG")

// Get full lineage view
lineage, _ := repo.GetAttributeLineage("UBO_CONCENTRATION_SCORE")

// Validate derivation sources
err := repo.ValidateDerivationSources([]string{"TAX_RESIDENCY_COUNTRY", "UBO_NAME"})
```

### Add Custom Derivation

```go
derivation := ontology.AttributeDerivation{
    DerivedAttributeCode: "CUSTOM_RISK_FLAG",
    SourceAttributeCode: "BUSINESS_ACTIVITY",
    RuleExpression: "(if (in BUSINESS_ACTIVITY ['CRYPTO' 'GAMBLING']) true false)",
    RuleType: "Boolean",
    Description: "Flags high-risk business activities",
}
err := repo.InsertAttributeDerivation(derivation)
```

## Migration & Setup

### 1. Run Schema Migration

```bash
psql -d kyc_dsl -f internal/storage/migrations/004_attribute_derivations.sql
```

### 2. Load Seed Data

```bash
psql -d kyc_dsl -f internal/ontology/seeds/derived_attributes_seed.sql
```

### 3. Verify Installation

```sql
-- Count public vs private attributes
SELECT attribute_class, COUNT(*) 
FROM kyc_attributes 
GROUP BY attribute_class;

-- Show derivation lineage
SELECT * FROM attribute_lineage LIMIT 10;
```

## Use Cases

### 1. Risk Scoring
Compute risk scores transparently from public attributes:
- Jurisdiction risk
- PEP exposure
- Sanctions screening
- Ownership concentration

### 2. Compliance Checks
Automated compliance flags:
- High-risk jurisdiction detection
- Complex structure identification
- Document completeness validation

### 3. Quality Metrics
Data quality indicators:
- Completeness scores
- Validation status
- Data freshness

### 4. AI Explainability
Enable AI agents to:
- Trace how flags were computed
- Explain risk decisions
- Validate computations
- Suggest improvements

## Benefits Summary

| Capability | What It Enables |
|-----------|----------------|
| **Data Lineage** | Every private value traceable to public, document-sourced data |
| **Configurability** | Derived attributes defined by declarative rules, not code |
| **Auditability** | Auditors can reconstruct how a flag was computed |
| **Explainability** | AI agents can reason about derivations |
| **Cross-Jurisdiction** | Same pattern, different rules per jurisdiction |
| **Versioning** | Rules stored in DSL, version-controlled with cases |
| **Transparency** | No "black box" — all logic explicit and queryable |

## Future Enhancements

1. **Rule Engine**: Full interpreter for rule expressions
2. **Real-Time Computation**: Compute derived attributes on case load
3. **Rule Validation**: Parse and validate rule syntax
4. **Visual Lineage**: GraphQL API for lineage visualization
5. **Rule Versioning**: Track changes to derivation rules over time
6. **Conditional Rules**: Different rules based on case context
7. **Cross-Attribute Rules**: Derive from other derived attributes (with cycle detection)

## Related Documentation

- `REGULATORY_ONTOLOGY.md` - Ontology structure
- `ONTOLOGY_VALIDATION.md` - Validation rules
- `CLAUDE.md` - Call trees and architecture
- `derived_attributes_example.dsl` - Complete working example

---

**Version**: 1.3  
**Status**: Production Ready  
**Grammar**: Extended with derived-attributes support