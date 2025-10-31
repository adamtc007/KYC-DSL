# KYC-DSL Call Tree

**Architecture:** Rust DSL Service + Go Data Layer  
**Purpose:** Complete execution path mapping for dead code identification

---

## 1. CLI Entry Point

```
main()
  └─ cmd/kycctl/main.go
      └─ cli.Run(os.Args[1:])
           └─ internal/cli/cli.go
```

---

## 2. CLI Command Dispatch

### Main Router: `cli.Run(args []string)`

```
cli.Run(args)
  ├─ "grammar"          → RunGrammarCommand()
  ├─ "ontology"         → RunOntologyCommand()
  ├─ "validate"         → RunValidateCommand(caseName, actor)
  ├─ "amend"            → RunAmendCommand(caseName, step)
  ├─ "seed-metadata"    → RunSeedMetadataCommand()
  ├─ "search-metadata"  → RunSearchMetadataCommand(query, limit)
  ├─ "similar-attributes" → RunSimilarAttributesCommand(code, limit)
  ├─ "text-search"      → RunTextSearchCommand(term)
  ├─ "metadata-stats"   → RunMetadataStatsCommand()
  ├─ "help"             → ShowUsage()
  └─ <file.dsl>         → RunProcessCommand(filePath)
```

---

## 3. Command Execution Flows

### 3.1 Grammar Command

```
RunGrammarCommand()
  ├─ rustclient.NewDslClient("")
  │   └─ grpc.DialContext() → Rust DSL Service (port 50060)
  ├─ rustClient.GetGrammar()
  │   └─ RPC: kyc.dsl.DslService/GetGrammar
  ├─ storage.ConnectPostgres()
  │   └─ sqlx.Connect("postgres", dsn)
  └─ storage.InsertGrammar(db, name, version, ebnf)
      └─ db.Exec("INSERT INTO grammar_versions...")
```

**External Calls:**
- ✅ Rust gRPC: `GetGrammar()`
- ✅ PostgreSQL: `InsertGrammar()`

---

### 3.2 Process DSL File Command

```
RunProcessCommand(filePath)
  ├─ os.ReadFile(filePath)
  ├─ rustclient.NewDslClient("")
  │   └─ grpc.DialContext() → Rust DSL Service
  ├─ rustClient.ParseDSL(dslText)
  │   └─ RPC: kyc.dsl.DslService/Parse
  │       └─ Rust: parser::parse() → AST
  ├─ rustClient.ValidateDSL(dslText)
  │   └─ RPC: kyc.dsl.DslService/Validate
  │       └─ Rust: compiler::compile() + validation
  ├─ storage.ConnectPostgres()
  ├─ storage.SaveCaseVersion(db, caseName, dslText)
  │   ├─ GetNextVersion(db, caseName)
  │   │   └─ db.Get("SELECT MAX(version)...")
  │   ├─ sha256Hex(dslText)
  │   └─ db.Exec("INSERT INTO case_versions...")
  └─ displayParsedCaseInfo(parseResp.Cases[0])
```

**External Calls:**
- ✅ Rust gRPC: `Parse()`, `Validate()`
- ✅ PostgreSQL: `SaveCaseVersion()`, `GetNextVersion()`

---

### 3.3 Validate Command

```
RunValidateCommand(caseName, actor)
  ├─ storage.ConnectPostgres()
  ├─ storage.GetLatestDSL(db, caseName)
  │   └─ db.Get("SELECT dsl_snapshot FROM case_versions...")
  ├─ rustclient.NewDslClient("")
  ├─ rustClient.ValidateDSL(dsl)
  │   └─ RPC: kyc.dsl.DslService/Validate
  └─ Display validation results
```

**External Calls:**
- ✅ Rust gRPC: `Validate()`
- ✅ PostgreSQL: `GetLatestDSL()`

---

### 3.4 Amendment Command

```
RunAmendCommand(caseName, step)
  ├─ storage.ConnectPostgres()
  ├─ IF step == "document-discovery"
  │   ├─ ontology.NewRepository(db)
  │   └─ amend.ApplyAmendment(db, caseName, step, mutation)
  │       ├─ getLatestVersion(db, caseName)
  │       ├─ rustclient.NewDslClient("")
  │       ├─ rustClient.ParseDSL(oldSnapshot)
  │       ├─ mutationFn(kycCase) ← Local mutation
  │       ├─ rustClient.SerializeCase(case)
  │       ├─ rustClient.ValidateDSL(newSnapshot)
  │       ├─ storage.SaveCaseVersion(db, caseName, newSnapshot)
  │       └─ storage.InsertAmendment(db, caseName, step, ...)
  └─ ELSE (all other amendments)
      ├─ rustclient.NewDslClient("")
      ├─ rustClient.AmendCase(caseName, step)
      │   └─ RPC: kyc.dsl.DslService/Amend
      │       └─ Rust: Apply amendment logic
      ├─ storage.SaveCaseVersion(db, caseName, amendResp.UpdatedDsl)
      └─ storage.InsertAmendment(db, caseName, step, ...)
```

**External Calls:**
- ✅ Rust gRPC: `ParseDSL()`, `AmendCase()`, `SerializeCase()`, `ValidateDSL()`
- ✅ PostgreSQL: `SaveCaseVersion()`, `InsertAmendment()`, `getLatestVersion()`
- ✅ Ontology: `NewRepository()` (only for document-discovery)

---

### 3.5 Ontology Command

```
RunOntologyCommand()
  ├─ storage.ConnectPostgres()
  ├─ ontology.NewRepository(db)
  └─ repo.DebugPrintOntologySummary()
      ├─ db.Query("SELECT * FROM kyc_regulations...")
      ├─ db.Query("SELECT * FROM kyc_documents...")
      ├─ db.Query("SELECT * FROM kyc_attributes...")
      └─ Display formatted output
```

**External Calls:**
- ✅ PostgreSQL: Multiple SELECT queries via ontology repository

---

### 3.6 RAG Commands

#### Seed Metadata
```
RunSeedMetadataCommand()
  ├─ storage.ConnectPostgres()
  ├─ rag.NewEmbedder(apiKey, model)
  ├─ ontology.NewMetadataRepository(db, embedder)
  └─ metaRepo.SeedAllMetadata()
      ├─ FOR EACH attribute:
      │   ├─ embedder.GenerateEmbedding(text)
      │   │   └─ openai.CreateEmbeddings()
      │   └─ db.Exec("INSERT INTO kyc_attribute_metadata...")
      └─ Display seeding results
```

**External Calls:**
- ✅ OpenAI API: `CreateEmbeddings()`
- ✅ PostgreSQL: `INSERT INTO kyc_attribute_metadata`

---

#### Search Metadata
```
RunSearchMetadataCommand(query, limit)
  ├─ storage.ConnectPostgres()
  ├─ rag.NewEmbedder(apiKey, model)
  ├─ ontology.NewMetadataRepository(db, embedder)
  └─ metaRepo.SearchAttributes(query, limit)
      ├─ embedder.GenerateEmbedding(query)
      │   └─ openai.CreateEmbeddings()
      └─ db.Query("SELECT ... ORDER BY embedding <-> $1 LIMIT $2")
          └─ pgvector cosine similarity search
```

**External Calls:**
- ✅ OpenAI API: `CreateEmbeddings()`
- ✅ PostgreSQL: Vector similarity query with pgvector

---

#### Similar Attributes
```
RunSimilarAttributesCommand(attributeCode, limit)
  ├─ storage.ConnectPostgres()
  ├─ rag.NewEmbedder(apiKey, model)
  ├─ ontology.NewMetadataRepository(db, embedder)
  └─ metaRepo.FindSimilarAttributes(attributeCode, limit)
      └─ db.Query("SELECT ... WHERE code != $1 ORDER BY embedding <-> (SELECT embedding FROM ... WHERE code = $1) LIMIT $2")
```

**External Calls:**
- ✅ PostgreSQL: Vector similarity query

---

#### Text Search
```
RunTextSearchCommand(searchTerm)
  ├─ storage.ConnectPostgres()
  ├─ ontology.NewMetadataRepository(db, nil)
  └─ metaRepo.TextSearchAttributes(searchTerm)
      └─ db.Query("SELECT ... WHERE code ILIKE $1 OR description ILIKE $1 OR category ILIKE $1")
```

**External Calls:**
- ✅ PostgreSQL: Text search query

---

#### Metadata Stats
```
RunMetadataStatsCommand()
  ├─ storage.ConnectPostgres()
  ├─ ontology.NewMetadataRepository(db, nil)
  └─ metaRepo.GetMetadataStats()
      └─ db.Query("SELECT COUNT(*), COUNT(embedding IS NOT NULL), AVG(...)...")
```

**External Calls:**
- ✅ PostgreSQL: Aggregate statistics query

---

## 4. Rust Client API

### All Rust Client Methods

```
internal/rustclient/dsl_client.go
  ├─ NewDslClient(addr) → *DslClient
  │   └─ grpc.DialContext() → Connects to Rust service
  │
  ├─ ParseDSL(dsl) → *pb.ParseResponse
  │   └─ RPC: kyc.dsl.DslService/Parse
  │
  ├─ ValidateDSL(dsl) → *pb.ValidationResult
  │   └─ RPC: kyc.dsl.DslService/Validate
  │
  ├─ ValidateCaseByID(caseID) → *pb.ValidationResult
  │   └─ RPC: kyc.dsl.DslService/Validate
  │
  ├─ ExecuteCase(caseID, function) → *pb.ExecuteResponse
  │   └─ RPC: kyc.dsl.DslService/Execute
  │
  ├─ AmendCase(caseName, amendType) → *pb.AmendResponse
  │   └─ RPC: kyc.dsl.DslService/Amend
  │
  ├─ SerializeCase(case) → *pb.SerializeResponse
  │   └─ RPC: kyc.dsl.DslService/Serialize
  │
  ├─ GetGrammar() → *pb.GrammarResponse
  │   └─ RPC: kyc.dsl.DslService/GetGrammar
  │
  ├─ ListAmendments() → *pb.ListAmendmentsResponse
  │   └─ RPC: kyc.dsl.DslService/ListAmendments
  │
  ├─ HealthCheck() → error
  │   └─ RPC: kyc.dsl.DslService/GetGrammar (as health check)
  │
  └─ Close() → error
      └─ conn.Close()
```

**Called By:**
- ✅ `internal/cli/cli.go` - All CLI commands
- ✅ `internal/amend/amend.go` - Amendment system

---

## 5. Storage Layer API

### Core Storage Functions

```
internal/storage/postgres.go
  ├─ ConnectPostgres() → *sqlx.DB
  │   └─ sqlx.Connect("postgres", dsn)
  │
  ├─ InsertCase(db, name) → error
  │   └─ db.Exec("INSERT INTO kyc_cases...")
  │
  ├─ InsertVersion(db, caseName, version, dsl) → error
  │   └─ db.Exec("INSERT INTO case_versions...")
  │
  ├─ GetNextVersion(db, caseName) → (int, error)
  │   └─ db.Get("SELECT MAX(version)...")
  │
  ├─ SaveCaseVersion(db, caseName, dsl) → error
  │   ├─ GetNextVersion()
  │   ├─ sha256Hex(dsl)
  │   ├─ InsertCase() if new
  │   └─ InsertVersion()
  │
  ├─ GetLatestDSL(db, caseName) → (string, error)
  │   └─ db.Get("SELECT dsl_snapshot FROM case_versions...")
  │
  ├─ InsertAmendment(db, caseName, step, changeType, diff) → error
  │   └─ db.Exec("INSERT INTO case_amendments...")
  │
  ├─ GetAmendments(db, caseName) → ([]Amendment, error)
  │   └─ db.Select("SELECT * FROM case_amendments...")
  │
  ├─ LogAmendment(db, caseName, step, diff) → error
  │   └─ InsertAmendment()
  │
  ├─ InsertGrammar(db, name, version, ebnf) → error
  │   └─ db.Exec("INSERT INTO grammar_versions...")
  │
  ├─ GetGrammar(db, name) → (string, error)
  │   └─ db.Get("SELECT ebnf FROM grammar_versions...")
  │
  ├─ RecordValidationResult(db, validation) → error
  │   └─ db.Exec("INSERT INTO case_validations...")
  │
  ├─ RecordValidationFinding(db, finding) → error
  │   └─ db.Exec("INSERT INTO validation_findings...")
  │
  ├─ GetValidationHistory(db, caseName) → ([]Validation, error)
  │   └─ db.Select("SELECT * FROM case_validations...")
  │
  ├─ RecordLineageEvaluation(db, caseName, version, result) → error
  │   └─ db.Exec("INSERT INTO lineage_evaluations...")
  │
  └─ GetLineageEvaluations(db, caseName) → ([]map, error)
      └─ db.Select("SELECT * FROM lineage_evaluations...")
```

**Called By:**
- ✅ `internal/cli/cli.go` - All CLI commands
- ✅ `internal/amend/amend.go` - Amendment system

---

## 6. Ontology Repository API

```
internal/ontology/repository.go
  ├─ NewRepository(db) → *Repository
  │
  ├─ DebugPrintOntologySummary() → error
  │   ├─ db.Query("SELECT * FROM kyc_regulations...")
  │   ├─ db.Query("SELECT * FROM kyc_documents...")
  │   └─ db.Query("SELECT * FROM kyc_attributes...")
  │
  ├─ GetAttribute(code) → (*Attribute, error)
  │   └─ db.Get("SELECT * FROM kyc_attributes WHERE code = $1")
  │
  ├─ GetDocument(code) → (*Document, error)
  │   └─ db.Get("SELECT * FROM kyc_documents WHERE code = $1")
  │
  └─ GetDocumentsForJurisdiction(jurisdiction) → ([]Document, error)
      └─ db.Select("SELECT * FROM kyc_documents JOIN kyc_doc_reg_links...")
```

**Called By:**
- ✅ `internal/cli/cli.go` - Ontology command
- ✅ `internal/amend/amend.go` - Document discovery

---

## 7. Metadata Repository API (RAG)

```
internal/ontology/metadata_repo.go
  ├─ NewMetadataRepository(db, embedder) → *MetadataRepository
  │
  ├─ SeedAllMetadata() → error
  │   ├─ db.Select("SELECT * FROM kyc_attributes...")
  │   ├─ FOR EACH attribute:
  │   │   ├─ embedder.GenerateEmbedding(text)
  │   │   └─ db.Exec("INSERT INTO kyc_attribute_metadata...")
  │   └─ Display results
  │
  ├─ SearchAttributes(query, limit) → ([]AttributeMetadata, error)
  │   ├─ embedder.GenerateEmbedding(query)
  │   └─ db.Query("SELECT ... ORDER BY embedding <-> $1 LIMIT $2")
  │
  ├─ FindSimilarAttributes(code, limit) → ([]AttributeMetadata, error)
  │   └─ db.Query("SELECT ... ORDER BY embedding <-> ...")
  │
  ├─ TextSearchAttributes(term) → ([]AttributeMetadata, error)
  │   └─ db.Query("SELECT ... WHERE code ILIKE $1 OR description ILIKE $1...")
  │
  └─ GetMetadataStats() → (*MetadataStats, error)
      └─ db.Query("SELECT COUNT(*), COUNT(embedding)...")
```

**Called By:**
- ✅ `internal/cli/cli.go` - All RAG commands

---

## 8. RAG Embedder API

```
internal/rag/embedder.go
  ├─ NewEmbedder(apiKey, model) → *Embedder
  │
  ├─ GenerateEmbedding(text) → ([]float32, error)
  │   └─ openai.CreateEmbeddings(ctx, EmbeddingRequest)
  │       └─ HTTP POST to OpenAI API
  │
  └─ GetModel() → string
```

**Called By:**
- ✅ `internal/ontology/metadata_repo.go` - All RAG operations

---

## 9. Amendment System API

```
internal/amend/amend.go
  ├─ ApplyAmendment(db, caseName, step, mutationFn) → error
  │   ├─ getLatestVersion(db, caseName)
  │   ├─ IF mutationFn != nil (ontology-aware):
  │   │   ├─ rustclient.NewDslClient()
  │   │   ├─ rustClient.ParseDSL()
  │   │   ├─ protoToModelCase()
  │   │   ├─ mutationFn(kycCase) ← Local mutation
  │   │   ├─ rustClient.SerializeCase()
  │   │   ├─ rustClient.ValidateDSL()
  │   │   ├─ storage.SaveCaseVersion()
  │   │   └─ storage.InsertAmendment()
  │   └─ ELSE (standard amendments):
  │       ├─ rustclient.NewDslClient()
  │       ├─ rustClient.AmendCase() ← Rust handles mutation
  │       ├─ storage.SaveCaseVersion()
  │       └─ storage.InsertAmendment()
  │
  ├─ getLatestVersion(db, caseName) → (*CaseVersion, error)
  │   └─ db.Get("SELECT * FROM case_versions ORDER BY version DESC LIMIT 1")
  │
  ├─ generateSimpleDiff(old, new) → string
  │   └─ Line-by-line diff comparison
  │
  ├─ detectChangeType(case, step) → string
  │   └─ Map step to change type
  │
  └─ protoToModelCase(protoCase) → *KycCase
      └─ Convert protobuf to internal model
```

**Called By:**
- ✅ `internal/cli/cli.go` - Amendment command

---

## 10. Amendment Mutations API

```
internal/amend/mutations.go
  ├─ AddPolicyDiscovery(case) → void
  ├─ AddDocumentSolicitation(case) → void
  ├─ AddDocumentDiscovery(case, repo) → error
  ├─ AddOwnershipStructure(case) → void
  ├─ AddRiskAssessment(case) → void
  ├─ AddRegulatorNotification(case) → void
  ├─ ApproveCase(case) → void
  ├─ DeclineCase(case) → void
  └─ RequestReviewCase(case) → void
```

**Called By:**
- ✅ `internal/cli/cli.go` - Document discovery amendment
- ⚠️  Others NOT called (handled by Rust service)

---

## 11. Rust DSL Service (External)

```
rust/kyc_dsl_service/src/main.rs (Port 50060)
  
  RPC: Parse
    ├─ parser::parse(dsl) → Result<Expr>
    │   └─ nom parser combinators
    ├─ extract_case_info(ast)
    └─ Return ParseResponse
  
  RPC: Validate
    ├─ compiler::compile(dsl) → Result<Plan>
    │   └─ AST validation + semantic checks
    └─ Return ValidationResult
  
  RPC: Execute
    ├─ compiler::compile(dsl)
    ├─ executor::execute(plan) → Result<String>
    └─ Return ExecuteResponse
  
  RPC: Amend
    ├─ Generate amended DSL
    ├─ Calculate hash
    └─ Return AmendResponse
  
  RPC: Serialize
    ├─ serialize_case(case)
    └─ Return SerializeResponse
  
  RPC: GetGrammar
    └─ Return EBNF grammar definition
  
  RPC: ListAmendments
    └─ Return available amendment types
```

**Called By:**
- ✅ `internal/rustclient/dsl_client.go` - All operations

---

## 12. Dead Code Analysis

### ✅ ACTIVELY USED

**CLI Commands:**
- ✅ `RunGrammarCommand`
- ✅ `RunProcessCommand`
- ✅ `RunValidateCommand`
- ✅ `RunAmendCommand`
- ✅ `RunOntologyCommand`
- ✅ `RunSeedMetadataCommand`
- ✅ `RunSearchMetadataCommand`
- ✅ `RunSimilarAttributesCommand`
- ✅ `RunTextSearchCommand`
- ✅ `RunMetadataStatsCommand`

**Rust Client:**
- ✅ `NewDslClient`
- ✅ `ParseDSL`
- ✅ `ValidateDSL`
- ✅ `AmendCase`
- ✅ `GetGrammar`
- ✅ `SerializeCase` (used by amend)
- ⚠️  `ExecuteCase` - NOT called by CLI
- ⚠️  `ValidateCaseByID` - NOT called
- ⚠️  `ListAmendments` - NOT called
- ⚠️  `HealthCheck` - NOT called

**Storage:**
- ✅ All core storage functions actively used

**Ontology:**
- ✅ `NewRepository`
- ✅ `DebugPrintOntologySummary`
- ⚠️  `GetAttribute`, `GetDocument` - Likely unused
- ✅ `NewMetadataRepository` - Used by RAG

**Amendment:**
- ✅ `ApplyAmendment`
- ✅ `AddDocumentDiscovery` (only mutation called from Go)
- ⚠️  Other mutations - ORPHANED (Rust handles these now)

### ⚠️  POTENTIAL DEAD CODE

**Amendment Mutations (internal/amend/mutations.go):**
- ⚠️  `AddPolicyDiscovery` - NOT called (Rust handles)
- ⚠️  `AddDocumentSolicitation` - NOT called (Rust handles)
- ⚠️  `AddOwnershipStructure` - NOT called (Rust handles)
- ⚠️  `AddRiskAssessment` - NOT called (Rust handles)
- ⚠️  `AddRegulatorNotification` - NOT called (Rust handles)
- ⚠️  `ApproveCase` - NOT called (Rust handles)
- ⚠️  `DeclineCase` - NOT called (Rust handles)
- ⚠️  `RequestReviewCase` - NOT called (Rust handles)

**Recommendation:** Delete `internal/amend/mutations.go` except `AddDocumentDiscovery`

---

## 13. Complete Execution Paths

### Path 1: Process DSL File
```
User types: ./kycctl sample_case.dsl

main() 
  → cli.Run(["sample_case.dsl"])
    → RunProcessCommand("sample_case.dsl")
      → os.ReadFile()
      → rustclient.NewDslClient() → gRPC connect to Rust (50060)
      → rustClient.ParseDSL() → RPC to Rust parser
      → rustClient.ValidateDSL() → RPC to Rust validator
      → storage.ConnectPostgres() → Connect to PostgreSQL
      → storage.SaveCaseVersion() → INSERT INTO case_versions
```

### Path 2: Apply Amendment
```
User types: ./kycctl amend CASE --step=policy-discovery

main()
  → cli.Run(["amend", "CASE", "--step=policy-discovery"])
    → RunAmendCommand("CASE", "policy-discovery")
      → storage.ConnectPostgres()
      → rustclient.NewDslClient()
      → rustClient.AmendCase("CASE", "policy-discovery") → RPC to Rust
      → storage.SaveCaseVersion() → INSERT INTO case_versions
      → storage.InsertAmendment() → INSERT INTO case_amendments
```

### Path 3: Semantic Search
```
User types: ./kycctl search-metadata "tax residency"

main()
  → cli.Run(["search-metadata", "tax residency"])
    → RunSearchMetadataCommand("tax residency", 10)
      → storage.ConnectPostgres()
      → rag.NewEmbedder() → OpenAI API key
      → ontology.NewMetadataRepository()
        → metaRepo.SearchAttributes("tax residency", 10)
          → embedder.GenerateEmbedding() → OpenAI API call
          → db.Query() → pgvector similarity search
```

---

## 14. External Service Dependencies

### Rust DSL Service (Port 50060)
**Called by:** Go CLI via gRPC  
**Calls:** None (stateless computation)  
**Status:** ✅ ACTIVE

### PostgreSQL (Port 5432)
**Called by:** All storage/ontology/RAG operations  
**Status:** ✅ ACTIVE

### OpenAI API
**Called by:** RAG embedder  
**Operations:** Generate embeddings  
**Status:** ✅ ACTIVE (when RAG features used)

### Go Data Service (Port 50070)
**Status:** ⚠️  OPTIONAL (not used by CLI currently)

---

## 15. Summary

**Total CLI Commands:** 10  
**Active Go Functions:** ~50  
**Active Rust RPC Calls:** 7  
**Database Tables:** 21  
**External APIs:** 2 (Rust gRPC, OpenAI)

**Clean Architecture:** ✅  
**Dead Code:** ⚠️  8 mutation functions in `internal/amend/mutations.go`

**Recommendation:** Delete unused mutation functions as Rust service now handles all standard amendments.

---

**Last Updated:** 2024-10-31  
**Architecture:** Rust DSL Service + Go Data Layer  
**Status:** Production Ready