# KYC-DSL Amendment System

## Overview

The Amendment System enables incremental, auditable evolution of KYC cases through a defined lifecycle. Each amendment is a deterministic change that loads the latest DSL version, applies a mutation, validates, and saves as a new version.

## Architecture

```
┌──────────────────────────────┐
│ Initial DSL creation         │
│ (nature, purpose, CBU)       │
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Policy discovery phase       │
│ (append policies + functions)│
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Document solicitation phase  │
│ (append obligations)         │
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Ownership & Control mapping  │
│ (append ownership-structure) │
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Risk & Approval              │
│ (append assess-risk, token)  │
└──────────────────────────────┘
```

## Package Structure

```
internal/amend/
├── amend.go          - Core engine: load → mutate → validate → save → version
├── mutations.go      - Predefined case evolution steps
└── transitions.go    - Lifecycle phase definitions and allowed transitions
```

## Lifecycle Phases

### Phase 1: Case Creation
**DSL Additions:** `kyc-case`, `nature-purpose`, `client-business-unit`

Initial case setup with basic information.

```lisp
(kyc-case CASE-NAME
  (nature-purpose
    (nature "Description")
    (purpose "Purpose statement")
  )
  (client-business-unit UNIT-NAME)
  (kyc-token "pending")
)
```

### Phase 2: Policy Discovery
**DSL Additions:** `function DISCOVER-POLICIES`, multiple `policy` nodes

Discovers which policies apply based on product & jurisdiction.

**Command:**
```bash
kycctl amend CASE-NAME --step=policy-discovery
```

**Changes:**
- Adds `DISCOVER-POLICIES` function
- Injects policy codes (KYCPOL-UK-2025, KYCPOL-EU-2025, AML-GLOBAL-BASE)

### Phase 3: Document Solicitation
**DSL Additions:** `function SOLICIT-DOCUMENTS`, multiple `obligation` nodes

Requests proofs (W8/W9, UBO declarations, etc.).

**Command:**
```bash
kycctl amend CASE-NAME --step=document-solicitation
```

**Changes:**
- Adds `SOLICIT-DOCUMENTS` function
- Adds obligations (OBL-W8BEN, OBL-W9, OBL-UBO-DECLARATION, OBL-PEP-001)

### Phase 4: Ownership & Control
**DSL Additions:** `function BUILD-OWNERSHIP-TREE`, `function VERIFY-OWNERSHIP`, ownership structure

Builds legal & beneficial ownership graph + operational control roles.

**Command:**
```bash
kycctl amend CASE-NAME --step=ownership-discovery
```

**Changes:**
- Adds `BUILD-OWNERSHIP-TREE` function
- Adds `VERIFY-OWNERSHIP` function
- Creates ownership structure with:
  - Legal owners (registered shareholders)
  - Beneficial owners (economic interest/voting rights)
  - Controllers (significant control/influence)
  - Operational roles (key management personnel)

**Ownership Types:**

| Type | Description | Example |
|------|-------------|---------|
| Legal ownership | Registered shareholders or beneficial owners (>25%) | BLACKROCK-PLC 100% |
| Beneficial ownership | Economic interest or voting rights without legal title | LARRY-FINK 35% voting rights |
| Significant control | Natural persons with operational/managerial authority | JANE-DOE "Senior Managing Official" |
| Operational management | Key control persons in day-to-day management | MARY-JONES "Chief Compliance Officer" |

### Phase 5: Risk & Review
**DSL Additions:** `function ASSESS-RISK`, optionally `function REGULATOR-NOTIFY`

Computes KYC score; escalates or approves.

**Commands:**
```bash
kycctl amend CASE-NAME --step=risk-assessment
kycctl amend CASE-NAME --step=regulator-notify  # optional
```

**Changes:**
- Adds `ASSESS-RISK` function
- Optionally adds `REGULATOR-NOTIFY` function

### Phase 6: Finalization
**DSL Additions:** `kyc-token` status update

Token issuance / completion.

**Commands:**
```bash
kycctl amend CASE-NAME --step=approve   # Token: "approved"
kycctl amend CASE-NAME --step=decline   # Token: "declined"
kycctl amend CASE-NAME --step=review    # Token: "review"
```

**Changes:**
- Updates `kyc-token` status
- Marks case as complete

## Amendment Engine Flow

Each amendment follows this deterministic process:

1. **Load** - Fetch latest serialized DSL from `kyc_case_versions`
2. **Parse** - Convert DSL text to AST
3. **Bind** - Transform AST to typed `model.KycCase`
4. **Mutate** - Apply user-provided mutation function
5. **Validate** - Check grammar and semantic rules
6. **Serialize** - Convert model back to DSL text
7. **Version** - Save as next version with SHA-256 hash
8. **Log** - Record amendment in `kyc_case_amendments`

## Database Schema

### `kyc_case_versions`
Stores complete DSL snapshots for each version.

```sql
CREATE TABLE kyc_case_versions (
    id SERIAL PRIMARY KEY,
    case_name TEXT NOT NULL,
    version INT NOT NULL,
    dsl_snapshot TEXT,
    hash TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### `kyc_case_amendments`
Audit trail of all amendments.

```sql
CREATE TABLE kyc_case_amendments (
    id SERIAL PRIMARY KEY,
    case_name TEXT NOT NULL,
    step TEXT NOT NULL,
    change_type TEXT NOT NULL,
    diff TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Usage Examples

### Complete Lifecycle Example

```bash
# Initial case creation
kycctl sample_case.dsl

# Phase 1: Discover policies
kycctl amend AVIVA-EU-EQUITY-FUND --step=policy-discovery

# Phase 2: Solicit documents
kycctl amend AVIVA-EU-EQUITY-FUND --step=document-solicitation

# Phase 3: Build ownership structure
kycctl amend AVIVA-EU-EQUITY-FUND --step=ownership-discovery

# Phase 4: Risk assessment
kycctl amend AVIVA-EU-EQUITY-FUND --step=risk-assessment

# Phase 5: Finalize (approve)
kycctl amend AVIVA-EU-EQUITY-FUND --step=approve
```

### View Amendment History

```sql
-- View all amendments for a case
SELECT step, change_type, created_at 
FROM kyc_case_amendments 
WHERE case_name = 'AVIVA-EU-EQUITY-FUND' 
ORDER BY created_at;

-- View version history with hashes
SELECT version, LEFT(hash, 12) as hash, created_at 
FROM kyc_case_versions 
WHERE case_name = 'AVIVA-EU-EQUITY-FUND' 
ORDER BY version;

-- View specific version snapshot
SELECT dsl_snapshot 
FROM kyc_case_versions 
WHERE case_name = 'AVIVA-EU-EQUITY-FUND' 
AND version = 7;
```

## Predefined Mutations

All mutations are defined in `internal/amend/mutations.go`:

### Policy Discovery
```go
func AddPolicyDiscovery(c *model.KycCase) {
    c.Functions = append(c.Functions, model.Function{
        Action: "DISCOVER-POLICIES",
        Status: model.Pending,
    })
    c.Policies = append(c.Policies,
        model.KycPolicy{Code: "KYCPOL-UK-2025"},
        model.KycPolicy{Code: "KYCPOL-EU-2025"},
        model.KycPolicy{Code: "AML-GLOBAL-BASE"},
    )
}
```

### Document Solicitation
```go
func AddDocumentSolicitation(c *model.KycCase) {
    c.Functions = append(c.Functions, model.Function{
        Action: "SOLICIT-DOCUMENTS",
        Status: model.Pending,
    })
    c.Obligations = append(c.Obligations,
        model.KycObligation{PolicyCode: "OBL-W8BEN"},
        model.KycObligation{PolicyCode: "OBL-W9"},
        model.KycObligation{PolicyCode: "OBL-UBO-DECLARATION"},
        model.KycObligation{PolicyCode: "OBL-PEP-001"},
    )
}
```

### Ownership Structure
```go
func AddOwnershipStructure(c *model.KycCase) {
    c.Functions = append(c.Functions,
        model.Function{Action: "BUILD-OWNERSHIP-TREE", Status: model.Pending},
        model.Function{Action: "VERIFY-OWNERSHIP", Status: model.Pending},
    )
    
    if c.Ownership == nil {
        c.Ownership = &model.OwnershipStructure{Entity: c.Name}
    }
    
    // Add legal owners
    c.Ownership.LegalOwners = append(c.Ownership.LegalOwners,
        model.Owner{Name: "BLACKROCK-PLC", Percentage: 100.0},
    )
    
    // Add beneficial owners
    c.Ownership.BeneficialOwners = append(c.Ownership.BeneficialOwners,
        model.BeneficialOwner{
            Name: "LARRY-FINK",
            Percentage: 35.0,
            Interest: "voting rights",
        },
    )
    
    // Add controllers
    c.Ownership.Controllers = append(c.Ownership.Controllers,
        model.Controller{Name: "JANE-DOE", Role: "Senior Managing Official"},
        model.Controller{Name: "JOHN-SMITH", Role: "Director"},
    )
    
    // Add operational roles
    c.Ownership.OperationalRoles = append(c.Ownership.OperationalRoles,
        model.OperationalRole{
            Name: "MARY-JONES",
            Title: "Chief Compliance Officer",
            Function: "compliance oversight",
        },
    )
}
```

## Custom Mutations

You can create custom mutations by implementing the signature:

```go
func CustomMutation(c *model.KycCase) {
    // Make changes to the case
    c.Functions = append(c.Functions, model.Function{
        Action: "CUSTOM-FUNCTION",
        Status: model.Pending,
    })
}
```

Then use it via the API:

```go
err := amend.ApplyAmendment(db, caseName, "custom-step", CustomMutation)
```

## Validation

All amendments are validated before saving:

1. **Grammar validation** - DSL must match EBNF grammar
2. **Semantic validation** - Functions, policies, and tokens must be valid
3. **Policy validation** - Policy codes must exist in `kyc_policies` table
4. **Function validation** - Function names must be in whitelist

## Audit Trail Benefits

The amendment system provides:

✅ **Determinism** - Identical mutations produce identical DSL
✅ **Provenance** - Each change is timestamped and logged
✅ **Replayability** - Rebuild case state from any historical snapshot
✅ **Traceability** - Complete audit trail of all changes
✅ **Version control** - SHA-256 hashes detect duplicate versions
✅ **Compliance** - Immutable record of case evolution

## Integration with Existing System

The amendment system integrates seamlessly:

- **Parser** - Uses existing `parser.Parse()` and `parser.Bind()`
- **Serializer** - Uses `parser.SerializeCases()` for round-trip
- **Validator** - Uses `parser.ValidateDSL()` for checks
- **Storage** - Uses `storage.SaveCaseVersion()` for persistence
- **Engine** - Can be used alongside `engine.RunCase()`

## Future Enhancements

Planned features:

1. **Diff generation** - Semantic diffs between versions (`internal/amend/diff.go`)
2. **Transition validation** - Enforce allowed phase transitions
3. **Rollback support** - Restore to previous version
4. **Approval workflows** - Multi-step approval process
5. **Automated amendments** - Trigger amendments based on events
6. **Amendment templates** - Pre-configured mutation sets

## See Also

- `REFACTORING_SUMMARY.md` - CLI refactoring details
- `README.md` - Main project documentation
- `CLAUDE.md` - Project guidance for Claude AI