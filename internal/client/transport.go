// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
)

// defaultTLSConfig returns a TLS configuration restricted to AEAD cipher suites
// and TLS 1.2+. This prevents weak-cipher negotiation attacks (VULN-008).
func defaultTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
	}
}

// authTransport injects the Bearer token Authorization header into each request.
// The token is held as []byte to avoid persistent string allocations in the
// Client struct (VULN-009). One string copy is created per RoundTrip — this is
// unavoidable due to Go's http.Header requiring string values — but it is
// immediately eligible for GC after the request completes, unlike the previous
// SecureString pattern which accumulated copies across all retries.
type authTransport struct {
	token []byte
	next  http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request before mutation — RoundTrip contract requires this.
	clone := req.Clone(req.Context())
	clone.Header.Set(HeaderAuthorization, BearerPrefix+string(t.token))
	return t.next.RoundTrip(clone)
}

// tlsEnforcedTransport wraps a non-standard RoundTripper and ensures only HTTPS
// requests are sent. When WithHTTPClient provides a transport that is not
// *http.Transport (so we cannot apply cipher suite restrictions directly),
// this wrapper at minimum blocks cleartext HTTP to prevent credential leakage
// (VULN-011, VULN-016).
type tlsEnforcedTransport struct {
	next http.RoundTripper
}

func (t *tlsEnforcedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" {
		return nil, fmt.Errorf("tlsEnforcedTransport: HTTPS required for %s, got %s://", req.URL.Host, req.URL.Scheme)
	}
	return t.next.RoundTrip(req)
}

// buildTransportChain constructs the layered transport:
//
//	baseTransport → [TLS enforcement] → authTransport
//
// For *http.Transport: TLS config is applied directly (cipher suites + min version).
// For other transports: wrapped in tlsEnforcedTransport to block non-HTTPS.
// Localhost URLs (testing) are exempt from TLS enforcement.
func buildTransportChain(apiKey []byte, baseTransport http.RoundTripper, baseURL string) http.RoundTripper {
	enforced := enforceTLS(baseTransport, baseURL)
	return &authTransport{
		token: apiKey,
		next:  enforced,
	}
}

// enforceTLS applies TLS restrictions to the transport based on the base URL.
// Localhost targets (used in tests with httptest) are exempt.
func enforceTLS(transport http.RoundTripper, baseURL string) http.RoundTripper {
	isLocalhost := strings.Contains(baseURL, "localhost") ||
		strings.Contains(baseURL, "127.0.0.1") ||
		strings.Contains(baseURL, "[::1]")

	if isLocalhost {
		return transport
	}

	if !strings.HasPrefix(baseURL, "https://") {
		// Non-HTTPS, non-localhost: wrap to block cleartext
		return &tlsEnforcedTransport{next: transport}
	}

	// HTTPS target: apply TLS config
	httpTransport, ok := transport.(*http.Transport)
	if !ok {
		// Custom transport — wrap to enforce HTTPS-only
		return &tlsEnforcedTransport{next: transport}
	}

	// Apply our TLS config to the standard transport
	httpTransport.TLSClientConfig = defaultTLSConfig()
	return httpTransport
}
