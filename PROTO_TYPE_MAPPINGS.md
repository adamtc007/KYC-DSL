# Protocol Buffer Type Mappings

**Last Updated**: 2025-10-31  
**Purpose**: Reference guide for proto type mappings between Proto3, Go, and Rust

---

## Proto File Organization

### Directory Structure

```
KYC-DSL/
├── api/proto/                    # Rust DSL Service protos
│   ├── dsl_service.proto        # Main DSL operations (Rust server)
│   ├── kyc_case.proto           # Case data structures
│   ├── cbu_graph.proto          # Business unit graphs
│   └── rag_service.proto        # RAG/vector search
│
└── proto_shared/                 # Shared Go/Rust protos
    ├── data_service.proto       # Data/Dictionary/Case services (Go server)
    └── ontology_service.proto   # Ontology operations (Go server)
```

### Package Mappings

| Proto File | Package | Go Package | Rust Module |
|------------|---------|------------|-------------|
| `api/proto/dsl_service.proto` | `kyc.dsl` | `github.com/adamtc007/KYC-DSL/api/pb` | `kyc::dsl` |
| `api/proto/kyc_case.proto` | `kyc` | `github.com/adamtc007/KYC-DSL/api/pb` | `kyc` |
| `api/proto/cbu_graph.proto` | `kyc.cbu` | `github.com/adamtc007/KYC-DSL/api/pb` | `kyc::cbu` |
| `api/proto/rag_service.proto` | `kyc.rag` | `github.com/adamtc007/KYC-DSL/api/pb` | `kyc::rag` |
| `proto_shared/data_service.proto` | `kyc.data` | `github.com/adamtc007/KYC-DSL/api/pb/kycdata` | `kyc::data` |
| `proto_shared/ontology_service.proto` | `kyc.ontology` | `github.com/adamtc007/KYC-DSL/api/pb/kycontology` | `kyc::ontology` |

---

## Service Definitions

### Rust DSL Service (Port 50060)

**Proto**: `api/proto/dsl_service.proto`  
**Package**: `kyc.dsl`  
**Server**: Rust (`rust/kyc_dsl_service`)  
**Clients**: Go (`internal/rustclient`)

```protobuf
service DslService {
  rpc Execute(ExecuteRequest) returns (ExecuteResponse);
  rpc Validate(ValidateRequest) returns (ValidationResult);
  rpc Parse(ParseRequest) returns (ParseResponse);
  rpc Serialize(SerializeRequest) returns (SerializeResponse);
  rpc Amend(AmendRequest) returns (AmendResponse);
  rpc ListAmendments(ListAmendmentsRequest) returns (ListAmendmentsResponse);
  rpc GetGrammar(GetGrammarRequest) returns (GrammarResponse);
}
```

### Go Data Service (Port 50070)

**Proto**: `proto_shared/data_service.proto`  
**Package**: `kyc.data`  
**Server**: Go (`internal/dataservice`)  
**Clients**: Go (`internal/dataclient`), Rust (`rust/kyc_ontology_client`)

```protobuf
service DictionaryService {
  rpc GetAttribute(GetAttributeRequest) returns (Attribute);
  rpc ListAttributes(ListAttributesRequest) returns (AttributeList);
  rpc GetDocument(GetDocumentRequest) returns (Document);
  rpc ListDocuments(ListDocumentsRequest) returns (DocumentList);
}

service CaseService {
  rpc SaveCaseVersion(CaseVersionRequest) returns (CaseVersionResponse);
  rpc GetCaseVersion(GetCaseRequest) returns (CaseVersion);
  rpc ListCaseVersions(ListCaseVersionsRequest) returns (CaseVersionList);
  rpc ListAllCases(ListAllCasesRequest) returns (CaseList);
}
```

---

## Field Type Mappings

### Proto3 → Go → Rust

| Proto Type | Go Type | Rust Type | Notes |
|------------|---------|-----------|-------|
| `string` | `string` | `String` | UTF-8 encoded |
| `int32` | `int32` | `i32` | Signed 32-bit |
| `int64` | `int64` | `i64` | Signed 64-bit |
| `uint32` | `uint32` | `u32` | Unsigned 32-bit |
| `uint64` | `uint64` | `u64` | Unsigned 64-bit |
| `float` | `float32` | `f32` | IEEE 754 |
| `double` | `float64` | `f64` | IEEE 754 |
| `bool` | `bool` | `bool` | Boolean |
| `bytes` | `[]byte` | `Vec<u8>` | Binary data |
| `repeated T` | `[]T` | `Vec<T>` | List/array |
| `map<K,V>` | `map[K]V` | `HashMap<K,V>` | Hash map |
| `google.protobuf.Timestamp` | `*timestamppb.Timestamp` | `prost_types::Timestamp` | RFC 3339 |

### Snake Case → CamelCase Conversion

**Proto (snake_case)** → **Go (PascalCase)** → **Rust (snake_case)**

| Proto Field | Go Field | Rust Field |
|-------------|----------|------------|
| `case_id` | `CaseId` | `case_id` |
| `dsl_source` | `DslSource` | `dsl_source` |
| `version_count` | `VersionCount` | `version_count` |
| `last_updated` | `LastUpdated` | `last_updated` |
| `created_at` | `CreatedAt` | `created_at` |
| `compiled_json` | `CompiledJson` | `compiled_json` |

**⚠️ IMPORTANT**: Go uses PascalCase for exported fields, Rust keeps snake_case!

---

## Message Type Examples

### CaseVersion (data_service.proto)

**Proto Definition**:
```protobuf
message CaseVersion {
  string id = 1;
  string case_id = 2;
  string dsl_source = 3;
  string compiled_json = 4;
  string status = 5;
  string created_at = 6;
}
```

**Go Generated** (`api/pb/kycdata/data_service.pb.go`):
```go
type CaseVersion struct {
    Id           string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
    CaseId       string `protobuf:"bytes,2,opt,name=case_id,json=caseId,proto3" json:"case_id,omitempty"`
    DslSource    string `protobuf:"bytes,3,opt,name=dsl_source,json=dslSource,proto3" json:"dsl_source,omitempty"`
    CompiledJson string `protobuf:"bytes,4,opt,name=compiled_json,json=compiledJson,proto3" json:"compiled_json,omitempty"`
    Status       string `protobuf:"bytes,5,opt,name=status,proto3" json:"status,omitempty"`
    CreatedAt    string `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
}
```

**Go Usage**:
```go
cv := &pb.CaseVersion{
    CaseId:    "AVIVA-EU-EQUITY-FUND",  // ✅ PascalCase
    DslSource: "(kyc-case ...)",
    Status:    "approved",
}
fmt.Println(cv.CaseId)  // ✅ Access with PascalCase
```

**Rust Generated**:
```rust
pub struct CaseVersion {
    pub id: String,
    pub case_id: String,      // ✅ snake_case
    pub dsl_source: String,
    pub compiled_json: String,
    pub status: String,
    pub created_at: String,
}
```

**Rust Usage**:
```rust
let cv = CaseVersion {
    case_id: "AVIVA-EU-EQUITY-FUND".to_string(),  // ✅ snake_case
    dsl_source: "(kyc-case ...)".to_string(),
    status: "approved".to_string(),
    ..Default::default()
};
println!("{}", cv.case_id);  // ✅ Access with snake_case
```

---

## Common Type Patterns

### Optional Fields

**Proto**:
```protobuf
message Foo {
  string required_field = 1;
  string optional_field = 2;  // Empty string = unset in Proto3
}
```

**Go**: All fields are pointers or zero values
```go
foo := &pb.Foo{
    RequiredField: "value",
    OptionalField: "",  // Empty string means unset
}
```

**Rust**: Use `Option<T>` for true optionals
```rust
// With prost default behavior
pub struct Foo {
    pub required_field: String,
    pub optional_field: String,  // Empty = unset
}
```

### Repeated Fields (Lists)

**Proto**:
```protobuf
message CaseList {
  repeated CaseSummary cases = 1;
  int32 total_count = 2;
}
```

**Go**:
```go
list := &pb.CaseList{
    Cases:      []*pb.CaseSummary{},  // ✅ Slice of pointers
    TotalCount: 42,
}
```

**Rust**:
```rust
let list = CaseList {
    cases: vec![],  // ✅ Vec of owned structs
    total_count: 42,
};
```

### Maps

**Proto**:
```protobuf
message ExecuteRequest {
  string case_id = 1;
  map<string, string> arguments = 2;
}
```

**Go**:
```go
req := &pb.ExecuteRequest{
    CaseId:    "CASE-1",
    Arguments: map[string]string{  // ✅ Native Go map
        "key": "value",
    },
}
```

**Rust**:
```rust
use std::collections::HashMap;

let mut args = HashMap::new();
args.insert("key".to_string(), "value".to_string());

let req = ExecuteRequest {
    case_id: "CASE-1".to_string(),
    arguments: args,  // ✅ HashMap
};
```

### Timestamps

**Proto**:
```protobuf
import "google/protobuf/timestamp.proto";

message GrammarResponse {
  string ebnf = 1;
  google.protobuf.Timestamp created_at = 2;
}
```

**Go**:
```go
import "google.golang.org/protobuf/types/known/timestamppb"

resp := &pb.GrammarResponse{
    Ebnf:      "grammar text",
    CreatedAt: timestamppb.Now(),  // ✅ Use timestamppb package
}
```

**Rust**:
```rust
use prost_types::Timestamp;

let resp = GrammarResponse {
    ebnf: "grammar text".to_string(),
    created_at: Some(Timestamp {
        seconds: 1698765432,
        nanos: 0,
    }),
};
```

---

## Generation Commands

### Go Proto Generation

```bash
# Generate to api/pb/kycdata/
protoc --go_out=. --go_opt=module=github.com/adamtc007/KYC-DSL \
    --go-grpc_out=. --go-grpc_opt=module=github.com/adamtc007/KYC-DSL \
    proto_shared/data_service.proto

# Generate to api/pb/
protoc --go_out=api/pb --go_opt=paths=source_relative \
    --go-grpc_out=api/pb --go-grpc_opt=paths=source_relative \
    api/proto/*.proto

# Or use make
make proto
```

### Rust Proto Generation

Rust uses `build.rs` to compile protos at build time:

**File**: `rust/kyc_dsl_service/build.rs`
```rust
fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::configure()
        .build_server(true)
        .build_client(false)
        .compile_protos(&["../../api/proto/dsl_service.proto"], 
                       &["../../api/proto"])?;
    Ok(())
}
```

**Usage**:
```bash
cd rust && cargo build --release
# Proto code generated automatically in target/debug/build/
```

---

## Common Gotchas

### 1. Field Name Casing

❌ **WRONG** (Go):
```go
cv.case_id = "value"  // Error: undefined field
cv.CaseName = "value" // Error: field doesn't exist
```

✅ **CORRECT** (Go):
```go
cv.CaseId = "value"  // ✅ Proto case_id → Go CaseId
```

### 2. Proto Package vs Go Package

**Proto**:
```protobuf
package kyc.data;
option go_package = "github.com/adamtc007/KYC-DSL/api/pb/kycdata";
```

**Go Import** (use go_package path):
```go
import pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"  // ✅
// NOT: import "kyc.data"  // ❌
```

### 3. Zero Values vs Unset

In Proto3, there's no distinction between unset and zero value:
- Empty string `""` = unset
- `0` for numbers = unset
- `false` for bool = unset

**Workaround**: Use wrapper types or separate `has_field` booleans

### 4. Timestamp Handling

**Proto field**:
```protobuf
string created_at = 6;  // RFC3339 string
```

**Database** → **Go** → **Proto**:
```go
var createdAt time.Time
err := row.Scan(&createdAt)  // From DB

cv.CreatedAt = createdAt.Format(time.RFC3339)  // ✅ Convert to string
```

**Proto** → **Go** → **Display**:
```go
// Parse back to time.Time if needed
t, err := time.Parse(time.RFC3339, cv.CreatedAt)
```

### 5. Nil vs Empty Slices

**Go**: Distinguish between nil and empty
```go
var cases []*pb.CaseSummary  // nil slice
cases := []*pb.CaseSummary{} // empty slice (length 0)

// Both serialize to empty list in proto!
```

**Rust**: No nil concept
```rust
let cases: Vec<CaseSummary> = vec![];  // Always initialized
```

---

## Data Flow Examples

### Case Retrieval Flow

```
CLI (Go)
  │ client.GetCaseVersion("CASE-1", 0)
  ├─ pb.GetCaseRequest { CaseId: "CASE-1" }
  │
  ▼ gRPC → localhost:50070
  │
Data Service (Go)
  │ s.GetCaseVersion(ctx, req)
  ├─ DB.QueryRow("SELECT ... FROM case_versions WHERE case_id = $1")
  ├─ row.Scan(&cv.Id, &cv.CaseId, &cv.DslSource, ...)
  ├─ cv.CreatedAt = createdAt.Format(time.RFC3339)
  │
  ▼ return &pb.CaseVersion{...}
  │
CLI (Go)
  │ caseVersion := resp
  └─ fmt.Printf("Case: %s\n", caseVersion.CaseId)
```

### DSL Parsing Flow

```
CLI (Go)
  │ rustClient.ParseDSL(dslText)
  ├─ pb.ParseRequest { Dsl: "(kyc-case ...)" }
  │
  ▼ gRPC → localhost:50060
  │
Rust DSL Service
  │ self.parse(request)
  ├─ parser::parse(&req.dsl)  // Rust nom parser
  ├─ Convert AST → ParsedCase struct
  │
  ▼ return ParseResponse { cases: vec![...] }
  │
CLI (Go)
  │ resp := parseResp
  └─ for _, case := range resp.Cases { ... }
```

---

## Testing Proto Compatibility

### Go Client → Go Server
```bash
# Test data service
go run cmd/kycctl/main.go get CASE-NAME
```

### Go Client → Rust Server
```bash
# Test Rust DSL service
go run cmd/kycctl/main.go sample_case.dsl
```

### Rust Client → Go Server
```bash
# From Rust code
cd rust/kyc_ontology_client
cargo test
```

### Manual gRPC Testing
```bash
# List services
grpcurl -plaintext localhost:50070 list
grpcurl -plaintext localhost:50060 list

# Test RPC
grpcurl -plaintext -d '{"case_id": "TEST"}' \
  localhost:50070 kyc.data.CaseService/GetCaseVersion
```

---

## Best Practices

### 1. Always Use Generated Types
❌ Don't create your own structs that mirror proto messages  
✅ Import and use generated types

### 2. Consistent Naming
- Proto: `snake_case`
- Go: `PascalCase` for public, `camelCase` for private
- Rust: `snake_case` everywhere

### 3. Version Proto Files
```protobuf
// Add version to package or comments
package kyc.data.v1;  // Versioned package
```

### 4. Document Field Constraints
```protobuf
message CaseVersion {
  string id = 1;           // UUID format
  string case_id = 2;      // Alphanumeric, max 100 chars
  string status = 5;       // One of: draft, approved, declined
}
```

### 5. Use Buf or Prototool
Consider using `buf` for linting and breaking change detection:
```yaml
# buf.yaml
version: v1
lint:
  use:
    - DEFAULT
```

---

## References

- **Proto3 Language Guide**: https://protobuf.dev/programming-guides/proto3/
- **Go Generated Code**: https://protobuf.dev/reference/go/go-generated/
- **Rust prost**: https://github.com/tokio-rs/prost
- **gRPC Go**: https://grpc.io/docs/languages/go/
- **tonic (Rust)**: https://github.com/hyperium/tonic

---

**Maintained by**: KYC-DSL Development Team  
**Last Reviewed**: 2025-10-31