# API Documentation Scraping Tools

Comprehensive tools for scraping Hyperping API documentation from JavaScript-rendered pages.

## Tools Created

| Tool | Purpose | Tech Stack |
|------|---------|------------|
| `scrape_api_docs.py` | Scrapes all Hyperping.com API docs | Python + Playwright |
| `scrape_notion.sh` | Scrapes Notion API documentation | Playwright + html2text |
| `scrape_comprehensive.sh` | All-in-one scraper with conversion | Orchestrates both |

## Quick Start

### Option 1: Scrape Everything (Recommended)

```bash
cd tools
./scrape_comprehensive.sh
```

**Output:**
- `docs_scraped/*.json` - Raw JSON data
- `docs_scraped/markdown/*.md` - Converted markdown files

### Option 2: Notion Page Only

```bash
cd tools
./scrape_notion.sh
```

**Output:**
- `docs_scraped/notion/api_docs.txt` - Plain text
- `docs_scraped/notion/api_docs.html` - HTML
- `docs_scraped/notion/api_docs.md` - Markdown

### Option 3: Python Script Directly

```bash
cd tools
pip install playwright
playwright install chromium
python3 scrape_api_docs.py
```

## Installation

### Prerequisites

```bash
# Python 3.8+
python3 --version

# pip
pip3 --version
```

### Install Dependencies

```bash
# Install Python packages
pip3 install playwright html2text markdown-it-py

# Install Playwright browsers
playwright install chromium
```

## Research Findings

Based on extensive research, here are the key technologies:

### 1. **Playwright vs Puppeteer (2026)**

**Winner: Playwright** for our use case because:
- ✅ Cross-browser support (Chromium, Firefox, WebKit)
- ✅ Built-in auto-waiting for elements
- ✅ Better parallel execution
- ✅ Modern API with TypeScript support
- ✅ Active development by Microsoft

**Sources:**
- [Playwright vs Puppeteer Comparison](https://www.browsercat.com/post/playwright-vs-puppeteer-web-scraping-comparison)
- [Playwright Web Scraping Guide 2026](https://www.browserstack.com/guide/playwright-web-scraping)
- [Research Comparison 2026](https://research.aimultiple.com/playwright-vs-puppeteer/)

### 2. **Scraping JavaScript-Rendered Sites**

Modern docs sites use React/Next.js. Key techniques:

**Next.js Data Extraction:**
- Check for `__NEXT_DATA__` script tag (contains pre-rendered JSON)
- Extract without headless browser if available

**React Crawling:**
- Use headless browser to execute JavaScript
- Wait for content to render before extraction

**Sources:**
- [Scraping Next.js Sites](https://brightdata.com/blog/how-tos/web-scraping-with-next-js)
- [React Crawling Guide](https://www.zenrows.com/blog/react-crawling)
- [Scraping without Headless Browser](https://peterrauscher.com/posts/the-secret-to-efficiently-scraping-react-apps-without-a-headless-browser/)

### 3. **Notion Public Pages**

Notion requires special handling:

**Tools Available:**
- `notion-exporter` (CLI) - For API-based export
- `notion4ever` (Python) - Official API with nested pages
- Custom scraping - For public pages without API token

**Challenges:**
- Notion loads content dynamically with lazy-loading
- Requires 3-5 second wait after page load
- Content is in `.notion-page-content` or `.notion-frame`

**Sources:**
- [Notion Exporter (PyPI)](https://pypi.org/project/notion-exporter/)
- [notion4ever GitHub](https://github.com/MerkulovDaniil/notion4ever)
- [Reverse Engineering Notion API](https://blog.kowalczyk.info/article/88aee8f43620471aa9dbcad28368174c/how-i-reverse-engineered-notion-api.html)

## How the Scripts Work

### 1. `scrape_api_docs.py`

**Flow:**
```
1. Launch headless Chromium browser
2. For each URL:
   - Navigate to page
   - Wait for network idle (JS executed)
   - Wait for content selector
   - Extract text + HTML via JavaScript
   - Save as JSON
3. Close browser
```

**Key Features:**
- Waits for `networkidle` (all network requests complete)
- Removes `<script>` and `<style>` tags
- Saves both text and HTML for flexibility

### 2. `scrape_notion.sh`

**Flow:**
```
1. Launch browser
2. Navigate to Notion page
3. Wait for .notion-page-content to appear
4. Extra 5-second wait (lazy-loading)
5. Remove Notion UI elements
6. Extract content
7. Convert HTML → Markdown
```

**Key Features:**
- Extended timeout for Notion (60s vs 30s)
- Removes Notion topbar/sidebar
- Uses html2text for clean Markdown conversion

## Output Structure

```
docs_scraped/
├── overview.json              # Base API info
├── monitors.json              # Monitor endpoints
├── statuspages.json          # Status page endpoints
├── maintenance.json          # Maintenance endpoints
├── incidents.json            # Incident endpoints
├── outages.json              # Outage endpoints
├── healthchecks.json         # Healthcheck endpoints
├── reports.json              # Reporting endpoints
├── markdown/
│   ├── overview.md
│   ├── monitors.md
│   └── ...
└── notion/
    ├── api_docs.txt
    ├── api_docs.html
    └── api_docs.md
```

## Troubleshooting

### Browser Not Found

```bash
playwright install chromium
```

### Timeout Errors

Increase timeout in script:
```python
await page.goto(url, timeout=60000)  # 60 seconds
```

### Notion Content Not Loading

Increase wait time:
```python
await page.wait_for_timeout(10000)  # 10 seconds
```

### Empty Output

Check selectors in browser DevTools:
```javascript
// In browser console
document.querySelector('.notion-page-content')
```

## Advanced Usage

### Custom Selectors

Edit `scrape_api_docs.py`:
```python
await page.wait_for_selector("YOUR_SELECTOR", timeout=10000)
```

### Save Screenshots

Add to script:
```python
await page.screenshot(path="debug.png")
```

### Extract Specific Elements

```python
content = await page.evaluate("""
    () => document.querySelector('.api-endpoint').innerText
""")
```

## Best Practices

1. **Respect Rate Limits**: Add delays between requests
2. **User-Agent**: Use realistic user-agent strings
3. **Headless Mode**: Use `headless=False` for debugging
4. **Error Handling**: Wrap in try/except for robustness
5. **Incremental**: Save after each page (in case of failure)

## Alternative Tools

If these scripts don't work, try:

1. **Browser Extension**: [Web Scraper](https://webscraper.io/)
2. **Cloud Service**: [ScrapingBee](https://www.scrapingbee.com/), [ScraperAPI](https://www.scraperapi.com/)
3. **Manual Export**: Browser DevTools → Copy outerHTML
4. **Notion API**: For Notion pages with API token

## Next Steps

After scraping:
1. Parse the markdown/JSON for API endpoints
2. Extract schemas, examples, constraints
3. Update provider implementation
4. Update tests with real examples
5. Update documentation

## Contributing

Improvements welcome:
- Better selectors for specific sites
- Parallel page loading
- Retry logic for failures
- Export to OpenAPI/Swagger format
