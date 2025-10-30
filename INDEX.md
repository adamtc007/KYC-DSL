# KYC-DSL v1.5 - Documentation Index

Complete guide to all documentation, organized by feature and use case.

---

## 🚀 Getting Started

**New to KYC-DSL?** Start here:

1. [README.md](README.md) - Project overview and quick start
2. [QUICKSTART.md](QUICKSTART.md) - Get up and running in 5 minutes
3. [CLAUDE.md](CLAUDE.md) - AI-friendly project guide

---

## 📦 Core Features

### DSL & Parsing
- [README.md](README.md) - DSL syntax and examples
- [CLAUDE.md](CLAUDE.md) - Parser architecture and usage

### Regulatory Ontology
- [REGULATORY_ONTOLOGY.md](REGULATORY_ONTOLOGY.md) - Complete ontology documentation
- [ONTOLOGY_VALIDATION.md](ONTOLOGY_VALIDATION.md) - Validation rules

### Amendments & Versioning
- [AMENDMENT_SYSTEM.md](AMENDMENT_SYSTEM.md) - Amendment workflow
- [OWNERSHIP_CONTROL.md](OWNERSHIP_CONTROL.md) - Ownership structures

### Attribute Lineage
- [LINEAGE_EVALUATOR.md](LINEAGE_EVALUATOR.md) - Lineage & derivation
- [DERIVED_ATTRIBUTES.md](DERIVED_ATTRIBUTES.md) - Derived attributes

### Validation & Audit
- [VALIDATION_AUDIT.md](VALIDATION_AUDIT.md) - Validation & audit trail

---

## 🤖 RAG & Vector Search (v1.4+)

### Core RAG Features
- [RAG_VECTOR_SEARCH.md](RAG_VECTOR_SEARCH.md) - Complete RAG documentation
- [RAG_QUICKSTART.md](RAG_QUICKSTART.md) - Quick start guide
- [RAG_IMPLEMENTATION_SUMMARY.md](RAG_IMPLEMENTATION_SUMMARY.md) - Technical details

### Feedback Loop (v1.5)
- [RAG_FEEDBACK.md](RAG_FEEDBACK.md) - **Complete feedback guide (400+ lines)**
- [RAG_FEEDBACK_QUICKREF.md](RAG_FEEDBACK_QUICKREF.md) - **Quick reference card**
- [RAG_FEEDBACK_IMPLEMENTATION.md](RAG_FEEDBACK_IMPLEMENTATION.md) - **Implementation details**

---

## 🌐 gRPC Service Layer (v1.5)

### Setup & Configuration
- [SETUP_GRPC.md](SETUP_GRPC.md) - **Step-by-step setup guide**
- [GRPC_GUIDE.md](GRPC_GUIDE.md) - **Complete usage guide**

### Technical Documentation
- [GRPC_IMPLEMENTATION_SUMMARY.md](GRPC_IMPLEMENTATION_SUMMARY.md) - **Implementation details**
- [api/proto/*.proto](api/proto/) - Protocol Buffer definitions

---

## 🧪 Testing

### Test Guides
- [TESTING_GUIDE.md](TESTING_GUIDE.md) - Complete testing guide
- [scripts/test_feedback.sh](scripts/test_feedback.sh) - Feedback loop tests
- [scripts/example_feedback_workflow.sh](scripts/example_feedback_workflow.sh) - Interactive demo
- [scripts/test_semantic_search.sh](scripts/test_semantic_search.sh) - RAG search tests

---

## 🔧 API Documentation

### REST API (Port 8080)
- [API_DOCUMENTATION.md](API_DOCUMENTATION.md) - Complete REST API reference
- Server: `cmd/kycserver/main.go`

### gRPC API (Port 50051)
- [GRPC_GUIDE.md](GRPC_GUIDE.md) - gRPC API guide
- Proto definitions: `api/proto/*.proto`
- Server: `cmd/server/main.go`

---

## 🏗️ Architecture & Implementation

### System Design
- [CLAUDE.md](CLAUDE.md) - Architecture overview
- [COMPLETE_IMPLEMENTATION_SUMMARY.md](COMPLETE_IMPLEMENTATION_SUMMARY.md) - **v1.5 summary**

### Database
- [internal/storage/migrations/](internal/storage/migrations/) - All migrations
- `007_rag_feedback.sql` - Feedback loop schema

### Scripts
- [scripts/migrate_feedback.sh](scripts/migrate_feedback.sh) - Apply feedback migration
- [scripts/init_ontology.sh](scripts/init_ontology.sh) - Initialize ontology

---

## 📝 Release Notes

- [CHANGES_v1.5.md](CHANGES_v1.5.md) - v1.5 release notes
- [COMPLETE_IMPLEMENTATION_SUMMARY.md](COMPLETE_IMPLEMENTATION_SUMMARY.md) - Complete v1.5 summary

---

## 🎯 Quick Reference by Use Case

### "I want to add feedback to search results"
→ [RAG_FEEDBACK_QUICKREF.md](RAG_FEEDBACK_QUICKREF.md)

### "I want to set up gRPC"
→ [SETUP_GRPC.md](SETUP_GRPC.md)

### "I want to understand RAG search"
→ [RAG_QUICKSTART.md](RAG_QUICKSTART.md)

### "I want to build a client"
→ [GRPC_GUIDE.md](GRPC_GUIDE.md) + [API_DOCUMENTATION.md](API_DOCUMENTATION.md)

### "I want to contribute"
→ [TESTING_GUIDE.md](TESTING_GUIDE.md) + [CLAUDE.md](CLAUDE.md)

### "I want to understand the architecture"
→ [CLAUDE.md](CLAUDE.md) + [COMPLETE_IMPLEMENTATION_SUMMARY.md](COMPLETE_IMPLEMENTATION_SUMMARY.md)

---

## 📊 Documentation Statistics

- **Total Documents**: 25+
- **Total Lines**: ~10,000+
- **Guides**: 15
- **API Docs**: 3
- **Test Scripts**: 5
- **Proto Files**: 3

---

## 🔍 Search by Topic

### Feedback
- RAG_FEEDBACK.md
- RAG_FEEDBACK_QUICKREF.md
- RAG_FEEDBACK_IMPLEMENTATION.md
- scripts/test_feedback.sh

### gRPC
- SETUP_GRPC.md
- GRPC_GUIDE.md
- GRPC_IMPLEMENTATION_SUMMARY.md
- api/proto/*.proto

### Vector Search
- RAG_VECTOR_SEARCH.md
- RAG_QUICKSTART.md
- RAG_IMPLEMENTATION_SUMMARY.md

### DSL
- README.md
- CLAUDE.md
- AMENDMENT_SYSTEM.md

### Ontology
- REGULATORY_ONTOLOGY.md
- ONTOLOGY_VALIDATION.md
- OWNERSHIP_CONTROL.md

---

## 🎓 Learning Path

### Beginner
1. README.md
2. QUICKSTART.md
3. RAG_QUICKSTART.md

### Intermediate
1. CLAUDE.md
2. RAG_VECTOR_SEARCH.md
3. API_DOCUMENTATION.md

### Advanced
1. REGULATORY_ONTOLOGY.md
2. LINEAGE_EVALUATOR.md
3. GRPC_IMPLEMENTATION_SUMMARY.md

### Expert
1. COMPLETE_IMPLEMENTATION_SUMMARY.md
2. All migration files
3. Internal implementation files

---

## 📞 Getting Help

1. Check relevant guide above
2. Run test scripts
3. View API documentation
4. Check implementation summaries

---

**Last Updated**: 2024  
**Version**: 1.5  
**Status**: Complete & Production Ready
