# KYC-DSL Call Trees

**Version**: 1.5  
**Architecture**: Dual Go/Rust with Shared Protobuf

This document provides comprehensive call tree traces for all major workflows in the KYC-DSL system.

---

## Table of Contents

1. [Go CLI Workflows](#go-cli-workflows)
2. [Parser & Validation](#parser--validation)
3. [Database Operations](#database-operations)
4. [gRPC Services](#grpc-services)
5. [RAG & Semantic Search](#rag--semantic-search)
6. [Amendment System](#amendment-system)
7. [Rust Service Workflows](#rust-service-workflows)

---

## Go CLI Workflows

### 1. Process DSL File: `./kycctl sample_case.dsl`

```
main()                                          [cmd/kycctl/main.go]
└── cli.Run(args)                              [internal/cli/cli.go]
    └── cli.RunProcessCommand("sample_case.dsl")
        ├── parser.ParseFile("sample_case.dsl")    [internal/parser/parse.go]
        │   ├── ioutil.ReadFile()                  [reads DSL file]
        │   ├── tokenize(dslText)                  [splits into tokens]
        │   ├── readExpr()                         [recursive S-expression parsing]
        │   └── returns []SExpr                    [AST list]
        │
        ├── parser.Bind(cases)                     [internal/parser/bind.go]
        │   ├── bindKycCase(sexpr)
        │   ├── bindNaturePurpose()
        │   ├── bindOwnershipStructure()
        │   ├── bindDataDictionary()
        │   └── returns []*model.KycCase           [typed models]
        │
        ├── storage.ConnectPostgres()              [internal/storage/postgres.go]
        │   ├── sql.Open("postgres", connStr)
        │   ├── db.Ping()
        │   └── createSchema()
        │
        ├── storage.GetGrammar(db)                 [internal/storage/grammar.go]
        │   └── db.QueryRow("SELECT * FROM grammar_versions ORDER BY created_at DESC LIMIT 1")
        │
        ├── parser.ValidateDSL(cases, grammar, db) [internal/parser/validate.go]
        │   ├── validateGrammar()                  [syntax checking]
        │   ├── validateSemantics()                [semantic rules]
        │   ├── validateOwnership()                [sum rules, controllers]
        │   │   ├── calculateOwnershipSum()
        │   │   ├── validateControllers()
        │   │   └── checkBeneficialOwners()
        │   └── validateOntologyReferences(db)     [check documents/attributes exist]
        │       ├── ontology.NewRepository(db)
        │       ├── repo.GetDocument(code)
        │       └── repo.GetAttribute(code)
        │
        ├── parser.SerializeCases(boundCases)      [internal/parser/serialize.go]
        │   ├── serializeKycCase()
        │   ├── serializeNaturePurpose()
        │   ├── serializeOwnership()
        │   └── returns dslText                    [reconstructed DSL]
        │
        ├── displayCaseInfo(kycCase)               [prints case details]
        │
        └── engine.NewExecutor(db).RunCase(kycCase) [internal/engine/executor.go]
            ├── executor.processCase()
            ├── executor.applyFunctions()
            ├── storage.InsertCase(db, case)       [internal/storage/case.go]
            │   ├── computeSHA256(dslText)
            │   ├── db.Exec("INSERT INTO kyc_cases ...")
            │   └── storage.InsertCaseVersion()
            │       └── db.Exec("INSERT INTO case_versions ...")
            └── returns executionResult
```

---

### 2. Grammar Command: `./kycctl grammar`

```
main()
└── cli.Run(["grammar"])
    └── cli.RunGrammarCommand()                    [internal/cli/grammar.go]
        ├── storage.ConnectPostgres()
        ├── parser.CurrentGrammarEBNF()            [internal/parser/grammar.go]
        │   └── returns grammarText                [EBNF specification]
        └── storage.InsertGrammar(db, grammar)     [internal/storage/grammar.go]
            ├── computeSHA256(grammarText)
            └── db.Exec("INSERT INTO grammar_versions ...")
```

---

### 3. Ontology Command: `./kycctl ontology`

```
main()
└── cli.Run(["ontology"])
    └── cli.RunOntologyCommand()                   [internal/cli/ontology.go]
        ├── storage.ConnectPostgres()
        ├── ontology.NewRepository(db)             [internal/ontology/repository.go]
        └── repo.DebugPrintOntologySummary()
            ├── repo.ListRegulations()
            │   └── db.Query("SELECT * FROM kyc_regulations")
            ├── repo.ListDocumentsByRegulation(regCode)
            │   └── db.Query("SELECT d.* FROM kyc_documents d JOIN kyc_doc_reg_links ...")
            └── fmt.Printf()                       [prints ontology tree]
```

---

### 4. Amendment: `./kycctl amend CASE-NAME --step=policy-discovery`

```
main()
└── cli.Run(["amend", "CASE-NAME", "--step=policy-discovery"])
    └── cli.RunAmendCommand("CASE-NAME", "policy-discovery") [internal/cli/amend.go]
        ├── storage.ConnectPostgres()
        ├── amend.GetMutationFunction("policy-discovery")     [internal/amend/mutations.go]
        │   └── returns mutationFunc                          [function pointer]
        │
        └── amend.ApplyAmendment(db, caseName, mutationFunc)  [internal/amend/amend.go]
            ├── storage.GetLatestCase(db, caseName)           [internal/storage/case.go]
            │   └── db.QueryRow("SELECT * FROM kyc_cases WHERE name = ? ORDER BY version DESC")
            │
            ├── parser.ParseFile(currentDSL)                  [re-parse existing]
            ├── parser.Bind(cases)
            │
            ├── mutationFunc(kycCase)                         [apply changes]
            │   └── amend.AddPolicyDiscovery(case)            [adds policies]
            │       ├── case.Policies = append(...)
            │       └── case.Functions = append("DISCOVER-POLICIES")
            │
            ├── parser.SerializeCases(modifiedCase)           [back to DSL]
            ├── storage.InsertCase(db, newDSL)                [new version]
            └── storage.InsertAmendment(db, amendmentRecord)  [internal/storage/amendment.go]
                └── db.Exec("INSERT INTO case_amendments ...")
```

---

### 5. Ontology-Aware Amendment: `./kycctl amend CASE --step=document-discovery`

```
main()
└── cli.Run(["amend", "CASE", "--step=document-discovery"])
    └── cli.RunAmendCommand("CASE", "document-discovery")
        ├── storage.ConnectPostgres()
        ├── ontology.NewRepository(db)                       [needed for ontology queries]
        │
        └── amend.ApplyAmendment(db, caseName, mutationFunc)
            ├── storage.GetLatestCase(db, caseName)
            ├── parser.ParseFile(currentDSL)
            ├── parser.Bind(cases)
            │
            ├── amend.AddDocumentDiscovery(case, repo)       [ontology-aware mutation]
            │   ├── repo.ListDocumentsByRegulation("AMLD5")  [query ontology]
            │   │   └── db.Query("SELECT d.* FROM kyc_documents d ...")
            │   │
            │   ├── repo.GetDocumentSources("UBO_NAME")      [get attribute sources]
            │   │   └── db.Query("SELECT doc_code, source_tier FROM kyc_attr_doc_links ...")
            │   │
            │   ├── case.DataDictionary = append(...)        [populate data dictionary]
            │   └── case.DocumentRequirements = append(...)  [populate doc requirements]
            │
            ├── parser.SerializeCases(modifiedCase)
            ├── storage.InsertCase(db, newDSL)
            └── storage.InsertAmendment(db, amendmentRecord)
```

---

## Parser & Validation

### Parser Call Tree

```
parser.ParseFile(filename)                         [internal/parser/parse.go]
├── ioutil.ReadFile(filename)
├── tokenize(dslText)
│   ├── scanner := bufio.NewScanner()
│   ├── for scanner.Scan()                         [iterate lines]
│   │   ├── handleComment(line)
│   │   ├── splitTokens(line)                      [split by parens/whitespace]
│   │   └── tokens = append(token)
│   └── returns []string
│
├── readExpr(tokens, &pos)                         [recursive descent parser]
│   ├── if token == "("                            [start of S-expression]
│   │   ├── head := tokens[pos+1]
│   │   ├── pos += 2
│   │   ├── children := []SExpr{}
│   │   ├── for tokens[pos] != ")"
│   │   │   ├── child := readExpr(tokens, &pos)   [RECURSIVE CALL]
│   │   │   └── children = append(child)
│   │   └── returns SExpr{Type: "call", Head: head, Children: children}
│   │
│   └── else                                        [atom/literal]
│       └── returns SExpr{Type: "atom", Value: token}
│
└── returns []SExpr                                 [list of parsed cases]
```

---

### Validation Call Tree

```
parser.ValidateDSL(cases, grammar, db)             [internal/parser/validate.go]
├── validateGrammar(cases, grammar)
│   ├── checkRequiredFields()
│   │   ├── hasField("nature-purpose")
│   │   ├── hasField("client-business-unit")
│   │   └── returns errors
│   └── checkFieldTypes()
│
├── validateSemantics(cases)
│   ├── validateFunctionNames()
│   ├── validatePolicyCodes()
│   └── validateObligationCodes()
│
├── validateOwnership(cases)                        [internal/parser/ownership.go]
│   ├── for each case with ownership
│   │   ├── sum := 0.0
│   │   ├── for each owner
│   │   │   └── sum += owner.Percentage
│   │   ├── if sum != 100.0
│   │   │   └── return error("ownership sum must equal 100%")
│   │   │
│   │   ├── hasController := false
│   │   ├── for each controller
│   │   │   └── hasController = true
│   │   └── if !hasController
│   │       └── return error("at least one controller required")
│   └── return nil or errors
│
└── validateOntologyReferences(cases, db)           [internal/parser/ontology_validate.go]
    ├── ontology.NewRepository(db)
    ├── for each case
    │   ├── for each document in data_dictionary
    │   │   ├── repo.GetDocument(docCode)
    │   │   │   └── db.QueryRow("SELECT * FROM kyc_documents WHERE code = ?")
    │   │   └── if not found
    │   │       └── errors = append("invalid document code")
    │   │
    │   └── for each attribute in data_dictionary
    │       ├── repo.GetAttribute(attrCode)
    │       │   └── db.QueryRow("SELECT * FROM kyc_attributes WHERE code = ?")
    │       └── if not found
    │           └── errors = append("invalid attribute code")
    └── return errors
```

---

## Database Operations

### Insert Case with Versioning

```
storage.InsertCase(db, kycCase)                    [internal/storage/case.go]
├── dslText := parser.SerializeCases([]*kycCase)
├── hash := computeSHA256(dslText)
│   ├── h := sha256.New()
│   ├── h.Write([]byte(dslText))
│   └── return hex.EncodeToString(h.Sum(nil))
│
├── tx, _ := db.Begin()                            [start transaction]
│
├── result := tx.Exec(`
│       INSERT INTO kyc_cases (name, status, last_updated)
│       VALUES (?, 'pending', NOW())
│       ON CONFLICT (name) DO UPDATE SET last_updated = NOW()
│   `, kycCase.Name)
│
├── caseID := result.LastInsertId()
│
├── tx.Exec(`
│       INSERT INTO case_versions (case_id, version, dsl_text, hash, created_at)
│       VALUES (?, 
│           (SELECT COALESCE(MAX(version), 0) + 1 FROM case_versions WHERE case_id = ?),
│           ?, ?, NOW()
│       )
│   `, caseID, caseID, dslText, hash)
│
├── tx.Commit()                                    [commit transaction]
└── return nil
```

---

### Query Ontology

```
ontology.NewRepository(db)                         [internal/ontology/repository.go]
├── repo := &Repository{db: db}
└── return repo

repo.ListDocumentsByRegulation(regCode)
├── query := `
│       SELECT d.code, d.name, d.type, d.jurisdiction
│       FROM kyc_documents d
│       JOIN kyc_doc_reg_links l ON d.code = l.doc_code
│       WHERE l.regulation_code = ?
│   `
├── rows, _ := db.Query(query, regCode)
├── defer rows.Close()
├── documents := []Document{}
├── for rows.Next()
│   ├── doc := Document{}
│   ├── rows.Scan(&doc.Code, &doc.Name, &doc.Type, &doc.Jurisdiction)
│   └── documents = append(doc)
└── return documents, nil

repo.GetDocumentSources(attrCode)
├── query := `
│       SELECT doc_code, source_tier
│       FROM kyc_attr_doc_links
│       WHERE attr_code = ?
│       ORDER BY source_tier ASC
│   `
├── rows, _ := db.Query(query, attrCode)
└── return sources, nil
```

---

## gRPC Services

### Go gRPC Server: Execute RPC

```
[gRPC Client Request] → grpc.Execute(ExecuteRequest)
│
[Server Handler: cmd/server/main.go]
└── server.Execute(ctx, req)                       [cmd/server/service.go]
    ├── caseName := req.CaseId
    ├── functionName := req.FunctionName
    │
    ├── storage.ConnectPostgres()
    ├── storage.GetLatestCase(db, caseName)
    │   └── db.QueryRow("SELECT * FROM kyc_cases ...")
    │
    ├── parser.ParseFile(caseDSL)
    ├── parser.Bind(cases)
    │
    ├── engine.NewExecutor(db)
    └── executor.ExecuteFunction(kycCase, functionName)
        ├── switch functionName
        │   case "DISCOVER-POLICIES":
        │       └── executor.discoverPolicies(case)
        │   case "SOLICIT-DOCUMENTS":
        │       └── executor.solicitDocuments(case)
        │   case "VERIFY-OWNERSHIP":
        │       └── executor.verifyOwnership(case)
        │
        ├── parser.SerializeCases(modifiedCase)
        ├── storage.InsertCase(db, updatedDSL)
        │
        └── return &pb.ExecuteResponse{
                UpdatedDsl: updatedDSL,
                Success: true,
                NewVersion: version,
            }
```

---

### Rust gRPC Server: Execute RPC

```
[gRPC Client Request] → grpc.Execute(ExecuteRequest)
│
[Server Handler: rust/kyc_dsl_service/src/main.rs]
└── RustDslServer::execute(self, request)
    ├── req := request.into_inner()
    ├── case_id := req.case_id
    ├── function_name := req.function_name
    │
    ├── dsl_source := format!(
    │       "(kyc-case {} (function {}))",
    │       case_id, function_name
    │   )
    │
    ├── kyc_dsl_core::compile_dsl(&dsl_source)    [rust/kyc_dsl_core/src/lib.rs]
    │   ├── parser::parse(src)                     [rust/kyc_dsl_core/src/parser.rs]
    │   │   ├── parse_expr(input)                  [nom parser combinators]
    │   │   │   ├── alt((parse_call, parse_atom))
    │   │   │   ├── parse_call:
    │   │   │   │   ├── tag("(")
    │   │   │   │   ├── parse_symbol()             [function name]
    │   │   │   │   ├── many0(parse_expr)          [RECURSIVE: parse children]
    │   │   │   │   └── tag(")")
    │   │   │   └── parse_atom:
    │   │   │       └── take_while1(is_symbol_char)
    │   │   └── returns Result<Expr, ParseError>
    │   │
    │   ├── compiler::compile(ast)                 [rust/kyc_dsl_core/src/compiler.rs]
    │   │   ├── match expr
    │   │   │   Expr::Call(name, args) =>
    │   │   │       ├── instructions.push(Instruction{name, args})
    │   │   │       └── for arg in args
    │   │   │           └── compile(arg)          [RECURSIVE]
    │   │   │   Expr::Atom(value) =>
    │   │   │       └── instructions.push(Instruction{name: "push", args: [value]})
    │   │   └── returns Vec<Instruction>
    │   │
    │   └── serde_json::to_string(&plan)           [serialize to JSON]
    │
    ├── kyc_dsl_core::execute_plan(&plan_json)     [rust/kyc_dsl_core/src/executor.rs]
    │   ├── plan := serde_json::from_str(plan_json)
    │   ├── context := ExecutionContext::new()
    │   ├── for instruction in plan
    │   │   ├── match instruction.name
    │   │   │   "init-case" =>
    │   │   │       └── context.set("case_id", args[0])
    │   │   │   "nature" =>
    │   │   │       └── context.set("nature", args[0])
    │   │   │   "function" =>
    │   │   │       └── context.execute_function(args[0])
    │   │   └── ...
    │   └── returns Result<String, DslError>
    │
    └── return Response::new(ExecuteResponse{
            updated_dsl: dsl_source,
            message: format!("Executed {} on {}", function_name, case_id),
            success: true,
            new_version: 1,
        })
```

---

## RAG & Semantic Search

### Seed Metadata with Embeddings

```
main()
└── cli.Run(["seed-metadata"])
    └── cli.RunSeedMetadataCommand()               [internal/cli/rag.go]
        ├── storage.ConnectPostgres()
        ├── ontology.NewMetadataRepo(db)           [internal/ontology/metadata.go]
        ├── rag.NewEmbedder(apiKey)                [internal/rag/embedder.go]
        │   └── returns &Embedder{client: openai.NewClient(apiKey)}
        │
        └── for each sample attribute
            ├── metadata := AttributeMetadata{
            │       Code: "UBO_NAME",
            │       Synonyms: ["beneficial owner name", ...],
            │       Context: "Legal name of beneficial owner...",
            │       Citations: ["AMLD5 Article 3", ...],
            │       RiskLevel: "HIGH",
            │   }
            │
            ├── embedder.GenerateEmbedding(metadata)  [internal/rag/embedder.go]
            │   ├── text := formatMetadataForEmbedding(metadata)
            │   │   └── returns "Code: UBO_NAME\nSynonyms: ...\nContext: ..."
            │   │
            │   ├── req := openai.EmbeddingRequest{
            │   │       Input: text,
            │   │       Model: "text-embedding-3-large",
            │   │   }
            │   │
            │   ├── resp := client.CreateEmbeddings(ctx, req)  [OpenAI API call]
            │   └── returns resp.Data[0].Embedding             [[]float32, 1536 dims]
            │
            └── repo.UpsertMetadata(metadata, embedding)       [internal/ontology/metadata.go]
                └── db.Exec(`
                        INSERT INTO kyc_attribute_metadata 
                        (attribute_code, synonyms, business_context, regulatory_citations, 
                         risk_level, embedding, last_updated)
                        VALUES (?, ?, ?, ?, ?, ?, NOW())
                        ON CONFLICT (attribute_code) DO UPDATE SET
                            synonyms = EXCLUDED.synonyms,
                            embedding = EXCLUDED.embedding,
                            last_updated = NOW()
                    `, code, synonyms, context, citations, risk, embedding)
```

---

### Semantic Search

```
main()
└── cli.Run(["search-metadata", "tax compliance requirements"])
    └── cli.RunSearchMetadataCommand("tax compliance requirements", limit)
        ├── storage.ConnectPostgres()
        ├── ontology.NewMetadataRepo(db)
        ├── rag.NewEmbedder(apiKey)
        │
        ├── embedder.GenerateEmbeddingFromText(query)  [internal/rag/embedder.go]
        │   ├── req := openai.EmbeddingRequest{
        │   │       Input: "tax compliance requirements",
        │   │       Model: "text-embedding-3-large",
        │   │   }
        │   ├── resp := client.CreateEmbeddings(ctx, req)
        │   └── returns queryVector                    [[]float32, 1536 dims]
        │
        └── repo.SearchByVector(queryVector, limit)    [internal/ontology/metadata.go]
            ├── query := `
            │       SELECT 
            │           attribute_code,
            │           synonyms,
            │           business_context,
            │           regulatory_citations,
            │           risk_level,
            │           1 - (embedding <=> $1) AS similarity
            │       FROM kyc_attribute_metadata
            │       WHERE embedding IS NOT NULL
            │       ORDER BY embedding <=> $1
            │       LIMIT $2
            │   `
            │   [<=> is pgvector cosine distance operator]
            │
            ├── rows := db.Query(query, queryVector, limit)
            ├── results := []SearchResult{}
            ├── for rows.Next()
            │   ├── result := SearchResult{}
            │   ├── rows.Scan(&result.Code, &result.Similarity, ...)
            │   └── results = append(result)
            │
            └── return results                         [sorted by similarity]
```

---

### Submit Feedback

```
[HTTP Client] → POST /rag/feedback
│
[API Handler: internal/api/rag.go]
└── handler.SubmitFeedback(w, r)
    ├── json.Decode(r.Body, &request)
    ├── feedback := FeedbackRequest{
    │       QueryText: "beneficial owner name",
    │       AttributeCode: "UBO_NAME",
    │       Feedback: "positive",
    │       Confidence: 0.9,
    │       AgentType: "human",
    │   }
    │
    ├── ontology.NewFeedbackRepo(db)               [internal/ontology/feedback.go]
    └── repo.InsertFeedback(feedback)
        ├── tx := db.Begin()
        │
        ├── tx.Exec(`
        │       INSERT INTO rag_feedback 
        │       (query_text, attribute_code, feedback, confidence, agent_type)
        │       VALUES (?, ?, ?, ?, ?)
        │   `, feedback.QueryText, feedback.AttributeCode, 
        │      feedback.Feedback, feedback.Confidence, feedback.AgentType)
        │
        │   [Trigger fires automatically: trig_feedback_relevance]
        │   [PostgreSQL Trigger Logic:]
        │   └── FUNCTION update_relevance()
        │       ├── IF NEW.feedback = 'positive' THEN
        │       │   └── UPDATE kyc_attribute_metadata
        │       │       SET relevance_score = relevance_score + (0.05 * NEW.confidence)
        │       │       WHERE attribute_code = NEW.attribute_code
        │       │
        │       └── ELSIF NEW.feedback = 'negative' THEN
        │           └── UPDATE kyc_attribute_metadata
        │               SET relevance_score = relevance_score - (0.05 * NEW.confidence)
        │               WHERE attribute_code = NEW.attribute_code
        │
        ├── tx.Commit()
        └── return nil
```

---

## Amendment System

### Full Amendment Workflow

```
main()
└── cli.Run(["amend", "AVIVA-EU-EQUITY-FUND", "--step=policy-discovery"])
    └── cli.RunAmendCommand("AVIVA-EU-EQUITY-FUND", "policy-discovery")
        │
        ├── amend.GetMutationFunction("policy-discovery")
        │   ├── mutations := map[string]MutationFunc{
        │   │       "policy-discovery": AddPolicyDiscovery,
        │   │       "document-solicitation": AddDocumentSolicitation,
        │   │       "document-discovery": AddDocumentDiscovery,
        │   │       "ownership-discovery": AddOwnershipDiscovery,
        │   │       "risk-assessment": AddRiskAssessment,
        │   │       "approve": Approve,
        │   │       "decline": Decline,
        │   │   }
        │   └── return mutations["policy-discovery"]
        │
        └── amend.ApplyAmendment(db, caseName, mutationFunc)
            │
            ├── storage.GetLatestCase(db, "AVIVA-EU-EQUITY-FUND")
            │   ├── query := `
            │   │       SELECT cv.dsl_text, cv.version
            │   │       FROM case_versions cv
            │   │       JOIN kyc_cases kc ON cv.case_id = kc.id
            │   │       WHERE kc.name = ?
            │   │       ORDER BY cv.version DESC
            │   │       LIMIT 1
            │   │   `
            │   └── returns dslText, version
            │
            ├── parser.ParseFile(dslText)
            │   └── returns []SExpr
            │
            ├── parser.Bind(cases)
            │   └── returns []*model.KycCase
            │
            ├── kycCase := cases[0]
            │
            ├── mutationFunc(kycCase)                  [AddPolicyDiscovery]
            │   └── amend.AddPolicyDiscovery(case)     [internal/amend/mutations.go]
            │       ├── case.Functions = append("DISCOVER-POLICIES")
            │       ├── case.Policies = append(
            │       │       "KYCPOL-UK-2025",
            │       │       "KYCPOL-EU-2025",
            │       │   )
            │       └── case.KycToken = "policy-discovery-complete"
            │
            ├── parser.SerializeCases([]*kycCase)      [convert back to DSL]
            │   ├── dsl := "(kyc-case AVIVA-EU-EQUITY-FUND\n"
            │   ├── dsl += "  (nature-purpose ...)\n"
            │   ├── dsl += "  (function DISCOVER-POLICIES)\n"
            │   ├── dsl += "  (policy KYCPOL-UK-2025)\n"
            │   ├── dsl += "  (policy KYCPOL-EU-2025)\n"
            │   ├── dsl += "  (kyc-token \"policy-discovery-complete\")\n"
            │   └── dsl += ")"
            │
            ├── storage.InsertCase(db, updatedDSL)     [new version: N+1]
            │   ├── computeSHA256(updatedDSL)
            │   ├── tx.Exec("INSERT INTO case_versions ...")
            │   └── tx.Commit()
            │
            └── storage.InsertAmendment(db, amendmentRecord)
                └── db.Exec(`
                        INSERT INTO case_amendments 
                        (case_id, from_version, to_version, amendment_type, applied_at)
                        VALUES (?, ?, ?, 'policy-discovery', NOW())
                    `, caseID, oldVersion, newVersion)
```

---

## Rust Service Workflows

### Rust Service Startup

```
main()                                             [rust/kyc_dsl_service/src/main.rs]
├── tokio::main
└── async fn main()
    ├── addr := "[::1]:50060".parse()
    ├── service := RustDslServer::default()
    │
    ├── reflection_service := ReflectionBuilder::configure()
    │   ├── register_encoded_file_descriptor_set(...)
    │   └── build_v1()
    │
    └── Server::builder()
        ├── add_service(DslServiceServer::new(service))
        ├── add_service(reflection_service)
        └──