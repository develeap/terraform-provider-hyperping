// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package discovery locates Hyperping API documentation URLs via sitemap.xml.
// This replaces the headless-browser approach: the sitemap already enumerates
// every /docs/api/* path, so no browser-based navigation is required.
package discovery

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	defaultSitemapURL = "https://hyperping.com/sitemap.xml"
	apiPathPrefix     = "/docs/api/"
	httpTimeout       = 15 * time.Second
)

// DiscoveredURL is a single API documentation page found in the sitemap.
type DiscoveredURL struct {
	URL     string // Full URL, e.g. "https://hyperping.com/docs/api/monitors/create"
	DocPath string // Relative path used as mapping key, e.g. "monitors/create"
	Section string // Top-level resource section, e.g. "monitors"
}

// sitemapXML is used to decode sitemap.xml.
type sitemapXML struct {
	URLs []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

// DiscoverFromSitemap fetches the Hyperping sitemap and returns all /docs/api/* entries.
// Pass an empty string to use the default sitemap URL.
func DiscoverFromSitemap(sitemapURL string) ([]DiscoveredURL, error) {
	if sitemapURL == "" {
		sitemapURL = defaultSitemapURL
	}

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(sitemapURL)
	if err != nil {
		return nil, fmt.Errorf("discovery: fetch sitemap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery: sitemap returned HTTP %d", resp.StatusCode)
	}

	var sm sitemapXML
	if err := xml.NewDecoder(resp.Body).Decode(&sm); err != nil {
		return nil, fmt.Errorf("discovery: parse sitemap XML: %w", err)
	}

	return filterAPIURLs(sm), nil
}

// filterAPIURLs extracts /docs/api/* entries and converts them to DiscoveredURL.
func filterAPIURLs(sm sitemapXML) []DiscoveredURL {
	var results []DiscoveredURL
	seen := make(map[string]bool)

	for _, u := range sm.URLs {
		loc := u.Loc
		idx := strings.Index(loc, apiPathPrefix)
		if idx == -1 {
			continue
		}

		docPath := loc[idx+len(apiPathPrefix):]
		docPath = strings.TrimSuffix(docPath, "/")
		if docPath == "" || seen[docPath] {
			continue
		}

		seen[docPath] = true
		section := strings.SplitN(docPath, "/", 2)[0]

		results = append(results, DiscoveredURL{
			URL:     loc,
			DocPath: docPath,
			Section: section,
		})
	}

	return results
}
