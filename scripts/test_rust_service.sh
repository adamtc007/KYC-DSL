#!/bin/bash
# Test script for Rust gRPC service verification
# Tests the kyc_dsl_service on port 50060

set -e

RUST_SERVICE_PORT=50060
RUST_SERVICE_ADDR="localhost:${RUST_SERVICE_PORT}"
SERVICE_NAME="kyc.dsl.DslService"
RUST_BIN="./rust/target/release/kyc_dsl_service"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🦀 Rust gRPC Service Test${NC}"
echo "=================================="
echo ""

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo -e "${RED}❌ grpcurl is not installed${NC}"
    echo "Install with: brew install grpcurl"
    exit 1
fi

# Build the Rust service
echo -e "${YELLOW}📦 Building Rust service...${NC}"
cd rust
cargo build --release -p kyc_dsl_service
cd ..

if [ ! -f "$RUST_BIN" ]; then
    echo -e "${RED}❌ Build failed: $RUST_BIN not found${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Build successful${NC}"
echo ""

# Start the Rust service in background
echo -e "${YELLOW}🚀 Starting Rust gRPC service on port ${RUST_SERVICE_PORT}...${NC}"
$RUST_BIN > /tmp/rust_service.log 2>&1 &
RUST_PID=$!

# Function to cleanup on exit
cleanup() {
    if [ -n "$RUST_PID" ]; then
        echo ""
        echo -e "${YELLOW}🧹 Stopping Rust service (PID: $RUST_PID)...${NC}"
        kill $RUST_PID 2>/dev/null || true
        wait $RUST_PID 2>/dev/null || true
    fi
    rm -f /tmp/rust_service.log
}

trap cleanup EXIT

# Wait for service to start
echo -e "${YELLOW}⏳ Waiting for service to be ready...${NC}"
MAX_WAIT=10
WAIT_COUNT=0
while ! lsof -i :$RUST_SERVICE_PORT > /dev/null 2>&1; do
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
    if [ $WAIT_COUNT -ge $MAX_WAIT ]; then
        echo -e "${RED}❌ Service failed to start after ${MAX_WAIT}s${NC}"
        echo "Service log:"
        cat /tmp/rust_service.log
        exit 1
    fi
done

echo -e "${GREEN}✅ Service is running (PID: $RUST_PID)${NC}"
echo ""

# Test 1: List available services
echo -e "${BLUE}Test 1: List gRPC services${NC}"
if grpcurl -plaintext $RUST_SERVICE_ADDR list > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Service discovery works${NC}"
    grpcurl -plaintext $RUST_SERVICE_ADDR list
else
    echo -e "${RED}❌ Service discovery failed${NC}"
    exit 1
fi
echo ""

# Test 2: List methods for DslService
echo -e "${BLUE}Test 2: List service methods${NC}"
if grpcurl -plaintext $RUST_SERVICE_ADDR list $SERVICE_NAME > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Method listing works${NC}"
    grpcurl -plaintext $RUST_SERVICE_ADDR list $SERVICE_NAME
else
    echo -e "${RED}❌ Method listing failed${NC}"
    exit 1
fi
echo ""

# Test 3: Execute RPC
echo -e "${BLUE}Test 3: Execute DSL function${NC}"
EXECUTE_REQUEST='{
  "case_id": "TEST-CASE-001",
  "function_name": "DISCOVER-POLICIES"
}'

if EXECUTE_RESPONSE=$(grpcurl -plaintext -d "$EXECUTE_REQUEST" $RUST_SERVICE_ADDR ${SERVICE_NAME}/Execute 2>&1); then
    echo -e "${GREEN}✅ Execute RPC successful${NC}"
    echo "Response:"
    echo "$EXECUTE_RESPONSE" | jq '.' 2>/dev/null || echo "$EXECUTE_RESPONSE"
else
    echo -e "${RED}❌ Execute RPC failed${NC}"
    echo "$EXECUTE_RESPONSE"
    exit 1
fi
echo ""

# Test 4: Validate RPC
echo -e "${BLUE}Test 4: Validate DSL syntax${NC}"
VALIDATE_REQUEST='{
  "dsl_source": "(kyc-case TEST (nature-purpose (nature \"test\") (purpose \"test\")))"
}'

if VALIDATE_RESPONSE=$(grpcurl -plaintext -d "$VALIDATE_REQUEST" $RUST_SERVICE_ADDR ${SERVICE_NAME}/Validate 2>&1); then
    echo -e "${GREEN}✅ Validate RPC successful${NC}"
    echo "Response:"
    echo "$VALIDATE_RESPONSE" | jq '.' 2>/dev/null || echo "$VALIDATE_RESPONSE"
else
    echo -e "${RED}❌ Validate RPC failed${NC}"
    echo "$VALIDATE_RESPONSE"
    exit 1
fi
echo ""

# Test 5: Parse RPC
echo -e "${BLUE}Test 5: Parse DSL to AST${NC}"
PARSE_REQUEST='{
  "dsl_source": "(kyc-case PARSE-TEST (function VERIFY))"
}'

if PARSE_RESPONSE=$(grpcurl -plaintext -d "$PARSE_REQUEST" $RUST_SERVICE_ADDR ${SERVICE_NAME}/Parse 2>&1); then
    echo -e "${GREEN}✅ Parse RPC successful${NC}"
    echo "Response:"
    echo "$PARSE_RESPONSE" | jq '.' 2>/dev/null || echo "$PARSE_RESPONSE"
else
    echo -e "${RED}❌ Parse RPC failed${NC}"
    echo "$PARSE_RESPONSE"
    exit 1
fi
echo ""

# Summary
echo "=================================="
echo -e "${GREEN}✅ All Rust gRPC service tests passed!${NC}"
echo ""
echo "Service Details:"
echo "  Address: $RUST_SERVICE_ADDR"
echo "  Service: $SERVICE_NAME"
echo "  Binary: $RUST_BIN"
echo ""
echo "Next Steps:"
echo "  1. Test Go → Rust interop"
echo "  2. Begin validator + audit chain phase"
echo "  3. Add integration tests"
echo ""
