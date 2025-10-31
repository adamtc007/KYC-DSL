# NO SIDE DOORS Policy

**Status**: ENFORCED  
**Version**: 1.5  
**Effective Date**: 2024

---

## 🚫 Core Principle

**ALL database access MUST go through the Go gRPC Data Service API.**

**NO direct SQL connections are permitted from:**
- CLI tools (`kycctl`)
- External clients
- Application code outside the data service
- Test scripts (except for data service tests)

---

## Architecture Enforcement

### ✅ CORRECT: Through Data Service

```
┌─────────────┐
│   kycctl    │  (CLI)
│   (Go)      │
└─────┬───────┘
      │
      │ gRPC only
      │
      ├──────────────────┬─────────────────┐
      │                  │                 │
      ▼                  ▼                 ▼
┌──────────────┐   ┌──────────────┐  ┌──────────────┐
│ Rust Service │   │ Data Service │  │  Other APIs  │
│ (port 50060) │   │ (port 50070) │  │              │
└──────────────┘   └──────┬───────┘  └──────────────┘
                          │
                          │ SQL (ONLY HERE)
                          ▼
                    ┌──────────────┐
                    │  PostgreSQL  │
                    │  (port 5432) │
                    └──────────────┘
```

### ❌ FORBIDDEN: Side Doors

```
┌─────────────┐
│   kycctl    │  ❌ NO DIRECT SQL!
└─────┬───────┘
      │
      │ ❌ storage.ConnectPostgres()
      │ ❌ sqlx.Open()
      │ ❌ db.Query()
      ▼
┌──────────────┐
│  PostgreSQL  │
└──────────────┘
```

---

## What This Means

### ❌ FORBIDDEN Patterns

```go
// ❌ BAD: Direct database connection in CLI
func RunGetCaseCommand(caseName string) error {
    db, err := storage.ConnectPostgres()  // ❌ NO!
    defer db.Close()
    
    var dsl string
    db.Get(&dsl, "SELECT dsl_snapshot FROM cases...")  // ❌ NO!
    return nil
}
```

```go
// ❌ BAD: Direct SQL import in CLI
import (
    "github.com/jmoiron/sqlx"  // ❌ NO!
    _ "github.com/lib/pq"      // ❌ NO!
)
```

### ✅ CORRECT Patterns

```go
// ✅ GOOD: Use Data Service client
func RunGetCaseCommand(caseName string) error {
    client, err := dataclient.NewDataClient("localhost:50070")  // ✅ YES!
    defer client.Close()
    
    caseVersion, err := client.GetLatestCaseVersion(caseName)  // ✅ YES!
    fmt.Println(caseVersion.DslSource)
    return nil
}
```

```go
// ✅ GOOD: Use gRPC client in CLI
import (
    "github.com/adamtc007/KYC-DSL/internal/dataclient"  // ✅ YES!
    pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"   // ✅ YES!
)
```

---

## Service Boundaries

### Data Service (Port 50070) - THE ONLY DATABASE ACCESS POINT

**Responsibility**: All PostgreSQL operations

**Allowed to:**
- ✅ Import `github.com/jmoiron/sqlx`
- ✅ Import `github.com/lib/pq`
- ✅ Use `storage.ConnectPostgres()`
- ✅ Execute SQL queries
- ✅ Manage connection pools
- ✅ Handle transactions

**Location**: `cmd/dataserver/`, `internal/dataservice/`, `internal/storage/`

### CLI (kycctl)

**Responsibility**: User interface and command routing

**Allowed to:**
- ✅ Import `internal/dataclient`
- ✅ Import `internal/rustclient`
- ✅ Call gRPC services
- ❌ **NEVER** import `sqlx`, `lib/pq`, or `storage`
- ❌ **NEVER** connect directly to PostgreSQL

**Location**: `cmd/kycctl/`, `internal/cli/`

### Rust DSL Service (Port 50060)

**Responsibility**: DSL parsing, validation, compilation

**Allowed to:**
- ✅ Parse DSL syntax
- ✅ Validate DSL semantics
- ✅ Compile DSL to instructions
- ❌ **NEVER** access PostgreSQL

**Location**: `rust/kyc_dsl_service/`

---

## Migration Checklist

### Phase 1: Audit Current Violations ✅

- [x] Identify all `storage.ConnectPostgres()` calls in CLI
- [x] Identify all direct SQL in non-data-service code
- [x] Document current violations

### Phase 2: Create Data Service Client ✅

- [x] Create `internal/dataclient/` package
- [x] Implement gRPC client wrapper
- [x] Add convenience methods for common operations

### Phase 3: Refactor CLI Commands

- [ ] Replace `storage.ConnectPostgres()` with `dataclient.NewDataClient()`
- [ ] Replace `storage.GetLatestDSL()` with `client.GetLatestCaseVersion()`
- [ ] Replace `storage.InsertCase()` with `client.SaveCaseVersion()`
- [ ] Replace `ontology.NewRepository()` with ontology gRPC calls

### Phase 4: Extend Data Service API

Add missing RPCs to Data Service:

- [ ] `ListAllCases` - list all cases with metadata
- [ ] `ListCaseVersions` - list all versions of a case
- [ ] `GetCaseByName` - get case metadata
- [ ] `DeleteCase` - delete a case (if needed)
- [ ] `SearchCases` - search cases by criteria

### Phase 5: Remove Direct SQL Access

- [ ] Remove `internal/storage` imports from CLI
- [ ] Remove `sqlx` imports from CLI
- [ ] Add linter rules to prevent future violations

---

## Required Data Service RPCs

### Current RPCs (Port 50070)

```protobuf
service DictionaryService {
  rpc GetAttribute(GetAttributeRequest) returns (Attribute);
  rpc ListAttributes(ListAttributesRequest) returns (AttributeList);
}

service CaseService {
  rpc SaveCaseVersion(CaseVersionRequest) returns (CaseVersionResponse);
  rpc GetCaseVersion(GetCaseRequest) returns (CaseVersion);
}

service OntologyService {
  // ... ontology operations
}
```

### Missing RPCs (Need to Add)

```protobuf
service CaseService {
  // Existing
  rpc SaveCaseVersion(CaseVersionRequest) returns (CaseVersionResponse);
  rpc GetCaseVersion(GetCaseRequest) returns (CaseVersion);
  
  // TO ADD:
  rpc ListAllCases(ListCasesRequest) returns (CaseList);
  rpc ListCaseVersions(ListVersionsRequest) returns (VersionList);
  rpc GetCaseMetadata(GetCaseMetadataRequest) returns (CaseMetadata);
  rpc DeleteCaseVersion(DeleteVersionRequest) returns (DeleteResponse);
  rpc SearchCases(SearchCasesRequest) returns (CaseList);
}
```

---

## Enforcement Mechanisms

### 1. Code Review Checklist

- [ ] No `storage.ConnectPostgres()` outside of `internal/storage/` or `cmd/dataserver/`
- [ ] No `sqlx` imports outside of data service
- [ ] No `lib/pq` imports outside of data service
- [ ] CLI uses `dataclient` only
- [ ] All database operations go through gRPC

### 2. Automated Linting

```bash
# Add to CI/CD pipeline
make lint-side-doors
```

```makefile
# Makefile target
lint-side-doors:
	@echo "🔍 Checking for side door violations..."
	@! grep -r "storage.ConnectPostgres" internal/cli/ cmd/kycctl/ || \
		(echo "❌ Found direct database access in CLI!" && exit 1)
	@! grep -r '"github.com/jmoiron/sqlx"' internal/cli/ cmd/kycctl/ || \
		(echo "❌ Found sqlx import in CLI!" && exit 1)
	@! grep -r '"github.com/lib/pq"' internal/cli/ cmd/kycctl/ || \
		(echo "❌ Found pq import in CLI!" && exit 1)
	@echo "✅ No side doors detected"
```

### 3. Import Restrictions

Add to `.golangci.yml`:

```yaml
linters-settings:
  depguard:
    rules:
      cli-no-db:
        files:
          - "**/internal/cli/**"
          - "**/cmd/kycctl/**"
        deny:
          - pkg: "github.com/jmoiron/sqlx"
            desc: "CLI must not access database directly. Use dataclient instead."
          - pkg: "github.com/lib/pq"
            desc: "CLI must not access database directly. Use dataclient instead."
          - pkg: "github.com/adamtc007/KYC-DSL/internal/storage"
            desc: "CLI must not import storage package. Use dataclient instead."
```

---

## Benefits of This Architecture

### 1. **Single Source of Truth**
- All database logic in one place
- Consistent error handling
- Centralized connection pooling

### 2. **Security**
- No connection strings in CLI
- Centralized authentication/authorization
- Audit all database access at service boundary

### 3. **Scalability**
- Data service can be scaled independently
- Connection pool managed centrally
- Easy to add caching layer

### 4. **Testing**
- Mock data service in tests
- No need for test databases in CLI tests
- Integration tests focus on service boundary

### 5. **Polyglot Support**
- Python clients can use same gRPC API
- JavaScript/TypeScript clients via grpc-web
- No language-specific database drivers needed

---

## Example: Correct CLI Implementation

### Before (❌ Side Door)

```go
func RunGetCaseCommand(caseName string) error {
    // ❌ Direct database access
    db, err := storage.ConnectPostgres()
    if err != nil {
        return err
    }
    defer db.Close()
    
    var dsl string
    err = db.Get(&dsl, "SELECT dsl_snapshot FROM kyc_case_versions WHERE case_name=$1 ORDER BY version DESC LIMIT 1", caseName)
    if err != nil {
        return err
    }
    
    fmt.Println(dsl)
    return nil
}
```

### After (✅ Through Data Service)

```go
func RunGetCaseCommand(caseName string) error {
    // ✅ Use Data Service client
    client, err := dataclient.NewDataClient("")
    if err != nil {
        return fmt.Errorf("failed to connect to data service: %w", err)
    }
    defer client.Close()
    
    caseVersion, err := client.GetLatestCaseVersion(caseName)
    if err != nil {
        return fmt.Errorf("failed to get case: %w", err)
    }
    
    fmt.Printf("📦 Case: %s\n", caseVersion.CaseName)
    fmt.Printf("📌 Version: %d\n", caseVersion.Version)
    fmt.Printf("🔑 Hash: %s\n", caseVersion.Hash)
    fmt.Println("─────────────────────────────────────────────")
    fmt.Println(caseVersion.DslSource)
    
    return nil
}
```

---

## Exception Policy

### When Direct Database Access IS Allowed

1. **Inside `internal/storage/`** - Database abstraction layer
2. **Inside `internal/dataservice/`** - Data service implementation
3. **Inside `cmd/dataserver/`** - Data service server
4. **Data service unit tests** - Testing storage layer directly
5. **Database migration scripts** - Schema management

### When Direct Database Access is FORBIDDEN

1. ❌ CLI commands (`internal/cli/`, `cmd/kycctl/`)
2. ❌ REST API handlers (`cmd/kycserver/`)
3. ❌ Rust code (should never touch PostgreSQL)
4. ❌ Client libraries
5. ❌ Example code
6. ❌ Integration tests (use gRPC services)

---

## Migration Status

| Component | Status | Notes |
|-----------|--------|-------|
| CLI Grammar Command | ✅ COMPLIANT | Uses Rust gRPC |
| CLI Process Command | ⚠️ MIXED | Uses Rust + direct SQL |
| CLI Amend Command | ⚠️ MIXED | Uses Rust + direct SQL |
| CLI Ontology Command | ❌ VIOLATION | Direct SQL |
| CLI RAG Commands | ❌ VIOLATION | Direct SQL |
| CLI Get Command | 🚧 IN PROGRESS | New command, needs data service |
| REST API | ⚠️ MIXED | Some direct SQL |
| Data Service | ✅ COMPLIANT | Owns all SQL |
| Rust Service | ✅ COMPLIANT | No database access |

---

## Action Items

### Immediate (P0)
1. ⚠️ **DO NOT** add new direct SQL in CLI
2. ⚠️ **DO NOT** merge PRs with `storage.ConnectPostgres()` in CLI
3. ✅ Use `dataclient` for all new CLI commands

### Short-term (P1)
1. Add missing RPCs to Data Service
2. Migrate CLI commands to use `dataclient`
3. Add linter rules to prevent violations

### Long-term (P2)
1. Remove `internal/storage` imports from all CLI code
2. Add authentication to Data Service
3. Add authorization/RBAC to Data Service
4. Add caching layer in Data Service

---

## Questions?

**Q: Why is this important?**  
A: Single source of truth, security, scalability, and testability.

**Q: What if the Data Service is down?**  
A: That's by design. If the data layer is down, nothing should work. This makes failures explicit and easier to debug.

**Q: Isn't gRPC overhead too much?**  
A: No. gRPC is highly optimized and adds negligible latency compared to direct SQL over the network. The architectural benefits far outweigh any minimal overhead.

**Q: Can I add one quick SQL query in the CLI?**  
A: NO. Add the RPC to the Data Service instead. This keeps the architecture clean.

---

**Last Updated**: 2024  
**Enforced By**: Code review + linting  
**Exceptions**: Documented above  
**Violations**: Report to architecture team