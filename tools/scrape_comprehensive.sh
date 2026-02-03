#!/bin/bash
# Comprehensive API Documentation Scraper
# Supports both Hyperping.com and Notion pages

set -e

echo "🔍 Installing dependencies..."

# Install Python dependencies
pip3 install -q playwright markdown-it-py html2text

# Install Playwright browsers
playwright install chromium

# Install Notion scraper
pip3 install -q notion-exporter

echo "✅ Dependencies installed"
echo ""

# Run the scraper
python3 "$(dirname "$0")/scrape_api_docs.py"

echo ""
echo "📚 Converting to markdown..."

# Convert JSON output to markdown
python3 << 'PYTHON_SCRIPT'
import json
import os
from pathlib import Path

output_dir = Path("docs_scraped")
markdown_dir = output_dir / "markdown"
markdown_dir.mkdir(exist_ok=True)

for json_file in output_dir.glob("*.json"):
    with open(json_file) as f:
        data = json.load(f)

    md_file = markdown_dir / f"{json_file.stem}.md"
    with open(md_file, 'w') as f:
        f.write(f"# {data['title']}\n\n")
        f.write(f"**Source:** {data['url']}\n\n")
        f.write("---\n\n")
        f.write(data['text'])

    print(f"✅ Converted {json_file.name} to markdown")

print(f"\n📁 Markdown files saved to: {markdown_dir}")
PYTHON_SCRIPT

echo ""
echo "🎉 Scraping complete!"
echo "📂 Output:"
echo "   - JSON: ./docs_scraped/*.json"
echo "   - Markdown: ./docs_scraped/markdown/*.md"
