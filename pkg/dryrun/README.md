# Dry-Run Package

Shared utilities for migration dry-run preview and compatibility analysis.

## Overview

The `dryrun` package provides reusable components for building comprehensive migration preview reports. It's used by all migration tools (Better Stack, UptimeRobot, Pingdom) to provide consistent dry-run functionality.

## Components

### Core Types (`types.go`)

Defines the data structures for dry-run reports:

- `Report` - Complete dry-run report
- `Summary` - Migration statistics
- `ResourceComparison` - Side-by-side comparison
- `CompatibilityScore` - Migration compatibility analysis
- `Warning` - Migration warnings and manual steps
- `PerformanceEstimates` - Time and resource estimates

### Compatibility Analyzer (`compatibility.go`)

Calculates migration compatibility scores and complexity:

```go
analyzer := dryrun.NewCompatibilityAnalyzer()
score := analyzer.AnalyzeCompatibility(comparisons, warnings)

fmt.Printf("Score: %.1f%%\n", score.OverallScore)
fmt.Printf("Complexity: %s\n", score.Complexity)
```

Features:
- Overall compatibility score (0-100%)
- Per-resource-type breakdown
- Complexity rating (Simple/Medium/Complex)
- Manual effort estimation

### Diff Formatter (`diff.go`)

Creates side-by-side comparisons:

```go
formatter := dryrun.NewDiffFormatter(useColors)
output := formatter.FormatComparison(comparison, maxWidth)
```

Features:
- ASCII table formatting
- Transformation indicators (✓, →, ~, ⚠)
- Text wrapping for long values
- Color support

### Preview Generator (`preview.go`)

Generates Terraform configuration previews:

```go
generator := dryrun.NewPreviewGenerator(useColors)
preview := generator.GeneratePreview(tfContent, resourceCount, previewLimit)
```

Features:
- Sample resource extraction
- Syntax highlighting (when colors enabled)
- Resource breakdown
- File size estimation

### Reporter (`reporter.go`)

Orchestrates all components to generate complete reports:

```go
reporter := dryrun.NewReporter(useColors)
report := reporter.GenerateReport(
    sourcePlatform,
    comparisons,
    warnings,
    tfContent,
    estimates,
    summary,
)

// Print to terminal
reporter.PrintReport(os.Stderr, report, opts)

// Or format as JSON
json, err := reporter.FormatJSON(report)
```

Features:
- Formatted terminal output
- JSON output for automation
- Configurable verbosity
- Section-based layout

### Bridge (`bridge.go`)

Adapters for connecting migration tool converters to dryrun package:

```go
// Implement this interface in your migration tool
type BridgeConverter interface {
    GetResourceName() string
    GetResourceType() string
    GetIssues() []string
    GetSourceData() map[string]interface{}
    GetTargetData() map[string]interface{}
}

// Build comparisons
comparisons := dryrun.BuildComparisons(converters)
```

## Usage Example

```go
import "github.com/develeap/terraform-provider-hyperping/pkg/dryrun"

// 1. Build resource comparisons
var bridges []dryrun.BridgeConverter
// ... populate with your converted resources

comparisons := dryrun.BuildComparisons(bridges)

// 2. Convert issues to warnings
warnings := convertIssuesToWarnings(conversionIssues)

// 3. Build summary
summary := dryrun.BuildSummary(
    monitorCount,
    healthcheckCount,
    tfContent,
    resourceBreakdown,
)

// 4. Estimate performance
estimates := dryrun.EstimatePerformance(
    resourceCount,
    sourceAPICalls,
    len(tfContent),
)

// 5. Generate report
reporter := dryrun.NewReporter(true)
report := reporter.GenerateReport(
    "Better Stack",
    comparisons,
    warnings,
    tfContent,
    estimates,
    summary,
)

// 6. Print report
opts := dryrun.Options{
    Verbose: verbose,
    PreviewLimit: 3,
}
reporter.PrintReport(os.Stderr, report, opts)
```

## Testing

```bash
go test ./pkg/dryrun/... -v
go test ./pkg/dryrun/... -cover
```

Test coverage: ~50% (focused on core logic)

## Integration

To integrate with a new migration tool:

1. **Implement BridgeConverter** for your source platform types
2. **Convert issues** to standardized warnings
3. **Build summary** with resource counts
4. **Generate report** using the reporter
5. **Print or serialize** the report

See `cmd/migrate-betterstack/dryrun.go` for a complete example.

## Design Principles

- **Reusable**: Shared across all migration tools
- **Flexible**: Supports different source platforms
- **Informative**: Provides actionable insights
- **Terminal-friendly**: Formatted for human readability
- **Automation-ready**: JSON output for CI/CD

## Future Enhancements

Potential improvements:

- HTML report generation
- Diff export to file
- Interactive mode (prompt for decisions)
- Historical comparison (track score over time)
- Custom warning rules
- Plugin system for platform-specific analysis
