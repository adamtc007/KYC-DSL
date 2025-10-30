#!/bin/bash

# ===========================================================
# example_feedback_workflow.sh
# Complete RAG Feedback Loop Workflow Example
# ===========================================================
#
# This script demonstrates a complete feedback workflow:
# 1. Perform semantic search
# 2. Review results
# 3. Submit feedback
# 4. Verify learning effect
# 5. View analytics
#
# ===========================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
API_BASE="http://localhost:8080"
DB_NAME="${PGDATABASE:-kyc_dsl}"

# Helper functions
print_header() {
    echo -e "\n${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║ $1${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}\n"
}

print_step() {
    echo -e "${CYAN}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

print_result() {
    echo -e "${MAGENTA}$1${NC}"
}

pause() {
    echo ""
    read -p "Press Enter to continue..."
}

# Check if server is running
check_server() {
    if ! curl -s "$API_BASE/rag/health" > /dev/null 2>&1; then
        echo -e "${RED}✗ Server is not running at $API_BASE${NC}"
        echo "Please start the server with: go run cmd/kycserver/main.go"
        exit 1
    fi
}

# Main workflow
main() {
    clear
    print_header "RAG Feedback Loop - Complete Workflow Example"

    echo "This demonstration shows how the feedback loop enables continuous learning."
    echo "We'll search for attributes, submit feedback, and observe how the system improves."
    echo ""

    check_server
    print_success "Server is running at $API_BASE"
    pause

    # ============================================
    # STEP 1: Initial Search
    # ============================================
    print_header "STEP 1: Perform Semantic Search"
    print_step "Searching for: 'beneficial owner name'"

    search_response=$(curl -s "$API_BASE/rag/attribute_search?q=beneficial%20owner%20name&limit=3")

    echo ""
    echo "$search_response" | jq -r '.results[] | "  • \(.code) (similarity: \(.similarity_score | tonumber | . * 100 | round / 100))\n    \(.description)"' 2>/dev/null || echo "$search_response"

    pause

    # ============================================
    # STEP 2: Check Initial Relevance Score
    # ============================================
    print_header "STEP 2: Check Current Relevance Scores"
    print_step "Querying database for UBO_NAME relevance scores..."

    echo ""
    print_info "Current relevance scores for UBO_NAME:"
    psql -d "$DB_NAME" -c "
        SELECT
            attribute_code,
            document_code,
            ROUND(relevance_score::numeric, 4) as relevance_score
        FROM kyc_attr_doc_links
        WHERE attribute_code = 'UBO_NAME'
        ORDER BY relevance_score DESC
        LIMIT 5;
    " 2>/dev/null || print_info "Install psql to view database scores"

    pause

    # ============================================
    # STEP 3: Submit Positive Feedback (Human)
    # ============================================
    print_header "STEP 3: Submit Positive Feedback"
    print_step "A human compliance officer found UBO_NAME highly relevant..."

    echo ""
    print_info "Submitting positive feedback with high confidence (human agent)"

    feedback_response=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "query_text": "beneficial owner name",
            "attribute_code": "UBO_NAME",
            "document_code": "UBO_DECLARATION",
            "regulation_code": "AMLD5",
            "feedback": "positive",
            "confidence": 0.95,
            "agent_name": "compliance_officer_jane",
            "agent_type": "human"
        }')

    echo ""
    echo "$feedback_response" | jq '.' 2>/dev/null || echo "$feedback_response"
    print_success "Feedback submitted successfully!"

    pause

    # ============================================
    # STEP 4: Verify Learning Effect
    # ============================================
    print_header "STEP 4: Verify Learning Effect"
    print_step "Checking if relevance scores increased..."

    echo ""
    print_info "Updated relevance scores for UBO_NAME:"
    psql -d "$DB_NAME" -c "
        SELECT
            attribute_code,
            document_code,
            ROUND(relevance_score::numeric, 4) as relevance_score,
            CASE
                WHEN relevance_score > 0.5 THEN '📈 INCREASED'
                ELSE '📉 DECREASED'
            END as trend
        FROM kyc_attr_doc_links
        WHERE attribute_code = 'UBO_NAME'
        ORDER BY relevance_score DESC
        LIMIT 5;
    " 2>/dev/null || print_info "Install psql to view database scores"

    print_success "Relevance scores have been adjusted!"
    print_info "Expected: UBO_DECLARATION score increased by ~0.048 (0.05 × 0.95 confidence)"

    pause

    # ============================================
    # STEP 5: Submit AI Agent Feedback
    # ============================================
    print_header "STEP 5: AI Agent Provides Feedback"
    print_step "An AI agent (Claude) evaluates another search result..."

    echo ""
    print_info "Submitting feedback from AI agent (medium confidence)"

    ai_feedback=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "query_text": "company ownership percentage",
            "attribute_code": "UBO_PERCENTAGE",
            "document_code": "UBO_DECLARATION",
            "feedback": "positive",
            "confidence": 0.75,
            "agent_name": "claude-3-opus",
            "agent_type": "ai"
        }')

    echo ""
    echo "$ai_feedback" | jq '.' 2>/dev/null || echo "$ai_feedback"
    print_success "AI feedback submitted!"

    pause

    # ============================================
    # STEP 6: Submit Negative Feedback
    # ============================================
    print_header "STEP 6: Submit Negative Feedback"
    print_step "User finds an irrelevant result..."

    echo ""
    print_info "Submitting negative feedback (low relevance match)"

    neg_feedback=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "query_text": "tax reporting requirements",
            "attribute_code": "COMPANY_NAME",
            "document_code": "W8BEN",
            "feedback": "negative",
            "confidence": 0.8,
            "agent_name": "user_123",
            "agent_type": "human"
        }')

    echo ""
    echo "$neg_feedback" | jq '.' 2>/dev/null || echo "$neg_feedback"
    print_success "Negative feedback recorded!"
    print_info "System will reduce relevance score for this mapping"

    pause

    # ============================================
    # STEP 7: Submit Automated Feedback
    # ============================================
    print_header "STEP 7: Automated System Feedback"
    print_step "An automated validator provides feedback..."

    echo ""
    print_info "Submitting feedback from automated system (low confidence)"

    auto_feedback=$(curl -s -X POST "$API_BASE/rag/feedback" \
        -H "Content-Type: application/json" \
        -d '{
            "query_text": "sanctions screening",
            "attribute_code": "SANCTIONS_SCREENING_STATUS",
            "feedback": "positive",
            "confidence": 0.4,
            "agent_name": "relevance_validator_v2",
            "agent_type": "automated"
        }')

    echo ""
    echo "$auto_feedback" | jq '.' 2>/dev/null || echo "$auto_feedback"
    print_success "Automated feedback recorded!"
    print_info "Lower confidence = smaller impact on scores"

    pause

    # ============================================
    # STEP 8: View Recent Feedback
    # ============================================
    print_header "STEP 8: View Recent Feedback"
    print_step "Retrieving the last 5 feedback entries..."

    echo ""
    recent=$(curl -s "$API_BASE/rag/feedback/recent?limit=5")
    echo "$recent" | jq '.' 2>/dev/null || echo "$recent"

    pause

    # ============================================
    # STEP 9: View Analytics
    # ============================================
    print_header "STEP 9: View Feedback Analytics"
    print_step "Getting comprehensive analytics..."

    echo ""
    analytics=$(curl -s "$API_BASE/rag/feedback/analytics?top=5")
    echo "$analytics" | jq '{
        total_feedback,
        positive_count,
        negative_count,
        neutral_count,
        avg_confidence,
        by_agent_type,
        top_attributes: .top_attributes[0:3]
    }' 2>/dev/null || echo "$analytics"

    pause

    # ============================================
    # STEP 10: View Summary
    # ============================================
    print_header "STEP 10: Feedback Summary"
    print_step "Getting aggregated summary by sentiment and agent type..."

    echo ""
    summary=$(curl -s "$API_BASE/rag/feedback/summary")
    echo "$summary" | jq '.' 2>/dev/null || echo "$summary"

    pause

    # ============================================
    # STEP 11: Demonstrate Improvement
    # ============================================
    print_header "STEP 11: Demonstrate System Improvement"
    print_step "Performing the same search again..."

    echo ""
    print_info "Re-searching for: 'beneficial owner name'"

    improved_search=$(curl -s "$API_BASE/rag/attribute_search?q=beneficial%20owner%20name&limit=3")

    echo ""
    echo "$improved_search" | jq -r '.results[] | "  • \(.code) (similarity: \(.similarity_score | tonumber | . * 100 | round / 100))\n    \(.description)"' 2>/dev/null || echo "$improved_search"

    print_success "Results are now influenced by accumulated feedback!"

    pause

    # ============================================
    # Final Summary
    # ============================================
    print_header "🎉 Workflow Complete!"

    echo "You've just witnessed the RAG feedback loop in action:"
    echo ""
    echo "  ✅ Performed semantic searches"
    echo "  ✅ Submitted feedback from multiple agent types (human, AI, automated)"
    echo "  ✅ Observed real-time relevance score adjustments"
    echo "  ✅ Viewed analytics and trends"
    echo "  ✅ Demonstrated continuous learning"
    echo ""
    echo "Key Takeaways:"
    echo ""
    echo "  📊 Feedback automatically adjusts relevance scores"
    echo "  🤖 Different agent types have different confidence levels"
    echo "  📈 System continuously improves from usage patterns"
    echo "  🔍 Analytics provide insights into search quality"
    echo "  🔄 Self-correcting mechanism enables domain adaptation"
    echo ""

    print_info "Next Steps:"
    echo ""
    echo "  1. Integrate feedback into your application:"
    echo "     • Add thumbs-up/thumbs-down buttons to search results"
    echo "     • Submit feedback automatically from AI agents"
    echo "     • Use automated validators for quality control"
    echo ""
    echo "  2. Monitor feedback analytics:"
    echo "     curl $API_BASE/rag/feedback/analytics"
    echo ""
    echo "  3. Track relevance score changes:"
    echo "     psql -d $DB_NAME -c 'SELECT * FROM kyc_attr_doc_links ORDER BY relevance_score DESC LIMIT 10;'"
    echo ""
    echo "  4. Review documentation:"
    echo "     • RAG_FEEDBACK.md - Complete guide"
    echo "     • RAG_VECTOR_SEARCH.md - Vector search details"
    echo "     • API_DOCUMENTATION.md - Full API reference"
    echo ""

    print_header "Thank you for exploring the RAG Feedback Loop!"
}

# Run main function
main
