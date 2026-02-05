// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/utils"
)

// EndpointMismatch represents a version mismatch between docs and provider
type EndpointMismatch struct {
	Resource        string
	DocsEndpoint    string
	ProviderPath    string
	DocsVersion     string
	ProviderVersion string
	FilePath        string
	LineNumber      int
	Suggestion      string
}

// EndpointVersionReport contains all endpoint version analysis results
type EndpointVersionReport struct {
	Mismatches []EndpointMismatch
	Matches    int
	Total      int
}

// endpointPattern matches API endpoint paths like /v1/monitors, /v2/statuspages
var endpointPattern = regexp.MustCompile(`/v(\d+)/([a-z-]+)`)

// ExtractEndpointsFromDocs extracts API endpoints from scraped documentation
func ExtractEndpointsFromDocs(snapshotDir string) (map[string]string, error) {
	endpoints := make(map[string]string) // resource -> endpoint path with version

	files, err := filepath.Glob(filepath.Join(snapshotDir, "*.json"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		content, err := utils.SafeReadFile(snapshotDir, file)
		if err != nil {
			continue
		}

		// Find all endpoint patterns in the file
		matches := endpointPattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) >= 3 {
				fullPath := match[0] // e.g., /v2/statuspages
				resource := match[2] // e.g., statuspages

				// Store the highest version found for each resource
				if existing, ok := endpoints[resource]; ok {
					existingVersion := extractVersion(existing)
					newVersion := extractVersion(fullPath)
					if newVersion > existingVersion {
						endpoints[resource] = fullPath
					}
				} else {
					endpoints[resource] = fullPath
				}
			}
		}
	}

	return endpoints, nil
}

// ExtractEndpointsFromProvider extracts API endpoints from provider client code
func ExtractEndpointsFromProvider(providerDir string) (map[string]EndpointInfo, error) {
	endpoints := make(map[string]EndpointInfo)

	clientDir := filepath.Join(providerDir, "client")
	files, err := filepath.Glob(filepath.Join(clientDir, "*.go"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// Skip test files
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		fileEndpoints, err := extractEndpointsFromFile(file)
		if err != nil {
			continue
		}

		for resource, info := range fileEndpoints {
			endpoints[resource] = info
		}
	}

	return endpoints, nil
}

// EndpointInfo contains information about an endpoint in provider code
type EndpointInfo struct {
	Path       string
	FilePath   string
	LineNumber int
	ConstName  string
}

// extractEndpointsFromFile parses a Go file and extracts endpoint constants
func extractEndpointsFromFile(filePath string) (map[string]EndpointInfo, error) {
	endpoints := make(map[string]EndpointInfo)

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Look for const declarations with "BasePath" or endpoint patterns
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.CONST {
				for _, spec := range x.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}

					for i, name := range valueSpec.Names {
						// Look for constants ending in BasePath
						if !strings.HasSuffix(name.Name, "BasePath") {
							continue
						}

						if i < len(valueSpec.Values) {
							if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
								path := strings.Trim(lit.Value, `"`)

								// Extract resource name from constant name
								// e.g., statuspagesBasePath -> statuspages
								resource := strings.TrimSuffix(name.Name, "BasePath")
								resource = strings.ToLower(resource)

								// Also try to extract from the path itself
								if matches := endpointPattern.FindStringSubmatch(path); len(matches) >= 3 {
									resource = matches[2]
								}

								pos := fset.Position(name.Pos())
								endpoints[resource] = EndpointInfo{
									Path:       path,
									FilePath:   filePath,
									LineNumber: pos.Line,
									ConstName:  name.Name,
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	return endpoints, nil
}

// CompareEndpointVersions compares endpoints from docs against provider code
func CompareEndpointVersions(docsEndpoints map[string]string, providerEndpoints map[string]EndpointInfo) *EndpointVersionReport {
	report := &EndpointVersionReport{}

	for resource, docsPath := range docsEndpoints {
		report.Total++

		providerInfo, exists := providerEndpoints[resource]
		if !exists {
			// Resource not implemented in provider - skip
			continue
		}

		docsVersion := extractVersion(docsPath)
		providerVersion := extractVersion(providerInfo.Path)

		if docsVersion != providerVersion {
			report.Mismatches = append(report.Mismatches, EndpointMismatch{
				Resource:        resource,
				DocsEndpoint:    docsPath,
				ProviderPath:    providerInfo.Path,
				DocsVersion:     fmt.Sprintf("v%d", docsVersion),
				ProviderVersion: fmt.Sprintf("v%d", providerVersion),
				FilePath:        providerInfo.FilePath,
				LineNumber:      providerInfo.LineNumber,
				Suggestion: fmt.Sprintf(
					"Update %s line %d: change %q to %q",
					filepath.Base(providerInfo.FilePath),
					providerInfo.LineNumber,
					providerInfo.Path,
					docsPath,
				),
			})
		} else {
			report.Matches++
		}
	}

	return report
}

// extractVersion extracts the version number from an endpoint path
func extractVersion(path string) int {
	matches := endpointPattern.FindStringSubmatch(path)
	if len(matches) >= 2 {
		var version int
		if _, err := fmt.Sscanf(matches[1], "%d", &version); err != nil {
			return 0
		}
		return version
	}
	return 0
}

// FormatEndpointReport generates a human-readable report
func FormatEndpointReport(report *EndpointVersionReport) string {
	var sb strings.Builder

	sb.WriteString("# Endpoint Version Analysis\n\n")

	if len(report.Mismatches) == 0 {
		sb.WriteString("✅ All endpoint versions match between docs and provider!\n")
		sb.WriteString(fmt.Sprintf("\nChecked %d endpoints, %d matches.\n", report.Total, report.Matches))
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("⚠️ Found %d endpoint version mismatches:\n\n", len(report.Mismatches)))

	for _, m := range report.Mismatches {
		sb.WriteString(fmt.Sprintf("## %s\n\n", m.Resource))
		sb.WriteString(fmt.Sprintf("| Source | Endpoint | Version |\n"))
		sb.WriteString(fmt.Sprintf("|--------|----------|--------|\n"))
		sb.WriteString(fmt.Sprintf("| **Docs** | `%s` | %s |\n", m.DocsEndpoint, m.DocsVersion))
		sb.WriteString(fmt.Sprintf("| **Provider** | `%s` | %s |\n", m.ProviderPath, m.ProviderVersion))
		sb.WriteString(fmt.Sprintf("\n**Fix:** %s\n\n", m.Suggestion))
	}

	sb.WriteString(fmt.Sprintf("---\n\nSummary: %d mismatches, %d matches out of %d endpoints\n",
		len(report.Mismatches), report.Matches, report.Total))

	return sb.String()
}
