#!/bin/bash
# Check if annotations.json is up to date
# This script regenerates the annotations and compares with the committed file

set -e

echo "Checking if annotations.json is up to date..."

# Build typemux binary if it doesn't exist
if [ ! -f "./typemux" ]; then
    echo "Building typemux binary..."
    go build -o typemux ./cmd/typemux
fi

# Generate fresh annotations
./typemux annotations > annotations.json.tmp

# Compare with committed file
if ! diff -q annotations.json annotations.json.tmp > /dev/null 2>&1; then
    echo ""
    echo "❌ ERROR: annotations.json is out of date!"
    echo ""
    echo "The committed annotations.json does not match the current registry."
    echo "Please regenerate it with:"
    echo ""
    echo "  ./typemux annotations > annotations.json"
    echo ""
    echo "Then commit the updated file."
    echo ""
    echo "Differences:"
    diff annotations.json annotations.json.tmp || true
    rm -f annotations.json.tmp
    exit 1
fi

rm -f annotations.json.tmp
echo "✅ annotations.json is up to date"
