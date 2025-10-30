# Rust gRPC Service Manual Testing

## Quick Start

### 1. Build the Service

```bash
cd rust
cargo build --release -p kyc_dsl_service
cd ..
```

### 2. Start the Service

In **Terminal 1**:
```bash
./rust/target/release/kyc_dsl_service
```

You should see:
```
🦀 Rust DSL gRPC Service
========================
Listening on: [::1]:50060
Protocol: gRPC (HTTP/2)
Service: kyc.dsl.DslService

Available RPCs:
  - Execute
  - Validate
  - Parse
  - Serialize
  - Amend
  - ListAmendments
  - GetGrammar

Ready to accept connections...
```

### 3. Test the Service

In **Terminal 2**, run these commands:

#### Test 1: Service Discovery
```bash
grpcurl -plaintext localhost:50060 list
```

Expected output:
```
grpc.reflection.v1.ServerReflection
kyc.dsl.DslService
```

#### Test 2: List Methods
```bash
grpcurl -plaintext localhost:50060 list kyc.dsl.DslService
```

Expected output:
```
kyc.dsl.DslService.Amend
kyc.dsl.DslService.Execute
kyc.dsl.DslService.GetGrammar
kyc.dsl.DslService.ListAmendments
kyc.dsl.DslService.Parse
kyc.dsl.DslService.Serialize
kyc.dsl.DslService.Validate
```

#### Test 3: Execute Function
```bash
grpcurl -plaintext -d '{
  "case_id": "TEST-CASE-001",
  "function_name": "DISCOVER-POLICIES"
}' localhost:50060 kyc.dsl.DslService/Execute
```

Expected response:
```json
{
  "updatedDsl": "(kyc-case TEST-CASE-001 (function DISCOVER-POLICIES))",
  "message": "Executed function 'DISCOVER-POLICIES' on case 'TEST-CASE-001'",
  "success": true,
  "caseId": "TEST-CASE-001",
  "newVersion": 1
}
```

#### Test 4: Validate DSL
```bash
grpcurl -plaintext -d '{
  "dsl_source": "(kyc-case TEST (nature-purpose (nature \"test\") (purpose \"test\")))"
}' localhost:50060 kyc.dsl.DslService/Validate
```

Expected response:
```json
{
  "valid": true,
  "errors": [],
  "warnings": [],
  "issues": []
}
```

#### Test 5: Parse DSL
```bash
grpcurl -plaintext -d '{
  "dsl_source": "(kyc-case PARSE-TEST (function VERIFY))"
}' localhost:50060 kyc.dsl.DslService/Parse
```

Expected response:
```json
{
  "success": true,
  "message": "Parsed successfully",
  "cases": [
    {
      "name": "PARSE-TEST",
      "naturePurpose": {},
      "clientBusinessUnit": "",
      "policy": "",
      "obligation": "",
      "kycToken": "",
      "functions": []
    }
  ],
  "errors": []
}
```

## Verification Checklist

- [ ] Service starts without errors
- [ ] gRPC reflection works (service discovery)
- [ ] Execute RPC returns success
- [ ] Validate RPC accepts DSL source
- [ ] Parse RPC parses S-expressions
- [ ] All RPCs return proper JSON responses

## Troubleshooting

### Port Already in Use
```bash
lsof -i :50060
kill -9 <PID>
```

### Service Won't Start
```bash
# Check for build errors
cd rust
cargo check -p kyc_dsl_service

# Look at detailed logs
RUST_LOG=debug ./target/release/kyc_dsl_service
```

### grpcurl Not Found
```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

## Port Configuration

- **Rust gRPC Service**: `[::1]:50060` (IPv6 localhost)
- **Go gRPC Service**: `localhost:50051`
- **Go REST API**: `localhost:8080`

To connect from other machines, change `[::1]:50060` to `0.0.0.0:50060` in `main.rs`.

## Next Steps

Once all tests pass:

1. ✅ Test Go → Rust interop
2. ✅ Begin validator + audit chain phase
3. ✅ Add integration tests
4. ✅ Performance benchmarking

## Status

- **Build**: ✅ Compiles without warnings
- **Reflection**: ✅ Enabled with tonic-reflection
- **Proto Path**: ✅ Correctly references `../../api/proto/dsl_service.proto`
- **Dependencies**: ✅ All versions aligned
- **Integration**: Ready for testing