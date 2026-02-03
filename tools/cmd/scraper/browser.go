package main

import (
	"fmt"
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// launchBrowser launches a new browser instance with resource blocking and returns cleanup function
// The cleanup function ensures browser is always closed, even on panic
func launchBrowser(config ScraperConfig) (*rod.Browser, func(), error) {
	launchURL, err := launcher.New().
		Headless(config.Headless).
		Launch()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(launchURL)
	if err := browser.Connect(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	// Setup resource blocking to improve performance (3-5x faster)
	if config.ResourceBlocking {
		setupResourceBlocking(browser)
	}

	// Cleanup function that always runs
	cleanup := func() {
		if browser != nil {
			if err := browser.Close(); err != nil {
				log.Printf("⚠️  Browser cleanup error: %v\n", err)
			} else {
				log.Println("✅ Browser closed successfully")
			}
		}
	}

	return browser, cleanup, nil
}

// setupResourceBlocking blocks unnecessary resources (images, CSS, fonts)
// This significantly improves scraping performance and reduces bandwidth
func setupResourceBlocking(browser *rod.Browser) {
	router := browser.HijackRequests()

	// Block images
	router.MustAdd("*.png", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})
	router.MustAdd("*.jpg", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})
	router.MustAdd("*.jpeg", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})
	router.MustAdd("*.gif", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})
	router.MustAdd("*.webp", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})
	router.MustAdd("*.svg", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})

	// Block CSS
	router.MustAdd("*.css", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})

	// Block fonts
	router.MustAdd("*.woff", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})
	router.MustAdd("*.woff2", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})
	router.MustAdd("*.ttf", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})
	router.MustAdd("*.otf", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})
	router.MustAdd("*.eot", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})

	// Start the hijack router in background
	go router.Run()

	log.Println("🚫 Resource blocking enabled (images, CSS, fonts)")
}
