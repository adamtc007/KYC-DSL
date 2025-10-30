#!/bin/bash
# verify.sh - Comprehensive verification script for KYC-DSL project

set -e

echo "=========================================="
echo "🔍 KYC-DSL Verification Script"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track overall status
ERRORS=0

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        ERRORS=$((ERRORS + 1))
    fi
}

# 1. Check Go installation
echo "📦 Checking Go installation..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    echo "   $GO_VERSION"
    print_status 0 "Go installed"
else
    print_status 1 "Go not found"
    exit 1
fi
echo ""

# 2. Check dependencies
echo "📚 Checking dependencies..."
go mod verify &> /dev/null
print_status $? "Go modules verified"
echo ""

# 3. Build check
echo "🏗️  Building project..."
make clean > /dev/null 2>&1
if make build > /dev/null 2>&1; then
    print_status 0 "Build successful (with greenteagc)"
else
    print_status 1 "Build failed"
fi
echo ""

# 4. Go vet
echo "🔎 Running go vet..."
if go vet ./... 2>&1 | grep -v "no Go files" > /dev/null; then
    print_status 1 "go vet found issues"
else
    print_status 0 "go vet passed"
fi
echo ""

# 5. Go fmt
echo "📝 Checking code formatting..."
UNFORMATTED=$(gofmt -l . 2>&1 | grep -v "vendor/" | grep "\.go$" || true)
if [ -z "$UNFORMATTED" ]; then
    print_status 0 "Code properly formatted"
else
    print_status 1 "Code needs formatting"
    echo "   Files needing formatting:"
    echo "$UNFORMATTED" | sed 's/^/   - /'
fi
echo ""

# 6. Tests
echo "🧪 Running tests..."
if GOEXPERIMENT=greenteagc go test ./... 2>&1 | grep -q "FAIL"; then
    print_status 1 "Tests failed"
else
    TEST_OUTPUT=$(GOEXPERIMENT=greenteagc go test ./... 2>&1)
    if echo "$TEST_OUTPUT" | grep -q "ok"; then
        print_status 0 "Tests passed"
    else
        print_status 0 "No test failures (some packages have no tests)"
    fi
fi
echo ""

# 7. Golangci-lint (if available)
echo "🔍 Running golangci-lint..."
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run 2>&1 | grep -E "(error|Error)" > /dev/null; then
        print_status 1 "golangci-lint found issues"
    else
        print_status 0 "golangci-lint passed"
    fi
else
    echo -e "${YELLOW}   ⚠️  golangci-lint not installed (skipping)${NC}"
fi
echo ""

# 8. Check for common issues
echo "🔬 Checking for common issues..."

# Check for TODO/FIXME
TODO_COUNT=$(grep -r "TODO\|FIXME" --include="*.go" . 2>/dev/null | wc -l | tr -d ' ')
if [ "$TODO_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}   ⚠️  Found $TODO_COUNT TODO/FIXME comments${NC}"
else
    print_status 0 "No TODO/FIXME comments"
fi

# Check for hardcoded credentials (look for actual password strings)
if grep -rE 'password.*=.*"[^"]+[A-Za-z0-9]{8,}"' --include="*.go" . 2>/dev/null | grep -v "PGPASSWORD" | grep -v "test" > /dev/null; then
    print_status 1 "Potential hardcoded credentials found"
else
    print_status 0 "No hardcoded credentials detected"
fi
echo ""

# 9. Binary check
echo "🔧 Verifying binary..."
if [ -f "./bin/kycctl" ]; then
    print_status 0 "Binary exists at bin/kycctl"

    # Test help command
    if ./bin/kycctl help > /dev/null 2>&1; then
        print_status 0 "Binary executes successfully"
    else
        print_status 1 "Binary execution failed"
    fi
else
    print_status 1 "Binary not found"
fi
echo ""

# 10. Database schema check
echo "🗄️  Database configuration..."
if psql -d kyc_dsl -c "SELECT COUNT(*) FROM kyc_grammar;" > /dev/null 2>&1; then
    print_status 0 "Database connection successful"
else
    echo -e "${YELLOW}   ⚠️  Database not accessible (this is OK if not set up yet)${NC}"
fi
echo ""

# Final summary
echo "=========================================="
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed!${NC}"
    echo "=========================================="
    exit 0
else
    echo -e "${RED}❌ Found $ERRORS issue(s)${NC}"
    echo "=========================================="
    exit 1
fi
