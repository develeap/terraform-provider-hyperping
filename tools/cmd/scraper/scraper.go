package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/go-rod/rod"
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

	// Remove noise elements.
	page.Eval(`() => { document.querySelectorAll('script,style,noscript').forEach(e=>e.remove()); }`)

	title := ""
	if el, err := page.Timeout(5 * time.Second).Element("title"); err == nil {
		title, _ = el.Text()
	}
	body := ""
	if el, err := page.Timeout(5 * time.Second).Element("body"); err == nil {
		body, _ = el.Text()
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
