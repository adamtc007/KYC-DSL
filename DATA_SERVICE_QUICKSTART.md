# KYC-DSL Data Service - Quick Reference

**Port**: 50070 | **Protocol**: gRPC | **Database**: PostgreSQL | **Driver**: pgx/v5

---

## 🚀 Quick Start (60 seconds)

```bash
# 1. Initialize database schema
make init-dataserver

# 2. Build the service
make build-dataserver

# 3. Run the service
make run-dataserver

# 4. Test it (in another terminal)
grpcurl -plaintext localhost:50070 list
```

---

## 📁 Project Structure

```
KYC-DSL/
├── proto_shared/
│   └── data_service.proto          ← gRPC contract (shared with Rust)
├── api/pb/kycdata/
│   ├── data_service.pb.go          ← Generated protobuf code
│   └── data_service_grpc.pb.go     ← Generated gRPC server/client
├── internal/dataservice/
│   ├── db.go                       ← pgx connection pool manager
│   └── data_service.go             ← gRPC service implementation
├── cmd/dataserver/
│   └── main.go                     ← Server entry point
└── scripts/
    ├── init_data_service.sh        ← Database initialization
    ├── init_data_service_tables.sql
    └── test_data_service.sh        ← Integration tests
```

---

## 🎯 Makefile Targets

```bash
make proto-data         # Regenerate protobuf bindings
make build-dataserver   # Build the server binary
make run-dataserver     # Run the server (port 50070)
make init-dataserver    # Initialize database schema
```

---

## 🗄️ Database Setup

### Environment Variables
```bash
export PGHOST=localhost
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=postgres
export PGDATABASE=kyc_dsl
```

### Manual Initialization
```bash
# Create database
createdb kyc_dsl

# Run schema + seed data
psql -d kyc_dsl -f scripts/init_data_service_tables.sql

# Verify
psql -d kyc_dsl -c "SELECT COUNT(*) FROM kyc_attributes;"
psql -d kyc_dsl -c "SELECT COUNT(*) FROM kyc_documents;"
```

---

## 📡 gRPC Services

### DictionaryService (Ontology Data)

| Method | Description | Example |
|--------|-------------|---------|
| `GetAttribute` | Get single attribute by code | `CLIENT_NAME`, `UBO_PERCENT` |
| `ListAttributes` | Paginated attribute list | limit/offset |
| `GetDocument` | Get single document by code | `DOC_PASSPORT`, `DOC_W9` |
| `ListDocuments` | Paginated document list | filter by jurisdiction |

### CaseService (Version Control)

| Method | Description | Example |
|--------|-------------|---------|
| `SaveCaseVersion` | Create new case version | DSL source + compiled JSON |
| `GetCaseVersion` | Get latest case version | Returns most recent |
| `ListCaseVersions` | Get case history | All versions, paginated |

---

## 🧪 Test Commands (grpcurl)

```bash
# List all services
grpcurl -plaintext localhost:50070 list

# Get an attribute
grpcurl -plaintext localhost:50070 \
  kyc.data.DictionaryService/GetAttribute \
  -d '{"id": "CLIENT_NAME"}'

# List attributes
grpcurl -plaintext localhost:50070 \
  kyc.data.DictionaryService/ListAttributes \
  -d '{"limit": 10, "offset": 0}'

# Get a document
grpcurl -plaintext localhost:50070 \
  kyc.data.DictionaryService/GetDocument \
  -d '{"id": "DOC_PASSPORT"}'

# List US documents
grpcurl -plaintext localhost:50070 \
  kyc.data.DictionaryService/ListDocuments \
  -d '{"limit": 10, "jurisdiction": "US"}'

# Save a case
grpcurl -plaintext localhost:50070 \
  kyc.data.CaseService/SaveCaseVersion \
  -d '{
    "case_id": "ACME-001",
    "dsl_source": "(kyc-case ACME-CORP ...)",
    "compiled_json": "{}",
    "status": "draft"
  }'

# Get latest case version
grpcurl -plaintext localhost:50070 \
  kyc.data.CaseService/GetCaseVersion \
  -d '{"case_id": "ACME-001"}'
```

---

## 🔧 Integration Tests

```bash
# Run full test suite
./scripts/test_data_service.sh

# Expected output: 20+ tests passing
# Tests: Dictionary queries, Case CRUD, error handling
```

---

## 🦀 Rust Client Example

```rust
use kyc_data::dictionary_service_client::DictionaryServiceClient;
use kyc_data::ListAttributesRequest;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut client = DictionaryServiceClient::connect(
        "http://localhost:50070"
    ).await?;
    
    let req = tonic::Request::new(ListAttributesRequest {
        limit: 10,
        offset: 0,
    });
    
    let resp = client.list_attributes(req).await?;
    println!("Attributes: {:?}", resp.into_inner());
    Ok(())
}
```

---

## 🔗 Go Client Example

```go
import (
    pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

conn, _ := grpc.Dial("localhost:50070",
    grpc.WithTransportCredentials(insecure.NewCredentials()))
defer conn.Close()

client := pb.NewDictionaryServiceClient(conn)
resp, _ := client.ListAttributes(ctx, &pb.ListAttributesRequest{
    Limit: 10,
})
```

---

## 🐛 Troubleshooting

| Problem | Solution |
|---------|----------|
| **Port 50070 in use** | `lsof -i :50070` then `kill -9 <PID>` |
| **Database connection error** | Check `PGHOST`, `PGPORT`, PostgreSQL running |
| **Missing tables** | Run `make init-dataserver` |
| **grpcurl not found** | `brew install grpcurl` (macOS) |
| **Regenerate protos** | `make proto-data` |

---

## 🌐 Port Allocation

| Port | Service |
|------|---------|
| **50070** | **Data Service (Dictionary + Case)** ← YOU ARE HERE |
| 50051 | Main gRPC Service (KycCase, DSL, RAG, CBU) |
| 50060 | Rust DSL Engine |
| 8080 | REST API Gateway |
| 5432 | PostgreSQL |

---

## 📚 Database Schema

### kyc_attributes (36 rows)
- `attribute_code` - Unique ID (CLIENT_NAME, UBO_PERCENT, etc.)
- `attribute_name` - Display name
- `description` - Full description
- `attribute_type` - text, number, date, etc.
- `jurisdiction` - US, EU, GLOBAL
- `regulation_code` - FATCA, CRS, AMLD5, etc.

### kyc_documents (27 rows)
- `document_code` - Unique ID (DOC_PASSPORT, DOC_W9, etc.)
- `document_name` - Display name
- `jurisdiction` - US, EU, GLOBAL
- `category` - Identity, Tax, Address, Entity
- `reference_url` - Official source (if available)

### case_versions
- `id` - Auto-increment version ID
- `case_id` - Case identifier (multiple versions per case)
- `dsl_source` - Original S-expression DSL
- `compiled_json` - Compiled representation
- `status` - draft, validated, approved, rejected
- `created_at` - Timestamp

---

## 🎓 Key Features

✅ **Single connection pool** - No connection sprawl  
✅ **Type-safe gRPC** - Shared protos with Rust  
✅ **Pagination** - All list operations support limit/offset  
✅ **Filtering** - Documents by jurisdiction  
✅ **Version control** - Full case history  
✅ **Health checks** - Connection pool monitoring  
✅ **pgx/v5** - High-performance native driver  

---

## 📖 Full Documentation

- **Comprehensive Guide**: `DATA_SERVICE_GUIDE.md`
- **Proto Definition**: `proto_shared/data_service.proto`
- **Architecture**: `CLAUDE.md` (search "Data Service")

---

## 🚦 Status Indicators

```bash
# Service is ready when you see:
✅ Connected to PostgreSQL
📊 Connection pool: max=20, min=5
🌐 gRPC server listening on :50070

# Database is ready when:
✅ kyc_attributes table (5+ rows)
✅ kyc_documents table (5+ rows)
✅ case_versions table (created)
```

---

**Last Updated**: 2024  
**Version**: 1.0  
**Maintainer**: KYC-DSL Team