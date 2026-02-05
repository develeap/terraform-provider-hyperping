// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import "regexp"

// HTTP header names.
const (
	// HeaderContentType is the Content-Type HTTP header.
	HeaderContentType = "Content-Type"

	// HeaderAuthorization is the Authorization HTTP header.
	HeaderAuthorization = "Authorization"

	// HeaderAccept is the Accept HTTP header.
	HeaderAccept = "Accept"
)

// Content type values.
const (
	// ContentTypeJSON is the MIME type for JSON content.
	ContentTypeJSON = "application/json"
)

// Authentication constants.
const (
	// BearerPrefix is the prefix for Bearer token authentication.
	BearerPrefix = "Bearer "

	// APIKeyPrefix is the prefix for Hyperping API keys.
	APIKeyPrefix = "sk_"

	// APIKeyPatternString is the regex pattern string for matching API keys.
	APIKeyPatternString = `sk_[a-zA-Z0-9_-]+` // #nosec G101 -- regex pattern for validation, not a credential
)

// APIKeyPattern is the compiled regex for matching Hyperping API keys.
var APIKeyPattern = regexp.MustCompile(APIKeyPatternString)
