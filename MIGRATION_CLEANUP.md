# Go vs Rust Code Migration Status

## 🎯 Current Architecture Status

### ✅ COMPLETED: Data Layer (Go)
- **Location**: `internal/dataservice/`, `cmd/dataserver/`
- **Purpose**: Single source of truth for PostgreSQL
- **Port**: 50070
- **Services**:
  - `kyc.data.DictionaryService` - Basic dictionary operations
  - `kyc.data.CaseService` - Case version storage
  - `kyc.ontology.OntologyService` - Full ontology API
- **Status**: ✅ Production-ready, no conflicts

### ⚠️ RESIDUAL: Go DSL Parser/Engine (DEPRECATED)
- **Location**: `internal/parser/`, `internal/engine/`, `internal/service/dsl_service.go`
- **Used by**: `cmd/server/main.go` (port 50051 - OLD gRPC server)
- **Problem**: Duplicates Rust functionality
- **Status**: ⚠️ Should be removed or marked deprecated

**Files to Review/Remove**:
```
internal/parser/parser.go           (5.7 KB) - DSL parsing
internal/parser/binder.go           (5.7 KB) - AST binding
internal/parser/validator.go        (4.7 KB) - Validation
internal/parser/validator_ontology.go (9.2 KB) - Ontology validation
internal/parser/serializer.go       (4.1 KB) - Serialization
internal/parser/grammar.go          (1.6 KB) - Grammar definition
internal/engine/engine.go           (998 B)  - Execution engine
internal/service/dsl_service.go     - gRPC wrapper (USES Go parser)
internal/service/kyc_case_service.go - Case service (USES Go parser)
```

### ✅ COMPLETED: Rust DSL Engine
- **Location**: `rust/kyc_dsl_core/`, `rust/kyc_dsl_service/`
- **Purpose**: DSL parsing, compilation, execution
- **Port**: 50060
- **Status**: ✅ Production-ready

**Rust Implementation**:
```
rust/kyc_dsl_core/src/
  ├── parser.rs       - nom-based S-expression parser
  ├── compiler.rs     - AST → instruction compilation
  ├── executor.rs     - Stateful execution engine
  └── lib.rs          - Public API

rust/kyc_dsl_service/src/
  └── main.rs         - gRPC service (port 50060)
```

## 📊 Comparison

| Feature | Go Implementation | Rust Implementation | Winner |
|---------|-------------------|---------------------|--------|
| **DSL Parsing** | `internal/parser/parser.go` | `kyc_dsl_core/parser.rs` | 🦀 Rust |
| **Validation** | `internal/parser/validator.go` | Built into executor | 🦀 Rust |
| **Execution** | `internal/engine/engine.go` | `kyc_dsl_core/executor.rs` | 🦀 Rust |
| **Data Access** | `internal/storage/` (direct SQL) | gRPC client → Go service | 🐹 Go |
| **Type Safety** | Runtime checks | Compile-time guarantees | 🦀 Rust |
| **Performance** | Good | Excellent | 🦀 Rust |

## 🔧 Recommended Actions

### 1. Deprecate Go DSL Code (HIGH PRIORITY)
```bash
# Mark files as deprecated
echo "// DEPRECATED: Use Rust DSL service on port 50060" > internal/parser/DEPRECATED.md
echo "// DEPRECATED: Use Rust DSL service on port 50060" > internal/engine/DEPRECATED.md

# Or remove entirely
rm -rf internal/parser/
rm -rf internal/engine/
```

### 2. Update Old gRPC Server (cmd/server/main.go)
Currently serves on port 50051 and uses Go parser. Options:

**Option A**: Remove it entirely (recommended)
```bash
rm -rf cmd/server/
```

**Option B**: Redirect to Rust service
```go
// cmd/server/main.go - proxy to Rust
type DslService struct {
    rustClient dsl_service_client.DslServiceClient // Port 50060
}

func (s *DslService) Execute(ctx, req) {
    // Forward to Rust service
    return s.rustClient.Execute(ctx, req)
}
```

### 3. Update internal/service/
The `internal/service/` directory has OLD gRPC services that use Go parser:
- `dsl_service.go` - DEPRECATED (uses Go parser)
- `kyc_case_service.go` - DEPRECATED (uses Go parser)
- `rag_service.go` - ⚠️ Needs review
- `cbu_graph_service.go` - ✅ OK (just data access)

**Action**: Either remove or update to call Rust service via gRPC.

### 4. CLI Updates (cmd/kycctl/)
Current CLI might use Go parser. Should use gRPC clients:
```go
// OLD (if it exists)
import "github.com/adamtc007/KYC-DSL/internal/parser"

// NEW
import pb "github.com/adamtc007/KYC-DSL/api/pb/kycontology"
client := pb.NewOntologyServiceClient(conn)
```

## 🏗️ Target Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     PostgreSQL Database                      │
│                   (Single Source of Truth)                   │
└─────────────────────────────────────────────────────────────┘
                              ▲
                              │ pgxpool
                              │
                     ┌────────────────────┐
                     │   Go Data Service  │  Port 50070
                     │  (OntologyService) │
                     └────────────────────┘
                       ▲                ▲
                       │                │
            gRPC       │                │  gRPC
       ┌───────────────┘                └──────────────────┐
       │                                                    │
┌──────────────────┐                            ┌──────────────────┐
│ Rust DSL Engine  │  Port 50060                │  Go CLI/UI       │
│ (parse/execute)  │                            │  (admin tools)   │
│                  │                            │                  │
│ - Parser         │◄────gRPC─────────────────► │ - Query data     │
│ - Compiler       │   (get ontology data)      │ - Admin ops      │
│ - Executor       │                            │                  │
└──────────────────┘                            └──────────────────┘
```

## ✅ Clean Separation of Concerns

| Layer | Responsibility | Technology | Port |
|-------|----------------|------------|------|
| **Data** | PostgreSQL access, ontology, versioning | Go (pgx) | 50070 |
| **Compute** | DSL parsing, validation, execution | Rust (nom) | 50060 |
| **CLI** | User interface, admin tools | Go (gRPC client) | N/A |

## 🚨 Conflicts to Resolve

1. **Port 50051** - Old Go gRPC server with deprecated parser
   - **Action**: Remove or redirect to Rust (port 50060)

2. **internal/service/** - Old services using Go parser
   - **Action**: Remove or update to proxy

3. **internal/parser/** & **internal/engine/** - Duplicate Rust code
   - **Action**: Remove entirely

4. **cmd/server/main.go** - Registers old services
   - **Action**: Remove or refactor

## 📝 Migration Checklist

- [ ] Audit all imports of `internal/parser`
- [ ] Audit all imports of `internal/engine`
- [ ] Remove or deprecate `internal/parser/`
- [ ] Remove or deprecate `internal/engine/`
- [ ] Update `internal/service/dsl_service.go` to proxy to Rust
- [ ] Update `internal/service/kyc_case_service.go`
- [ ] Update `cmd/server/main.go` or remove it
- [ ] Ensure `cmd/kycctl/` uses gRPC clients only
- [ ] Update documentation (README, CLAUDE.md)
- [ ] Add deprecation warnings to old code
- [ ] Create migration guide for users

## 🎯 End State

After cleanup:
```
KYC-DSL/
├── cmd/
│   ├── dataserver/       ✅ Data service (port 50070)
│   └── kycctl/           ✅ CLI (gRPC client)
├── internal/
│   ├── dataservice/      ✅ Data layer only
│   └── cli/              ✅ CLI logic
├── rust/
│   ├── kyc_dsl_core/     ✅ Core DSL engine
│   ├── kyc_dsl_service/  ✅ gRPC service (port 50060)
│   └── kyc_ontology_client/ ✅ Rust gRPC client
└── proto_shared/         ✅ Shared proto definitions
```

**NO MORE**: `internal/parser/`, `internal/engine/`, `cmd/server/` (old)

## 🔍 Current Usage Analysis

Files still importing deprecated Go parser/engine:

1. **internal/cli/cli.go** - CLI implementation
2. **internal/amend/amend.go** - Amendment system
3. **internal/service/kyc_case_service.go** - Case service (port 50051)
4. **internal/service/dsl_service.go** - DSL service (port 50051)

## 🎯 Immediate Action Plan

### Step 1: Mark as Deprecated
```bash
# Add deprecation notices
cat > internal/parser/DEPRECATED.txt << 'NOTICE'
⚠️ DEPRECATED: This Go parser is replaced by Rust implementation

Use the Rust DSL service instead:
- Port: 50060
- Location: rust/kyc_dsl_service/
- gRPC API: api/proto/dsl_service.proto

The Rust implementation provides:
- Better performance
- Type safety
- Memory safety
- Modern parser (nom-based)
NOTICE

cat > internal/engine/DEPRECATED.txt << 'NOTICE'
⚠️ DEPRECATED: This Go engine is replaced by Rust implementation

Use the Rust DSL service instead (port 50060)
NOTICE
```

### Step 2: Clean Architecture Decision

**OPTION A: Nuclear Option (Recommended for clean slate)**
```bash
# Remove all deprecated Go DSL code
rm -rf internal/parser/
rm -rf internal/engine/
rm -rf internal/service/   # Old services
rm -rf cmd/server/          # Old server (port 50051)

# Keep only:
# - internal/dataservice/ (NEW)
# - cmd/dataserver/ (NEW - port 50070)
```

**OPTION B: Gradual Migration**
```bash
# Move old code to deprecated directory
mkdir -p deprecated/go-dsl
mv internal/parser deprecated/go-dsl/
mv internal/engine deprecated/go-dsl/
mv internal/service deprecated/go-dsl/
mv cmd/server deprecated/go-dsl/

# Update imports to fail with helpful message
```

**OPTION C: Keep as Reference (Not Recommended)**
```bash
# Just add deprecation warnings in each file
for f in internal/parser/*.go; do
  sed -i '' '1i\
// DEPRECATED: Use Rust DSL service on port 50060\
' "$f"
done
```

## 📊 Port Usage Summary

| Port | Service | Status | Action |
|------|---------|--------|--------|
| 50051 | Old Go gRPC Server | ⚠️ DEPRECATED | Remove or disable |
| 50060 | Rust DSL Service | ✅ ACTIVE | Keep, this is correct |
| 50070 | Go Data Service (NEW) | ✅ ACTIVE | Keep, this is correct |
| 8080 | REST API | ❓ Unknown | Need to check |

## ✅ Recommended Final State

```
Production Services:
├── Port 50060: Rust DSL Engine (parse/compile/execute)
├── Port 50070: Go Data Service (PostgreSQL access)
└── Port 8080: REST Gateway (optional)

Development Tools:
├── cmd/kycctl: Go CLI (gRPC client)
└── rust/kyc_ontology_client: Rust CLI (gRPC client)

Removed:
├── ❌ Port 50051 (old Go gRPC server)
├── ❌ internal/parser/ (replaced by Rust)
├── ❌ internal/engine/ (replaced by Rust)
└── ❌ internal/service/ (old services using Go parser)
```

## 🚀 Next Steps

1. **Decide**: Choose Option A, B, or C above
2. **Clean**: Remove/move deprecated code
3. **Update**: Fix remaining imports (cli.go, amend.go)
4. **Test**: Ensure Rust service works
5. **Document**: Update README and CLAUDE.md
6. **Deploy**: Roll out clean architecture

