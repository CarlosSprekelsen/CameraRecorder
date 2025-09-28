#!/bin/bash

# Pre-commit hook for documentation validation
# Ensures documentation is updated when code changes

set -e

echo "🔍 Running pre-commit documentation validation..."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository"
    exit 1
fi

# Get list of changed files
CHANGED_FILES=$(git diff --cached --name-only)

# Check if any WebSocket-related files changed
WEBSOCKET_CHANGED=false
if echo "$CHANGED_FILES" | grep -q "internal/websocket/"; then
    WEBSOCKET_CHANGED=true
fi

# Check if any documentation files changed
DOCS_CHANGED=false
if echo "$CHANGED_FILES" | grep -q "docs/"; then
    DOCS_CHANGED=true
fi

# If WebSocket code changed but docs didn't, warn user
if [ "$WEBSOCKET_CHANGED" = true ] && [ "$DOCS_CHANGED" = false ]; then
    echo "⚠️  WebSocket code changed but documentation wasn't updated"
    echo "   Please consider updating documentation for any new methods or changes"
    echo "   Documentation files: docs/api/json_rpc_methods.md"
    echo ""
    echo "   Changed files:"
    echo "$CHANGED_FILES" | grep "internal/websocket/" | sed 's/^/     /'
    echo ""
    echo "   To skip this check, use: git commit --no-verify"
    echo "   To update documentation, edit: docs/api/json_rpc_methods.md"
    echo ""
    read -p "Continue with commit? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Commit aborted"
        exit 1
    fi
fi

# If documentation changed, validate it
if [ "$DOCS_CHANGED" = true ]; then
    echo "📚 Documentation files changed, validating..."
    
    # Run documentation validation
    if [ -f "scripts/validate-documentation.sh" ]; then
        chmod +x scripts/validate-documentation.sh
        if ./scripts/validate-documentation.sh; then
            echo "✅ Documentation validation passed"
        else
            echo "❌ Documentation validation failed"
            echo "   Please fix documentation issues before committing"
            exit 1
        fi
    else
        echo "⚠️  Documentation validation script not found"
        echo "   Skipping validation..."
    fi
fi

# Check for new methods that need documentation
if [ "$WEBSOCKET_CHANGED" = true ]; then
    echo "🔍 Checking for new methods that need documentation..."
    
    # Extract new methods from staged changes
    NEW_METHODS=$(git diff --cached internal/websocket/ | grep -o 'Method.*func' | grep -o '"[^"]*"' | sed 's/"//g' | sort | uniq)
    
    if [ -n "$NEW_METHODS" ]; then
        echo "📝 New methods detected:"
        echo "$NEW_METHODS" | sed 's/^/     /'
        echo ""
        echo "   Please ensure these methods are documented in:"
        echo "   docs/api/json_rpc_methods.md"
        echo ""
        echo "   Required sections for each method:"
        echo "   - ### method_name"
        echo "   - #### Parameters"
        echo "   - #### Response"
        echo "   - #### Example"
    fi
fi

echo "✅ Pre-commit documentation validation completed"
