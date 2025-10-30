# RAG & Vector Search Quickstart Guide

**Version**: 1.4  
**Time to Complete**: 10 minutes  
**Prerequisites**: PostgreSQL with pgvector, OpenAI API Key

---

## 🚀 5-Minute Setup

### 1. Install Dependencies

```bash
cd KYC-DSL

# Install Go dependencies
go get github.com/sashabaranov/go-openai
go mod tidy

# Build the binary
make build
```

### 2. Configure Environment

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="sk-..."

# Verify PostgreSQL connection
export PGDATABASE=kyc_dsl
psql -c "SELECT version();"
```

### 3. Enable pgvector Extension

```bash
# Install pgvector (if not already installed)
# macOS:
brew install pgvector

# Ubuntu/Debian:
sudo apt install postgresql-15-pgvector

# Enable in database
psql -d kyc_dsl -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

### 4. Run Migration

```bash
# Apply the vector search migration
psql -d kyc_dsl -f internal/storage/migrations/006_attribute_metadata.sql

# Verify tables created
psql -d kyc_dsl -c "\dt kyc_attribute_metadata"
```

### 5. Seed Metadata with Embeddings

```bash
# Generate embeddings for all attributes (takes ~15-30 seconds)
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
...

✅ Successfully seeded: 20 attributes
⏱️  Total time: 15.3s
```

---

## 🔍 Quick Examples

### Example 1: Find Tax-Related Attributes

```bash
./bin/kycctl search-metadata "tax reporting requirements"
```

**What you'll get**:
- TAX_RESIDENCY_COUNTRY (similarity: 0.87)
- FATCA_STATUS (similarity: 0.85)
- CRS_CLASSIFICATION (similarity: 0.83)
- INCORPORATION_COUNTRY (similarity: 0.72)

Each result includes:
- Similarity score (0-1, higher is more relevant)
- Risk level (LOW, MEDIUM, HIGH, CRITICAL)
- Business context explanation
- Regulatory citations (FATCA, CRS, AMLD5, etc.)
- Example values

---

### Example 2: Find Similar Attributes

```bash
./bin/kycctl similar-attributes UBO_NAME
```

**What you'll get**:
- UBO_OWNERSHIP_PERCENT (similarity: 0.91)
- DIRECTOR_NAME (similarity: 0.83)
- REGISTERED_NAME (similarity: 0.76)

Perfect for:
- Discovering related data points
- Building attribute clusters
- Suggesting additional due diligence
- Agent context expansion

---

### Example 3: Quick Text Search

```bash
./bin/kycctl text-search "ownership"
```

**What you'll get**:
- Keyword matches in attribute codes
- Synonym matches
- Business context matches

No API calls required!

---

### Example 4: Repository Statistics

```bash
./bin/kycctl metadata-stats
```

**What you'll get**:
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
```

---

## 💡 Real-World Use Cases

### Use Case 1: AI Agent Context Retrieval

**Scenario**: Your AI agent needs to build a KYC case for an EU investment fund.

```python
# Agent workflow
query = "What attributes do I need for EU investment fund KYC?"

# 1. Semantic search
attributes = search_vector_db(query, limit=10)

# 2. Build context
context = {
    "attributes": [
        {
            "code": "TAX_RESIDENCY_COUNTRY",
            "definition": "Jurisdiction where entity is tax resident...",
            "regulations": ["FATCA", "CRS", "AMLD5"],
            "risk_level": "HIGH"
        },
        # ... more attributes
    ]
}

# 3. Generate DSL
agent.generate_dsl(context)
```

**Result**: Agent generates compliant DSL with proper citations.

---

### Use Case 2: Regulatory Explainability

**Scenario**: Compliance officer asks "Why do we need UBO information?"

```bash
./bin/kycctl search-metadata "ultimate beneficial owner requirements"
```

**Agent Response**:
> "Under AMLD5 Article 3 and FATF Recommendation 24, financial institutions must 
> identify ultimate beneficial owners (UBOs) who own or control more than 25% of 
> an entity. This is a CRITICAL risk attribute required for all EU-regulated entities."

**Citations**: Pulled directly from `regulatory_citations` field.

---

### Use Case 3: Synonym Resolution

**Scenario**: User says "Company Name" but system needs `REGISTERED_NAME`.

```bash
./bin/kycctl text-search "Company Name"
```

**Result**:
```
🏷️  Code: REGISTERED_NAME
🔤 Synonyms: Legal Name, Company Name, Entity Name, Corporate Name
```

**Agent maps**: "Company Name" → `REGISTERED_NAME` → DSL attribute code

---

### Use Case 4: Risk-Based Prioritization

**Scenario**: Agent needs to prioritize data collection for high-risk entity.

```bash
./bin/kycctl search-metadata "money laundering risk indicators"
```

**Result**: CRITICAL/HIGH risk attributes first
- PEP_STATUS (CRITICAL)
- SANCTIONS_SCREENING_STATUS (CRITICAL)
- SOURCE_OF_FUNDS (CRITICAL)
- ADVERSE_MEDIA_FLAG (HIGH)

**Agent action**: Applies enhanced due diligence, requests additional documents.

---

## 🧪 Test Your Setup

Run the comprehensive test suite:

```bash
./scripts/test_semantic_search.sh
```

**Tests Include**:
1. ✅ Database connectivity
2. ✅ pgvector extension
3. ✅ Metadata statistics
4. ✅ Semantic search (tax, ownership, risk)
5. ✅ Similar attributes
6. ✅ Text search
7. ✅ Direct SQL vector queries
8. ✅ Embedding coverage
9. ✅ Cluster detection
10. ✅ Synonym resolution

**Expected Duration**: 2-3 minutes

---

## 📊 Direct SQL Queries

### Query 1: Find Similar Attributes

```sql
SELECT 
    attribute_code, 
    risk_level,
    ROUND((1 - (embedding <=> (
        SELECT embedding 
        FROM kyc_attribute_metadata 
        WHERE attribute_code = 'UBO_NAME'
    )))::numeric, 4) as similarity
FROM kyc_attribute_metadata
WHERE attribute_code != 'UBO_NAME'
  AND embedding IS NOT NULL
ORDER BY similarity DESC
LIMIT 5;
```

---

### Query 2: Embedding Coverage by Risk Level

```sql
SELECT 
    risk_level,
    COUNT(*) as total,
    COUNT(embedding) as with_embedding,
    ROUND(100.0 * COUNT(embedding) / COUNT(*), 1) as coverage_pct
FROM kyc_attribute_metadata
GROUP BY risk_level
ORDER BY coverage_pct DESC;
```

---

### Query 3: Detect Attribute Clusters

```sql
WITH similarity_counts AS (
    SELECT 
        a1.attribute_code,
        COUNT(*) as similar_count
    FROM kyc_attribute_metadata a1
    CROSS JOIN kyc_attribute_metadata a2
    WHERE a1.attribute_code != a2.attribute_code
      AND 1 - (a1.embedding <=> a2.embedding) > 0.8
    GROUP BY a1.attribute_code
)
SELECT * FROM similarity_counts
ORDER BY similar_count DESC
LIMIT 10;
```

---

## 🐍 Python Integration Example

```python
import psycopg2
import openai
import numpy as np

# Initialize
openai.api_key = "sk-..."
conn = psycopg2.connect("dbname=kyc_dsl")

def semantic_search(query: str, limit: int = 10):
    # Generate query embedding
    response = openai.Embedding.create(
        model="text-embedding-3-large",
        input=query
    )
    query_embedding = response['data'][0]['embedding']
    
    # Search database
    cur = conn.cursor()
    cur.execute("""
        SELECT 
            attribute_code,
            business_context,
            regulatory_citations,
            risk_level,
            1 - (embedding <=> %s::vector) as similarity
        FROM kyc_attribute_metadata
        WHERE embedding IS NOT NULL
        ORDER BY embedding <=> %s::vector
        LIMIT %s
    """, (query_embedding, query_embedding, limit))
    
    return cur.fetchall()

# Example usage
results = semantic_search("tax compliance requirements")

for attr_code, context, citations, risk, similarity in results:
    print(f"{attr_code}: {similarity:.3f}")
    print(f"  Risk: {risk}")
    print(f"  Citations: {citations}")
    print()
```

---

## 🔧 Troubleshooting

### Issue 1: No Results from Semantic Search

**Symptoms**:
```
❌ No results found.
```

**Solutions**:
1. Check embeddings exist:
   ```sql
   SELECT COUNT(*) FROM kyc_attribute_metadata WHERE embedding IS NOT NULL;
   ```
2. Re-run seeding:
   ```bash
   ./bin/kycctl seed-metadata
   ```

---

### Issue 2: OpenAI API Rate Limit

**Symptoms**:
```
❌ Failed to generate embedding: rate limit exceeded
```

**Solutions**:
1. Wait 60 seconds and retry
2. Reduce batch size in code
3. Check API usage at platform.openai.com

---

### Issue 3: pgvector Extension Error

**Symptoms**:
```sql
ERROR: type "vector" does not exist
```

**Solutions**:
```bash
# Install pgvector
brew install pgvector  # macOS
sudo apt install postgresql-15-pgvector  # Ubuntu

# Enable in database
psql -d kyc_dsl -c "CREATE EXTENSION vector;"

# Verify
psql -d kyc_dsl -c "SELECT * FROM pg_extension WHERE extname='vector';"
```

---

### Issue 4: Slow Vector Queries

**Symptoms**: Queries taking >1 second

**Solutions**:
```sql
-- Rebuild index
DROP INDEX IF EXISTS idx_attribute_metadata_embedding;
CREATE INDEX idx_attribute_metadata_embedding
    ON kyc_attribute_metadata 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Analyze table
ANALYZE kyc_attribute_metadata;

-- Verify index usage
EXPLAIN ANALYZE
SELECT * FROM kyc_attribute_metadata
ORDER BY embedding <=> '[...]'::vector
LIMIT 10;
```

---

## 📚 Next Steps

### 1. Read Full Documentation
- [RAG_VECTOR_SEARCH.md](RAG_VECTOR_SEARCH.md) - Complete reference
- [REGULATORY_ONTOLOGY.md](REGULATORY_ONTOLOGY.md) - Ontology details
- [CLAUDE.md](CLAUDE.md) - Development guide

### 2. Explore Advanced Features
- Auto-clustering algorithms
- Multi-modal embeddings
- Hybrid search (vector + text)
- Agent SDK integration

### 3. Integrate with Your Application
- REST API wrapper (coming soon)
- Python/TypeScript SDKs (planned)
- GraphQL endpoint (planned)

### 4. Contribute
- Add more attribute metadata
- Improve embedding prompts
- Build agent integrations
- Share use cases

---

## 🎯 Key Takeaways

✅ **Setup is fast**: 5-10 minutes from zero to semantic search  
✅ **Powerful queries**: Natural language → relevant attributes  
✅ **Regulatory grounded**: Every result includes citations  
✅ **Agent-ready**: Perfect for RAG patterns and LLM context  
✅ **Production-ready**: Battle-tested with real KYC workflows  

---

## 🆘 Need Help?

- **Documentation**: See [RAG_VECTOR_SEARCH.md](RAG_VECTOR_SEARCH.md)
- **Examples**: Run `./scripts/test_semantic_search.sh`
- **Issues**: Check [Troubleshooting](#troubleshooting) section
- **Questions**: Review inline code documentation

---

**Last Updated**: 2024  
**Version**: 1.4  
**Status**: Production Ready 🚀