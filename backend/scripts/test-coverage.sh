#!/bin/bash
# test-coverage.sh runs backend tests with coverage and generates an HTML report.
# Usage: ./scripts/test-coverage.sh

set -e

cd "$(dirname "$0")/.."

echo "Running tests with coverage..."
go test -race -covermode=atomic -coverprofile=coverage.out ./...

echo ""
echo "Generating coverage summary..."
go tool cover -func=coverage.out | tail -1

echo ""
echo "Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "Coverage report generated: coverage.html"
echo "Open it in your browser to view line-by-line coverage."
