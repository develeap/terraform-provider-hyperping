#!/bin/bash
# Run the Go-based API documentation scraper

set -e

SCRAPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/cmd/scraper"

echo "📦 Installing dependencies..."
cd "$SCRAPER_DIR"
go mod download

echo ""
echo "🔨 Building scraper..."
go build -o scraper main.go

echo ""
echo "🚀 Running scraper..."
./scraper

echo ""
echo "📚 Converting to Markdown..."
cd ../../..
go run ./tools/cmd/scraper/converter

echo ""
echo "✅ Done! Check tools/docs_scraped/ for output"
