# Ontology Validation

## Overview

The Ontology Validation system ensures that all DSL references (documents, attributes, regulations) correspond to valid entries in the regulatory data ontology. This provides **compile-time safety** for compliance requirements.

## What It Validates

### 1. Document References

All document codes in `(data-dictionary ...)` and `(document-requirements ...)` must exist in `kyc_documents` table.

**Validates:**
- Primary source documents
- Secondary source documents  
- Tertiary source documents
- Required documents by jurisdiction

**Example Error:**
```
❌ unknown document code 'W8BENZ' in jurisdiction 'EU'
```

### 2. Attribute References

All attribute codes in `(data-dictionary ...)` must exist in `kyc_attributes` table.

**Example Error:**
```
❌ unknown attribute 'FAKE_ATTRIBUTE_XYZ' in data-dictionary
```

### 3. Document-Regulation Linkage

All required documents must be linked to at least one regulation via `kyc_doc_reg_links` table.

**Example Error:**
```
❌ document 'CUSTOM-DOC' not linked to any regulation in ontology
```

### 4. Jurisdiction Completeness

All `(document-requirements ...)` sections must specify a jurisdiction.

**Example Error:**
```
❌ document-requirements section missing jurisdiction
```

## Validation Flow

```
DSL File → Parse → Bind → Validate Syntax → Validate Semantics → Validate Ontology → Persist
                                                                        ↑
                                                           checks against kyc_* tables
```

The validation runs **after** syntax and semantic checks but **before** version persistence.

## Success Output

When all references are valid:

```
✔ Case BLACKROCK-GLOBAL-EQUITY-FUND passed ontology validation
   - All 5 documents valid
   - All 3 attributes resolved
   - Policy references validated
```

## Implementation

### File Structure

**`internal/parser/validator_ontology.go`** - Main validation logic

**`internal/ontology/repository.go`** - Helper methods:
- `AllDocumentCodes()` - Returns all valid document codes
- `AllAttributeCodes()` - Returns all valid attribute codes
- `AllRegulationCodes()` - Returns all valid regulation codes
- `DocumentLinkedToRegulation(code)` - Checks if document has regulation link

### Integration Point

In `internal/parser/validator.go`:

```go
func ValidateDSL(db *sqlx.DB, cases []*model.KycCase, ebnf string) error {
    for _, c := range cases {
        // 1. Structure validation
        if err := validateCaseStructure(c); err != nil { ... }
        
        // 2. Semantic validation
        if err := validateCaseSemantics(db, c); err != nil { ... }
        
        // 3. Ontology validation (NEW)
        if err := ValidateOntologyRefs(db, c); err != nil { ... }
    }
}
```

## Test Cases

### Valid Case

**File:** `ontology_example.dsl`

Uses valid codes:
- `CERT-INC` (Certificate of Incorporation) ✓
- `UBO-DECL` (UBO Declaration) ✓
- `REGISTERED_NAME` attribute ✓
- `UBO_NAME` attribute ✓

### Invalid Document Code

**File:** `test_invalid_ontology_doc.dsl`

Contains:
```lisp
(data-dictionary
  (attribute UBO_NAME
    (primary-source (document W8BENZ)))  ; ← Typo: should be W8BEN
)
```

**Expected Error:**
```
❌ ontology reference validation failed: unknown primary-source document 'W8BENZ' for attribute 'UBO_NAME'
```

### Invalid Attribute Code

**File:** `test_invalid_ontology_attr.dsl`

Contains:
```lisp
(data-dictionary
  (attribute FAKE_ATTRIBUTE_XYZ    ; ← Does not exist
    (primary-source (document UBO-DECL)))
)
```

**Expected Error:**
```
❌ ontology reference validation failed: unknown attribute 'FAKE_ATTRIBUTE_XYZ' in data-dictionary
```

## Running Validation Tests

### Manual Testing

```bash
# Valid case (should succeed)
./bin/kycctl ontology_example.dsl

# Invalid document (should fail)
./bin/kycctl test_invalid_ontology_doc.dsl

# Invalid attribute (should fail)
./bin/kycctl test_invalid_ontology_attr.dsl
```

### Automated Test Suite

```bash
./test_ontology_validation.sh
```

This runs all validation scenarios and reports:
- ✓ Valid case accepted with detailed feedback
- ✓ Invalid document code detected
- ✓ Invalid attribute code detected

## Benefits

### 1. Early Error Detection

Catches typos and invalid references at parse time, not runtime.

**Before:**
```
Case processed → Stored in DB → Execution fails → Manual investigation
```

**After:**
```
Case rejected at validation → Clear error message → Fix and retry
```

### 2. Ontology Integrity

Ensures DSL cases only reference data that exists in the regulatory ontology.

### 3. Compliance Safety

Prevents cases from referencing non-existent regulations or unsupported documents.

### 4. Developer Experience

Clear, actionable error messages:
- Shows which code is invalid
- Shows where it appears (attribute, document, tier)
- Suggests the correct format

## Case Sensitivity

All validation is **case-insensitive** to handle variations:
- `W8BEN` = `w8ben` = `W8Ben` ✓
- `UBO_NAME` = `ubo_name` = `Ubo_Name` ✓

Internal storage uses uppercase for consistency.

## Extending Validation

### Add New Check

Edit `internal/parser/validator_ontology.go`:

```go
func ValidateOntologyRefs(db *sqlx.DB, c *model.KycCase) error {
    // ... existing checks ...
    
    // NEW: Check document expiry
    for _, dr := range c.DocumentRequirements {
        for _, d := range dr.Documents {
            doc, _ := repo.GetDocumentByCode(d.Code)
            if doc.ValidityYears > 0 {
                // Warn about documents with expiration
            }
        }
    }
}
```

### Add New Repository Method

Edit `internal/ontology/repository.go`:

```go
func (r *Repository) GetDocumentsExpiringSoon() ([]Document, error) {
    // Custom query for expiring documents
}
```

## Performance

Validation is efficient:
- **3 SQL queries** to load all codes (cached per validation run)
- **Map lookups** (O(1)) for each reference
- **~1-5ms** for typical case with 10 references

## Error Message Reference

| Error | Cause | Solution |
|-------|-------|----------|
| `unknown document code 'X'` | Document not in ontology | Check `kyc_documents` table or fix typo |
| `unknown attribute 'X'` | Attribute not in ontology | Check `kyc_attributes` table or fix typo |
| `document 'X' not linked to regulation` | Missing `kyc_doc_reg_links` entry | Add regulation link in seed data |
| `missing jurisdiction` | No `(jurisdiction X)` in section | Add jurisdiction to `document-requirements` |

## SQL Queries for Debugging

### Check if document exists:
```sql
SELECT code, name FROM kyc_documents WHERE code = 'W8BEN';
```

### Check if attribute exists:
```sql
SELECT code, name FROM kyc_attributes WHERE code = 'UBO_NAME';
```

### Check document-regulation links:
```sql
SELECT dr.document_code, dr.regulation_code, r.name
FROM kyc_doc_reg_links dr
JOIN kyc_regulations r ON r.code = dr.regulation_code
WHERE dr.document_code = 'UBO-DECL';
```

### View all valid codes:
```sql
SELECT 'Document' as type, code FROM kyc_documents
UNION ALL
SELECT 'Attribute', code FROM kyc_attributes
UNION ALL
SELECT 'Regulation', code FROM kyc_regulations
ORDER BY type, code;
```

## Future Enhancements

1. **Jurisdiction Validation**: Check if jurisdiction is valid (US, EU, SG, HK, etc.)
2. **Tier Validation**: Warn if no primary source specified
3. **Regulation Applicability**: Validate document-regulation combinations match jurisdiction
4. **Deprecation Warnings**: Flag documents marked as deprecated in ontology
5. **Completeness Checks**: Suggest missing required documents for a jurisdiction

## Related Documentation

- `REGULATORY_ONTOLOGY.md` - Ontology structure and content
- `CLAUDE.md` - Call trees showing validation flow
- `QUICKSTART.md` - Testing validation examples

---

**Version**: 1.2  
**Status**: Production Ready  
**Validation Scope**: Documents, Attributes, Regulations, Jurisdictions