# TODO: No Side Doors Refactor

**Status**: In Progress  
**Goal**: All database access through gRPC Data Service (port 50070)  
**Architecture**: CLI → gRPC → Data Service → PostgreSQL

---

## ✅ COMPLETED

- [x] Documentation created
  - [x] NO_SIDE_DOORS.md - Architecture policy
  - [x] CONTRACT_FIRST.md - Development principle
  - [x] PROTOBUF_VALIDATION.md - Contract testing guide
- [x] Data Service is running (port 50070)
- [x] Discovered existing RPCs:
  - CaseService: GetCaseVersion, ListCaseVersions, SaveCaseVersion
  - DictionaryService: GetAttribute, ListAttributes, GetDocument, ListDocuments
  - OntologyService: 25 RPCs for ontology operations
- [x] Created `internal/dataclient/` package with core methods
  - [x] GetAttribute, ListAttributes (Dictionary)
  - [x] SaveCaseVersion, GetCaseVersion, GetLatestCaseVersion (Cases)
  - [x] ListCaseVersions (Cases)
- [x] Created `internal/storage/case_retrieval.go` with helper functions
- [x] Data Service confirmed running (port 50070) ✅
- [x] **MIGRATED CLI COMMANDS** (2025-10-31):
  - [x] `RunGetCaseCommand()` - Now uses dataclient ✅
  - [x] `RunListCaseVersionsCommand()` - Now uses dataclient ✅
  - [x] Added CLI router entries for `get`, `versions`, `list` commands
  - [x] Fixed proto field name mismatches in dataclient
  - [x] Resolved DATABASE_URL connection issues

---

## 🚨 CURRENT STATUS (Updated)

**Services Running:**
- ✅ Data Service: Port 50070 (RUNNING) - Use `DATABASE_URL` env var
- ✅ Rust DSL Service: Port 50060 (RUNNING for testing)

**CLI Files Migration Status:**
- 🟢 `internal/cli/get_case.go` - **2 of 3 functions migrated!** ✅
  - ✅ `RunGetCaseCommand()` - Using gRPC
  - ✅ `RunListCaseVersionsCommand()` - Using gRPC
  - ❌ `RunListAllCasesCommand()` - Needs ListAllCases RPC first
- ❌ `internal/cli/search_metadata.go` - 5 functions use `storage.ConnectPostgres()`
- ❌ `internal/cli/seed_metadata.go` - 1 function uses `storage.ConnectPostgres()`

**Dataclient Status:**
- ✅ Package exists at `internal/dataclient/client.go`
- ✅ Core methods implemented (GetAttribute, ListAttributes, SaveCaseVersion, GetCaseVersion)
- ✅ ListCaseVersions added ✅
- ✅ Proto field names fixed (Id, CaseId, etc.) ✅
- ❌ Missing ListAllCases wrapper (RPC doesn't exist yet)
- ❌ No tests yet

**Key Discovery:**
- ⚠️ Data Service requires `DATABASE_URL` env var (not individual PG* vars)
- ⚠️ Table naming inconsistency: `kyc_case_versions` vs `case_versions`

---

## 🎯 NEXT SESSION - START HERE

### STEP 1: Complete dataclient Package (20 min)

**File**: `internal/dataclient/client.go`

Add missing wrapper methods:

```go
// ListCaseVersions retrieves all versions of a case
func (c *DataClient) ListCaseVersions(caseName string) ([]*pb.CaseVersion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	req := &pb.ListCaseVersionsRequest{
		CaseName: caseName,
	}

	resp, err := c.caseClient.ListCaseVersions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list case versions for %s: %w", caseName, err)
	}

	return resp.Versions, nil
}

// TODO: Add ListAllCases when RPC is available
```

---

### STEP 2: Migrate `internal/cli/get_case.go` (30 min)

**Priority**: This file has 3 functions that are good candidates for migration.

#### Function 1: RunGetCaseCommand

**BEFORE** (lines 10-53):
```go
func RunGetCaseCommand(caseName string, version int) error {
	db, err := storage.ConnectPostgres()  // ❌ REMOVE
	defer db.Close()
	
	// ... direct SQL calls ...
	dsl, hash, err = storage.GetCaseVersion(db, caseName, version)
	dsl, actualVersion, hash, err = storage.GetLatestCaseWithMetadata(db, caseName)
```

**AFTER**:
```go
func RunGetCaseCommand(caseName string, version int) error {
	client, err := dataclient.NewDataClient("")  // ✅ USE THIS
	if err != nil {
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer client.Close()

	var caseVersion *pb.CaseVersion
	if version > 0 {
		caseVersion, err = client.GetCaseVersion(caseName, int32(version))
	} else {
		caseVersion, err = client.GetLatestCaseVersion(caseName)
	}
	
	if err != nil {
		return fmt.Errorf("failed to retrieve case: %w", err)
	}

	fmt.Printf("📦 Case: %s\n", caseVersion.CaseName)
	fmt.Printf("📌 Version: %d\n", caseVersion.Version)
	fmt.Printf("🔑 Hash: %s\n", caseVersion.Hash)
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println(caseVersion.DslSource)
	
	return nil
}
```

#### Function 2: RunListCaseVersionsCommand

Replace `storage.ConnectPostgres()` with `dataclient.NewDataClient()`  
Replace `storage.ListCaseVersions(db, caseName)` with `client.ListCaseVersions(caseName)`

#### Function 3: RunListAllCasesCommand

**BLOCKER**: Needs `ListAllCases` RPC in Data Service  
**Action**: Skip for now OR add the RPC (see Step 4)

---

### STEP 3: Test Migrated Command (5 min)

```bash
# Ensure Data Service is running
lsof -i :50070

# If not running, start it
./bin/dataserver &

# Rebuild CLI with changes
make build

# Test the migrated command
./kycctl get AVIVA-EU-EQUITY-FUND

# Check for errors
echo $?  # Should be 0
```

**Expected Output:**
```
📦 Case: AVIVA-EU-EQUITY-FUND
📌 Version: 1
🔑 Hash: abc123...
─────────────────────────────────────────────
(kyc-case AVIVA-EU-EQUITY-FUND
  ...
)
```

---

### STEP 4: Add Missing RPC: ListAllCases (OPTIONAL - 45 min)

---

### 4. Add Missing RPC: ListAllCases (1 hour)

**Only do this if you have time and want to unblock `./kycctl list` command**

#### A. Find and update proto file

```bash
# Find the proto file
find . -name "*.proto" -exec grep -l "CaseService" {} \;
# Likely: api/proto/data_service.proto or proto_shared/data_service.proto
```

**Add to CaseService**:

```protobuf
service CaseService {
  // ... existing ...
  
  rpc ListAllCases(ListAllCasesRequest) returns (CaseList);
}

message ListAllCasesRequest {
  int32 limit = 1;
  int32 offset = 2;
  string status_filter = 3;  // optional
}

message CaseList {
  repeated CaseMetadata cases = 1;
  int32 total_count = 2;
}

message CaseMetadata {
  string case_id = 1;
  int32 version_count = 2;
  string status = 3;
  google.protobuf.Timestamp last_updated = 4;
}
```

#### B. Regenerate stubs

```bash
make proto
```

#### C. Implement in Data Service

**File**: `internal/dataservice/case_service.go`

```go
func (s *DataService) ListAllCases(ctx context.Context, req *pb.ListAllCasesRequest) (*pb.CaseList, error) {
    // Query database
    // Return CaseList
}
```

#### D. Test it

```bash
grpcurl -plaintext -d '{"limit": 10}' \
  localhost:50070 kyc.data.CaseService/ListAllCases
```

---

### 5. Migrate ONE CLI Command (30 min)

**Pick**: `./kycctl get CASE-NAME` (uses existing RPC)

#### Before:
```go
// internal/cli/get_case.go
func RunGetCaseCommand(caseName string) error {
    db := storage.ConnectPostgres()  // ❌ OLD WAY
    defer db.Close()
    // ... direct SQL ...
}
```

#### After:
```go
// internal/cli/get_case.go
func RunGetCaseCommand(caseName string) error {
    client := dataclient.NewDataClient("")  // ✅ NEW WAY
    defer client.Close()
    
    caseVersion, err := client.GetCaseVersion(caseName)
    if err != nil {
        return err
    }
    
    fmt.Printf("📦 Case: %s\n", caseVersion.CaseId)
    fmt.Printf("📌 Version: %d\n", caseVersion.Version)
    fmt.Println(caseVersion.DslSource)
    return nil
}
```

#### Test:
```bash
# Ensure Data Service is running
./bin/dataserver &

# Test migrated command
./kycctl get AVIVA-EU-EQUITY-FUND
```

---

## 📋 FULL MIGRATION CHECKLIST

### CLI Commands to Migrate

- [x] `./kycctl get CASE-NAME` - ✅ **COMPLETED** (Uses GetCaseVersion - RPC exists)
- [x] `./kycctl versions CASE-NAME` - ✅ **COMPLETED** (Uses ListCaseVersions - RPC exists)
- [ ] `./kycctl list` - ⚠️ NEEDS ListAllCases RPC (add it first) - 🎯 **NEXT PRIORITY**
- [ ] `./kycctl ontology` - Uses OntologyService/ListRegulations
- [ ] `./kycctl sample_case.dsl` - Uses SaveCaseVersion
- [ ] `./kycctl amend CASE --step=X` - Uses SaveCaseVersion + amendments
- [ ] `./kycctl search-metadata QUERY` - ⚠️ RAG (needs vector search RPC)
- [ ] `./kycctl seed-metadata` - ⚠️ RAG (needs metadata RPCs)
</text>

<old_text line=272>
| Task | Time | Status |
|------|------|--------|
| Document architecture | 2h | ✅ Done |
| Survey existing RPCs | 15m | ✅ Done |
| Complete dataclient | 30m | ✅ Done (ListCaseVersions added) |
| Add missing RPCs | 1h each | ⏳ Not Started (ListAllCases needed) |
| Migrate CLI commands | 20m each | 🔄 15% Done (2 of 13 functions) |
| Add contract tests | 1h | ⏳ Not Started |
| Cleanup & lint rules | 30m | ⏳ Not Started |

**Current Status**: 45% complete (Phase 1 finished!)

**What Just Happened**: 
- ✅ Migrated `RunGetCaseCommand()` and `RunListCaseVersionsCommand()`
- ✅ Fixed proto field name bugs in dataclient
- ✅ Resolved DATABASE_URL connection issues
- ✅ Commands tested and working!

**Immediate Next Step**: 
1. Add `ListAllCases` RPC to Data Service (Est. 45 min)
2. Migrate `RunListAllCasesCommand()` (Est. 15 min)
3. Then move to search_metadata.go

### Data Service RPCs to Add

- [ ] ListAllCases(ListAllCasesRequest) → CaseList
- [ ] SearchCases(SearchRequest) → CaseList
- [ ] DeleteCase(DeleteCaseRequest) → DeleteResponse
- [ ] GetAmendments(GetAmendmentsRequest) → AmendmentList

### Cleanup Tasks

- [ ] Remove `internal/storage` imports from CLI files
- [ ] Remove `sqlx` imports from CLI files
- [ ] Add linter rules to prevent future violations
- [ ] Update CLAUDE.md with new architecture

---

## 🧪 VALIDATION COMMANDS

After each migration, run:

```bash
# 1. Does it compile?
make build

# 2. Does Data Service work?
grpcurl -plaintext localhost:50070 list

# 3. Does the command work?
./kycctl <command>

# 4. No direct DB imports?
grep -r "storage.ConnectPostgres" internal/cli/
grep -r "sqlx.Open" internal/cli/
# Should return nothing!
```

---

## 🚀 QUICK START (Next Session)

```bash
# 1. Check what's running
lsof -i :50070  # Data Service
lsof -i :50060  # Rust Service

# 2. Start services if needed
./bin/dataserver &
./rust/target/release/kyc_dsl_service &

# 3. Test one RPC
grpcurl -plaintext -d '{"case_id": "TEST"}' \
  localhost:50070 kyc.data.CaseService/GetCaseVersion

# 4. Work on dataclient wrapper or migrate a command
```

---

## 📊 PROGRESS TRACKER

**Estimated Total Work**: 4-6 coding sessions

| Task | Time | Status |
|------|------|--------|
| Document architecture | 2h | ✅ Done |
| Survey existing RPCs | 15m | ✅ Done |
| Complete dataclient | 30m | 🔄 In Progress |
| Add missing RPCs | 1h each | ⏳ Not Started |
| Migrate CLI commands | 20m each | ⏳ Not Started |
| Add contract tests | 1h | ⏳ Not Started |
| Cleanup & lint rules | 30m | ⏳ Not Started |

**Current Status**: 20% complete

---

## 🎯 SUCCESS CRITERIA

✅ **Refactor Complete When:**
1. Zero `storage.ConnectPostgres()` calls in `internal/cli/`
2. Zero `sqlx` or `lib/pq` imports in `internal/cli/`
3. All CLI commands work via gRPC
4. Contract tests pass
5. Documentation updated

---

---

## 🎬 QUICK COPY-PASTE COMMANDS

```bash
# 1. Check services
lsof -i :50070  # Data Service (should be running)
lsof -i :50060  # Rust Service (optional for this work)

# 2. Start Data Service if needed
./bin/dataserver &

# 3. Check which CLI files need migration
grep -l "storage.ConnectPostgres" internal/cli/*.go

# 4. After making changes, rebuild
make build

# 5. Test
./kycctl get AVIVA-EU-EQUITY-FUND
```

---

**Last Updated**: 2025-10-31  
**Current Focus**: ✅ PHASE 1 COMPLETE! 2 CLI functions migrated successfully  
**Next Action**: Add `ListAllCases` RPC to Data Service, then migrate remaining functions

**Session Summary**: See `MIGRATION_SESSION_SUMMARY.md` for full details

---

## 📝 CONCRETE MIGRATION EXAMPLES

### Example 1: RunGetCaseCommand (internal/cli/get_case.go)

**BEFORE** (Current - Lines 10-53):
```go
func RunGetCaseCommand(caseName string, version int) error {
	// Connect to database
	db, err := storage.ConnectPostgres()  // ❌ REMOVE
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database: %v", closeErr)
		}
	}()

	// Get DSL based on version flag
	var dsl string
	var actualVersion int
	var hash string

	if version > 0 {
		// Get specific version
		dsl, hash, err = storage.GetCaseVersion(db, caseName, version)  // ❌ REMOVE
		if err != nil {
			return fmt.Errorf("failed to retrieve case '%s' version %d: %w", caseName, version, err)
		}
		actualVersion = version
	} else {
		// Get latest version
		dsl, actualVersion, hash, err = storage.GetLatestCaseWithMetadata(db, caseName)  // ❌ REMOVE
		if err != nil {
			return fmt.Errorf("failed to retrieve latest case '%s': %w", caseName, err)
		}
	}

	// Display metadata
	fmt.Printf("📦 Case: %s\n", caseName)
	fmt.Printf("📌 Version: %d\n", actualVersion)
	fmt.Printf("🔑 Hash: %s\n", hash)
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()

	// Display DSL content
	fmt.Println(dsl)
	fmt.Println()

	return nil
}
```

**AFTER** (Migrated):
```go
func RunGetCaseCommand(caseName string, version int) error {
	// Connect to data service
	client, err := dataclient.NewDataClient("")  // ✅ NEW WAY
	if err != nil {
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer client.Close()

	// Get case version via gRPC
	var caseVersion *pb.CaseVersion
	if version > 0 {
		caseVersion, err = client.GetCaseVersion(caseName, int32(version))  // ✅ gRPC
	} else {
		caseVersion, err = client.GetLatestCaseVersion(caseName)  // ✅ gRPC
	}
	
	if err != nil {
		return fmt.Errorf("failed to retrieve case: %w", err)
	}

	// Display metadata
	fmt.Printf("📦 Case: %s\n", caseVersion.CaseId)
	fmt.Printf("🔑 ID: %s\n", caseVersion.Id)
	fmt.Printf("📅 Created: %s\n", caseVersion.CreatedAt)
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()

	// Display DSL content
	fmt.Println(caseVersion.DslSource)
	fmt.Println()

	return nil
}
```

**Changes Required:**
1. Replace `storage.ConnectPostgres()` → `dataclient.NewDataClient("")`
2. Replace `storage.GetCaseVersion()` → `client.GetCaseVersion()`
3. Replace `storage.GetLatestCaseWithMetadata()` → `client.GetLatestCaseVersion()`
4. Change return types from `(dsl, hash, version)` → `*pb.CaseVersion` struct
5. Update imports: Add `dataclient` and `pb`, remove direct `storage` calls
6. ⚠️ **CRITICAL**: Use correct proto field names: `CaseId`, `DslSource`, `Id`, `CreatedAt`

---

### Example 2: RunListCaseVersionsCommand (internal/cli/get_case.go)

**BEFORE** (Current - Lines 56-98):
```go
func RunListCaseVersionsCommand(caseName string) error {
	// Connect to database
	db, err := storage.ConnectPostgres()  // ❌ REMOVE
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database: %v", closeErr)
		}
	}()

	// Get version list
	versions, err := storage.ListCaseVersions(db, caseName)  // ❌ REMOVE
	if err != nil {
		return fmt.Errorf("failed to list versions for case '%s': %w", caseName, err)
	}
	
	// ... display logic ...
}
```

**AFTER** (Migrated):
```go
func RunListCaseVersionsCommand(caseName string) error {
	// Connect to data service
	client, err := dataclient.NewDataClient("")  // ✅ NEW WAY
	if err != nil {
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer client.Close()

	// Get version list via gRPC
	versions, err := client.ListCaseVersions(caseName)  // ✅ gRPC
	if err != nil {
		return fmt.Errorf("failed to list versions for case '%s': %w", caseName, err)
	}
	
	// ... display logic remains the same ...
}
```

**Changes Required:**
1. Replace `storage.ConnectPostgres()` → `dataclient.NewDataClient("")`
2. Replace `storage.ListCaseVersions(db, caseName)` → `client.ListCaseVersions(caseName)`
3. **NOTE**: Need to add `ListCaseVersions()` method to `internal/dataclient/client.go` first!

---

### Example 3: Import Changes Required

**BEFORE** (Current - Lines 1-8):
```go
package cli

import (
	"fmt"
	"log"

	"github.com/adamtc007/KYC-DSL/internal/storage"  // ❌ REMOVE THIS LINE
)
```

**AFTER** (Migrated):
```go
package cli

import (
	"fmt"
	"log"

	pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"  // ✅ ADD
	"github.com/adamtc007/KYC-DSL/internal/dataclient"  // ✅ ADD
)
```

---

## 🎯 MIGRATION CHECKLIST FOR get_case.go

- [ ] 1. Add `ListCaseVersions()` method to `internal/dataclient/client.go`
- [ ] 2. Update imports in `internal/cli/get_case.go`
- [ ] 3. Migrate `RunGetCaseCommand()` (lines 10-53)
- [ ] 4. Migrate `RunListCaseVersionsCommand()` (lines 56-98)
- [ ] 5. Skip `RunListAllCasesCommand()` for now (needs new RPC)
- [ ] 6. **Double-check proto field names**: Use `CaseId` not `CaseName`!
- [ ] 7. Test: `make build && ./kycctl get AVIVA-EU-EQUITY-FUND`
- [ ] 8. Test: `./kycctl versions AVIVA-EU-EQUITY-FUND`
- [ ] 9. Verify: `grep "storage.ConnectPostgres" internal/cli/get_case.go` returns nothing

**Estimated Time**: 25 minutes

---

## 💡 TIPS FOR MIGRATION

### ⚠️ CRITICAL: Proto Field Naming

**The proto file uses `snake_case` which becomes `CamelCase` in Go:**

From `proto_shared/data_service.proto`:
```protobuf
message CaseVersion {
  string id = 1;
  string case_id = 2;           // ← becomes CaseId in Go
  string dsl_source = 3;        // ← becomes DslSource in Go
  string compiled_json = 4;
  string status = 5;
  string created_at = 6;
}
```

**In Go code, access like this:**
```go
caseVersion.CaseId      // ✅ NOT caseVersion.CaseName ❌
caseVersion.DslSource   // ✅ NOT caseVersion.Dsl ❌
caseVersion.Id          // ✅ NOT caseVersion.VersionId ❌
caseVersion.CreatedAt   // ✅ NOT caseVersion.CreateDate ❌
```

**⚠️ NOTE**: The proto doesn't have separate `version` (int) or `hash` fields!
- `id` field is the unique identifier (UUID/string)
- `case_id` is the case name
- If you need version numbers or hashes, they may need to be added to the proto

1. **Always connect to Data Service first**:
   ```go
   client, err := dataclient.NewDataClient("")
   if err != nil {
       return fmt.Errorf("failed to connect to data service: %w", err)
   }
   defer client.Close()
   ```

2. **Proto field mappings**:
   - Proto: `case_id` → Go: `CaseId` (string)
   - Proto: `dsl_source` → Go: `DslSource` (string)
   - Proto: `created_at` → Go: `CreatedAt` (string, not timestamp!)
   - All versions use `int32` (not `int`)
</text>

<old_text line=488>
	// Display metadata
	fmt.Printf("📦 Case: %s\n", caseVersion.CaseName)
	fmt.Printf("📌 Version: %d\n", caseVersion.Version)
	fmt.Printf("🔑 Hash: %s\n", caseVersion.Hash)
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()

	// Display DSL content
	fmt.Println(caseVersion.DslSource)

3. **No more SQL queries in CLI code** - everything goes through gRPC

4. **If RPC doesn't exist**:
   - Option A: Add it to Data Service (1 hour)
   - Option B: Skip that command for now

5. **Test after each function migration**:
   ```bash
   make build
   ./kycctl <command>
   ```

---

**End of TODO - Ready to Start Coding! 🚀**