#!/bin/bash
echo "=== ANALYZING CALL TREE ==="
echo ""

echo "1. CLI Entry Points:"
grep -n "^func Run" internal/cli/cli.go | head -20

echo ""
echo "2. Rust Client Calls:"
grep -n "func.*Client" internal/rustclient/dsl_client.go | head -20

echo ""
echo "3. Storage Layer Functions:"
grep -n "^func" internal/storage/*.go | grep -v "test" | head -30

echo ""
echo "4. Who calls rustclient?"
grep -r "rustclient\." internal/ --include="*.go" | cut -d: -f1 | sort -u

echo ""
echo "5. Dead code check - functions never called:"
for file in internal/*/*.go; do
  funcs=$(grep -o "^func [A-Z][a-zA-Z]*" "$file" | cut -d' ' -f2)
  for func in $funcs; do
    count=$(grep -r "\.$func\|$func(" --include="*.go" . | grep -v "^$file" | wc -l | tr -d ' ')
    if [ "$count" -eq 0 ]; then
      echo "  ⚠️  $file::$func - NOT CALLED"
    fi
  done
done 2>/dev/null | head -20
