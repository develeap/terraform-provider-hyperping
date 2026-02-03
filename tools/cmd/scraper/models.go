package main

import (
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
)

// CacheEntry stores metadata for change detection
type CacheEntry struct {
	Filename     string    `json:"filename"`
	URL          string    `json:"url"`
	ContentHash  string    `json:"content_hash"`
	Size         int       `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

// Cache stores all cache entries with creation timestamp
type Cache struct {
	Entries   map[string]CacheEntry `json:"entries"`
	CreatedAt time.Time             `json:"created_at"`
}

// ParameterChange represents a modification to an existing parameter
type ParameterChange struct {
	Name        string      `json:"name"`
	OldType     string      `json:"old_type"`
	NewType     string      `json:"new_type"`
	OldRequired bool        `json:"old_required"`
	NewRequired bool        `json:"new_required"`
	OldDefault  interface{} `json:"old_default,omitempty"`
	NewDefault  interface{} `json:"new_default,omitempty"`
}

// APIDiff represents changes detected between two versions of an API endpoint
type APIDiff struct {
	Section          string                    `json:"section"` // monitors, statuspages, incidents, etc.
	Endpoint         string                    `json:"endpoint"`
	Method           string                    `json:"method"`
	AddedParams      []extractor.APIParameter  `json:"added_params,omitempty"`
	RemovedParams    []extractor.APIParameter  `json:"removed_params,omitempty"`
	ModifiedParams   []ParameterChange         `json:"modified_params,omitempty"`
	Breaking         bool                      `json:"breaking"`
	RawContentChange bool                      `json:"raw_content_change"` // True if hash changed but no semantic diff found
}

// DiffReport aggregates all changes for a scraping run
type DiffReport struct {
	Timestamp      time.Time  `json:"timestamp"`
	TotalPages     int        `json:"total_pages"`
	ChangedPages   int        `json:"changed_pages"`
	UnchangedPages int        `json:"unchanged_pages"`
	APIDiffs       []APIDiff  `json:"api_diffs,omitempty"`
	Breaking       bool       `json:"breaking"` // True if any diff is breaking
	Summary        string     `json:"summary"`
}

// DiscoveredURL represents a URL found during navigation discovery
type DiscoveredURL struct {
	URL      string `json:"url"`
	Section  string `json:"section"`  // Parent section (monitors, statuspages, etc.)
	Method   string `json:"method"`   // HTTP method or operation (list, create, update, etc.)
	IsParent bool   `json:"is_parent"` // True if this is a parent page with children
}
