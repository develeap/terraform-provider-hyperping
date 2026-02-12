# Filter Implementation Reference

Quick reference for the filtering framework implementation.

## Quick Stats

- **Data Sources with Filtering**: 6
- **Filter Types**: 7 (name_regex, protocol, paused, status, severity, monitor_uuid, hostname)
- **Integration Tests**: 16
- **Performance Tests**: 7
- **Benchmarks**: 4
- **Documentation Pages**: 1 (1500+ lines)

## Files Overview

### Core Implementation (Agent 1)
```
internal/provider/
├── filter_models.go          # Filter model structs
├── filter_helpers.go         # Reusable filter functions
├── filter_helpers_test.go    # Unit tests for helpers
├── filter_examples.go        # Usage examples and patterns
└── *_data_source.go          # Data sources with filtering
    ├── monitors_data_source.go
    ├── incidents_data_source.go
    ├── maintenance_windows_data_source.go
    ├── healthchecks_data_source.go
    ├── outages_data_source.go
    └── statuspages_data_source.go
```

### Testing & Documentation (Agent 2)
```
internal/provider/
├── filters_integration_test.go    # 16 integration tests
├── filters_performance_test.go    # 7 performance tests + 4 benchmarks
└── *_data_source_filter_test.go   # Individual data source tests

internal/client/
└── test_helpers.go                # Test data generators

docs/guides/
└── filtering-data-sources.md      # User guide (1500+ lines)
```

## Filter Types by Data Source

| Data Source | name_regex | protocol | paused | status | severity | monitor_uuid | hostname |
|-------------|------------|----------|--------|--------|----------|--------------|----------|
| monitors | ✓ | ✓ | ✓ | - | - | - | - |
| incidents | ✓ | - | - | ✓ | ✓ | - | - |
| maintenance_windows | ✓ | - | - | ✓ | - | - | - |
| healthchecks | ✓ | - | - | ✓ | - | - | - |
| outages | ✓ | - | - | - | - | ✓ | - |
| statuspages | ✓ | - | - | - | - | - | ✓ |

## Implementation Pattern

### 1. Add Filter to Schema
```go
func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "filter": MonitorFilterSchema(), // Use appropriate schema
            "resources": schema.ListNestedAttribute{
                // ... resource definition
            },
        },
    }
}
```

### 2. Add Filter to Model
```go
type DataSourceModel struct {
    Filter    *MonitorFilterModel `tfsdk:"filter"` // Use appropriate filter model
    Resources []ResourceModel     `tfsdk:"resources"`
}
```

### 3. Apply Filter in Read
```go
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var config DataSourceModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

    // Fetch all from API
    allResources, err := d.client.ListResources(ctx)
    if err != nil {
        resp.Diagnostics.AddError("Error", err.Error())
        return
    }

    // Apply filtering
    var filtered []Resource
    if config.Filter != nil {
        for _, resource := range allResources {
            if d.shouldIncludeResource(&resource, config.Filter, &resp.Diagnostics) {
                filtered = append(filtered, resource)
            }
        }
    } else {
        filtered = allResources
    }

    // Map to model and save state
    config.Resources = mapToModel(filtered)
    resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
```

### 4. Implement shouldInclude Method
```go
func (d *DataSource) shouldIncludeResource(resource *client.Resource, filter *FilterModel, diags *diag.Diagnostics) bool {
    return ApplyAllFilters(
        func() bool {
            match, err := MatchesNameRegex(resource.Name, filter.NameRegex)
            if err != nil {
                diags.AddError("Invalid regex", err.Error())
                return false
            }
            return match
        },
        func() bool {
            return MatchesExact(resource.Field, filter.Field)
        },
        func() bool {
            return MatchesBool(resource.BoolField, filter.BoolField)
        },
    )
}
```

## Available Filter Helpers

### String Matching
- `MatchesNameRegex(name string, pattern types.String) (bool, error)` - Regex matching
- `MatchesExact(value string, filter types.String) bool` - Case-sensitive exact match
- `MatchesExactCaseInsensitive(value string, filter types.String) bool` - Case-insensitive match
- `ContainsSubstring(value string, filter types.String) bool` - Substring match

### Boolean Matching
- `MatchesBool(value bool, filter types.Bool) bool` - Boolean comparison

### Numeric Matching
- `MatchesInt64(value int64, filter types.Int64) bool` - Exact int64 match
- `MatchesInt64Range(value int64, min, max types.Int64) bool` - Range check

### Slice Matching
- `MatchesStringSlice(values []string, filter types.String) bool` - Case-sensitive
- `MatchesStringSliceCaseInsensitive(values []string, filter types.String) bool` - Case-insensitive

### Combinator
- `ApplyAllFilters(filterFuncs ...func() bool) bool` - Combines multiple filters with AND logic

## Performance Benchmarks

### Results (100 resources)
- Exact Match: ~10.8µs per operation (115,514 ops/sec)
- Boolean Match: ~11.8µs per operation (85,010 ops/sec)
- Regex Match: ~388µs per operation (2,642 ops/sec)
- Combined Filters: ~295µs per operation (4,383 ops/sec)

### Optimization Tips
1. Use exact matching over regex when possible (35x faster)
2. Anchor regex patterns (`^...$`)
3. Avoid negative lookaheads
4. Combine filters: fast filters first (protocol, paused), then regex

## Common Regex Patterns

```hcl
# Environment prefix
name_regex = "^\\[(PROD|STAGING)\\]-.*"

# Service type suffix
name_regex = ".*-API$"

# Contains keyword
name_regex = ".*(DATABASE|DB).*"

# Standard format
name_regex = "^\\[ENV\\]-TYPE-NAME(-VERSION)?$"

# Case insensitive (using Go syntax)
name_regex = "(?i)keyword"
```

## Testing Checklist

When adding filtering to a new data source:

- [ ] Add filter schema to Schema()
- [ ] Add filter field to model struct
- [ ] Implement shouldInclude method
- [ ] Apply filtering in Read()
- [ ] Add unit tests for shouldInclude
- [ ] Add integration tests
- [ ] Update documentation
- [ ] Run performance benchmarks

## Troubleshooting

### Common Issues

1. **Regex compilation error**
   - Check for escaped backslashes (`\\` not `\`)
   - Test pattern with online regex tool

2. **No matches**
   - Verify regex is case-sensitive
   - Check for typos in filter values
   - Test with empty filter first

3. **Performance issues**
   - Use exact matching over regex
   - Simplify regex patterns
   - Add anchors to regex

4. **Null vs Unknown**
   - Null/unknown filters match all (no filtering)
   - To filter null values, use local processing

## Future Enhancements

### Potential Improvements
1. Server-side filtering (if API adds support)
2. Regex caching (pre-compile patterns)
3. Parallel filtering (goroutines for large datasets)
4. Index-based filtering
5. Filter statistics (how many matched)

### Backward Compatibility
- Adding new filters: ✅ Compatible (optional fields)
- Removing filters: ❌ Breaking change
- Changing filter behavior: ⚠️ Major version bump

## References

- [User Guide](./guides/filtering-data-sources.md) - Complete user documentation
- [Filter Models](../internal/provider/filter_models.go) - Filter structs
- [Filter Helpers](../internal/provider/filter_helpers.go) - Matching functions
- [Filter Examples](../internal/provider/filter_examples.go) - Implementation examples
- [Integration Tests](../internal/provider/filters_integration_test.go) - Test scenarios
- [Performance Tests](../internal/provider/filters_performance_test.go) - Benchmarks
