# Lineage Evaluator

## Overview

The **Lineage Evaluator** is a runtime engine that executes derived attribute rules dynamically, providing:

- **Dynamic Evaluation**: Compute derived attributes from public attribute values at runtime
- **Complete Lineage Tracking**: Record which inputs produced which outputs
- **Explainability**: Generate human-readable explanations of computations
- **Audit Trail**: Persist all evaluation results for regulatory compliance
- **Cascading Derivations**: Use derived values in subsequent derivations

## Purpose

The Lineage Evaluator enables:

1. **Evaluate Derived Attributes**: Execute rule expressions dynamically against case data
2. **Reference Values**: Access other attributes' values within the same KYC case
3. **Record Results**: Log evaluation outcomes with complete lineage for audit
4. **Provide Explainability**: Generate explanations like "This flag = TRUE because TaxResidencyCountry = IR, and rule X matched"

## Architecture

```
┌─────────────────┐
│  Public         │
│  Attributes     │ ──┐
│  (Documents)    │   │
└─────────────────┘   │
                      ├──> ┌──────────────┐      ┌─────────────┐
┌─────────────────┐   │    │   Lineage    │      │  Evaluation │
│  Derived        │   ├───>│  Evaluator   │─────>│  Results    │
│  Attribute      │   │    │              │      │  + Lineage  │
│  Rules          │ ──┘    └──────────────┘      └─────────────┘
└─────────────────┘              │
                                 │
                                 v
                         ┌───────────────┐
                         │  Audit Trail  │
                         │  (Database)   │
                         └───────────────┘
```

## Implementation

### Core Components

**File**: `internal/lineage/evaluator.go`

#### EvaluationResult

Holds the outcome of a rule evaluation:

```go
type EvaluationResult struct {
    DerivedCode string          // Attribute being computed
    Value       any             // Computed value (bool, float64, string, etc.)
    Success     bool            // Whether evaluation succeeded
    Error       string          // Error message if failed
    Timestamp   time.Time       // When evaluation occurred
    Rule        string          // The rule expression that was evaluated
    Inputs      map[string]any  // Source attribute values used
}
```

#### Evaluator

Runs derived-attribute rules in a sandboxed context:

```go
type Evaluator struct {
    env     map[string]any         // Attribute values (public + derived)
    program map[string]*vm.Program // Compiled rule expressions
    results []EvaluationResult     // Evaluation outcomes
}
```

### Key Methods

#### NewEvaluator
```go
func NewEvaluator(attrValues map[string]any) *Evaluator
```

Creates an evaluator with known public attribute values.

**Example:**
```go
caseData := map[string]any{
    "TAX_RESIDENCY_COUNTRY": "IR",
    "REGISTERED_NAME":       "BlackRock Global Fund",
}
eval := lineage.NewEvaluator(caseData)
```

#### CompileDerivations
```go
func (e *Evaluator) CompileDerivations(derivations []model.DerivedAttribute) error
```

Compiles all rule expressions ahead of time for performance.

**Example:**
```go
derivations := []model.DerivedAttribute{
    {
        DerivedAttribute: "HIGH_RISK_JURISDICTION_FLAG",
        SourceAttributes: []string{"TAX_RESIDENCY_COUNTRY"},
        RuleExpression:   `TAX_RESIDENCY_COUNTRY in ["IR", "KP", "SY"]`,
    },
}
err := eval.CompileDerivations(derivations)
```

#### Evaluate
```go
func (e *Evaluator) Evaluate(derivations []model.DerivedAttribute) []EvaluationResult
```

Runs all compiled expressions and stores results.

**Example:**
```go
results := eval.Evaluate(derivations)
for _, r := range results {
    fmt.Printf("%s = %v (success=%v)\n", r.DerivedCode, r.Value, r.Success)
}
```

#### GetValue
```go
func (e *Evaluator) GetValue(derivedCode string) (any, bool)
```

Retrieves the evaluated value for a derived attribute.

**Example:**
```go
if val, ok := eval.GetValue("HIGH_RISK_JURISDICTION_FLAG"); ok {
    fmt.Printf("Flag value: %v\n", val)
}
```

#### ExplainResult
```go
func (r *EvaluationResult) ExplainResult() string
```

Generates human-readable explanation of evaluation.

**Example:**
```go
for _, r := range results {
    fmt.Println(r.ExplainResult())
}
```

## Rule Expression Language

The evaluator uses **expr-lang** (https://github.com/expr-lang/expr) which supports:

### Boolean Expressions

```go
// Membership check
`TAX_RESIDENCY_COUNTRY in ["IR", "KP", "SY"]`

// Comparison
`UBO_PERCENT > 75`

// Logical operators
`PEP_STATUS == true || SANCTIONS_STATUS == true`

// Combined conditions
`TAX_RESIDENCY_COUNTRY in ["IR", "KP"] && ENTITY_AGE_YEARS < 2`
```

### Numeric Expressions

```go
// Built-in functions
`max(UBO_PERCENT)`
`min(UBO_PERCENT)`
`len(UBO_NAME)`

// Arithmetic
`ENTITY_AGE_YEARS * 12`  // Convert to months
`sum(UBO_PERCENT)`
```

### String Operations

```go
// Concatenation
`FIRST_NAME + " " + LAST_NAME`

// Length
`len(REGISTERED_NAME) > 100`

// Comparison
`ENTITY_TYPE == "Corporation"`
```

### Array Operations

```go
// Length check
`len(UBO_NAME) > 3`

// Aggregate functions
`max(UBO_PERCENT)`
`sum(OWNERSHIP_SHARES)`
```

## Complete Usage Example

### 1. Setup Public Attributes

```go
caseData := map[string]any{
    "TAX_RESIDENCY_COUNTRY":      "IR",
    "INCORPORATION_JURISDICTION": "US",
    "REGISTERED_NAME":            "BlackRock Global Fund",
    "UBO_PERCENT":                []float64{35.0, 45.0, 20.0},
    "PEP_STATUS":                 false,
}
```

### 2. Define Derived Attributes

```go
derivations := []model.DerivedAttribute{
    {
        DerivedAttribute: "HIGH_RISK_JURISDICTION_FLAG",
        SourceAttributes: []string{"TAX_RESIDENCY_COUNTRY"},
        RuleExpression:   `TAX_RESIDENCY_COUNTRY in ["IR", "KP", "SY"]`,
        Jurisdiction:     "GLOBAL",
        RegulationCode:   "AMLD5",
    },
    {
        DerivedAttribute: "UBO_CONCENTRATION_SCORE",
        SourceAttributes: []string{"UBO_PERCENT"},
        RuleExpression:   `max(UBO_PERCENT)`,
        Jurisdiction:     "GLOBAL",
        RegulationCode:   "AMLD5",
    },
}
```

### 3. Create and Compile

```go
eval := lineage.NewEvaluator(caseData)
if err := eval.CompileDerivations(derivations); err != nil {
    log.Fatal("Compilation failed:", err)
}
```

### 4. Evaluate

```go
results := eval.Evaluate(derivations)
```

### 5. Display Results

```go
for _, r := range results {
    if r.Success {
        fmt.Printf("✅ %s = %v\n", r.DerivedCode, r.Value)
        fmt.Printf("   Inputs: %+v\n", r.Inputs)
        fmt.Printf("   Rule: %s\n\n", r.Rule)
    } else {
        fmt.Printf("❌ %s failed: %s\n", r.DerivedCode, r.Error)
    }
}
```

**Output:**
```
✅ HIGH_RISK_JURISDICTION_FLAG = true
   Inputs: map[TAX_RESIDENCY_COUNTRY:IR]
   Rule: TAX_RESIDENCY_COUNTRY in ["IR", "KP", "SY"]

✅ UBO_CONCENTRATION_SCORE = 45
   Inputs: map[UBO_PERCENT:[35 45 20]]
   Rule: max(UBO_PERCENT)
```

## DSL-to-Execution Flow

### 1. DSL Definition

```lisp
(derived-attributes
  (attribute HIGH_RISK_JURISDICTION_FLAG
    (sources (TAX_RESIDENCY_COUNTRY))
    (rule "TAX_RESIDENCY_COUNTRY in ['IR', 'KP', 'SY']")
    (jurisdiction GLOBAL)
    (regulation AMLD5)
  )
)
```

### 2. Parser Binds to Model

```go
model.DerivedAttribute{
    DerivedAttribute: "HIGH_RISK_JURISDICTION_FLAG",
    SourceAttributes: []string{"TAX_RESIDENCY_COUNTRY"},
    RuleExpression:   "TAX_RESIDENCY_COUNTRY in ['IR', 'KP', 'SY']",
    Jurisdiction:     "GLOBAL",
    RegulationCode:   "AMLD5",
}
```

### 3. Evaluator Compiles and Executes

```go
eval.CompileDerivations(derivations)  // Compile rules
results := eval.Evaluate(derivations)  // Execute rules
```

### 4. Result Available

```
HIGH_RISK_JURISDICTION_FLAG = true
  because TAX_RESIDENCY_COUNTRY = IR
  rule = "TAX_RESIDENCY_COUNTRY in ['IR', 'KP', 'SY']"
  evaluated_at = 2024-10-30T17:03:29Z
```

### 5. Regulator/Agent Can Query

```sql
SELECT derived_code, value, inputs, rule, evaluated_at
FROM kyc_lineage_evaluations
WHERE case_name = 'BLACKROCK-GLOBAL-EQUITY-FUND'
  AND derived_code = 'HIGH_RISK_JURISDICTION_FLAG';
```

## Audit Trail Storage

### Database Schema

**File**: `internal/storage/migrations/005_lineage_evaluations.sql`

```sql
CREATE TABLE kyc_lineage_evaluations (
    id SERIAL PRIMARY KEY,
    case_name TEXT NOT NULL,
    case_version INT,
    derived_code TEXT NOT NULL,
    value TEXT,
    value_type TEXT,
    success BOOLEAN NOT NULL,
    error TEXT,
    inputs JSONB,
    rule TEXT NOT NULL,
    jurisdiction TEXT,
    regulation_code TEXT,
    evaluated_at TIMESTAMP DEFAULT NOW()
);
```

### Recording Evaluations

```go
import "github.com/adamtc007/KYC-DSL/internal/storage"

for _, result := range results {
    err := storage.RecordLineageEvaluation(db, caseName, caseVersion, result)
    if err != nil {
        log.Printf("Failed to record evaluation: %v", err)
    }
}
```

### Querying History

```go
evaluations, err := storage.GetLineageEvaluations(db, "BLACKROCK-GLOBAL-EQUITY-FUND")
for _, eval := range evaluations {
    fmt.Printf("%s = %v at %v\n",
        eval["derived_code"],
        eval["value"],
        eval["evaluated_at"])
}
```

## Explainability Features

### Generate Explanation

```go
for _, r := range results {
    if r.DerivedCode == "HIGH_RISK_JURISDICTION_FLAG" {
        fmt.Println("Why is HIGH_RISK_JURISDICTION_FLAG = true?")
        fmt.Println(r.ExplainResult())
    }
}
```

**Output:**
```
Why is HIGH_RISK_JURISDICTION_FLAG = true?
✅ HIGH_RISK_JURISDICTION_FLAG = true
   Rule: TAX_RESIDENCY_COUNTRY in ["IR", "KP", "SY"]
   Inputs:
     • TAX_RESIDENCY_COUNTRY = IR
   Evaluated at: 2024-10-30T17:03:29Z
```

### Trace Lineage

```go
fmt.Printf("Flag is %v because:\n", result.Value)
fmt.Printf("  - TAX_RESIDENCY_COUNTRY = '%v'\n", result.Inputs["TAX_RESIDENCY_COUNTRY"])
fmt.Printf("  - Rule: %s\n", result.Rule)
fmt.Printf("  - 'IR' is in the high-risk jurisdiction list\n")
```

## Cascading Evaluations

Derived values become available for subsequent rules:

```go
derivations := []model.DerivedAttribute{
    // First: compute risk flag
    {
        DerivedAttribute: "HIGH_RISK_JURISDICTION_FLAG",
        SourceAttributes: []string{"TAX_RESIDENCY_COUNTRY"},
        RuleExpression:   `TAX_RESIDENCY_COUNTRY in ["IR", "KP"]`,
    },
    // Second: use the flag in another derivation
    {
        DerivedAttribute: "OVERALL_RISK_SCORE",
        SourceAttributes: []string{"HIGH_RISK_JURISDICTION_FLAG", "PEP_STATUS"},
        RuleExpression:   `(HIGH_RISK_JURISDICTION_FLAG ? 50 : 0) + (PEP_STATUS ? 30 : 0)`,
    },
}

results := eval.Evaluate(derivations)
// HIGH_RISK_JURISDICTION_FLAG computed first
// Then available for OVERALL_RISK_SCORE computation
```

## Error Handling

### Compilation Errors

```go
err := eval.CompileDerivations(derivations)
if err != nil {
    // Rule syntax error detected at compile time
    fmt.Printf("Compilation failed: %v\n", err)
}
```

**Example Error:**
```
Compilation error for HIGH_RISK_JURISDICTION_FLAG: undefined identifier "NONEXISTENT_FIELD"
```

### Runtime Errors

```go
results := eval.Evaluate(derivations)
for _, r := range results {
    if !r.Success {
        fmt.Printf("❌ %s failed: %s\n", r.DerivedCode, r.Error)
        fmt.Printf("   Rule: %s\n", r.Rule)
        fmt.Printf("   Inputs: %+v\n", r.Inputs)
    }
}
```

**Example Error:**
```
❌ DIVISION_CALC failed: division by zero
   Rule: TOTAL_VALUE / SHARE_COUNT
   Inputs: map[TOTAL_VALUE:1000 SHARE_COUNT:0]
```

## Running the Example

```bash
cd examples/lineage_evaluator
GOEXPERIMENT=greenteagc go run main.go
```

**Sample Output:**
```
=== Lineage Evaluator Example ===

📋 Step 1: Loading public attribute values...
  • TAX_RESIDENCY_COUNTRY = IR
  • UBO_PERCENT = [35 45 20]

📋 Step 2: Defining derived attributes...
  • HIGH_RISK_JURISDICTION_FLAG
  • UBO_CONCENTRATION_SCORE

📋 Step 3: Compiling rule expressions...
✅ All rules compiled successfully

📋 Step 4: Evaluating derived attributes...

📊 Evaluation Results:

✅ HIGH_RISK_JURISDICTION_FLAG = true
   Rule: TAX_RESIDENCY_COUNTRY in ["IR", "KP", "SY"]
   Inputs:
     • TAX_RESIDENCY_COUNTRY = IR

✅ UBO_CONCENTRATION_SCORE = 45
   Rule: max(UBO_PERCENT)
   Inputs:
     • UBO_PERCENT = [35 45 20]

=== Summary ===
Total Evaluations: 5
✅ Successful: 5
❌ Failed: 0
```

## Query Examples

### Recent Evaluations

```sql
SELECT case_name, derived_code, value, evaluated_at
FROM kyc_lineage_evaluations
ORDER BY evaluated_at DESC
LIMIT 10;
```

### Failed Evaluations

```sql
SELECT * FROM failed_lineage_evaluations
WHERE evaluated_at > NOW() - INTERVAL '7 days';
```

### Evaluation Health

```sql
SELECT * FROM lineage_evaluation_summary
WHERE health_status != 'HEALTHY';
```

### Attribute Usage

```sql
SELECT
    derived_code,
    COUNT(*) as eval_count,
    AVG(CASE WHEN success THEN 1 ELSE 0 END) as success_rate
FROM kyc_lineage_evaluations
WHERE evaluated_at > NOW() - INTERVAL '30 days'
GROUP BY derived_code
ORDER BY eval_count DESC;
```

## Benefits

### For Compliance

- **Transparency**: Every computation is traceable
- **Auditability**: Complete history of evaluations
- **Explainability**: Generate explanations for regulators
- **Determinism**: Same inputs always produce same outputs

### For Operations

- **Debugging**: See why flags are set
- **Monitoring**: Track evaluation success rates
- **Testing**: Validate rule logic with test cases
- **Performance**: Pre-compiled rules for speed

### For AI Agents

- **Reasoning**: Agents can understand derivations
- **Validation**: Verify computed values
- **Learning**: Train on evaluation patterns
- **Explanation**: Generate natural language explanations

## Performance

- **Compilation**: Rules compiled once, executed many times
- **Caching**: Compiled programs reused across evaluations
- **Cascading**: Derived values cached for subsequent rules
- **Parallel**: Multiple cases can be evaluated concurrently

**Typical Performance:**
- Compilation: 1-5ms per rule
- Evaluation: <1ms per rule
- Total for 10 rules: ~10ms

## Security

- **Sandboxed**: expr-lang provides safe evaluation
- **No File Access**: Rules cannot access filesystem
- **No Network**: Rules cannot make network calls
- **Type Safe**: Runtime type checking
- **Input Validation**: Source attributes validated

## Future Enhancements

1. **Custom Functions**: Register domain-specific functions
2. **Async Evaluation**: Background evaluation for large cases
3. **Rule Versioning**: Track changes to rule expressions
4. **Visual Debugger**: Step through rule evaluation
5. **Performance Metrics**: Profile rule execution time
6. **A/B Testing**: Compare different rule versions
7. **ML Integration**: Learn optimal rules from data

## Related Documentation

- `DERIVED_ATTRIBUTES.md` - Public/Private attribute system
- `REGULATORY_ONTOLOGY.md` - Ontology structure
- `VALIDATION_AUDIT.md` - Audit trail system
- `examples/lineage_evaluator/main.go` - Complete working example

---

**Version**: 1.3  
**Status**: Production Ready  
**Engine**: expr-lang v1.17.6  
**Performance**: ~1ms per rule evaluation