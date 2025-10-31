# KYC-DSL Data Service Guide

**Version**: 1.0  
**Port**: 50070  
**Protocol**: gRPC  
**Language**: Go with pgx/v5  

## Overview

The **Data Service** is a centralized gRPC microservice that owns all PostgreSQL connections and exposes a typed API for dictionary (ontology) data and case version management. It serves as the single source of truth for database operations, consumed by:

- **Go CLI** (`kycctl`) - case processing and ontology queries
- **Rust DSL Engine** (`kyc_dsl_service`) - ontology validation and data access
- **UI/Frontend** - case data and dictionary browsing
- **REST API** - indirect access through API gateway

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Data Service (Port 50070)                 │
│  ┌────────────────────┐      ┌─────────────────────────┐   │
│  │ DictionaryService  │      │    CaseService          │   │
│  │  - GetAttribute    │      │  - SaveCaseVersion      │   │
│  │  - ListAttributes  │      │  - GetCaseVersion       │   │
│  │  - GetDocument     │      │  - ListCaseVersions     │   │
│  │  - ListDocuments   │      └─────────────────────────┘   │
│  └────────────────────┘                                     │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │      pgx/v5 Connection Pool (5-20 connections)       │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                  ┌───────────────────────┐
                  │   PostgreSQL (5432)    │
                  │  ┌──────────────────┐  │
                  │  │ kyc_attributes   │  │
                  │  │ kyc_documents    │  │
                  │  │ case_versions    │  │
                  │  └──────────────────┘  │
                  └───────────────────────┘

            ┌──────────────────┐  ┌──────────────────┐
            │   Go CLI/UI      │  │  Rust DSL Engine │
            │  (gRPC Client)   │  │  (gRPC Client)   │
            └──────────────────┘  └──────────────────┘
```

## Key Features

### 🎯 Single Responsibility
- **One service, one database pool** - eliminates connection sprawl
- Centralized connection management with configurable pooling
- Health checks and automatic reconnection

### 🔍 Dictionary Service
- **Attributes**: Regulatory attributes (CLIENT_NAME, UBO_PERCENT, etc.)
- **Documents**: KYC document types (PASSPORT, W-9, etc.)
- Pagination support (limit/offset)
- Jurisdiction filtering for documents

### 📦 Case Service
- **Version control**: Full case history with SHA-256 hashing
- **State management**: draft, validated, approved, rejected
- **Audit trail**: Timestamps and lineage tracking

### 🚀 Performance
- pgx/v5 native driver (faster than database/sql)
- Connection pooling (5 min, 20 max connections)
- Prepared statement caching
- Efficient batch operations

## Quick Start

### 1. Prerequisites

```bash
# PostgreSQL running
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:16

# Or use existing installation
export PGHOST=localhost
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=postgres
export PGDATABASE=kyc_dsl
```

### 2. Initialize Database

```bash
# Create database (if needed)
createdb kyc_dsl

# Run initialization script (creates tables + sample data)
make init-dataserver

# Or manually:
./scripts/init_data_service.sh
```

**Tables Created**:
- `kyc_attributes` - 36 regulatory attributes
- `kyc_documents` - 27 document types
- `case_versions` - case version history

### 3. Build and Run

```bash
# Build the Data Service
make build-dataserver

# Run the service (port 50070)
make run-dataserver

# Or run directly:
./bin/dataserver
```

**Expected Output**:
```
🚀 Starting KYC Data Service...

📊 Initializing database connection pool...
✅ Connected to PostgreSQL
📊 Connection pool: max=20, min=5
✅ Data Service initialized successfully

📋 Available services:
   • kyc.data.DictionaryService - Ontology data (attributes, documents)
   • kyc.data.CaseService - Case version management

🌐 gRPC server listening on :50070
```

## API Reference

### DictionaryService

#### GetAttribute
Retrieve a single attribute by code.

```bash
grpcurl -plaintext localhost:50070 \
  kyc.data.DictionaryService/GetAttribute \
  -d '{"id": "CLIENT_NAME"}'
```

**Response**:
```json
{
  "id": "CLIENT_NAME",
  "name": "Client Name",
  "description": "Legal name of the client entity",
  "attrType": "text",
  "jurisdiction": "GLOBAL",
  "regulation": "GENERAL"
}
```

#### ListAttributes
Paginated list of all attributes.

```bash
grpcurl -plaintext localhost:50070 \
  kyc.data.DictionaryService/ListAttributes \
  -d '{"limit": 10, "offset": 0}'
```

**Response**:
```json
{
  "attributes": [
    {
      "id": "CLIENT_LEI",
      "name": "Legal Entity Identifier",
      "description": "LEI code for the client",
      "attrType": "text",
      "jurisdiction": "GLOBAL",
      "regulation": "FATCA"
    }
  ],
  "totalCount": 36
}
```

#### GetDocument
Retrieve a single document by code.

```bash
grpcurl -plaintext localhost:50070 \
  kyc.data.DictionaryService/GetDocument \
  -d '{"id": "DOC_PASSPORT"}'
```

**Response**:
```json
{
  "id": "DOC_PASSPORT",
  "title": "Passport",
  "jurisdiction": "GLOBAL",
  "category": "Identity",
  "description": "Government-issued passport",
  "url": ""
}
```

#### ListDocuments
Paginated list of documents with optional jurisdiction filter.

```bash
# All documents
grpcurl -plaintext localhost:50070 \
  kyc.data.DictionaryService/ListDocuments \
  -d '{"limit": 10, "offset": 0}'

# US documents only
grpcurl -plaintext localhost:50070 \
  kyc.data.DictionaryService/ListDocuments \
  -d '{"limit": 10, "offset": 0, "jurisdiction": "US"}'
```

### CaseService

#### SaveCaseVersion
Create a new case version.

```bash
grpcurl -plaintext localhost:50070 \
  kyc.data.CaseService/SaveCaseVersion \
  -d '{
    "case_id": "CASE-001",
    "dsl_source": "(kyc-case ACME-CORP ...)",
    "compiled_json": "{\"case_id\": \"ACME-CORP\"}",
    "status": "draft"
  }'
```

**Response**:
```json
{
  "success": true,
  "error": "",
  "versionId": "123"
}
```

#### GetCaseVersion
Retrieve the latest version of a case.

```bash
grpcurl -plaintext localhost:50070 \
  kyc.data.CaseService/GetCaseVersion \
  -d '{"case_id": "CASE-001"}'
```

**Response**:
```json
{
  "id": "123",
  "caseId": "CASE-001",
  "dslSource": "(kyc-case ACME-CORP ...)",
  "compiledJson": "{\"case_id\": \"ACME-CORP\"}",
  "status": "draft",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

#### ListCaseVersions
Get all versions of a case.

```bash
grpcurl -plaintext localhost:50070 \
  kyc.data.CaseService/ListCaseVersions \
  -d '{"case_id": "CASE-001", "limit": 10, "offset": 0}'
```

## Go Client Integration

### Example: Query Attributes

```go
package main

import (
    "context"
    "log"
    
    pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // Connect to Data Service
    conn, err := grpc.Dial("localhost:50070", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()
    
    // Create Dictionary client
    client := pb.NewDictionaryServiceClient(conn)
    
    // List attributes
    resp, err := client.ListAttributes(context.Background(), 
        &pb.ListAttributesRequest{
            Limit:  10,
            Offset: 0,
        })
    if err != nil {
        log.Fatalf("ListAttributes failed: %v", err)
    }
    
    log.Printf("Found %d attributes (total: %d)", 
        len(resp.Attributes), resp.TotalCount)
    
    for _, attr := range resp.Attributes {
        log.Printf("  %s: %s (%s)", attr.Id, attr.Name, attr.Regulation)
    }
}
```

### Example: Save Case Version

```go
func saveCaseVersion(client pb.CaseServiceClient, caseID, dslSource string) error {
    resp, err := client.SaveCaseVersion(context.Background(),
        &pb.CaseVersionRequest{
            CaseId:       caseID,
            DslSource:    dslSource,
            CompiledJson: compileToJSON(dslSource),
            Status:       "draft",
        })
    
    if err != nil {
        return fmt.Errorf("save failed: %w", err)
    }
    
    if !resp.Success {
        return fmt.Errorf("save failed: %s", resp.Error)
    }
    
    log.Printf("✅ Saved version %s", resp.VersionId)
    return nil
}
```

## Rust Client Integration

### Add Dependencies

```toml
[dependencies]
tonic = "0.11"
prost = "0.12"
tokio = { version = "1", features = ["full"] }
```

### Example: Query Documents

```rust
use kyc_data::dictionary_service_client::DictionaryServiceClient;
use kyc_data::{ListDocumentsRequest};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Connect to Data Service
    let mut client = DictionaryServiceClient::connect("http://localhost:50070").await?;
    
    // List documents
    let request = tonic::Request::new(ListDocumentsRequest {
        limit: 20,
        offset: 0,
        jurisdiction: "US".to_string(),
    });
    
    let response = client.list_documents(request).await?;
    let docs = response.into_inner();
    
    println!("Found {} US documents (total: {})", 
        docs.documents.len(), docs.total_count);
    
    for doc in docs.documents {
        println!("  {}: {} ({})", doc.id, doc.title, doc.category);
    }
    
    Ok(())
}
```

## Environment Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | - | Full PostgreSQL connection string |
| `PGHOST` | `localhost` | PostgreSQL host |
| `PGPORT` | `5432` | PostgreSQL port |
| `PGUSER` | `postgres` | PostgreSQL user |
| `PGPASSWORD` | `postgres` | PostgreSQL password |
| `PGDATABASE` | `kyc_dsl` | Database name |
| `PGSSLMODE` | `disable` | SSL mode (disable/require/verify-ca/verify-full) |

### Connection Pool Settings

Configured in `internal/service/db.go`:

```go
cfg.MaxConns = 20                      // Maximum connections
cfg.MinConns = 5                       // Minimum connections
cfg.MaxConnLifetime = time.Hour        // Max connection age
cfg.MaxConnIdleTime = 30 * time.Minute // Max idle time
cfg.HealthCheckPeriod = time.Minute    // Health check interval
```

## Database Schema

### kyc_attributes

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key |
| `attribute_code` | VARCHAR(100) | Unique attribute code (e.g., CLIENT_NAME) |
| `attribute_name` | VARCHAR(255) | Display name |
| `description` | TEXT | Full description |
| `attribute_type` | VARCHAR(50) | Data type (text, number, date, etc.) |
| `jurisdiction` | VARCHAR(10) | Jurisdiction (US, EU, GLOBAL) |
| `regulation_code` | VARCHAR(50) | Regulation (FATCA, CRS, AMLD5) |
| `created_at` | TIMESTAMP | Creation timestamp |
| `updated_at` | TIMESTAMP | Last update timestamp |

### kyc_documents

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key |
| `document_code` | VARCHAR(100) | Unique document code (e.g., DOC_PASSPORT) |
| `document_name` | VARCHAR(255) | Display name |
| `jurisdiction` | VARCHAR(10) | Jurisdiction |
| `category` | VARCHAR(100) | Category (Identity, Tax, Address) |
| `description` | TEXT | Full description |
| `reference_url` | TEXT | Reference URL (optional) |
| `created_at` | TIMESTAMP | Creation timestamp |
| `updated_at` | TIMESTAMP | Last update timestamp |

### case_versions

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key |
| `case_id` | VARCHAR(255) | Case identifier |
| `dsl_source` | TEXT | Original DSL source code |
| `compiled_json` | TEXT | Compiled JSON representation |
| `status` | VARCHAR(50) | Status (draft, validated, approved, rejected) |
| `created_at` | TIMESTAMP | Creation timestamp |
| `updated_at` | TIMESTAMP | Last update timestamp |

**Indexes**:
- `idx_case_versions_case_id` - Fast lookup by case ID
- `idx_case_versions_case_id_created` - Sorted version history
- `idx_case_versions_status` - Filter by status

## Testing

### Unit Tests

```bash
# Test database connection
go test ./internal/service -v -run TestInitDB

# Test gRPC service methods
go test ./internal/service -v -run TestDataService
```

### Integration Testing with grpcurl

```bash
# Health check (list services)
grpcurl -plaintext localhost:50070 list

# Test Dictionary Service
grpcurl -plaintext localhost:50070 \
  kyc.data.DictionaryService/ListAttributes \
  -d '{"limit": 5}'

# Test Case Service
grpcurl -plaintext localhost:50070 \
  kyc.data.CaseService/SaveCaseVersion \
  -d '{
    "case_id": "TEST-001",
    "dsl_source": "(kyc-case TEST ...)",
    "compiled_json": "{}",
    "status": "draft"
  }'

grpcurl -plaintext localhost:50070 \
  kyc.data.CaseService/GetCaseVersion \
  -d '{"case_id": "TEST-001"}'
```

### Performance Testing

```bash
# Install ghz (gRPC benchmarking tool)
go install github.com/bojand/ghz/cmd/ghz@latest

# Benchmark ListAttributes
ghz --insecure \
  --proto proto_shared/data_service.proto \
  --call kyc.data.DictionaryService.ListAttributes \
  -d '{"limit": 10}' \
  -n 1000 -c 10 \
  localhost:50070

# Expected: ~500-1000 req/sec with 10 concurrent connections
```

## Troubleshooting

### Connection Errors

**Error**: `Failed to connect to database`

```bash
# Check PostgreSQL is running
psql -h localhost -U postgres -d kyc_dsl -c '\l'

# Verify environment variables
echo $PGHOST $PGPORT $PGUSER $PGDATABASE

# Check connection string
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/kyc_dsl?sslmode=disable"
```

### Port Already in Use

**Error**: `Failed to listen on :50070`

```bash
# Find process using port 50070
lsof -i :50070

# Kill the process
kill -9 <PID>

# Or use a different port
# Edit cmd/dataserver/main.go and change ":50070" to ":50071"
```

### Missing Tables

**Error**: `relation "kyc_attributes" does not exist`

```bash
# Run initialization script
make init-dataserver

# Or manually
psql -h localhost -U postgres -d kyc_dsl -f scripts/init_data_service_tables.sql
```

### gRPC Reflection Not Working

```bash
# Ensure reflection is registered (already done in main.go)
# Test with grpcurl
grpcurl -plaintext localhost:50070 list

# If empty, check server logs for errors
./bin/dataserver
```

## Port Allocation Reference

| Service | Port | Description |
|---------|------|-------------|
| **Data Service** | **50070** | Dictionary + Case services |
| Main gRPC Service | 50051 | KycCase, DSL, RAG, CBU |
| Rust DSL Service | 50060 | Rust compiler service |
| REST API | 8080 | HTTP/JSON gateway |
| PostgreSQL | 5432 | Database |

## Monitoring and Observability

### Health Check Endpoint

```go
// Add to cmd/dataserver/main.go for HTTP health check
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    if err := service.HealthCheck(ctx); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("unhealthy"))
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("healthy"))
})
go http.ListenAndServe(":8071", nil)
```

### Metrics

Connection pool stats are available via `pgxpool.Stat()`:

```go
stats := service.DB.Stat()
log.Printf("Pool stats: total=%d idle=%d active=%d",
    stats.TotalConns(), stats.IdleConns(), stats.AcquiredConns())
```

## Best Practices

### Connection Management
- ✅ Use the global `DB` pool from `service.InitDB()`
- ✅ Always use context for timeouts
- ✅ Close connections gracefully on shutdown
- ❌ Don't create new connection pools

### Error Handling
- ✅ Return gRPC status codes (NotFound, InvalidArgument)
- ✅ Log errors with context (case ID, attribute code)
- ✅ Return user-friendly error messages
- ❌ Don't expose internal database errors

### Performance
- ✅ Use pagination (limit/offset) for large datasets
- ✅ Add indexes for frequently queried columns
- ✅ Use prepared statements (pgx does this automatically)
- ❌ Don't return unbounded result sets

## Migration from Old Architecture

If migrating from the existing `internal/storage/postgres.go`:

1. **Update imports**:
   ```go
   // Old
   import "github.com/adamtc007/KYC-DSL/internal/storage"
   
   // New
   import pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"
   ```

2. **Replace direct SQL queries**:
   ```go
   // Old
   storage.QueryAttributes(db, filter)
   
   // New
   client := pb.NewDictionaryServiceClient(conn)
   client.ListAttributes(ctx, &pb.ListAttributesRequest{...})
   ```

3. **Update connection initialization**:
   ```go
   // Old (in main.go)
   db, err := storage.ConnectPostgres()
   
   // New (Data Service handles this)
   conn, err := grpc.Dial("localhost:50070", ...)
   ```

## Future Enhancements

- [ ] Add caching layer (Redis) for frequently accessed attributes
- [ ] Implement full-text search on documents and attributes
- [ ] Add batch operations (bulk insert/update)
- [ ] Implement gRPC streaming for large result sets
- [ ] Add authentication/authorization (JWT tokens)
- [ ] Implement rate limiting per client
- [ ] Add distributed tracing (OpenTelemetry)
- [ ] Create GraphQL gateway layer

## References

- **Protocol Buffers**: `proto_shared/data_service.proto`
- **Go Implementation**: `internal/service/data_service.go`
- **Database Schema**: `scripts/init_data_service_tables.sql`
- **Main Server**: `cmd/dataserver/main.go`
- **Makefile Targets**: `make help` (search for "dataserver")

## Support

For issues or questions:
1. Check server logs: `./bin/dataserver` output
2. Verify database connectivity: `psql -h localhost -U postgres -d kyc_dsl`
3. Test with grpcurl: `grpcurl -plaintext localhost:50070 list`
4. Review this guide's troubleshooting section

---

**Last Updated**: 2024  
**Maintainer**: KYC-DSL Team  
**License**: MIT