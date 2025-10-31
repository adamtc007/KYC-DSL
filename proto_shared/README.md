# proto_shared - Shared Protocol Buffer Definitions

This directory contains Protocol Buffer definitions that are **shared between Go and Rust** implementations.

---

## Purpose

Protocol Buffers in this directory define gRPC APIs that can be consumed by:
- **Go** services and clients
- **Rust** services and clients
- Any other language supporting gRPC (Python, Java, etc.)

Unlike `api/proto/` which is Go-specific, `proto_shared/` is designed for cross-language compatibility.

---

## Current Services

### data_service.proto

**Port**: 50070  
**Implemented in**: Go  
**Consumers**: Go CLI, Rust DSL Engine, UI/Frontend

Services:
- `DictionaryService` - Ontology data (attributes, documents)
- `CaseService` - Case version management

---

## Generating Bindings

### Go (Already Generated)

```bash
# From repository root
make proto-data

# Or manually:
protoc --go_out=. --go-grpc_out=. \
  --go_opt=module=github.com/adamtc007/KYC-DSL \
  --go-grpc_opt=module=github.com/adamtc007/KYC-DSL \
  proto_shared/data_service.proto
```

**Output**:
- `api/pb/kycdata/data_service.pb.go`
- `api/pb/kycdata/data_service_grpc.pb.go`

### Rust (TODO)

#### Option 1: Using tonic-build (Recommended)

1. Add to `rust/Cargo.toml`:
   ```toml
   [build-dependencies]
   tonic-build = "0.11"
   ```

2. Create `rust/build.rs`:
   ```rust
   fn main() -> Result<(), Box<dyn std::error::Error>> {
       tonic_build::compile_protos("../proto_shared/data_service.proto")?;
       Ok(())
   }
   ```

3. Build:
   ```bash
   cd rust
   cargo build
   ```

4. Use in code:
   ```rust
   pub mod kyc_data {
       tonic::include_proto!("kyc.data");
   }
   
   use kyc_data::dictionary_service_client::DictionaryServiceClient;
   ```

#### Option 2: Manual Generation

```bash
# Install protoc and protoc-gen-rust
cargo install protobuf-codegen

# Generate from proto_shared
protoc --rust_out=rust/src \
  --grpc_out=rust/src \
  --plugin=protoc-gen-grpc=`which grpc_rust_plugin` \
  proto_shared/data_service.proto
```

---

## Using the Generated Code

### Go Client Example

```go
import (
    pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// Connect to Data Service
conn, err := grpc.Dial("localhost:50070",
    grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Create client
client := pb.NewDictionaryServiceClient(conn)

// Call method
resp, err := client.GetAttribute(ctx, &pb.GetAttributeRequest{
    Id: "CLIENT_NAME",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Attribute: %s\n", resp.Name)
```

### Rust Client Example

```rust
use kyc_data::dictionary_service_client::DictionaryServiceClient;
use kyc_data::GetAttributeRequest;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Connect to Data Service
    let mut client = DictionaryServiceClient::connect(
        "http://localhost:50070"
    ).await?;
    
    // Create request
    let request = tonic::Request::new(GetAttributeRequest {
        id: "CLIENT_NAME".to_string(),
    });
    
    // Call method
    let response = client.get_attribute(request).await?;
    let attr = response.into_inner();
    
    println!("Attribute: {}", attr.name);
    
    Ok(())
}
```

---

## Adding New Shared Services

1. **Create proto file** in `proto_shared/`:
   ```protobuf
   syntax = "proto3";
   package kyc.myservice;
   
   option go_package = "github.com/adamtc007/KYC-DSL/api/pb/myservice";
   
   service MyService {
     rpc DoSomething(Request) returns (Response);
   }
   ```

2. **Update Makefile** to generate bindings:
   ```makefile
   proto-myservice:
       protoc --go_out=. --go-grpc_out=. \
         --go_opt=module=github.com/adamtc007/KYC-DSL \
         --go-grpc_opt=module=github.com/adamtc007/KYC-DSL \
         proto_shared/myservice.proto
   ```

3. **Generate for both languages**:
   ```bash
   make proto-myservice          # Go
   cd rust && cargo build        # Rust (if tonic-build configured)
   ```

4. **Implement server** (Go or Rust)

5. **Create clients** (Go, Rust, or both)

---

## Design Guidelines

### ✅ DO

- Use clear, descriptive message names
- Include pagination in list operations (limit, offset)
- Add total_count to list responses
- Use standard field types (string, int32, bool)
- Document all fields with comments
- Version services when breaking changes occur
- Use consistent naming (PascalCase for messages, camelCase for fields)

### ❌ DON'T

- Use language-specific types (e.g., `interface{}` in Go)
- Embed implementation details in messages
- Create circular dependencies between services
- Use deprecated protobuf features
- Mix concerns (keep services focused)

---

## Proto Style Guide

```protobuf
syntax = "proto3";

// Package name: lowercase, dot-separated
package kyc.service;

// Go package option: required for Go
option go_package = "github.com/adamtc007/KYC-DSL/api/pb/pkgname";

// Service: PascalCase, verb-based methods
service MyService {
  // RPC method: PascalCase
  rpc GetItem(GetItemRequest) returns (Item);
  rpc ListItems(ListItemsRequest) returns (ItemList);
}

// Message: PascalCase
message Item {
  // Fields: snake_case, numbered sequentially
  string item_id = 1;
  string item_name = 2;
  int32 item_count = 3;
}

// Request/Response: descriptive names
message GetItemRequest {
  string id = 1;
}

message ListItemsRequest {
  int32 limit = 1;
  int32 offset = 2;
}

message ItemList {
  repeated Item items = 1;
  int32 total_count = 2;
}
```

---

## Testing Proto Changes

### Verify Go Generation

```bash
make proto-data
go build ./api/pb/kycdata/...
```

### Verify Rust Generation (when configured)

```bash
cd rust
cargo build --features proto-validation
```

### Test with grpcurl

```bash
# Start server
make run-dataserver

# Test in another terminal
grpcurl -plaintext localhost:50070 list
grpcurl -plaintext localhost:50070 kyc.data.DictionaryService.GetAttribute \
  -d '{"id": "CLIENT_NAME"}'
```

---

## Versioning

When making breaking changes to protos:

1. **Create new version**:
   ```
   proto_shared/
   ├── data_service.proto      # v1 (current)
   └── data_service_v2.proto   # v2 (new)
   ```

2. **Update package name**:
   ```protobuf
   package kyc.data.v2;
   option go_package = ".../api/pb/kycdata/v2";
   ```

3. **Maintain both versions** during migration

4. **Deprecate old version** after all clients migrate

---

## Troubleshooting

### "protoc: command not found"

```bash
# macOS
brew install protobuf

# Linux
apt-get install -y protobuf-compiler

# Verify
protoc --version
```

### "Plugin not found: protoc-gen-go"

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### "Module path mismatch"

Ensure `--go_opt=module=...` matches your go.mod:
```bash
--go_opt=module=github.com/adamtc007/KYC-DSL
```

### Rust compilation errors

Check `tonic` and `prost` versions are compatible:
```toml
tonic = "0.11"
prost = "0.12"
tonic-build = "0.11"
```

---

## References

- [Protocol Buffers Guide](https://protobuf.dev/)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/)
- [tonic (Rust gRPC)](https://github.com/hyperium/tonic)
- [Data Service Guide](../DATA_SERVICE_GUIDE.md)

---

**Last Updated**: 2024  
**Maintained by**: KYC-DSL Team