# CONTRACT-FIRST Development

**Status**: ENFORCED  
**Version**: 1.5  
**Principle**: All development is done against Protocol Buffer contracts, NOT the database.

---

## 🎯 Core Principle

> **"Code against interfaces (protobuf), not implementations (database)"**

All application logic MUST interact with services through Protocol Buffer-defined gRPC interfaces. Direct database access is forbidden outside the Data Service.

---

## Why Contract-First?

### 1. **Decoupling**
- Application logic doesn't care about database schema
- Can swap PostgreSQL for another database
- Services can be rewritten in different languages

### 2. **Testability**
- Mock gRPC services in tests
- No need for test databases
- Fast, reliable unit tests

### 3. **Documentation**
- Proto files ARE the documentation
- Self-describing APIs
- Tool support (grpcurl, grpc-ui, Postman)

### 4. **Versioning**
- API versions tracked in proto files
- Backward compatibility enforced
- Clear migration paths

### 5. **Polyglot Support**
- Any language can consume gRPC
- Generated clients in Go, Rust, Python, etc.
- No language-specific database drivers needed

---

## The Three Services

### 1. **Rust DSL Service** (Port 50060)
**Contract**: `api/proto/dsl_service.proto`

```protobuf
service DslService {
  rpc Parse(ParseRequest) returns (ParseResponse);
  rpc Validate(ValidateRequest) returns (ValidationResult);
  rpc Execute(ExecuteRequest) returns (ExecuteResponse);
  rpc Amend(AmendRequest) returns (AmendResponse);
}
```

**Responsibility**: DSL language operations (parse, compile, validate)

### 2. **Go Data Service** (Port 50070)
**Contracts**: 
- `api/proto/data_service.proto` (Dictionary & Cases)
- `api/proto/ontology_service.proto` (Ontology)

```protobuf
service CaseService {
  rpc SaveCaseVersion(CaseVersionRequest) returns (CaseVersionResponse);
  rpc GetCaseVersion(GetCaseRequest) returns (CaseVersion);
  rpc ListAllCases(ListCasesRequest) returns (CaseList);
  rpc ListCaseVersions(ListVersionsRequest) returns (VersionList);
}

service DictionaryService {
  rpc GetAttribute(GetAttributeRequest) returns (Attribute);
  rpc ListAttributes(ListAttributesRequest) returns (AttributeList);
}
```

**Responsibility**: ALL database operations

### 3. **Go CLI** (kycctl)
**Role**: Client of both services

```go
// ✅ Correct: Uses gRPC clients
rustClient := rustclient.NewDslClient("localhost:50060")
dataClient := dataclient.NewDataClient("localhost:50070")
```

---

## Development Workflow

### Step 1: Define the Contract (Proto File)

```bash
# Edit proto file FIRST
vim api/proto/case_service.proto
```

```protobuf
// Add new RPC
service CaseService {
  // Existing methods...
  
  // NEW: Search cases by status
  rpc SearchCasesByStatus(SearchRequest) returns (CaseList);
}

message SearchRequest {
  string status = 1;
  int32 limit = 2;
  int32 offset = 3;
}

message CaseList {
  repeated CaseMetadata cases = 1;
  int32 total_count = 2;
}
```

### Step 2: Generate Stubs

```bash
# Generate Go stubs
make proto

# Generate Rust stubs
cd rust && cargo build
```

### Step 3: Implement Server (Data Service)

```go
// internal/dataservice/case_service.go
func (s *DataService) SearchCasesByStatus(ctx context.Context, req *pb.SearchRequest) (*pb.CaseList, error) {
    // Database query here (ONLY place SQL is allowed)
    cases := []CaseMetadata{}
    err := s.db.Select(&cases, `
        SELECT name, status, last_updated 
        FROM kyc_cases 
        WHERE status = $1 
        LIMIT $2 OFFSET $3
    `, req.Status, req.Limit, req.Offset)
    
    if err != nil {
        return nil, err
    }
    
    // Convert to proto messages
    pbCases := []*pb.CaseMetadata{}
    for _, c := range cases {
        pbCases = append(pbCases, &pb.CaseMetadata{
            Name: c.Name,
            Status: c.Status,
            LastUpdated: timestamppb.New(c.LastUpdated),
        })
    }
    
    return &pb.CaseList{Cases: pbCases}, nil
}
```

### Step 4: Implement Client (CLI)

```go
// internal/dataclient/client.go
func (c *DataClient) SearchCasesByStatus(status string, limit int) ([]*pb.CaseMetadata, error) {
    req := &pb.SearchRequest{
        Status: status,
        Limit: int32(limit),
    }
    
    resp, err := c.caseClient.SearchCasesByStatus(context.Background(), req)
    if err != nil {
        return nil, err
    }
    
    return resp.Cases, nil
}
```

```go
// internal/cli/search.go
func RunSearchCommand(status string) error {
    client, err := dataclient.NewDataClient("")
    if err != nil {
        return err
    }
    defer client.Close()
    
    cases, err := client.SearchCasesByStatus(status, 10)
    if err != nil {
        return err
    }
    
    for _, c := range cases {
        fmt.Printf("%s - %s\n", c.Name, c.Status)
    }
    return nil
}
```

### Step 5: Test the Contract

```bash
# Test with grpcurl
grpcurl -plaintext -d '{"status": "pending", "limit": 10}' \
  localhost:50070 kyc.data.CaseService/SearchCasesByStatus

# Test with CLI
./kycctl search --status=pending
```

---

## ✅ Correct Patterns

### Adding a New Feature

```
1. Update proto file         ✅ Define contract
2. Generate stubs            ✅ make proto && cargo build
3. Implement in Data Service ✅ SQL here, and ONLY here
4. Implement client wrapper  ✅ dataclient/
5. Use in CLI                ✅ Uses dataclient
6. Test                      ✅ grpcurl + CLI
```

### Querying Data from CLI

```go
// ✅ CORRECT: Via gRPC client
func GetCase(caseName string) error {
    client, _ := dataclient.NewDataClient("")
    defer client.Close()
    
    caseData, err := client.GetCaseVersion(caseName, 0)
    fmt.Println(caseData.DslSource)
    return nil
}
```

### Rust Calling Data Service

```rust
// ✅ CORRECT: Via gRPC client
use tonic::Request;
use data_service_client::CaseServiceClient;

async fn save_case(&self, case_name: String, dsl: String) -> Result<i32> {
    let mut client = CaseServiceClient::connect("http://localhost:50070").await?;
    
    let request = Request::new(CaseVersionRequest {
        case_name,
        dsl_source: dsl,
    });
    
    let response = client.save_case_version(request).await?;
    Ok(response.into_inner().version)
}
```

---

## ❌ Anti-Patterns

### Direct Database Access in CLI

```go
// ❌ WRONG: Direct SQL in CLI
func GetCase(caseName string) error {
    db, _ := sqlx.Open("postgres", connStr)  // ❌ NO!
    defer db.Close()
    
    var dsl string
    db.Get(&dsl, "SELECT dsl_snapshot FROM cases WHERE name=$1", caseName)  // ❌ NO!
    fmt.Println(dsl)
    return nil
}
```

### Mixing Database and gRPC

```go
// ❌ WRONG: Some via gRPC, some via DB
func ProcessCase(caseName string) error {
    // Parse via gRPC (correct)
    rustClient.Parse(dsl)
    
    // ❌ But then save directly to DB (WRONG!)
    db, _ := storage.ConnectPostgres()
    db.Exec("INSERT INTO cases ...")
    
    return nil
}
```

### Bypassing Contracts

```go
// ❌ WRONG: Importing storage in CLI
import (
    "github.com/jmoiron/sqlx"  // ❌ NO!
    "github.com/adamtc007/KYC-DSL/internal/storage"  // ❌ NO!
)
```

---

## Testing Strategy

### Unit Tests (Mock gRPC)

```go
// cli_test.go
func TestGetCase(t *testing.T) {
    // Mock the gRPC client
    mockClient := &MockDataClient{
        GetCaseVersionFunc: func(name string, version int) (*pb.CaseVersion, error) {
            return &pb.CaseVersion{
                CaseName: "TEST",
                Version: 1,
                DslSource: "(test)",
            }, nil
        },
    }
    
    // Test CLI logic without needing database
    result := GetCaseWithClient(mockClient, "TEST")
    assert.NoError(t, result)
}
```

### Integration Tests (Real gRPC)

```go
// integration_test.go
func TestGetCaseIntegration(t *testing.T) {
    // Start real Data Service
    dataService := startDataService(t)
    defer dataService.Stop()
    
    // Use real gRPC client
    client, _ := dataclient.NewDataClient("localhost:50070")
    defer client.Close()
    
    // Test actual RPC call
    result, err := client.GetCaseVersion("TEST", 1)
    assert.NoError(t, err)
    assert.Equal(t, "TEST", result.CaseName)
}
```

### Contract Tests

```bash
# tests/contract/validate_contracts.sh
#!/bin/bash

# Ensure proto files compile
make proto
cd rust && cargo build

# Ensure services start
./bin/dataserver &
./rust/target/release/kyc_dsl_service &

# Test cross-service calls
grpcurl -plaintext localhost:50070 list
grpcurl -plaintext localhost:50060 list

# Test actual RPC
./kycctl get TEST-CASE

# All services must respond correctly
```

---

## Adding New Data Operations

### Example: Add "Delete Case" Feature

#### 1. Update Proto

```protobuf
// api/proto/case_service.proto
service CaseService {
  // ... existing methods ...
  
  rpc DeleteCase(DeleteCaseRequest) returns (DeleteCaseResponse);
}

message DeleteCaseRequest {
  string case_name = 1;
  bool force = 2;  // Delete all versions
}

message DeleteCaseResponse {
  bool success = 1;
  string message = 2;
  int32 versions_deleted = 3;
}
```

#### 2. Generate Stubs

```bash
make proto
cd rust && cargo build
```

#### 3. Implement Server

```go
// internal/dataservice/case_service.go
func (s *DataService) DeleteCase(ctx context.Context, req *pb.DeleteCaseRequest) (*pb.DeleteCaseResponse, error) {
    tx, _ := s.db.Begin()
    defer tx.Rollback()
    
    // Delete versions
    result, err := tx.Exec("DELETE FROM kyc_case_versions WHERE case_name = $1", req.CaseName)
    versionsDeleted := result.RowsAffected()
    
    // Delete case
    _, err = tx.Exec("DELETE FROM kyc_cases WHERE name = $1", req.CaseName)
    
    tx.Commit()
    
    return &pb.DeleteCaseResponse{
        Success: true,
        Message: fmt.Sprintf("Deleted case %s", req.CaseName),
        VersionsDeleted: int32(versionsDeleted),
    }, nil
}
```

#### 4. Implement Client

```go
// internal/dataclient/client.go
func (c *DataClient) DeleteCase(caseName string, force bool) error {
    req := &pb.DeleteCaseRequest{
        CaseName: caseName,
        Force: force,
    }
    
    resp, err := c.caseClient.DeleteCase(context.Background(), req)
    if err != nil {
        return err
    }
    
    fmt.Printf("✅ %s\n", resp.Message)
    return nil
}
```

#### 5. Add CLI Command

```go
// internal/cli/delete.go
func RunDeleteCommand(caseName string, force bool) error {
    client, _ := dataclient.NewDataClient("")
    defer client.Close()
    
    return client.DeleteCase(caseName, force)
}
```

#### 6. Test

```bash
# Manual test
grpcurl -plaintext -d '{"case_name": "TEST", "force": true}' \
  localhost:50070 kyc.data.CaseService/DeleteCase

# CLI test
./kycctl delete TEST-CASE --force
```

---

## Benefits Summary

| Aspect | Without Contracts | With Contracts |
|--------|------------------|----------------|
| **Coupling** | Tight (direct DB) | Loose (via API) |
| **Testing** | Needs test DB | Mock gRPC |
| **Documentation** | Scattered | Proto files |
| **Versioning** | Manual tracking | Proto versions |
| **Languages** | Go only | Any language |
| **Database** | Exposed everywhere | Encapsulated |
| **Changes** | Ripple through code | Contained |

---

## Enforcement

### Code Review Checklist

- [ ] Proto file updated FIRST
- [ ] Stubs regenerated (Go + Rust)
- [ ] Server implementation in Data Service
- [ ] Client wrapper added to dataclient/
- [ ] CLI uses dataclient, not storage
- [ ] No direct DB imports in CLI
- [ ] Contract tests pass

### Automated Checks

```bash
# Add to CI/CD
make lint-contracts
make test-contracts
make proto-check
```

### Import Restrictions

```yaml
# .golangci.yml
depguard:
  rules:
    cli-no-db:
      files:
        - "**/internal/cli/**"
        - "**/cmd/kycctl/**"
      deny:
        - pkg: "github.com/jmoiron/sqlx"
        - pkg: "github.com/lib/pq"
        - pkg: "github.com/adamtc007/KYC-DSL/internal/storage"
```

---

## Migration Checklist

### Existing Code (Fix These)

- [ ] CLI grammar command ✅ (already uses Rust gRPC)
- [ ] CLI process command ⚠️ (uses direct DB for storage)
- [ ] CLI amend command ⚠️ (uses direct DB)
- [ ] CLI ontology command ❌ (direct DB)
- [ ] CLI RAG commands ❌ (direct DB)
- [ ] REST API handlers ⚠️ (mixed)

### New Code (Follow These)

- [x] Always define proto FIRST
- [x] Implement in Data Service
- [x] Use dataclient in CLI
- [x] Never import storage in CLI
- [x] Test contracts

---

## Golden Rules

1. **Proto files are the source of truth**
2. **No SQL outside internal/dataservice/**
3. **No sqlx/lib/pq imports outside data service**
4. **CLI uses gRPC clients only**
5. **Rust uses gRPC clients for data access**
6. **All new features start with proto definition**
7. **Contract tests must pass before merge**

---

**Last Updated**: 2024  
**Version**: 1.5  
**Status**: ENFORCED  
**Principle**: Contract-First, Always.