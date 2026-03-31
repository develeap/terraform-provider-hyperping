package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	tfjson "github.com/hashicorp/terraform-json"
	"golang.org/x/time/rate"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/analyzer"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/contract"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/coverage"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/diff"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/discovery"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/notify"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/openapi"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/utils"
)

// Command line flags.
var (
	analyzeMode   bool
	syncCheck     bool
	snapshotDir   string
	cassetteDir   string
	schemaFile    string // path to `terraform providers schema -json` output
	forceBaseline bool   // bypass enum-regression guard; use when API genuinely removed a value
)

func main() {
	flag.BoolVar(&analyzeMode, "analyze", false, "Run coverage analysis comparing API spec to provider schema")
	flag.BoolVar(&syncCheck, "sync", false, "Quick sync check — exits 1 if provider is out of sync with API")
	flag.StringVar(&snapshotDir, "snapshot-dir", "./snapshots", "Path to snapshot directory")
	flag.StringVar(&cassetteDir, "cassette-dir", "", "Path to VCR cassettes for contract testing (optional)")
	flag.StringVar(&schemaFile, "schema-file", "", "Path to 'terraform providers schema -json' output (required for -analyze/-sync)")
	flag.BoolVar(&forceBaseline, "force-baseline", false, "Accept new snapshot even if enum regression is detected (use when the API has genuinely removed a value)")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\n🛑 Shutting down...")
		cancel()
	}()

	var code int
	switch {
	case syncCheck:
		code = runSyncCheck()
	case analyzeMode:
		code = runAnalyze()
	default:
		code = runScrape(ctx)
	}
	os.Exit(code) //nolint:gocritic // exitAfterDefer: cancel() is a no-op after os.Exit anyway
}

// runSyncCheck exits 0 if coverage ≥ 80%, otherwise exits 1.
func runSyncCheck() int {
	log.SetFlags(0)
	spec, tfSchemas, err := loadSpecAndSchema()
	if err != nil {
		fmt.Println("❌ Sync check failed:", err)
		return 1
	}
	report := coverage.Analyze(spec, tfSchemas)
	PrintSyncStatus(report)
	if report.CoveragePercent < 80 {
		return 1
	}
	return 0
}

// runAnalyze prints a full coverage report and saves it as markdown.
func runAnalyze() int {
	log.SetFlags(0)
	fmt.Println("🔍 Hyperping Provider Coverage Analyzer")
	fmt.Println(strings.Repeat("=", 60))

	spec, tfSchemas, err := loadSpecAndSchema()
	if err != nil {
		fmt.Println("❌", err)
		return 1
	}

	report := coverage.Analyze(spec, tfSchemas)

	fmt.Printf("\nOverall Coverage: %.1f%%\n", report.CoveragePercent)
	fmt.Printf("API Fields: %d  |  Covered: %d  |  Missing: %d  |  Stale: %d\n\n",
		report.TotalAPIFields, report.CoveredFields, report.MissingFields, report.StaleFields)

	for _, g := range report.Gaps {
		icon := "⚠️"
		if g.GapType == "missing" {
			icon = "❌"
		}
		fmt.Printf("%s [%s] %s → %s: %s\n", icon, g.GapType, g.APIField, g.TFField, g.Details)
	}

	// Save markdown coverage report using existing formatter.
	if len(report.Gaps) > 0 {
		md := formatCoverageMarkdown(report)
		fname := "coverage_report_" + time.Now().Format("2006-01-02_15-04-05") + ".md"
		if err := utils.SaveToFile(fname, md); err != nil {
			log.Printf("⚠️  Failed to save report: %v\n", err)
		} else {
			fmt.Println("\n✅ Report saved:", fname)
		}
	}
	return 0
}

// runScrape is the default mode: discover → scrape → generate OAS → diff → notify.
func runScrape(ctx context.Context) int {
	log.SetFlags(0)
	fmt.Println("🚀 Hyperping API Documentation Scraper")
	fmt.Println(strings.Repeat("=", 60))

	config := DefaultConfig()
	if err := os.MkdirAll(config.OutputDir, 0o750); err != nil {
		log.Printf("❌ Failed to create output dir: %v\n", err)
		return 1
	}

	cache, err := LoadCache(config.CacheFile)
	if err != nil {
		log.Printf("❌ Failed to load cache: %v\n", err)
		return 1
	}

	// Discover URLs from sitemap — no browser required.
	log.Println("\n🔍 Discovering API docs from sitemap...")
	discovered, err := discovery.DiscoverFromSitemap("")
	if err != nil {
		log.Printf("❌ URL discovery failed: %v\n", err)
		return 1
	}
	log.Printf("   Found %d API documentation pages\n", len(discovered))

	// Launch browser for JS-rendered content.
	browser, cleanup, err := launchBrowser(config)
	if err != nil {
		log.Printf("❌ Browser launch failed: %v\n", err)
		return 1
	}
	defer cleanup()

	// Scrape pages and extract parameters.
	apiParams, newCache, stats, err := scrapeAndExtract(ctx, browser, discovered, cache, config)
	if err != nil {
		log.Printf("❌ Failed to initialise scrape page: %v\n", err)
		return 1
	}
	saveResults(discovered, stats, newCache, config.CacheFile)

	// Generate OpenAPI spec from scraped params.
	apiVersion := time.Now().Format("2006-01-02")
	spec := openapi.Generate(apiParams, apiVersion)

	// Save OAS YAML snapshots.
	snapshotMgr := NewSnapshotManager(snapshotDir)

	// Identify previous snapshot BEFORE saving the new one so we compare
	// the freshly generated spec against the most recent cached snapshot.
	prevSpecPath, prevErr := snapshotMgr.GetLatestSnapshot()

	// Guard: if enum values shrank vs. the previous baseline the scrape was likely
	// degraded (lazy-loaded DOM nodes not yet visible when HTML was captured).
	// We track consecutive degraded runs; once the count reaches DegradedAcceptThreshold
	// we accept the snapshot as a genuine API change so real removals are never
	// silenced permanently. Use -force-baseline to accept immediately.
	isDegraded := false
	if prevErr == nil && !forceBaseline {
		regressions, regErr := DetectEnumRegression(prevSpecPath, spec)
		switch {
		case regErr != nil:
			// Proceed without the guard rather than blocking on a corrupt prev spec.
			log.Printf("⚠️  Enum regression check failed (proceeding without guard): %v\n", regErr)
		case len(regressions) > 0:
			state, stateErr := snapshotMgr.LoadDegradedState()
			if stateErr != nil {
				log.Printf("⚠️  Could not load degraded state (%v) — resetting counter\n", stateErr)
				state = &DegradedState{}
			}

			// Distinguish "same regression persisting" from "new regression pattern".
			if regressionSetsMatch(state.Regressions, regressions) {
				state.ConsecutiveCount++
			} else {
				state.ConsecutiveCount = 1
			}
			state.Regressions = regressions

			if saveErr := snapshotMgr.SaveDegradedState(state); saveErr != nil {
				log.Printf("⚠️  Failed to persist degraded state: %v\n", saveErr)
			}

			if state.ConsecutiveCount >= DegradedAcceptThreshold {
				// Seen consistently — accept as genuine API change, notify for human review.
				log.Printf("⚠️  Enum regression seen for %d consecutive runs — accepting as genuine API change\n",
					state.ConsecutiveCount)
				log.Println("   ℹ️  Review the diff carefully; use -force-baseline to suppress the guard if this is a scrape issue.")
				if resetErr := snapshotMgr.ResetDegradedState(); resetErr != nil {
					log.Printf("⚠️  Failed to reset degraded state: %v\n", resetErr)
				}
				// isDegraded stays false — snapshot will be saved and diff run normally.
			} else {
				isDegraded = true
				log.Printf("⚠️  Degraded scrape (%d/%d consecutive) — baseline NOT updated:\n",
					state.ConsecutiveCount, DegradedAcceptThreshold)
				for _, r := range regressions {
					log.Printf("   %s %s .%s: missing [%s]\n",
						r.Method, r.Path, r.Field,
						strings.Join(missingEnumValues(r.OldValues, r.NewValues), ", "))
				}
				log.Println("   ℹ️  Root cause: likely lazy-loaded content not fully rendered.")
				log.Printf("   ℹ️  Diff skipped. Will accept after %d total consecutive runs, or use -force-baseline.\n",
					DegradedAcceptThreshold)
			}
		default:
			// Clean run — reset any accumulated degraded count.
			if resetErr := snapshotMgr.ResetDegradedState(); resetErr != nil {
				log.Printf("⚠️  Failed to reset degraded state: %v\n", resetErr)
			}
		}
	} else if forceBaseline {
		log.Println("   ℹ️  -force-baseline set: enum regression guard bypassed.")
	}

	if !isDegraded {
		if err := snapshotMgr.SaveSnapshot(time.Now(), spec); err != nil {
			log.Printf("⚠️  Snapshot save failed: %v\n", err)
		}
	}

	// Always write the latest OAS copy for CI inspection, even on degraded runs.
	if err := SaveLatestOpenAPI(spec, config.OutputDir); err != nil {
		log.Printf("⚠️  Failed to save latest OAS: %v\n", err)
	}

	// Diff only when the snapshot was successfully updated.
	if !isDegraded {
		newSpecPath, newErr := snapshotMgr.GetLatestSnapshot()
		if prevErr == nil && newErr == nil && prevSpecPath != newSpecPath {
			if result, err := diff.Compare(prevSpecPath, newSpecPath); err != nil {
				log.Printf("⚠️  Diff failed: %v\n", err)
			} else if result.HasChanges {
				log.Println("\n📝 API changes detected:")
				fmt.Println(result.Summary)
				if result.HasPathChanges {
					notifyAPIChange(result)
				} else {
					log.Println("   ⏭️  Metadata-only changes — skipping GitHub issue creation")
				}
			} else {
				log.Println("\n✅ No API changes detected")
			}
		} else if prevErr != nil {
			log.Println("   ℹ️  No previous snapshot found — skipping diff (first run?)")
		}
	}

	// Contract validation (optional).
	if cassetteDir != "" {
		latestSpec, _ := snapshotMgr.GetLatestSnapshot() //nolint:errcheck
		if errs, err := contract.ValidateCassettes(latestSpec, cassetteDir); err != nil {
			log.Printf("⚠️  Contract validation error: %v\n", err)
		} else {
			log.Printf("🔬 Contract validation: %d issues\n", len(errs))
			for _, e := range errs {
				log.Printf("   %s %s %s: %s\n", e.CassetteFile, e.Method, e.Path, e.Message)
			}
		}
	}

	snapshotMgr.CleanupOldSnapshots(DefaultSnapshotRetention) //nolint:errcheck
	log.Println("\n✅ Done!")
	return 0
}

// --- helpers ---

func loadSpecAndSchema() (*openapi.Spec, *tfjson.ProviderSchemas, error) {
	snapshotMgr := NewSnapshotManager(snapshotDir)
	specPath, err := snapshotMgr.GetLatestSnapshot()
	if err != nil {
		return nil, nil, fmt.Errorf("no snapshot found (run scraper first): %w", err)
	}
	spec, err := openapi.Load(specPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load OAS spec: %w", err)
	}
	if schemaFile == "" {
		return nil, nil, fmt.Errorf("-schema-file is required (run: terraform providers schema -json > schema.json)")
	}
	tfSchema, err := coverage.LoadProviderSchema(schemaFile)
	if err != nil {
		return nil, nil, fmt.Errorf("load provider schema: %w", err)
	}
	return spec, tfSchema, nil
}

// scrapeAndExtract scrapes each discovered URL and extracts API parameters.
// Returns a map keyed by doc path (e.g., "monitors/create") → []APIParameter.
func scrapeAndExtract(
	ctx context.Context,
	browser *rod.Browser,
	discovered []discovery.DiscoveredURL,
	cache Cache,
	config ScraperConfig,
) (map[string][]extractor.APIParameter, Cache, ScrapeStats, error) {
	log.Println("\n🔄 Scraping API documentation pages...")

	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, Cache{}, ScrapeStats{}, fmt.Errorf("failed to create browser page: %w", err)
	}
	defer func() {
		if closeErr := page.Close(); closeErr != nil {
			log.Printf("⚠️  Page close error: %v\n", closeErr)
		}
	}()

	limiter := rate.NewLimiter(rate.Limit(config.RateLimit), 1)
	newCache := Cache{Entries: make(map[string]CacheEntry), CreatedAt: time.Now()}
	apiParams := make(map[string][]extractor.APIParameter, len(discovered))
	stats := ScrapeStats{}
	start := time.Now()

	for i, d := range discovered {
		if ctx.Err() != nil {
			break
		}
		if err := limiter.Wait(ctx); err != nil {
			break
		}

		log.Printf("[%d/%d] %s\n", i+1, len(discovered), d.URL)

		pageData, err := scrapeWithRetry(ctx, page, d.URL, config.Retries, config.Timeout)
		if err != nil {
			log.Printf("  ❌ Failed: %v\n", err)
			stats.Failed++
			continue
		}

		filename := URLToFilename(d.URL)
		if HasChanged(cache, filename, pageData.Text) {
			stats.Scraped++
			log.Printf("  ✅ Scraped (%d chars) — CHANGED\n", len(pageData.Text))
		} else {
			stats.Skipped++
			log.Printf("  ⏭️  Unchanged — SKIPPED\n")
		}

		UpdateCache(&newCache, filename, pageData)
		params := extractor.ExtractAPIParameters(pageData)
		if len(params) > 0 {
			apiParams[d.DocPath] = params
		}
	}

	stats.Duration = time.Since(start)
	return apiParams, newCache, stats, nil
}

func saveResults(discovered []discovery.DiscoveredURL, stats ScrapeStats, newCache Cache, cacheFile string) {
	log.Printf("\n📊 Scraping complete — total=%d scraped=%d skipped=%d failed=%d duration=%v\n",
		len(discovered), stats.Scraped, stats.Skipped, stats.Failed, stats.Duration.Round(time.Second))
	if err := SaveCache(cacheFile, newCache); err != nil {
		log.Printf("⚠️  Cache save failed: %v\n", err)
	}
}

// regressionSetsMatch returns true when prev and curr describe the same set of
// enum regressions (same path/method/field/missing-values, order-independent).
// Used to decide whether to increment the consecutive counter or reset it.
func regressionSetsMatch(prev, curr []EnumRegression) bool {
	if len(prev) != len(curr) {
		return false
	}
	prevSigs := make(map[string]bool, len(prev))
	for _, r := range prev {
		prevSigs[regressionSig(r)] = true
	}
	for _, r := range curr {
		if !prevSigs[regressionSig(r)] {
			return false
		}
	}
	return true
}

// regressionSig builds a stable, order-independent key for an EnumRegression.
func regressionSig(r EnumRegression) string {
	missing := missingEnumValues(r.OldValues, r.NewValues)
	sort.Strings(missing)
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s", r.Path, r.Method, r.Field, strings.Join(missing, ","))
}

// missingEnumValues returns values present in old but absent in curr.
func missingEnumValues(old, curr []string) []string {
	newSet := make(map[string]bool, len(curr))
	for _, v := range curr {
		newSet[v] = true
	}
	var missing []string
	for _, v := range old {
		if !newSet[v] {
			missing = append(missing, v)
		}
	}
	return missing
}

func notifyAPIChange(result *diff.Result) {
	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_REPO_OWNER")
	repo := os.Getenv("GITHUB_REPO_NAME")
	if token == "" || owner == "" || repo == "" {
		log.Println("   ℹ️  GITHUB_TOKEN/OWNER/REPO not set — skipping GitHub notification")
		return
	}

	ctx := context.Background()
	cfg := notify.Config{Token: token, Owner: owner, Repo: repo}
	client := notify.NewClient(ctx, cfg)

	if err := client.EnsureLabels(ctx, RequiredLabels); err != nil {
		log.Printf("⚠️  Failed to ensure labels: %v\n", err)
	}

	labels := []string{"api-change", "automated"}
	if result.Breaking {
		labels = append(labels, "breaking-change")
	} else {
		labels = append(labels, "non-breaking")
	}

	title := "API Change Detected — " + time.Now().Format("2006-01-02")
	if err := client.NotifyAPIChange(ctx, title, result.Summary, labels); err != nil {
		log.Printf("⚠️  GitHub notification failed: %v\n", err)
	} else {
		log.Println("   ✅ GitHub issue created/updated")
	}
}

// formatCoverageMarkdown produces a simple markdown coverage summary.
// Delegates to analyzer.FormatCoverageReportMarkdown when available.
func formatCoverageMarkdown(r *coverage.Report) string {
	// Re-wrap into analyzer.CoverageReport for the existing formatter.
	ar := &analyzer.CoverageReport{
		Timestamp:       time.Now(),
		TotalAPIFields:  r.TotalAPIFields,
		CoveredFields:   r.CoveredFields,
		MissingFields:   r.MissingFields,
		StaleFields:     r.StaleFields,
		CoveragePercent: r.CoveragePercent,
	}
	for _, g := range r.Gaps {
		ar.Gaps = append(ar.Gaps, analyzer.CoverageGap{
			Type:     analyzer.CoverageGapType(g.GapType),
			Details:  g.Details,
			APIField: g.APIField,
			TFField:  g.TFField,
			Resource: g.Resource,
			Severity: g.Severity,
		})
	}
	return analyzer.FormatCoverageReportMarkdown(ar)
}

// PrintSyncStatus prints a concise sync status to stdout.
func PrintSyncStatus(report *coverage.Report) {
	if report.CoveragePercent >= 80 {
		fmt.Printf("✅ SYNC OK — Provider coverage: %.1f%%\n", report.CoveragePercent)
		return
	}
	fmt.Printf("❌ SYNC FAILED — Coverage %.1f%% < 80%%\n", report.CoveragePercent)
	fmt.Printf("   Missing: %d  |  Stale: %d\n", report.MissingFields, report.StaleFields)
}
