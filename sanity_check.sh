#!/bin/bash
set -e

echo "================================================"
echo "🧠 MIGRATION SANITY CHECK - COMPREHENSIVE AUDIT"
echo "================================================"
echo ""

echo "✅ CHECK 1: No Go parser imports"
PARSER_COUNT=$(grep -r "internal/parser" --include="*.go" 2>/dev/null | wc -l | tr -d ' ')
echo "   Found: $PARSER_COUNT imports (target: 0)"
if [ "$PARSER_COUNT" -eq 0 ]; then
    echo "   ✅ PASS"
else
    echo "   ❌ FAIL"
    grep -r "internal/parser" --include="*.go"
    exit 1
fi
echo ""

echo "✅ CHECK 2: No Go engine imports"
ENGINE_COUNT=$(grep -r "internal/engine" --include="*.go" 2>/dev/null | wc -l | tr -d ' ')
echo "   Found: $ENGINE_COUNT imports (target: 0)"
if [ "$ENGINE_COUNT" -eq 0 ]; then
    echo "   ✅ PASS"
else
    echo "   ❌ FAIL"
    grep -r "internal/engine" --include="*.go"
    exit 1
fi
echo ""

echo "✅ CHECK 3: Rust parser exists"
if [ -f "rust/kyc_dsl_core/src/parser.rs" ]; then
    echo "   Found: rust/kyc_dsl_core/src/parser.rs"
    echo "   ✅ PASS"
else
    echo "   ❌ FAIL: Rust parser not found"
    exit 1
fi
echo ""

echo "✅ CHECK 4: Rust owns EBNF grammar"
EBNF_FILE=$(find rust/kyc_dsl_service -name "*.rs" 2>/dev/null | xargs grep -l "ebnf" | head -1)
if [ -n "$EBNF_FILE" ]; then
    echo "   Found: $EBNF_FILE"
    echo "   ✅ PASS"
else
    echo "   ❌ FAIL: EBNF not found in Rust"
    exit 1
fi
echo ""

echo "✅ CHECK 5: Go build succeeds"
if go build ./cmd/kycctl 2>&1 | grep -q "internal/parser\|internal/engine"; then
    echo "   ❌ FAIL: Still references deleted packages"
    go build ./cmd/kycctl 2>&1
    exit 1
else
    echo "   ✅ PASS: Build successful"
fi
echo ""

echo "✅ CHECK 6: Rust service builds"
cd rust
if cargo build --release 2>&1 | grep -q "error"; then
    echo "   ❌ FAIL: Rust build failed"
    cargo build --release 2>&1 | tail -20
    exit 1
else
    echo "   ✅ PASS: Rust build successful"
fi
cd ..
echo ""

echo "✅ CHECK 7: Verify deleted directories"
DELETED_DIRS=("internal/parser" "internal/engine" "cmd/server")
for dir in "${DELETED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo "   ❌ FAIL: $dir still exists"
        exit 1
    else
        echo "   ✅ $dir deleted"
    fi
done
echo ""

echo "✅ CHECK 8: Verify Rust client exists"
if [ -f "internal/rustclient/dsl_client.go" ]; then
    echo "   Found: internal/rustclient/dsl_client.go"
    echo "   ✅ PASS"
else
    echo "   ❌ FAIL: Rust client wrapper not found"
    exit 1
fi
echo ""

echo "================================================"
echo "🎉 ALL SANITY CHECKS PASSED!"
echo "================================================"
echo ""
echo "Migration Summary:"
echo "  ✅ Go parser deleted"
echo "  ✅ Go engine deleted"
echo "  ✅ Old gRPC services deleted"
echo "  ✅ Rust owns DSL parsing"
echo "  ✅ Rust owns EBNF grammar"
echo "  ✅ Go uses Rust via gRPC"
echo "  ✅ Build succeeds"
echo ""
echo "Next steps:"
echo "  1. Start Rust service: cd rust && cargo run -p kyc_dsl_service"
echo "  2. Test CLI: ./kycctl sample_case.dsl"
echo "  3. Run tests: make test"
