# Go-to-Rust Migration Plan: DSL Parser & Engine Cleanup

**Version:** 1.0  
**Date:** 2024  
**Status:** Ready for Execution

## Executive Summary

This document outlines the complete migration from legacy Go DSL parsing/execution code to the Rust implementation, eliminating code duplication and modernizing the architecture.

**Goal:** Remove all legacy Go parser/engine code and route all DSL operations through the Rust gRPC service (port 50060).

---

## Current Architecture Problems

### 1. Code Duplication
- **Go Parser**: `internal/parser/` (S-expression parser with participle)
- **Rust Parser**: `rust/kyc_dsl_core/src/parser.rs` (nom-based parser)
- **Result**: Two separate parsing implementations that must be kept in sync

### 2. Maintenance Burden
- Grammar changes require updates in both Go and Rust
- Validation logic duplicated across codebases
- Testing overhead for both implementations

### 3. Architecture Confusion
- Multiple gRPC services serving similar purposes
- Unclear separation of concerns
- Port proliferation (50051, 50060, 50070)

---

## Target Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Client Layer                         │
│  (kycctl CLI, REST clients, external services)         │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│              Rust DSL Service (Port 50060)              │
│  - Parse DSL                                            │
│  - Validate DSL                                         │
│  - Execute Functions                                    │
│  - Serialize Cases                                      │
│  - Apply Amendments                                     │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│          Go Data Service (Port 50070)                   │
│  - Database Access (PostgreSQL)                         │
│  - Dictionary Service                                   │
│  - Case Version Control                                 │
│  - Ontology Repository                                  │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │  PostgreSQL  │
                    └──────────────┘
```

**Key Principle:** Rust owns computation, Go owns data.

---

## Files Affected

### Files to MODIFY (Update to use Rust gRPC)

1. **`internal/cli/cli.go`** (Heavy modification)
   - Current: Calls `parser.ParseFile()`, `engine.RunCase()`
   - Target: Call Rust gRPC service for parsing/execution
   - Keep: Database operations, RAG commands

2. **`internal/amend/amend.go`** (Moderate modification)
   - Current: Calls `parser.Parse()`, `parser.Bind()`, `parser.SerializeCases()`
   - Target: Call Rust gRPC `Amend()` RPC
   - Keep: Database version tracking

3. **`cmd/kycctl/main.go`** (No changes needed)
   - Already just delegates to `internal/cli`

### Files to DELETE (Legacy Code)

1. **`internal/parser/`** (Entire directory)
   - `parser.go` - S-expression parser
   - `bind.go` - AST to model binding
   - `validate.go` - DSL validation
   - `serialize.go` - Model to DSL serialization
   - `grammar.go` - EBNF grammar definition

2. **`internal/engine/`** (Entire directory)
   - `engine.go` - Case execution engine
   - `executor.go` - Function executor

3. **`internal/service/dsl_service.go`** (Delete or deprecate)
   - Duplicates Rust DSL service functionality
   - Old gRPC service on port 50051

4. **`internal/service/kyc_case_service.go`** (Delete or deprecate)
   - Case management now handled by Rust + DataService

5. **`cmd/server/`** (Delete entire directory)
   - Old gRPC server (port 50051)
   - Replaced by Rust service (port 50060)

6. **`cmd/kycserver/`** (Optional - keep if REST needed)
   - REST API wrapper
   - Could be rewritten to proxy to Rust gRPC

### Files to KEEP (Data Access Layer)

✅ **`internal/storage/`** - PostgreSQL operations  
✅ **`internal/dataservice/`** - Data service implementation  
✅ **`internal/ontology/`** - Ontology repository  
✅ **`internal/rag/`** - RAG/vector search  
✅ **`internal/model/`** - Data models  
✅ **`internal/token/`** - Token management  
✅ **`cmd/dataserver/`** - Data service gRPC server  

---

## Migration Steps

### Phase 1: Setup & Validation (Day 1)

**1.1 Verify Rust Service Functionality**
```bash
# Start Rust DSL service
cd rust
cargo build --release
cargo run -p kyc_dsl_service

# In another terminal, test with grpcurl
grpcurl -plaintext \
  -d '{"dsl": "(kyc-case TEST)"}' \
  localhost:50060 \
  kyc.dsl.DslService/Parse
```

**1.2 Create gRPC Client Helper in Go**
```bash
# Create new file: internal/rustclient/dsl_client.go
```

This client will wrap Rust gRPC calls for use in Go CLI.

### Phase 2: Update CLI to Use Rust (Day 2-3)

**2.1 Create Rust gRPC Client Wrapper**

File: `internal/rustclient/dsl_client.go`
- `ParseDSL(dsl string) (*pb.ParseResponse, error)`
- `ValidateDSL(dsl string) (*pb.ValidationResult, error)`
- `ExecuteCase(caseID, function string) (*pb.ExecuteResponse, error)`
- `AmendCase(caseName, amendType string) (*pb.AmendResponse, error)`
- `SerializeCase(case *pb.ParsedCase) (string, error)`

**2.2 Update `internal/cli/cli.go`**

Replace:
```go
// OLD
dsl, err := parser.ParseFile(filePath)
cases, err := parser.Bind(dsl)
err = parser.ValidateDSL(db, cases, ebnf)
exec := engine.NewExecutor(db)
err = exec.RunCase(cases[0].Name, serialized)
```

With:
```go
// NEW
rustClient := rustclient.NewDslClient("localhost:50060")
resp, err := rustClient.ParseDSL(dslContent)
valResult, err := rustClient.ValidateDSL(dslContent)
execResp, err := rustClient.ExecuteCase(caseName, "process")
// Save to database via DataService
```

**2.3 Update `internal/amend/amend.go`**

Replace:
```go
// OLD
parsedDSL, err := parser.Parse(strings.NewReader(latestVersion.DslSnapshot))
cases, err := parser.Bind(parsedDSL)
mutationFn(kycCase)
newSnapshot := parser.SerializeCases([]*model.KycCase{kycCase})
```

With:
```go
// NEW
rustClient := rustclient.NewDslClient("localhost:50060")
amendResp, err := rustClient.AmendCase(caseName, step)
// Save new version via DataService
```

### Phase 3: Testing (Day 4)

**3.1 Integration Tests**
```bash
# Test all CLI commands
./kycctl grammar
./kycctl sample_case.dsl
./kycctl validate TEST-CASE
./kycctl amend TEST-CASE --step=policy-discovery

# Test RAG commands (should still work)
./kycctl seed-metadata
./kycctl search-metadata "tax residency"
```

**3.2 Verify Rust Service Handles All Cases**
```bash
# Test each DSL file
./kycctl sample_case.dsl
./kycctl ontology_example.dsl
./kycctl ownership_case.dsl
./kycctl derived_attributes_example.dsl
```

**3.3 Performance Benchmarks**
```bash
# Compare Go vs Rust parsing speed
time ./kycctl sample_case.dsl  # Before
time ./kycctl sample_case.dsl  # After (with Rust)
```

### Phase 4: Code Deletion (Day 5)

**4.1 Delete Legacy Parser**
```bash
rm -rf internal/parser/
```

**4.2 Delete Legacy Engine**
```bash
rm -rf internal/engine/
```

**4.3 Delete Old gRPC Services**
```bash
rm -rf cmd/server/
rm internal/service/dsl_service.go
rm internal/service/kyc_case_service.go
```

**4.4 Update Dependencies**
```bash
# Remove unused dependencies from go.mod
go mod tidy
```

**4.5 Update Import Statements**
```bash
# Find and remove all imports of deleted packages
grep -r "internal/parser" . --include="*.go"
grep -r "internal/engine" . --include="*.go"
```

### Phase 5: Documentation Update (Day 6)

**5.1 Update README.md**
- Remove references to Go parser
- Update architecture diagram
- Document Rust service as primary DSL engine

**5.2 Update CLAUDE.md**
- Remove Go parser/engine from architecture
- Update port allocation
- Clarify Rust as computation layer

**5.3 Update Makefile**
```makefile
# Remove Go parser targets
# Add Rust service startup to common workflows
```

**5.4 Create Migration Guide**
- Document breaking changes
- Provide upgrade path for external clients

---

## Detailed Code Changes

### 1. New File: `internal/rustclient/dsl_client.go`

```go
package rustclient

import (
	"context"
	"fmt"
	"time"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DslClient struct {
	conn   *grpc.ClientConn
	client pb.DslServiceClient
}

func NewDslClient(addr string) (*DslClient, error) {
	conn, err := grpc.Dial(addr, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Rust DSL service: %w", err)
	}
	
	return &DslClient{
		conn:   conn,
		client: pb.NewDslServiceClient(conn),
	}, nil
}

func (c *DslClient) Close() error {
	return c.conn.Close()
}

func (c *DslClient) ParseDSL(dsl string) (*pb.ParseResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return c.client.Parse(ctx, &pb.ParseRequest{Dsl: dsl})
}

func (c *DslClient) ValidateDSL(dsl string) (*pb.ValidationResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return c.client.Validate(ctx, &pb.ValidateRequest{Dsl: dsl})
}

func (c *DslClient) ExecuteCase(caseID, function string) (*pb.ExecuteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	return c.client.Execute(ctx, &pb.ExecuteRequest{
		CaseId:       caseID,
		FunctionName: function,
	})
}

func (c *DslClient) AmendCase(caseName, amendType string) (*pb.AmendResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	return c.client.Amend(ctx, &pb.AmendRequest{
		CaseName:      caseName,
		AmendmentType: amendType,
	})
}

func (c *DslClient) GetGrammar() (*pb.GrammarResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return c.client.GetGrammar(ctx, &pb.GetGrammarRequest{})
}

func (c *DslClient) ListAmendments() (*pb.ListAmendmentsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return c.client.ListAmendments(ctx, &pb.ListAmendmentsRequest{})
}
```

### 2. Updated: `internal/cli/cli.go` (RunProcessCommand)

```go
func RunProcessCommand(filePath string) error {
	// Read DSL file
	dslContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Connect to Rust DSL service
	rustClient, err := rustclient.NewDslClient("localhost:50060")
	if err != nil {
		return fmt.Errorf("failed to connect to Rust DSL service: %w", err)
	}
	defer rustClient.Close()

	// Parse via Rust
	parseResp, err := rustClient.ParseDSL(string(dslContent))
	if err != nil || !parseResp.Success {
		return fmt.Errorf("parse error: %v", err)
	}

	// Validate via Rust
	valResult, err := rustClient.ValidateDSL(string(dslContent))
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !valResult.Valid {
		return fmt.Errorf("validation failed: %v", valResult.Errors)
	}
	fmt.Println("✅ DSL validated successfully via Rust service")

	// Connect to database for persistence
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	defer db.Close()

	// Extract case name from parse response
	if len(parseResp.Cases) == 0 {
		return fmt.Errorf("no cases parsed")
	}
	caseName := parseResp.Cases[0].Name

	// Display case info
	displayParsedCaseInfo(parseResp.Cases[0])

	// Save to database via storage layer
	if err := storage.SaveCaseVersion(db, caseName, string(dslContent)); err != nil {
		return fmt.Errorf("failed to save case: %w", err)
	}

	fmt.Printf("\n🧾 DSL snapshot stored successfully (case: %s)\n", caseName)
	return nil
}
```

### 3. Updated: `internal/amend/amend.go`

```go
func ApplyAmendment(db *sqlx.DB, caseName string, step string) error {
	// Connect to Rust DSL service
	rustClient, err := rustclient.NewDslClient("localhost:50060")
	if err != nil {
		return fmt.Errorf("failed to connect to Rust DSL service: %w", err)
	}
	defer rustClient.Close()

	// Apply amendment via Rust
	amendResp, err := rustClient.AmendCase(caseName, step)
	if err != nil || !amendResp.Success {
		return fmt.Errorf("amendment failed: %v", err)
	}

	// Save new version to database
	if err := storage.SaveCaseVersion(db, caseName, amendResp.UpdatedDsl); err != nil {
		return fmt.Errorf("failed to save amended version: %w", err)
	}

	// Log amendment
	if err := storage.InsertAmendment(db, caseName, step, "rust-applied", "Applied via Rust service"); err != nil {
		return fmt.Errorf("failed to log amendment: %w", err)
	}

	fmt.Printf("✅ Amendment applied: %s → %s\n", caseName, step)
	return nil
}
```

---

## Service Port Allocation (Post-Migration)

| Port  | Service                | Purpose                      | Status |
|-------|------------------------|------------------------------|--------|
| 50060 | Rust DSL Service       | Parse, validate, execute DSL | ✅ Primary |
| 50070 | Go Data Service        | Database access, ontology    | ✅ Active |
| 8080  | REST API (optional)    | HTTP gateway to gRPC         | 🔄 Update |
| ~~50051~~ | ~~Go gRPC Service~~ | ~~Legacy DSL service~~      | ❌ DELETE |

---

## Risk Mitigation

### Risk 1: Rust Service Downtime
**Mitigation:** Keep Go parser code in a `deprecated/` directory for 1 release cycle
```bash
mkdir deprecated
mv internal/parser deprecated/
mv internal/engine deprecated/
```

### Risk 2: Feature Parity Gaps
**Mitigation:** Feature comparison checklist
- [ ] Parse all DSL syntax
- [ ] Validate ownership structures
- [ ] Validate ontology references
- [ ] Handle all amendment types
- [ ] Support serialization
- [ ] Grammar retrieval

### Risk 3: Performance Regression
**Mitigation:** Benchmark suite
```bash
./scripts/benchmark_dsl_processing.sh
```

### Risk 4: External Client Breakage
**Mitigation:** Deprecation warnings + dual deployment
- Add deprecation warning to old service
- Run both services in parallel for 1 month
- Provide migration guide

---

## Testing Checklist

### Unit Tests
- [ ] Rust client wrapper tests
- [ ] CLI command tests
- [ ] Amendment system tests

### Integration Tests
- [ ] Parse sample_case.dsl via Rust
- [ ] Validate ontology_example.dsl
- [ ] Process ownership_case.dsl
- [ ] Apply all amendment types
- [ ] Round-trip serialize/parse

### Performance Tests
- [ ] Parse 100 cases (latency)
- [ ] Concurrent validation (throughput)
- [ ] gRPC connection pooling

### Regression Tests
- [ ] All existing test scripts pass
- [ ] RAG commands unaffected
- [ ] Database schemas unchanged
- [ ] CLI UX identical

---

## Rollback Plan

If critical issues arise:

1. **Immediate Rollback (< 1 hour)**
   ```bash
   git revert <migration-commit>
   make build
   ./kycctl sample_case.dsl
   ```

2. **Partial Rollback**
   - Restore `internal/parser/` from `deprecated/`
   - Update CLI to use local parser
   - Keep Rust service running alongside

3. **Full Rollback**
   - Restore all deleted code from git history
   - Revert Makefile and documentation
   - Resume dual Go/Rust operations

---

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Code reduction | -40% (parser/engine) | Lines of code |
| Build time | -20% | `make build` duration |
| Parse latency | < 50ms | `time kycctl parse` |
| Test coverage | > 80% | `go test -cover` |
| gRPC throughput | > 1000 req/s | Load test |
| Memory usage | < 100MB | Process RSS |

---

## Post-Migration Tasks

### Week 1: Monitoring
- [ ] Set up Rust service monitoring (Prometheus)
- [ ] Track gRPC error rates
- [ ] Monitor parse latency (p50, p95, p99)
- [ ] Log analysis for failures

### Week 2: Optimization
- [ ] Profile Rust service for bottlenecks
- [ ] Optimize gRPC connection pooling
- [ ] Add caching layer if needed
- [ ] Tune PostgreSQL queries

### Month 1: Documentation
- [ ] Update all technical docs
- [ ] Create video demo of new architecture
- [ ] Write blog post about migration
- [ ] Update API documentation

---

## Architecture Decision Records (ADRs)

### ADR-001: Rust for DSL Parsing
**Decision:** Use Rust as the primary DSL parsing engine  
**Rationale:** 
- Superior performance (3-5x faster parsing)
- Memory safety eliminates entire class of bugs
- Better type system for AST manipulation
- Easier to maintain single parser implementation

### ADR-002: Go for Data Access
**Decision:** Keep Go for all database operations  
**Rationale:**
- Mature PostgreSQL ecosystem (sqlx, pgx)
- Existing investment in Go storage layer
- No performance benefit from Rust for I/O-bound operations
- Team expertise in Go database programming

### ADR-003: gRPC as Integration Layer
**Decision:** Use gRPC for Go-Rust communication  
**Rationale:**
- Type-safe contract (protobuf)
- Excellent performance (HTTP/2)
- Language-agnostic
- Built-in service discovery support

---

## Timeline Summary

| Phase | Duration | Key Deliverable |
|-------|----------|-----------------|
| 1. Setup | 1 day | Rust client wrapper |
| 2. CLI Migration | 2 days | Updated CLI using Rust |
| 3. Testing | 1 day | All tests passing |
| 4. Code Deletion | 1 day | Clean codebase |
| 5. Documentation | 1 day | Updated docs |
| **Total** | **6 days** | **Production-ready migration** |

---

## Approval & Sign-off

- [ ] Technical Lead Review
- [ ] Security Review (gRPC connections)
- [ ] Performance Benchmarks Approved
- [ ] Stakeholder Sign-off
- [ ] Rollback Plan Tested

---

## Questions & Answers

**Q: Why not keep both implementations?**  
A: Maintenance burden, inconsistency risk, code bloat

**Q: What if we need Go-specific features?**  
A: Add them to the Go data service layer, not parsing layer

**Q: How do we handle breaking changes?**  
A: Protobuf versioning + deprecation warnings

**Q: What about offline/CLI-only mode?**  
A: Rust service can be embedded or run locally

---

## References

- [Rust DSL Core](rust/kyc_dsl_core/)
- [Rust DSL Service](rust/kyc_dsl_service/)
- [Proto Definitions](api/proto/dsl_service.proto)
- [Data Service Guide](DATA_SERVICE_GUIDE.md)
- [GRPC Services Documentation](GRPC_SERVICES_COMPLETE.md)

---

**Status:** Ready for Execution  
**Next Action:** Create `internal/rustclient/dsl_client.go`  
**Owner:** Engineering Team  
**Target Date:** Q1 2024