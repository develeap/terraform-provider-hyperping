# Hyperping API Documentation Scraper

Automated tool for monitoring changes to Hyperping's API documentation with intelligent diff detection and GitHub integration.

## What It Does

1. **Discovers** all API documentation pages by crawling navigation
2. **Scrapes** each page using headless browser (JavaScript support)
3. **Caches** content to skip unchanged pages on subsequent runs
4. **Extracts** API parameters from HTML documentation
5. **Compares** with previous snapshot to detect changes
6. **Generates** detailed diff reports for breaking changes
7. **Creates** GitHub issues automatically (optional)

## Quick Start

```bash
# Clone and navigate
cd tools/cmd/scraper

# First run (downloads Chromium browser ~132 MB)
go run main.go

# Expected output:
# - docs_scraped/*.json       (scraped pages)
# - snapshots/2026-02-03.../  (versioned snapshot)
# - api_changes_*.md          (diff report if changes found)
```

### First Run Output
```
🚀 Hyperping API Documentation Scraper - Plan A
============================================================
📋 Configuration:
   Base URL: https://hyperping.com/docs/api
   Output Dir: docs_scraped
   Rate Limit: 1.0 req/sec
   Timeout: 30s
   Retries: 3

🌐 Launching browser...
✅ Browser connected

============================================================
🔍 Stage 1: Discovering parent API sections...
   Found 8 parent sections

🔍 Stage 2: Discovering child method pages...
   [1/8] Checking monitors for child pages...
      Found 6 child pages
   ...

📊 Discovery Summary:
   Total URLs: 42
   By section:
      monitors: 7
      incidents: 6
      maintenance: 5
      ...

============================================================
🔄 Starting smart scraping...

[1/42] https://hyperping.com/docs/api/overview
        Section: overview, Method:
        ✅ Scraped (2341 chars) - CHANGED

[2/42] https://hyperping.com/docs/api/monitors
        Section: monitors, Method:
        ⏭️  Unchanged (1823 chars) - SKIPPED
...

============================================================
📊 Scraping Complete!
   Total URLs: 42
   Scraped: 15 (content changed)
   Skipped: 27 (unchanged)
   Failed: 0
   Duration: 1m 23s
   Time Savings: ~64%

✅ Done!
```

## Installation

### Prerequisites

- **Go 1.22+** - [Download](https://go.dev/dl/)
- **Internet connection** - For downloading Chromium and scraping
- **2 GB RAM** - Browser automation is memory-intensive
- **500 MB disk space** - For snapshots

### Setup

```bash
# Install dependencies
go mod download

# Verify setup
go build -o scraper
./scraper --version
```

## Configuration

### Environment Variables (Optional)

For automated GitHub issue creation:

```bash
# Option 1: Combined
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GITHUB_REPOSITORY="develeap/terraform-provider-hyperping"

# Option 2: Separate
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GITHUB_OWNER="develeap"
export GITHUB_REPO="terraform-provider-hyperping"
```

**Obtaining GitHub Token:**
1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Required permissions: `repo`, `issues:write`
4. Copy token (starts with `ghp_`)

### Default Settings

Configured in `config.go`:

| Setting | Default | Description |
|---------|---------|-------------|
| Base URL | `https://hyperping.com/docs/api` | Starting point for discovery |
| Output Dir | `docs_scraped/` | Where JSON files are saved |
| Cache File | `.scraper_cache.json` | Change detection cache |
| Rate Limit | 1.0 req/sec | Respectful to Hyperping servers |
| Timeout | 30 seconds | Per-page timeout |
| Retries | 3 attempts | With exponential backoff |
| Headless | `true` | Run browser in background |

### Customization

To change defaults, edit `config.go` and rebuild:

```go
const (
    DefaultRateLimit = 2.0  // Faster scraping (use cautiously)
    DefaultTimeout   = 60 * time.Second  // Longer timeout
)
```

## Usage Examples

### Basic Usage

```bash
# Run scraper
go run main.go

# Check for errors
echo $?  # 0 = success
```

### GitHub Integration

```bash
# With GitHub issue creation
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GITHUB_REPOSITORY="owner/repo"
go run main.go

# Expected output (if changes detected):
# ✅ GitHub Issue Created: #42
#    URL: https://github.com/owner/repo/issues/42
```

### Scheduled Runs (CI/CD)

```yaml
# .github/workflows/scrape-api.yml
name: Monitor API Changes
on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours
  workflow_dispatch:  # Manual trigger

jobs:
  scrape:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Run scraper
        working-directory: tools/cmd/scraper
        run: go run main.go
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GITHUB_REPOSITORY: ${{ github.repository }}

      - name: Upload artifacts
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: scraper-output
          path: |
            tools/cmd/scraper/docs_scraped/
            tools/cmd/scraper/api_changes_*.md
```

## Output Files

### docs_scraped/
Scraped documentation in JSON format:

```json
{
  "url": "https://hyperping.com/docs/api/monitors/create",
  "title": "Create Monitor - Hyperping API",
  "text": "Full text content...",
  "html": "<html>Full HTML...</html>",
  "timestamp": "2026-02-03T12:34:56Z"
}
```

### snapshots/
Timestamped snapshots for version history:

```
snapshots/
├── 2026-02-03_10-30-00/  (previous)
│   ├── monitors.json
│   ├── monitors_list.json
│   └── ...
└── 2026-02-03_14-51-20/  (current)
    ├── monitors.json
    ├── monitors_list.json
    └── ...
```

### api_changes_*.md
Diff reports with actionable items:

```markdown
# API Changes Detected

**Date:** 2026-02-03 14:51:20 IST
**Summary:** 3 endpoint(s) changed (1 breaking)

⚠️ **WARNING: Breaking changes detected!**

## Monitors API

### Create

⚠️ **BREAKING CHANGE**

#### ❌ Removed Parameters
- **timeout** (integer, required)

## Action Items
- [ ] Review all changes
- [ ] Update Terraform provider schema
- [ ] Update tests
- [ ] ⚠️ Plan migration strategy
```

## Troubleshooting

### Common Issues

#### "Failed to launch browser"
**Cause:** Chromium not installed

**Solution:**
```bash
# Let scraper auto-download (happens on first run)
go run main.go

# Or manually install
go get -u github.com/go-rod/rod
```

#### "Failed to load cache"
**Cause:** Corrupted cache file

**Solution:**
```bash
# Delete cache and re-scrape
rm .scraper_cache.json
go run main.go  # Will re-scrape all pages
```

#### "GitHub API error: 401"
**Cause:** Invalid or missing GITHUB_TOKEN

**Solution:**
1. Check token is valid: https://github.com/settings/tokens
2. Ensure token has `repo` and `issues:write` permissions
3. Re-export environment variable

#### "Rate limit exceeded"
**Cause:** Too many requests to Hyperping

**Solution:**
- Wait 10 minutes and retry
- Scraper will auto-retry with backoff
- Consider increasing `DefaultRateLimit` in config

#### Build Error: "main redeclared"
**Cause:** `test_github_preview.go` conflicts with `main.go`

**Solution:**
```bash
# Temporary fix: rename test file
mv test_github_preview.go test_github_preview.go.bak

# Or build specific file
go run main.go cache.go config.go discovery.go differ.go github.go models.go snapshots.go
```

### Getting Help

1. Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for detailed solutions
2. Review logs in `scraper.log`
3. Open an issue with:
   - Command you ran
   - Full error output
   - Your Go version: `go version`

## Architecture

### High-Level Flow

```
┌─────────────┐
│  Discovery  │  Find all API page URLs
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Cache      │  Skip unchanged pages
│  Check      │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Scraping   │  Download with headless browser
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Extraction  │  Parse API parameters from HTML
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Diffing    │  Compare with previous snapshot
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Snapshot   │  Save timestamped version
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  GitHub     │  Create issue if breaking changes
│  Issue      │
└─────────────┘
```

### Directory Structure

```
tools/cmd/scraper/
├── main.go              # Entry point & orchestration
├── config.go            # Configuration constants
├── discovery.go         # URL crawling logic
├── cache.go             # Change detection
├── differ.go            # Semantic diff comparison
├── snapshots.go         # Snapshot management
├── github.go            # GitHub API integration
├── models.go            # Data structures
├── extractor/           # HTML parsing
│   ├── extractor.go     # Main extraction logic
│   ├── hyperping.go     # Hyperping-specific parsing
│   ├── tables.go        # Table extraction
│   ├── json.go          # JSON example extraction
│   └── filters.go       # Content filtering
└── utils/               # Helper functions
    ├── files.go         # File operations
    └── hash.go          # Content hashing
```

## Advanced Usage

### Manual Snapshot Comparison

```bash
# List snapshots
ls -lt snapshots/

# Compare specific snapshots
# (Requires code modification - see snapshots.go:107)
```

### Rebuilding Cache from Disk

```go
// If cache is corrupted, rebuild from existing files
cache, err := BuildCacheFromDisk("docs_scraped")
if err != nil {
    log.Fatal(err)
}
SaveCache(".scraper_cache.json", cache)
```

### Testing Locally

```bash
# Test scraper without GitHub integration
unset GITHUB_TOKEN
go run main.go

# Dry run mode (future feature)
# go run main.go --dry-run
```

## Performance

### Resource Usage
- **Memory:** ~1.2 GB (headless browser)
- **CPU:** 2 cores recommended
- **Network:** ~10 MB download per run (if all pages changed)
- **Disk:** ~50 MB per snapshot

### Run Times
- **First run:** 2-3 minutes (downloads browser + scrapes all pages)
- **Incremental run:** 30-60 seconds (with caching)
- **No changes:** 20-30 seconds (cache hits)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

### Key Areas for Contribution

1. **Extraction Logic** - Improve parameter parsing accuracy
2. **Error Handling** - Better error messages and recovery
3. **Testing** - Add unit tests and test fixtures
4. **Documentation** - Improve this README and add examples

## License

Part of [terraform-provider-hyperping](https://github.com/develeap/terraform-provider-hyperping) project.

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - Detailed design decisions
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common errors and solutions
- [CONFIGURATION.md](CONFIGURATION.md) - All configuration options
- [DOCUMENTATION_AUDIT.md](DOCUMENTATION_AUDIT.md) - Documentation quality report

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

---

**Questions?** Open an issue or contact the maintainers.
