# MVP Scraper - Quick Start Guide

## What This Does

Scrapes **all 50 Hyperping API documentation pages** and saves them as JSON files.

**Features**:
- ✅ All 50 URLs hardcoded (no discovery needed)
- ✅ Sequential scraping (simple, reliable)
- ✅ Rate limiting (1 req/sec)
- ✅ Retry logic (3 attempts with exponential backoff)
- ✅ Progress tracking
- ✅ Error handling

## Quick Run

```bash
cd /home/khaleds/projects/terraform-provider-hyperping/tools/cmd/scraper
go run scraper_mvp.go
```

## Expected Output

```
🚀 Hyperping API Documentation Scraper - MVP
📊 Total URLs to scrape: 50

🌐 Launching browser...

[1/50] 📄 https://hyperping.com/docs/api/overview
  ✅ Saved to overview.json (2341 chars)

[2/50] 📄 https://hyperping.com/docs/api/monitors
  ✅ Saved to monitors.json (1823 chars)

[3/50] 📄 https://hyperping.com/docs/api/monitors/list
  ✅ Saved to monitors_list.json (4567 chars)

...

============================================================
✅ Scraping Complete!
⏱️  Duration: 2m15s
📊 Success: 50/50
📂 Output: docs_scraped_mvp/
============================================================
```

## Output Structure

```
docs_scraped_mvp/
├── overview.json
├── monitors.json
├── monitors_list.json
├── monitors_get.json
├── monitors_create.json
├── monitors_update.json
├── monitors_delete.json
├── statuspages.json
├── statuspages_list.json
├── statuspages_get.json
├── statuspages_create.json
├── statuspages_update.json
├── statuspages_delete.json
├── statuspages_subscribers_list.json
├── statuspages_subscribers_create.json
├── statuspages_subscribers_delete.json
├── maintenance.json
├── maintenance_list.json
├── ... (50 total files)
```

## JSON Format

Each file contains:
```json
{
  "url": "https://hyperping.com/docs/api/monitors/create",
  "title": "Create Monitor - Hyperping API",
  "text": "Full text content...",
  "html": "<html>Full HTML...</html>",
  "timestamp": "2026-02-03T12:34:56Z"
}
```

## Verification

Check that all 50 pages were scraped:
```bash
ls docs_scraped_mvp/*.json | wc -l
# Should output: 50
```

Check file sizes:
```bash
ls -lh docs_scraped_mvp/*.json
```

## Expected Time

- **First run**: 2-3 minutes (downloads Chromium ~132 MB)
- **Subsequent runs**: 2-3 minutes

## Troubleshooting

### Browser download fails
```bash
# Manually trigger browser download
go run scraper_mvp.go
# Browser downloads automatically on first run
```

### Some pages fail
- Check internet connection
- Check if Hyperping docs are accessible
- Review error messages in output
- Failed pages will be retried 3 times automatically

### Memory issues
- MVP uses ~1.2 GB RAM
- Close other applications if needed

## Next Steps

After verifying all 50 pages scrape successfully:

1. **Validate content** - Spot check a few JSON files
2. **Move to Plan A** - Implement dynamic URL discovery
3. **Add to scrape.sh** - Integrate with existing workflow

## Files

- `scraper_mvp.go` - Main scraper (150 lines)
- `MVP_README.md` - This file
- `ARCHITECTURE_PLANS.md` - Full architecture documentation
