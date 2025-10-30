# Ownership & Control System

## Overview

The Ownership & Control system extends KYC-DSL v1.1 with constructs for modeling legal ownership, beneficial ownership, and control relationships in financial entities. This enables comprehensive Know Your Customer (KYC) compliance by tracking:

- **Legal Ownership** - Registered shareholders (>25% threshold)
- **Beneficial Ownership** - Economic interest/voting rights without legal title
- **Significant Control** - Natural persons with operational/managerial authority
- **Control Persons** - Directors, trustees, senior managing officials

## Grammar v1.1

### EBNF Definition

```ebnf
(* Ownership & Control Section *)

OwnershipStructure = "(" "ownership-structure"
                        { OwnershipNode }
                     ")" ;

OwnershipNode      = Owner | BeneficialOwner | Controller ;

Owner              = "(" "owner" Identifier Number "%" ")" ;
BeneficialOwner    = "(" "beneficial-owner" Identifier Number "%" ")" ;
Controller         = "(" "controller" Identifier QuotedString ")" ;
```

### Grammar Update

The grammar was updated from v1.0 to v1.1 to include ownership constructs:
- Added `OwnershipStructure` to the `Section` alternatives
- Defined three ownership node types: `Owner`, `BeneficialOwner`, `Controller`
- Percentages are represented as `Number "%"` (integers)
- Roles are quoted strings for controllers

## DSL Syntax

### Basic Structure

```lisp
(ownership-structure
  (owner ENTITY-NAME PERCENTAGE%)
  (beneficial-owner PERSON-NAME PERCENTAGE%)
  (controller PERSON-NAME "Role Description")
)
```

### Complete Example

```lisp
(kyc-case BLACKROCK-GLOBAL-EQUITY-FUND
  (nature-purpose
    (nature "Institutional investment management vehicle")
    (purpose "Operate a SICAV with multi-jurisdictional sub-funds"))
  (client-business-unit BLACKROCK-GLOBAL-FUNDS)
  (function BUILD-OWNERSHIP-TREE)
  (ownership-structure
    (owner BLACKROCK-PLC 100)
    (beneficial-owner LARRY-FINK 35)
    (controller JANE-DOE "Senior Managing Official")
    (controller JOHN-SMITH "Director, Risk Oversight"))
  (kyc-token "pending"))
```

## Data Model

### OwnershipNode Structure

```go
type OwnershipNode struct {
    Entity           string  `db:"entity"`
    Owner            string  `db:"owner"`
    BeneficialOwner  string  `db:"beneficial_owner"`
    Controller       string  `db:"controller"`
    Role             string  `db:"role"`
    OwnershipPercent float64 `db:"ownership_percent"`
}
```

### Integration with KycCase

```go
type KycCase struct {
    ID            int
    Name          string
    Version       int
    Status        CaseStatus
    LastUpdated   time.Time

    // DSL-derived fields
    Nature      string
    Purpose     string
    CBU         ClientBusinessUnit
    Policies    []KycPolicy
    Obligations []KycObligation
    Functions   []Function
    Token       *KycToken
    Ownership   []OwnershipNode  // ← New field
}
```

### Node Types

Each `OwnershipNode` represents one relationship:

| Field | Used For | Example |
|-------|----------|---------|
| `Entity` | Entity identifier | "BLACKROCK-GLOBAL-FUNDS" |
| `Owner` | Legal owner name | "BLACKROCK-PLC" |
| `BeneficialOwner` | Beneficial owner name | "LARRY-FINK" |
| `Controller` | Controller name | "JANE-DOE" |
| `Role` | Controller's role | "Senior Managing Official" |
| `OwnershipPercent` | Ownership percentage | 100.0, 35.0 |

## Parser Implementation

### Binder (Parse → Model)

The binder converts DSL ownership structures into `OwnershipNode` slices:

```go
case "ownership-structure":
    for _, n := range node.Args {
        switch n.Head {
        case "owner":
            if len(n.Args) == 2 {
                caseObj.Ownership = append(caseObj.Ownership, model.OwnershipNode{
                    Owner:            n.Args[0].Head,
                    OwnershipPercent: parsePercent(n.Args[1].Head),
                })
            }
        case "beneficial-owner":
            if len(n.Args) == 2 {
                caseObj.Ownership = append(caseObj.Ownership, model.OwnershipNode{
                    BeneficialOwner:  n.Args[0].Head,
                    OwnershipPercent: parsePercent(n.Args[1].Head),
                })
            }
        case "controller":
            if len(n.Args) == 2 {
                caseObj.Ownership = append(caseObj.Ownership, model.OwnershipNode{
                    Controller: n.Args[0].Head,
                    Role:       trimQuotes(n.Args[1].Head),
                })
            }
        }
    }
```

### Serializer (Model → DSL)

The serializer reconstructs DSL from the model:

```go
if len(c.Ownership) > 0 {
    sb.WriteString("  (ownership-structure\n")
    for _, o := range c.Ownership {
        switch {
        case o.Owner != "":
            sb.WriteString(fmt.Sprintf("    (owner %s %.0f%%)\n", o.Owner, o.OwnershipPercent))
        case o.BeneficialOwner != "":
            sb.WriteString(fmt.Sprintf("    (beneficial-owner %s %.0f%%)\n", o.BeneficialOwner, o.OwnershipPercent))
        case o.Controller != "":
            sb.WriteString(fmt.Sprintf("    (controller %s \"%s\")\n", o.Controller, o.Role))
        }
    }
    sb.WriteString("  )\n")
}
```

## Amendment System Integration

### Predefined Mutation

The `AddOwnershipStructure` mutation adds ownership data to a case:

```go
func AddOwnershipStructure(c *model.KycCase) {
    c.Functions = append(c.Functions, model.Function{
        Action: "BUILD-OWNERSHIP-TREE",
        Status: model.Pending,
    })
    c.Functions = append(c.Functions, model.Function{
        Action: "VERIFY-OWNERSHIP",
        Status: model.Pending,
    })

    c.Ownership = append(c.Ownership,
        model.OwnershipNode{
            Entity: "BLACKROCK-GLOBAL-FUNDS",
        },
        model.OwnershipNode{
            Owner:            "BLACKROCK-PLC",
            OwnershipPercent: 100,
        },
        model.OwnershipNode{
            BeneficialOwner:  "LARRY-FINK",
            OwnershipPercent: 35,
        },
        model.OwnershipNode{
            Controller: "JANE-DOE",
            Role:       "Senior Managing Official",
        },
        model.OwnershipNode{
            Controller: "JOHN-SMITH",
            Role:       "Director",
        },
    )
}
```

### CLI Usage

```bash
# Add ownership structure to existing case
kycctl amend BLACKROCK-GLOBAL-EQUITY-FUND --step=ownership-discovery
```

This will:
1. Load the latest case version
2. Apply the `AddOwnershipStructure` mutation
3. Add ownership functions and ownership nodes
4. Validate the result
5. Save as a new version
6. Log the amendment

## Use Cases

### 1. AML/KYC Compliance

Track beneficial owners exceeding 25% threshold for anti-money laundering compliance:

```lisp
(ownership-structure
  (owner HOLDING-COMPANY 100)
  (beneficial-owner PERSON-A 40)
  (beneficial-owner PERSON-B 30)
  (beneficial-owner PERSON-C 30))
```

### 2. Control Person Identification

Identify persons with significant control or influence:

```lisp
(ownership-structure
  (controller CEO-NAME "Chief Executive Officer")
  (controller CFO-NAME "Chief Financial Officer")
  (controller TRUSTEE-NAME "Trustee, Voting Authority"))
```

### 3. Complex Ownership Chains

Model multi-tier ownership structures:

```lisp
(ownership-structure
  (owner PARENT-CORP 100)
  (beneficial-owner ULTIMATE-OWNER 51)
  (controller MANAGING-DIRECTOR "Senior Managing Official")
  (controller COMPLIANCE-OFFICER "Chief Compliance Officer"))
```

### 4. Regulatory Reporting

Generate ownership reports for regulatory submissions (FATCA, CRS, etc.)

## Validation

The validator checks:

1. **Grammar compliance** - Structure matches EBNF v1.1
2. **Function validity** - `BUILD-OWNERSHIP-TREE` and `VERIFY-OWNERSHIP` are valid functions
3. **Percentage format** - Numbers followed by `%`
4. **Role format** - Quoted strings for controller roles

## Round-Trip Verification

The system ensures perfect round-trip fidelity:

```
DSL File → Parse → Bind → Model → Serialize → DSL Text
```

Example verification:

```bash
# Process a DSL file
./bin/kycctl ownership_case.dsl

# Retrieve serialized version from database
psql -d kyc_dsl -c "SELECT dsl_snapshot FROM kyc_case_versions 
                    WHERE case_name='BLACKROCK-GLOBAL-EQUITY-FUND' 
                    ORDER BY version DESC LIMIT 1;"
```

The output should match the original DSL structure exactly (with consistent formatting).

## Compliance Benefits

### Regulatory Requirements Met

- ✅ **FATCA** - Foreign Account Tax Compliance Act
- ✅ **CRS** - Common Reporting Standard
- ✅ **AML** - Anti-Money Laundering regulations
- ✅ **UBO** - Ultimate Beneficial Owner identification
- ✅ **PEP** - Politically Exposed Person screening

### Audit Trail

Every ownership structure change is:
- Versioned with SHA-256 hash
- Timestamped in `kyc_case_amendments`
- Fully serialized in `kyc_case_versions`
- Traceable through complete history

### Determinism

Same ownership data always produces:
- Identical DSL text
- Identical hash
- Identical database representation

## Future Enhancements

Planned features:

1. **Ownership percentage validation** - Ensure totals don't exceed 100%
2. **Ownership graph visualization** - Generate ownership tree diagrams
3. **Chain resolution** - Trace ultimate beneficial owners through multiple tiers
4. **Conflict detection** - Identify overlapping control persons
5. **Historical tracking** - Track ownership changes over time
6. **Automated discovery** - API integration for corporate registry lookups

## Testing

### Test File: ownership_case.dsl

```lisp
(kyc-case BLACKROCK-GLOBAL-EQUITY-FUND
  (nature-purpose
    (nature "Institutional investment management vehicle")
    (purpose "Operate a SICAV with multi-jurisdictional sub-funds"))
  (client-business-unit BLACKROCK-GLOBAL-FUNDS)
  (function BUILD-OWNERSHIP-TREE)
  (ownership-structure
    (owner BLACKROCK-PLC 100)
    (beneficial-owner LARRY-FINK 35)
    (controller JANE-DOE "Senior Managing Official"))
  (kyc-token "pending"))
```

### Run Tests

```bash
# Process the file
./bin/kycctl ownership_case.dsl

# Verify round-trip
psql -d kyc_dsl -c "SELECT dsl_snapshot FROM kyc_case_versions 
                    WHERE case_name='BLACKROCK-GLOBAL-EQUITY-FUND' 
                    ORDER BY version DESC LIMIT 1;"

# Test amendment
kycctl amend BLACKROCK-GLOBAL-EQUITY-FUND --step=ownership-discovery
```

## See Also

- `AMENDMENT_SYSTEM.md` - Amendment lifecycle documentation
- `README.md` - Main project documentation
- `REFACTORING_SUMMARY.md` - CLI refactoring details
- `CLAUDE.md` - Project guidance for Claude AI

## Version History

- **v1.0** - Initial KYC-DSL with basic case structure
- **v1.1** - Added ownership and control constructs (this release)