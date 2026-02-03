# Hyperping API Scraper - Comprehensive Remediation Plan

**Generated:** 2026-02-03
**Review Sources:** 8 senior expert panels (Software Engineer, DevOps, Architect, SRE, Security, Domain Expert, Go Expert, Scraping Expert)
**Total Issues Found:** 60+ (14 Critical, 23 High, 16 Medium, 10+ Low)

---

## Executive Summary

**Current State:** Functional MVP with significant technical debt
**Risk Assessment:** HIGH - Not production-ready
**Estimated Total Effort:** 3-4 weeks (120-160 hours)
**Recommended Approach:** Phased implementation with immediate critical fixes

---

## Phase 0: Emergency Security Fixes (IMMEDIATE - 4 hours)

**Priority:** P0 (BLOCKER)
**Must complete before any deployment or commit**

### 0.1 Prevent Secret Exposure
**Issue:** No .gitignore, world-readable files with potential secrets

**Tasks:**
1. Create `.gitignore` file
2. Change file permissions from 0644 → 0600 for sensitive files
3. Scan git history for accidentally committed secrets

**DoD:**
- [ ] `.gitignore` created with comprehensive exclusions:
  - `.env`, `.env.*`, `.env.local`
  - `.scraper_cache.json`
  - `scraper`, `scraper_*` (binaries)
  - `*.log`
  - `docs_scraped/`
  - `snapshots/`
  - `api_changes_*.md`
- [ ] All sensitive files (cache, snapshots, reports) use 0600 permissions
- [ ] `git status` shows no sensitive files staged
- [ ] Run `truffleHog` or `gitleaks` on repo history
- [ ] If secrets found, rotate ALL affected credentials
- [ ] Commit .gitignore as first commit in remediation branch

**Code Changes:**
```go
// utils/files.go - Update all WriteFile calls
const (
    filePermPublic  = 0644 // Public docs, reports
    filePermPrivate = 0600 // Cache, snapshots, logs
)

// Use private perms for sensitive data
os.WriteFile(filename, data, filePermPrivate)
```

### 0.2 Fix GitHub Token Validation
**Issue:** No token format validation, injection risk

**Tasks:**
1. Add token format validation
2. Validate before first API call

**DoD:**
- [ ] Token format validated with regex: `^(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{36,}$`
- [ ] Invalid tokens rejected with clear error message
- [ ] Test with malformed tokens confirms rejection
- [ ] No token logged or exposed in error messages

**Code Changes:**
```go
// github.go:353 - Add validation
func LoadGitHubConfig() (*GitHubClient, error) {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        return nil, fmt.Errorf("GITHUB_TOKEN not set")
    }

    // Validate format
    validToken := regexp.MustCompile(`^(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{36,}$`)
    if !validToken.MatchString(token) {
        return nil, fmt.Errorf("invalid GITHUB_TOKEN format")
    }

    // ... rest of function
}
```

---

## Phase 1: Critical Reliability Fixes (Week 1 - 32 hours)

**Priority:** P0 (CRITICAL)
**Prevents data corruption and resource leaks**

### 1.1 Implement Graceful Shutdown
**Issue:** No signal handling, defer bypassed by log.Fatalf(), zombie browser processes

**Research:**
- Go signal handling: `os/signal` package
- Best practice: `context.WithCancel` for propagation
- Browser cleanup: Ensure `browser.Close()` always runs

**Tasks:**
1. Add signal handler for SIGTERM, SIGINT
2. Replace all `log.Fatalf()` with proper error returns
3. Propagate context through entire scraper pipeline

**DoD:**
- [ ] Signal handler catches SIGTERM, SIGINT, triggers graceful shutdown
- [ ] Context passed from main() through all I/O functions
- [ ] Zero `log.Fatalf()` calls remain (use error returns)
- [ ] Browser cleanup verified even on interrupt (manual test with Ctrl+C)
- [ ] No zombie chromium processes after any exit path
- [ ] Exit codes meaningful: 0 (success), 1 (error), 2 (panic/interrupted)
- [ ] Test: `kill -TERM <pid>` triggers clean shutdown
- [ ] Test: Ctrl+C during scraping closes browser

**Code Changes:**
```go
// main.go - New structure
func main() {
    // Setup signal handling
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Shutdown signal received, cleaning up...")
        cancel()
    }()

    // Run scraper with context
    exitCode := run(ctx)
    os.Exit(exitCode)
}

func run(ctx context.Context) int {
    config := DefaultConfig()

    browser, err := launchBrowser(ctx, config)
    if err != nil {
        log.Printf("Failed to launch browser: %v", err)
        return 1
    }
    defer cleanupBrowser(browser) // Always runs

    // All operations use ctx
    if err := scrapeAll(ctx, browser, config); err != nil {
        if ctx.Err() != nil {
            log.Println("Interrupted")
            return 2
        }
        log.Printf("Error: %v", err)
        return 1
    }

    return 0
}
```

### 1.2 Atomic File Writes
**Issue:** Direct writes cause corruption on crashes, no transactional safety

**Research:**
- Unix atomic rename: `os.Rename()` is atomic on Unix
- Pattern: Write to `.tmp` → `fsync()` → `rename()`
- Windows compatibility: Atomic on NTFS

**Tasks:**
1. Create atomic write wrapper
2. Update all `os.WriteFile()` calls to use wrapper
3. Add fsync for durability guarantee

**DoD:**
- [ ] `AtomicWriteFile()` function created in utils/files.go
- [ ] All cache writes use atomic pattern
- [ ] All snapshot saves use atomic pattern
- [ ] All report writes use atomic pattern
- [ ] Test: Kill -9 during write leaves old file intact (not corrupted)
- [ ] Test: Verify with `ls -la` shows no partial `.tmp` files after clean run
- [ ] Windows compatibility verified (if applicable)

**Code Changes:**
```go
// utils/files.go - New function
func AtomicWriteFile(filename string, data []byte, perm os.FileMode) error {
    // Create temp file in same directory (ensures same filesystem)
    dir := filepath.Dir(filename)
    tmpFile, err := os.CreateTemp(dir, ".tmp-*")
    if err != nil {
        return fmt.Errorf("create temp: %w", err)
    }
    tmpPath := tmpFile.Name()

    // Cleanup temp file on any error
    defer func() {
        tmpFile.Close()
        if _, err := os.Stat(tmpPath); err == nil {
            os.Remove(tmpPath)
        }
    }()

    // Write data
    if _, err := tmpFile.Write(data); err != nil {
        return fmt.Errorf("write temp: %w", err)
    }

    // Sync to disk (durability guarantee)
    if err := tmpFile.Sync(); err != nil {
        return fmt.Errorf("sync: %w", err)
    }

    // Close before rename
    if err := tmpFile.Close(); err != nil {
        return fmt.Errorf("close: %w", err)
    }

    // Set permissions
    if err := os.Chmod(tmpPath, perm); err != nil {
        return fmt.Errorf("chmod: %w", err)
    }

    // Atomic rename (this is the atomic operation)
    if err := os.Rename(tmpPath, filename); err != nil {
        return fmt.Errorf("rename: %w", err)
    }

    return nil
}

// Update all writes:
// OLD: os.WriteFile(filename, data, 0644)
// NEW: AtomicWriteFile(filename, data, filePermPrivate)
```

### 1.3 Browser Resource Management
**Issue:** Unbounded memory growth, no resource blocking, no crash recovery

**Research:**
- Rod resource blocking: `browser.HijackRequests()`
- Memory limits: `ulimit` or Go memory profiling
- Crash recovery: Detect and restart browser

**Tasks:**
1. Implement resource blocking (images, CSS, fonts)
2. Add memory monitoring
3. Add browser crash recovery
4. Add memory limits via cgroup (Docker) or ulimit

**DoD:**
- [ ] Resource blocking implemented for: PNG, JPG, CSS, WOFF/WOFF2, SVG
- [ ] Only HTML and JavaScript loaded
- [ ] Bandwidth reduction verified: ≥50% reduction measured
- [ ] Memory usage monitored via `runtime.ReadMemStats()`
- [ ] Browser restarted if memory exceeds 2GB
- [ ] Test: Run on 100+ pages without memory growth
- [ ] Test: Performance improvement measured (baseline vs. optimized)
- [ ] Browser crash (manual kill) triggers restart and retry

**Code Changes:**
```go
// main.go - Add resource blocking
func launchBrowser(ctx context.Context, config ScraperConfig) (*rod.Browser, error) {
    launchURL := launcher.New().
        Headless(config.Headless).
        MustLaunch()

    browser := rod.New().
        ControlURL(launchURL).
        MustConnect()

    // Resource blocking
    router := browser.HijackRequests()
    router.MustAdd("*.png", func(ctx *rod.Hijack) {
        ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
    })
    router.MustAdd("*.jpg", func(ctx *rod.Hijack) {
        ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
    })
    router.MustAdd("*.jpeg", func(ctx *rod.Hijack) {
        ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
    })
    router.MustAdd("*.css", func(ctx *rod.Hijack) {
        ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
    })
    router.MustAdd("*.woff*", func(ctx *rod.Hijack) {
        ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
    })
    router.MustAdd("*.svg", func(ctx *rod.Hijack) {
        ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
    })
    go router.Run()

    return browser, nil
}

// Add memory monitoring
func monitorMemory(browser *rod.Browser) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)

        if m.Alloc > 2*1024*1024*1024 { // 2GB
            log.Printf("WARNING: Memory usage high: %d MB", m.Alloc/1024/1024)
            // Could trigger browser restart here
        }
    }
}
```

### 1.4 Context Propagation
**Issue:** `context.Background()` used in wrong places, no cancellation support

**Research:**
- Go context best practices: Accept context, don't create
- Context should flow from top (main) to bottom (I/O)
- Rate limiter respects context cancellation

**Tasks:**
1. Add context parameter to all I/O functions
2. Check rate limiter error return value
3. Propagate context through scraping pipeline

**DoD:**
- [ ] All I/O functions accept `context.Context` as first parameter
- [ ] `limiter.Wait(ctx)` error checked
- [ ] Context cancellation tested (triggers within 1 second)
- [ ] No `context.Background()` used except in main()
- [ ] Run `go vet` confirms no context issues
- [ ] Test: Cancel context mid-scrape, verify cleanup within 1s

**Code Changes:**
```go
// Update function signatures
func scrapeWithRetry(ctx context.Context, page *rod.Page, url string, maxRetries int, timeout time.Duration) (*extractor.PageData, error)
func scrapePage(ctx context.Context, page *rod.Page, url string) (*extractor.PageData, error)
func DiscoverURLs(ctx context.Context, browser *rod.Browser, baseURL string) ([]DiscoveredURL, error)

// main.go:107 - Fix rate limiter
if err := limiter.Wait(ctx); err != nil {
    if ctx.Err() != nil {
        return ctx.Err() // Canceled or deadline exceeded
    }
    return fmt.Errorf("rate limit error: %w", err)
}
```

---

## Phase 2: Code Quality & Testing (Week 2 - 40 hours)

**Priority:** P1 (HIGH)
**Improves maintainability and prevents regressions**

### 2.1 Refactor God Function
**Issue:** main() is 256 lines, impossible to test, violates SRP

**Research:**
- Single Responsibility Principle
- Extract method refactoring
- Testable architecture patterns

**Tasks:**
1. Extract 6 phases into separate functions
2. Create Scraper struct with dependencies
3. Each phase returns error, no side effects

**DoD:**
- [ ] main() reduced to <50 lines (orchestration only)
- [ ] 6 phase functions extracted with clear responsibilities:
  - `setupConfig()` - Configuration loading
  - `launchBrowser(ctx, config)` - Browser initialization
  - `discoverURLs(ctx, browser, config)` - URL discovery
  - `scrapePages(ctx, browser, urls, cache, config)` - Scraping phase
  - `analyzeChanges(ctx, cache, config)` - Diff analysis
  - `reportResults(ctx, diffs, config)` - GitHub integration
- [ ] Each function <80 lines
- [ ] Each function independently testable
- [ ] Run `golangci-lint` shows complexity ≤15 for all functions

**Code Changes:**
```go
// main.go - Refactored structure
type Scraper struct {
    config  ScraperConfig
    browser *rod.Browser
    cache   Cache
}

func main() {
    ctx, cancel := setupShutdown()
    defer cancel()

    config := setupConfig()

    scraper, err := NewScraper(ctx, config)
    if err != nil {
        log.Fatalf("Setup failed: %v", err)
    }
    defer scraper.Cleanup()

    if err := scraper.Run(ctx); err != nil {
        log.Fatalf("Scraper failed: %v", err)
    }
}

func (s *Scraper) Run(ctx context.Context) error {
    urls, err := s.discoverURLs(ctx)
    if err != nil {
        return fmt.Errorf("discovery: %w", err)
    }

    newCache, err := s.scrapePages(ctx, urls)
    if err != nil {
        return fmt.Errorf("scraping: %w", err)
    }

    diffs, err := s.analyzeChanges(ctx, newCache)
    if err != nil {
        return fmt.Errorf("analysis: %w", err)
    }

    if err := s.reportResults(ctx, diffs); err != nil {
        return fmt.Errorf("reporting: %w", err)
    }

    return nil
}
```

### 2.2 Add Unit Tests (80%+ Coverage Target)
**Issue:** Zero test files, no automated verification

**Research:**
- Go testing best practices
- Table-driven tests
- Mock patterns (httptest, rod mocking)

**Tasks:**
1. Create test files for all packages
2. Mock browser for scraper tests
3. Mock HTTP for GitHub tests
4. Add table-driven tests for extractors
5. Achieve ≥80% coverage (go test -cover)

**DoD:**
- [ ] Test files created for all packages:
  - `main_test.go`
  - `cache_test.go`
  - `discovery_test.go`
  - `differ_test.go`
  - `snapshots_test.go`
  - `github_test.go`
  - `extractor/extractor_test.go`
  - `extractor/hyperping_test.go`
  - `extractor/tables_test.go`
  - `utils/files_test.go`
  - `utils/hash_test.go`
- [ ] Coverage ≥80% (`go test -cover ./... | grep "coverage"`)
- [ ] All tests pass: `go test ./... -count=1`
- [ ] Race detector clean: `go test -race ./...`
- [ ] Tests run in <10 seconds
- [ ] CI integration (GitHub Actions) runs tests on PR

**Test Structure Example:**
```go
// extractor/hyperping_test.go
func TestParseHyperpingParameter(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        want     APIParameter
        wantErr  bool
    }{
        {
            name:  "basic string required",
            input: "namestringrequired",
            want: APIParameter{
                Name:     "name",
                Type:     "string",
                Required: true,
            },
        },
        {
            name:  "number optional",
            input: "check_frequencynumberoptional",
            want: APIParameter{
                Name:     "check_frequency",
                Type:     "number",
                Required: false,
            },
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := parseHyperpingParameter(tt.input)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %+v, want %+v", got, tt.want)
            }
        })
    }
}
```

### 2.3 Fix Go Anti-Patterns
**Issue:** Multiple Go-specific issues (variable shadowing, unused returns, etc.)

**Research:**
- Go Code Review Comments (official guide)
- Common Go mistakes
- golangci-lint configuration

**Tasks:**
1. Remove all `os.Stdout.Sync()` calls
2. Fix variable shadowing (`filepath` → `filePath`)
3. Pre-compile regexes at package level
4. Remove duplicate functions
5. Check all error returns

**DoD:**
- [ ] Zero `os.Stdout.Sync()` calls
- [ ] No variable shadowing warnings
- [ ] All regexes pre-compiled as package-level vars
- [ ] Duplicate functions removed (use utils versions)
- [ ] All error returns checked (no `_` except justified)
- [ ] `go vet ./...` passes with zero warnings
- [ ] `golangci-lint run` passes with zero errors
- [ ] `staticcheck ./...` passes

**Code Changes:**
```go
// extractor/filters.go - Pre-compile regexes
var (
    paramStartRegex = regexp.MustCompile(`^[a-zA-Z_]`)
    paramBodyRegex  = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

func isValidParameterName(name string) bool {
    return paramStartRegex.MatchString(name) &&
           paramBodyRegex.MatchString(name)
}

// Remove duplicate fileExists from main.go
// Use utils.FileExists everywhere
```

---

## Phase 3: Scraping Robustness (Week 2 - 24 hours)

**Priority:** P1 (HIGH)
**Prevents fragile scraping failures**

### 3.1 Improve Content Waiting Strategy
**Issue:** Hardcoded sleeps, no verification content loaded, race conditions

**Research:**
- Rod's WaitStable() vs WaitLoad() vs Element waiting
- Best practices: Combine multiple strategies
- Timeout hierarchies

**Tasks:**
1. Replace all `time.Sleep()` with proper waits
2. Implement multi-strategy content waiting
3. Add content verification after wait

**DoD:**
- [ ] Zero hardcoded `time.Sleep()` calls for waiting (config timeouts OK)
- [ ] Use `page.MustWaitStable()` for dynamic content
- [ ] Use `page.Element()` with timeout for specific content
- [ ] Fallback strategy: Network idle → DOM stable → Element present
- [ ] Test: Slow-loading pages still scraped correctly
- [ ] Test: Fast pages don't wait unnecessarily
- [ ] Average page time reduced by 30%

**Code Changes:**
```go
// main.go - Replace sleep with proper wait
func waitForContent(page *rod.Page) error {
    // Strategy 1: Wait for specific content marker
    if _, err := page.Timeout(5 * time.Second).Element("article.docs"); err == nil {
        return nil
    }

    // Strategy 2: Wait for DOM stability
    if err := page.WaitStable(2 * time.Second); err != nil {
        return fmt.Errorf("dom not stable: %w", err)
    }

    // Strategy 3: Verify JavaScript ready
    ready, _ := page.Eval(`() => document.readyState === 'complete'`)
    if !ready.Bool() {
        return fmt.Errorf("document not ready")
    }

    return nil
}

// Replace:
// time.Sleep(500 * time.Millisecond)
// With:
if err := waitForContent(page); err != nil {
    log.Printf("Warning: %v", err)
    // Continue anyway
}
```

### 3.2 Classify Retryable Errors
**Issue:** Retries 404s and network errors equally, wastes time and resources

**Research:**
- HTTP status code semantics
- Transient vs. permanent errors
- Exponential backoff with jitter

**Tasks:**
1. Create error classification function
2. Skip retries for permanent errors (404, 403)
3. Retry only transient errors (5xx, timeout, connection refused)
4. Add jitter to backoff

**DoD:**
- [ ] `isRetryableError()` function created
- [ ] 4xx errors (except 429) not retried
- [ ] 5xx errors retried with backoff
- [ ] Network timeouts retried
- [ ] Connection errors retried
- [ ] Jitter added to prevent thundering herd
- [ ] Max backoff capped at 30 seconds
- [ ] Test: 404 returns immediately (no retries)
- [ ] Test: 503 retries 3 times
- [ ] Retry logs show jitter variation

**Code Changes:**
```go
// main.go - Error classification
func isRetryableError(err error) bool {
    if err == nil {
        return false
    }

    errStr := err.Error()

    // Network errors - retry
    if strings.Contains(errStr, "timeout") ||
       strings.Contains(errStr, "connection refused") ||
       strings.Contains(errStr, "network") {
        return true
    }

    // HTTP 5xx - retry
    if strings.Contains(errStr, "500") ||
       strings.Contains(errStr, "502") ||
       strings.Contains(errStr, "503") ||
       strings.Contains(errStr, "504") {
        return true
    }

    // HTTP 429 - retry (rate limit)
    if strings.Contains(errStr, "429") {
        return true
    }

    // Client errors - don't retry
    return false
}

func scrapeWithRetry(ctx context.Context, page *rod.Page, url string, maxRetries int) (*extractor.PageData, error) {
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        if attempt > 0 {
            // Only retry if error is retryable
            if !isRetryableError(lastErr) {
                return nil, fmt.Errorf("non-retryable error: %w", lastErr)
            }

            // Exponential backoff with jitter and cap
            backoff := time.Duration(1<<uint(attempt)) * time.Second
            if backoff > 30*time.Second {
                backoff = 30 * time.Second
            }
            jitter := time.Duration(rand.Intn(1000)) * time.Millisecond

            log.Printf("⏳ Retry %d/%d after %v (+ jitter)...", attempt+1, maxRetries, backoff)
            time.Sleep(backoff + jitter)
        }

        pageData, err := scrapePage(ctx, page, url)
        if err == nil {
            return pageData, nil
        }

        lastErr = err
        log.Printf("⚠️  Attempt %d failed: %v", attempt+1, err)
    }

    return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}
```

### 3.3 Add Extraction Confidence Scoring
**Issue:** No validation if extraction succeeded, false positives undetected

**Research:**
- Multi-strategy extraction voting
- Confidence scoring algorithms
- Sanity checks for extracted data

**Tasks:**
1. Add confidence field to extraction results
2. Compare results from different strategies
3. Flag low-confidence extractions
4. Add sanity checks (reserved words, empty results)

**DoD:**
- [ ] `ExtractionResult` struct includes `Confidence float64` field
- [ ] Each strategy returns confidence score:
  - Hyperping format: 0.9 (high confidence)
  - Table format: 0.7 (medium)
  - JSON format: 0.5 (low)
- [ ] Multiple strategies agree → boost confidence
- [ ] Sanity checks implemented:
  - 0 params + "parameter" in text → flag warning
  - Reserved words extracted (type, required, etc.) → flag warning
  - Duplicate param names → flag warning
- [ ] Warnings logged for manual review
- [ ] Test: Known good page → confidence ≥0.8
- [ ] Test: Malformed page → confidence <0.5, warning logged

**Code Changes:**
```go
// extractor/extractor.go - Add confidence
type ExtractionResult struct {
    Parameters []APIParameter
    Confidence float64
    Strategy   string
    Warnings   []string
}

func ExtractWithConfidence(pageData *PageData) ExtractionResult {
    strategies := []struct {
        name       string
        confidence float64
        extract    func(*PageData) []APIParameter
    }{
        {"hyperping", 0.9, extractFromHyperpingFormat},
        {"table", 0.7, extractFromTables},
        {"json", 0.5, extractFromJSONExamples},
    }

    var best ExtractionResult

    for _, strat := range strategies {
        params := strat.extract(pageData)
        result := ExtractionResult{
            Parameters: params,
            Confidence: strat.confidence,
            Strategy:   strat.name,
        }

        // Sanity checks
        if len(params) == 0 && strings.Contains(pageData.Text, "parameter") {
            result.Warnings = append(result.Warnings, "No params extracted but 'parameter' found in text")
            result.Confidence *= 0.5
        }

        for _, p := range params {
            if p.Name == "type" || p.Name == "required" {
                result.Warnings = append(result.Warnings, fmt.Sprintf("Suspicious param name: %s", p.Name))
                result.Confidence *= 0.8
            }
        }

        if result.Confidence > best.Confidence {
            best = result
        }
    }

    return best
}
```

---

## Phase 4: Operational Excellence (Week 3 - 32 hours)

**Priority:** P2 (MEDIUM)
**Enables production monitoring and debugging**

### 4.1 Add Structured Logging
**Issue:** Printf-style logs, no levels, hard to parse, emojis break tooling

**Research:**
- Go structured logging: `log/slog` (Go 1.21+) or `logrus`
- JSON format for machine parsing
- Log levels: DEBUG, INFO, WARN, ERROR

**Tasks:**
1. Replace `log.Printf` with structured logger
2. Add log levels
3. Output JSON format
4. Remove emojis from log output

**DoD:**
- [ ] All `log.Printf()` replaced with structured logging
- [ ] Log levels used appropriately:
  - DEBUG: Scraping progress, extraction details
  - INFO: Phase completion, summary stats
  - WARN: Retries, skipped pages, low confidence
  - ERROR: Fatal errors, critical failures
- [ ] JSON output format for machine parsing
- [ ] Zero emojis in log output (optional colored output for terminal)
- [ ] Logs include context: timestamp, level, component, message
- [ ] Test: Parse logs with `jq` successfully
- [ ] Environment variable `LOG_LEVEL` controls verbosity

**Code Changes:**
```go
// main.go - Setup structured logging
import "log/slog"

func setupLogging() *slog.Logger {
    level := os.Getenv("LOG_LEVEL")
    var logLevel slog.Level
    switch level {
    case "DEBUG":
        logLevel = slog.LevelDebug
    case "WARN":
        logLevel = slog.LevelWarn
    case "ERROR":
        logLevel = slog.LevelError
    default:
        logLevel = slog.LevelInfo
    }

    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: logLevel,
    })

    return slog.New(handler)
}

// Usage:
logger := setupLogging()
logger.Info("Starting scraper",
    "config", config,
    "urls_discovered", len(urls))

logger.Warn("Retry attempt",
    "url", url,
    "attempt", attempt,
    "error", err)

logger.Error("Fatal error",
    "component", "github",
    "error", err)
```

### 4.2 Add Metrics and Health Checks
**Issue:** No observability, can't monitor scraper health or performance

**Research:**
- Prometheus metrics (histograms, counters, gauges)
- Health check patterns
- Metrics to track: duration, success rate, cache hit ratio

**Tasks:**
1. Add Prometheus metrics endpoint
2. Track key metrics (duration, errors, cache hits)
3. Add /health endpoint
4. Add /metrics endpoint

**DoD:**
- [ ] Metrics exposed on `:9090/metrics` (Prometheus format)
- [ ] Key metrics tracked:
  - `scraper_run_duration_seconds` (histogram)
  - `scraper_pages_scraped_total{status="success|failure|skipped"}` (counter)
  - `scraper_cache_hit_ratio` (gauge)
  - `scraper_errors_total{type="network|parse|github"}` (counter)
  - `scraper_last_success_timestamp` (gauge)
- [ ] Health check endpoint `:8080/health` returns JSON:
  ```json
  {
    "status": "healthy",
    "last_run": "2026-02-03T10:00:00Z",
    "uptime_seconds": 3600
  }
  ```
- [ ] Test: `curl http://localhost:9090/metrics` returns Prometheus format
- [ ] Test: `curl http://localhost:8080/health` returns 200 OK
- [ ] Optional: Grafana dashboard template created

**Code Changes:**
```go
// metrics.go - New file
import "github.com/prometheus/client_golang/prometheus"

var (
    runDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "scraper_run_duration_seconds",
            Help: "Duration of scraper runs",
        },
    )

    pagesScraped = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "scraper_pages_scraped_total",
            Help: "Total pages scraped",
        },
        []string{"status"},
    )

    cacheHitRatio = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "scraper_cache_hit_ratio",
            Help: "Cache hit ratio (0.0 to 1.0)",
        },
    )
)

func init() {
    prometheus.MustRegister(runDuration, pagesScraped, cacheHitRatio)
}

// main.go - Expose metrics
func startMetricsServer() {
    http.Handle("/metrics", promhttp.Handler())
    go http.ListenAndServe(":9090", nil)
}

// Track metrics during scraping
timer := prometheus.NewTimer(runDuration)
defer timer.ObserveDuration()

pagesScraped.WithLabelValues("success").Inc()
cacheHitRatio.Set(float64(cached) / float64(total))
```

### 4.3 CI/CD Integration
**Issue:** No automated testing, no release pipeline

**Research:**
- GitHub Actions for Go projects
- Test, build, release workflow
- Artifact publishing

**Tasks:**
1. Create GitHub Actions workflow
2. Run tests on every PR
3. Run linters (golangci-lint, gosec)
4. Build binaries for releases

**DoD:**
- [ ] `.github/workflows/test.yml` created
- [ ] Workflow runs on: push, pull_request
- [ ] Workflow steps:
  - Checkout code
  - Setup Go 1.22
  - Run `go test ./... -race -cover`
  - Run `golangci-lint run`
  - Run `gosec ./...`
  - Upload coverage report
- [ ] PR status checks enforced (tests must pass)
- [ ] Release workflow builds binaries on git tags
- [ ] Test: Create PR, verify CI runs
- [ ] Test: Tag v0.1.0, verify binary built

**GitHub Actions Workflow:**
```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
          cache: true

      - name: Cache Chromium
        uses: actions/cache@v3
        with:
          path: ~/.cache/rod
          key: rod-chromium-${{ runner.os }}

      - name: Run tests
        working-directory: tools/cmd/scraper
        run: |
          go test ./... -race -cover -coverprofile=coverage.out

      - name: Run linters
        uses: golangci/golangci-lint-action@v3
        with:
          working-directory: tools/cmd/scraper

      - name: Security scan
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec ./...
        working-directory: tools/cmd/scraper

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./tools/cmd/scraper/coverage.out
```

---

## Phase 5: Architecture Improvements (Week 3-4 - 24 hours)

**Priority:** P2 (MEDIUM)
**Long-term maintainability and extensibility**

### 5.1 Dependency Injection & Interfaces
**Issue:** Tight coupling, hard to test, no interfaces

**Research:**
- Go interface design principles
- Dependency injection patterns
- Testing with interfaces

**Tasks:**
1. Define interfaces for major components
2. Accept interfaces in constructors
3. Enable mock implementations for testing

**DoD:**
- [ ] Interfaces defined:
  - `BrowserManager` interface (Launch, NewPage, Close)
  - `CacheRepository` interface (Load, Save, HasChanged)
  - `SnapshotRepository` interface (Save, Load, Compare)
  - `GitHubService` interface (CreateIssue, CreateLabels)
- [ ] All functions accept interfaces, not concrete types
- [ ] Mock implementations created for testing
- [ ] Test: Scraper can run with mock dependencies (no network)
- [ ] Example mock test passes

**Code Changes:**
```go
// interfaces.go - New file
type BrowserManager interface {
    Launch(ctx context.Context, config ScraperConfig) error
    NewPage(ctx context.Context) (Page, error)
    Close() error
}

type Page interface {
    Navigate(ctx context.Context, url string) error
    WaitLoad(ctx context.Context) error
    HTML() (string, error)
    Element(selector string) (Element, error)
}

type CacheRepository interface {
    Load(ctx context.Context, filename string) (Cache, error)
    Save(ctx context.Context, filename string, cache Cache) error
    HasChanged(filename string, content string) bool
}

// Scraper uses interfaces
type Scraper struct {
    browser BrowserManager
    cache   CacheRepository
    github  GitHubService
    config  ScraperConfig
}

// main.go - Wire dependencies
browser := &RodBrowser{}
cache := &FilesystemCache{}
github := &GitHubClient{}

scraper := NewScraper(browser, cache, github, config)
```

### 5.2 Configuration Validation
**Issue:** No validation of config values, negative timeouts possible

**Research:**
- Go validation patterns
- Constraint checking
- Default value handling

**Tasks:**
1. Add `Validate()` method to ScraperConfig
2. Check bounds on all numeric fields
3. Verify required fields set

**DoD:**
- [ ] `ScraperConfig.Validate()` method created
- [ ] Validation checks:
  - RateLimit > 0
  - Timeout > 0
  - Retries ≥ 0
  - BaseURL is valid URL
  - OutputDir is not empty
- [ ] Invalid config returns clear error message
- [ ] Test: Negative timeout → validation error
- [ ] Test: Invalid URL → validation error
- [ ] Validation called before scraper starts

**Code Changes:**
```go
// config.go - Add validation
func (c ScraperConfig) Validate() error {
    if c.RateLimit <= 0 {
        return fmt.Errorf("rate_limit must be positive, got %f", c.RateLimit)
    }

    if c.Timeout <= 0 {
        return fmt.Errorf("timeout must be positive, got %v", c.Timeout)
    }

    if c.Retries < 0 {
        return fmt.Errorf("retries cannot be negative, got %d", c.Retries)
    }

    if _, err := url.Parse(c.BaseURL); err != nil {
        return fmt.Errorf("invalid base_url: %w", err)
    }

    if c.OutputDir == "" {
        return fmt.Errorf("output_dir cannot be empty")
    }

    return nil
}

// main.go - Validate before use
config := DefaultConfig()
if err := config.Validate(); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}
```

---

## Phase 6: Domain Accuracy (Week 4 - 16 hours)

**Priority:** P2 (MEDIUM)
**Ensures scraper extracts correct API structure**

### 6.1 Protocol-Aware Extraction
**Issue:** Doesn't understand HTTP vs Port vs ICMP monitors, conditional requirements

**Research:**
- Hyperping API monitor types
- Conditional field requirements
- Parameter validation rules

**Tasks:**
1. Add protocol detection to extraction
2. Track conditional requirements
3. Validate parameter combinations

**DoD:**
- [ ] Extractor detects monitor type (HTTP, Port, ICMP)
- [ ] Conditional requirements tracked in data model
- [ ] Example: `port` marked as "required when protocol=port"
- [ ] Validation rules implemented
- [ ] Test: HTTP monitor parameters extracted correctly
- [ ] Test: Port monitor parameters include port field
- [ ] Test: ICMP monitor parameters exclude HTTP-specific fields

**Code Changes:**
```go
// extractor/extractor.go - Enhanced parameter
type ConditionalRequirement struct {
    Field     string   // "protocol"
    Operator  string   // "equals", "not_equals"
    Values    []string // ["port"]
}

type APIParameter struct {
    Name                   string
    Type                   string
    Required               bool
    ConditionalRequirement *ConditionalRequirement
    Default                interface{}
    Description            string
    ValidValues            []string
    Deprecated             bool
}

// Extraction logic
if strings.Contains(description, "when protocol") {
    param.ConditionalRequirement = &ConditionalRequirement{
        Field:    "protocol",
        Operator: "equals",
        Values:   []string{"port"},
    }
}
```

### 6.2 Enum Value Extraction
**Issue:** ValidValues field never populated, no enum validation

**Research:**
- Enum detection patterns
- Hyperping enum formats
- Validation generation

**Tasks:**
1. Extract enum values from "Available options:" sections
2. Parse comma-separated lists
3. Parse concatenated region strings

**DoD:**
- [ ] Enum extraction implemented
- [ ] Patterns recognized:
  - "Available options: X, Y, Z"
  - "Valid values: X | Y | Z"
  - Concatenated: "londonfrankfurtsydney" → ["london", "frankfurt", "sydney"]
- [ ] ValidValues populated for known enums:
  - protocol: ["http", "port", "icmp"]
  - http_method: ["GET", "POST", "PUT", "PATCH", "DELETE"]
  - regions: [list of all regions]
- [ ] Test: Known enum page → ValidValues populated
- [ ] Test: Regions concatenated string → correctly split

**Code Changes:**
```go
// extractor/hyperping.go - Enum extraction
func extractEnumValues(line string) []string {
    // Pattern 1: "Available options: X, Y, Z"
    if strings.Contains(line, "Available options:") {
        after := strings.Split(line, "Available options:")[1]
        return splitClean(after, ",")
    }

    // Pattern 2: "Valid values: X | Y | Z"
    if strings.Contains(line, "Valid values:") {
        after := strings.Split(line, "Valid values:")[1]
        return splitClean(after, "|")
    }

    // Pattern 3: Concatenated regions
    if strings.Contains(line, "londonfrankfurt") {
        return splitRegions(line)
    }

    return nil
}

func splitRegions(concatenated string) []string {
    knownRegions := []string{
        "london", "frankfurt", "singapore", "sydney",
        "virginia", "oregon", "saopaulo", "tokyo",
        "bahrain", "sanfrancisco", "nyc", "paris",
        "seoul", "mumbai", "bangalore", "california",
        "toronto", "amsterdam",
    }

    var found []string
    remaining := strings.ToLower(concatenated)

    for _, region := range knownRegions {
        if strings.Contains(remaining, region) {
            found = append(found, region)
            remaining = strings.Replace(remaining, region, "", 1)
        }
    }

    return found
}
```

---

## Phase 7: Documentation & Handoff (Week 4 - 8 hours)

**Priority:** P3 (LOW)
**Knowledge transfer and maintenance**

### 7.1 Comprehensive Documentation
**Issue:** No runbook, no architecture docs, no troubleshooting guide

**Tasks:**
1. Create ARCHITECTURE.md
2. Create RUNBOOK.md
3. Create TROUBLESHOOTING.md
4. Update README.md

**DoD:**
- [ ] ARCHITECTURE.md documents:
  - System overview diagram
  - Component responsibilities
  - Data flow
  - Technology choices and rationale
- [ ] RUNBOOK.md includes:
  - Deployment instructions
  - Configuration options
  - Monitoring and alerting
  - Backup and recovery
- [ ] TROUBLESHOOTING.md covers:
  - Common failure scenarios
  - Debug techniques
  - Resolution steps
  - Contact information
- [ ] README.md updated with:
  - Quick start guide
  - Prerequisites
  - Basic usage examples
  - Link to detailed docs

### 7.2 Code Documentation
**Issue:** Minimal comments, no package docs, no examples

**Tasks:**
1. Add package-level documentation
2. Add godoc comments to exported functions
3. Add examples for key functions

**DoD:**
- [ ] All packages have package-level doc comment
- [ ] All exported functions have godoc comments
- [ ] Example functions created for key use cases
- [ ] `godoc -http=:6060` serves readable documentation
- [ ] Test: `go doc` shows useful info for each package

---

## Success Metrics & Validation

### Phase 0-1 Validation (Week 1 End)
- [ ] Zero secrets in git history (`gitleaks` clean)
- [ ] Zero zombie browser processes after any exit
- [ ] Cache corruption impossible (atomic writes verified)
- [ ] Context cancellation responds within 1 second
- [ ] Exit codes meaningful (0=success, 1=error, 2=interrupted)

### Phase 2-3 Validation (Week 2 End)
- [ ] Code coverage ≥80% (`go test -cover`)
- [ ] All linters pass (go vet, golangci-lint, staticcheck)
- [ ] main() function <50 lines
- [ ] Average page scrape time reduced by 30%
- [ ] 404 errors don't retry (immediate failure)
- [ ] Extraction confidence logged for all pages

### Phase 4-5 Validation (Week 3 End)
- [ ] Metrics endpoint returns Prometheus format
- [ ] Health check endpoint returns 200 OK
- [ ] CI passes on every PR
- [ ] Structured logs parseable with `jq`
- [ ] Scraper runs with mock dependencies (no network)

### Final Validation (Week 4 End)
- [ ] Full scraper run completes in <90 seconds (50 pages)
- [ ] Memory usage stays below 1GB
- [ ] Zero false positives in diff detection
- [ ] All documentation complete and reviewed
- [ ] Handoff session with team complete

---

## Rollout Plan

### Pre-Rollout
1. Create feature branch: `feat/scraper-remediation`
2. Set up project tracking (GitHub Issues or Jira)
3. Schedule daily standups for duration of remediation
4. Set up monitoring dashboard (Grafana + Prometheus)

### Week 1 (Critical Fixes)
- Day 1: Phase 0 (Security - 4h)
- Day 2-3: Phase 1.1-1.2 (Graceful shutdown, atomic writes - 16h)
- Day 4-5: Phase 1.3-1.4 (Browser resources, context - 16h)
- **Checkpoint:** Run scraper, verify no crashes, no zombies, clean shutdown

### Week 2 (Quality & Scraping)
- Day 1-2: Phase 2.1-2.2 (Refactoring, tests - 20h)
- Day 3: Phase 2.3 (Go anti-patterns - 8h)
- Day 4-5: Phase 3.1-3.3 (Scraping improvements - 24h)
- **Checkpoint:** 80% test coverage, all linters pass, scraper 30% faster

### Week 3 (Operations)
- Day 1-2: Phase 4.1-4.2 (Logging, metrics - 20h)
- Day 3: Phase 4.3 (CI/CD - 12h)
- Day 4-5: Phase 5.1-5.2 (Architecture - 24h)
- **Checkpoint:** CI passing, metrics exported, health checks working

### Week 4 (Polish & Handoff)
- Day 1-2: Phase 6.1-6.2 (Domain accuracy - 16h)
- Day 3: Phase 7.1-7.2 (Documentation - 8h)
- Day 4: Integration testing, bug fixes
- Day 5: Code review, handoff session
- **Checkpoint:** All validation criteria met, docs complete

### Post-Rollout
- Merge to main after code review approval
- Deploy to staging environment
- Run for 1 week in staging with monitoring
- Deploy to production
- Monitor for 2 weeks, address any issues
- Schedule retrospective

---

## Risk Management

### High Risks
1. **Breaking existing functionality**
   - Mitigation: 80% test coverage before refactoring
   - Mitigation: Feature flags for new behavior
   - Mitigation: Staging environment testing

2. **Timeline slip (missing deadlines)**
   - Mitigation: Daily progress tracking
   - Mitigation: Prioritized phases (can skip P3)
   - Mitigation: Pair programming on complex tasks

3. **Incomplete testing leading to production issues**
   - Mitigation: Chaos testing (kill -9, disk full, network partition)
   - Mitigation: Load testing (100+ pages)
   - Mitigation: Staging environment mirrors production

### Medium Risks
4. **Team knowledge gaps (Go, testing, rod library)**
   - Mitigation: Pair programming with senior engineers
   - Mitigation: Code review process
   - Mitigation: Documentation as we go

5. **Scope creep (adding features instead of fixing issues)**
   - Mitigation: Strict adherence to remediation plan
   - Mitigation: New features tracked separately
   - Mitigation: Weekly scope review

---

## Resource Requirements

### Personnel
- **Primary Developer:** Full-time (4 weeks)
- **Code Reviewer:** 4 hours/week
- **QA Tester:** 8 hours (week 4)
- **DevOps Support:** 4 hours (CI/CD setup)

### Infrastructure
- **Staging Environment:** Required
- **CI/CD Pipeline:** GitHub Actions (included)
- **Monitoring:** Prometheus + Grafana (can use cloud providers)

### Tools
- Go 1.22+
- golangci-lint
- gosec
- truffleHog or gitleaks
- Prometheus client library
- Rod (already in use)

---

## Appendix A: Complete Checklist

Copy this to track progress:

```markdown
## Phase 0: Emergency Security (4h)
- [ ] 0.1 .gitignore created
- [ ] 0.2 File permissions fixed (0600)
- [ ] 0.3 Token validation added
- [ ] 0.4 Git history scanned for secrets

## Phase 1: Critical Reliability (32h)
- [ ] 1.1 Signal handling implemented
- [ ] 1.2 Atomic file writes implemented
- [ ] 1.3 Resource blocking implemented
- [ ] 1.4 Context propagation fixed
- [ ] 1.5 Rate limiter errors checked

## Phase 2: Code Quality (40h)
- [ ] 2.1 main() refactored (<50 lines)
- [ ] 2.2 Unit tests added (≥80% coverage)
- [ ] 2.3 Go anti-patterns fixed
- [ ] 2.4 All linters pass

## Phase 3: Scraping Robustness (24h)
- [ ] 3.1 Content waiting improved
- [ ] 3.2 Error classification implemented
- [ ] 3.3 Confidence scoring added

## Phase 4: Operational Excellence (32h)
- [ ] 4.1 Structured logging implemented
- [ ] 4.2 Metrics endpoint added
- [ ] 4.3 CI/CD pipeline created

## Phase 5: Architecture (24h)
- [ ] 5.1 Interfaces defined
- [ ] 5.2 Config validation added

## Phase 6: Domain Accuracy (16h)
- [ ] 6.1 Protocol-aware extraction
- [ ] 6.2 Enum value extraction

## Phase 7: Documentation (8h)
- [ ] 7.1 Architecture docs written
- [ ] 7.2 Code documentation added
```

---

## Appendix B: Testing Checklist

### Security Testing
- [ ] Run gitleaks/truffleHog on repo
- [ ] Verify no secrets in git history
- [ ] Test invalid GitHub token rejection
- [ ] Verify file permissions (0600 for sensitive files)

### Reliability Testing
- [ ] Test Ctrl+C during scraping (clean shutdown)
- [ ] Test kill -TERM (graceful shutdown)
- [ ] Test kill -9 during cache write (no corruption)
- [ ] Test OOM scenario (memory limit enforcement)
- [ ] Run 100 page scrape without memory growth

### Scraping Testing
- [ ] Test with slow-loading pages
- [ ] Test with 404 pages (no retry)
- [ ] Test with 503 pages (retry with backoff)
- [ ] Test with network timeout (retry)
- [ ] Verify resource blocking (measure bandwidth)
- [ ] Test extraction confidence scoring

### Integration Testing
- [ ] Run full scraper with all phases
- [ ] Verify GitHub issue creation
- [ ] Test snapshot comparison
- [ ] Test cache hit/miss scenarios

### Performance Testing
- [ ] Baseline: Time 50 page scrape (before optimization)
- [ ] After Phase 1: Measure improvement
- [ ] After Phase 3: Measure improvement
- [ ] Memory profiling: `go test -memprofile`
- [ ] CPU profiling: `go test -cpuprofile`

---

**END OF REMEDIATION PLAN**

This plan addresses all 60+ issues found by the review panel. Estimated total effort: 180 hours (4.5 weeks at 40h/week). Can be compressed to 3 weeks with focused effort or extended to 6 weeks for comprehensive testing.
