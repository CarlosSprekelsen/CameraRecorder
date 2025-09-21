#!/bin/bash

# Test Naming Convention Migration Script
# This script helps identify and migrate test function names to the standardized format

echo "🔍 MediaMTX Test Naming Convention Migration Tool"
echo "================================================="

# Find all test files
echo "📁 Scanning for test files..."
find . -name "*_test.go" -type f | wc -l | xargs echo "Found test files:"

echo ""
echo "🔍 Analyzing current test function patterns..."

# Show current patterns that need migration
echo ""
echo "❌ Functions missing requirement tags:"
grep -r "^func Test.*\(t \*testing\.T\)" --include="*_test.go" . | grep -v "_Req" | head -10

echo ""
echo "❌ Functions with inconsistent naming:"
grep -r "^func Test.*\(t \*testing\.T\)" --include="*_test.go" . | grep -E "(Creation|StartStop|DoubleStart)" | head -10

echo ""
echo "✅ Functions already following standard:"
grep -r "^func Test.*_Req.*_.*\(t \*testing\.T\)" --include="*_test.go" . | head -5

echo ""
echo "📊 Migration Progress:"
total_tests=$(grep -r "^func Test.*\(t \*testing\.T\)" --include="*_test.go" . | wc -l)
standardized_tests=$(grep -r "^func Test.*_Req.*_.*\(t \*testing\.T\)" --include="*_test.go" . | wc -l)
percentage=$((standardized_tests * 100 / total_tests))

echo "Total test functions: $total_tests"
echo "Standardized functions: $standardized_tests"
echo "Completion: $percentage%"

echo ""
echo "🎯 Next files to migrate (by priority):"
echo "1. internal/mediamtx/controller_test.go (18 functions remaining)"
echo "2. internal/websocket/server_test.go (19 functions remaining)" 
echo "3. internal/websocket/methods_test.go (18 functions remaining)"
echo "4. internal/camera/*_test.go (8 files)"
echo "5. internal/security/*_test.go (6 files)"

echo ""
echo "📝 To continue migration:"
echo "1. Review docs/development/test-naming-conventions.md"
echo "2. Apply standard format: TestComponent_Method_Requirement_Scenario"
echo "3. Update function names in batches by file"
echo "4. Run tests to ensure no breakage"
