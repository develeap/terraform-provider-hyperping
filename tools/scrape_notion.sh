#!/bin/bash
# Scrape Notion API Documentation Page
# Uses notion-snapshot (doesn't require API token for public pages)

set -e

NOTION_URL="https://hyperping.notion.site/Hyperping-API-documentation-a0dc48fb818e4542a8f7fb4163ede2c3"
OUTPUT_DIR="docs_scraped/notion"

echo "🔍 Scraping Notion page..."
echo "URL: $NOTION_URL"
echo ""

# Method 1: Using Playwright with extra wait time for Notion
python3 << 'PYTHON_SCRIPT'
import asyncio
from playwright.async_api import async_playwright
import os

NOTION_URL = "https://hyperping.notion.site/Hyperping-API-documentation-a0dc48fb818e4542a8f7fb4163ede2c3"
OUTPUT_DIR = "docs_scraped/notion"

async def scrape_notion():
    os.makedirs(OUTPUT_DIR, exist_ok=True)

    async with async_playwright() as p:
        print("🚀 Launching browser...")
        browser = await p.chromium.launch(headless=True)
        context = await browser.new_context(
            user_agent="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"
        )
        page = await context.new_page()

        print("📄 Loading Notion page...")
        await page.goto(NOTION_URL, wait_until="networkidle", timeout=60000)

        # Wait for Notion to fully render
        print("⏳ Waiting for content to render...")
        await page.wait_for_selector(".notion-page-content, .notion-frame", timeout=30000)
        await page.wait_for_timeout(5000)  # Extra wait for lazy-loaded content

        # Extract content
        print("📝 Extracting content...")
        content = await page.evaluate("""
            () => {
                const removeElements = (selector) => {
                    document.querySelectorAll(selector).forEach(el => el.remove());
                };

                // Remove Notion UI elements
                removeElements('script, style, .notion-topbar, .notion-sidebar');

                // Get main content
                const main = document.querySelector('.notion-page-content') ||
                             document.querySelector('.notion-frame') ||
                             document.body;

                return {
                    text: main.innerText,
                    html: main.innerHTML,
                    title: document.title
                };
            }
        """)

        # Save as text
        with open(f"{OUTPUT_DIR}/api_docs.txt", 'w', encoding='utf-8') as f:
            f.write(content['text'])

        # Save as HTML
        with open(f"{OUTPUT_DIR}/api_docs.html", 'w', encoding='utf-8') as f:
            f.write(f"<h1>{content['title']}</h1>\n")
            f.write(content['html'])

        await browser.close()

        print(f"\n✅ Saved to {OUTPUT_DIR}/")
        print(f"   - api_docs.txt (plain text)")
        print(f"   - api_docs.html (HTML)")

asyncio.run(scrape_notion())
PYTHON_SCRIPT

# Convert HTML to Markdown
echo ""
echo "📚 Converting to Markdown..."
python3 << 'PYTHON_SCRIPT'
import html2text
import os

OUTPUT_DIR = "docs_scraped/notion"

# Read HTML
with open(f"{OUTPUT_DIR}/api_docs.html", 'r', encoding='utf-8') as f:
    html_content = f.read()

# Convert to markdown
h = html2text.HTML2Text()
h.ignore_links = False
h.ignore_images = False
markdown = h.handle(html_content)

# Save markdown
with open(f"{OUTPUT_DIR}/api_docs.md", 'w', encoding='utf-8') as f:
    f.write(markdown)

print(f"✅ Converted to Markdown: {OUTPUT_DIR}/api_docs.md")
PYTHON_SCRIPT

echo ""
echo "🎉 Done!"
