# Load Testing Suite

This directory contains comprehensive load tests for the migration tools, validating performance with enterprise-scale datasets (100+ monitors).

## Quick Start

```bash
# Run all load tests
go test -tags=load ./test/load/... -v -timeout=30m

# Run specific platform tests
go test -tags=load ./test/load/... -v -run TestBetterStack
go test -tags=load ./test/load/... -v -run TestUptimeRobot
go test -tags=load ./test/load/... -v -run TestPingdom

# Run benchmarks
go test -tags=load ./test/load/... -bench=. -benchmem -run=^$

# Run with profiling
go test -tags=load ./test/load/... -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Test Categories

### Scale Tests
- **Small Scale**: 10 monitors (baseline)
- **Medium Scale**: 50 monitors
- **Large Scale**: 100 monitors
- **XL Scale**: 200 monitors (requires `-short=false`)

### Specialized Tests
- **Memory Leak Detection**: 10 iterations with GC checks
- **File Size Validation**: Ensures generated files < 10MB
- **Parallel Conversion**: Tests for race conditions
- **Rate Limiting**: Validates backoff behavior
- **Scaling Analysis**: Verifies linear scaling characteristics

## Test Structure

```
test/load/
├── README.md                      # This file
├── helpers.go                     # Shared test utilities
├── load_test.go                   # Comprehensive test suite
├── betterstack_load_test.go       # Better Stack specific tests
├── uptimerobot_load_test.go       # UptimeRobot specific tests
└── pingdom_load_test.go           # Pingdom specific tests
```

## Performance Expectations

Based on test system (Intel i7-12700H, 20 cores):

### Better Stack
- **100 monitors + 50 healthchecks**: ~3ms, ~2MB memory
- **Throughput**: ~50,000 resources/sec
- **E2E benchmark**: ~8.6ms per migration (100 monitors)

### UptimeRobot
- **100 monitors + 50 healthchecks**: ~1ms, ~1.2MB memory
- **Throughput**: ~150,000 resources/sec
- **E2E benchmark**: ~1.3ms per migration (100 monitors)

### Pingdom
- **100 checks**: ~5ms, ~2.1MB memory
- **Throughput**: ~20,000 checks/sec
- **E2E benchmark**: ~11ms per migration (100 checks)

## Build Tags

Tests use the `load` build tag to separate them from regular tests:

```go
//go:build load
```

This prevents load tests from running during normal development:
```bash
go test ./...           # Skips load tests
go test -tags=load ./... # Includes load tests
```

## CI/CD Integration

Add to GitHub Actions for weekly performance regression testing:

```yaml
name: Load Tests
on:
  schedule:
    - cron: '0 2 * * 1'  # Monday 2am
  workflow_dispatch:

jobs:
  load-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run Load Tests
        run: |
          go test -tags=load ./test/load/... -v -bench=. -benchmem > results.txt

      - name: Upload Results
        uses: actions/upload-artifact@v4
        with:
          name: load-test-results
          path: results.txt
```

## Memory Profiling

Analyze memory usage:

```bash
# Generate memory profile
go test -tags=load ./test/load/... -run TestBetterStackLoad_LargeScale -memprofile=mem.prof

# Analyze with pprof
go tool pprof mem.prof

# Commands in pprof:
# top10              - Top 10 memory allocators
# list functionName  - Show source code with allocations
# web                - Generate visual graph (requires graphviz)
```

## CPU Profiling

Identify performance bottlenecks:

```bash
# Generate CPU profile
go test -tags=load ./test/load/... -run TestBetterStackLoad_LargeScale -cpuprofile=cpu.prof

# Analyze with pprof
go tool pprof cpu.prof

# Commands in pprof:
# top10              - Top 10 CPU consumers
# list functionName  - Show source code with CPU time
# web                - Generate visual graph
```

## Race Detection

Test for concurrency issues:

```bash
go test -tags=load ./test/load/... -race -v
```

## Acceptance Criteria

All tests must meet these criteria:

- ✅ **Memory**: < 500MB for 100 monitors
- ✅ **Performance**: < 5 minutes for 100 monitors
- ✅ **File Size**: Generated files < 10MB
- ✅ **No Memory Leaks**: Stable across 10 iterations
- ✅ **Linear Scaling**: 2x data → ~2x time (±50%)
- ✅ **No Race Conditions**: Passes with `-race` flag

## Troubleshooting

### Test Failures

**Symptom**: `Memory usage exceeds 500 MB`
```bash
# Profile memory usage
go test -tags=load ./test/load/... -run TestMemoryLeak -memprofile=mem.prof
go tool pprof mem.prof
```

**Symptom**: `Execution time exceeds maximum`
```bash
# Profile CPU usage
go test -tags=load ./test/load/... -run TestLargeScale -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

**Symptom**: `Race condition detected`
```bash
# Run with race detector
go test -tags=load ./test/load/... -race -v
# Fix shared mutable state
```

### Performance Degradation

Compare benchmark results over time:

```bash
# Baseline
go test -tags=load ./test/load/... -bench=BenchmarkBetterStackE2E -benchmem > baseline.txt

# After changes
go test -tags=load ./test/load/... -bench=BenchmarkBetterStackE2E -benchmem > current.txt

# Compare
benchstat baseline.txt current.txt
```

## Contributing

When adding new migration tools:

1. Create `{platform}_load_test.go` following existing patterns
2. Add test data generators (e.g., `generate{Platform}Monitors`)
3. Implement all test categories (small/medium/large scale)
4. Add benchmarks (Conversion, Generation, E2E)
5. Update PERFORMANCE.md with results
6. Ensure all acceptance criteria pass

## See Also

- [PERFORMANCE.md](../../docs/PERFORMANCE.md) - Detailed performance documentation
- [Integration Tests](../integration/README.md) - API integration tests
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Development guidelines
