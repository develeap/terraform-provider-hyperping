// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

// =============================================================================
// Status Page Models
// =============================================================================

// StatusPage represents a Hyperping status page.
// API: ... /v2/statuspages/{uuid}
type StatusPage struct {
	UUID              string              `json:"uuid"` // e.g., "sp_abc123xyz"
	Name              string              `json:"name"`
	Hostname          *string             `json:"hostname"`        // custom domain
	HostedSubdomain   string              `json:"hostedsubdomain"` // e.g., "mycompany.hyperping.app"
	URL               string              `json:"url"`             // active URL
	PasswordProtected bool                `json:"password_protected"`
	Settings          StatusPageSettings  `json:"settings"`
	Sections          []StatusPageSection `json:"sections"`
}

// StatusPageSettings represents the settings for a status page.
type StatusPageSettings struct {
	Name                  string                           `json:"name"`
	Website               string                           `json:"website,omitempty"`
	Description           map[string]string                `json:"description,omitempty"` // language -> text
	Languages             []string                         `json:"languages"`
	DefaultLanguage       string                           `json:"default_language"`
	Theme                 string                           `json:"theme"` // light, dark, system
	Font                  string                           `json:"font"`
	AccentColor           string                           `json:"accent_color"` // hex color
	AutoRefresh           bool                             `json:"auto_refresh"`
	BannerHeader          bool                             `json:"banner_header"`
	Logo                  *string                          `json:"logo"`
	LogoHeight            string                           `json:"logo_height"`
	Favicon               *string                          `json:"favicon"`
	HidePoweredBy         bool                             `json:"hide_powered_by"`
	HideFromSearchEngines bool                             `json:"hide_from_search_engines"`
	GoogleAnalytics       *string                          `json:"google_analytics"`
	Subscribe             StatusPageSubscribeSettings      `json:"subscribe"`
	Authentication        StatusPageAuthenticationSettings `json:"authentication"`
}

// StatusPageSubscribeSettings represents subscription settings for a status page.
type StatusPageSubscribeSettings struct {
	Enabled bool `json:"enabled"`
	Email   bool `json:"email"`
	Slack   bool `json:"slack"`
	Teams   bool `json:"teams"`
	SMS     bool `json:"sms"`
}

// StatusPageAuthenticationSettings represents authentication settings for a status page.
type StatusPageAuthenticationSettings struct {
	PasswordProtection bool     `json:"password_protection"`
	GoogleSSO          bool     `json:"google_sso"`
	SAMLSSO            bool     `json:"saml_sso"`
	AllowedDomains     []string `json:"allowed_domains"`
}

// StatusPageSection represents a section on a status page.
type StatusPageSection struct {
	Name     map[string]string   `json:"name"` // language -> text
	IsSplit  bool                `json:"is_split"`
	Services []StatusPageService `json:"services"`
}

// StatusPageService represents a service (monitor or component) in a section.
type StatusPageService struct {
	ID                string              `json:"id"`
	UUID              string              `json:"uuid"`
	Name              map[string]string   `json:"name"` // language -> text
	IsGroup           bool                `json:"is_group"`
	ShowUptime        bool                `json:"show_uptime"`
	ShowResponseTimes bool                `json:"show_response_times"`
	Services          []StatusPageService `json:"services,omitempty"` // nested services if is_group
}

// StatusPagePaginatedResponse represents a paginated list of status pages.
// API: ... /v2/statuspages with pagination
type StatusPagePaginatedResponse struct {
	StatusPages    []StatusPage `json:"statuspages"`
	HasNextPage    bool         `json:"hasNextPage"`
	Total          int          `json:"total"`
	Page           int          `json:"page"`
	ResultsPerPage int          `json:"resultsPerPage"`
}

// CreateStatusPageRequest represents a request to create a status page.
// API: ... /v2/statuspages
type CreateStatusPageRequest struct {
	Name                  string                                  `json:"name"`
	Subdomain             *string                                 `json:"subdomain,omitempty"`
	Hostname              *string                                 `json:"hostname,omitempty"`
	Website               *string                                 `json:"website,omitempty"`
	Description           map[string]string                       `json:"description,omitempty"`
	Languages             []string                                `json:"languages,omitempty"`
	Theme                 *string                                 `json:"theme,omitempty"`
	Font                  *string                                 `json:"font,omitempty"`
	AccentColor           *string                                 `json:"accent_color,omitempty"`
	AutoRefresh           *bool                                   `json:"auto_refresh,omitempty"`
	BannerHeader          *bool                                   `json:"banner_header,omitempty"`
	Logo                  *string                                 `json:"logo,omitempty"`
	LogoHeight            *string                                 `json:"logo_height,omitempty"`
	Favicon               *string                                 `json:"favicon,omitempty"`
	HidePoweredBy         *bool                                   `json:"hide_powered_by,omitempty"`
	HideFromSearchEngines *bool                                   `json:"hide_from_search_engines,omitempty"`
	GoogleAnalytics       *string                                 `json:"google_analytics,omitempty"`
	Password              *string                                 `json:"password,omitempty"`
	Subscribe             *CreateStatusPageSubscribeSettings      `json:"subscribe,omitempty"`
	Authentication        *CreateStatusPageAuthenticationSettings `json:"authentication,omitempty"`
	Sections              []CreateStatusPageSection               `json:"sections,omitempty"`
}

// CreateStatusPageSubscribeSettings represents subscription settings in create requests.
type CreateStatusPageSubscribeSettings struct {
	Enabled *bool `json:"enabled,omitempty"`
	Email   *bool `json:"email,omitempty"`
	Slack   *bool `json:"slack,omitempty"`
	Teams   *bool `json:"teams,omitempty"`
	SMS     *bool `json:"sms,omitempty"`
}

// CreateStatusPageAuthenticationSettings represents authentication settings in create requests.
type CreateStatusPageAuthenticationSettings struct {
	PasswordProtection *bool    `json:"password_protection,omitempty"`
	GoogleSSO          *bool    `json:"google_sso,omitempty"`
	SAMLSSO            *bool    `json:"saml_sso,omitempty"`
	AllowedDomains     []string `json:"allowed_domains,omitempty"`
}

// CreateStatusPageSection represents a section in create requests.
type CreateStatusPageSection struct {
	Name     string                    `json:"name"`
	IsSplit  *bool                     `json:"is_split,omitempty"`
	Services []CreateStatusPageService `json:"services,omitempty"`
}

// CreateStatusPageService represents a service in create requests.
type CreateStatusPageService struct {
	MonitorUUID       string                    `json:"monitor_uuid"`
	NameShown         *string                   `json:"name_shown,omitempty"`
	ShowUptime        *bool                     `json:"show_uptime,omitempty"`
	ShowResponseTimes *bool                     `json:"show_response_times,omitempty"`
	IsGroup           *bool                     `json:"is_group,omitempty"`
	Services          []CreateStatusPageService `json:"services,omitempty"` // nested services
}

// Validate checks input lengths on CreateStatusPageRequest fields.
func (r CreateStatusPageRequest) Validate() error {
	if err := validateStringLength("name", r.Name, maxNameLength); err != nil {
		return err
	}
	if r.Website != nil {
		if err := validateStringLength("website", *r.Website, maxURLLength); err != nil {
			return err
		}
	}
	return nil
}

// UpdateStatusPageRequest represents a request to update a status page.
// API: ... /v2/statuspages/{uuid}
type UpdateStatusPageRequest struct {
	Name                  *string                                 `json:"name,omitempty"`
	Subdomain             *string                                 `json:"subdomain,omitempty"`
	Hostname              *string                                 `json:"hostname,omitempty"`
	Website               *string                                 `json:"website,omitempty"`
	Description           map[string]string                       `json:"description,omitempty"`
	Languages             []string                                `json:"languages,omitempty"`
	Theme                 *string                                 `json:"theme,omitempty"`
	Font                  *string                                 `json:"font,omitempty"`
	AccentColor           *string                                 `json:"accent_color,omitempty"`
	AutoRefresh           *bool                                   `json:"auto_refresh,omitempty"`
	BannerHeader          *bool                                   `json:"banner_header,omitempty"`
	Logo                  *string                                 `json:"logo,omitempty"`
	LogoHeight            *string                                 `json:"logo_height,omitempty"`
	Favicon               *string                                 `json:"favicon,omitempty"`
	HidePoweredBy         *bool                                   `json:"hide_powered_by,omitempty"`
	HideFromSearchEngines *bool                                   `json:"hide_from_search_engines,omitempty"`
	GoogleAnalytics       *string                                 `json:"google_analytics,omitempty"`
	Password              *string                                 `json:"password,omitempty"`
	Subscribe             *CreateStatusPageSubscribeSettings      `json:"subscribe,omitempty"`
	Authentication        *CreateStatusPageAuthenticationSettings `json:"authentication,omitempty"`
	Sections              []CreateStatusPageSection               `json:"sections,omitempty"`
}
