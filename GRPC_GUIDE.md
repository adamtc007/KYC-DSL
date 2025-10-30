# gRPC Service Layer - Setup Guide

**Version**: 1.5  
**Status**: Ready for Development  
**Prerequisites**: Protocol Buffers compiler (protoc)

---

## 📋 Overview

The KYC-DSL project now includes a complete gRPC service layer providing type-safe, high-performance APIs for:

- **KYC Case Management** - Create, read, update, delete cases with versioning
- **DSL Operations** - Parse, validate, execute, and amend DSL cases
- **RAG Services** - Semantic search, feedback loop, and metadata operations

---

## 🚀 Quick Start

### 1. Install Protocol Buffers Compiler

**macOS**:
```bash
brew install protobuf
brew install protoc-gen-go protoc-gen-go-grpc
```

**Ubuntu/Debian**:
```bash
sudo apt install -y protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

**Verify installation**:
```bash
protoc --version  # Should show libprotoc 3.x or higher
```

### 2. Install Go Protobuf Plugins

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Add to PATH (add to ~/.zshrc or ~/.bashrc):
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### 3. Generate Proto Code

```bash
make proto
```

This generates Go code in `api/pb/` from the `.proto` definitions.

### 4. Build gRPC Server

```bash
make build-grpc
```

### 5. Run gRPC Server

```bash
# Set environment variables
export OPENAI_API_KEY="sk-..."
export PGDATABASE="kyc_dsl"

# Start server
make run-grpc
```

Server runs on:
- **gRPC**: `:50051`
- **REST Gateway** (optional): `:8080`

---

## 🔌 API Services

### 1. KYC Case Service (`kyc.KycCaseService`)

Operations for managing KYC cases:

```protobuf
service KycCaseService {
  rpc GetCase (GetCaseRequest) returns (KycCase);
  rpc UpdateCase (UpdateCaseRequest) returns (KycCase);
  rpc ListCases (ListCasesRequest) returns (stream KycCase);
  rpc CreateCase (CreateCaseRequest) returns (KycCase);
  rpc DeleteCase (DeleteCaseRequest) returns (DeleteCaseResponse);
  rpc GetCaseVersions (GetCaseVersionsRequest) returns (stream KycCaseVersion);
}
```

### 2. DSL Service (`kyc.dsl.DslService`)

DSL execution and validation:

```protobuf
service DslService {
  rpc Execute (ExecuteRequest) returns (ExecuteResponse);
  rpc Validate (ValidateRequest) returns (ValidationResult);
  rpc Parse (ParseRequest) returns (ParseResponse);
  rpc Serialize (SerializeRequest) returns (SerializeResponse);
  rpc Amend (AmendRequest) returns (AmendResponse);
  rpc ListAmendments (ListAmendmentsRequest) returns (ListAmendmentsResponse);
  rpc GetGrammar (GetGrammarRequest) returns (GrammarResponse);
}
```

### 3. RAG Service (`kyc.rag.RagService`)

Semantic search and feedback:

```protobuf
service RagService {
  rpc AttributeSearch (RagSearchRequest) returns (RagSearchResponse);
  rpc SimilarAttributes (SimilarAttributesRequest) returns (RagSearchResponse);
  rpc TextSearch (TextSearchRequest) returns (RagSearchResponse);
  rpc GetAttribute (GetAttributeRequest) returns (AttributeMetadata);
  rpc SubmitFeedback (RagFeedbackRequest) returns (RagFeedbackResponse);
  rpc GetRecentFeedback (GetRecentFeedbackRequest) returns (stream RagFeedback);
  rpc GetFeedbackAnalytics (GetFeedbackAnalyticsRequest) returns (FeedbackAnalytics);
  rpc GetMetadataStats (GetMetadataStatsRequest) returns (MetadataStats);
  rpc HealthCheck (HealthCheckRequest) returns (HealthCheckResponse);
}
```

---

## 🧪 Testing with grpcurl

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

### List Methods

```bash
grpcurl -plaintext localhost:50051 list kyc.rag.RagService
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
  "timestamp": "2024-12-20T10:00:00Z",
  "databaseStatus": "connected",
  "embedderStatus": "ready"
}
```

### Semantic Search

```bash
grpcurl -plaintext -d '{"query":"beneficial owner name","limit":3}' \
  localhost:50051 kyc.rag.RagService/AttributeSearch
```

### Submit Feedback

```bash
grpcurl -plaintext -d '{
  "query_text":"tax requirements",
  "attribute_code":"TAX_RESIDENCY_COUNTRY",
  "feedback":"positive",
  "confidence":0.9,
  "agent_type":"human"
}' localhost:50051 kyc.rag.RagService/SubmitFeedback
```

---

## 📁 Project Structure

```
KYC-DSL/
├── api/
│   ├── proto/                    # Protocol Buffer definitions
│   │   ├── kyc_case.proto       # Case management service
│   │   ├── dsl_service.proto    # DSL operations service
│   │   └── rag_service.proto    # RAG & feedback service
│   └── pb/                      # Generated Go code (gitignored)
│       ├── *.pb.go              # Protobuf messages
│       └── *_grpc.pb.go         # gRPC service stubs
├── internal/service/
│   ├── kyc_case_service.go      # Case service implementation
│   ├── dsl_service.go           # DSL service implementation
│   └── rag_service.go           # RAG service implementation
├── cmd/server/
│   └── main.go                  # gRPC server entry point
└── Makefile                     # Build automation
```

---

## 🔧 Development

### Generate Proto Code

```bash
make proto
```

### Build Server

```bash
make build-grpc
```

### Run Server

```bash
make run-grpc
```

### Clean Build

```bash
make clean
make proto
make build-grpc
```

---

## 🐛 Troubleshooting

### "protoc: command not found"

Install Protocol Buffers compiler:
```bash
brew install protobuf  # macOS
sudo apt install protobuf-compiler  # Ubuntu
```

### "protoc-gen-go: program not found"

Install Go protobuf plugins:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Add GOPATH/bin to PATH:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### "api/pb does not exist"

Generate proto code first:
```bash
make proto
```

### Import Errors

Run:
```bash
go mod tidy
go mod download
```

---

## 🎯 Next Steps

1. **Generate Proto Code**: `make proto`
2. **Build Server**: `make build-grpc`
3. **Start Server**: `make run-grpc`
4. **Test with grpcurl**: See examples above
5. **Integrate with Clients**: Use generated `api/pb` package

---

## 📚 Resources

- [gRPC Go Quick Start](https://grpc.io/docs/languages/go/quickstart/)
- [Protocol Buffers Guide](https://protobuf.dev/getting-started/gotutorial/)
- [grpcurl Documentation](https://github.com/fullstorydev/grpcurl)

---

**Last Updated**: 2024  
**Version**: 1.5  
**Status**: Ready for Development
