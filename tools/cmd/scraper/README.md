# Hyperping API Documentation Scraper

Monitors Hyperping API documentation for changes and analyzes provider coverage gaps.

## Features

- **Scraping**: Discovers and scrapes all API documentation pages
- **Change Detection**: Compares snapshots to detect API changes
- **Coverage Analysis**: Compares API docs against provider schema
- **GitHub Integration**: Creates issues for API changes and coverage gaps
- **Automation**: Runs daily via GitHub Actions

## Quick Start

```bash
# Build
cd tools/cmd/scraper
go build -o scraper

# Scrape API docs
./scraper

# Analyze provider coverage
./scraper -analyze -provider-dir=../../../internal

# With GitHub integration
export GITHUB_TOKEN=ghp_xxx
./scraper
```

## Commands

| Flag | Description |
|------|-------------|
| `-analyze` | Run coverage analysis instead of scraping |
| `-provider-dir` | Path to provider internal directory |
| `-snapshot-dir` | Path to snapshots directory |
| `-log-level` | Log level: debug, info, warn, error |

## Output

```
snapshots/           # Historical API snapshots
docs_scraped/        # Scraped documentation (JSON)
.scraper_cache.json  # Cache for change detection
```

## GitHub Actions

The scraper runs automatically via `.github/workflows/scraper.yml`:
- **Daily at 2 AM UTC**: Scrapes API docs, detects changes
- **Manual trigger**: Available from Actions tab
- **Creates issues**: When API changes or coverage gaps detected

## Configuration

Edit `config.go` for:
- Base URL, rate limits, timeouts
- Output directories
- Retry settings

## Testing

```bash
go test ./...
```

## License

MPL-2.0
