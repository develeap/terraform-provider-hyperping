package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
)

// scrapeWithRetry attempts to scrape a page with exponential backoff.
func scrapeWithRetry(ctx context.Context, page *rod.Page, url string, maxRetries int, timeout time.Duration) (*extractor.PageData, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			log.Printf("  ⏳ Retry %d/%d after %v...\n", attempt+1, maxRetries, backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		pageData, err := scrapePage(ctx, page, url, timeout)
		if err == nil {
			return pageData, nil
		}

		lastErr = err
		if !isRetryableError(err) {
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}
		log.Printf("  ⚠️  Attempt %d failed: %v\n", attempt+1, err)
	}

	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

// scrapePage navigates to url, waits for JS render, and extracts page content.
func scrapePage(ctx context.Context, page *rod.Page, url string, timeout time.Duration) (*extractor.PageData, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := page.Context(timeoutCtx).Navigate(url); err != nil {
		return nil, fmt.Errorf("navigation failed: %w", err)
	}
	if err := page.Context(timeoutCtx).WaitLoad(); err != nil {
		return nil, fmt.Errorf("page load timeout: %w", err)
	}
	if err := page.Context(timeoutCtx).WaitStable(100 * time.Millisecond); err != nil {
		log.Printf("  ⚠️  Page stability timeout (continuing)\n")
	}

	// Scroll incrementally to trigger IntersectionObserver callbacks on Notion-based
	// docs pages. Without this, lazily-rendered elements (e.g. enum option nodes near
	// the bottom of a long parameter list) are never added to the DOM and are missed
	// by the HTML extractor.
	scrollToRevealContent(timeoutCtx, page)

	// Remove noise elements. Failure here is non-fatal; the page can still be scraped.
	if _, err := page.Eval(`() => { document.querySelectorAll('script,style,noscript').forEach(e=>e.remove()); }`); err != nil {
		GetLogger().Warn("DOM cleanup eval failed", map[string]interface{}{"error": err.Error()})
	}

	title := ""
	if el, err := page.Timeout(5 * time.Second).Element("title"); err == nil {
		title, _ = el.Text() //nolint:errcheck
	}
	body := ""
	if el, err := page.Timeout(5 * time.Second).Element("body"); err == nil {
		body, _ = el.Text() //nolint:errcheck
	}
	html, err := page.HTML()
	if err != nil {
		return nil, fmt.Errorf("get HTML: %w", err)
	}

	return &extractor.PageData{
		URL:       url,
		Title:     strings.TrimSpace(title),
		Text:      strings.TrimSpace(body),
		HTML:      html,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// scrollToRevealContent scrolls the page incrementally so that IntersectionObserver
// callbacks fire for every element, forcing full DOM materialisation on lazy-rendered
// pages (e.g. Hyperping's Notion-based API docs).
//
// A dedicated 8 s sub-context caps scroll time so the outer page timeout is not
// exhausted. Scrolling stops as soon as the page bottom is reached, so short pages
// (the common case) finish in well under 1 s instead of always spending ~5 s.
func scrollToRevealContent(ctx context.Context, page *rod.Page) {
	// Cap scroll budget independently of the outer page timeout.
	// 8 s allows up to 40 × 200 ms steps before giving up.
	scrollCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	const stepPx = 800

	for {
		if scrollCtx.Err() != nil {
			return
		}
		if _, err := page.Context(scrollCtx).Eval(fmt.Sprintf(`() => window.scrollBy(0, %d)`, stepPx)); err != nil {
			log.Printf("  ⚠️  Scroll step failed: %v\n", err)
			return
		}

		// Wait for IntersectionObserver callbacks triggered by the scroll.
		select {
		case <-time.After(200 * time.Millisecond):
		case <-scrollCtx.Done():
			return
		}

		// Stop once we've reached the bottom — avoids wasting time on short pages.
		ret, err := page.Context(scrollCtx).Eval(
			`() => window.scrollY + window.innerHeight >= document.documentElement.scrollHeight - 10`)
		if err == nil && ret.Value.Bool() {
			break
		}
	}

	// Return to top so downstream HTML extraction starts from a known position.
	if _, err := page.Context(scrollCtx).Eval(`() => window.scrollTo(0, 0)`); err != nil {
		log.Printf("  ⚠️  Scroll reset failed: %v\n", err)
	}
	// Brief settle to allow any trailing renders triggered by the scroll reset.
	select {
	case <-time.After(300 * time.Millisecond):
	case <-scrollCtx.Done():
	}
}

// isRetryableError returns false for client errors (4xx) that should not be retried.
func isRetryableError(err error) bool {
	s := strings.ToLower(err.Error())
	for _, pattern := range []string{"404", "403", "401", "400", "not found", "forbidden", "unauthorized", "bad request"} {
		if strings.Contains(s, pattern) {
			return false
		}
	}
	return true
}
