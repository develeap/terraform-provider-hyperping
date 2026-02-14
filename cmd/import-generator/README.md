# Import Generator v2.0

Enhanced bulk import tool for Hyperping Terraform resources with filtering, parallel execution, drift detection, and rollback capabilities.

## Quick Start

```bash
# Build
go build -o import-generator

# Generate import commands
export HYPERPING_API_KEY="sk_your_api_key"
./import-generator

# Execute parallel import with filtering
./import-generator --execute \
  --filter-name="PROD-.*" \
  --parallel=10 \
  --detect-drift
```

## Key Features

- **Filtering:** Import specific resource subsets by name/type
- **Parallel Execution:** 5-8x faster with concurrent imports
- **Drift Detection:** Pre/post-import terraform plan checks
- **Checkpoint/Resume:** Auto-save progress, resume after failures
- **Rollback:** Undo imports with one command
- **Progress Tracking:** Real-time progress bars

## Documentation

See [IMPORT_GENERATOR_GUIDE.md](../../docs/IMPORT_GENERATOR_GUIDE.md) for complete documentation.

## Examples

### Filter PROD resources
```bash
./import-generator --execute --filter-name="^PROD-.*"
```

### Parallel import with drift detection
```bash
./import-generator --execute --parallel=10 --detect-drift --abort-on-drift
```

### Resume after interruption
```bash
./import-generator --execute --resume
```

### Rollback failed import
```bash
./import-generator --rollback
```

## Performance

| Resources | Sequential | Parallel (10) | Speedup |
|-----------|-----------|---------------|---------|
| 50        | 2m 30s    | 30s           | 5x      |
| 100       | 5m        | 45s           | 6.7x    |
| 500       | 25m       | 3m            | 8.3x    |

## Testing

```bash
go test -v ./...
```

## Changelog

### v2.0 (2026-02-14)
- Added parallel import execution (5-8x faster)
- Added resource filtering (name, type, exclusion)
- Added drift detection (pre/post-import)
- Added checkpoint/resume capability
- Added rollback functionality
- Added progress tracking
- Added dry-run mode

### v1.0
- Initial release
- Sequential import generation
- Basic HCL generation

## License

MPL-2.0
