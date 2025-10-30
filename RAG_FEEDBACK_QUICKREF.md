# RAG Feedback Loop - Quick Reference

**Version**: 1.5 | **Status**: Production Ready

---

## 🚀 Quick Start

```bash
# 1. Apply migration
./scripts/migrate_feedback.sh

# 2. Start server
go run cmd/kycserver/main.go

# 3. Submit feedback
curl -X POST http://localhost:8080/rag/feedback \
  -H "Content-Type: application/json" \
  -d '{"query_text":"tax requirements","attribute_code":"TAX_RESIDENCY_COUNTRY","feedback":"positive","confidence":0.9,"agent_type":"human"}'

# 4. View analytics
curl http://localhost:8080/rag/feedback/analytics
```

---

## 📌 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/rag/feedback` | Submit feedback |
| GET | `/rag/feedback/recent?limit=N` | Recent feedback entries |
| GET | `/rag/feedback/analytics?top=N` | Comprehensive analytics |
| GET | `/rag/feedback/attribute/{code}` | Feedback for attribute |
| GET | `/rag/feedback/summary` | Aggregated summary |

---

## 📝 Request Format

```json
{
  "query_text": "beneficial owner name",     // Required: original search query
  "attribute_code": "UBO_NAME",              // Optional: attribute being rated
  "document_code": "W8BEN",                  // Optional: document being rated
  "regulation_code": "AMLD5",                // Optional: regulation being rated
  "feedback": "positive",                    // positive | negative | neutral (default: positive)
  "confidence": 0.9,                         // 0.0-1.0 weight (default: 1.0)
  "agent_name": "jane_doe",                  // Optional: feedback provider name
  "agent_type": "human"                      // human | ai | automated (default: human)
}
```

**Note**: At least one of `attribute_code`, `document_code`, or `regulation_code` must be provided.

---

## ✅ Response Format

```json
{
  "status": "ok",
  "id": 123,
  "feedback": "positive",
  "agent_name": "jane_doe",
  "created_at": "2024-10-30T09:41:05Z"
}
```

---

## 🤖 Agent Types

| Type | Use Case | Typical Confidence | Impact |
|------|----------|-------------------|--------|
| **human** | Manual review, expert judgment | 0.8-1.0 | High |
| **ai** | Claude, GPT-4, ML models | 0.6-0.9 | Medium |
| **automated** | Rule-based, heuristics, A/B tests | 0.3-0.6 | Low |

---

## 📊 Feedback Sentiments

| Sentiment | Effect | Score Adjustment |
|-----------|--------|------------------|
| **positive** | Increase relevance | +0.05 × confidence |
| **negative** | Decrease relevance | -0.05 × confidence |
| **neutral** | No change | 0.00 |

**Examples**:
- Positive @ confidence=1.0 → +0.050
- Positive @ confidence=0.5 → +0.025
- Negative @ confidence=0.8 → -0.040

---

## 💡 Common Use Cases

### 1. Human Expert Validation
```bash
curl -X POST http://localhost:8080/rag/feedback \
  -d '{"query_text":"beneficial owner","attribute_code":"UBO_NAME","feedback":"positive","confidence":1.0,"agent_type":"human","agent_name":"compliance_officer"}'
```

### 2. AI Agent Evaluation
```bash
curl -X POST http://localhost:8080/rag/feedback \
  -d '{"query_text":"tax compliance","attribute_code":"TAX_RESIDENCY_COUNTRY","feedback":"positive","confidence":0.75,"agent_type":"ai","agent_name":"claude-3"}'
```

### 3. Automated Testing
```bash
curl -X POST http://localhost:8080/rag/feedback \
  -d '{"query_text":"ownership structure","attribute_code":"UBO_PERCENTAGE","feedback":"positive","confidence":0.4,"agent_type":"automated","agent_name":"ab_test_v1"}'
```

---

## 🔍 Analytics Queries

### Get Recent Feedback
```bash
curl "http://localhost:8080/rag/feedback/recent?limit=10"
```

### Get Analytics Dashboard
```bash
curl "http://localhost:8080/rag/feedback/analytics?top=20" | jq
```

### Get Feedback for Specific Attribute
```bash
curl "http://localhost:8080/rag/feedback/attribute/UBO_NAME" | jq
```

### Get Summary Statistics
```bash
curl "http://localhost:8080/rag/feedback/summary" | jq
```

---

## 🗄️ Database Queries

### View All Feedback
```sql
SELECT * FROM rag_feedback ORDER BY created_at DESC LIMIT 10;
```

### Check Relevance Scores
```sql
SELECT attribute_code, document_code, relevance_score
FROM kyc_attr_doc_links
WHERE attribute_code = 'UBO_NAME'
ORDER BY relevance_score DESC;
```

### Feedback by Sentiment
```sql
SELECT feedback, COUNT(*) as count, AVG(confidence) as avg_conf
FROM rag_feedback
GROUP BY feedback;
```

### Top Attributes by Feedback
```sql
SELECT attribute_code, COUNT(*) as feedback_count
FROM rag_feedback
WHERE attribute_code IS NOT NULL
GROUP BY attribute_code
ORDER BY feedback_count DESC
LIMIT 10;
```

### Agent Performance
```sql
SELECT agent_type, agent_name, COUNT(*) as total, AVG(confidence) as avg_conf
FROM rag_feedback
GROUP BY agent_type, agent_name
ORDER BY total DESC;
```

---

## ⚙️ Configuration

### Environment Variables
```bash
export OPENAI_API_KEY="sk-..."    # Required for embeddings
export PGDATABASE="kyc_dsl"       # Database name
export PGHOST="localhost"          # Database host
export PGPORT="5432"               # Database port
export PORT="8080"                 # API server port
```

---

## 🧪 Testing

### Run Full Test Suite
```bash
./scripts/test_feedback.sh
```

### Run Example Workflow
```bash
./scripts/example_feedback_workflow.sh
```

### Manual Test
```bash
# Submit feedback
curl -X POST http://localhost:8080/rag/feedback \
  -H "Content-Type: application/json" \
  -d '{"query_text":"test query","attribute_code":"UBO_NAME","feedback":"positive","confidence":0.9,"agent_type":"human"}'

# Verify in database
psql -d kyc_dsl -c "SELECT * FROM rag_feedback ORDER BY created_at DESC LIMIT 1;"
```

---

## 🐛 Troubleshooting

### Server Not Running
```bash
# Check if server is up
curl http://localhost:8080/rag/health

# Start server
go run cmd/kycserver/main.go
```

### Migration Not Applied
```bash
# Check if table exists
psql -d kyc_dsl -c "\dt rag_feedback"

# Apply migration
./scripts/migrate_feedback.sh
```

### Trigger Not Firing
```sql
-- Check trigger status
SELECT tgname, tgenabled FROM pg_trigger WHERE tgname = 'trig_feedback_relevance';

-- Verify function exists
\df update_relevance
```

### Invalid Confidence Value
```
Error: "confidence must be between 0 and 1"
Fix: Use confidence values in range [0.0, 1.0]
```

### Missing Query Text
```
Error: "query_text is required"
Fix: Always include query_text in feedback submissions
```

---

## 📐 Learning Formula

```
new_score = CLAMP(current_score + adjustment, 0.0, 1.0)

where:
  adjustment = base_delta × confidence
  base_delta = +0.05 (positive) | -0.05 (negative) | 0.00 (neutral)
  confidence ∈ [0.0, 1.0]
```

**Example Calculations**:

| Current | Feedback | Confidence | Adjustment | New Score |
|---------|----------|------------|------------|-----------|
| 0.50 | positive | 1.0 | +0.05 | 0.55 |
| 0.50 | positive | 0.5 | +0.025 | 0.525 |
| 0.50 | negative | 0.8 | -0.04 | 0.46 |
| 0.98 | positive | 1.0 | +0.02* | 1.00 (clamped) |
| 0.03 | negative | 1.0 | -0.03* | 0.00 (clamped) |

\*Clamped to [0.0, 1.0] range

---

## 🔗 Related Documentation

- [RAG_FEEDBACK.md](RAG_FEEDBACK.md) - Complete documentation
- [RAG_VECTOR_SEARCH.md](RAG_VECTOR_SEARCH.md) - Vector search details
- [RAG_QUICKSTART.md](RAG_QUICKSTART.md) - RAG quick start
- [API_DOCUMENTATION.md](API_DOCUMENTATION.md) - Full API reference

---

## 📞 Quick Help

```bash
# View API documentation
open http://localhost:8080/

# Check server health
curl http://localhost:8080/rag/health

# Get metadata stats
curl http://localhost:8080/rag/stats

# Run semantic search
curl "http://localhost:8080/rag/attribute_search?q=tax%20requirements"
```

---

**Last Updated**: 2024 | **Version**: 1.5 | **Status**: Production Ready