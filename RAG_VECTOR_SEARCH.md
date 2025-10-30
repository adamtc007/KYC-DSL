# RAG & Vector Search Documentation

**Version**: 1.4  
**Feature**: Vectorization & RAG Readiness for KYC Ontology  
**Status**: Production Ready 🚀

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Database Schema](#database-schema)
4. [Setup & Configuration](#setup--configuration)
5. [CLI Commands](#cli-commands)
6. [API Usage](#api-usage)
7. [Semantic Search Examples](#semantic-search-examples)
8. [Agent Integration](#agent-integration)
9. [Performance & Optimization](#performance--optimization)
10. [Troubleshooting](#troubleshooting)

---

## Overview

The KYC-DSL RAG (Retrieval-Augmented Generation) system provides semantic search capabilities over the regulatory data ontology. It enables:

- **Semantic Search**: Find attributes by meaning, not just keywords
- **Attribute Clustering**: Auto-generate logical groupings based on embeddings
- **Agent Integration**: Power AI agents with contextual regulatory knowledge
- **Synonym Resolution**: Map natural language to formal attribute codes
- **Regulatory Context**: Ground agent responses in actual regulations

### Key Benefits

✅ **Semantic Understanding**: Query "tax reporting requirements" → finds FATCA_STATUS, CRS_CLASSIFICATION, TAX_RESIDENCY_COUNTRY  
✅ **Explainability**: Every attribute includes regulatory citations and business context  
✅ **Jurisdiction-Aware**: Filter attributes by regulatory regime (FATCA, CRS, AMLD5, etc.)  
✅ **Version Control**: All embeddings tracked with metadata timestamps  
✅ **Incremental Updates**: Generate embeddings for new attributes on-demand  

---

## Architecture

### Component Stack

```
┌─────────────────────────────────────────────────────────────┐
│                     AI Agent / Application                   │
├─────────────────────────────────────────────────────────────┤
│              CLI Commands (kycctl search-metadata)           │
├─────────────────────────────────────────────────────────────┤
│         Go Application Layer (internal/rag)                  │
│  ┌──────────────────┐  ┌──────────────────────────────┐    │
│  │  Embedder        │  │  MetadataRepo                 │    │
│  │  - OpenAI API    │  │  - Vector Search              │    │
│  │  - Batch Gen     │  │  - Similarity Queries         │    │
│  │  - Retry Logic   │  │  - Text Search                │    │
│  └──────────────────┘  └──────────────────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│              PostgreSQL + pgvector Extension                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  kyc_attribute_metadata                              │  │
│  │  - attribute_code (PK)                               │  │
│  │  - synonyms TEXT[]                                   │  │
│  │  - business_context TEXT                             │  │
│  │  - regulatory_citations TEXT[]                       │  │
│  │  - embedding vector(1536)  ← OpenAI embeddings      │  │
│  │  - risk_level, data_type, examples, etc.            │  │
│  └──────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│              OpenAI API (text-embedding-3-large)             │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Embedding Generation**:
   ```
   Attribute Metadata → Concatenate Text → OpenAI API → 1536-dim Vector → PostgreSQL
   ```

2. **Semantic Search**:
   ```
   User Query → Generate Query Embedding → Vector Similarity Search → Ranked Results
   ```

3. **Similarity Finding**:
   ```
   Source Attribute → Retrieve Embedding → Compare with All Embeddings → Top-K Similar
   ```

---

## Database Schema

### Migration: `006_attribute_metadata.sql`

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Main metadata table with embeddings
CREATE TABLE IF NOT EXISTS kyc_attribute_metadata (
    id SERIAL PRIMARY KEY,
    attribute_code TEXT UNIQUE NOT NULL REFERENCES kyc_attributes(code),
    synonyms TEXT[],
    data_type TEXT,
    domain_values TEXT[],
    risk_level TEXT CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    example_values TEXT[],
    regulatory_citations TEXT[],
    business_context TEXT,
    embedding vector(1536),  -- OpenAI text-embedding-3-large
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Vector similarity index (IVFFlat for cosine distance)
CREATE INDEX idx_attribute_metadata_embedding
    ON kyc_attribute_metadata 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Standard indexes
CREATE INDEX idx_attribute_metadata_code ON kyc_attribute_metadata(attribute_code);
CREATE INDEX idx_attribute_metadata_risk ON kyc_attribute_metadata(risk_level);
CREATE INDEX idx_attribute_metadata_synonyms ON kyc_attribute_metadata USING GIN(synonyms);
```

### Embedding Text Format

Each attribute's embedding is generated from concatenated text:

```
<ATTRIBUTE_CODE>. Definition: <BUSINESS_CONTEXT>. Synonyms: <SYNONYMS>. Citations: <REGULATORY_CITATIONS>. Examples: <EXAMPLE_VALUES>
```

**Example**:
```
TAX_RESIDENCY_COUNTRY. Definition: Jurisdiction where the entity is considered tax resident under FATCA/CRS regulations. Synonyms: Tax Country, Country of Tax Residence, Tax Jurisdiction. Citations: FATCA §1471(b)(1)(D), CRS Common Reporting Standard. Examples: US, GB, HK, SG
```

---

## Setup & Configuration

### Prerequisites

1. **PostgreSQL with pgvector**:
   ```bash
   # Install pgvector extension
   brew install pgvector  # macOS
   # or
   sudo apt install postgresql-15-pgvector  # Ubuntu
   ```

2. **OpenAI API Key**:
   ```bash
   export OPENAI_API_KEY="sk-..."
   ```

3. **Go Dependencies**:
   ```bash
   cd KYC-DSL
   go get github.com/sashabaranov/go-openai
   go mod tidy
   ```

### Database Initialization

```bash
# Run migrations
psql -d kyc_dsl -f internal/storage/migrations/006_attribute_metadata.sql

# Verify pgvector is enabled
psql -d kyc_dsl -c "SELECT * FROM pg_extension WHERE extname='vector';"
```

### Build & Install

```bash
make build
make install
```

---

## CLI Commands

### 1. Seed Metadata with Embeddings

Generate embeddings for all attributes in the ontology:

```bash
./kycctl seed-metadata
```

**Output**:
```
🌱 Seeding Attribute Metadata with Embeddings...
================================================

📊 Processing 20 attributes...

[1/20] Processing: REGISTERED_NAME
  ✅ Seeded with 1536-dimensional embedding
[2/20] Processing: TAX_RESIDENCY_COUNTRY
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
```

**Features**:
- Automatic retry on API failures (3 attempts)
- Rate limiting (200ms delay between requests)
- Progress tracking
- Comprehensive error reporting

---

### 2. Semantic Search

Find attributes by semantic meaning:

```bash
./kycctl search-metadata "tax reporting requirements"
./kycctl search-metadata "beneficial ownership information"
./kycctl search-metadata "politically exposed persons" --limit=5
```

**Example Output**:
```
🔍 Semantic Search: "tax reporting requirements"
================================================

⚡ Generating query embedding...
🔎 Searching for top 10 matches...

📊 Found 10 matches:

─────────────────────────────────────────────────
Rank #1
─────────────────────────────────────────────────
🏷️  Code:           TAX_RESIDENCY_COUNTRY
📈 Similarity:      0.8742 (distance: 0.1258)
⚠️  Risk Level:      HIGH
📝 Data Type:       enum(ISO 3166-1 Alpha-2)
🔤 Synonyms:        Tax Country, Country of Tax Residence, Tax Jurisdiction
📖 Context:         Jurisdiction where the entity is considered tax resident under FATCA/CRS regulations...
📜 Citations:       FATCA §1471(b)(1)(D), CRS Common Reporting Standard
💡 Examples:        US, GB, HK, SG

─────────────────────────────────────────────────
Rank #2
─────────────────────────────────────────────────
🏷️  Code:           FATCA_STATUS
📈 Similarity:      0.8521 (distance: 0.1479)
⚠️  Risk Level:      HIGH
📝 Data Type:       enum
🔤 Synonyms:        FATCA Classification, Chapter 4 Status
📖 Context:         Entity classification under FATCA determining US withholding obligations...
📜 Citations:       FATCA §1471-1474, IRS Publication 5190

...
```

**How it works**:
1. Generates embedding for your query using OpenAI
2. Compares query embedding against all attribute embeddings
3. Returns top-K matches ranked by cosine similarity
4. Displays comprehensive metadata for each match

---

### 3. Find Similar Attributes

Find attributes semantically related to a given attribute:

```bash
./kycctl similar-attributes UBO_NAME
./kycctl similar-attributes SANCTIONS_SCREENING_STATUS --limit=5
```

**Example Output**:
```
🔍 Finding Similar Attributes to: UBO_NAME
================================================

📋 Source Attribute:
  Code:        UBO_NAME
  Risk Level:  CRITICAL
  Data Type:   string
  Synonyms:    Ultimate Beneficial Owner Name, Beneficial Owner, UBO
  Context:     Full legal name of the ultimate beneficial owner...

🔎 Finding top 10 similar attributes...

📊 Found 10 similar attributes:

─────────────────────────────────────────────────
Rank #1
─────────────────────────────────────────────────
🏷️  Code:           UBO_OWNERSHIP_PERCENT
📈 Similarity:      0.9123 (distance: 0.0877)
⚠️  Risk Level:      CRITICAL
📝 Data Type:       float
🔤 Synonyms:        Ownership Percentage, Beneficial Ownership %
📖 Context:         Percentage of ownership or voting rights held by the UBO...

─────────────────────────────────────────────────
Rank #2
─────────────────────────────────────────────────
🏷️  Code:           DIRECTOR_NAME
📈 Similarity:      0.8345 (distance: 0.1655)
⚠️  Risk Level:      HIGH
📝 Data Type:       string
🔤 Synonyms:        Board Member, Director, Company Director

...

================================================
💡 Clustering Suggestion:
   These attributes could form a cluster with UBO_NAME
   based on semantic similarity.
```

**Use Cases**:
- Discover related attributes for agent context
- Auto-generate attribute clusters
- Validate completeness of data collection
- Suggest additional due diligence items

---

### 4. Text Search (Traditional)

Search by keyword or synonym (no embedding required):

```bash
./kycctl text-search "ownership"
./kycctl text-search "PEP"
./kycctl text-search "residence"
```

**Example Output**:
```
🔍 Text Search: "ownership"
================================================

🔎 Searching attributes and synonyms...

📊 Found 3 matches:

─────────────────────────────────────────────────
Result #1
─────────────────────────────────────────────────
🏷️  Code:           UBO_OWNERSHIP_PERCENT
⚠️  Risk Level:      CRITICAL
📝 Data Type:       float
🔤 Synonyms:        Ownership Percentage, Beneficial Ownership %
📖 Context:         Percentage of ownership held by the UBO...
💡 Examples:        26.5, 50.0, 100.0

...
```

**When to use**:
- Exact keyword matching needed
- Quick lookups without API calls
- Synonym resolution
- Metadata validation

---

### 5. Metadata Statistics

View repository health and coverage:

```bash
./kycctl metadata-stats
```

**Example Output**:
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

================================================
✅ Statistics retrieved successfully.
```

---

## API Usage

### Programmatic Access

#### Generate Embeddings

```go
import (
    "context"
    "github.com/adamtc007/KYC-DSL/internal/rag"
    "github.com/adamtc007/KYC-DSL/internal/model"
)

// Initialize embedder
embedder := rag.NewEmbedder()

// Create metadata
metadata := model.AttributeMetadata{
    AttributeCode:       "CUSTOM_ATTRIBUTE",
    Synonyms:            []string{"Alternative Name"},
    BusinessContext:     "Description of the attribute",
    RegulatoryCitations: []string{"Regulation ABC"},
    ExampleValues:       []string{"Example1", "Example2"},
}

// Generate embedding
embedding, err := embedder.GenerateEmbedding(context.Background(), metadata)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Generated %d-dimensional embedding\n", len(embedding))
```

#### Semantic Search

```go
import (
    "github.com/adamtc007/KYC-DSL/internal/ontology"
    "github.com/adamtc007/KYC-DSL/internal/storage"
)

// Connect to database
db, _ := storage.ConnectPostgres()
repo := ontology.NewMetadataRepo(db)

// Generate query embedding
embedder := rag.NewEmbedder()
queryEmbedding, _ := embedder.GenerateEmbeddingFromText(ctx, "tax compliance")

// Search
results, _ := repo.SearchByVector(ctx, queryEmbedding, 10)

for _, result := range results {
    fmt.Printf("%s: similarity=%.4f\n", 
        result.AttributeCode, 
        result.SimilarityScore)
}
```

#### Find Similar Attributes

```go
// Find attributes similar to UBO_NAME
results, err := repo.FindSimilarAttributes(ctx, "UBO_NAME", 10)

for i, result := range results {
    fmt.Printf("#%d: %s (similarity: %.4f)\n",
        i+1,
        result.AttributeCode,
        result.SimilarityScore)
}
```

---

## Semantic Search Examples

### Tax & Reporting

**Query**: `"tax reporting requirements"`

**Top Results**:
1. TAX_RESIDENCY_COUNTRY (0.8742)
2. FATCA_STATUS (0.8521)
3. CRS_CLASSIFICATION (0.8312)
4. INCORPORATION_COUNTRY (0.7234)

---

### Ownership & Control

**Query**: `"who controls this entity"`

**Top Results**:
1. UBO_NAME (0.8956)
2. UBO_OWNERSHIP_PERCENT (0.8734)
3. DIRECTOR_NAME (0.8123)
4. REGISTERED_ADDRESS (0.6234)

---

### Risk Assessment

**Query**: `"money laundering risk factors"`

**Top Results**:
1. PEP_STATUS (0.9012)
2. SANCTIONS_SCREENING_STATUS (0.8845)
3. SOURCE_OF_FUNDS (0.8623)
4. ADVERSE_MEDIA_FLAG (0.8401)
5. CUSTOMER_RISK_RATING (0.8234)

---

### Identity Verification

**Query**: `"verify customer identity"`

**Top Results**:
1. REGISTERED_NAME (0.8512)
2. INCORPORATION_COUNTRY (0.8234)
3. REGISTERED_ADDRESS (0.8123)
4. DIRECTOR_NAME (0.7845)

---

## Agent Integration

### RAG Pattern for AI Agents

```python
# Example: Claude/GPT agent with KYC-DSL RAG

def handle_kyc_query(user_question: str):
    # 1. Generate embedding for user question
    query_embedding = generate_embedding(user_question)
    
    # 2. Retrieve relevant attributes from vector DB
    relevant_attributes = search_vector_db(query_embedding, limit=5)
    
    # 3. Construct context for LLM
    context = build_context(relevant_attributes)
    
    # 4. Generate response with grounded context
    prompt = f"""
    User Question: {user_question}
    
    Relevant Regulatory Attributes:
    {context}
    
    Please answer based on the provided regulatory context.
    """
    
    return llm.generate(prompt)

def build_context(attributes):
    context = []
    for attr in attributes:
        context.append(f"""
        Attribute: {attr.code}
        Definition: {attr.business_context}
        Regulations: {', '.join(attr.regulatory_citations)}
        Risk Level: {attr.risk_level}
        """)
    return "\n\n".join(context)
```

### Use Cases

1. **Agent-Driven Case Assembly**:
   - Query: "What attributes do I need for EU investment fund KYC?"
   - RAG retrieves: AMLD5-related attributes
   - Agent generates: DSL `(data-dictionary ...)` section

2. **Regulatory Explainability**:
   - Query: "Why do we need UBO information?"
   - RAG retrieves: UBO_NAME with citations
   - Agent explains: AMLD5 Article 3, FATF Rec 24 requirements

3. **Synonym Resolution**:
   - User says: "Company Name"
   - RAG finds: REGISTERED_NAME (synonym match)
   - Agent maps to correct attribute code

4. **Risk-Based Questioning**:
   - Agent identifies HIGH/CRITICAL risk attributes
   - Prioritizes data collection accordingly
   - Applies enhanced due diligence rules

---

## Performance & Optimization

### Vector Index Tuning

The `ivfflat` index uses approximate nearest neighbor (ANN) search:

```sql
-- Adjust 'lists' parameter based on dataset size
-- Rule of thumb: lists = sqrt(total_rows)
-- For 1000 attributes: lists = 32
-- For 10000 attributes: lists = 100

CREATE INDEX idx_attribute_metadata_embedding
    ON kyc_attribute_metadata 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
```

### Query Performance

**Typical Performance**:
- Embedding generation: 200-800ms per attribute
- Vector search (10K attributes): 5-20ms
- Batch embedding (20 attributes): 15-30 seconds

**Optimization Tips**:
1. Use batch operations when possible
2. Cache frequent query embeddings
3. Adjust `lists` parameter for index
4. Consider `hnsw` index for larger datasets (>50K attributes)

### Rate Limiting

OpenAI API limits:
- **text-embedding-3-large**: 5,000 requests/min
- **Embeddings/min**: ~1M tokens/min

Current implementation:
- 200ms delay between requests (safe rate)
- 3 retry attempts with exponential backoff
- Batch processing with progress tracking

---

## Troubleshooting

### pgvector Extension Not Found

```bash
# Error: extension "vector" does not exist

# Solution: Install pgvector
brew install pgvector  # macOS
sudo apt install postgresql-15-pgvector  # Ubuntu

# Then in psql:
CREATE EXTENSION vector;
```

---

### OpenAI API Key Not Set

```bash
# Error: OPENAI_API_KEY environment variable not set

# Solution:
export OPENAI_API_KEY="sk-..."

# Or add to ~/.bashrc or ~/.zshrc
echo 'export OPENAI_API_KEY="sk-..."' >> ~/.bashrc
```

---

### Slow Vector Search

```sql
-- Check if index exists
SELECT indexname FROM pg_indexes 
WHERE tablename = 'kyc_attribute_metadata' 
  AND indexname LIKE '%embedding%';

-- Rebuild index if needed
DROP INDEX IF EXISTS idx_attribute_metadata_embedding;
CREATE INDEX idx_attribute_metadata_embedding
    ON kyc_attribute_metadata 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Analyze table
ANALYZE kyc_attribute_metadata;
```

---

### Embedding Dimension Mismatch

```bash
# Error: vector dimension mismatch

# Cause: Using different OpenAI model than configured

# Solution: Ensure consistency
# - Database schema: vector(1536)
# - OpenAI model: text-embedding-3-large (1536 dims)

# If using text-embedding-3-small (512 dims):
ALTER TABLE kyc_attribute_metadata 
ALTER COLUMN embedding TYPE vector(512);
```

---

### Low Similarity Scores

**Issue**: All similarity scores below 0.5

**Possible Causes**:
1. Query too generic
2. Insufficient metadata in business_context
3. Missing synonyms
4. Wrong embedding model

**Solutions**:
1. Be more specific in queries
2. Enrich business_context with more detail
3. Add comprehensive synonyms
4. Verify using `text-embedding-3-large`

---

## SQL Reference

### Direct Vector Queries

```sql
-- Find attributes similar to UBO_NAME
SELECT 
    attribute_code, 
    risk_level,
    1 - (embedding <=> (
        SELECT embedding 
        FROM kyc_attribute_metadata 
        WHERE attribute_code = 'UBO_NAME'
    )) as similarity
FROM kyc_attribute_metadata
WHERE attribute_code != 'UBO_NAME'
  AND embedding IS NOT NULL
ORDER BY embedding <=> (
    SELECT embedding 
    FROM kyc_attribute_metadata 
    WHERE attribute_code = 'UBO_NAME'
)
LIMIT 10;
```

### Embedding Coverage Report

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

### Attribute Clustering by Similarity

```sql
-- Find dense clusters (attributes with many similar neighbors)
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

## Next Steps

### Planned Enhancements

1. **Auto-Clustering Algorithm**:
   - K-means clustering on embeddings
   - Automatic cluster naming
   - Cluster quality metrics

2. **Multi-Modal Embeddings**:
   - Document text embeddings
   - Regulation section embeddings
   - Cross-modal similarity search

3. **Embedding Cache**:
   - Redis cache for frequent queries
   - Reduce API calls
   - Faster response times

4. **Hybrid Search**:
   - Combine vector + text search
   - Weighted scoring
   - Best of both approaches

5. **Agent SDK**:
   - Python/TypeScript wrappers
   - Pre-built RAG patterns
   - Streaming search results

---

## References

- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [OpenAI Embeddings Guide](https://platform.openai.com/docs/guides/embeddings)
- [Vector Similarity Search](https://www.postgresql.org/docs/current/functions-vector.html)
- [RAG Best Practices](https://www.pinecone.io/learn/retrieval-augmented-generation/)

---

**Last Updated**: 2024  
**Version**: 1.4  
**Status**: Production Ready 🚀