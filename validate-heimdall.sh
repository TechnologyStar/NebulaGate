#!/bin/bash

# Simple syntax validation for Go files
echo "Checking Go file syntax..."

# Check for basic syntax issues in the main Heimdall files
files=(
    "middleware/heimdall.go"
    "middleware/heimdall_config.go"
    "middleware/heimdall_test.go"
    "middleware/heimdall_integration_test.go"
    "router/heimdall-relay-router.go"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "Checking $file..."
        
        # Basic syntax checks
        if grep -q "package middleware" "$file" 2>/dev/null || grep -q "package router" "$file" 2>/dev/null; then
            echo "  ✓ Package declaration found"
        else
            echo "  ✗ Package declaration missing"
        fi
        
        # Check for balanced braces
        open_braces=$(grep -o '{' "$file" | wc -l)
        close_braces=$(grep -o '}' "$file" | wc -l)
        if [ "$open_braces" -eq "$close_braces" ]; then
            echo "  ✓ Braces balanced"
        else
            echo "  ✗ Braces not balanced: $open_braces open, $close_braces close"
        fi
        
        # Check for import statements
        if grep -q 'import(' "$file" || grep -q 'import "' "$file"; then
            echo "  ✓ Import statements found"
        else
            echo "  ⚠ No import statements (may be intentional)"
        fi
        
    else
        echo "File $file not found"
    fi
    echo ""
done

echo "Syntax validation complete."
echo ""
echo "Files created:"
echo "- middleware/heimdall.go (main Heimdall middleware)"
echo "- middleware/heimdall_config.go (configuration management)"
echo "- middleware/heimdall_test.go (unit tests)"
echo "- middleware/heimdall_integration_test.go (integration tests)"
echo "- router/heimdall-relay-router.go (Heimdall-enabled router)"
echo "- .env.heimdall.example (configuration example)"
echo "- docs/HEIMDALL.md (comprehensive documentation)"
echo ""
echo "Integration points:"
echo "- Updated main.go to initialize Heimdall configuration"
echo "- Updated router/main.go to use Heimdall when enabled"
echo "- Compatible with existing token validation system"
echo "- Uses existing Redis infrastructure"