# 🚀 START HERE - Next Coding Session

**Date**: 2025-01-20  
**Task**: Migrate CLI to use gRPC Data Service (No Side Doors Refactor)  
**Estimated Time**: 30-45 minutes  
**Full Details**: See `TODO_NO_SIDE_DOORS.md`

---

## ✅ PRE-FLIGHT CHECK (2 min)

```bash
# 1. Check if Data Service is running
lsof -i :50070

# 2. If not running, start it
./bin/dataserver &

# 3. Wait 2 seconds, then verify
sleep 2 && lsof -i :50070 | grep LISTEN

# 4. Check which files need migration
grep -l "storage.ConnectPostgres" internal/cli/*.go
```

**Expected Output:**
```
internal/cli/get_case.go
internal/cli/search_metadata.go
internal/cli/seed_metadata.go
```

---

## 🎯 TODAY'S GOAL

**Migrate**: `internal/cli/get_case.go` (2 functions)  
**Skip**: `RunListAllCasesCommand` (needs new RPC)  
**Result**: Remove all direct database access from get_case.go

---

## 📝 STEP 1: Add Missing Method to dataclient (5 min)

**File**: `internal/dataclient/client.go`

**Add this method at the end** (after `GetLatestCaseVersion`):

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
```

---

## 📝 STEP 2: Update Imports in get_case.go (2 min)

**File**: `internal/cli/get_case.go`

**REPLACE** lines 1-8:
```go
package cli

import (
	"fmt"
	"log"

	"github.com/adamtc007/KYC-DSL/internal/storage"  // ❌ REMOVE
)
```

**WITH**:
```go
package cli

import (
	"fmt"
	"log"

	pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"          // ✅ ADD
	"github.com/adamtc007/KYC-DSL/internal/dataclient"        // ✅ ADD
)
```

---

## 📝 STEP 3: Migrate RunGetCaseCommand (10 min)

**File**: `internal/cli/get_case.go` (lines 10-53)

**REPLACE** the entire function with:

```go
func RunGetCaseCommand(caseName string, version int) error {
	// Connect to data service
	client, err := dataclient.NewDataClient("")
	if err != nil {
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer client.Close()

	// Get case version via gRPC
	var caseVersion *pb.CaseVersion
	if version > 0 {
		caseVersion, err = client.GetCaseVersion(caseName, int32(version))
	} else {
		caseVersion, err = client.GetLatestCaseVersion(caseName)
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

---

## 📝 STEP 4: Migrate RunListCaseVersionsCommand (10 min)

**File**: `internal/cli/get_case.go` (lines 56-98)

**REPLACE** lines 56-71 with:

```go
func RunListCaseVersionsCommand(caseName string) error {
	// Connect to data service
	client, err := dataclient.NewDataClient("")
	if err != nil {
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer client.Close()

	// Get version list via gRPC
	versions, err := client.ListCaseVersions(caseName)
	if err != nil {
		return fmt.Errorf("failed to list versions for case '%s': %w", caseName, err)
	}

	// ... KEEP THE REST (display logic) ...
```

**Keep** the display logic (lines 72-98) unchanged.

---

## ✅ STEP 5: Test Your Changes (5 min)

```bash
# 1. Rebuild
make build

# 2. Test get command
./kycctl get AVIVA-EU-EQUITY-FUND

# 3. Test versions command  
./kycctl versions AVIVA-EU-EQUITY-FUND

# 4. Verify no direct DB access
grep "storage.ConnectPostgres" internal/cli/get_case.go
# Should return ONLY from RunListAllCasesCommand (which we're skipping)
```

---

## ✅ SUCCESS CRITERIA

- [ ] `make build` completes without errors
- [ ] `./kycctl get CASE-NAME` works and displays case info
- [ ] `./kycctl versions CASE-NAME` works and shows version list
- [ ] Only `RunListAllCasesCommand` still has `storage.ConnectPostgres`
- [ ] No compile errors about undefined methods

---

## ⚠️ COMMON ISSUES

**Issue**: "undefined: pb"  
**Fix**: Add import: `pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"`

**Issue**: "client.ListCaseVersions undefined"  
**Fix**: Did you add the method to `internal/dataclient/client.go`?

**Issue**: "caseVersion.CaseName undefined"  
**Fix**: Use `caseVersion.CaseId` (proto field names!)

**Issue**: Data Service not responding  
**Fix**: `./bin/dataserver &` then wait 2 seconds

---

## 📊 PROGRESS AFTER THIS SESSION

- ✅ `RunGetCaseCommand` - Migrated
- ✅ `RunListCaseVersionsCommand` - Migrated  
- ⏭️ `RunListAllCasesCommand` - Skipped (needs new RPC)
- ⏳ `internal/cli/search_metadata.go` - Next session
- ⏳ `internal/cli/seed_metadata.go` - Next session

**Progress**: 30% → 45% complete

---

## 🎉 DONE? COMMIT YOUR WORK

```bash
git add internal/dataclient/client.go internal/cli/get_case.go
git commit -m "refactor: migrate get_case.go to use gRPC data service

- Add ListCaseVersions method to dataclient
- Migrate RunGetCaseCommand to use dataclient
- Migrate RunListCaseVersionsCommand to use dataclient
- Remove direct database access (storage.ConnectPostgres)
- Part of No Side Doors refactor

Ref: TODO_NO_SIDE_DOORS.md"
```

---

## 📚 REFERENCES

- **Full TODO**: `TODO_NO_SIDE_DOORS.md`
- **Architecture**: `NO_SIDE_DOORS.md`
- **Proto Definitions**: `proto_shared/data_service.proto`
- **Dataclient Package**: `internal/dataclient/client.go`

---

**You got this! 💪 Start with Step 1 and work through sequentially.**