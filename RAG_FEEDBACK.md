# RAG Feedback Loop System

## Overview

The RAG Feedback Loop enables **continuous learning** and improvement of search relevance scores through user and AI agent feedback. This self-correcting system automatically adjusts attribute-document relationship weights based on real-world usage patterns.

**Version**: 1.5  
**Status**: Production Ready  
**Dependencies**: PostgreSQL with pgvector, OpenAI Embeddings

---

## 🎯 Key Features

- **Self-Learning**: Automatically adjusts relevance scores based on feedback
- **Multi-Agent Support**: Accepts feedback from humans, AI agents, and automated systems
- **Confidence Weighting**: Scales impact based on feedback confidence (0.0-1.0)
- **Real-Time Updates**: Database triggers apply changes immediately
- **Analytics Dashboard**: Track sentiment trends, agent performance, and learning progress
- **Audit Trail**: Complete history of all feedback events

---

## 🏗️ Architecture

### Data Flow

```
User/Agent → API Endpoint → Feedback Validation → Database Insert
                                                         ↓
                                                   Trigger Fires
                                                         ↓
                                        Update Relevance Scores (kyc_attr_doc_links)
```

### Components

1. **Database Schema** (`007_rag_feedback.sql`)
   - `rag_feedback` table for storing feedback entries
   - `update_relevance()` trigger function for automatic score adjustment
   - Views for analytics and summaries

2. **Go Models** (`internal/model/rag_feedback.go`)
   - `Feedback` - Core feedback structure
   - `FeedbackRequest` - API request model
   - `FeedbackAnalytics` - Analytics aggregation

3. **Repository Layer** (`internal/ontology/feedback_repo.go`)
   - Database operations
   - Analytics queries
   - Summary generation

4. **API Handler** (`internal/api/rag_handler.go`)
   - HTTP endpoints
   - Request validation
   - Response formatting

---

## 📊 Database Schema

### Main Table: `rag_feedback`

```sql
CREATE TABLE rag_feedback (
    id SERIAL PRIMARY KEY,
    query_text TEXT NOT NULL,              -- Original search query
    attribute_code TEXT,                   -- FK to kyc_attributes
    document_code TEXT,                    -- FK to kyc_documents
    regulation_code TEXT,                  -- FK to kyc_regulations
    feedback feedback_sentiment,           -- positive/negative/neutral
    confidence FLOAT DEFAULT 1.0,          -- Impact weight (0.0-1.0)
    agent_name TEXT,                       -- Who provided feedback
    agent_type TEXT,                       -- human/ai/automated
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Trigger Function: `update_relevance()`

Automatically adjusts `kyc_attr_doc_links.relevance_score` on feedback insert:

```sql
relevance_score = CASE
    WHEN feedback = 'positive' THEN relevance_score + (0.05 * confidence)
    WHEN feedback = 'negative' THEN relevance_score - (0.05 * confidence)
    ELSE relevance_score
END
```

**Impact Examples:**
- Positive feedback (confidence=1.0): +0.05 per event
- Positive feedback (confidence=0.5): +0.025 per event
- Negative feedback (confidence=0.8): -0.04 per event

---

## 🔌 API Endpoints

### 1. Submit Feedback

**POST** `/rag/feedback`

Submit feedback on search results to improve relevance.

**Request Body:**
```json
{
  "query_text": "beneficial owner name",        // Required
  "attribute_code": "UBO_NAME",                 // Optional
  "document_code": "W8BEN",                     // Optional
  "regulation_code": "AMLD5",                   // Optional
  "feedback": "positive",                       // positive/negative/neutral
  "confidence": 0.9,                            // 0.0-1.0
  "agent_name": "adam",                         // Optional
  "agent_type": "human"                         // human/ai/automated
}
```

**Response:**
```json
{
  "status": "ok",
  "id": 42,
  "feedback": "positive",
  "agent_name": "adam",
  "created_at": "2024-10-30T09:41:05Z"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/rag/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "query_text": "beneficial owner name",
    "attribute_code": "UBO_NAME",
    "document_code": "W8BEN",
    "feedback": "positive",
    "confidence": 0.9,
    "agent_name": "claude",
    "agent_type": "ai"
  }'
```

---

### 2. Get Recent Feedback

**GET** `/rag/feedback/recent?limit=50`

Retrieve the most recent feedback entries.

**Parameters:**
- `limit` (optional) - Max results, default: 50

**Response:**
```json
{
  "count": 10,
  "feedbacks": [
    {
      "id": 42,
      "query_text": "beneficial owner name",
      "attribute_code": "UBO_NAME",
      "document_code": "W8BEN",
      "feedback": "positive",
      "confidence": 0.9,
      "agent_name": "adam",
      "agent_type": "human",
      "created_at": "2024-10-30T09:41:05Z"
    }
  ]
}
```

**Example:**
```bash
curl http://localhost:8080/rag/feedback/recent?limit=10
```

---

### 3. Get Feedback Analytics

**GET** `/rag/feedback/analytics?top=10`

Retrieve comprehensive feedback analytics including sentiment distribution and trends.

**Parameters:**
- `top` (optional) - Top N attributes to include, default: 10

**Response:**
```json
{
  "total_feedback": 156,
  "positive_count": 120,
  "negative_count": 25,
  "neutral_count": 11,
  "avg_confidence": 0.85,
  "by_agent_type": {
    "human": 80,
    "ai": 60,
    "automated": 16
  },
  "top_attributes": [
    {
      "attribute_code": "UBO_NAME",
      "feedback": "positive",
      "feedback_count": 15,
      "avg_confidence": 0.9,
      "agent_types": "human, ai"
    }
  ],
  "recent_feedback": [...],
  "sentiment_trend": [...]
}
```

**Example:**
```bash
curl http://localhost:8080/rag/feedback/analytics?top=20
```

---

### 4. Get Feedback by Attribute

**GET** `/rag/feedback/attribute/{code}`

Retrieve all feedback for a specific attribute.

**Example:**
```bash
curl http://localhost:8080/rag/feedback/attribute/UBO_NAME
```

---

### 5. Get Feedback Summary

**GET** `/rag/feedback/summary?limit=20`

Get aggregated feedback summary by sentiment and agent type.

**Parameters:**
- `limit` (optional) - Max attributes, default: 20

**Example:**
```bash
curl http://localhost:8080/rag/feedback/summary
```

---

## 🚀 Quick Start

### 1. Apply Migration

```bash
psql -d kyc_dsl -f internal/storage/migrations/007_rag_feedback.sql
```

### 2. Start Server

```bash
export OPENAI_API_KEY="sk-..."
go run cmd/kycserver/main.go
```

### 3. Submit Test Feedback

```bash
curl -X POST http://localhost:8080/rag/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "query_text": "tax compliance requirements",
    "attribute_code": "TAX_RESIDENCY_COUNTRY",
    "document_code": "W8BEN",
    "feedback": "positive",
    "confidence": 0.9,
    "agent_type": "human"
  }'
```

### 4. Verify Learning Effect

```sql
-- Check initial score
SELECT attribute_code, document_code, relevance_score
FROM kyc_attr_doc_links
WHERE attribute_code='TAX_RESIDENCY_COUNTRY';

-- Submit positive feedback (via API)

-- Check updated score
SELECT attribute_code, document_code, relevance_score
FROM kyc_attr_doc_links
WHERE attribute_code='TAX_RESIDENCY_COUNTRY';
-- Score should have increased by ~0.05
```

---

## 🧪 Testing

### Run Test Suite

```bash
chmod +x scripts/test_feedback.sh
./scripts/test_feedback.sh
```

The test suite includes:

1. ✅ Positive feedback (human agent)
2. ✅ Negative feedback (AI agent)
3. ✅ Neutral feedback (automated agent)
4. ✅ High confidence feedback (1.0)
5. ✅ Low confidence feedback (0.3)
6. ✅ Recent feedback retrieval
7. ✅ Analytics generation
8. ✅ Feedback by attribute
9. ✅ Summary statistics
10. ✅ Learning effect verification
11. ✅ Error handling - missing fields
12. ✅ Error handling - invalid confidence
13. ✅ Batch submission

---

## 🤖 Agent Types

### Human Agents

**Use Case**: Manual compliance review, expert judgment  
**Confidence**: Typically 0.8-1.0  
**Impact**: High - trusted human expertise

```json
{
  "agent_type": "human",
  "agent_name": "compliance_officer",
  "confidence": 1.0
}
```

---

### AI Agents

**Use Case**: Claude, GPT-4, custom models evaluating search quality  
**Confidence**: Typically 0.6-0.9  
**Impact**: Medium - ML-based assessment

```json
{
  "agent_type": "ai",
  "agent_name": "claude-3-opus",
  "confidence": 0.8
}
```

---

### Automated Agents

**Use Case**: Rule-based validation, heuristics, A/B testing  
**Confidence**: Typically 0.3-0.7  
**Impact**: Low-Medium - algorithmic evaluation

```json
{
  "agent_type": "automated",
  "agent_name": "relevance_validator",
  "confidence": 0.5
}
```

---

## 📈 Confidence Weighting

The `confidence` parameter allows fine-grained control over feedback impact:

| Confidence | Interpretation | Use Case |
|-----------|---------------|----------|
| 1.0 | Absolute certainty | Expert human validation |
| 0.9 | Very high confidence | Human review, high-quality AI |
| 0.7-0.8 | High confidence | Standard AI evaluation |
| 0.5-0.6 | Medium confidence | Automated heuristics |
| 0.3-0.4 | Low confidence | Exploratory signals |
| < 0.3 | Very low confidence | Weak signals, testing |

**Formula:**
```
score_adjustment = base_adjustment (0.05) × confidence
```

**Examples:**
- Positive at confidence=1.0: +0.050
- Positive at confidence=0.5: +0.025
- Negative at confidence=0.8: -0.040

---

## 📊 Analytics & Monitoring

### Key Metrics

1. **Total Feedback Count** - Volume of feedback received
2. **Sentiment Distribution** - Positive/Negative/Neutral breakdown
3. **Agent Type Distribution** - Human/AI/Automated mix
4. **Average Confidence** - Mean confidence across all feedback
5. **Top Attributes** - Most-rated attributes
6. **Sentiment Trends** - Changes over time

### SQL Queries

**Feedback by Sentiment:**
```sql
SELECT feedback, COUNT(*) as count, AVG(confidence) as avg_conf
FROM rag_feedback
GROUP BY feedback;
```

**Top Attributes by Feedback:**
```sql
SELECT attribute_code, COUNT(*) as feedback_count
FROM rag_feedback
WHERE attribute_code IS NOT NULL
GROUP BY attribute_code
ORDER BY feedback_count DESC
LIMIT 10;
```

**Relevance Score Changes:**
```sql
SELECT attribute_code, document_code,
       relevance_score,
       (SELECT COUNT(*) FROM rag_feedback rf
        WHERE rf.attribute_code = adl.attribute_code
          AND rf.feedback = 'positive') as positive_votes,
       (SELECT COUNT(*) FROM rag_feedback rf
        WHERE rf.attribute_code = adl.attribute_code
          AND rf.feedback = 'negative') as negative_votes
FROM kyc_attr_doc_links adl
ORDER BY relevance_score DESC
LIMIT 10;
```

---

## 🎯 Use Cases

### 1. AI Agent Context Improvement

**Scenario**: Claude searches for "company ownership" but gets irrelevant attributes

```bash
# AI agent submits negative feedback
curl -X POST http://localhost:8080/rag/feedback \
  -d '{
    "query_text": "company ownership",
    "attribute_code": "COMPANY_NAME",
    "feedback": "negative",
    "confidence": 0.7,
    "agent_type": "ai",
    "agent_name": "claude-3"
  }'

# System reduces relevance score for COMPANY_NAME in ownership contexts
```

---

### 2. Human Expert Validation

**Scenario**: Compliance officer validates that UBO_NAME is correct for "beneficial owner" queries

```bash
curl -X POST http://localhost:8080/rag/feedback \
  -d '{
    "query_text": "beneficial owner identification",
    "attribute_code": "UBO_NAME",
    "document_code": "UBO_DECLARATION",
    "regulation_code": "AMLD5",
    "feedback": "positive",
    "confidence": 1.0,
    "agent_type": "human",
    "agent_name": "jane_doe"
  }'

# System increases relevance score with high confidence
```

---

### 3. A/B Testing & Experimentation

**Scenario**: Automated system tests different attribute recommendations

```bash
# Test variant A
curl -X POST http://localhost:8080/rag/feedback \
  -d '{
    "query_text": "tax reporting",
    "attribute_code": "TAX_RESIDENCY_COUNTRY",
    "feedback": "positive",
    "confidence": 0.4,
    "agent_type": "automated",
    "agent_name": "ab_test_variant_a"
  }'
```

---

## 🔒 Best Practices

### 1. Confidence Calibration

- **Human experts**: 0.9-1.0
- **AI agents**: 0.6-0.9 (calibrate based on model performance)
- **Automated systems**: 0.3-0.6
- **Exploratory signals**: < 0.3

### 2. Agent Naming

Use descriptive agent names for auditability:
```
✅ Good: "claude-3-opus", "compliance_team_lead", "relevance_validator_v2"
❌ Bad: "ai", "user", "system"
```

### 3. Feedback Granularity

Provide feedback at the most specific level:
```
✅ Ideal: attribute + document + regulation
✅ Good: attribute + document
✅ Acceptable: attribute only
❌ Avoid: No entity codes (just query_text)
```

### 4. Feedback Frequency

- Don't flood the system with duplicate feedback
- Aggregate similar feedback before submission
- Use higher confidence for repeated confirmations

### 5. Error Handling

Always check response status:
```bash
response=$(curl -X POST ... | jq -r '.status')
if [ "$response" != "ok" ]; then
    echo "Feedback failed"
fi
```

---

## 🔮 Future Enhancements

### Phase 2: Advanced Learning

| Feature | Description | Priority |
|---------|-------------|----------|
| **Feedback Clustering** | Aggregate votes by query term for concept-level learning | High |
| **Weighted Agent Types** | Different base weights for human vs AI vs automated | Medium |
| **Feedback Decay** | Reduce impact of old feedback over time | Medium |
| **Explainability** | Link feedback events to retrieved records for audit | High |
| **Feedback Dashboard** | Real-time visualization of trends and accuracy | Low |

---

### Phase 3: Intelligence Layer

- **Synonym Learning**: Automatically map natural language to attribute codes
- **Query Expansion**: Use feedback to suggest related attributes
- **Conflict Resolution**: Handle contradictory feedback intelligently
- **Personalization**: Per-user or per-agent relevance tuning

---

## 🐛 Troubleshooting

### Issue: Feedback not affecting relevance scores

**Check trigger is active:**
```sql
SELECT tgname, tgenabled FROM pg_trigger
WHERE tgname = 'trig_feedback_relevance';
```

**Verify trigger function exists:**
```sql
\df update_relevance
```

**Check relevance scores before/after:**
```sql
SELECT * FROM kyc_attr_doc_links WHERE attribute_code='UBO_NAME';
```

---

### Issue: Duplicate feedback entries

**Solution**: Add uniqueness constraint (optional)
```sql
CREATE UNIQUE INDEX idx_unique_feedback
ON rag_feedback(query_text, attribute_code, agent_name, DATE(created_at))
WHERE attribute_code IS NOT NULL;
```

---

### Issue: Relevance score out of bounds

**Trigger already clamps to [0.0, 1.0]:**
```sql
GREATEST(0.0, LEAST(1.0, new_score))
```

**Verify:**
```sql
SELECT MIN(relevance_score), MAX(relevance_score)
FROM kyc_attr_doc_links;
```

---

## 📚 Related Documentation

- [RAG_VECTOR_SEARCH.md](RAG_VECTOR_SEARCH.md) - Vector search implementation
- [RAG_QUICKSTART.md](RAG_QUICKSTART.md) - Quick start guide
- [API_DOCUMENTATION.md](API_DOCUMENTATION.md) - Complete API reference
- [REGULATORY_ONTOLOGY.md](REGULATORY_ONTOLOGY.md) - Data ontology

---

## 📞 Support

For questions or issues:

1. Check test suite: `./scripts/test_feedback.sh`
2. Review analytics: `curl http://localhost:8080/rag/feedback/analytics`
3. Inspect database: `psql -d kyc_dsl -c "SELECT * FROM rag_feedback ORDER BY created_at DESC LIMIT 10;"`

---

**Last Updated**: 2024  
**Version**: 1.5  
**Status**: Production Ready