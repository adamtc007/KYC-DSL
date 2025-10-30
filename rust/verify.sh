#!/bin/bash
# Rust workspace verification script for KYC-DSL

set -e

echo "🦀 KYC-DSL Rust Workspace Verification"
echo "======================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Track status
FAILURES=0

# Helper functions
check_pass() {
    echo -e "${GREEN}✓${NC} $1"
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    FAILURES=$((FAILURES + 1))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

section() {
    echo
    echo -e "${BLUE}▶${NC} $1"
    echo "-----------------------------------"
}

# 1. Check prerequisites
section "Checking Prerequisites"

if command -v rustc &> /dev/null; then
    RUST_VERSION=$(rustc --version)
    check_pass "Rust installed: $RUST_VERSION"
else
    check_fail "Rust not found. Install from: https://rustup.rs"
fi

if command -v cargo &> /dev/null; then
    CARGO_VERSION=$(cargo --version)
    check_pass "Cargo installed: $CARGO_VERSION"
else
    check_fail "Cargo not found"
fi

if command -v protoc &> /dev/null; then
    PROTOC_VERSION=$(protoc --version)
    check_pass "Protocol Buffers compiler: $PROTOC_VERSION"
else
    check_fail "protoc not found. Install with: brew install protobuf"
fi

# 2. Check workspace structure
section "Checking Workspace Structure"

if [ -f "Cargo.toml" ]; then
    check_pass "Workspace Cargo.toml exists"
else
    check_fail "Workspace Cargo.toml not found"
fi

if [ -d "kyc_dsl_core" ]; then
    check_pass "kyc_dsl_core crate directory exists"
else
    check_fail "kyc_dsl_core directory not found"
fi

if [ -d "kyc_dsl_service" ]; then
    check_pass "kyc_dsl_service crate directory exists"
else
    check_fail "kyc_dsl_service directory not found"
fi

if [ -f "../api/proto/dsl_service.proto" ]; then
    check_pass "Protocol buffer definition found"
else
    check_fail "dsl_service.proto not found at ../api/proto/"
fi

# 3. Check source files
section "Checking Source Files"

REQUIRED_FILES=(
    "kyc_dsl_core/src/lib.rs"
    "kyc_dsl_core/src/parser.rs"
    "kyc_dsl_core/src/compiler.rs"
    "kyc_dsl_core/src/executor.rs"
    "kyc_dsl_service/src/main.rs"
    "kyc_dsl_service/build.rs"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        check_pass "$file"
    else
        check_fail "$file missing"
    fi
done

# 4. Build workspace
section "Building Workspace"

echo "Running: cargo build --release"
if cargo build --release 2>&1 | tee /tmp/rust-build.log; then
    check_pass "Workspace built successfully"
else
    check_fail "Build failed. Check /tmp/rust-build.log"
    cat /tmp/rust-build.log
fi

# 5. Check build artifacts
section "Checking Build Artifacts"

if [ -f "target/release/kyc_dsl_service" ]; then
    SIZE=$(ls -lh target/release/kyc_dsl_service | awk '{print $5}')
    check_pass "Binary created: kyc_dsl_service ($SIZE)"
else
    check_fail "Binary not found: target/release/kyc_dsl_service"
fi

if [ -f "target/release/libkyc_dsl_core.rlib" ]; then
    check_pass "Library compiled: libkyc_dsl_core.rlib"
else
    check_warn "Library artifact not found (may be optimized away)"
fi

# 6. Run tests
section "Running Tests"

echo "Running: cargo test"
if cargo test 2>&1 | tee /tmp/rust-test.log; then
    TEST_COUNT=$(grep -o "test result: ok" /tmp/rust-test.log | wc -l || echo "0")
    check_pass "All tests passed (test suites: $TEST_COUNT)"
else
    check_fail "Some tests failed. Check /tmp/rust-test.log"
fi

# 7. Code quality checks
section "Code Quality Checks"

if command -v cargo-fmt &> /dev/null || cargo fmt --version &> /dev/null 2>&1; then
    if cargo fmt -- --check &> /dev/null; then
        check_pass "Code formatting OK"
    else
        check_warn "Code needs formatting. Run: cargo fmt"
    fi
else
    check_warn "rustfmt not installed. Run: rustup component add rustfmt"
fi

if command -v cargo-clippy &> /dev/null || cargo clippy --version &> /dev/null 2>&1; then
    if cargo clippy -- -D warnings &> /dev/null 2>&1; then
        check_pass "Clippy checks passed"
    else
        check_warn "Clippy warnings found. Run: cargo clippy"
    fi
else
    check_warn "clippy not installed. Run: rustup component add clippy"
fi

# 8. Verify gRPC service can start
section "Service Startup Test"

echo "Testing if service binary runs (timeout 3s)..."
timeout 3s ./target/release/kyc_dsl_service &> /tmp/rust-service-test.log || true

if grep -q "Listening on" /tmp/rust-service-test.log; then
    check_pass "Service starts successfully"
    PORT=$(grep -o "\[::\]:50060\|127.0.0.1:50060" /tmp/rust-service-test.log | head -1)
    echo "  Service port: $PORT"
else
    check_warn "Could not verify service startup"
fi

# 9. Dependencies check
section "Dependency Information"

echo "Direct dependencies:"
cargo tree --depth 1 | head -20

# 10. Summary
echo
echo "======================================="
echo -e "${BLUE}Verification Summary${NC}"
echo "======================================="

if [ $FAILURES -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed!${NC}"
    echo
    echo "Next steps:"
    echo "  1. Start the service: make run-rust"
    echo "  2. Test with grpcurl: grpcurl -plaintext localhost:50060 list"
    echo "  3. Integrate with Go: Update internal/service/dsl_client.go"
    echo
    exit 0
else
    echo -e "${RED}✗ $FAILURES check(s) failed${NC}"
    echo
    echo "Please fix the issues above and run again."
    echo
    exit 1
fi
