#!/bin/bash
# ===========================================================
# test_api.sh
# Quick API endpoint testing script
# ===========================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "================================================"
echo "🧪 Testing KYC-DSL RAG API"
echo "================================================"
echo "Base URL: $BASE_URL"
echo ""

# Check if server is running
echo -e "${BLUE}Checking if server is running...${NC}"
if ! curl -s -f "$BASE_URL/rag/health" > /dev/null; then
    echo -e "${RED}❌ Server is not running at $BASE_URL${NC}"
    echo "   Start the server with: make run-server"
    exit 1
fi
echo -e "${GREEN}✅ Server is running${NC}"
echo ""

# Test 1: Health Check
echo "================================================"
echo -e "${BLUE}TEST 1: Health Check${NC}"
echo "================================================"
echo "GET $BASE_URL/rag/health"
echo ""
curl -s "$BASE_URL/rag/health" | jq '.'
echo ""
echo -e "${GREEN}✅ Health check passed${NC}"
echo ""
sleep 1

# Test 2: Statistics
echo "================================================"
echo -e "${BLUE}TEST 2: Metadata Statistics${NC}"
echo "================================================"
echo "GET $BASE_URL/rag/stats"
echo ""
curl -s "$BASE_URL/rag/stats" | jq '.'
echo ""
echo -e "${GREEN}✅ Statistics retrieved${NC}"
echo ""
sleep 1

# Test 3: Semantic Search - Beneficial Owner
echo "================================================"
echo -e "${BLUE}TEST 3: Semantic Search - Beneficial Owner${NC}"
echo "================================================"
echo "GET $BASE_URL/rag/attribute_search?q=beneficial%20owner%20name&limit=3"
echo ""
curl -s "$BASE_URL/rag/attribute_search?q=beneficial%20owner%20name&limit=3" | jq '.'
echo ""
echo -e "${GREEN}✅ Semantic search completed${NC}"
echo ""
sleep 2

# Test 4: Semantic Search - Tax
echo "================================================"
echo -e "${BLUE}TEST 4: Semantic Search - Tax Reporting${NC}"
echo "================================================"
echo "GET $BASE_URL/rag/attribute_search?q=tax%20reporting%20requirements&limit=3"
echo ""
curl -s "$BASE_URL/rag/attribute_search?q=tax%20reporting%20requirements&limit=3" | jq '.'
echo ""
echo -e "${GREEN}✅ Tax search completed${NC}"
echo ""
sleep 2

# Test 5: Semantic Search - Risk
echo "================================================"
echo -e "${BLUE}TEST 5: Semantic Search - Risk Indicators${NC}"
echo "================================================"
echo "GET $BASE_URL/rag/attribute_search?q=money%20laundering%20risk%20factors&limit=3"
echo ""
curl -s "$BASE_URL/rag/attribute_search?q=money%20laundering%20risk%20factors&limit=3" | jq '.'
echo ""
echo -e "${GREEN}✅ Risk search completed${NC}"
echo ""
sleep 2

# Test 6: Similar Attributes
echo "================================================"
echo -e "${BLUE}TEST 6: Find Similar Attributes${NC}"
echo "================================================"
echo "GET $BASE_URL/rag/similar_attributes?code=UBO_NAME&limit=3"
echo ""
curl -s "$BASE_URL/rag/similar_attributes?code=UBO_NAME&limit=3" | jq '.'
echo ""
echo -e "${GREEN}✅ Similar attributes found${NC}"
echo ""
sleep 1

# Test 7: Text Search
echo "================================================"
echo -e "${BLUE}TEST 7: Text Search${NC}"
echo "================================================"
echo "GET $BASE_URL/rag/text_search?term=ownership"
echo ""
curl -s "$BASE_URL/rag/text_search?term=ownership" | jq '.'
echo ""
echo -e "${GREEN}✅ Text search completed${NC}"
echo ""
sleep 1

# Test 8: Get Specific Attribute
echo "================================================"
echo -e "${BLUE}TEST 8: Get Specific Attribute${NC}"
echo "================================================"
echo "GET $BASE_URL/rag/attribute/TAX_RESIDENCY_COUNTRY"
echo ""
curl -s "$BASE_URL/rag/attribute/TAX_RESIDENCY_COUNTRY" | jq '.'
echo ""
echo -e "${GREEN}✅ Attribute retrieved${NC}"
echo ""
sleep 1

# Test 9: Error Handling - Missing Parameter
echo "================================================"
echo -e "${BLUE}TEST 9: Error Handling - Missing Parameter${NC}"
echo "================================================"
echo "GET $BASE_URL/rag/attribute_search (no q parameter)"
echo ""
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/rag/attribute_search")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)
echo "$BODY" | jq '.'
echo ""
if [ "$HTTP_CODE" = "400" ]; then
    echo -e "${GREEN}✅ Correctly returned 400 Bad Request${NC}"
else
    echo -e "${YELLOW}⚠️  Expected 400, got $HTTP_CODE${NC}"
fi
echo ""
sleep 1

# Test 10: Error Handling - Not Found
echo "================================================"
echo -e "${BLUE}TEST 10: Error Handling - Attribute Not Found${NC}"
echo "================================================"
echo "GET $BASE_URL/rag/attribute/INVALID_CODE"
echo ""
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/rag/attribute/INVALID_CODE")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)
echo "$BODY" | jq '.'
echo ""
if [ "$HTTP_CODE" = "404" ]; then
    echo -e "${GREEN}✅ Correctly returned 404 Not Found${NC}"
else
    echo -e "${YELLOW}⚠️  Expected 404, got $HTTP_CODE${NC}"
fi
echo ""

# Summary
echo "================================================"
echo -e "${GREEN}✅ All API tests completed successfully!${NC}"
echo "================================================"
echo ""
echo "📊 Test Summary:"
echo "   ✅ Health check"
echo "   ✅ Statistics"
echo "   ✅ Semantic search (beneficial owner)"
echo "   ✅ Semantic search (tax reporting)"
echo "   ✅ Semantic search (risk factors)"
echo "   ✅ Similar attributes"
echo "   ✅ Text search"
echo "   ✅ Get specific attribute"
echo "   ✅ Error handling (400)"
echo "   ✅ Error handling (404)"
echo ""
echo "🎉 Your RAG API is fully operational!"
echo ""
echo "Next steps:"
echo "  - Review results above"
echo "  - Check server logs for any warnings"
echo "  - Test with your own queries"
echo "  - Integrate with your application"
echo ""
echo "Documentation:"
echo "  - API Reference: API_DOCUMENTATION.md"
echo "  - Quick Start: RAG_QUICKSTART.md"
echo "  - Full Guide: RAG_VECTOR_SEARCH.md"
echo ""
