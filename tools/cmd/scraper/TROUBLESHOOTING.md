# Troubleshooting Guide

Common issues and solutions for the Hyperping API Documentation Scraper.

## Table of Contents

- [Build & Compilation Issues](#build--compilation-issues)
- [Browser & Scraping Issues](#browser--scraping-issues)
- [Network & Connectivity Issues](#network--connectivity-issues)
- [GitHub Integration Issues](#github-integration-issues)
- [Cache & State Issues](#cache--state-issues)
- [Performance Issues](#performance-issues)
- [Debugging Tips](#debugging-tips)

---

## Build & Compilation Issues

### ❌ Error: "main redeclared in this block"

**Full Error:**
```
# github.com/develeap/terraform-provider-hyperping/tools/scraper
./test_github_preview.go:10:6: main redeclared in this block
	./main.go:19:6: other declaration of main
```

**Cause:** Multiple files with `func main()` in the same package.

**Solutions:**

**Quick Fix:**
```bash
# Rename conflicting test file
mv test_github_preview.go test_github_preview.go.bak
go run main.go
```

**Proper Fix:**
```bash
# Move to separate directory
mkdir -p cmd/preview
mv test_github_preview.go cmd/preview/
cd cmd/preview
go run test_github_preview.go
```

**Or build specific files:**
```bash
go run main.go cache.go config.go discovery.go differ.go github.go models.go snapshots.go
```

---

### ❌ Error: "undefined: DefaultConfig"

**Cause:** Missing dependency imports or package initialization.

**Solution:**
```bash
# Ensure all files are in the same directory
ls *.go

# Should see: main.go, config.go, cache.go, discovery.go, etc.

# Rebuild module cache
go mod tidy
go build
```

---

### ❌ Error: "cannot find package"

**Full Error:**
```
go: github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor:
    module path mismatch
```

**Cause:** Module path doesn't match directory structure.

**Solution:**
```bash
# Check go.mod
cat go.mod
# Should show: module github.com/develeap/terraform-provider-hyperping/tools/scraper

# If wrong, recreate module
rm go.mod go.sum
go mod init github.com/develeap/terraform-provider-hyperping/tools/scraper
go mod tidy
```

---

## Browser & Scraping Issues

### ❌ Error: "Failed to launch browser"

**Cause:** Chromium browser not downloaded.

**Solution:**

Browser downloads automatically on first run (132 MB). If it fails:

```bash
# Check internet connection
ping google.com

# Check disk space
df -h

# Manually trigger download
go run main.go
# Browser downloads automatically

# Or use playwright CLI
go get -u github.com/go-rod/rod
```

**Advanced:** Set custom browser path
```go
// In main.go, before launcher.New()
launcher := launcher.New().
    Bin("/path/to/chromium").  // Custom browser
    Headless(true).
    MustLaunch()
```

---

### ❌ Error: "Page load timeout"

**Full Error:**
```
❌ Failed after 3 retries: page load timeout: context deadline exceeded
```

**Causes:**
1. Slow network connection
2. Hyperping website is slow/down
3. Timeout too short for complex pages

**Solutions:**

**Increase timeout:**
```go
// In config.go
const DefaultTimeout = 60 * time.Second  // Was 30s
```

**Check network:**
```bash
# Test connectivity to Hyperping
curl -I https://hyperping.com/docs/api
# Should return 200 OK

# Check DNS resolution
nslookup hyperping.com
```

**Retry manually:**
```bash
# Scraper auto-retries 3 times with backoff
# If still failing, run again after 5 minutes
go run main.go
```

---

### ⚠️ Warning: "Failed to scrape page, retrying..."

**Output:**
```
[15/42] https://hyperping.com/docs/api/monitors/create
        ⚠️  Attempt 1 failed: navigation failed
        ⏳ Retry 2/3 after 2s...
```

**Causes:**
1. Temporary network glitch
2. Rate limiting by Hyperping
3. Page structure changed

**Actions:**

**If < 5 failures:** Normal, scraper will retry automatically

**If > 10 failures:**
```bash
# Check if Hyperping docs are accessible
curl https://hyperping.com/docs/api

# Reduce rate limit (be more polite)
# Edit config.go:
const DefaultRateLimit = 0.5  // Was 1.0 (slower)
```

---

### ❌ Error: "No parameters extracted"

**Output:**
```
[10/42] https://hyperping.com/docs/api/monitors/create
        ✅ Scraped (4231 chars) - CHANGED
        📋 No parameters extracted
```

**Causes:**
1. Page HTML structure changed
2. JavaScript didn't render in time
3. Extraction logic needs update

**Solutions:**

**Increase wait time:**
```go
// In main.go:332
time.Sleep(1000 * time.Millisecond)  // Was 500ms
```

**Inspect HTML manually:**
```bash
# Save HTML for inspection
# Modify main.go to log HTML:
log.Printf("HTML: %s\n", pageData.HTML)
```

**Update extraction logic:**
See `extractor/hyperping.go` for parsing rules.

---

## Network & Connectivity Issues

### ❌ Error: "Connection refused"

**Full Error:**
```
❌ Failed to navigate to https://hyperping.com/docs/api:
    connection refused
```

**Causes:**
1. No internet connection
2. Firewall blocking outbound connections
3. Hyperping website is down

**Solutions:**

**Check connectivity:**
```bash
# Test internet
ping google.com

# Test Hyperping specifically
curl https://hyperping.com
# Should return HTML

# Check DNS
nslookup hyperping.com
```

**Check firewall:**
```bash
# Allow outbound HTTPS
# (firewall commands vary by OS)

# Test with proxy if needed
export HTTPS_PROXY=http://proxy.company.com:8080
go run main.go
```

**Verify website status:**
Visit https://status.hyperping.com or https://downdetector.com

---

### ❌ Error: "Rate limit exceeded"

**Output:**
```
❌ Failed after 3 retries: HTTP 429 Too Many Requests
Retry-After: 60
```

**Cause:** Too many requests to Hyperping in short time.

**Solutions:**

**Wait and retry:**
```bash
# Wait duration specified in Retry-After header (usually 60 seconds)
sleep 60
go run main.go
```

**Reduce rate limit:**
```go
// In config.go
const DefaultRateLimit = 0.5  // Was 1.0 req/sec
```

**Spread out runs:**
Instead of running every hour, run every 6 hours.

---

### ⚠️ Warning: "TLS handshake timeout"

**Cause:** Network or proxy issues with HTTPS.

**Solutions:**

**Update Go and certificates:**
```bash
# Update Go
go version  # Should be 1.22+

# Update CA certificates (Linux)
sudo apt-get update
sudo apt-get install ca-certificates

# macOS
brew upgrade ca-certificates
```

**Test TLS connection:**
```bash
openssl s_client -connect hyperping.com:443
# Should complete handshake successfully
```

---

## GitHub Integration Issues

### ❌ Error: "GITHUB_TOKEN environment variable not set"

**Output:**
```
🐙 GitHub Integration...
   ⏭️  GitHub credentials not found
   Running in PREVIEW MODE...
```

**This is not an error!** The scraper runs fine without GitHub integration. It just won't create issues automatically.

**To enable GitHub integration:**

1. Generate token: https://github.com/settings/tokens
2. Required permissions: `repo`, `issues:write`
3. Export environment variable:
```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export GITHUB_REPOSITORY="owner/repo"
```

4. Re-run scraper:
```bash
go run main.go
```

---

### ❌ Error: "GitHub API error (status 401)"

**Full Error:**
```
❌ Failed to create issue: GitHub API error (status 401):
    Bad credentials
```

**Causes:**
1. Invalid GitHub token
2. Expired token
3. Token lacks required permissions

**Solutions:**

**Verify token:**
```bash
# Test token with GitHub API
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
     https://api.github.com/user
# Should return your user info

# If invalid, generate new token:
# https://github.com/settings/tokens
```

**Check permissions:**
Token needs:
- ✅ `repo` (full control of private repositories)
- ✅ `issues:write` (create/edit issues)

**Refresh token:**
```bash
# Generate new token and export
export GITHUB_TOKEN="ghp_NEW_TOKEN_HERE"
go run main.go
```

---

### ❌ Error: "GitHub API error (status 404)"

**Full Error:**
```
❌ Failed to create issue: GitHub API error (status 404): Not Found
```

**Causes:**
1. Repository doesn't exist
2. Token doesn't have access to repository
3. Wrong repository name format

**Solutions:**

**Verify repository:**
```bash
# Check format (must be owner/repo)
echo $GITHUB_REPOSITORY
# Should output: develeap/terraform-provider-hyperping

# Test repository access
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
     https://api.github.com/repos/$GITHUB_REPOSITORY
# Should return repository info
```

**Fix format:**
```bash
# Correct format
export GITHUB_REPOSITORY="owner/repo"

# NOT:
# export GITHUB_REPOSITORY="https://github.com/owner/repo"
# export GITHUB_REPOSITORY="owner"
```

---

### ❌ Error: "GitHub API error (status 422)"

**Full Error:**
```
❌ Failed to create issue: GitHub API error (status 422):
    Validation Failed
```

**Causes:**
1. Issue title or body too long
2. Invalid label names
3. Repository doesn't allow issues

**Solutions:**

**Check repository settings:**
- Go to repository on GitHub
- Settings → General → Features
- Ensure "Issues" is checked ✅

**Reduce diff size:**
If diff report is huge, GitHub might reject it.

**Check labels:**
Scraper creates these labels automatically:
- `api-change`
- `breaking-change`
- `automated`
- `monitors-api`, `incidents-api`, etc.

If these exist with different colors, update them or delete to let scraper recreate.

---

### ⚠️ Warning: "Failed to create label"

**Output:**
```
⚠️  Failed to create label api-change: already exists
```

**This is normal!** Labels only need to be created once. Subsequent runs will skip creation.

---

## Cache & State Issues

### ❌ Error: "Failed to load cache"

**Full Error:**
```
❌ Failed to load cache: failed to parse cache file:
    invalid character 'x' looking for beginning of value
```

**Cause:** Corrupted `.scraper_cache.json` file.

**Solutions:**

**Delete and rebuild:**
```bash
# Remove corrupted cache
rm .scraper_cache.json

# Re-run (will rebuild cache)
go run main.go
# This will re-scrape all pages (slower first run)
```

**Or rebuild from existing files:**
```go
// In main.go, replace LoadCache with:
cache, err := BuildCacheFromDisk(config.OutputDir)
if err != nil {
    log.Fatalf("❌ Failed to rebuild cache: %v\n", err)
}
```

---

### ⚠️ Warning: "All pages showing as changed"

**Output:**
```
📊 Scraping Complete!
   Scraped: 42 (content changed)  ← All pages!
   Skipped: 0 (unchanged)
```

**Causes:**
1. First run (no cache exists)
2. Cache was deleted/corrupted
3. Cache file is outdated

**This is normal for:**
- ✅ First run ever
- ✅ After deleting cache
- ✅ After major Hyperping website changes

**Subsequent runs should show:**
```
   Scraped: 3 (content changed)
   Skipped: 39 (unchanged)
```

---

### ❌ Snapshots directory growing too large

**Problem:** Disk space filling up with old snapshots.

**Check size:**
```bash
du -sh snapshots/
# Output: 2.5G	snapshots/
```

**Solution:**

Scraper auto-cleans old snapshots, keeping last 10.

**Manual cleanup:**
```bash
# Keep only last 5 snapshots
cd snapshots
ls -t | tail -n +6 | xargs rm -rf
cd ..
```

**Adjust retention:**
```go
// In main.go:270
snapshotMgr.CleanupOldSnapshots(5)  // Was 10
```

---

## Performance Issues

### ⚠️ Scraper is very slow

**Expected times:**
- First run: 2-3 minutes (downloads browser + scrapes all)
- Incremental: 30-60 seconds
- No changes: 20-30 seconds

**If slower than this:**

**Check rate limiting:**
```go
// In config.go - reduce if being too aggressive
const DefaultRateLimit = 1.0  // requests per second
```

**Check network speed:**
```bash
# Test download speed
curl -o /dev/null https://hyperping.com/docs/api
```

**Check disk I/O:**
```bash
# Slow disk can impact snapshot saving
iostat -x 1
```

**Disable resource blocking (speeds up rendering):**
```go
// In config.go
const DefaultResourceBlocking = false  // Was true
```

---

### ⚠️ High memory usage

**Expected:** ~1.2 GB RAM (headless browser)

**If using > 3 GB:**

**Solutions:**

**Close browser properly:**
Ensure `defer browser.MustClose()` is called.

**Run on machine with more RAM:**
Browser automation is inherently memory-intensive.

**Use server with swap:**
```bash
# Enable swap on Linux
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

---

### ⚠️ CPU usage at 100%

**Normal during scraping** (browser rendering JavaScript)

**If stuck at 100% for > 5 minutes:**

1. Kill process: `killall scraper`
2. Check for infinite loops in logs
3. Re-run with single page to test:

```go
// In main.go, limit discovered URLs
discovered = discovered[:1]  // Test with just first page
```

---

## Debugging Tips

### Enable verbose logging

**Modify main.go:**
```go
// At top of main()
log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)  // Show file/line
log.Println("DEBUG: Starting scraper")
```

**Add debug output in critical areas:**
```go
// In scrapePage function
log.Printf("DEBUG: Navigating to %s\n", url)
log.Printf("DEBUG: Page title: %s\n", title)
log.Printf("DEBUG: Text length: %d\n", len(text))
```

---

### Inspect scraped HTML

**Save HTML to file for manual inspection:**
```go
// In main.go after scraping
os.WriteFile("debug.html", []byte(pageData.HTML), 0644)
log.Println("DEBUG: Saved HTML to debug.html")
```

**Open in browser:**
```bash
open debug.html  # macOS
xdg-open debug.html  # Linux
```

---

### Test extraction logic

**Test parameter extraction on specific page:**
```go
// In extractor/extractor.go
func TestExtraction() {
    pageData := &PageData{
        URL: "https://hyperping.com/docs/api/monitors/create",
        HTML: /* paste HTML here */,
    }

    params := ExtractAPIParameters(pageData)
    log.Printf("Extracted %d parameters:\n", len(params))
    for _, p := range params {
        log.Printf("  - %s (%s, required=%v)\n", p.Name, p.Type, p.Required)
    }
}
```

---

### Compare snapshots manually

**List available snapshots:**
```bash
ls -lt snapshots/
```

**Compare two snapshots:**
```bash
# Install diff tool
brew install colordiff  # macOS
sudo apt-get install colordiff  # Linux

# Compare specific files
colordiff \
  snapshots/2026-02-03_10-30-00/monitors_create.json \
  snapshots/2026-02-03_14-51-20/monitors_create.json
```

---

### Check GitHub API directly

**Test token:**
```bash
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
     https://api.github.com/user
```

**Test repository access:**
```bash
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
     https://api.github.com/repos/$GITHUB_REPOSITORY
```

**Create test issue:**
```bash
curl -X POST \
  -H "Authorization: Bearer $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  https://api.github.com/repos/$GITHUB_REPOSITORY/issues \
  -d '{"title":"Test Issue","body":"Testing API access"}'
```

---

## Still Stuck?

### Collect diagnostic information

```bash
# System info
go version
uname -a
df -h

# Project info
cd tools/cmd/scraper
ls -la
cat go.mod

# Run scraper with full output
go run main.go 2>&1 | tee scraper-debug.log

# Check logs
cat scraper.log
cat scraper-debug.log
```

### Open an issue

Include:
1. **What you tried:** Full command you ran
2. **What happened:** Complete error output
3. **System info:** OS, Go version, available RAM
4. **Logs:** Attach `scraper-debug.log`

Open issue at: https://github.com/develeap/terraform-provider-hyperping/issues

---

## Prevention Tips

### Regular maintenance

```bash
# Weekly: Clean old snapshots
cd snapshots && ls -t | tail -n +11 | xargs rm -rf

# Monthly: Rebuild cache
rm .scraper_cache.json && go run main.go

# Quarterly: Rotate GitHub token
# Generate new token, update environment variable
```

### Monitoring

**Set up monitoring for:**
- ✅ Run duration > 10 minutes (may be stuck)
- ✅ Failed pages > 10 (scraping issues)
- ✅ Memory usage > 3 GB (memory leak)
- ✅ Disk usage in snapshots/ > 5 GB (cleanup needed)

### Best practices

1. **Run in cron/scheduled job** - Consistent execution
2. **Commit .env.example** - Never commit .env
3. **Keep snapshots in .gitignore** - Too large for git
4. **Rotate GitHub tokens** - Security best practice
5. **Monitor GitHub issue labels** - Ensure consistency

---

**Last Updated:** 2026-02-03
