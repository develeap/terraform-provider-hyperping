// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure HyperpingProvider satisfies the provider.Provider interface.
var _ provider.Provider = &HyperpingProvider{}

// HyperpingProvider defines the provider implementation.
type HyperpingProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// HyperpingProviderModel describes the provider data model.
type HyperpingProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

// Metadata returns the provider type name.
func (p *HyperpingProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hyperping"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *HyperpingProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Hyperping provider allows you to manage Hyperping monitors, incidents, and maintenance windows.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Hyperping API key (starts with `sk_`). Can also be set via `HYPERPING_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Hyperping API base URL. Defaults to `https://api.hyperping.io`.",
				Optional:            true,
			},
		},
	}
}

// Configure prepares a Hyperping API client for data sources and resources.
func (p *HyperpingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config HyperpingProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values from environment variables
	apiKey := os.Getenv("HYPERPING_API_KEY")
	baseURL := client.DefaultBaseURL

	// Override with config values if provided
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()

		// SECURITY: Enforce domain allowlist AND HTTPS requirement to prevent credential theft.
		// isAllowedBaseURL checks both the domain (*.hyperping.io only) and protocol (HTTPS required,
		// except for localhost which is exempt for testing).
		if !isAllowedBaseURL(baseURL) {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_url"),
				"Invalid Base URL",
				"The base_url must be an HTTPS URL pointing to a Hyperping API domain (*.hyperping.io) "+
					"to protect your API credentials from being sent to unauthorized servers. "+
					fmt.Sprintf("Provided: %s. Expected: https://api.hyperping.io", baseURL),
			)
			return
		}
	}

	// Validate API key is set
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Hyperping API Key",
			"The provider cannot create the Hyperping API client as there is a missing or empty value for the Hyperping API key. "+
				"Set the api_key value in the configuration or use the HYPERPING_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	// Mask sensitive fields in logs to prevent API key leaks in debug output
	// Note: Context masking applies to logs in this Configure function only
	_ = tflog.MaskAllFieldValuesRegexes(
		tflog.MaskFieldValuesWithFieldKeys(ctx, "api_key"),
		client.APIKeyPattern,
	)

	// Create client with tflog integration for debugging
	hyperpingClient := client.NewClient(
		apiKey,
		client.WithBaseURL(baseURL),
		client.WithLogger(NewTFLogAdapter()),
		client.WithVersion(p.version),
	)

	// Make the client available to data sources and resources
	resp.DataSourceData = hyperpingClient
	resp.ResourceData = hyperpingClient
}

// Resources defines the resources implemented in the provider.
func (p *HyperpingProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMonitorResource,
		NewIncidentResource,
		NewIncidentUpdateResource,
		NewMaintenanceResource,
		NewOutageResource,
		NewHealthcheckResource,
		NewStatusPageResource,
		NewStatusPageSubscriberResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *HyperpingProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMonitorDataSource,
		NewMonitorsDataSource,
		NewIncidentDataSource,
		NewIncidentsDataSource,
		NewMaintenanceWindowDataSource,
		NewMaintenanceWindowsDataSource,
		NewMonitorReportDataSource,
		NewMonitorReportsDataSource,
		NewOutageDataSource,
		NewOutagesDataSource,
		NewHealthcheckDataSource,
		NewHealthchecksDataSource,
		NewStatusPageDataSource,
		NewStatusPagesDataSource,
		NewStatusPageSubscribersDataSource,
	}
}

// isAllowedBaseURL validates that a base URL points to an allowed Hyperping domain
// and uses HTTPS (VULN-016: reject http:// to prevent credential leakage in cleartext).
// Only *.hyperping.io and localhost (for testing) are permitted to prevent
// credential theft via SSRF attacks.
func isAllowedBaseURL(baseURL string) bool {
	lower := strings.ToLower(baseURL)

	// Extract domain, handling both http:// and https:// prefixes
	domain := lower
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	// Remove path, query, and fragment
	if idx := strings.IndexAny(domain, "/:?#"); idx != -1 {
		domain = domain[:idx]
	}

	// Allow localhost for local testing (127.0.0.1, ::1, localhost)
	// Localhost is exempt from HTTPS requirement (httptest uses HTTP)
	if domain == "localhost" || domain == "127.0.0.1" || domain == "[::1]" {
		return true
	}
	if strings.HasPrefix(domain, "127.0.0.1:") || strings.HasPrefix(domain, "localhost:") {
		return true
	}

	// Non-localhost targets MUST use HTTPS to prevent credential leakage (VULN-016)
	if !strings.HasPrefix(lower, "https://") {
		return false
	}

	// Allow official Hyperping domains (*.hyperping.io)
	if domain == "hyperping.io" || strings.HasSuffix(domain, ".hyperping.io") {
		return true
	}

	return false
}

// New creates a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HyperpingProvider{
			version: version,
		}
	}
}
