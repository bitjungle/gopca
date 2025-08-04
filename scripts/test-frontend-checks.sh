#!/bin/bash
#
# test-frontend-checks.sh - Test the frontend validation logic
#
# This script tests the frontend checking logic that will be added to pre-commit hooks

set -e

echo "Testing frontend validation logic..."
echo ""

# Function to check a frontend directory
check_frontend() {
    local dir=$1
    local name=$2
    
    if [ -d "$dir" ] && [ -f "$dir/package.json" ]; then
        echo "Checking $name..."
        
        # Check if node_modules exists (could be local or at workspace root)
        if [ ! -d "$dir/node_modules" ] && [ ! -d "./node_modules" ]; then
            echo "❌ $name: node_modules not found. Run 'npm install' from project root"
            return 1
        fi
        
        # Run TypeScript compilation check
        if [ -f "$dir/tsconfig.json" ]; then
            echo "  → Running TypeScript check..."
            # Use tsconfig.check.json if it exists for type checking only
            if [ -f "$dir/tsconfig.check.json" ]; then
                if ! (cd "$dir" && npx tsc --project tsconfig.check.json 2>&1); then
                    echo "❌ $name: TypeScript compilation failed"
                    echo "Run 'npm run build' in $dir to see detailed errors."
                    return 1
                fi
            else
                if ! (cd "$dir" && npx tsc --noEmit 2>&1); then
                    echo "❌ $name: TypeScript compilation failed"
                    echo "Run 'npm run build' in $dir to see detailed errors."
                    return 1
                fi
            fi
            echo "  ✓ $name TypeScript OK"
        fi
        
        # Run ESLint if configured
        if [ -f "$dir/.eslintrc.json" ] || [ -f "$dir/.eslintrc.js" ]; then
            echo "  → Running ESLint check..."
            if ! (cd "$dir" && npx eslint src --ext .ts,.tsx,.js,.jsx --max-warnings 0 2>&1); then
                echo "❌ $name: ESLint found issues"
                return 1
            fi
            echo "  ✓ $name ESLint OK"
        fi
    else
        echo "⚠️  $name: Directory not found or no package.json"
    fi
    
    return 0
}

# Check each frontend project
FRONTEND_OK=true

echo "1. Checking shared UI components..."
if ! check_frontend "packages/ui-components" "UI Components"; then
    FRONTEND_OK=false
fi

echo ""
echo "2. Checking GoPCA Desktop frontend..."
if ! check_frontend "cmd/gopca-desktop/frontend" "GoPCA Desktop"; then
    FRONTEND_OK=false
fi

echo ""
echo "3. Checking GoCSV frontend..."
if ! check_frontend "cmd/gocsv/frontend" "GoCSV"; then
    FRONTEND_OK=false
fi

echo ""
if [ "$FRONTEND_OK" = true ]; then
    echo "✅ All frontend checks passed!"
    exit 0
else
    echo "❌ Frontend validation failed"
    exit 1
fi