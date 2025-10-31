# Protocol Buffer Contract Validation

**Version**: 1.5  
**Status**: REQUIRED FOR ALL REFACTORING  
**Architecture**: Dual Go/Rust with Shared Protobuf

---

## 🎯 Purpose

When refactoring the "No Side Doors" architecture, we MUST validate that Protocol Buffer contracts work correctly between:
- **Go CLI** → **Go Data Service** (port 50070)
- **Go CLI** → **Rust DSL Service** (port 50060)
- **Rust DSL Service** → **Go Data Service** (port 50070)

---

## Architecture Contract Points

```
┌──────────────┐
│   CLI (Go)   │
└──────┬───────┘
       │
       │ ① gRPC/Protobuf
       │
       ├─────────────────────┬──────────────────┐
       │                     │                  │
       ▼                     ▼                  │
┌──────────────┐      ┌──────────────┐         │
│ Rust Service │  ②  │ Data Service │         │
│  (50060)     │─────>│  (50070)     │◄────────┘
│              │ gRPC │              │    ③
└──────────────┘      └──────┬───────┘
                             │
                             │ SQL
                             ▼
                       ┌──────────────┐
                       │  PostgreSQL  │
                       └──────────────┘

Contract Points:
① CLI (Go) ←→ Rust Service (proto: dsl_service.proto)
② Rust Service ←→ Data Service (proto: data_service.proto, case_service.proto)
③ CLI (Go) ←→ Data Service (proto: data_service.proto, case_service.proto)
```

---

## Validation Strategy

### Level 1: Proto Compilation ✅
**Validates**: Proto files are syntactically correct

```bash
# Generate Go stubs
protoc --go_out=. --go-grpc_out=. api/proto/*.proto

# Generate Rust stubs (via build.rs)
cd rust && cargo build
```

**Pass Criteria**: No compilation errors

---

### Level 2: Type Compatibility ✅
**Validates**: Generated types match between Go and Rust

```bash
# Check Go generated files exist
ls api/pb/*.pb.go

# Check Rust generated files exist
ls rust/target/debug/build/*/out/*.rs
```

**Pass Criteria**: All expected files generated

---

### Level 3: Service Discovery ✅
**Validates**: gRPC reflection works, services are discoverable

```bash
# Test Rust Service
grpcurl -plaintext localhost:50060 list

# Test Data Service
grpcurl -plaintext localhost:50070 list
```

**Expected Output**:
```
# Rust Service (50060)
grpc.reflection.v1.ServerReflection
kyc.dsl.DslService

# Data Service (50070)
grpc.reflection.v1.ServerReflection
kyc.data.DictionaryService
kyc.data.CaseService
kyc.ontology.OntologyService
```

---

### Level 4: Cross-Language Contract Tests 🎯
**Validates**: Go client can call Rust service, Rust client can call Go service

#### Test 1: Go CLI → Rust Service

```bash
#!/bin/bash
# tests/contract/test_go_to_rust.sh

echo "🧪 Testing: Go CLI → Rust DSL Service"

# Start Rust service
./rust/target/release/kyc_dsl_service &
RUST_PID=$!
sleep 2

# Test with Go CLI
./kycctl grammar

# Cleanup
kill $RUST_PID

if [ $? -eq 0 ]; then
    echo "✅ Go → Rust contract valid"
    exit 0
else
    echo "❌ Go → Rust contract broken"
    exit 1
fi
```

#### Test 2: Rust Service → Go Data Service

```bash
#!/bin/bash
# tests/contract/test_rust_to_go.sh

echo "🧪 Testing: Rust DSL Service → Go Data Service"

# Start Data Service
./bin/dataserver &
DATA_PID=$!
sleep 2

# Start Rust service (will try to call Data Service)
./rust/target/release/kyc_dsl_service &
RUST_PID=$!
sleep 2

# Make a call that triggers Rust → Go communication
grpcurl -plaintext -d '{
  "case_id": "TEST",
  "function_name": "DISCOVER-POLICIES"
}' localhost:50060 kyc.dsl.DslService/Execute

RESULT=$?

# Cleanup
kill $RUST_PID $DATA_PID

if [ $RESULT -eq 0 ]; then
    echo "✅ Rust → Go contract valid"
    exit 0
else
    echo "❌ Rust → Go contract broken"
    exit 1
fi
```

#### Test 3: Go CLI → Go Data Service

```bash
#!/bin/bash
# tests/contract/test_go_to_go.sh

echo "🧪 Testing: Go CLI → Go Data Service"

# Start Data Service
./bin/dataserver &
DATA_PID=$!
sleep 2

# Test with Go CLI
./kycctl list

RESULT=$?

# Cleanup
kill $DATA_PID

if [ $RESULT -eq 0 ]; then
    echo "✅ Go → Go contract valid"
    exit 0
else
    echo "❌ Go → Go contract broken"
    exit 1
fi
```

---

### Level 5: End-to-End Integration Tests 🎯
**Validates**: Complete workflows work across all services

```bash
#!/bin/bash
# tests/contract/test_e2e_flow.sh

echo "🧪 Testing: Complete E2E Flow"

# Start all services
./bin/dataserver &
DATA_PID=$!

./rust/target/release/kyc_dsl_service &
RUST_PID=$!

sleep 3

# Test complete flow: Parse → Validate → Store
echo "1️⃣ Processing DSL file..."
./kycctl sample_case.dsl

echo "2️⃣ Retrieving case..."
./kycctl get AVIVA-EU-EQUITY-FUND

echo "3️⃣ Amending case..."
./kycctl amend AVIVA-EU-EQUITY-FUND --step=policy-discovery

echo "4️⃣ Listing versions..."
./kycctl versions AVIVA-EU-EQUITY-FUND

RESULT=$?

# Cleanup
kill $RUST_PID $DATA_PID

if [ $RESULT -eq 0 ]; then
    echo "✅ E2E flow successful"
    exit 0
else
    echo "❌ E2E flow failed"
    exit 1
fi
```

---

## Validation Checklist

### Before Refactoring
- [ ] Document current proto contracts
- [ ] Create baseline contract tests
- [ ] Run all tests and capture output
- [ ] Tag current working state in git

### During Refactoring
- [ ] Make changes incrementally
- [ ] Run contract tests after each change
- [ ] Fix failures immediately
- [ ] Do not proceed if contracts break

### After Refactoring
- [ ] Run full contract test suite
- [ ] Compare against baseline
- [ ] Test all three contract points
- [ ] Update documentation

---

## Contract Test Makefile Targets

Add to `Makefile`:

```makefile
# Protobuf contract validation
.PHONY: test-contracts test-contract-go-rust test-contract-rust-go test-contract-go-go test-contract-e2e

test-contracts: test-contract-go-rust test-contract-rust-go test-contract-go-go test-contract-e2e
	@echo "✅ All contract tests passed"

test-contract-go-rust:
	@echo "🧪 Testing Go → Rust contract..."
	@./tests/contract/test_go_to_rust.sh

test-contract-rust-go:
	@echo "🧪 Testing Rust → Go contract..."
	@./tests/contract/test_rust_to_go.sh

test-contract-go-go:
	@echo "🧪 Testing Go → Go contract..."
	@./tests/contract/test_go_to_go.sh

test-contract-e2e:
	@echo "🧪 Testing E2E flow..."
	@./tests/contract/test_e2e_flow.sh

# Quick proto regeneration check
proto-check:
	@echo "🔍 Checking proto compilation..."
	@make proto
	@cd rust && cargo check -p kyc_dsl_service
	@echo "✅ Proto files compile successfully"
```

---

## Common Contract Failures

### 1. Field Name Mismatch

**Problem**: Go uses `case_id`, Rust expects `caseId`

```protobuf
// ❌ BAD: Inconsistent naming
message ExecuteRequest {
  string case_id = 1;     // Go sees this
  string caseId = 2;      // Rust sees this (WRONG!)
}

// ✅ GOOD: Consistent naming
message ExecuteRequest {
  string case_id = 1;     // Both see "case_id"
}
```

**Detection**: Rust compilation will fail with "unknown field"

---

### 2. Message Type Mismatch

**Problem**: Go sends `ExecuteRequest`, Rust expects `ExecuteRequestV2`

```protobuf
// ❌ BAD: Type changed without updating both sides
rpc Execute (ExecuteRequest) returns (ExecuteResponse);    // Go
rpc Execute (ExecuteRequestV2) returns (ExecuteResponse);  // Rust (WRONG!)

// ✅ GOOD: Same types
rpc Execute (ExecuteRequest) returns (ExecuteResponse);    // Both
```

**Detection**: gRPC will return "method not found"

---

### 3. Package Name Mismatch

**Problem**: Go uses `package kyc.dsl`, Rust uses `package kyc.dsl.v2`

```protobuf
// ❌ BAD: Different packages
package kyc.dsl;      // Go
package kyc.dsl.v2;   // Rust (WRONG!)

// ✅ GOOD: Same package
package kyc.dsl;      // Both
```

**Detection**: Service discovery will show different names

---

### 4. Enum Value Mismatch

**Problem**: Go adds enum value, Rust doesn't know about it

```protobuf
enum Status {
  PENDING = 0;
  APPROVED = 1;
  REJECTED = 2;
  // REVIEWED = 3;  // Added in Go, missing in Rust
}
```

**Detection**: Rust will receive unknown enum value

---

## Proto Change Workflow

### 1. Update Proto File
```bash
# Edit api/proto/dsl_service.proto
vim api/proto/dsl_service.proto
```

### 2. Regenerate Go Stubs
```bash
make proto
```

### 3. Regenerate Rust Stubs
```bash
cd rust
cargo clean
cargo build -p kyc_dsl_service
```

### 4. Validate Compilation
```bash
make proto-check
```

### 5. Run Contract Tests
```bash
make test-contracts
```

### 6. Fix Any Failures
```bash
# If Go code fails
vim internal/cli/cli.go

# If Rust code fails
vim rust/kyc_dsl_service/src/main.rs
```

### 7. Commit When All Pass
```bash
git add api/proto/ internal/ rust/
git commit -m "refactor: update proto contracts [contracts validated]"
```

---

## Continuous Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/contract-tests.yml
name: Contract Tests

on: [push, pull_request]

jobs:
  proto-contracts:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: kyc_dsl
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Setup Rust
        uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
      
      - name: Install protoc
        run: |
          sudo apt-get update
          sudo apt-get install -y protobuf-compiler
      
      - name: Install grpcurl
        run: |
          go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
      
      - name: Build Go services
        run: |
          make build
          make build-dataserver
      
      - name: Build Rust service
        run: |
          cd rust
          cargo build --release -p kyc_dsl_service
      
      - name: Run contract tests
        run: |
          make test-contracts
        env:
          PGHOST: localhost
          PGPORT: 5432
          PGUSER: postgres
          PGPASSWORD: postgres
          PGDATABASE: kyc_dsl
```

---

## Contract Test Output Examples

### ✅ Success Output

```
🧪 Testing: Go CLI → Rust DSL Service
Starting Rust DSL service...
🦀 Rust DSL gRPC Service
========================
Listening on: [::1]:50060

Testing grammar command...
✅ Grammar (v1.0) inserted into Postgres via Rust service.

✅ Go → Rust contract valid

🧪 Testing: Rust DSL Service → Go Data Service
Starting Data Service...
🚀 Starting KYC Data Service...
✅ Data Service initialized successfully

Starting Rust service...
Testing Execute RPC...
{
  "success": true,
  "message": "Executed function 'DISCOVER-POLICIES' on case 'TEST'",
  "newVersion": 1
}

✅ Rust → Go contract valid

🧪 Testing: Go CLI → Go Data Service
Starting Data Service...
Testing list command...
📋 Total Cases: 3

✅ Go → Go contract valid

✅ All contract tests passed
```

### ❌ Failure Output

```
🧪 Testing: Go CLI → Rust DSL Service
Starting Rust DSL service...
Error: rpc error: code = Unimplemented desc = unknown method GetGrammar for service kyc.dsl.DslService

❌ Go → Rust contract broken

Analysis:
  - Go client calling method: GetGrammar
  - Rust service doesn't implement: GetGrammar
  - Check: api/proto/dsl_service.proto
  - Ensure: Rust implements all RPCs
```

---

## Manual Testing Commands

```bash
# 1. Check service discovery
grpcurl -plaintext localhost:50060 list
grpcurl -plaintext localhost:50070 list

# 2. List available methods
grpcurl -plaintext localhost:50060 list kyc.dsl.DslService
grpcurl -plaintext localhost:50070 list kyc.data.CaseService

# 3. Describe a method
grpcurl -plaintext localhost:50060 describe kyc.dsl.DslService.Execute
grpcurl -plaintext localhost:50070 describe kyc.data.CaseService.SaveCaseVersion

# 4. Test a method
grpcurl -plaintext -d '{"case_id": "TEST", "function_name": "TEST"}' \
  localhost:50060 kyc.dsl.DslService/Execute

grpcurl -plaintext -d '{"case_name": "TEST", "dsl_source": "(test)"}' \
  localhost:50070 kyc.data.CaseService/SaveCaseVersion
```

---

## Best Practices

### 1. **Version Proto Files**
```protobuf
syntax = "proto3";

package kyc.dsl.v1;  // Include version

option go_package = "github.com/adamtc007/KYC-DSL/api/pb/v1;pb";
```

### 2. **Use Field Numbers Carefully**
```protobuf
message ExecuteRequest {
  string case_id = 1;        // Never change field number 1
  string function_name = 2;  // Never change field number 2
  // Never reuse field numbers!
}
```

### 3. **Add Fields, Don't Remove**
```protobuf
// ✅ SAFE: Adding optional fields
message ExecuteRequest {
  string case_id = 1;
  string function_name = 2;
  map<string, string> metadata = 3;  // NEW field (OK!)
}

// ❌ DANGEROUS: Removing fields
message ExecuteRequest {
  string case_id = 1;
  // string function_name = 2;  // REMOVED (BREAKS COMPATIBILITY!)
}
```

### 4. **Test After Every Proto Change**
```bash
# Always run after editing .proto files
make proto-check
make test-contracts
```

---

## Troubleshooting

### Problem: "unknown service"
**Cause**: Service not registered or reflection not enabled  
**Fix**: Ensure `reflection.Register(grpcServer)` in both services

### Problem: "unknown method"
**Cause**: Method name mismatch or not implemented  
**Fix**: Check proto file matches implementation

### Problem: "type mismatch"
**Cause**: Proto regeneration needed  
**Fix**: Run `make proto` and `cargo clean && cargo build`

### Problem: Connection refused
**Cause**: Service not running or wrong port  
**Fix**: Check service is running: `lsof -i :50060` or `lsof -i :50070`

---

## Summary

✅ **Before refactoring**: Run `make test-contracts`  
✅ **After proto changes**: Run `make proto-check && make test-contracts`  
✅ **Before merging PR**: All contract tests must pass  
✅ **In CI/CD**: Contract tests run automatically  

**Golden Rule**: If contract tests fail, the refactoring is not complete.

---

**Last Updated**: 2024  
**Enforced By**: CI/CD + Code Review  
**Test Coverage**: 3 contract points + E2E flow