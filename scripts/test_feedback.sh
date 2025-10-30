#!/bin/bash

# ===========================================================
# test_feedback.sh
# Comprehensive test script for RAG Feedback Loop
# ===========================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE="http://localhost:8080"
DELAY=1  # Delay between requests in seconds

# Helper functions
print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Check if server is running
check_server() {
    print_header "🔍 Checking Server Status"
    if curl -s "$API_BASE/rag/health" > /dev/null 2>&1; then
        print_success "Server is running at $API_BASE"
    else
        print_error "Server is not running at $API_BASE"
        echo "Please start the server with: go run cmd/kycserver/main.go"
        exit 1
    fi
}

# Get initial relevance score
get_relevance_score() {
    local attr_code=$1
    local doc_code=$2

    print_header "📊 Initial Relevance Score"
    echo "Querying: attribute_code='$attr_code', document_code='$doc_code'"

    psql -d kyc_dsl -t -c "
        SELECT attribute_code, document_code, relevance_score
        FROM kyc_attr_doc_links
        WHERE attribute_code='$attr_code' AND document_code='$doc_code'
        LIMIT 1;
    " 2>/dev/null || echo "Note: Install psql to see relevance scores"
}

# Test 1: Submit positive feedback (human)
test_positive_feedback_human() {
    print_header "🧪 Test 1: Positive Feedback (Human Agent)"

    local response=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "query_text": "beneficial owner name",
            "attribute_code": "UBO_NAME",
            "document_code": "W8BEN",
            "feedback": "positive",
            "confidence": 0.9,
            "agent_name": "test_user",
            "agent_type": "human"
        }')

    if echo "$response" | grep -q '"status":"ok"'; then
        print_success "Positive feedback submitted successfully"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Failed to submit positive feedback"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 2: Submit negative feedback (AI agent)
test_negative_feedback_ai() {
    print_header "🧪 Test 2: Negative Feedback (AI Agent)"

    local response=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "query_text": "tax reporting requirements",
            "attribute_code": "TAX_RESIDENCY_COUNTRY",
            "document_code": "W9",
            "feedback": "negative",
            "confidence": 0.7,
            "agent_name": "claude_ai",
            "agent_type": "ai"
        }')

    if echo "$response" | grep -q '"status":"ok"'; then
        print_success "Negative feedback submitted successfully"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Failed to submit negative feedback"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 3: Submit neutral feedback (automated)
test_neutral_feedback_automated() {
    print_header "🧪 Test 3: Neutral Feedback (Automated Agent)"

    local response=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "query_text": "ownership structure",
            "attribute_code": "UBO_PERCENTAGE",
            "regulation_code": "AMLD5",
            "feedback": "neutral",
            "confidence": 0.5,
            "agent_name": "auto_validator",
            "agent_type": "automated"
        }')

    if echo "$response" | grep -q '"status":"ok"'; then
        print_success "Neutral feedback submitted successfully"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Failed to submit neutral feedback"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 4: High confidence positive feedback
test_high_confidence_feedback() {
    print_header "🧪 Test 4: High Confidence Positive Feedback"

    local response=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "query_text": "politically exposed person status",
            "attribute_code": "PEP_STATUS",
            "document_code": "PEP_DECLARATION",
            "regulation_code": "AMLD5",
            "feedback": "positive",
            "confidence": 1.0,
            "agent_name": "compliance_officer",
            "agent_type": "human"
        }')

    if echo "$response" | grep -q '"status":"ok"'; then
        print_success "High confidence feedback submitted successfully"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Failed to submit high confidence feedback"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 5: Low confidence feedback
test_low_confidence_feedback() {
    print_header "🧪 Test 5: Low Confidence Feedback"

    local response=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "query_text": "sanctions screening",
            "attribute_code": "SANCTIONS_SCREENING_STATUS",
            "feedback": "positive",
            "confidence": 0.3,
            "agent_type": "ai"
        }')

    if echo "$response" | grep -q '"status":"ok"'; then
        print_success "Low confidence feedback submitted successfully"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Failed to submit low confidence feedback"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 6: Get recent feedback
test_get_recent_feedback() {
    print_header "🧪 Test 6: Get Recent Feedback"

    local response=$(curl -s "$API_BASE/rag/feedback/recent?limit=10")

    if echo "$response" | grep -q '"count"'; then
        print_success "Retrieved recent feedback"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Failed to get recent feedback"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 7: Get feedback analytics
test_get_analytics() {
    print_header "🧪 Test 7: Get Feedback Analytics"

    local response=$(curl -s "$API_BASE/rag/feedback/analytics?top=5")

    if echo "$response" | grep -q '"total_feedback"'; then
        print_success "Retrieved feedback analytics"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Failed to get feedback analytics"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 8: Get feedback by attribute
test_get_feedback_by_attribute() {
    print_header "🧪 Test 8: Get Feedback by Attribute (UBO_NAME)"

    local response=$(curl -s "$API_BASE/rag/feedback/attribute/UBO_NAME")

    if echo "$response" | grep -q '"count"'; then
        print_success "Retrieved feedback for UBO_NAME"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Failed to get feedback by attribute"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 9: Get feedback summary
test_get_summary() {
    print_header "🧪 Test 9: Get Feedback Summary"

    local response=$(curl -s "$API_BASE/rag/feedback/summary?limit=10")

    if echo "$response" | grep -q '"overall_summary"'; then
        print_success "Retrieved feedback summary"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Failed to get feedback summary"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 10: Verify learning effect
test_learning_effect() {
    print_header "🧪 Test 10: Verify Learning Effect"

    print_info "Checking relevance score after feedback..."

    psql -d kyc_dsl -t -c "
        SELECT
            attribute_code,
            document_code,
            relevance_score,
            CASE
                WHEN relevance_score > 0.5 THEN 'INCREASED ✓'
                WHEN relevance_score < 0.5 THEN 'DECREASED ↓'
                ELSE 'UNCHANGED'
            END as status
        FROM kyc_attr_doc_links
        WHERE attribute_code='UBO_NAME' AND document_code='W8BEN'
        LIMIT 1;
    " 2>/dev/null || echo "Note: Install psql to see relevance scores"

    print_info "Expected: Score should have increased due to positive feedback"
}

# Test 11: Error handling - missing required fields
test_error_missing_query() {
    print_header "🧪 Test 11: Error Handling - Missing Query Text"

    local response=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "attribute_code": "UBO_NAME",
            "feedback": "positive"
        }')

    if echo "$response" | grep -q "query_text is required"; then
        print_success "Correctly rejected missing query_text"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Did not properly validate missing query_text"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 12: Error handling - invalid confidence
test_error_invalid_confidence() {
    print_header "🧪 Test 12: Error Handling - Invalid Confidence"

    local response=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "query_text": "test query",
            "attribute_code": "UBO_NAME",
            "feedback": "positive",
            "confidence": 1.5
        }')

    if echo "$response" | grep -q "confidence must be between 0 and 1"; then
        print_success "Correctly rejected invalid confidence value"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_error "Did not properly validate confidence range"
        echo "$response"
    fi

    sleep $DELAY
}

# Test 13: Batch feedback submission
test_batch_feedback() {
    print_header "🧪 Test 13: Batch Feedback Submission"

    local count=0
    local success=0

    for i in {1..5}; do
        count=$((count + 1))
        response=$(curl -s -X POST "$API_BASE/rag/feedback" \
            -H "Content-Type: application/json" \
            -d "{
                \"query_text\": \"test query $i\",
                \"attribute_code\": \"TAX_RESIDENCY_COUNTRY\",
                \"feedback\": \"positive\",
                \"confidence\": 0.8,
                \"agent_type\": \"automated\"
            }")

        if echo "$response" | grep -q '"status":"ok"'; then
            success=$((success + 1))
        fi
        sleep 0.5
    done

    print_success "Batch submission: $success/$count succeeded"
}

# Main execution
main() {
    print_header "🚀 RAG Feedback Loop Test Suite"
    echo "Testing feedback system with comprehensive scenarios..."
    echo ""

    # Check prerequisites
    check_server

    # Get initial state
    get_relevance_score "UBO_NAME" "W8BEN"

    # Run all tests
    test_positive_feedback_human
    test_negative_feedback_ai
    test_neutral_feedback_automated
    test_high_confidence_feedback
    test_low_confidence_feedback
    test_get_recent_feedback
    test_get_analytics
    test_get_feedback_by_attribute
    test_get_summary
    test_learning_effect
    test_error_missing_query
    test_error_invalid_confidence
    test_batch_feedback

    # Final summary
    print_header "✅ Test Suite Complete"
    echo "All feedback tests have been executed."
    echo ""
    echo "Next steps:"
    echo "  1. Check feedback analytics: curl $API_BASE/rag/feedback/analytics"
    echo "  2. View recent feedback: curl $API_BASE/rag/feedback/recent"
    echo "  3. Check relevance scores: psql -d kyc_dsl -c 'SELECT * FROM kyc_attr_doc_links LIMIT 10;'"
    echo ""
}

# Run main function
main
