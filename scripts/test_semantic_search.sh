#!/bin/bash
# ===========================================================
# test_semantic_search.sh
# Test script for RAG & Vector Search functionality
# ===========================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY="$PROJECT_DIR/bin/kycctl"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "================================================"
echo "🧪 Testing RAG & Vector Search Functionality"
echo "================================================"
echo ""

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo -e "${YELLOW}⚠️  Binary not found, building...${NC}"
    cd "$PROJECT_DIR"
    make build
    echo ""
fi

# Check for OPENAI_API_KEY
if [ -z "$OPENAI_API_KEY" ]; then
    echo -e "${RED}❌ Error: OPENAI_API_KEY environment variable not set${NC}"
    echo "   Please set your OpenAI API key:"
    echo "   export OPENAI_API_KEY='sk-...'"
    exit 1
fi

# Check database connection
echo -e "${BLUE}1. Checking database connection...${NC}"
if ! psql -d kyc_dsl -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${RED}❌ Cannot connect to database 'kyc_dsl'${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Database connection successful${NC}"
echo ""

# Check pgvector extension
echo -e "${BLUE}2. Checking pgvector extension...${NC}"
if ! psql -d kyc_dsl -c "SELECT * FROM pg_extension WHERE extname='vector'" | grep -q vector; then
    echo -e "${YELLOW}⚠️  pgvector extension not found, installing...${NC}"
    psql -d kyc_dsl -c "CREATE EXTENSION IF NOT EXISTS vector"
fi
echo -e "${GREEN}✅ pgvector extension enabled${NC}"
echo ""

# Seed metadata (if not already seeded)
echo -e "${BLUE}3. Seeding attribute metadata...${NC}"
EMBEDDING_COUNT=$(psql -d kyc_dsl -t -c "SELECT COUNT(*) FROM kyc_attribute_metadata WHERE embedding IS NOT NULL" 2>/dev/null || echo "0")
EMBEDDING_COUNT=$(echo $EMBEDDING_COUNT | tr -d ' ')

if [ "$EMBEDDING_COUNT" -lt 5 ]; then
    echo -e "${YELLOW}   Generating embeddings for attributes...${NC}"
    "$BINARY" seed-metadata
    echo ""
else
    echo -e "${GREEN}✅ Already seeded ($EMBEDDING_COUNT attributes with embeddings)${NC}"
    echo ""
fi

# Test 1: Metadata Statistics
echo "================================================"
echo -e "${BLUE}TEST 1: Metadata Statistics${NC}"
echo "================================================"
"$BINARY" metadata-stats
echo ""
sleep 1

# Test 2: Semantic Search - Tax Compliance
echo "================================================"
echo -e "${BLUE}TEST 2: Semantic Search - Tax Compliance${NC}"
echo "================================================"
"$BINARY" search-metadata "tax reporting requirements" --limit=5
echo ""
sleep 1

# Test 3: Semantic Search - Ownership
echo "================================================"
echo -e "${BLUE}TEST 3: Semantic Search - Beneficial Ownership${NC}"
echo "================================================"
"$BINARY" search-metadata "who owns this company" --limit=5
echo ""
sleep 1

# Test 4: Semantic Search - Risk Assessment
echo "================================================"
echo -e "${BLUE}TEST 4: Semantic Search - Risk Assessment${NC}"
echo "================================================"
"$BINARY" search-metadata "money laundering risk factors" --limit=5
echo ""
sleep 1

# Test 5: Find Similar Attributes
echo "================================================"
echo -e "${BLUE}TEST 5: Find Similar Attributes to UBO_NAME${NC}"
echo "================================================"
"$BINARY" similar-attributes UBO_NAME --limit=5
echo ""
sleep 1

# Test 6: Text Search
echo "================================================"
echo -e "${BLUE}TEST 6: Text Search - PEP${NC}"
echo "================================================"
"$BINARY" text-search "PEP"
echo ""
sleep 1

# Test 7: Direct SQL Vector Query
echo "================================================"
echo -e "${BLUE}TEST 7: Direct SQL Vector Query${NC}"
echo "================================================"
echo "Finding attributes similar to TAX_RESIDENCY_COUNTRY..."
psql -d kyc_dsl -c "
SELECT
    attribute_code,
    risk_level,
    ROUND((1 - (embedding <=> (
        SELECT embedding
        FROM kyc_attribute_metadata
        WHERE attribute_code = 'TAX_RESIDENCY_COUNTRY'
    )))::numeric, 4) as similarity
FROM kyc_attribute_metadata
WHERE attribute_code != 'TAX_RESIDENCY_COUNTRY'
  AND embedding IS NOT NULL
ORDER BY embedding <=> (
    SELECT embedding
    FROM kyc_attribute_metadata
    WHERE attribute_code = 'TAX_RESIDENCY_COUNTRY'
)
LIMIT 5;
"
echo ""

# Test 8: Embedding Coverage Report
echo "================================================"
echo -e "${BLUE}TEST 8: Embedding Coverage by Risk Level${NC}"
echo "================================================"
psql -d kyc_dsl -c "
SELECT
    COALESCE(risk_level, 'NULL') as risk_level,
    COUNT(*) as total,
    COUNT(embedding) as with_embedding,
    ROUND(100.0 * COUNT(embedding) / COUNT(*), 1) as coverage_pct
FROM kyc_attribute_metadata
GROUP BY risk_level
ORDER BY coverage_pct DESC;
"
echo ""

# Test 9: Cluster Detection
echo "================================================"
echo -e "${BLUE}TEST 9: Detecting Attribute Clusters${NC}"
echo "================================================"
echo "Finding attributes with many similar neighbors (>0.75 similarity)..."
psql -d kyc_dsl -c "
WITH similarity_counts AS (
    SELECT
        a1.attribute_code,
        COUNT(*) as similar_count
    FROM kyc_attribute_metadata a1
    CROSS JOIN kyc_attribute_metadata a2
    WHERE a1.attribute_code != a2.attribute_code
      AND a1.embedding IS NOT NULL
      AND a2.embedding IS NOT NULL
      AND 1 - (a1.embedding <=> a2.embedding) > 0.75
    GROUP BY a1.attribute_code
)
SELECT
    attribute_code,
    similar_count as neighbors
FROM similarity_counts
WHERE similar_count > 0
ORDER BY similar_count DESC
LIMIT 10;
"
echo ""

# Test 10: Synonym Resolution
echo "================================================"
echo -e "${BLUE}TEST 10: Synonym Resolution${NC}"
echo "================================================"
echo "Searching for 'Company Name' (should find REGISTERED_NAME)..."
"$BINARY" text-search "Company Name"
echo ""

# Summary
echo "================================================"
echo -e "${GREEN}✅ All tests completed successfully!${NC}"
echo "================================================"
echo ""
echo "🎉 RAG & Vector Search System is operational!"
echo ""
echo "Next steps:"
echo "  1. Try custom semantic queries:"
echo "     ./kycctl search-metadata \"your query here\""
echo ""
echo "  2. Find similar attributes:"
echo "     ./kycctl similar-attributes YOUR_ATTRIBUTE_CODE"
echo ""
echo "  3. Integrate with AI agents:"
echo "     - Use SearchByVector() for semantic retrieval"
echo "     - Build context from RegulatoryCitations"
echo "     - Ground LLM responses in regulatory data"
echo ""
echo "📖 See RAG_VECTOR_SEARCH.md for detailed documentation"
echo ""
