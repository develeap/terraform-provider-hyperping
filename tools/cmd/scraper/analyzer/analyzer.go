package analyzer

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/utils"
)

// CoverageGapType categorizes the type of discrepancy found
type CoverageGapType string

const (
	GapMissing         CoverageGapType = "missing"          // API has field, provider doesn't
	GapStale           CoverageGapType = "stale"            // Provider has field, API doesn't document
	GapTypeMismatch    CoverageGapType = "type_mismatch"    // Types differ
	GapRequiredChanged CoverageGapType = "required_changed" // Required/optional differs
)

// CoverageGap represents a discrepancy between API docs and provider
type CoverageGap struct {
	Type       CoverageGapType
	Resource   string
	APIField   string
	TFField    string
	APIType    string
	TFType     string
	Details    string
	Severity   string
	Suggestion string
	FilePath   string
	CodeHint   string
}

// ResourceCoverage represents coverage for a single resource
type ResourceCoverage struct {
	Resource          string
	TerraformResource string
	APIFields         int
	ImplementedFields int
	MissingFields     int
	StaleFields       int
	CoveragePercent   float64
}

// CoverageReport aggregates all coverage analysis
type CoverageReport struct {
	Timestamp       time.Time
	Resources       []ResourceCoverage
	TotalAPIFields  int
	CoveredFields   int
	MissingFields   int
	StaleFields     int
	CoveragePercent float64
	Gaps            []CoverageGap
}

// Analyzer performs API-to-provider schema comparison
type Analyzer struct {
	ProviderDir     string
	ProviderSchemas []ResourceSchema
	AllowedValues   map[string][]string
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(providerDir string) (*Analyzer, error) {
	a := &Analyzer{
		ProviderDir: providerDir,
	}

	// Extract provider schemas via AST parsing
	schemas, err := ExtractProviderSchemas(providerDir)
	if err != nil {
		return nil, fmt.Errorf("failed to extract provider schemas: %w", err)
	}
	a.ProviderSchemas = schemas

	// Extract allowed values from models_common.go
	allowedValues, err := ExtractAllowedValues(providerDir)
	if err != nil {
		// Non-fatal, continue without allowed values
		a.AllowedValues = make(map[string][]string)
	} else {
		a.AllowedValues = allowedValues
	}

	return a, nil
}

// AnalyzeCoverage compares API parameters against provider schemas
func (a *Analyzer) AnalyzeCoverage(apiParams map[string][]extractor.APIParameter) *CoverageReport {
	report := &CoverageReport{
		Timestamp: time.Now(),
	}

	// Group API params by resource section
	apiByResource := groupAPIParamsByResource(apiParams)

	// Analyze each resource
	for _, mapping := range ResourceMappings {
		resourceCoverage := a.analyzeResource(mapping, apiByResource[mapping.APISection])
		report.Resources = append(report.Resources, resourceCoverage)

		// Accumulate totals
		report.TotalAPIFields += resourceCoverage.APIFields
		report.CoveredFields += resourceCoverage.ImplementedFields
		report.MissingFields += resourceCoverage.MissingFields
		report.StaleFields += resourceCoverage.StaleFields
	}

	// Calculate overall coverage
	if report.TotalAPIFields > 0 {
		report.CoveragePercent = float64(report.CoveredFields) / float64(report.TotalAPIFields) * 100
	}

	// Collect all gaps
	report.Gaps = a.findAllGaps(apiByResource)

	return report
}

// analyzeResource compares API params for a single resource with provider schema
func (a *Analyzer) analyzeResource(mapping ResourceMapping, apiParams []extractor.APIParameter) ResourceCoverage {
	coverage := ResourceCoverage{
		Resource:          mapping.APISection,
		TerraformResource: mapping.TerraformResource,
	}

	// Find the provider schema for this resource
	var providerSchema *ResourceSchema
	for i := range a.ProviderSchemas {
		if a.ProviderSchemas[i].Name == mapping.TerraformResource {
			providerSchema = &a.ProviderSchemas[i]
			break
		}
	}

	// Build sets for comparison
	apiFieldSet := make(map[string]extractor.APIParameter)
	for _, param := range apiParams {
		if IsSkippedField(param.Name) {
			continue
		}
		normalizedName := CamelToSnake(param.Name)
		apiFieldSet[normalizedName] = param
	}

	tfFieldSet := make(map[string]SchemaField)
	if providerSchema != nil {
		for _, field := range providerSchema.Fields {
			if IsSkippedField(field.TerraformName) {
				continue
			}
			tfFieldSet[field.TerraformName] = field
		}
	}

	coverage.APIFields = len(apiFieldSet)

	// Count implemented fields (fields in both API and TF)
	for apiName := range apiFieldSet {
		tfName := MapAPIFieldToTerraform(apiName, &mapping)
		if _, exists := tfFieldSet[tfName]; exists {
			coverage.ImplementedFields++
		} else if _, exists := tfFieldSet[apiName]; exists {
			// Also check direct match
			coverage.ImplementedFields++
		} else if IsNestedField(apiName, &mapping) {
			// Field is implemented but nested inside another TF attribute (e.g., settings.website)
			coverage.ImplementedFields++
		} else {
			coverage.MissingFields++
		}
	}

	// Count stale fields (fields in TF but not in API)
	for tfName := range tfFieldSet {
		found := false
		for apiName := range apiFieldSet {
			mappedName := MapAPIFieldToTerraform(apiName, &mapping)
			if tfName == mappedName || tfName == apiName {
				found = true
				break
			}
		}
		if !found && !IsSkippedField(tfName) {
			// Check if it's an expected stale field (computed-only or known API doc gap)
			if IsExpectedStaleField(tfName, &mapping) {
				continue
			}
			// Also check AST-extracted schema for computed-only fields
			if field, ok := tfFieldSet[tfName]; ok && field.Computed && !field.Required && !field.Optional {
				continue
			}
			coverage.StaleFields++
		}
	}

	// Calculate coverage percentage
	if coverage.APIFields > 0 {
		coverage.CoveragePercent = float64(coverage.ImplementedFields) / float64(coverage.APIFields) * 100
	} else {
		coverage.CoveragePercent = 100 // No API fields = 100% coverage
	}

	return coverage
}

// findAllGaps collects all coverage gaps across resources
func (a *Analyzer) findAllGaps(apiByResource map[string][]extractor.APIParameter) []CoverageGap {
	var gaps []CoverageGap

	for _, mapping := range ResourceMappings {
		apiParams := apiByResource[mapping.APISection]
		resourceGaps := a.findResourceGaps(mapping, apiParams)
		gaps = append(gaps, resourceGaps...)
	}

	return gaps
}

// findResourceGaps finds gaps for a single resource
func (a *Analyzer) findResourceGaps(mapping ResourceMapping, apiParams []extractor.APIParameter) []CoverageGap {
	var gaps []CoverageGap

	// Find the provider schema for this resource
	var providerSchema *ResourceSchema
	for i := range a.ProviderSchemas {
		if a.ProviderSchemas[i].Name == mapping.TerraformResource {
			providerSchema = &a.ProviderSchemas[i]
			break
		}
	}

	// Build lookup maps
	tfFieldMap := make(map[string]SchemaField)
	if providerSchema != nil {
		for _, field := range providerSchema.Fields {
			tfFieldMap[field.TerraformName] = field
		}
	}

	// Check for missing fields (API has, provider doesn't)
	for _, apiParam := range apiParams {
		if IsSkippedField(apiParam.Name) {
			continue
		}

		tfName := MapAPIFieldToTerraform(apiParam.Name, &mapping)

		// Check if field is implemented via nested mapping (e.g., API "website" -> TF "settings.website")
		if IsNestedField(apiParam.Name, &mapping) {
			// Field is implemented but nested - not a gap
			continue
		}

		if _, exists := tfFieldMap[tfName]; !exists {
			// Also try direct match
			if _, exists := tfFieldMap[apiParam.Name]; !exists {
				gap := CoverageGap{
					Type:     GapMissing,
					Resource: mapping.APISection,
					APIField: apiParam.Name,
					TFField:  tfName,
					APIType:  apiParam.Type,
					Details:  fmt.Sprintf("API documents field '%s' but provider doesn't implement it", apiParam.Name),
					Severity: "warning",
				}

				// Generate suggestion
				gap.Suggestion = fmt.Sprintf("Add '%s' to %s schema", tfName, mapping.TerraformResource)
				if providerSchema != nil {
					gap.FilePath = providerSchema.FilePath
				}
				gap.CodeHint = generateCodeHint(tfName, apiParam)

				gaps = append(gaps, gap)
			}
		} else {
			// Field exists - check for type mismatch
			tfField := tfFieldMap[tfName]
			apiType := NormalizeTypeName(apiParam.Type)
			tfType := NormalizeTypeName(tfField.Type)

			// Only flag as mismatch if types are truly incompatible
			if !areTypesCompatible(apiType, tfType) {
				gap := CoverageGap{
					Type:       GapTypeMismatch,
					Resource:   mapping.APISection,
					APIField:   apiParam.Name,
					TFField:    tfName,
					APIType:    apiParam.Type,
					TFType:     tfField.Type,
					Details:    fmt.Sprintf("Type mismatch: API='%s', Terraform='%s'", apiParam.Type, tfField.Type),
					Severity:   "info",
					Suggestion: "Verify type mapping is intentional",
				}
				if providerSchema != nil {
					gap.FilePath = providerSchema.FilePath
				}
				gaps = append(gaps, gap)
			}
		}
	}

	return gaps
}

// groupAPIParamsByResource groups API parameters by their resource section
func groupAPIParamsByResource(apiParams map[string][]extractor.APIParameter) map[string][]extractor.APIParameter {
	result := make(map[string][]extractor.APIParameter)

	for endpoint, params := range apiParams {
		section := ExtractAPIEndpointSection(endpoint)
		if section == "" {
			continue
		}

		// Deduplicate params by name (same param may appear in create and update)
		existingNames := make(map[string]bool)
		for _, existing := range result[section] {
			existingNames[existing.Name] = true
		}

		for _, param := range params {
			if !existingNames[param.Name] {
				result[section] = append(result[section], param)
				existingNames[param.Name] = true
			}
		}
	}

	return result
}

// generateCodeHint generates a code snippet for adding a missing field
func generateCodeHint(tfName string, apiParam extractor.APIParameter) string {
	tfType := mapAPITypeToTerraformAttribute(apiParam.Type)
	required := "Optional"
	if apiParam.Required {
		required = "Required"
	}

	hint := fmt.Sprintf(`"%s": schema.%sAttribute{
    %s: true,
    MarkdownDescription: "%s",
}`, tfName, tfType, required, escapeDescription(apiParam.Description))

	return hint
}

// mapAPITypeToTerraformAttribute maps API types to Terraform schema attribute types
func mapAPITypeToTerraformAttribute(apiType string) string {
	normalized := NormalizeTypeName(apiType)
	switch normalized {
	case "string", "enum":
		return "String"
	case "number":
		return "Int64"
	case "boolean":
		return "Bool"
	case "array":
		return "List"
	case "object":
		return "SingleNested"
	default:
		return "String"
	}
}

// areTypesCompatible checks if two normalized types are compatible
// Some API types map to different TF types by design:
// - API "enum" -> TF "string" (with validators)
// - API "object" -> TF "string" (for localized strings like title/text)
func areTypesCompatible(apiType, tfType string) bool {
	// Same type is always compatible
	if apiType == tfType {
		return true
	}

	// Unknown types are compatible (can't verify)
	if apiType == "unknown" || tfType == "unknown" {
		return true
	}

	// enum -> string is the standard Terraform pattern (use validators for constraints)
	if apiType == "enum" && tfType == "string" {
		return true
	}

	// object -> string is used for localized strings (title, text with language keys)
	// The API docs may show these as objects with language keys, but TF accepts simple strings
	if apiType == "object" && tfType == "string" {
		return true
	}

	return false
}

// escapeDescription escapes special characters in descriptions
func escapeDescription(desc string) string {
	desc = strings.ReplaceAll(desc, `"`, `\"`)
	desc = strings.ReplaceAll(desc, "\n", " ")
	if len(desc) > 100 {
		desc = desc[:97] + "..."
	}
	return desc
}

// LoadAPIParamsFromSnapshot loads API parameters from a snapshot directory
func LoadAPIParamsFromSnapshot(snapshotDir string) (map[string][]extractor.APIParameter, error) {
	result := make(map[string][]extractor.APIParameter)

	files, err := filepath.Glob(filepath.Join(snapshotDir, "*.json"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		pageData, err := loadPageData(snapshotDir, file)
		if err != nil {
			continue
		}

		params := extractor.ExtractAPIParameters(pageData)
		filename := filepath.Base(file)
		result[filename] = params
	}

	return result, nil
}

// loadPageData reads a single page data JSON file
func loadPageData(baseDir, filePath string) (*extractor.PageData, error) {
	data, err := utils.SafeReadFile(baseDir, filePath)
	if err != nil {
		return nil, err
	}

	var pageData extractor.PageData
	if err := json.Unmarshal(data, &pageData); err != nil {
		return nil, err
	}

	return &pageData, nil
}
