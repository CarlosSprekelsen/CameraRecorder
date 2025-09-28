#!/bin/bash

# Documentation Validation Script
# Validates that documentation matches implementation

set -e

echo "🔍 Starting Documentation Validation..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}❌ $message${NC}"
            ;;
    esac
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo "📋 Checking prerequisites..."

if ! command_exists go; then
    print_status "ERROR" "Go is not installed"
    exit 1
fi

if ! command_exists grep; then
    print_status "ERROR" "grep is not available"
    exit 1
fi

print_status "SUCCESS" "Prerequisites check passed"

# Extract implemented methods from code
echo "🔍 Extracting implemented methods..."

IMPLEMENTED_METHODS_FILE="/tmp/implemented_methods.txt"
DOCUMENTED_METHODS_FILE="/tmp/documented_methods.txt"

# Extract methods from Go code
echo "Extracting methods from implementation..."
grep -r "Method.*func" ./internal/websocket/ | \
    grep -v test | \
    grep -o '"[^"]*"' | \
    sed 's/"//g' | \
    sort | uniq > "$IMPLEMENTED_METHODS_FILE"

# Extract methods from documentation
echo "Extracting methods from documentation..."
grep -n "### " docs/api/json_rpc_methods.md | \
    grep -v "Version\|Compatibility\|Indicators\|Process\|Flow\|Token\|Generation\|Levels\|Matrix\|Guarantees\|Methods\|Authentication Method\|Boolean Parameters\|Error Handling\|Error Response Fields\|GET /files\|Go Error Response Types\|Numeric Parameters\|Parameter Validation\|Required Fields\|Response Metadata\|Response Validation\|Role-Based Access Control\|Standard Error Response Format\|String Parameters\|Type Constraints" | \
    sed 's/.*### //' | \
    sort > "$DOCUMENTED_METHODS_FILE"

# Compare methods
echo "📊 Comparing implementation vs documentation..."

MISSING_DOCS_FILE="/tmp/missing_docs.txt"
EXTRA_DOCS_FILE="/tmp/extra_docs.txt"

# Find methods in implementation but not in documentation
comm -23 "$IMPLEMENTED_METHODS_FILE" "$DOCUMENTED_METHODS_FILE" > "$MISSING_DOCS_FILE"

# Find methods in documentation but not in implementation
comm -13 "$IMPLEMENTED_METHODS_FILE" "$DOCUMENTED_METHODS_FILE" > "$EXTRA_DOCS_FILE"

# Report results
echo ""
echo "📈 Documentation Coverage Report"
echo "================================"

IMPLEMENTED_COUNT=$(wc -l < "$IMPLEMENTED_METHODS_FILE")
DOCUMENTED_COUNT=$(wc -l < "$DOCUMENTED_METHODS_FILE")
MISSING_COUNT=$(wc -l < "$MISSING_DOCS_FILE")
EXTRA_COUNT=$(wc -l < "$EXTRA_DOCS_FILE")

echo "📊 Statistics:"
echo "  Implemented methods: $IMPLEMENTED_COUNT"
echo "  Documented methods: $DOCUMENTED_COUNT"
echo "  Missing documentation: $MISSING_COUNT"
echo "  Extra documentation: $EXTRA_COUNT"

# Calculate coverage percentage
if [ "$IMPLEMENTED_COUNT" -gt 0 ]; then
    COVERAGE=$(( (DOCUMENTED_COUNT - EXTRA_COUNT) * 100 / IMPLEMENTED_COUNT ))
    echo "  Documentation coverage: ${COVERAGE}%"
else
    COVERAGE=0
    echo "  Documentation coverage: 0%"
fi

# Report missing documentation
if [ "$MISSING_COUNT" -gt 0 ]; then
    echo ""
    print_status "ERROR" "Missing documentation for $MISSING_COUNT methods:"
    while IFS= read -r method; do
        echo "  - $method"
    done < "$MISSING_DOCS_FILE"
fi

# Report extra documentation
if [ "$EXTRA_COUNT" -gt 0 ]; then
    echo ""
    print_status "WARNING" "Extra documentation for $EXTRA_COUNT methods:"
    while IFS= read -r method; do
        echo "  - $method"
    done < "$EXTRA_DOCS_FILE"
fi

# Check for critical missing methods
CRITICAL_MISSING=0
if grep -q "add_external_stream\|remove_external_stream\|get_external_streams\|discover_external_streams" "$MISSING_DOCS_FILE"; then
    print_status "ERROR" "Critical external stream methods are undocumented!"
    CRITICAL_MISSING=1
fi

if grep -q "get_server_info\|get_status\|get_system_status" "$MISSING_DOCS_FILE"; then
    print_status "ERROR" "Critical system status methods are undocumented!"
    CRITICAL_MISSING=1
fi

# Validate API documentation structure
echo ""
echo "🔍 Validating API documentation structure..."

# Check for required sections
REQUIRED_SECTIONS=(
    "## JSON-RPC 2.0 Compliance"
    "## Connection"
    "## Authentication"
    "## Core Methods"
    "### Standard Error Response Format"
)

MISSING_SECTIONS=0
for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -q "$section" docs/api/json_rpc_methods.md; then
        print_status "ERROR" "Missing required section: $section"
        MISSING_SECTIONS=1
    fi
done

# Check for method documentation completeness
echo ""
echo "🔍 Validating method documentation completeness..."

INCOMPLETE_METHODS=0
while IFS= read -r method; do
    if grep -q "### $method" docs/api/json_rpc_methods.md; then
        # Check if method has required subsections (support both formats)
        if ! grep -A 50 "### $method" docs/api/json_rpc_methods.md | grep -q "#### Parameters\|#### Response\|#### Example\|**Parameters:**\|**Returns:**\|**Example:**"; then
            print_status "WARNING" "Method $method documentation is incomplete"
            INCOMPLETE_METHODS=1
        fi
    fi
done < "$DOCUMENTED_METHODS_FILE"

# Generate validation report
echo ""
echo "📋 Validation Summary"
echo "===================="

if [ "$COVERAGE" -ge 90 ]; then
    print_status "SUCCESS" "Documentation coverage is excellent ($COVERAGE%)"
elif [ "$COVERAGE" -ge 80 ]; then
    print_status "WARNING" "Documentation coverage is good ($COVERAGE%)"
else
    print_status "ERROR" "Documentation coverage is poor ($COVERAGE%)"
fi

if [ "$MISSING_SECTIONS" -eq 0 ]; then
    print_status "SUCCESS" "All required sections are present"
else
    print_status "ERROR" "Missing required sections"
fi

if [ "$INCOMPLETE_METHODS" -eq 0 ]; then
    print_status "SUCCESS" "All documented methods are complete"
else
    print_status "WARNING" "Some method documentation is incomplete"
fi

# Determine overall status
if [ "$COVERAGE" -ge 90 ] && [ "$MISSING_SECTIONS" -eq 0 ] && [ "$CRITICAL_MISSING" -eq 0 ]; then
    print_status "SUCCESS" "Documentation validation passed!"
    exit 0
elif [ "$COVERAGE" -ge 80 ] && [ "$CRITICAL_MISSING" -eq 0 ]; then
    print_status "WARNING" "Documentation validation passed with warnings"
    exit 0
else
    print_status "ERROR" "Documentation validation failed!"
    exit 1
fi
