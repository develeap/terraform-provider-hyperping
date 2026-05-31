// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNoControlCharacters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     types.String
		wantError bool
	}{
		{"clean header name", types.StringValue("X-Custom-Header"), false},
		{"clean header value", types.StringValue("Bearer token123"), false},
		{"LF injection", types.StringValue("X-Injected\nEvil: payload"), true},
		{"CR injection", types.StringValue("X-Injected\rEvil: payload"), true},
		{"CRLF injection", types.StringValue("X-Injected\r\nEvil: payload"), true},
		{"NULL byte", types.StringValue("X-Null\x00Byte"), true},
		{"empty string", types.StringValue(""), false},
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := NoControlCharacters("test message")
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tt.input,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("NoControlCharacters(%v): got error=%v, want error=%v",
					tt.input, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestReservedHeaderName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     types.String
		wantError bool
	}{
		{"Host blocked", types.StringValue("Host"), true},
		{"host lowercase blocked", types.StringValue("host"), true},
		{"Transfer-Encoding blocked", types.StringValue("Transfer-Encoding"), true},
		{"transfer-encoding lowercase blocked", types.StringValue("transfer-encoding"), true},
		{"Content-Length blocked (smuggling pair)", types.StringValue("Content-Length"), true},
		{"content-length lowercase blocked", types.StringValue("content-length"), true},
		{"Connection blocked (hop-by-hop)", types.StringValue("Connection"), true},
		{"Upgrade blocked (protocol switch)", types.StringValue("Upgrade"), true},
		{"TE blocked (TE negotiation)", types.StringValue("TE"), true},
		{"te lowercase blocked", types.StringValue("te"), true},
		{"Trailer blocked (smuggling-related)", types.StringValue("Trailer"), true},
		{"Expect blocked (100-continue)", types.StringValue("Expect"), true},
		{"Authorization allowed (issue #132)", types.StringValue("Authorization"), false},
		{"authorization lowercase allowed", types.StringValue("authorization"), false},
		{"Cookie allowed", types.StringValue("Cookie"), false},
		{"Set-Cookie allowed (response-only header, harmless)", types.StringValue("Set-Cookie"), false},
		{"Proxy-Authorize allowed (no proxy on probe path)", types.StringValue("Proxy-Authorize"), false},
		{"X-Forwarded-For allowed (forwarding metadata)", types.StringValue("X-Forwarded-For"), false},
		{"Forwarded allowed (forwarding metadata)", types.StringValue("Forwarded"), false},
		{"Range allowed (range requests)", types.StringValue("Range"), false},
		{"custom header allowed", types.StringValue("X-Custom"), false},
		{"Accept allowed", types.StringValue("Accept"), false},
		{"Content-Type allowed", types.StringValue("Content-Type"), false},
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := ReservedHeaderName()
			req := validator.StringRequest{
				Path:        path.Root("name"),
				ConfigValue: tt.input,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("ReservedHeaderName(%v): got error=%v, want error=%v",
					tt.input, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestISO8601(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     types.String
		wantError bool
	}{
		{"valid with Z", types.StringValue("2026-01-29T10:00:00Z"), false},
		{"valid with timezone", types.StringValue("2026-01-29T10:00:00+05:00"), false},
		{"valid end of year", types.StringValue("2026-12-31T23:59:59Z"), false},
		{"invalid - missing time", types.StringValue("2026-01-29"), true},
		{"invalid - space separator", types.StringValue("2026-01-29 10:00:00"), true},
		{"invalid - not a date", types.StringValue("invalid"), true},
		{"invalid - empty", types.StringValue(""), true},
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := ISO8601()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tt.input,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("ISO8601(%v): got error=%v, want error=%v",
					tt.input, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestUUIDFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     types.String
		wantError bool
	}{
		{"standard UUID", types.StringValue("550e8400-e29b-41d4-a716-446655440000"), false},
		{"monitor ID", types.StringValue("mon_abc123def456"), false},
		{"token ID", types.StringValue("tok_xyz789abc"), false},
		{"outage ID", types.StringValue("out_incident123"), false},
		{"incident ID", types.StringValue("inc_test123"), false},
		{"short monitor ID", types.StringValue("mon_123"), false},
		{"short statuspage ID", types.StringValue("sp_001"), false},
		{"invalid - no separators", types.StringValue("short"), true},
		{"invalid - no separators long", types.StringValue("nodashesorunderscores"), true},
		{"invalid - empty", types.StringValue(""), true},
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := UUIDFormat()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tt.input,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("UUIDFormat(%v): got error=%v, want error=%v",
					tt.input, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestURLFormatValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"valid https", types.StringValue("https://example.com"), false},
		{"valid http", types.StringValue("http://api.example.com/health"), false},
		{"valid with path", types.StringValue("https://example.com/api/v1/health"), false},
		{"valid with port", types.StringValue("https://example.com:8080"), false},
		{"valid with query", types.StringValue("https://example.com/api?key=value"), false},
		{"valid with fragment", types.StringValue("https://example.com/page#section"), false},
		{"invalid ftp", types.StringValue("ftp://example.com"), true},
		{"invalid no scheme", types.StringValue("example.com"), true},
		{"invalid typo", types.StringValue("htp://example.com"), true},
		{"invalid empty", types.StringValue(""), true},
		{"invalid just scheme", types.StringValue("https://"), true},
		{"invalid file scheme", types.StringValue("file:///path/to/file"), true},
		{"invalid javascript", types.StringValue("javascript:alert(1)"), true},
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := URLFormat()
			req := validator.StringRequest{
				Path:        path.Root("url"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("URLFormat(%v): got error=%v, want error=%v",
					tt.value, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestStringLengthValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		min       int
		max       int
		value     types.String
		wantError bool
	}{
		{"valid in range", 1, 255, types.StringValue("Valid Name"), false},
		{"valid at min", 1, 255, types.StringValue("X"), false},
		{"valid at max", 1, 255, types.StringValue(string(make([]byte, 255))), false},
		{"invalid too short", 1, 255, types.StringValue(""), true},
		{"invalid too long", 1, 255, types.StringValue(string(make([]byte, 256))), true},
		{"valid exact length", 5, 5, types.StringValue("hello"), false},
		{"invalid one under min", 5, 10, types.StringValue("test"), true},
		{"invalid one over max", 5, 10, types.StringValue("hello world"), true},
		{"valid unicode single char", 1, 10, types.StringValue("\U0001f600"), false},
		{"valid unicode multi char", 1, 10, types.StringValue("Hello \u4e16\u754c"), false},
		{"invalid unicode too long", 1, 5, types.StringValue("Hello \u4e16\u754c"), true},
		{"null value", 1, 255, types.StringNull(), false},
		{"unknown value", 1, 255, types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := StringLength(tt.min, tt.max)
			req := validator.StringRequest{
				Path:        path.Root("name"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("StringLength(%d,%d)(%v): got error=%v, want error=%v",
					tt.min, tt.max, tt.value, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestCronExpressionValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"valid daily midnight", types.StringValue("0 0 * * *"), false},
		{"valid every 15 min", types.StringValue("*/15 * * * *"), false},
		{"valid business hours", types.StringValue("0 9-17 * * 1-5"), false},
		{"valid specific time", types.StringValue("30 14 * * *"), false},
		{"valid every hour", types.StringValue("0 * * * *"), false},
		{"valid complex", types.StringValue("0,30 9-17 * * 1-5"), false},
		{"invalid text", types.StringValue("invalid"), true},
		{"invalid minute 60", types.StringValue("60 * * * *"), true},
		{"invalid day 32", types.StringValue("* * 32 * *"), true},
		{"invalid empty", types.StringValue(""), true},
		{"invalid too few fields", types.StringValue("0 0 *"), true},
		{"invalid too many fields", types.StringValue("0 0 * * * * *"), true},
		{"invalid month 13", types.StringValue("0 0 * 13 *"), true},
		{"invalid weekday 8", types.StringValue("0 0 * * 8"), true},
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := CronExpression()
			req := validator.StringRequest{
				Path:        path.Root("cron"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("CronExpression(%v): got error=%v, want error=%v",
					tt.value, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestTimezoneValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"valid UTC", types.StringValue("UTC"), false},
		{"valid America/New_York", types.StringValue("America/New_York"), false},
		{"valid Europe/London", types.StringValue("Europe/London"), false},
		{"valid Asia/Tokyo", types.StringValue("Asia/Tokyo"), false},
		{"valid America/Los_Angeles", types.StringValue("America/Los_Angeles"), false},
		{"valid Europe/Paris", types.StringValue("Europe/Paris"), false},
		{"valid Australia/Sydney", types.StringValue("Australia/Sydney"), false},
		{"valid Africa/Cairo", types.StringValue("Africa/Cairo"), false},
		{"valid EST", types.StringValue("EST"), false},
		{"valid Local", types.StringValue("Local"), false},
		{"invalid New York", types.StringValue("New York"), true},
		{"invalid random", types.StringValue("RandomTimezone"), true},
		{"invalid number", types.StringValue("12345"), true},
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := Timezone()
			req := validator.StringRequest{
				Path:        path.Root("timezone"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("Timezone(%v): got error=%v, want error=%v",
					tt.value, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestPortRangeValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.Int64
		wantError bool
	}{
		{"valid port 80", types.Int64Value(80), false},
		{"valid port 443", types.Int64Value(443), false},
		{"valid port 1", types.Int64Value(1), false},
		{"valid port 65535", types.Int64Value(65535), false},
		{"valid port 3000", types.Int64Value(3000), false},
		{"valid port 8080", types.Int64Value(8080), false},
		{"valid port 22", types.Int64Value(22), false},
		{"valid port 3306", types.Int64Value(3306), false},
		{"valid port 5432", types.Int64Value(5432), false},
		{"valid port 27017", types.Int64Value(27017), false},
		{"invalid port 0", types.Int64Value(0), true},
		{"invalid port -1", types.Int64Value(-1), true},
		{"invalid port -100", types.Int64Value(-100), true},
		{"invalid port 65536", types.Int64Value(65536), true},
		{"invalid port 100000", types.Int64Value(100000), true},
		{"invalid port 70000", types.Int64Value(70000), true},
		{"null value", types.Int64Null(), false},
		{"unknown value", types.Int64Unknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := PortRange()
			req := validator.Int64Request{
				Path:        path.Root("port"),
				ConfigValue: tt.value,
			}
			resp := &validator.Int64Response{}
			v.ValidateInt64(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("PortRange(%v): got error=%v, want error=%v",
					tt.value, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestHexColorValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"valid lowercase", types.StringValue("#ff5733"), false},
		{"valid uppercase", types.StringValue("#FF5733"), false},
		{"valid mixed case", types.StringValue("#Ff5733"), false},
		{"valid black", types.StringValue("#000000"), false},
		{"valid white", types.StringValue("#ffffff"), false},
		{"valid white uppercase", types.StringValue("#FFFFFF"), false},
		{"valid gray", types.StringValue("#808080"), false},
		{"valid default hyperping", types.StringValue("#36b27e"), false},
		{"valid all digits", types.StringValue("#123456"), false},
		{"valid all letters", types.StringValue("#abcdef"), false},
		{"valid mixed", types.StringValue("#a1b2c3"), false},
		{"invalid no hash", types.StringValue("ff5733"), true},
		{"invalid 3 digit", types.StringValue("#fff"), true},
		{"invalid 8 digit", types.StringValue("#ff5733aa"), true},
		{"invalid with alpha", types.StringValue("#ff5733ff"), true},
		{"invalid chars g", types.StringValue("#gggggg"), true},
		{"invalid chars z", types.StringValue("#zzzzzz"), true},
		{"invalid empty", types.StringValue(""), true},
		{"invalid just hash", types.StringValue("#"), true},
		{"invalid 5 digit", types.StringValue("#12345"), true},
		{"invalid 7 digit", types.StringValue("#1234567"), true},
		{"invalid space", types.StringValue("#ff 5733"), true},
		{"invalid special char", types.StringValue("#ff57@3"), true},
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := HexColor()
			req := validator.StringRequest{
				Path:        path.Root("accent_color"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("HexColor(%v): got error=%v, want error=%v",
					tt.value, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestEmailFormatValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"valid simple", types.StringValue("user@example.com"), false},
		{"valid subdomain", types.StringValue("user@mail.example.com"), false},
		{"valid plus", types.StringValue("user+tag@example.com"), false},
		{"valid dots", types.StringValue("first.last@example.com"), false},
		{"valid numbers", types.StringValue("user123@example456.com"), false},
		{"valid hyphen", types.StringValue("user@ex-ample.com"), false},
		{"valid underscore", types.StringValue("user_name@example.com"), false},
		{"valid percent", types.StringValue("user%test@example.com"), false},
		{"valid uppercase", types.StringValue("User@Example.COM"), false},
		{"invalid no @", types.StringValue("userexample.com"), true},
		{"invalid no domain", types.StringValue("user@"), true},
		{"invalid no user", types.StringValue("@example.com"), true},
		{"invalid no tld", types.StringValue("user@example"), true},
		{"invalid spaces", types.StringValue("user @example.com"), true},
		{"invalid double @", types.StringValue("user@@example.com"), true},
		{"invalid empty", types.StringValue(""), true},
		{"invalid special chars", types.StringValue("user<>@example.com"), true},
		{"invalid no domain part", types.StringValue("user@.com"), true},
		{"invalid tld too short", types.StringValue("user@example.c"), true},
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := EmailFormat()
			req := validator.StringRequest{
				Path:        path.Root("email"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("EmailFormat(%v): got error=%v, want error=%v",
					tt.value, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestAlertsWaitValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.Int64
		wantError bool
	}{
		{"valid -1 (disabled)", types.Int64Value(-1), false},
		{"valid 0 (immediate)", types.Int64Value(0), false},
		{"valid 1 minute", types.Int64Value(1), false},
		{"valid 2 minutes", types.Int64Value(2), false},
		{"valid 3 minutes", types.Int64Value(3), false},
		{"valid 5 minutes", types.Int64Value(5), false},
		{"valid 10 minutes", types.Int64Value(10), false},
		{"valid 30 minutes", types.Int64Value(30), false},
		{"valid 60 minutes", types.Int64Value(60), false},
		{"invalid 300 (original bug)", types.Int64Value(300), true},
		{"invalid -2", types.Int64Value(-2), true},
		{"invalid 100", types.Int64Value(100), true},
		{"invalid 7", types.Int64Value(7), true},
		{"invalid 4", types.Int64Value(4), true},
		{"invalid 15", types.Int64Value(15), true},
		{"invalid 120", types.Int64Value(120), true},
		{"null value", types.Int64Null(), false},
		{"unknown value", types.Int64Unknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := AlertsWait()
			req := validator.Int64Request{
				Path:        path.Root("alerts_wait"),
				ConfigValue: tt.value,
			}
			resp := &validator.Int64Response{}
			v.ValidateInt64(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("AlertsWait(%v): got error=%v, want error=%v",
					tt.value, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

func TestStatusCodePatternValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		// Valid specific codes
		{"valid specific 200", types.StringValue("200"), false},
		{"valid specific 404", types.StringValue("404"), false},
		{"valid specific 599", types.StringValue("599"), false},
		{"valid specific 100", types.StringValue("100"), false},
		{"valid specific 503", types.StringValue("503"), false},

		// Valid single wildcards
		{"valid wildcard 2xx", types.StringValue("2xx"), false},
		{"valid wildcard 5xx", types.StringValue("5xx"), false},
		{"valid wildcard 1xx", types.StringValue("1xx"), false},
		{"valid wildcard 3xx", types.StringValue("3xx"), false},
		{"valid wildcard 4xx", types.StringValue("4xx"), false},

		// Valid multi-range patterns
		{"valid range 1xx-3xx", types.StringValue("1xx-3xx"), false},
		{"valid range 2xx-5xx", types.StringValue("2xx-5xx"), false},
		{"valid range same class 2xx-2xx", types.StringValue("2xx-2xx"), false},
		{"valid range 1xx-5xx", types.StringValue("1xx-5xx"), false},
		{"valid range 3xx-4xx", types.StringValue("3xx-4xx"), false},

		// Invalid: inverted range
		{"invalid inverted range 3xx-1xx", types.StringValue("3xx-1xx"), true},
		{"invalid inverted range 5xx-2xx", types.StringValue("5xx-2xx"), true},

		// Invalid: class out of range
		{"invalid class 6xx", types.StringValue("6xx"), true},
		{"invalid class 0xx", types.StringValue("0xx"), true},
		{"invalid code 600", types.StringValue("600"), true},
		{"invalid code 099", types.StringValue("099"), true},

		// Invalid: malformed patterns
		{"invalid text abc", types.StringValue("abc"), true},
		{"invalid too long 2xxx", types.StringValue("2xxx"), true},
		{"invalid too short 20", types.StringValue("20"), true},
		{"invalid incomplete 2x", types.StringValue("2x"), true},
		{"invalid empty string", types.StringValue(""), true},
		{"invalid spaces", types.StringValue("2 xx"), true},
		{"invalid mixed 2xX", types.StringValue("2xX"), true},
		{"invalid range with specific 200-3xx", types.StringValue("200-3xx"), true},
		{"invalid uppercase 2XX", types.StringValue("2XX"), true},
		{"invalid trailing dash 1xx-", types.StringValue("1xx-"), true},
		{"invalid leading dash -3xx", types.StringValue("-3xx"), true},
		{"invalid range endpoint 1xx-6xx", types.StringValue("1xx-6xx"), true},
		{"invalid trailing dash after wildcard 2xx-", types.StringValue("2xx-"), true},

		// Null/Unknown
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := StatusCodePattern()
			req := validator.StringRequest{
				Path:        path.Root("expected_status_code"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("StatusCodePattern(%v): got error=%v, want error=%v",
					tt.value, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

// TestStringValidatorDescriptions verifies Description and MarkdownDescription
// for all string validators return non-empty values.
func TestStringValidatorDescriptions(t *testing.T) {
	t.Parallel()

	validators := []struct {
		name          string
		validator     validator.String
		descMatchesMD bool
	}{
		{"NoControlCharacters", NoControlCharacters("test"), true},
		{"ReservedHeaderName", ReservedHeaderName(), true},
		{"ISO8601", ISO8601(), true},
		{"UUIDFormat", UUIDFormat(), true},
		{"URLFormat", URLFormat(), true},
		{"StringLength", StringLength(1, 255), true},
		{"CronExpression", CronExpression(), false},
		{"Timezone", Timezone(), false},
		{"HexColor", HexColor(), false},
		{"EmailFormat", EmailFormat(), true},
		{"StatusCodePattern", StatusCodePattern(), false},
	}

	ctx := context.Background()
	for _, tt := range validators {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			desc := tt.validator.Description(ctx)
			if desc == "" {
				t.Error("expected non-empty description")
			}

			mdDesc := tt.validator.MarkdownDescription(ctx)
			if mdDesc == "" {
				t.Error("expected non-empty markdown description")
			}

			if tt.descMatchesMD && desc != mdDesc {
				t.Errorf("description and markdown description should match: %q != %q", desc, mdDesc)
			}
		})
	}
}

// TestInt64ValidatorDescriptions verifies Description and MarkdownDescription
// for all int64 validators return non-empty values.
func TestInt64ValidatorDescriptions(t *testing.T) {
	t.Parallel()

	validators := []struct {
		name          string
		validator     validator.Int64
		descMatchesMD bool
	}{
		{"PortRange", PortRange(), true},
		{"AlertsWait", AlertsWait(), true},
	}

	ctx := context.Background()
	for _, tt := range validators {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			desc := tt.validator.Description(ctx)
			if desc == "" {
				t.Error("expected non-empty description")
			}

			mdDesc := tt.validator.MarkdownDescription(ctx)
			if mdDesc == "" {
				t.Error("expected non-empty markdown description")
			}

			if tt.descMatchesMD && desc != mdDesc {
				t.Errorf("description and markdown description should match: %q != %q", desc, mdDesc)
			}
		})
	}
}
