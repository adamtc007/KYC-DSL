#!/usr/bin/env bash
# ============================================================================
# KYC Data Service - Integration Test Script
# ============================================================================
# Tests all gRPC endpoints of the Data Service
# Requires: grpcurl, running Data Service on port 50070
# ============================================================================

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Data Service endpoint
DATA_SERVICE="localhost:50070"

# Test case ID (unique per run)
TEST_CASE_ID="TEST-$(date +%s)"

echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}KYC Data Service - Integration Tests${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""
echo -e "${CYAN}Test Configuration:${NC}"
echo -e "  Service: ${DATA_SERVICE}"
echo -e "  Test Case ID: ${TEST_CASE_ID}"
echo ""

# ============================================================================
# Helper Functions
# ============================================================================

run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected_pattern="$3"

    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "${YELLOW}[TEST $TESTS_RUN] ${test_name}${NC}"

    # Run the command and capture output
    if output=$(eval "$test_cmd" 2>&1); then
        # Check if output matches expected pattern (if provided)
        if [ -z "$expected_pattern" ] || echo "$output" | grep -q "$expected_pattern"; then
            echo -e "${GREEN}  ✅ PASSED${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            echo -e "${RED}  ❌ FAILED - Expected pattern not found: $expected_pattern${NC}"
            echo -e "${RED}  Output: $output${NC}"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    else
        echo -e "${RED}  ❌ FAILED - Command error${NC}"
        echo -e "${RED}  Output: $output${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# ============================================================================
# Pre-flight Checks
# ============================================================================

echo -e "${CYAN}🔍 Pre-flight Checks${NC}"
echo ""

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo -e "${RED}❌ grpcurl not found${NC}"
    echo -e "${YELLOW}Install with: brew install grpcurl (macOS) or go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest${NC}"
    exit 1
fi
echo -e "${GREEN}✅ grpcurl installed${NC}"

# Check if Data Service is running
if ! grpcurl -plaintext "$DATA_SERVICE" list > /dev/null 2>&1; then
    echo -e "${RED}❌ Data Service not running on ${DATA_SERVICE}${NC}"
    echo -e "${YELLOW}Start with: make run-dataserver${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Data Service is running${NC}"

# List available services
echo ""
echo -e "${CYAN}📋 Available Services:${NC}"
grpcurl -plaintext "$DATA_SERVICE" list | while read -r service; do
    echo -e "  • $service"
done

echo ""
echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}Dictionary Service Tests${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""

# ============================================================================
# Dictionary Service - GetAttribute
# ============================================================================

run_test "GetAttribute - CLIENT_NAME" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/GetAttribute -d '{\"id\": \"CLIENT_NAME\"}'" \
    "Client Name"

run_test "GetAttribute - CLIENT_LEI" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/GetAttribute -d '{\"id\": \"CLIENT_LEI\"}'" \
    "Legal Entity Identifier"

run_test "GetAttribute - UBO_NAME" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/GetAttribute -d '{\"id\": \"UBO_NAME\"}'" \
    "Ultimate Beneficial Owner"

# ============================================================================
# Dictionary Service - ListAttributes
# ============================================================================

run_test "ListAttributes - First 5 attributes" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/ListAttributes -d '{\"limit\": 5, \"offset\": 0}'" \
    "attributes"

run_test "ListAttributes - With pagination (offset 2)" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/ListAttributes -d '{\"limit\": 3, \"offset\": 2}'" \
    "totalCount"

run_test "ListAttributes - Large limit (50)" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/ListAttributes -d '{\"limit\": 50}'" \
    "attributes"

# ============================================================================
# Dictionary Service - GetDocument
# ============================================================================

run_test "GetDocument - DOC_PASSPORT" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/GetDocument -d '{\"id\": \"DOC_PASSPORT\"}'" \
    "Passport"

run_test "GetDocument - DOC_W9" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/GetDocument -d '{\"id\": \"DOC_W9\"}'" \
    "IRS Form W-9"

run_test "GetDocument - DOC_CERT_INC" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/GetDocument -d '{\"id\": \"DOC_CERT_INC\"}'" \
    "Certificate of Incorporation"

# ============================================================================
# Dictionary Service - ListDocuments
# ============================================================================

run_test "ListDocuments - All documents" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/ListDocuments -d '{\"limit\": 10, \"offset\": 0}'" \
    "documents"

run_test "ListDocuments - US jurisdiction filter" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/ListDocuments -d '{\"limit\": 10, \"jurisdiction\": \"US\"}'" \
    "DOC_W"

run_test "ListDocuments - GLOBAL jurisdiction filter" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/ListDocuments -d '{\"limit\": 20, \"jurisdiction\": \"GLOBAL\"}'" \
    "DOC_PASSPORT"

echo ""
echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}Case Service Tests${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""

# ============================================================================
# Case Service - SaveCaseVersion
# ============================================================================

run_test "SaveCaseVersion - Create draft case" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.CaseService/SaveCaseVersion -d '{
        \"case_id\": \"$TEST_CASE_ID\",
        \"dsl_source\": \"(kyc-case TEST-CORP (nature-purpose (nature \\\"Technology\\\") (purpose \\\"Investment\\\")))\",
        \"compiled_json\": \"{\\\"case_id\\\": \\\"TEST-CORP\\\"}\",
        \"status\": \"draft\"
    }'" \
    "success.*true"

run_test "SaveCaseVersion - Create validated case" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.CaseService/SaveCaseVersion -d '{
        \"case_id\": \"${TEST_CASE_ID}-V2\",
        \"dsl_source\": \"(kyc-case TEST-CORP-V2 (policy POL001))\",
        \"compiled_json\": \"{\\\"case_id\\\": \\\"TEST-CORP-V2\\\", \\\"policy\\\": \\\"POL001\\\"}\",
        \"status\": \"validated\"
    }'" \
    "success"

run_test "SaveCaseVersion - Second version of same case" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.CaseService/SaveCaseVersion -d '{
        \"case_id\": \"$TEST_CASE_ID\",
        \"dsl_source\": \"(kyc-case TEST-CORP (kyc-token \\\"updated\\\"))\",
        \"compiled_json\": \"{\\\"case_id\\\": \\\"TEST-CORP\\\", \\\"updated\\\": true}\",
        \"status\": \"approved\"
    }'" \
    "versionId"

# ============================================================================
# Case Service - GetCaseVersion
# ============================================================================

run_test "GetCaseVersion - Retrieve latest version" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.CaseService/GetCaseVersion -d '{\"case_id\": \"$TEST_CASE_ID\"}'" \
    "caseId.*$TEST_CASE_ID"

run_test "GetCaseVersion - Check status is latest (approved)" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.CaseService/GetCaseVersion -d '{\"case_id\": \"$TEST_CASE_ID\"}'" \
    "approved"

run_test "GetCaseVersion - Retrieve second test case" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.CaseService/GetCaseVersion -d '{\"case_id\": \"${TEST_CASE_ID}-V2\"}'" \
    "validated"

# ============================================================================
# Case Service - ListCaseVersions
# ============================================================================

run_test "ListCaseVersions - Get all versions of test case" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.CaseService/ListCaseVersions -d '{
        \"case_id\": \"$TEST_CASE_ID\",
        \"limit\": 10,
        \"offset\": 0
    }'" \
    "versions"

run_test "ListCaseVersions - Check version count (should be 2)" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.CaseService/ListCaseVersions -d '{
        \"case_id\": \"$TEST_CASE_ID\",
        \"limit\": 10
    }'" \
    "totalCount.*2"

run_test "ListCaseVersions - Pagination with limit 1" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.CaseService/ListCaseVersions -d '{
        \"case_id\": \"$TEST_CASE_ID\",
        \"limit\": 1,
        \"offset\": 0
    }'" \
    "versions"

# ============================================================================
# Edge Cases and Error Handling
# ============================================================================

echo ""
echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}Edge Cases and Error Handling${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""

run_test "GetAttribute - Non-existent attribute" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/GetAttribute -d '{\"id\": \"NONEXISTENT_ATTR\"}' 2>&1 || true" \
    "not found"

run_test "GetDocument - Non-existent document" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/GetDocument -d '{\"id\": \"NONEXISTENT_DOC\"}' 2>&1 || true" \
    "not found"

run_test "GetCaseVersion - Non-existent case" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.CaseService/GetCaseVersion -d '{\"case_id\": \"NONEXISTENT-CASE-999\"}' 2>&1 || true" \
    "not found"

run_test "ListAttributes - Zero limit (should use default)" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/ListAttributes -d '{\"limit\": 0}'" \
    "attributes"

run_test "ListDocuments - Negative offset (should use 0)" \
    "grpcurl -plaintext $DATA_SERVICE kyc.data.DictionaryService/ListDocuments -d '{\"limit\": 5, \"offset\": -1}'" \
    "documents"

# ============================================================================
# Summary
# ============================================================================

echo ""
echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""

echo -e "${CYAN}Results:${NC}"
echo -e "  Total Tests:  ${TESTS_RUN}"
echo -e "  ${GREEN}Passed:       ${TESTS_PASSED}${NC}"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "  ${RED}Failed:       ${TESTS_FAILED}${NC}"
else
    echo -e "  ${GREEN}Failed:       0${NC}"
fi

echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}============================================================================${NC}"
    echo -e "${GREEN}✅ ALL TESTS PASSED${NC}"
    echo -e "${GREEN}============================================================================${NC}"
    exit 0
else
    PASS_RATE=$((TESTS_PASSED * 100 / TESTS_RUN))
    echo -e "${YELLOW}============================================================================${NC}"
    echo -e "${YELLOW}⚠️  SOME TESTS FAILED (Pass Rate: ${PASS_RATE}%)${NC}"
    echo -e "${YELLOW}============================================================================${NC}"
    exit 1
fi
