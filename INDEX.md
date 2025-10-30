# KYC-DSL Documentation Index

**Version**: 1.5  
**Architecture**: Dual Go/Rust with Shared Protobuf

---

## 🚀 Getting Started

**New to KYC-DSL?** Start here:

1. **[README.md](README.md)** - Project overview and architecture
2. **[QUICKSTART.md](QUICKSTART.md)** - Get up and running in 5 minutes
3. **[CLAUDE.md](CLAUDE.md)** - Comprehensive guide for AI agents and developers
4. **[CALL_TREES.md](CALL_TREES.md)** - Complete call tree traces for all workflows

---

## 📦 Core Features

### DSL Parser & Execution
- **[CLAUDE.md](CLAUDE.md)** - CLI commands, data flow, and architecture
- **[README.md](README.md)** - DSL syntax and examples

### Regulatory Ontology
- **[REGULATORY_ONTOLOGY.md](REGULATORY_ONTOLOGY.md)** - Complete ontology documentation (8 regulations, 27 documents, 36 attributes)
- **[ONTOLOGY_VALIDATION.md](ONTOLOGY_VALIDATION.md)** - Validation rules and semantic checking

### Amendments & Versioning
- **[AMENDMENT_SYSTEM.md](AMENDMENT_SYSTEM.md)** - Incremental amendment workflow
- **[OWNERSHIP_CONTROL.md](OWNERSHIP_CONTROL.md)** - Ownership structures and validation

### Attribute Lineage
- **[LINEAGE_EVALUATOR.md](LINEAGE_EVALUATOR.md)** - Attribute lineage and derivation engine
- **[DERIVED_ATTRIBUTES.md](DERIVED_ATTRIBUTES.md)** - Derived attribute examples

### Validation & Audit
- **[VALIDATION_AUDIT.md](VALIDATION_AUDIT.md)** - Validation rules and audit trail system

---

## 🤖 RAG & Vector Search (v1.4+)

### Core RAG Features
- **[RAG_VECTOR_SEARCH.md](RAG_VECTOR_SEARCH.md)** - Complete RAG & semantic search documentation
- **[RAG_QUICKSTART.md](RAG_QUICKSTART.md)** - Quick start guide (5 minutes)

### Feedback Loop (v1.5)
- **[RAG_FEEDBACK.md](RAG_FEEDBACK.md)** - Self-correcting feedback loop system

---

## 🌐 API & Services

### REST API (Port 8080)
- **[API_DOCUMENTATION.md](API_DOCUMENTATION.md)** - Complete REST API reference
- Server: `cmd/kycserver/main.go`

### gRPC API (Port 50051 Go, 50060 Rust)
- **[GRPC_GUIDE.md](GRPC_GUIDE.md)** - Complete gRPC setup and usage guide
- **[GRPC_SERVICES_COMPLETE.md](GRPC_SERVICES_COMPLETE.md)** - Service definitions and examples
- Proto definitions: `api/proto/*.proto`

### Client Libraries
- **[CBU_GRAPH_GUIDE.md](CBU_GRAPH_GUIDE.md)** - CBU graph operations
- **[GIO_CLIENT_GUIDE.md](GIO_CLIENT_GUIDE.md)** - Gio UI client guide

---

## 🦀 Rust Implementation

### Getting Started
- **[RUST_QUICKSTART.md](RUST_QUICKSTART.md)** - 5-minute Rust quickstart
- **[rust/README.md](rust/README.md)** - Rust architecture overview
- **[rust/QUICK_REFERENCE.md](rust/QUICK_REFERENCE.md)** - Essential commands

### Build & Deploy
- **[RUST_BUILD_PLAYBOOK.md](RUST_BUILD_PLAYBOOK.md)** - Complete build and sanity playbook
- **[rust/DEPENDENCIES.md](rust/DEPENDENCIES.md)** - Dependencies and preflight checklist
- **[RUST_MIGRATION_REPORT.md](RUST_MIGRATION_REPORT.md)** - Architecture and migration details

### Testing
- **[RUST_SERVICE_TEST.md](RUST_SERVICE_TEST.md)** - Manual testing instructions
- Run: `cd rust && ./preflight.sh`

---

## 🧪 Testing

### Test Guides
- **[TESTING_GUIDE.md](TESTING_GUIDE.md)** - Comprehensive testing guide

### Test Scripts
- `scripts/test_feedback.sh` - Feedback loop tests
- `scripts/test_semantic_search.sh` - RAG search tests
- `scripts/test_rust_service.sh` - Rust gRPC service tests
- `scripts/init_ontology.sh` - Initialize database ontology

---

## 🎯 Quick Reference by Use Case

### "I want to get started quickly"
→ **[QUICKSTART.md](QUICKSTART.md)** or **[RUST_QUICKSTART.md](RUST_QUICKSTART.md)**

### "I want to understand the architecture"
→ **[CLAUDE.md](CLAUDE.md)** + **[RUST_MIGRATION_REPORT.md](RUST_MIGRATION_REPORT.md)**

### "I want to set up gRPC"
→ **[GRPC_GUIDE.md](GRPC_GUIDE.md)**

### "I want to use semantic search"
→ **[RAG_QUICKSTART.md](RAG_QUICKSTART.md)**

### "I want to build a client"
→ **[API_DOCUMENTATION.md](API_DOCUMENTATION.md)** + **[GRPC_GUIDE.md](GRPC_GUIDE.md)**

### "I want to build/test Rust code"
→ **[RUST_BUILD_PLAYBOOK.md](RUST_BUILD_PLAYBOOK.md)**

### "I want to understand call flows"
→ **[CALL_TREES.md](CALL_TREES.md)**

### "I want to contribute"
→ **[TESTING_GUIDE.md](TESTING_GUIDE.md)** + **[CLAUDE.md](CLAUDE.md)**

---

## 📊 Documentation Statistics

- **Total Documents**: 24 main + 3 rust
- **Core Guides**: 9
- **API Docs**: 3
- **Rust Docs**: 6
- **Feature Guides**: 9

---

## 🏗️ Architecture Documents

### System Design
- **[CLAUDE.md](CLAUDE.md)** - Complete architecture and CLI reference
- **[CALL_TREES.md](CALL_TREES.md)** - Call tree traces for all workflows
- **[RUST_MIGRATION_REPORT.md](RUST_MIGRATION_REPORT.md)** - Dual Go/Rust architecture

### Database
- Migrations: `internal/storage/migrations/`
- Ontology seed: `internal/ontology/seeds/ontology_seed.sql`

---

## 🔍 Search by Topic

### **Feedback & RAG**
- RAG_FEEDBACK.md
- RAG_VECTOR_SEARCH.md
- RAG_QUICKSTART.md

### **gRPC & APIs**
- GRPC_GUIDE.md
- GRPC_SERVICES_COMPLETE.md
- API_DOCUMENTATION.md

### **DSL & Parsing**
- CLAUDE.md
- README.md
- AMENDMENT_SYSTEM.md

### **Ontology**
- REGULATORY_ONTOLOGY.md
- ONTOLOGY_VALIDATION.md
- OWNERSHIP_CONTROL.md

### **Rust**
- RUST_BUILD_PLAYBOOK.md
- RUST_QUICKSTART.md
- RUST_MIGRATION_REPORT.md
- RUST_SERVICE_TEST.md

### **Testing**
- TESTING_GUIDE.md
- rust/preflight.sh
- scripts/test_*.sh

---

## 🎓 Learning Path

### **Beginner (Day 1)**
1. README.md
2. QUICKSTART.md
3. RAG_QUICKSTART.md

### **Intermediate (Week 1)**
1. CLAUDE.md
2. CALL_TREES.md
3. RAG_VECTOR_SEARCH.md
4. GRPC_GUIDE.md

### **Advanced (Month 1)**
1. REGULATORY_ONTOLOGY.md
2. LINEAGE_EVALUATOR.md
3. RUST_MIGRATION_REPORT.md

### **Expert (Production)**
1. RUST_BUILD_PLAYBOOK.md
2. TESTING_GUIDE.md
3. All API documentation

---

## 📞 Getting Help

1. **Check the relevant guide** from sections above
2. **Trace call flows**: See **[CALL_TREES.md](CALL_TREES.md)**
3. **Run test scripts** in `scripts/`
4. **View API documentation** for integration
5. **Check CLAUDE.md** for comprehensive CLI reference
6. **Run preflight checks**: `cd rust && ./preflight.sh`

---

**Last Updated**: 2024  
**Version**: 1.5  
**Status**: Production Ready