#!/bin/bash
# Pre-commit CI validation script
# Run this before committing to ensure CI will pass

set -e

echo "=========================================="
echo "Running CI Pipeline Locally with act"
echo "=========================================="
echo ""
echo "This will run all CI jobs defined in .github/workflows/ci.yml"
echo ""

# Check if act is installed
if ! command -v act &> /dev/null; then
    echo "Error: 'act' is not installed."
    echo "Install it with: curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash"
    exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo "Error: Docker is not running."
    echo "Please start Docker and try again."
    exit 1
fi

# Run the CI workflow
# Note: Using catthehacker images which approximate GitHub's hosted runners
# GitHub runners are VMs, not containers, so these are community-maintained approximations
# - runner-latest: Most accurate simulation of GitHub runners (~18GB, includes all pre-installed tools)
# - act-latest: Medium size with common tools (~500MB) - good balance
# --no-cache-server disables the cache server to avoid cache-related errors
echo "Running CI workflow..."
echo ""
echo "Note: This uses Docker images that approximate GitHub's hosted runners."
echo "Some OS-specific tests (macOS, Windows) won't run locally."
echo ""

act push \
    --workflows .github/workflows/ci.yml \
    --platform ubuntu-latest=catthehacker/ubuntu:runner-latest \
    --platform ubuntu-22.04=catthehacker/ubuntu:runner-22.04 \
    --platform ubuntu-20.04=catthehacker/ubuntu:runner-20.04 \
    --no-cache-server \
    --verbose

# Check exit code
if [ $? -eq 0 ]; then
    echo ""
    echo "=========================================="
    echo "✅ All CI checks passed!"
    echo "Safe to commit and push."
    echo "=========================================="
else
    echo ""
    echo "=========================================="
    echo "❌ CI checks failed!"
    echo "Please fix the issues before committing."
    echo "=========================================="
    exit 1
fi
