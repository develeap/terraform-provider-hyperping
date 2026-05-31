// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import (
	"fmt"
	"regexp"
)

// safeUUIDPattern matches identifiers we are willing to interpolate into
// generated shell scripts. The Hyperping backend currently emits resource
// UUIDs as hex with optional hyphens or short prefixed identifiers
// (e.g. "mon_abc123"). This pattern is intentionally narrow: a single
// non-conforming byte rejects the entire value rather than attempt partial
// sanitisation.
var safeUUIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// SafeUUID returns (value, true) when the input is a plausibly server-issued
// UUID-shaped identifier (alphanumerics, underscores, hyphens). Otherwise it
// returns ("", false). This is defense-in-depth: the API is the source of
// truth for these identifiers, but if a future regression or partner-API
// forwarding ever surfaces an attacker-influenced value, we want callers to
// have an explicit boolean signal to refuse interpolating it into a shell
// script or HCL document.
func SafeUUID(s string) (string, bool) {
	if s == "" || !safeUUIDPattern.MatchString(s) {
		return "", false
	}
	return s, true
}

// QuoteShellUUID emits a UUID-shaped identifier as a Go-quoted (and thus
// bash double-quote-safe) literal. Inputs that do not match the safe UUID
// pattern are replaced with the sentinel "INVALID_UUID_<hex>" before quoting,
// so a malformed server response cannot smuggle command substitution
// ($(...), ``...``), variable expansion ($VAR), or statement chaining (;) into
// the generated shell script.
//
// The sentinel preserves the script's structure (Terraform will fail the
// import with a clear "resource not found" error), which is preferable to
// silently emitting a corrupted value.
func QuoteShellUUID(s string) string {
	if _, ok := SafeUUID(s); !ok {
		return fmt.Sprintf("%q", "INVALID_UUID_REJECTED_BY_SANITIZER")
	}
	// %q is safe here because the input is restricted to [A-Za-z0-9_-], which
	// contains no Go-string-escape characters and no bash metacharacters.
	return fmt.Sprintf("%q", s)
}
