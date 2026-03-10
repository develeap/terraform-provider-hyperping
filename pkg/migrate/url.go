// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import "strings"

// EnsureURLScheme prepends "https://" if the URL has no HTTP/HTTPS scheme.
// The Hyperping provider requires all URLs to have an HTTP/HTTPS scheme,
// even for ICMP and port monitors.
func EnsureURLScheme(rawURL string) string {
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}
	return "https://" + rawURL
}
