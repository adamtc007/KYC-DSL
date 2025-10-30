# gRPC Setup - Step by Step

Follow these steps to get the gRPC service layer running.

---

## ⚠️ Prerequisites

Before running `make proto`, you need:

### 1. Protocol Buffers Compiler

**macOS**:
```bash
brew install protobuf
```

**Ubuntu/Debian**:
```bash
sudo apt update
sudo apt install -y protobuf-compiler
```

**Verify**:
```bash
protoc --version
# Should show: libprotoc 3.x or higher
```

### 2. Go Protobuf Plugins

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 3. Add to PATH

Add this to your `~/.zshrc` or `~/.bashrc`:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Then reload:
```bash
source ~/.zshrc  # or ~/.bashrc
```

**Verify**:
```bash
which protoc-gen-go
which protoc-gen-go-grpc
# Both should show paths in $GOPATH/bin
```

---

## 🚀 Setup Steps

### Step 1: Generate Proto Code

```bash
make proto
```

Expected output:
```
Generating protobuf Go code...
✓ Proto files generated in api/pb
```

This creates:
- `api/pb/kyc_case.pb.go`
- `api/pb/kyc_case_grpc.pb.go`
- `api/pb/dsl_service.pb.go`
- `api/pb/dsl_service_grpc.pb.go`
- `api/pb/rag_service.pb.go`
- `api/pb/rag_service_grpc.pb.go`

### Step 2: Build gRPC Server

```bash
make build-grpc
```

Expected output:
```
Building grpcserver with GOEXPERIMENT=greenteagc...
```

Creates: `bin/grpcserver`

### Step 3: Setup Environment

```bash
export OPENAI_API_KEY="sk-..."
export PGDATABASE="kyc_dsl"
export PGHOST="localhost"
export PGPORT="5432"
```

### Step 4: Start Server

```bash
make run-grpc
```

Expected output:
```
🚀 Starting gRPC Server...
📊 Connecting to PostgreSQL...
✅ Database connected successfully
🌐 gRPC server listening on :50051

📋 Available services:
   • kyc.KycCaseService
   • kyc.dsl.DslService
   • kyc.rag.RagService

💡 Test with grpcurl:
   grpcurl -plaintext localhost:50051 list
```

---

## 🧪 Testing

### Install grpcurl

```bash
brew install grpcurl  # macOS
# or
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### List Services

```bash
grpcurl -plaintext localhost:50051 list
```

Output:
```
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
kyc.KycCaseService
kyc.dsl.DslService
kyc.rag.RagService
```

### Health Check

```bash
grpcurl -plaintext localhost:50051 kyc.rag.RagService/HealthCheck
```

Output:
```json
{
  "status": "healthy",
  "model": "text-embedding-3-large",
  "dimensions": 1536,
  "timestamp": "2024-10-30T19:00:00Z",
  "databaseStatus": "connected",
  "embedderStatus": "ready"
}
```

---

## 🐛 Troubleshooting

### Error: "protoc: command not found"

**Solution**:
```bash
brew install protobuf  # macOS
sudo apt install protobuf-compiler  # Ubuntu
```

### Error: "protoc-gen-go: program not found"

**Solution**:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Add to PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Error: "cannot find package"

**Solution**:
```bash
go mod download
go mod tidy
```

### Error: "api/pb does not exist"

**Solution**:
```bash
mkdir -p api/pb
make proto
```

---

## ✅ Verification

After setup, verify everything works:

```bash
# 1. Proto files exist
ls api/pb/*.pb.go

# 2. Server binary exists
ls bin/grpcserver

# 3. Server starts
make run-grpc &

# 4. Services respond
grpcurl -plaintext localhost:50051 list

# 5. Health check works
grpcurl -plaintext localhost:50051 kyc.rag.RagService/HealthCheck
```

---

## 📚 Next Steps

1. Read [GRPC_GUIDE.md](GRPC_GUIDE.md) for detailed usage
2. Explore [GRPC_IMPLEMENTATION_SUMMARY.md](GRPC_IMPLEMENTATION_SUMMARY.md)
3. Try the example queries in the guides
4. Integrate with your client applications

---

**Status**: Ready to use!  
**Support**: See [GRPC_GUIDE.md](GRPC_GUIDE.md) for troubleshooting
