# Architecture Plans: Hyperping API Documentation Scraper

## Executive Summary

**Goal**: Scrape 50 pages (8 parents + 41 children + 1 Notion) efficiently and reliably

**Key Challenges**:
1. JavaScript-rendered content (requires browser automation)
2. URL discovery (child pages need dynamic extraction)
3. Scale (50 pages vs current 9 pages = 5.5x increase)
4. Memory constraints (browser automation is RAM-intensive)
5. Rate limiting (respect Hyperping servers)
6. Change detection (identify what changed between scrapes)

**Research Insights Applied**:
- Browser contexts (not instances) for isolation
- Worker pool pattern for concurrency
- BFS traversal for documentation structure
- Resource blocking for 50-80% performance gain
- Circuit breaker + exponential backoff for reliability
- Content-based caching for incremental updates

---

## Plan A: Sequential Smart Scraper (Simple & Reliable)

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Sequential Smart Scraper                                     │
│                                                              │
│  1. Single Browser Instance                                 │
│     └── Reused Page (navigates to each URL)                │
│                                                              │
│  2. Two-Phase Approach                                      │
│     ├── Phase 1: Scrape Parent Pages                       │
│     │   └── Extract child URLs from navigation             │
│     └── Phase 2: Scrape All Child Pages                    │
│         └── Use extracted URLs                              │
│                                                              │
│  3. Optimizations                                           │
│     ├── Block images/CSS/fonts                              │
│     ├── Reuse page (no recreation overhead)                │
│     ├── Content-based caching (skip unchanged pages)       │
│     └── Rate limiting: 1 req/sec                            │
│                                                              │
│  4. Error Handling                                          │
│     ├── Retry failed pages 3x with exponential backoff     │
│     ├── Continue on single page failure                     │
│     └── Save partial results incrementally                  │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Flow

```go
func Sequential() {
    browser := rod.New().MustConnect()
    defer browser.Close()

    page := browser.MustPage()
    defer page.Close()

    setupResourceBlocking(page) // Block images/CSS

    // Phase 1: Scrape parents + extract child URLs
    parentURLs := []string{
        "/api/monitors",
        "/api/statuspages",
        "/api/maintenance",
        // ...
    }

    allURLs := []string{}
    for _, parentURL := range parentURLs {
        rateLimiter.Wait() // 1 req/sec

        // Scrape parent
        content := scrapePage(page, parentURL)
        save(parentURL, content)

        // Extract child URLs
        childURLs := extractNavigationLinks(page)
        allURLs = append(allURLs, childURLs...)
    }

    // Phase 2: Scrape all children
    for _, url := range allURLs {
        rateLimiter.Wait()

        // Check cache first
        if !hasChanged(url) {
            log.Printf("Skipping unchanged: %s", url)
            continue
        }

        content := scrapeWithRetry(page, url, 3)
        save(url, content)
    }
}
```

### Pros
✅ **Simple to implement** - ~200 lines of code
✅ **Minimal memory usage** - Single browser instance (~800 MB)
✅ **Easy to debug** - Sequential execution, clear logs
✅ **Reliable** - No concurrency issues, predictable behavior
✅ **Server-friendly** - Controlled 1 req/sec rate
✅ **Incremental caching** - Skip unchanged pages (saves time on re-runs)

### Cons
❌ **Slowest option** - 50 pages × 1 sec = 50 seconds minimum (plus scraping time)
❌ **Total time estimate**: ~2-3 minutes for full scrape
❌ **No parallelization** - Single-threaded execution

### Resource Requirements
- **Memory**: 1 GB (browser) + 200 MB (Go process) = **1.2 GB total**
- **CPU**: Low (single core utilized)
- **Time**: **2-3 minutes** per scrape
- **Network**: 1 req/sec = very conservative

### Best For
- Initial implementation and testing
- When reliability > speed
- Memory-constrained environments (< 4 GB RAM)
- When you need guaranteed success on all pages

### Estimated Implementation Time
**1-2 days** (includes URL discovery logic and error handling)

---

## Plan B: Concurrent Worker Pool (Balanced)

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Concurrent Worker Pool Scraper                               │
│                                                              │
│  1. Single Browser Instance                                 │
│     └── Multiple Browser Contexts (5 contexts)              │
│         ├── Worker 1 Context                                │
│         ├── Worker 2 Context                                │
│         ├── Worker 3 Context                                │
│         ├── Worker 4 Context                                │
│         └── Worker 5 Context                                │
│                                                              │
│  2. Two-Phase Approach                                      │
│     ├── Phase 1: Sequential parent scraping                 │
│     │   └── Collect all child URLs                          │
│     └── Phase 2: Concurrent child scraping                  │
│         └── 5 workers process children in parallel          │
│                                                              │
│  3. Worker Pool Pattern                                     │
│     ├── Job Queue (buffered channel, size 50)              │
│     ├── Result Channel (buffered channel, size 50)         │
│     ├── Error Channel (buffered channel, size 50)          │
│     └── Rate Limiter (shared across workers)                │
│                                                              │
│  4. Optimizations                                           │
│     ├── Block resources (images/CSS/fonts)                  │
│     ├── Browser context reuse per worker                    │
│     ├── Adaptive rate limiting (2-5 req/sec)               │
│     └── Content-based caching                               │
│                                                              │
│  5. Reliability                                             │
│     ├── Per-worker error isolation                          │
│     ├── Circuit breaker (gobreaker)                         │
│     ├── Retry with exponential backoff                      │
│     └── Partial success handling (save as you go)          │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Flow

```go
type WorkerPool struct {
    browser     *rod.Browser
    jobs        chan string
    results     chan ScrapedData
    errors      chan error
    rateLimiter *rate.Limiter
    workerCount int
}

func Concurrent() {
    browser := rod.New().MustConnect()
    defer browser.Close()

    pool := NewWorkerPool(browser, 5) // 5 workers

    // Phase 1: Sequential parent scraping
    parentURLs := getParentURLs()
    allChildURLs := []string{}

    for _, parentURL := range parentURLs {
        content := scrapePage(browser.MustPage(), parentURL)
        save(parentURL, content)

        childURLs := extractChildURLs(content)
        allChildURLs = append(allChildURLs, childURLs...)
    }

    // Phase 2: Concurrent child scraping
    pool.Start()

    for _, url := range allChildURLs {
        pool.Submit(url)
    }

    pool.Close()
    results := pool.CollectResults()
}

func (pool *WorkerPool) worker(workerID int) {
    // Create isolated context per worker
    ctx := pool.browser.MustIncognito()
    defer ctx.Close()

    page := ctx.MustPage()
    defer page.Close()

    setupResourceBlocking(page)

    for url := range pool.jobs {
        // Rate limiting (shared across workers)
        pool.rateLimiter.Wait(context.Background())

        // Scrape with retries
        data, err := scrapeWithRetry(page, url, 3)
        if err != nil {
            pool.errors <- err
            continue
        }

        pool.results <- data
    }
}
```

### Pros
✅ **5x faster** than sequential (for 41 child pages)
✅ **Balanced** - Good speed without overwhelming resources
✅ **Proven pattern** - Worker pool is production-ready pattern
✅ **Error isolation** - One worker failure doesn't affect others
✅ **Scalable** - Easy to adjust worker count (3-10)
✅ **Partial success** - Continues even if some pages fail

### Cons
❌ **More complex** - ~500 lines of code vs 200
❌ **Higher memory** - 5 contexts × 800 MB = 4 GB peak
❌ **Coordination overhead** - Channels, mutexes, goroutines
❌ **Debugging harder** - Concurrent logs interleaved

### Resource Requirements
- **Memory**: 5 contexts × 800 MB = **4 GB browser** + 500 MB Go = **4.5 GB total**
- **CPU**: Moderate (5 cores utilized)
- **Time**: **30-45 seconds** per scrape (vs 2-3 minutes sequential)
- **Network**: 2-5 req/sec (adaptive)

### Best For
- Production deployments
- When speed matters (automated weekly scraping)
- Servers with 8-16 GB RAM
- When you need balance of speed and reliability

### Estimated Implementation Time
**3-4 days** (includes worker pool, rate limiting, circuit breaker)

---

## Plan C: Hybrid Staged Pipeline (Most Efficient)

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Hybrid Staged Pipeline Scraper                               │
│                                                              │
│  Stage 1: Discovery (Sequential)                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │ • Scrape 8 parent pages sequentially                │    │
│  │ • Extract ALL child URLs (~41 URLs)                 │    │
│  │ • Build complete URL graph                           │    │
│  │ • Classify URLs by type (list/get/create/update)    │    │
│  │ • Time: ~10 seconds                                  │    │
│  └────────────────────────────────────────────────────┘    │
│                    ↓                                         │
│  Stage 2: Filtering (Smart Cache)                           │
│  ┌────────────────────────────────────────────────────┐    │
│  │ • Check cache for each URL                          │    │
│  │ • Use content hash to detect changes                │    │
│  │ • Filter out unchanged pages                         │    │
│  │ • Priority queue: changed pages first               │    │
│  │ • Time: ~2 seconds                                   │    │
│  └────────────────────────────────────────────────────┘    │
│                    ↓                                         │
│  Stage 3: Concurrent Scraping (Worker Pool)                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Workers (8 concurrent)                              │    │
│  │ ├── Priority 1: Changed pages (scrape first)       │    │
│  │ ├── Priority 2: New pages                           │    │
│  │ └── Priority 3: Periodic refresh (weekly)          │    │
│  │                                                      │    │
│  │ Optimizations:                                       │    │
│  │ • Block resources                                    │    │
│  │ • Shared browser, isolated contexts                 │    │
│  │ • Adaptive rate limiting                             │    │
│  │ • Stream results to disk (not memory)              │    │
│  │ • Time: ~20-30 seconds (only changed pages)        │    │
│  └────────────────────────────────────────────────────┘    │
│                    ↓                                         │
│  Stage 4: Diffing & Classification                          │
│  ┌────────────────────────────────────────────────────┐    │
│  │ • Compare with previous snapshot                    │    │
│  │ • Field-level diff (parameters, types, examples)   │    │
│  │ • Classify changes (breaking vs non-breaking)       │    │
│  │ • Generate structured changelog                      │    │
│  │ • Time: ~5 seconds                                   │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  Total Time: ~40 seconds (first run)                        │
│              ~15 seconds (incremental runs)                  │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Flow

```go
type PipelineScraper struct {
    browser      *rod.Browser
    cache        *ContentCache
    urlGraph     *URLGraph
    workerPool   *WorkerPool
}

func Pipeline() {
    scraper := NewPipelineScraper()

    // Stage 1: Discovery
    log.Println("Stage 1: Discovering URLs...")
    urlGraph := scraper.discoverURLs()
    log.Printf("Found %d total URLs", urlGraph.Count())

    // Stage 2: Filtering
    log.Println("Stage 2: Filtering unchanged pages...")
    toScrape := scraper.filterUnchanged(urlGraph)
    log.Printf("Need to scrape %d changed/new pages", len(toScrape))

    if len(toScrape) == 0 {
        log.Println("No changes detected!")
        return
    }

    // Stage 3: Concurrent scraping
    log.Println("Stage 3: Scraping changed pages...")
    results := scraper.scrapeConcurrent(toScrape, 8)

    // Stage 4: Diffing
    log.Println("Stage 4: Analyzing changes...")
    changes := scraper.compareWithPrevious(results)
    scraper.generateChangelog(changes)

    log.Printf("Complete! Scraped %d pages, found %d changes",
        len(results), len(changes))
}

func (ps *PipelineScraper) discoverURLs() *URLGraph {
    graph := NewURLGraph()

    // Scrape parents sequentially
    for _, parentURL := range parentURLs {
        content := scrapePage(parentURL)
        graph.AddParent(parentURL, content)

        // Extract child URLs from navigation
        childURLs := extractChildURLs(content)
        for _, childURL := range childURLs {
            graph.AddChild(parentURL, childURL)
        }
    }

    return graph
}

func (ps *PipelineScraper) filterUnchanged(graph *URLGraph) []string {
    toScrape := []string{}

    for _, url := range graph.AllURLs() {
        // Check if content changed (HEAD request or cache hash)
        if ps.cache.HasChanged(url) {
            toScrape = append(toScrape, url)
        } else {
            log.Printf("Unchanged: %s", url)
        }
    }

    // Sort by priority (changed pages first)
    sort.Slice(toScrape, func(i, j int) bool {
        return ps.cache.GetPriority(toScrape[i]) > ps.cache.GetPriority(toScrape[j])
    })

    return toScrape
}

func (ps *PipelineScraper) scrapeConcurrent(urls []string, workers int) []ScrapedData {
    pool := NewWorkerPool(ps.browser, workers)
    pool.Start()

    for _, url := range urls {
        pool.Submit(url)
    }

    return pool.CollectResults()
}
```

### Pros
✅ **Most efficient** - Only scrapes changed pages (typically 5-10%)
✅ **Fastest incremental** - 15 seconds for typical updates
✅ **Smart caching** - Content-based change detection
✅ **Scalable** - Can handle 500+ pages with same approach
✅ **Best resource usage** - Only allocates workers when needed
✅ **Detailed insights** - Knows exactly what changed
✅ **Production-ready** - All enterprise patterns included

### Cons
❌ **Most complex** - ~800 lines of code
❌ **Longer initial dev time** - 5-7 days implementation
❌ **Requires robust caching** - Complex cache management
❌ **More moving parts** - Stages, graph, cache, workers

### Resource Requirements
- **Memory**: 8 contexts × 800 MB = **6.4 GB browser** + 800 MB Go = **7.2 GB peak**
- **CPU**: High (8 cores utilized)
- **Time (first run)**: **40 seconds**
- **Time (incremental)**: **15 seconds** (only changed pages)
- **Network**: 3-8 req/sec (adaptive)

### Best For
- Production deployments at scale
- When you need maximum efficiency
- Frequent scraping (daily/hourly)
- When you have 16+ GB RAM available
- When you need detailed change tracking

### Estimated Implementation Time
**5-7 days** (includes all stages, caching, diffing, testing)

---

## Plan D: Minimal MVP (Quick Start)

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Minimal MVP Scraper                                          │
│                                                              │
│  Hardcoded URL List (No Discovery)                          │
│  ┌────────────────────────────────────────────────────┐    │
│  │ • Manually list all 50 URLs in code                │    │
│  │ • No dynamic URL extraction                         │    │
│  │ • Easy to verify completeness                       │    │
│  └────────────────────────────────────────────────────┘    │
│                    ↓                                         │
│  Sequential Scraping (No Concurrency)                       │
│  ┌────────────────────────────────────────────────────┐    │
│  │ • Single browser, single page                       │    │
│  │ • for url in urls: scrape(url)                     │    │
│  │ • 1 req/sec rate limit                              │    │
│  │ • Save to JSON files                                 │    │
│  └────────────────────────────────────────────────────┘    │
│                    ↓                                         │
│  Basic Error Handling                                        │
│  ┌────────────────────────────────────────────────────┐    │
│  │ • Retry 3x on failure                               │    │
│  │ • Continue on error                                  │    │
│  │ • Log errors to file                                 │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  Total: ~100 lines of code, 1 day implementation            │
└─────────────────────────────────────────────────────────────┘
```

### Implementation

```go
func MinimalMVP() {
    // Hardcoded list of ALL 50 URLs
    urls := []string{
        // Overview
        "https://hyperping.com/docs/api/overview",

        // Monitors (6 pages)
        "https://hyperping.com/docs/api/monitors",
        "https://hyperping.com/docs/api/monitors/list",
        "https://hyperping.com/docs/api/monitors/get",
        "https://hyperping.com/docs/api/monitors/create",
        "https://hyperping.com/docs/api/monitors/update",
        "https://hyperping.com/docs/api/monitors/delete",

        // Status Pages (9 pages)
        "https://hyperping.com/docs/api/statuspages",
        "https://hyperping.com/docs/api/statuspages/list",
        "https://hyperping.com/docs/api/statuspages/get",
        // ... all 50 URLs explicitly listed
    }

    browser := rod.New().MustConnect()
    defer browser.Close()

    page := browser.MustPage()
    defer page.Close()

    limiter := rate.NewLimiter(1.0, 1) // 1 req/sec

    for i, url := range urls {
        limiter.Wait(context.Background())

        log.Printf("[%d/%d] Scraping: %s", i+1, len(urls), url)

        content, err := scrapeWithRetry(page, url, 3)
        if err != nil {
            log.Printf("ERROR: %v", err)
            continue
        }

        filename := urlToFilename(url)
        os.WriteFile(filename, []byte(content), 0644)
    }
}
```

### Pros
✅ **Fastest to implement** - 4 hours to 1 day
✅ **Simplest code** - ~100 lines total
✅ **Easy to verify** - Manually check all 50 URLs scraped
✅ **No dynamic complexity** - Hardcoded = predictable
✅ **Minimal dependencies** - Just Rod + rate limiter

### Cons
❌ **Manual maintenance** - Add new URLs by hand
❌ **Slowest execution** - 2-3 minutes per run
❌ **No smart caching** - Scrapes everything every time
❌ **Not scalable** - Doesn't handle API growth

### Resource Requirements
- **Memory**: **1.2 GB total**
- **CPU**: Low
- **Time**: **2-3 minutes** per scrape
- **Network**: 1 req/sec

### Best For
- **Proof of concept**
- **Testing if scraping works** for all 50 pages
- **Validating URL patterns** before building dynamic discovery
- **Getting results FAST** (today)

### Estimated Implementation Time
**4 hours to 1 day**

---

## Comparison Matrix

| Criteria | Plan A: Sequential | Plan B: Worker Pool | Plan C: Pipeline | Plan D: MVP |
|----------|-------------------|---------------------|------------------|-------------|
| **Implementation Time** | 1-2 days | 3-4 days | 5-7 days | 4 hours |
| **Code Complexity** | Low (200 LOC) | Medium (500 LOC) | High (800 LOC) | Minimal (100 LOC) |
| **Memory Usage** | 1.2 GB | 4.5 GB | 7.2 GB | 1.2 GB |
| **First Run Time** | 2-3 min | 30-45 sec | 40 sec | 2-3 min |
| **Incremental Time** | 1-2 min | 20-30 sec | 15 sec | 2-3 min |
| **Reliability** | Excellent | Good | Good | Excellent |
| **Maintainability** | Easy | Medium | Complex | Very Easy |
| **Scalability** | Poor | Good | Excellent | Poor |
| **URL Discovery** | Dynamic | Dynamic | Dynamic | Hardcoded |
| **Caching** | Basic | Content-based | Smart | None |
| **Error Handling** | Robust | Robust | Robust | Basic |
| **Production Ready** | Yes | Yes | Yes | No |

---

## Recommended Phased Approach

### 🎯 **Phase 1: MVP (Today - Day 1)**
**Goal**: Prove scraping works for all 50 pages

**Plan**: Use **Plan D (MVP)** with hardcoded URLs
- Hardcode all 50 URLs explicitly
- Sequential scraping (no concurrency)
- Basic retry logic
- Save to JSON files

**Deliverable**: 50 JSON files with complete documentation

**Time**: 4 hours

**Success Criteria**: All 50 pages scraped successfully

---

### 🎯 **Phase 2: Production Sequential (Day 2-3)**
**Goal**: Make it reliable and automatic

**Plan**: Upgrade to **Plan A (Sequential Smart)** with dynamic discovery
- Implement URL discovery from navigation
- Add content-based caching
- Add resource blocking (images/CSS)
- Proper error handling + circuit breaker

**Deliverable**: Production-ready sequential scraper

**Time**: 1-2 days

**Success Criteria**:
- Discovers all URLs automatically
- Skips unchanged pages
- Handles failures gracefully
- < 2 minute scrape time

---

### 🎯 **Phase 3: Concurrent Optimization (Day 4-6) [Optional]**
**Goal**: Speed up for weekly automated runs

**Plan**: Upgrade to **Plan B (Worker Pool)** for concurrency
- Implement worker pool (5-8 workers)
- Add rate limiting
- Maintain all error handling from Phase 2

**Deliverable**: Fast production scraper

**Time**: 2-3 days

**Success Criteria**:
- < 45 second scrape time
- No increase in errors
- Works reliably in GitHub Actions

---

### 🎯 **Phase 4: Pipeline Efficiency (Future) [Optional]**
**Goal**: Maximum efficiency for frequent scraping

**Plan**: Upgrade to **Plan C (Pipeline)** for incremental updates
- Add smart filtering stage
- Implement priority queue
- Add detailed change tracking

**Deliverable**: Enterprise-grade scraper

**Time**: 3-4 days

**Success Criteria**:
- < 15 second incremental scrapes
- Detailed change reports
- Scales to 500+ pages

---

## Final Recommendation

### **Start Here: Plan D (MVP) → Plan A (Sequential Smart)**

**Rationale**:
1. **Validate assumptions** - Prove scraping works before investing in complexity
2. **Fast results** - Get all 50 pages scraped today
3. **Low risk** - Sequential is reliable and debuggable
4. **Easy migration** - Can upgrade to concurrent later if needed
5. **GitHub Actions friendly** - Sequential uses minimal resources

### **Implementation Order**:
```
Day 1:  Plan D (MVP) - Hardcoded 50 URLs, basic scraping
        → Deliverable: 50 JSON files

Day 2:  Plan A Part 1 - Add URL discovery from navigation
        → Deliverable: Dynamic URL extraction

Day 3:  Plan A Part 2 - Add caching + resource blocking
        → Deliverable: Production-ready scraper

Day 4+: [Optional] Plan B - Add worker pool for speed
        → Deliverable: 5x faster scraping
```

### **Why This Path?**
- ✅ Get results **immediately** (Day 1)
- ✅ Validate approach before heavy investment
- ✅ Each phase is **independently valuable**
- ✅ Can stop at any phase and have working solution
- ✅ Low risk, incremental improvement
- ✅ Matches GitHub Actions resource constraints

### **When to Skip Ahead**:
- If you have 16+ GB RAM available → Start with **Plan B**
- If you need maximum efficiency → Go straight to **Plan C**
- If you just need proof today → **Plan D only**

---

## Next Steps

**Ready to proceed?** Choose your starting point:

1. **"Let's validate first"** → Implement Plan D (MVP) with hardcoded URLs
2. **"Go production ready"** → Implement Plan A (Sequential Smart) with discovery
3. **"I need speed"** → Implement Plan B (Worker Pool) directly
4. **"Maximum efficiency"** → Implement Plan C (Pipeline) for enterprise-scale

Which approach should we implement first?
