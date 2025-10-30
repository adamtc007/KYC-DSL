# RAG & Vector Search Testing Guide

**Version**: 1.4  
**Last Updated**: 2024  
**Status**: Production Ready

---

## 📋 Quick Test Checklist

Run through these steps to verify your RAG system is working:

- [ ] Database connection
- [ ] pgvector extension enabled
- [ ] OpenAI API key configured
- [ ] Metadata seeded with embeddings
- [ ] CLI commands working
- [ ] API server running
- [ ] Endpoints responding correctly

---

## 🚀 Step-by-Step Testing

### Step 1: Prerequisites Check

```bash
# Check PostgreSQL is running
psql -d kyc_dsl -c "SELECT version();"

# Check pgvector extension
psql -d kyc_dsl -c "SELECT * FROM pg_extension WHERE extname='vector';"

# Should return a row with extname='vector'
# If not, install: CREATE EXTENSION vector;

# Check OpenAI API key
echo $OPENAI_API_KEY
# Should output: sk-...

# Check binaries are built
ls -la bin/
# Should show: kycctl and kycserver
```

**Expected Output**:
```
✅ PostgreSQL version: 12+
✅ pgvector extension: enabled
✅ OPENAI_API_KEY: configured
✅ Binaries: present
```

---

### Step 2: Seed Metadata

```bash
# Generate embeddings for sample attributes
./bin/kycctl seed-metadata
```

**Expected Output**:
```
🌱 Seeding Attribute Metadata with Embeddings...
================================================

📊 Processing 20 attributes...

[1/20] Processing: REGISTERED_NAME
  ✅ Seeded with 1536-dimensional embedding
[2/20] Processing: TAX_RESIDENCY_COUNTRY
  ✅ Seeded with 1536-dimensional embedding
[3/20] Processing: UBO_NAME
  ✅ Seeded with 1536-dimensional embedding
...

================================================
📈 Seeding Summary
================================================
✅ Successfully seeded: 20 attributes
❌ Failed: 0 attributes
⏱️  Total time: 15.3s
🚀 Average time per attribute: 765ms

================================================
📊 Repository Statistics
================================================
Total attributes with metadata: 20
Attributes with embeddings: 20
Embedding coverage: 100.0%

✅ Seeding complete!
```

**If it fails**:
- Check `OPENAI_API_KEY` is valid
- Verify internet connection
- Check OpenAI API rate limits
- Review error messages for details

---

### Step 3: Test CLI Commands

#### Test 3.1: Metadata Statistics

```bash
./bin/kycctl metadata-stats
```

**Expected Output**:
```
📊 Attribute Metadata Statistics
================================================

📈 Overview:
  Total Attributes:         20
  With Embeddings:          20
  Embedding Coverage:       100.0%

⚠️  Risk Level Distribution:
  CRITICAL      7 attributes
  HIGH          6 attributes
  MEDIUM        5 attributes
  LOW           2 attributes

✅ All attributes have embeddings!
```

---

#### Test 3.2: Semantic Search - Tax Compliance

```bash
./bin/kycctl search-metadata "tax reporting requirements" --limit=3
```

**Expected Output**:
```
🔍 Semantic Search: "tax reporting requirements"
================================================

⚡ Generating query embedding...
🔎 Searching for top 3 matches...

📊 Found 3 matches:

─────────────────────────────────────────────────
Rank #1
─────────────────────────────────────────────────
🏷️  Code:           TAX_RESIDENCY_COUNTRY
📈 Similarity:      0.8742 (distance: 0.1258)
⚠️  Risk Level:      HIGH
📝 Data Type:       enum(ISO 3166-1 Alpha-2)
🔤 Synonyms:        Tax Country, Country of Tax Residence, Tax Jurisdiction, Fiscal Residence
📖 Context:         Jurisdiction where the entity is considered tax resident under FATCA/CRS regulations...
📜 Citations:       FATCA §1471(b)(1)(D), CRS Common Reporting Standard, OECD Model Tax Convention
💡 Examples:        US, GB, HK, SG, DE

─────────────────────────────────────────────────
Rank #2
─────────────────────────────────────────────────
🏷️  Code:           FATCA_STATUS
📈 Similarity:      0.8521 (distance: 0.1479)
⚠️  Risk Level:      HIGH
📝 Data Type:       enum
🔤 Synonyms:        FATCA Classification, Chapter 4 Status, US Tax Status
📖 Context:         Entity classification under the Foreign Account Tax Compliance Act...
📜 Citations:       FATCA §1471-1474, IRS Publication 5190

─────────────────────────────────────────────────
Rank #3
─────────────────────────────────────────────────
🏷️  Code:           CRS_CLASSIFICATION
📈 Similarity:      0.8312 (distance: 0.1688)
⚠️  Risk Level:      HIGH
📝 Data Type:       enum
🔤 Synonyms:        CRS Status, AEOI Classification, Common Reporting Standard Type
```

**✅ Success Criteria**:
- TAX_RESIDENCY_COUNTRY appears in top 3
- Similarity scores > 0.8
- All results have tax-related context

---

#### Test 3.3: Semantic Search - Beneficial Ownership

```bash
./bin/kycctl search-metadata "who owns this company" --limit=3
```

**Expected Output**:
```
🔍 Semantic Search: "who owns this company"
================================================

📊 Found 3 matches:

─────────────────────────────────────────────────
Rank #1
─────────────────────────────────────────────────
🏷️  Code:           UBO_NAME
📈 Similarity:      0.8956 (distance: 0.1044)
⚠️  Risk Level:      CRITICAL
📝 Data Type:       string
🔤 Synonyms:        Ultimate Beneficial Owner Name, Beneficial Owner, UBO, Controller Name
📖 Context:         Full legal name of the ultimate beneficial owner who directly or indirectly owns or controls more than 25%...
📜 Citations:       AMLD5 Article 3, FATF Recommendation 24, MAS 626

─────────────────────────────────────────────────
Rank #2
─────────────────────────────────────────────────
🏷️  Code:           UBO_OWNERSHIP_PERCENT
📈 Similarity:      0.8734 (distance: 0.1266)
⚠️  Risk Level:      CRITICAL

─────────────────────────────────────────────────
Rank #3
─────────────────────────────────────────────────
🏷️  Code:           DIRECTOR_NAME
📈 Similarity:      0.8123 (distance: 0.1877)
⚠️  Risk Level:      HIGH
```

**✅ Success Criteria**:
- UBO_NAME is #1 result
- Ownership-related attributes in top 3
- Similarity scores > 0.8

---

#### Test 3.4: Find Similar Attributes

```bash
./bin/kycctl similar-attributes UBO_NAME --limit=3
```

**Expected Output**:
```
🔍 Finding Similar Attributes to: UBO_NAME
================================================

📋 Source Attribute:
  Code:        UBO_NAME
  Risk Level:  CRITICAL
  Data Type:   string
  Synonyms:    Ultimate Beneficial Owner Name, Beneficial Owner, UBO, Controller Name
  Context:     Full legal name of the ultimate beneficial owner...

🔎 Finding top 3 similar attributes...

📊 Found 3 similar attributes:

─────────────────────────────────────────────────
Rank #1
─────────────────────────────────────────────────
🏷️  Code:           UBO_OWNERSHIP_PERCENT
📈 Similarity:      0.9123 (distance: 0.0877)
⚠️  Risk Level:      CRITICAL
📝 Data Type:       float
🔤 Synonyms:        Ownership Percentage, Beneficial Ownership %, Control Percentage
📖 Context:         Percentage of ownership or voting rights held by the ultimate beneficial owner...

─────────────────────────────────────────────────
Rank #2
─────────────────────────────────────────────────
🏷️  Code:           DIRECTOR_NAME
📈 Similarity:      0.8345 (distance: 0.1655)

─────────────────────────────────────────────────
Rank #3
─────────────────────────────────────────────────
🏷️  Code:           REGISTERED_NAME
📈 Similarity:      0.7621 (distance: 0.2379)

================================================
💡 Clustering Suggestion:
   These attributes could form a cluster with UBO_NAME
   based on semantic similarity.
```

**✅ Success Criteria**:
- UBO_OWNERSHIP_PERCENT is most similar
- Ownership/control attributes ranked highly
- Top result has similarity > 0.9

---

#### Test 3.5: Text Search (Synonym Resolution)

```bash
./bin/kycctl text-search "Company Name"
```

**Expected Output**:
```
🔍 Text Search: "Company Name"
================================================

🔎 Searching attributes and synonyms...

📊 Found 1 matches:

─────────────────────────────────────────────────
Result #1
─────────────────────────────────────────────────
🏷️  Code:           REGISTERED_NAME
⚠️  Risk Level:      LOW
📝 Data Type:       string
🔤 Synonyms:        Legal Name, Company Name, Entity Name, Corporate Name
📖 Context:         Official legal name of the entity as registered with competent authority...
💡 Examples:        BlackRock Global Funds, HSBC Holdings PLC, Deutsche Bank AG
```

**✅ Success Criteria**:
- REGISTERED_NAME found (matches "Company Name" synonym)
- Returns without calling OpenAI API
- Fast response (<50ms)

---

### Step 4: Start API Server

```bash
# Terminal 1: Start server
make run-server

# Or directly:
./bin/kycserver
```

**Expected Output**:
```
🚀 Starting KYC-DSL RAG API Server...
📊 Connecting to PostgreSQL...
✅ Database connected successfully
🧠 Initializing OpenAI embedder...
   Model: text-embedding-3-large
   Dimensions: 1536
🌐 Server listening on http://localhost:8080

📋 Available endpoints:
   GET  /                                   - API documentation
   GET  /rag/health                         - Health check
   GET  /rag/stats                          - Metadata statistics
   GET  /rag/attribute_search?q=<query>     - Semantic search
   GET  /rag/similar_attributes?code=<code> - Similar attributes
   GET  /rag/text_search?term=<term>        - Text search
   GET  /rag/attribute/<code>               - Get attribute metadata
```

**Keep this terminal open!** The server is now running.

---

### Step 5: Test REST API Endpoints

Open a new terminal for testing.

#### Test 5.1: Health Check

```bash
curl http://localhost:8080/rag/health
```

**Expected Output**:
```json
{
  "status": "healthy",
  "embeddings_count": 20,
  "embedding_model": "text-embedding-3-large",
  "embedding_dimensions": 1536
}
```

**✅ Success Criteria**:
- Status: "healthy"
- Embeddings count: 20
- Model: "text-embedding-3-large"

---

#### Test 5.2: Metadata Statistics

```bash
curl http://localhost:8080/rag/stats
```

**Expected Output**:
```json
{
  "total_attributes": 20,
  "attributes_with_embeddings": 20,
  "embedding_coverage_percent": 100.0,
  "risk_distribution": [
    {"risk_level": "CRITICAL", "count": 7},
    {"risk_level": "HIGH", "count": 6},
    {"risk_level": "MEDIUM", "count": 5},
    {"risk_level": "LOW", "count": 2}
  ]
}
```

**✅ Success Criteria**:
- Coverage: 100%
- All risk levels represented

---

#### Test 5.3: Semantic Search - Beneficial Owner

```bash
curl "http://localhost:8080/rag/attribute_search?q=beneficial%20owner%20name&limit=3"
```

**Expected Output**:
```json
{
  "query": "beneficial owner name",
  "limit": 3,
  "count": 3,
  "results": [
    {
      "code": "UBO_NAME",
      "risk_level": "CRITICAL",
      "data_type": "string",
      "business_context": "Full legal name of the ultimate beneficial owner who directly or indirectly owns or controls more than 25% of the entity or exercises control through other means.",
      "synonyms": ["Ultimate Beneficial Owner Name", "Beneficial Owner", "UBO", "Controller Name"],
      "regulatory_citations": ["AMLD5 Article 3", "FATF Recommendation 24", "MAS 626"],
      "example_values": ["John Smith", "Jane Doe", "Michael Chen"],
      "similarity_score": 0.9234,
      "distance": 0.0766
    },
    {
      "code": "UBO_OWNERSHIP_PERCENT",
      "risk_level": "CRITICAL",
      "data_type": "float",
      "business_context": "Percentage of ownership or voting rights held by the ultimate beneficial owner. Threshold of 25% triggers reporting requirements under most AML regulations.",
      "synonyms": ["Ownership Percentage", "Beneficial Ownership %", "Control Percentage"],
      "regulatory_citations": ["AMLD5 Article 3(6)", "FATF Recommendation 10"],
      "example_values": ["26.5", "50.0", "100.0"],
      "similarity_score": 0.8812,
      "distance": 0.1188
    },
    {
      "code": "REGISTERED_NAME",
      "risk_level": "LOW",
      "data_type": "string",
      "business_context": "Official legal name of the entity as registered with the competent authority. This is the primary identifier for legal entity identification across all jurisdictions.",
      "synonyms": ["Legal Name", "Company Name", "Entity Name", "Corporate Name"],
      "regulatory_citations": ["AMLD5 Article 13", "MAS 626 Annex A", "Companies Act 2006 s.86"],
      "example_values": ["BlackRock Global Funds", "HSBC Holdings PLC", "Deutsche Bank AG"],
      "similarity_score": 0.7456,
      "distance": 0.2544
    }
  ]
}
```

**✅ Success Criteria**:
- UBO_NAME is #1 result
- Similarity score > 0.9 for top result
- All fields populated correctly
- Response time < 500ms

---

#### Test 5.4: Semantic Search - Tax Reporting

```bash
curl "http://localhost:8080/rag/attribute_search?q=tax%20reporting%20requirements&limit=3"
```

**Expected Output**:
```json
{
  "query": "tax reporting requirements",
  "limit": 3,
  "count": 3,
  "results": [
    {
      "code": "TAX_RESIDENCY_COUNTRY",
      "risk_level": "HIGH",
      "data_type": "enum(ISO 3166-1 Alpha-2)",
      "business_context": "Jurisdiction where the entity is considered tax resident under FATCA/CRS regulations. Critical for automatic exchange of information and tax reporting obligations.",
      "synonyms": ["Tax Country", "Country of Tax Residence", "Tax Jurisdiction", "Fiscal Residence"],
      "regulatory_citations": ["FATCA §1471(b)(1)(D)", "CRS Common Reporting Standard", "OECD Model Tax Convention"],
      "example_values": ["US", "GB", "HK", "SG", "DE"],
      "similarity_score": 0.8742,
      "distance": 0.1258
    },
    {
      "code": "FATCA_STATUS",
      "risk_level": "HIGH",
      "similarity_score": 0.8521,
      "distance": 0.1479
    },
    {
      "code": "CRS_CLASSIFICATION",
      "risk_level": "HIGH",
      "similarity_score": 0.8312,
      "distance": 0.1688
    }
  ]
}
```

**✅ Success Criteria**:
- TAX_RESIDENCY_COUNTRY is top result
- FATCA_STATUS and CRS_CLASSIFICATION in top 3
- All tax-related attributes

---

#### Test 5.5: Semantic Search - Risk Indicators

```bash
curl "http://localhost:8080/rag/attribute_search?q=money%20laundering%20risk%20factors&limit=5"
```

**Expected Output**:
```json
{
  "query": "money laundering risk factors",
  "limit": 5,
  "count": 5,
  "results": [
    {
      "code": "PEP_STATUS",
      "risk_level": "CRITICAL",
      "similarity_score": 0.9012,
      "business_context": "Indicator of whether the individual holds or has held a prominent public function. PEPs present higher money laundering and corruption risks requiring enhanced due diligence."
    },
    {
      "code": "SANCTIONS_SCREENING_STATUS",
      "risk_level": "CRITICAL",
      "similarity_score": 0.8845
    },
    {
      "code": "SOURCE_OF_FUNDS",
      "risk_level": "CRITICAL",
      "similarity_score": 0.8623
    },
    {
      "code": "ADVERSE_MEDIA_FLAG",
      "risk_level": "HIGH",
      "similarity_score": 0.8401
    },
    {
      "code": "CUSTOMER_RISK_RATING",
      "risk_level": "CRITICAL",
      "similarity_score": 0.8234
    }
  ]
}
```

**✅ Success Criteria**:
- All results are HIGH or CRITICAL risk
- PEP_STATUS, SANCTIONS, SOURCE_OF_FUNDS in top 3
- Similarity scores > 0.8

---

#### Test 5.6: Similar Attributes

```bash
curl "http://localhost:8080/rag/similar_attributes?code=PEP_STATUS&limit=3"
```

**Expected Output**:
```json
{
  "source_attribute": "PEP_STATUS",
  "limit": 3,
  "count": 3,
  "results": [
    {
      "code": "SANCTIONS_SCREENING_STATUS",
      "risk_level": "CRITICAL",
      "similarity_score": 0.8923,
      "distance": 0.1077
    },
    {
      "code": "ADVERSE_MEDIA_FLAG",
      "risk_level": "HIGH",
      "similarity_score": 0.8456,
      "distance": 0.1544
    },
    {
      "code": "SOURCE_OF_FUNDS",
      "risk_level": "CRITICAL",
      "similarity_score": 0.8012,
      "distance": 0.1988
    }
  ]
}
```

**✅ Success Criteria**:
- All related to risk/compliance screening
- Top result similarity > 0.85
- Response time < 50ms (no OpenAI call)

---

#### Test 5.7: Text Search

```bash
curl "http://localhost:8080/rag/text_search?term=ownership"
```

**Expected Output**:
```json
{
  "search_term": "ownership",
  "count": 2,
  "results": [
    {
      "code": "UBO_OWNERSHIP_PERCENT",
      "risk_level": "CRITICAL",
      "data_type": "float",
      "business_context": "Percentage of ownership or voting rights held by the ultimate beneficial owner...",
      "synonyms": ["Ownership Percentage", "Beneficial Ownership %", "Control Percentage"]
    },
    {
      "code": "UBO_NAME",
      "risk_level": "CRITICAL"
    }
  ]
}
```

**✅ Success Criteria**:
- Returns attributes with "ownership" in code/context
- Fast response (<20ms)
- No OpenAI API call

---

#### Test 5.8: Get Specific Attribute

```bash
curl http://localhost:8080/rag/attribute/TAX_RESIDENCY_COUNTRY
```

**Expected Output**:
```json
{
  "code": "TAX_RESIDENCY_COUNTRY",
  "risk_level": "HIGH",
  "data_type": "enum(ISO 3166-1 Alpha-2)",
  "business_context": "Jurisdiction where the entity is considered tax resident under FATCA/CRS regulations. Critical for automatic exchange of information and tax reporting obligations.",
  "synonyms": ["Tax Country", "Country of Tax Residence", "Tax Jurisdiction", "Fiscal Residence"],
  "regulatory_citations": ["FATCA §1471(b)(1)(D)", "CRS Common Reporting Standard", "OECD Model Tax Convention"],
  "example_values": ["US", "GB", "HK", "SG", "DE"]
}
```

**✅ Success Criteria**:
- Returns complete metadata
- All fields populated
- Response time < 10ms

---

### Step 6: Error Handling Tests

#### Test 6.1: Missing Query Parameter

```bash
curl "http://localhost:8080/rag/attribute_search"
```

**Expected Output**:
```json
{
  "error": "Bad Request",
  "message": "missing 'q' query parameter"
}
```

**✅ Status Code**: 400

---

#### Test 6.2: Attribute Not Found

```bash
curl "http://localhost:8080/rag/attribute/INVALID_CODE"
```

**Expected Output**:
```json
{
  "error": "Not Found",
  "message": "attribute not found: INVALID_CODE"
}
```

**✅ Status Code**: 404

---

#### Test 6.3: Invalid Similar Attributes Source

```bash
curl "http://localhost:8080/rag/similar_attributes?code=NONEXISTENT"
```

**Expected Output**:
```json
{
  "error": "Internal Server Error",
  "message": "failed to find similar attributes: metadata not found for attribute: NONEXISTENT"
}
```

**✅ Status Code**: 500 or 404

---

### Step 7: Automated Test Suite

Run the comprehensive test script:

```bash
./scripts/test_semantic_search.sh
```

**Expected Output**:
```
================================================
🧪 Testing RAG & Vector Search Functionality
================================================

✅ Database connection successful
✅ pgvector extension enabled
✅ Already seeded (20 attributes with embeddings)

================================================
TEST 1: Metadata Statistics
================================================
[Statistics output...]

================================================
TEST 2: Semantic Search - Tax Compliance
================================================
[Search results...]

[... 10 tests total ...]

================================================
✅ All tests completed successfully!
================================================

🎉 RAG & Vector Search System is operational!
```

**Duration**: 2-3 minutes

---

## 🐛 Troubleshooting

### Issue: "OPENAI_API_KEY environment variable not set"

**Solution**:
```bash
export OPENAI_API_KEY="sk-..."
# Add to ~/.bashrc or ~/.zshrc for persistence
```

---

### Issue: "type 'vector' does not exist"

**Solution**:
```bash
psql -d kyc_dsl -c "CREATE EXTENSION vector;"
```

If still fails, install pgvector:
```bash
# macOS
brew install pgvector

# Ubuntu
sudo apt install postgresql-15-pgvector
```

---

### Issue: "failed to generate embedding: rate limit exceeded"

**Solution**:
- Wait 60 seconds and retry
- Check OpenAI dashboard for usage
- Verify API key has sufficient credits

---

### Issue: "no results found" from semantic search

**Solution**:
```bash
# Check embeddings exist
psql -d kyc_dsl -c "SELECT COUNT(*) FROM kyc_attribute_metadata WHERE embedding IS NOT NULL;"

# Should return 20
# If 0, re-run: ./bin/kycctl seed-metadata
```

---

### Issue: Slow API responses

**Solution**:
```sql
-- Check vector index exists
SELECT indexname FROM pg_indexes 
WHERE tablename = 'kyc_attribute_metadata' 
  AND indexname LIKE '%embedding%';

-- Rebuild if needed
DROP INDEX IF EXISTS idx_attribute_metadata_embedding;
CREATE INDEX idx_attribute_metadata_embedding
    ON kyc_attribute_metadata 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
```

---

## 📊 Performance Benchmarks

### Expected Performance

| Operation                | Expected Time |
|--------------------------|---------------|
| Embedding generation     | 200-800ms     |
| Vector search (10K rows) | 5-20ms        |
| API health check         | 2-5ms         |
| Semantic search (w/ API) | 200-500ms     |
| Similar attributes       | 10-30ms       |
| Text search              | 5-15ms        |

### If Performance is Degraded

1. Check database connection pool
2. Verify vector index is being used (EXPLAIN ANALYZE)
3. Monitor OpenAI API latency
4. Check system resources (CPU, memory)
5. Review logs for errors

---

## ✅ Final Verification Checklist

Run through this checklist to ensure everything works:

- [ ] ✅ CLI: `seed-metadata` completes successfully
- [ ] ✅ CLI: `metadata-stats` shows 100% coverage
- [ ] ✅ CLI: `search-metadata` returns relevant results
- [ ] ✅ CLI: `similar-attributes` finds related items
- [ ] ✅ CLI: `text-search` works without API calls
- [ ] ✅ API: `/rag/health` returns "healthy"
- [ ] ✅ API: `/rag/stats` shows correct counts
- [ ] ✅ API: `/rag/attribute_search` returns ranked results
- [ ] ✅ API: `/rag/similar_attributes` finds similar items
- [ ] ✅ API: `/rag/text_search` finds keywords
- [ ] ✅ API: `/rag/attribute/{code}` returns metadata
- [ ] ✅ Errors: 400/404/500 responses work correctly
- [ ] ✅ Automated: Test script passes all 10 tests

---

## 🎯 Success Criteria

Your RAG system is working correctly if:

1. **Semantic Search Quality**:
   - Query "tax reporting" → TAX_RESIDENCY_COUNTRY is #1
   - Query "beneficial owner" → UBO_NAME is #1
   - Query "risk factors" → PEP_STATUS, SANCTIONS in top 3
   - Similarity scores > 0.8 for relevant results

2. **Performance**:
   - Embedding generation: <1s per attribute
   - Vector search: <50ms
   - API responses: <500ms (including OpenAI)

3. **Coverage**:
   - 20 attributes with embeddings (100%)
   - All risk levels represented
   - All searches return results

4. **Reliability**:
   - No errors in normal operations
   - Graceful error handling
   - Retry logic works for API failures

---

## 📚 Next Steps

After testing successfully:

1. **Production Deployment**: See [API_DOCUMENTATION.md](API_DOCUMENTATION.md#production-deployment)
2. **Agent Integration**: See [RAG_QUICKSTART.md](RAG_QUICKSTART.md#real-world-use-cases)
3. **Performance Tuning**: See [RAG_VECTOR_SEARCH.md](RAG_VECTOR_SEARCH.md#performance--optimization)
4. **Monitoring**: Set up metrics and alerts

---

**Testing Complete!** 🎉

Your KYC-DSL RAG system is now fully operational and ready for production use.