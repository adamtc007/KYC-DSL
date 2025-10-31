#!/bin/bash

echo "================================================================"
echo "🔍 COMPREHENSIVE GO FILE CHECK - KYC-DSL PROJECT"
echo "================================================================"
echo ""

echo "📊 Checking all packages..."
echo ""

PACKAGES=$(go list ./... 2>/dev/null | grep -v "/test")
TOTAL=0
SUCCESS=0
FAILED=0

for pkg in $PACKAGES; do
    TOTAL=$((TOTAL + 1))
    if go build -o /dev/null "$pkg" 2>/dev/null; then
        SUCCESS=$((SUCCESS + 1))
        echo "✅ $pkg"
    else
        FAILED=$((FAILED + 1))
        echo "❌ $pkg"
        go build -o /dev/null "$pkg" 2>&1 | head -3 | sed 's/^/   /'
    fi
done

echo ""
echo "================================================================"
echo "📊 SUMMARY"
echo "================================================================"
echo "Total packages:     $TOTAL"
echo "✅ Successful:      $SUCCESS"
echo "❌ Failed:          $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "🎉 ALL PACKAGES BUILD SUCCESSFULLY!"
else
    echo "⚠️  Some packages have errors (see details above)"
fi

echo ""
echo "================================================================"
echo "🔍 MIGRATION-SPECIFIC CHECKS"
echo "================================================================"

echo "Parser imports:   $(grep -r 'internal/parser' --include='*.go' 2>/dev/null | wc -l | tr -d ' ') (target: 0)"
echo "Engine imports:   $(grep -r 'internal/engine' --include='*.go' 2>/dev/null | wc -l | tr -d ' ') (target: 0)"
echo ""
echo "Deleted directories:"
[ ! -d "internal/parser" ] && echo "  ✅ internal/parser deleted" || echo "  ❌ internal/parser exists"
[ ! -d "internal/engine" ] && echo "  ✅ internal/engine deleted" || echo "  ❌ internal/engine exists"
[ ! -d "cmd/server" ] && echo "  ✅ cmd/server deleted" || echo "  ❌ cmd/server exists"

