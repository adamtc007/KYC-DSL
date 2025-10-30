#!/bin/bash
# Preflight checks for KYC-DSL Rust services
# Verifies all dependencies are installed and services are ready

set -e

echo "🔍 KYC-DSL Rust Preflight Checks"
echo "================================="
echo

FAIL=0
WARN=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

check_pass() {
    echo -e "${GREEN}✓${NC} $1"
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    FAIL=$((FAIL + 1))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    WARN=$((WARN + 1))
}

check_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

section() {
    echo
    echo -e "${BLUE}▶${NC} $1"
    echo "-----------------------------------"
}

# 1. Check Rust Toolchain
section "Rust Toolchain"

if command -v rustc > /dev/null 2>&1; then
    RUST_VERSION=$(rustc --version | awk '{print $2}')
    check_pass "Rust installed: $RUST_VERSION"

    # Check version is >= 1.70
    RUST_MAJOR=$(echo $RUST_VERSION | cut -d. -f1)
    RUST_MINOR=$(echo $RUST_VERSION | cut -d. -f2)
    if [ "$RUST_MAJOR" -gt 1 ] || ([ "$RUST_MAJOR" -eq 1 ] && [ "$RUST_MINOR" -ge 70 ]); then
        check_pass "Rust version is sufficient (>= 1.70)"
    else
        check_warn "Rust version $RUST_VERSION is older than recommended 1.70+"
    fi
else
    check_fail "Rust not found"
    echo "  Install with: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
fi

if command -v cargo > /dev/null 2>&1; then
    CARGO_VERSION=$(cargo --version)
    check_pass "Cargo installed: $CARGO_VERSION"
else
    check_fail "Cargo not found"
fi

# Check Rust components
if rustup component list 2>/dev/null | grep -q "rustfmt-.*installed"; then
    check_pass "rustfmt installed"
else
    check_warn "rustfmt not installed"
    echo "  Install with: rustup component add rustfmt"
fi

if rustup component list 2>/dev/null | grep -q "clippy-.*installed"; then
    check_pass "clippy installed"
else
    check_warn "clippy not installed"
    echo "  Install with: rustup component add clippy"
fi

# 2. Check Go Toolchain
section "Go Toolchain"

if command -v go > /dev/null 2>&1; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    check_pass "Go installed: $GO_VERSION"
else
    check_fail "Go not found"
    echo "  Install with: brew install go (macOS) or download from https://go.dev/dl/"
fi

# 3. Check Protocol Buffers
section "Protocol Buffers Compiler"

if command -v protoc > /dev/null 2>&1; then
    PROTOC_VERSION=$(protoc --version)
    check_pass "protoc installed: $PROTOC_VERSION"
else
    check_fail "protoc not found"
    echo "  macOS: brew install protobuf"
    echo "  Ubuntu: sudo apt install protobuf-compiler"
fi

# 4. Check PostgreSQL
section "PostgreSQL Database"

if command -v psql > /dev/null 2>&1; then
    PSQL_VERSION=$(psql --version)
    check_pass "PostgreSQL client installed: $PSQL_VERSION"
else
    check_fail "PostgreSQL client not found"
    echo "  macOS: brew install postgresql@15"
    echo "  Ubuntu: sudo apt install postgresql postgresql-contrib"
fi

# Check if PostgreSQL is running
if psql -d postgres -c "SELECT 1;" > /dev/null 2>&1; then
    check_pass "PostgreSQL server is running"

    # Check if kyc_dsl database exists
    if psql -lqt | cut -d \| -f 1 | grep -qw kyc_dsl 2>/dev/null; then
        check_pass "Database 'kyc_dsl' exists"
    else
        check_warn "Database 'kyc_dsl' does not exist"
        echo "  Create with: createdb kyc_dsl"
    fi

    # Check for pgvector extension
    if psql -d kyc_dsl -c "SELECT 1 FROM pg_extension WHERE extname='vector';" 2>/dev/null | grep -q 1; then
        check_pass "pgvector extension installed"
    else
        check_warn "pgvector extension not installed (needed for RAG features)"
        echo "  Install with: psql -d kyc_dsl -c \"CREATE EXTENSION vector;\""
    fi
else
    check_fail "PostgreSQL server is NOT running"
    echo "  macOS: brew services start postgresql@15"
    echo "  Linux: sudo systemctl start postgresql"
    echo "  Docker: docker run -d -p 5432:5432 -e POSTGRES_DB=kyc_dsl postgres:15"
fi

# 5. Check grpcurl (optional but recommended)
section "Testing Tools"

if command -v grpcurl > /dev/null 2>&1; then
    GRPCURL_VERSION=$(grpcurl --version 2>&1 | head -1)
    check_pass "grpcurl installed: $GRPCURL_VERSION"
else
    check_warn "grpcurl not installed (optional, but recommended for testing)"
    echo "  macOS: brew install grpcurl"
    echo "  Linux: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
fi

if command -v curl > /dev/null 2>&1; then
    check_pass "curl installed"
else
    check_warn "curl not found (needed for REST API testing)"
fi

# 6. Check port availability
section "Port Availability"

check_port() {
    PORT=$1
    SERVICE=$2
    if lsof -i :$PORT > /dev/null 2>&1; then
        check_warn "Port $PORT is in use ($SERVICE may already be running)"
        lsof -i :$PORT | grep LISTEN | awk '{print "  Process: " $1 " (PID: " $2 ")"}'
    else
        check_pass "Port $PORT is available"
    fi
}

check_port 50060 "Rust gRPC service"
check_port 50051 "Go gRPC service"
check_port 8080 "Go REST API"
check_port 5432 "PostgreSQL"

# 7. Check environment variables
section "Environment Variables"

if [ -n "$PGDATABASE" ]; then
    check_pass "PGDATABASE=$PGDATABASE"
else
    check_info "PGDATABASE not set (will use default 'kyc_dsl')"
fi

if [ -n "$PGHOST" ]; then
    check_info "PGHOST=$PGHOST"
else
    check_info "PGHOST not set (will use default 'localhost')"
fi

if [ -n "$PGPORT" ]; then
    check_info "PGPORT=$PGPORT"
else
    check_info "PGPORT not set (will use default '5432')"
fi

if [ -n "$PGUSER" ]; then
    check_info "PGUSER=$PGUSER"
else
    check_info "PGUSER not set (will use current user)"
fi

if [ -n "$OPENAI_API_KEY" ]; then
    check_pass "OPENAI_API_KEY is set"
else
    check_warn "OPENAI_API_KEY not set (RAG/semantic search features will not work)"
    echo "  Set with: export OPENAI_API_KEY=sk-..."
fi

if [ -n "$RUST_LOG" ]; then
    check_info "RUST_LOG=$RUST_LOG"
else
    check_info "RUST_LOG not set (default logging will be used)"
fi

# 8. Check workspace structure
section "Workspace Structure"

if [ -f "Cargo.toml" ]; then
    check_pass "Cargo workspace found"
else
    check_fail "Cargo.toml not found - are you in the rust/ directory?"
fi

if [ -d "kyc_dsl_core" ]; then
    check_pass "kyc_dsl_core crate exists"
else
    check_fail "kyc_dsl_core directory not found"
fi

if [ -d "kyc_dsl_service" ]; then
    check_pass "kyc_dsl_service crate exists"
else
    check_fail "kyc_dsl_service directory not found"
fi

if [ -f "../api/proto/dsl_service.proto" ]; then
    check_pass "Proto definition found"
else
    check_fail "dsl_service.proto not found at ../api/proto/"
fi

# 9. Validate workspace metadata
section "Workspace Metadata"

if cargo metadata --format-version=1 > /dev/null 2>&1; then
    check_pass "Workspace metadata is valid"

    WORKSPACE_MEMBERS=$(cargo metadata --format-version=1 2>/dev/null | grep -o '"kyc_dsl_[^"]*"' | wc -l)
    if [ "$WORKSPACE_MEMBERS" -ge 2 ]; then
        check_pass "Found $WORKSPACE_MEMBERS workspace members"
    else
        check_fail "Expected 2 workspace members, found $WORKSPACE_MEMBERS"
    fi
else
    check_fail "Workspace metadata validation failed"
fi

# 10. Cargo check for core crate
section "Core Crate Validation (kyc_dsl_core)"

if cargo check -p kyc_dsl_core > /tmp/cargo_check_core.log 2>&1; then
    check_pass "kyc_dsl_core compiles successfully"
else
    check_fail "kyc_dsl_core compilation failed"
    echo "  See errors in /tmp/cargo_check_core.log"
    tail -5 /tmp/cargo_check_core.log | sed 's/^/  /'
fi

# 11. Cargo check for service crate
section "Service Crate Validation (kyc_dsl_service)"

if cargo check -p kyc_dsl_service > /tmp/cargo_check_service.log 2>&1; then
    check_pass "kyc_dsl_service compiles successfully"
else
    check_fail "kyc_dsl_service compilation failed"
    echo "  See errors in /tmp/cargo_check_service.log"
    tail -5 /tmp/cargo_check_service.log | sed 's/^/  /'
fi

# 12. Run tests
section "Test Execution"

if cargo test --no-fail-fast > /tmp/cargo_test.log 2>&1; then
    TEST_COUNT=$(grep "test result:" /tmp/cargo_test.log | tail -1 | grep -o "[0-9]* passed" | cut -d' ' -f1)
    check_pass "All tests passed ($TEST_COUNT tests)"
else
    check_fail "Some tests failed"
    echo "  See details in /tmp/cargo_test.log"
    grep "FAILED" /tmp/cargo_test.log | head -5 | sed 's/^/  /'
fi

# 13. Cargo clippy check
section "Code Quality (Clippy)"

if command -v cargo-clippy > /dev/null 2>&1 || rustup component list | grep -q "clippy.*installed"; then
    if cargo clippy --all-targets --all-features -- -D warnings > /tmp/cargo_clippy.log 2>&1; then
        check_pass "No clippy warnings"
    else
        CLIPPY_WARNINGS=$(grep "warning:" /tmp/cargo_clippy.log | wc -l)
        if [ "$CLIPPY_WARNINGS" -gt 0 ]; then
            check_warn "$CLIPPY_WARNINGS clippy warning(s) found"
            echo "  Run 'cargo clippy --fix' to auto-fix"
            head -10 /tmp/cargo_clippy.log | sed 's/^/  /'
        else
            check_fail "Clippy check failed with errors"
            tail -10 /tmp/cargo_clippy.log | sed 's/^/  /'
        fi
    fi
else
    check_info "Clippy not available (skipping)"
fi

# 14. Cargo format check
section "Code Formatting (rustfmt)"

if command -v cargo-fmt > /dev/null 2>&1 || rustup component list | grep -q "rustfmt.*installed"; then
    if cargo fmt --all -- --check > /tmp/cargo_fmt.log 2>&1; then
        check_pass "Code is properly formatted"
    else
        check_warn "Code formatting issues found"
        echo "  Run 'cargo fmt' to auto-format"
    fi
else
    check_info "rustfmt not available (skipping)"
fi

# 15. Quick build check
section "Build Verification"

if [ -d "target" ]; then
    check_info "Build artifacts exist (cargo build has been run)"

    if [ -f "target/release/kyc_dsl_service" ]; then
        check_pass "Release binary exists: kyc_dsl_service"
    else
        check_info "Release binary not built yet (run 'cargo build --release')"
    fi
else
    check_info "No build artifacts (run 'cargo build' first time)"
fi

# Summary
echo
echo "================================="
echo -e "${BLUE}Summary${NC}"
echo "================================="
echo

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✓ All critical checks passed!${NC}"
    if [ $WARN -gt 0 ]; then
        echo -e "${YELLOW}⚠ $WARN warning(s) - non-critical issues${NC}"
    fi
    echo
    echo "✅ All systems operational! You can now:"
    echo
    echo "  ${GREEN}cargo build --release${NC}         # Build optimized binaries"
    echo "  ${GREEN}cargo run -p kyc_dsl_service${NC}  # Start Rust gRPC service (port 50060)"
    echo "  ${GREEN}cargo test${NC}                    # Run all tests"
    echo "  ${GREEN}cargo clippy --fix${NC}            # Auto-fix linting issues"
    echo "  ${GREEN}cargo fmt${NC}                     # Auto-format code"
    echo
    echo "Or use Make targets from project root:"
    echo "  ${GREEN}make rust-build${NC}               # Build Rust workspace"
    echo "  ${GREEN}make rust-test${NC}                # Run Rust tests"
    echo "  ${GREEN}make rust-service${NC}             # Start Rust service"
    echo
    echo "Test the service with:"
    echo "  ${GREEN}grpcurl -plaintext localhost:50060 list${NC}"
    echo
    exit 0
else
    echo -e "${RED}✗ $FAIL critical check(s) failed${NC}"
    if [ $WARN -gt 0 ]; then
        echo -e "${YELLOW}⚠ $WARN warning(s)${NC}"
    fi
    echo
    echo "Please fix the critical issues above and run again."
    echo
    echo "Quick fixes:"
    echo "  Rust:       curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
    echo "  Components: rustup component add clippy rustfmt"
    echo "  protoc:     brew install protobuf (macOS) or apt install protobuf-compiler (Ubuntu)"
    echo "  Go:         brew install go (macOS) or download from https://go.dev/dl/"
    echo "  PostgreSQL: brew install postgresql@15 && brew services start postgresql@15"
    echo "  grpcurl:    brew install grpcurl (macOS)"
    echo
    echo "Clean rebuild:"
    echo "  cargo clean && cargo build --release"
    echo
    exit 1
fi
