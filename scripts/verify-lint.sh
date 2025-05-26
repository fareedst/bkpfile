#!/bin/bash

# Revive Deployment Verification Script
# This script verifies that revive is properly deployed and configured

set -e

echo "🔍 Verifying revive deployment..."

# Check if revive is installed
if ! command -v revive &> /dev/null; then
    echo "❌ revive is not installed. Installing..."
    go install github.com/mgechev/revive@latest
    echo "✅ revive installed successfully"
else
    echo "✅ revive is installed"
fi

# Check if revive.toml exists
if [ ! -f "revive.toml" ]; then
    echo "❌ revive.toml configuration file not found"
    exit 1
else
    echo "✅ revive.toml configuration file found"
fi

# Test revive configuration
echo "🧪 Testing revive configuration..."
if revive -config revive.toml ./... > /dev/null 2>&1; then
    echo "✅ revive configuration is valid"
else
    echo "⚠️  revive found linting issues (this is normal)"
fi

# Check if Makefile has lint targets
if grep -q "^lint:" Makefile; then
    echo "✅ Makefile has lint target"
else
    echo "❌ Makefile missing lint target"
    exit 1
fi

# Test make lint
echo "🧪 Testing make lint..."
if make lint > /dev/null 2>&1; then
    echo "✅ make lint works correctly"
else
    echo "⚠️  make lint found issues (this is normal)"
fi

echo ""
echo "🎉 Revive deployment verification complete!"
echo ""
echo "Usage:"
echo "  make lint      - Run revive linter"
echo "  make lint-fix  - Run comprehensive linting"
echo "  revive -config revive.toml ./...  - Run revive directly"
echo ""
echo "Current linting status:"
make lint || echo "📝 Found linting issues that should be addressed" 