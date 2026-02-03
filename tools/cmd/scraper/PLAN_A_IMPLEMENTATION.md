# Plan A: Sequential Smart Scraper - Implementation Plan

## Status: Ready to Implement
**Estimated Time**: 3 days
**Approach**: Sequential build (MVP → URL Discovery → Semantic Diff → GitHub Integration)

---

## File Structure

```
tools/cmd/scraper/
├── main.go                    # NEW: Main scraper (replaces scraper_mvp.go)
├── discovery.go               # NEW: URL discovery from navigation
├── extractor.go               # NEW: Parameter extraction from HTML
├── differ.go                  # NEW: Semantic diff engine
├── cache.go                   # NEW: Smart caching system
├── github.go                  # NEW: GitHub issue creation
├── models.go                  # NEW: Data structures
├── go.mod                     # Existing
│
├── scraper_mvp.go            # Keep as reference
├── test_single.go            # Keep for debugging
├── compare_runs.go           # Keep for manual testing
│
└── snapshots/                 # NEW: Structured storage
    ├── cache.json            # Content hashes
    ├── 2026-02-03/           # Timestamped snapshots
    │   ├── monitors/
    │   │   ├── parent.json
    │   │   ├── list.json
    │   │   ├── create.json
    │   │   └── ...
    │   ├── statuspages/
    │   └── ...
    └── latest/               # Symlink to most recent
```

---

## Day 1: Core Scraping with URL Discovery

### models.go - Data Structures

```go
package main

// API documentation structure
type APISection struct {
    Name        string         `json:"name"`         // e.g., "monitors"
    ParentURL   string         `json:"parent_url"`
    ParentPage  PageData       `json:"parent_page"`
    ChildPages  []PageData     `json:"child_pages"`
    Endpoints   []APIEndpoint  `json:"endpoints"`
}

type PageData struct {
    URL         string          `json:"url"`
    Title       string          `json:"title"`
    Method      string          `json:"method"`       // GET, POST, PUT, DELETE
    Endpoint    string          `json:"endpoint"`     // /v1/monitors
    Text        string          `json:"text"`
    HTML        string          `json:"html"`
    Parameters  []APIParameter  `json:"parameters"`
    Timestamp   time.Time       `json:"timestamp"`
    ContentHash string          `json:"content_hash"`
}

type APIParameter struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"`          // string, boolean, integer, array, object
    Required    bool        `json:"required"`
    Default     interface{} `json:"default"`
    Description string      `json:"description"`
    ValidValues []string    `json:"valid_values"`  // For enums
    Deprecated  bool        `json:"deprecated"`
    Location    string      `json:"location"`      // body, query, path, header
}

type APIEndpoint struct {
    Section    string         `json:"section"`    // monitors, statuspages, etc.
    Method     string         `json:"method"`
    Path       string         `json:"path"`
    Title      string         `json:"title"`
    Parameters []APIParameter `json:"parameters"`
    URL        string         `json:"url"`        // Documentation URL
}

// Caching
type CacheEntry struct {
    URL          string    `json:"url"`
    ContentHash  string    `json:"content_hash"`
    LastModified time.Time `json:"last_modified"`
    Parameters   []string  `json:"parameters"`  // Parameter names for quick diff
}

type Cache struct {
    Entries   map[string]CacheEntry `json:"entries"`
    UpdatedAt time.Time             `json:"updated_at"`
}

// Diff results
type APIDiff struct {
    Section          string               `json:"section"`
    Endpoint         string               `json:"endpoint"`
    Method           string               `json:"method"`
    AddedParams      []APIParameter       `json:"added_parameters"`
    RemovedParams    []APIParameter       `json:"removed_parameters"`
    ModifiedParams   []ParameterChange    `json:"modified_parameters"`
    Breaking         bool                 `json:"breaking"`
    DocumentationURL string               `json:"documentation_url"`
}

type ParameterChange struct {
    Name            string      `json:"name"`
    WasRequired     bool        `json:"was_required"`
    NowRequired     bool        `json:"now_required"`
    OldType         string      `json:"old_type"`
    NewType         string      `json:"new_type"`
    OldDefault      interface{} `json:"old_default"`
    NewDefault      interface{} `json:"new_default"`
    Breaking        bool        `json:"breaking"`
    BreakingReason  string      `json:"breaking_reason"`
}

type DiffReport struct {
    Timestamp     time.Time   `json:"timestamp"`
    TotalChanges  int         `json:"total_changes"`
    BreakingCount int         `json:"breaking_count"`
    Diffs         []APIDiff   `json:"diffs"`
}
```

### discovery.go - URL Discovery

```go
package main

import (
    "fmt"
    "strings"
    "github.com/go-rod/rod"
)

// Parent pages to start discovery from
var parentPages = []string{
    "https://hyperping.com/docs/api/monitors",
    "https://hyperping.com/docs/api/statuspages",
    "https://hyperping.com/docs/api/maintenance",
    "https://hyperping.com/docs/api/incidents",
    "https://hyperping.com/docs/api/outages",
    "https://hyperping.com/docs/api/healthchecks",
    "https://hyperping.com/docs/api/reports",
}

type URLDiscoverer struct {
    browser *rod.Browser
}

func NewURLDiscoverer(browser *rod.Browser) *URLDiscoverer {
    return &URLDiscoverer{browser: browser}
}

// DiscoverAllURLs discovers all API documentation URLs
func (d *URLDiscoverer) DiscoverAllURLs() (map[string][]string, error) {
    sections := make(map[string][]string)

    for _, parentURL := range parentPages {
        sectionName := extractSectionName(parentURL)

        // Scrape parent page
        page := d.browser.MustPage(parentURL)
        page.MustWaitLoad()

        // Extract child URLs from navigation
        childURLs := d.extractChildURLs(page, parentURL)

        // Store parent + children
        sections[sectionName] = append([]string{parentURL}, childURLs...)

        page.Close()

        fmt.Printf("  ✅ %s: found %d child pages\n", sectionName, len(childURLs))
    }

    return sections, nil
}

// extractChildURLs extracts child documentation URLs from parent page
func (d *URLDiscoverer) extractChildURLs(page *rod.Page, parentURL string) []string {
    var childURLs []string

    // Look for navigation links in common locations
    selectors := []string{
        "nav a[href*='/api/']",           // Navigation links
        ".sidebar a[href*='/api/']",      // Sidebar links
        "[data-section] a[href]",         // Section-specific links
    }

    seen := make(map[string]bool)

    for _, selector := range selectors {
        elements, err := page.Elements(selector)
        if err != nil {
            continue
        }

        for _, el := range elements {
            href, err := el.Property("href")
            if err != nil {
                continue
            }

            url := href.String()

            // Filter: must be child of current parent
            if !strings.HasPrefix(url, parentURL+"/") {
                continue
            }

            // Deduplicate
            if seen[url] {
                continue
            }
            seen[url] = true

            childURLs = append(childURLs, url)
        }
    }

    return childURLs
}

func extractSectionName(url string) string {
    parts := strings.Split(url, "/")
    return parts[len(parts)-1]
}
```

### main.go - Core Scraper

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/go-rod/rod"
    "github.com/go-rod/rod/lib/launcher"
    "golang.org/x/time/rate"
)

type Scraper struct {
    browser    *rod.Browser
    limiter    *rate.Limiter
    cache      *Cache
    outputDir  string
}

func main() {
    fmt.Println("🚀 Hyperping API Scraper - Plan A")
    fmt.Println("=" + strings.Repeat("=", 59))

    // Initialize
    scraper := NewScraper()
    defer scraper.Close()

    // Phase 1: URL Discovery
    fmt.Println("\n📍 Phase 1: Discovering API URLs...")
    discoverer := NewURLDiscoverer(scraper.browser)
    sections, err := discoverer.DiscoverAllURLs()
    if err != nil {
        log.Fatalf("Discovery failed: %v", err)
    }

    totalURLs := 0
    for _, urls := range sections {
        totalURLs += len(urls)
    }
    fmt.Printf("✅ Discovered %d URLs across %d sections\n", totalURLs, len(sections))

    // Phase 2: Smart Scraping (with cache)
    fmt.Println("\n📥 Phase 2: Smart Scraping...")
    scrapedData := scraper.ScrapeAll(sections)
    fmt.Printf("✅ Scraped %d pages\n", len(scrapedData))

    // Phase 3: Parameter Extraction
    fmt.Println("\n🔍 Phase 3: Extracting API Parameters...")
    extractor := NewParameterExtractor()
    for i := range scrapedData {
        scrapedData[i].Parameters = extractor.Extract(scrapedData[i])
    }
    fmt.Printf("✅ Extracted parameters from %d pages\n", len(scrapedData))

    // Phase 4: Diff Detection
    fmt.Println("\n📊 Phase 4: Detecting Changes...")
    differ := NewDiffer()
    diffs := differ.Compare(scraper.cache, scrapedData)

    if len(diffs.Diffs) == 0 {
        fmt.Println("✅ No changes detected")
    } else {
        fmt.Printf("📝 Detected %d changes (%d breaking)\n",
            diffs.TotalChanges, diffs.BreakingCount)

        // Generate report
        report := GenerateMarkdownReport(diffs)
        os.WriteFile("diff-report.md", []byte(report), 0644)
        fmt.Println("✅ Diff report saved to diff-report.md")
    }

    // Phase 5: Save Results
    fmt.Println("\n💾 Phase 5: Saving...")
    scraper.Save(scrapedData)
    scraper.UpdateCache(scrapedData)

    fmt.Println("\n" + strings.Repeat("=", 60))
    fmt.Println("✅ Scraping Complete!")
}

func NewScraper() *Scraper {
    // Launch browser
    launchURL := launcher.New().Headless(true).MustLaunch()
    browser := rod.New().ControlURL(launchURL).MustConnect()

    // Rate limiter: 1 req/sec
    limiter := rate.NewLimiter(1.0, 1)

    // Load cache
    cache := LoadCache(".scraper_cache.json")

    return &Scraper{
        browser:   browser,
        limiter:   limiter,
        cache:     cache,
        outputDir: "snapshots",
    }
}

func (s *Scraper) Close() {
    s.browser.MustClose()
}

func (s *Scraper) ScrapeAll(sections map[string][]string) []PageData {
    var allData []PageData

    for section, urls := range sections {
        fmt.Printf("\n  📂 Section: %s (%d pages)\n", section, len(urls))

        for i, url := range urls {
            fmt.Printf("    [%d/%d] %s\n", i+1, len(urls), url)

            // Check cache
            if !s.cache.NeedsUpdate(url) {
                fmt.Println("      ⏭️  Cached (unchanged)")
                continue
            }

            // Rate limit
            s.limiter.Wait(context.Background())

            // Scrape
            page := s.browser.MustPage(url)
            data := ScrapePage(page)
            page.Close()

            allData = append(allData, data)
            fmt.Printf("      ✅ %d chars\n", len(data.Text))
        }
    }

    return allData
}
```

---

## Day 2: Semantic Diff Engine

### extractor.go - Parameter Extraction

```go
package main

import (
    "regexp"
    "strings"
)

type ParameterExtractor struct {
    // Regex patterns for parameter extraction
    paramNamePattern   *regexp.Regexp
    paramTypePattern   *regexp.Regexp
    requiredPattern    *regexp.Regexp
    defaultPattern     *regexp.Regexp
}

func NewParameterExtractor() *ParameterExtractor {
    return &ParameterExtractor{
        paramNamePattern:  regexp.MustCompile(`(?m)^([a-z_][a-z0-9_]*)\s+(string|boolean|integer|number|array|object)`),
        paramTypePattern:  regexp.MustCompile(`type:\s*(string|boolean|integer|array|object)`),
        requiredPattern:   regexp.MustCompile(`required|Required`),
        defaultPattern:    regexp.MustCompile(`[Dd]efault[s:]?\s*(.+)`),
    }
}

// Extract parameters from page content
func (e *ParameterExtractor) Extract(page PageData) []APIParameter {
    var params []APIParameter

    // Split text into sections
    sections := strings.Split(page.Text, "\n\n")

    for _, section := range sections {
        // Look for parameter definitions
        if e.looksLikeParameter(section) {
            param := e.parseParameter(section)
            if param != nil {
                params = append(params, *param)
            }
        }
    }

    return params
}

func (e *ParameterExtractor) looksLikeParameter(text string) bool {
    // Check if text looks like a parameter definition
    lower := strings.ToLower(text)
    return strings.Contains(lower, "string") ||
           strings.Contains(lower, "boolean") ||
           strings.Contains(lower, "integer") ||
           strings.Contains(lower, "required") ||
           strings.Contains(lower, "optional")
}

func (e *ParameterExtractor) parseParameter(text string) *APIParameter {
    // Parse parameter from text block
    // This is simplified - real implementation needs robust parsing

    param := &APIParameter{}

    // Extract name and type
    matches := e.paramNamePattern.FindStringSubmatch(text)
    if len(matches) >= 3 {
        param.Name = matches[1]
        param.Type = matches[2]
    } else {
        return nil // Not a valid parameter
    }

    // Check if required
    param.Required = e.requiredPattern.MatchString(text)

    // Extract default value
    defaultMatches := e.defaultPattern.FindStringSubmatch(text)
    if len(defaultMatches) >= 2 {
        param.Default = strings.TrimSpace(defaultMatches[1])
    }

    // Extract description (first line usually)
    lines := strings.Split(text, "\n")
    if len(lines) > 0 {
        param.Description = strings.TrimSpace(lines[0])
    }

    return param
}
```

### differ.go - Semantic Diff

```go
package main

type Differ struct{}

func NewDiffer() *Differ {
    return &Differ{}
}

// Compare old cache with new scraped data
func (d *Differ) Compare(cache *Cache, newData []PageData) DiffReport {
    report := DiffReport{
        Timestamp: time.Now(),
        Diffs:     []APIDiff{},
    }

    // Group by endpoint
    endpoints := d.groupByEndpoint(newData)

    for endpoint, newPage := range endpoints {
        // Get old version from cache
        oldEntry, exists := cache.Entries[newPage.URL]
        if !exists {
            // New endpoint
            report.Diffs = append(report.Diffs, d.createAddedDiff(newPage))
            continue
        }

        // Compare parameters
        diff := d.compareParameters(oldEntry, newPage)
        if diff != nil {
            report.Diffs = append(report.Diffs, *diff)
            if diff.Breaking {
                report.BreakingCount++
            }
        }
    }

    report.TotalChanges = len(report.Diffs)
    return report
}

func (d *Differ) compareParameters(old CacheEntry, new PageData) *APIDiff {
    // Compare parameter lists
    oldParams := d.parseParameterNames(old.Parameters)
    newParams := d.extractParameterNames(new.Parameters)

    added := difference(newParams, oldParams)
    removed := difference(oldParams, newParams)

    if len(added) == 0 && len(removed) == 0 {
        return nil // No changes
    }

    diff := &APIDiff{
        Section:  extractSectionFromURL(new.URL),
        Endpoint: new.Endpoint,
        Method:   new.Method,
        AddedParams:   d.filterParameters(new.Parameters, added),
        RemovedParams: d.filterParameters(old.Parameters, removed), // Needs old data
        Breaking:      len(removed) > 0, // Removed params are breaking
    }

    return diff
}

func difference(a, b []string) []string {
    m := make(map[string]bool)
    for _, item := range b {
        m[item] = true
    }

    var diff []string
    for _, item := range a {
        if !m[item] {
            diff = append(diff, item)
        }
    }
    return diff
}
```

---

## Day 3: GitHub Integration

### github.go - Issue Creation

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

type GitHubIssueCreator struct {
    token      string
    repo       string
    owner      string
}

func NewGitHubIssueCreator() *GitHubIssueCreator {
    return &GitHubIssueCreator{
        token: os.Getenv("GITHUB_TOKEN"),
        repo:  os.Getenv("GITHUB_REPOSITORY"),  // org/repo format
        owner: "", // Extracted from GITHUB_REPOSITORY
    }
}

func (g *GitHubIssueCreator) CreateIssue(report DiffReport) error {
    // Generate markdown body
    body := GenerateMarkdownReport(report)

    // Classify labels
    labels := []string{"api-change", "automated"}
    if report.BreakingCount > 0 {
        labels = append(labels, "breaking-change", "needs-review")
    } else {
        labels = append(labels, "non-breaking")
    }

    // Add per-API labels
    for _, diff := range report.Diffs {
        labels = append(labels, diff.Section+"-api")
    }

    // Create issue title
    title := g.generateTitle(report)

    // GitHub API call
    issueData := map[string]interface{}{
        "title":  title,
        "body":   body,
        "labels": labels,
    }

    jsonData, _ := json.Marshal(issueData)

    url := fmt.Sprintf("https://api.github.com/repos/%s/issues", g.repo)
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    req.Header.Set("Authorization", "Bearer "+g.token)
    req.Header.Set("Accept", "application/vnd.github.v3+json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 201 {
        return fmt.Errorf("failed to create issue: %d", resp.StatusCode)
    }

    fmt.Println("✅ GitHub issue created")
    return nil
}

func (g *GitHubIssueCreator) generateTitle(report DiffReport) string {
    if report.TotalChanges == 1 {
        diff := report.Diffs[0]
        if len(diff.AddedParams) == 1 {
            return fmt.Sprintf("[API Change] New parameter '%s' in %s %s",
                diff.AddedParams[0].Name, diff.Method, diff.Endpoint)
        }
    }

    if report.BreakingCount > 0 {
        return fmt.Sprintf("[API Change] %d breaking changes detected", report.BreakingCount)
    }

    return fmt.Sprintf("[API Change] %d API updates detected", report.TotalChanges)
}

func GenerateMarkdownReport(report DiffReport) string {
    var buf bytes.Buffer

    buf.WriteString("## Summary\n\n")
    buf.WriteString(fmt.Sprintf("API changes detected on %s\n\n",
        report.Timestamp.Format("2006-01-02 15:04:05 UTC")))
    buf.WriteString(fmt.Sprintf("- **Total changes**: %d\n", report.TotalChanges))
    buf.WriteString(fmt.Sprintf("- **Breaking changes**: %d\n\n", report.BreakingCount))

    for _, diff := range report.Diffs {
        buf.WriteString(fmt.Sprintf("### %s %s\n\n", diff.Method, diff.Endpoint))

        if len(diff.AddedParams) > 0 {
            buf.WriteString("#### 🆕 New Parameters\n\n")
            for _, param := range diff.AddedParams {
                buf.WriteString(fmt.Sprintf("- **%s** (%s, %s)\n",
                    param.Name,
                    param.Type,
                    ternary(param.Required, "required", "optional")))
                if param.Description != "" {
                    buf.WriteString(fmt.Sprintf("  - %s\n", param.Description))
                }
                if param.Default != nil {
                    buf.WriteString(fmt.Sprintf("  - Default: `%v`\n", param.Default))
                }
            }
            buf.WriteString("\n")
        }

        if len(diff.RemovedParams) > 0 {
            buf.WriteString("#### ❌ Removed Parameters\n\n")
            for _, param := range diff.RemovedParams {
                buf.WriteString(fmt.Sprintf("- **%s** (%s)\n", param.Name, param.Type))
            }
            buf.WriteString("\n")
        }

        buf.WriteString(fmt.Sprintf("**Impact**: %s\n\n",
            ternary(diff.Breaking, "⚠️ Breaking", "✅ Non-breaking")))

        buf.WriteString(fmt.Sprintf("**Documentation**: %s\n\n", diff.DocumentationURL))
        buf.WriteString("---\n\n")
    }

    buf.WriteString("## Action Items\n\n")
    buf.WriteString("- [ ] Review changes\n")
    buf.WriteString("- [ ] Update provider schema\n")
    buf.WriteString("- [ ] Update tests\n")
    buf.WriteString("- [ ] Update documentation\n\n")

    buf.WriteString("---\n")
    buf.WriteString("🤖 Auto-generated by Hyperping API scraper\n")

    return buf.String()
}

func ternary(condition bool, ifTrue, ifFalse string) string {
    if condition {
        return ifTrue
    }
    return ifFalse
}
```

---

## Testing Strategy

### Unit Tests
- Test URL discovery with mock HTML
- Test parameter extraction with known patterns
- Test diff detection with before/after snapshots

### Integration Test
```bash
# Run full scraper
go run main.go

# Verify output
ls snapshots/latest/
cat diff-report.md

# Manual test: modify API docs (if possible) and re-run
```

---

## Success Criteria

- [ ] Discovers all 50 URLs automatically
- [ ] Scrapes only changed pages (98% cache hit expected)
- [ ] Extracts parameters with 90%+ accuracy
- [ ] Detects added/removed parameters correctly
- [ ] Generates readable markdown reports
- [ ] Creates GitHub issues with proper labels
- [ ] Runs in < 60 seconds (with cache)
- [ ] Runs in < 120 seconds (cold start)

---

## Next: Start Implementation

Ready to build? I'll create the files in order:
1. models.go (data structures)
2. discovery.go (URL discovery)
3. extractor.go (parameter extraction)
4. differ.go (semantic diff)
5. cache.go (smart caching)
6. github.go (issue creation)
7. main.go (orchestration)

Shall I start implementing?
