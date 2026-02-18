package main

import (
	"net/url"
	"regexp"
	"strings"
)

// safeFilenameRe replaces any characters that are not alphanumeric, underscore,
// or hyphen with an underscore. Hyphens are preserved so that identifiers like
// "API-123" round-trip cleanly.
var safeFilenameRe = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

// URLToFilename converts a URL to a safe filename by stripping the base path
// and replacing unsafe characters with underscores.
func URLToFilename(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "unknown.json"
	}

	path := u.Path
	stripped := false

	// Strip known base paths.
	for _, prefix := range []string{"/docs/api/", "/docs/api"} {
		if strings.HasPrefix(path, prefix) {
			path = path[len(prefix):]
			stripped = true
			break
		}
	}

	// If nothing remains after stripping the base path, use "index".
	// If no base path was matched and path is empty, fall back to the
	// last path segment of the original URL.
	if path == "" {
		if stripped {
			path = "index"
		} else {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			path = parts[len(parts)-1]
		}
	}

	// Include query string in filename to avoid collisions.
	if u.RawQuery != "" {
		path += "_" + u.RawQuery
	}

	// Clean to safe characters (alphanumeric, underscore, hyphen).
	path = safeFilenameRe.ReplaceAllString(path, "_")
	path = strings.Trim(path, "_")

	if path == "" {
		path = "index"
	}

	return path + ".json"
}

// FilterNewURLs returns only those discovered URLs that are not yet in the cache.
func FilterNewURLs(discovered []DiscoveredURL, cache Cache) []DiscoveredURL {
	var filtered []DiscoveredURL
	for _, d := range discovered {
		filename := URLToFilename(d.URL)
		if _, exists := cache.Entries[filename]; !exists {
			filtered = append(filtered, d)
		}
	}
	return filtered
}
