# Validation Audit Trail

## Overview

The Validation Audit Trail system provides comprehensive logging of all validation attempts for regulatory compliance and audit purposes. This system is designed to meet requirements under:

- **FCA SYSC** (Senior Management Arrangements, Systems and Controls)
- **MAS 626 §4.2** (Record-keeping requirements)
- **HKMA AML §3.6** (Audit trail and documentation)
- **EU AMLD6 Article 30** (Record-keeping obligations)

## Key Features

- ✅ **Complete Audit Trail** - Every validation attempt is recorded
- ✅ **Actor Attribution** - Tracks who/what performed validation (System, User, Agent)
- ✅ **Detailed Findings** - Individual check results with severity levels
- ✅ **Version Tracking** - Links validations to specific case versions
- ✅ **Compliance Reports** - Pre-built views for regulatory reporting
- ✅ **Immutable Records** - Timestamped entries cannot be modified

## Database Schema

### Core Tables

#### `kyc_case_validations`
Main audit table recording each validation attempt.

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key |
| `case_name` | TEXT | Case identifier |
| `version` | INT | Case version number |
| `validation_time` | TIMESTAMP | When validation occurred |
| `grammar_version` | TEXT | DSL grammar version used |
| `ontology_version` | TEXT | Ontology version used |
| `validator_actor` | TEXT | Who performed validation |
| `validation_status` | TEXT | PASS or FAIL |
| `error_message` | TEXT | Error details if failed |
| `total_checks` | INT | Number of checks performed |
| `passed_checks` | INT | Number of checks passed |
| `failed_checks` | INT | Number of checks failed |
| `metadata` | JSONB | Additional context |
| `created_at` | TIMESTAMP | Record creation time |

#### `kyc_validation_findings`
Detailed findings for each validation check.

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key |
| `validation_id` | INT | Links to validation record |
| `check_type` | TEXT | Type of check (ontology_document, ownership_sum, etc.) |
| `check_name` | TEXT | Name of specific check |
| `check_status` | TEXT | PASS, WARN, or FAIL |
| `check_message` | TEXT | Result message |
| `entity_ref` | TEXT | Reference to entity checked |
| `severity` | TEXT | INFO, WARNING, ERROR, CRITICAL |
| `created_at` | TIMESTAMP | Record creation time |

### Views

#### `validation_summary`
Aggregated statistics per case.

```sql
SELECT
    case_name,
    COUNT(*) as total_validations,
    SUM(CASE WHEN validation_status = 'PASS' THEN 1 ELSE 0 END) as passed,
    SUM(CASE WHEN validation_status = 'FAIL' THEN 1 ELSE 0 END) as failed,
    MAX(validation_time) as last_validation,
    AVG(total_checks) as avg_checks_per_validation
FROM kyc_case_validations
GROUP BY case_name;
```

#### `compliance_audit_trail`
Complete audit view for regulatory reports.

```sql
SELECT
    v.case_name,
    v.version,
    v.validation_time,
    v.validator_actor,
    v.validation_status,
    v.grammar_version,
    v.ontology_version,
    COUNT(f.id) as finding_count,
    SUM(CASE WHEN f.severity = 'CRITICAL' THEN 1 ELSE 0 END) as critical_findings,
    SUM(CASE WHEN f.severity = 'ERROR' THEN 1 ELSE 0 END) as error_findings
FROM kyc_case_validations v
LEFT JOIN kyc_validation_findings f ON f.validation_id = v.id
GROUP BY v.id, v.case_name, v.version...
ORDER BY v.validation_time DESC;
```

## Usage

### CLI Command

Validate an existing case and record audit trail:

```bash
./kycctl validate CASE-NAME
```

With custom actor:

```bash
./kycctl validate CASE-NAME --actor="User:adam"
```

### Example Output

**Success:**
```
✔ Case BLACKROCK-GLOBAL-EQUITY-FUND passed ontology validation
   - All 5 documents valid
   - All 3 attributes resolved
   - Policy references validated
✅ Case BLACKROCK-GLOBAL-EQUITY-FUND validated successfully (3/3 checks passed)
   Audit trail recorded (actor: System)
✅ Case BLACKROCK-GLOBAL-EQUITY-FUND validated and audit logged.
```

**Failure:**
```
❌ ontology reference validation failed: unknown document code 'W8BENZ' in jurisdiction 'EU'
```

### Programmatic Usage

```go
import (
    "github.com/adamtc007/KYC-DSL/internal/parser"
    "github.com/adamtc007/KYC-DSL/internal/storage"
)

// Validate with audit
db, _ := storage.ConnectPostgres()
defer db.Close()

err := parser.ValidateCaseWithAudit(db, kycCase, "System")
if err != nil {
    // Validation failed - audit already recorded
    log.Fatal(err)
}

// Retrieve validation history
history, _ := storage.GetValidationHistory(db, "CASE-NAME")
for _, v := range history {
    fmt.Printf("%s: %s (%d/%d checks passed)\n",
        v.ValidationTime, v.ValidationStatus,
        v.PassedChecks, v.TotalChecks)
}
```

## Validation Checks

The system performs three types of checks:

### 1. Structure Validation
- Nature and purpose sections present
- Client business unit defined
- KYC token present

### 2. Semantic Validation
- Function names are valid
- Policy codes exist in registry
- Token status is valid
- Ownership percentages sum correctly
- Controllers specified when required

### 3. Ontology Validation
- Document codes exist in ontology
- Attribute codes exist in ontology
- Documents linked to regulations
- Jurisdictions specified

## Actor Types

The `validator_actor` field identifies who/what performed the validation:

| Actor | Description | Example |
|-------|-------------|---------|
| `System` | Automated system validation | `System` |
| `User:name` | Manual validation by user | `User:adam` |
| `Agent:name` | AI agent validation | `Agent:Claude` |
| `Service:name` | External service validation | `Service:ComplianceBot` |

## Query Examples

### Recent Validation History

```sql
SELECT case_name, version, validation_status, validator_actor, validation_time
FROM kyc_case_validations
WHERE case_name = 'BLACKROCK-GLOBAL-EQUITY-FUND'
ORDER BY validation_time DESC
LIMIT 10;
```

### Failed Validations Report

```sql
SELECT case_name, version, error_message, validator_actor, validation_time
FROM kyc_case_validations
WHERE validation_status = 'FAIL'
ORDER BY validation_time DESC;
```

### Validation Success Rate

```sql
SELECT
    case_name,
    COUNT(*) as total_attempts,
    SUM(CASE WHEN validation_status = 'PASS' THEN 1 ELSE 0 END) as passed,
    ROUND(100.0 * SUM(CASE WHEN validation_status = 'PASS' THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate
FROM kyc_case_validations
GROUP BY case_name
ORDER BY success_rate DESC;
```

### Critical Findings

```sql
SELECT v.case_name, v.version, f.check_type, f.check_message, f.entity_ref
FROM kyc_validation_findings f
JOIN kyc_case_validations v ON v.id = f.validation_id
WHERE f.severity = 'CRITICAL'
ORDER BY f.created_at DESC;
```

### Daily Validation Volume

```sql
SELECT
    DATE(validation_time) as date,
    COUNT(*) as total_validations,
    SUM(CASE WHEN validation_status = 'PASS' THEN 1 ELSE 0 END) as passed,
    SUM(CASE WHEN validation_status = 'FAIL' THEN 1 ELSE 0 END) as failed
FROM kyc_case_validations
GROUP BY DATE(validation_time)
ORDER BY date DESC;
```

## Compliance Reporting

### Regulatory Audit Report

For FCA, MAS, HKMA audits:

```sql
SELECT
    case_name,
    version,
    validation_time,
    validator_actor,
    validation_status,
    grammar_version,
    ontology_version,
    total_checks,
    passed_checks,
    failed_checks,
    error_message
FROM kyc_case_validations
WHERE validation_time >= '2024-01-01'
ORDER BY validation_time DESC;
```

Export to CSV:

```bash
psql -d kyc_dsl -c "COPY (
    SELECT * FROM compliance_audit_trail
    WHERE validation_time >= '2024-01-01'
) TO '/tmp/compliance_audit_2024.csv' CSV HEADER;"
```

## Data Retention

Validation records should be retained according to regulatory requirements:

- **MAS 626**: Minimum 5 years
- **HKMA**: Minimum 5 years
- **EU AMLD6**: Minimum 5 years
- **FCA**: Minimum 6 years (financial promotions)

Archive old records:

```sql
-- Create archive table
CREATE TABLE kyc_case_validations_archive (LIKE kyc_case_validations);

-- Archive records older than 7 years
INSERT INTO kyc_case_validations_archive
SELECT * FROM kyc_case_validations
WHERE validation_time < NOW() - INTERVAL '7 years';

-- Delete archived records (only after backup!)
DELETE FROM kyc_case_validations
WHERE validation_time < NOW() - INTERVAL '7 years';
```

## Migration

Initialize the audit trail schema:

```bash
psql -d kyc_dsl -f internal/storage/migrations/002_validation_audit.sql
```

## Benefits

### For Compliance
- **Audit Trail**: Complete record of all validation attempts
- **Accountability**: Tracks who performed validation
- **Evidence**: Demonstrates due diligence to regulators
- **Traceability**: Links validations to specific case versions

### For Operations
- **Debugging**: Identify recurring validation failures
- **Monitoring**: Track validation success rates over time
- **Quality Metrics**: Measure system health
- **Root Cause Analysis**: Detailed findings per validation

### For Development
- **Testing**: Verify validation logic works correctly
- **Regression Detection**: Catch validation regressions
- **Performance**: Track validation execution time
- **Coverage**: Ensure all checks are running

## Error Handling

All validation functions include proper error handling:

```go
// Database connection is checked
if db == nil {
    return fmt.Errorf("database connection is nil")
}

// Parameters are validated
if caseName == "" {
    return fmt.Errorf("case name is required")
}

// Errors are wrapped with context
if err := db.Exec(...); err != nil {
    return fmt.Errorf("record validation result failed (case=%s): %w", caseName, err)
}
```

Connection cleanup is guaranteed:

```go
defer func() {
    if closeErr := db.Close(); closeErr != nil {
        log.Printf("WARNING: failed to close database: %v", closeErr)
    }
}()
```

## Future Enhancements

1. **Real-time Alerts**: Notify on validation failures
2. **Validation Metrics Dashboard**: Grafana/Prometheus integration
3. **Automated Remediation**: Auto-fix common validation errors
4. **Machine Learning**: Predict validation failures before they occur
5. **Blockchain Anchoring**: Tamper-proof audit trail via blockchain
6. **Multi-tenant Support**: Separate audit trails per organization

## Related Documentation

- `REGULATORY_ONTOLOGY.md` - Ontology validation details
- `ONTOLOGY_VALIDATION.md` - Validation rules and checks
- `CLAUDE.md` - Call trees showing validation flow
- `QUICKSTART.md` - Getting started with validation

---

**Version**: 1.2  
**Status**: Production Ready  
**Compliance**: FCA SYSC, MAS 626 §4.2, HKMA AML §3.6, EU AMLD6 Article 30