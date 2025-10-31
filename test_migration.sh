#!/bin/bash
echo "==================================================="
echo "🧪 TESTING MIGRATION-SPECIFIC FILES"
echo "==================================================="
echo ""

echo "✅ TEST 1: Rust client compiles"
go build -o /dev/null internal/rustclient/dsl_client.go && echo "   ✅ PASS" || echo "   ❌ FAIL"

echo ""
echo "✅ TEST 2: CLI compiles (uses rustclient)"
go build -o /dev/null ./cmd/kycctl && echo "   ✅ PASS" || echo "   ❌ FAIL"

echo ""
echo "✅ TEST 3: No parser imports"
FOUND=$(grep -r "internal/parser" --include="*.go" 2>/dev/null | wc -l | tr -d ' ')
if [ "$FOUND" -eq 0 ]; then
    echo "   ✅ PASS - 0 imports found"
else
    echo "   ❌ FAIL - $FOUND imports found"
fi

echo ""
echo "✅ TEST 4: No engine imports"
FOUND=$(grep -r "internal/engine" --include="*.go" 2>/dev/null | wc -l | tr -d ' ')
if [ "$FOUND" -eq 0 ]; then
    echo "   ✅ PASS - 0 imports found"
else
    echo "   ❌ FAIL - $FOUND imports found"
fi

echo ""
echo "✅ TEST 5: Deleted directories verified"
! [ -d "internal/parser" ] && ! [ -d "internal/engine" ] && ! [ -d "cmd/server" ] && echo "   ✅ PASS - All deleted" || echo "   ❌ FAIL - Some exist"

echo ""
echo "==================================================="
echo "MIGRATION VERIFICATION COMPLETE"
echo "==================================================="
