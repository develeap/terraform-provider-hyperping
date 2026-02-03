#!/usr/bin/env python3
"""
Scrape Hyperping API documentation from JavaScript-rendered pages.
Requires: pip install playwright && playwright install chromium
"""

import asyncio
import json
from playwright.async_api import async_playwright

DOCS_URLS = [
    "https://hyperping.com/docs/api/overview",
    "https://hyperping.com/docs/api/monitors",
    "https://hyperping.com/docs/api/statuspages",
    "https://hyperping.com/docs/api/maintenance",
    "https://hyperping.com/docs/api/incidents",
    "https://hyperping.com/docs/api/outages",
    "https://hyperping.com/docs/api/healthchecks",
    "https://hyperping.com/docs/api/reports",
]

NOTION_URL = "https://hyperping.notion.site/Hyperping-API-documentation-a0dc48fb818e4542a8f7fb4163ede2c3"


async def scrape_page(page, url, output_file):
    """Scrape a single documentation page."""
    print(f"Scraping: {url}")

    try:
        # Navigate and wait for content to load
        await page.goto(url, wait_until="networkidle", timeout=30000)

        # Wait for main content to appear
        await page.wait_for_selector("main, article, .documentation, .notion-page-content", timeout=10000)

        # Extract all text content
        content = await page.evaluate("""
            () => {
                // Try different content selectors
                const main = document.querySelector('main') ||
                             document.querySelector('article') ||
                             document.querySelector('.documentation') ||
                             document.querySelector('.notion-page-content') ||
                             document.body;

                // Remove script and style tags
                const scripts = main.querySelectorAll('script, style');
                scripts.forEach(s => s.remove());

                return {
                    text: main.innerText,
                    html: main.innerHTML,
                    title: document.title
                };
            }
        """)

        # Save to file
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump({
                'url': url,
                'title': content['title'],
                'text': content['text'],
                'html': content['html']
            }, f, indent=2, ensure_ascii=False)

        print(f"✅ Saved to {output_file}")
        return content['text']

    except Exception as e:
        print(f"❌ Error scraping {url}: {e}")
        return None


async def main():
    """Scrape all documentation pages."""
    async with async_playwright() as p:
        # Launch browser
        browser = await p.chromium.launch(headless=True)
        context = await browser.new_context(
            user_agent="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"
        )
        page = await context.new_page()

        # Scrape each page
        results = {}
        for url in DOCS_URLS:
            # Extract page name from URL
            page_name = url.split("/")[-1]
            output_file = f"docs_scraped/{page_name}.json"

            content = await scrape_page(page, url, output_file)
            if content:
                results[page_name] = content

        # Scrape Notion page separately (might need different wait)
        print(f"\nScraping Notion page...")
        await page.goto(NOTION_URL, wait_until="networkidle", timeout=30000)
        await page.wait_for_timeout(3000)  # Extra wait for Notion

        notion_content = await page.evaluate("document.body.innerText")
        with open("docs_scraped/notion_api_docs.txt", 'w', encoding='utf-8') as f:
            f.write(notion_content)
        print(f"✅ Saved Notion docs to docs_scraped/notion_api_docs.txt")

        await browser.close()

        # Summary
        print(f"\n{'='*60}")
        print(f"Scraped {len(results)} pages successfully")
        print(f"Output directory: ./docs_scraped/")
        print(f"{'='*60}")


if __name__ == "__main__":
    # Create output directory
    import os
    os.makedirs("docs_scraped", exist_ok=True)

    # Run scraper
    asyncio.run(main())
