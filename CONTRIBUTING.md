# Contributing to terraform-provider-hyperping

Thank you for your interest in contributing! This document provides guidelines for contributing to this Terraform provider.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the behavior
- **Expected behavior** vs actual behavior
- **Terraform version** and provider version
- **Configuration files** (sanitized of sensitive data)
- **Debug logs** if applicable (`TF_LOG=DEBUG`)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Clear use case** and motivation
- **Proposed solution** (if you have one)
- **Alternatives considered**
- **Additional context** or examples

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Follow the coding standards** (see Development section)
3. **Write tests** for new functionality
4. **Update documentation** if needed
5. **Ensure tests pass** locally before submitting
6. **Write clear commit messages** following [Conventional Commits](https://www.conventionalcommits.org/)

#### Commit Message Format

```
<type>: <description>

[optional body]
[optional footer]
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`

**Examples:**
```
feat: add support for monitor groups
fix: resolve race condition in status page subscriber
docs: update README with new data source examples
test: add acceptance tests for healthcheck resource
```

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) 1.24 or later
- [Terraform](https://www.terraform.io/downloads.html) 1.0 or later
- [golangci-lint](https://golangci-lint.run/usage/install/) for linting
- [gosec](https://github.com/securego/gosec) for security scanning

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/terraform-provider-hyperping.git
cd terraform-provider-hyperping

# Install dependencies
go mod download

# Build
go build -v
```

### Running Tests

```bash
# Unit tests
go test -v ./...

# With coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out

# Acceptance tests (requires HYPERPING_API_KEY)
HYPERPING_API_KEY=your_key TF_ACC=1 go test -v ./internal/provider/

# Clean up test resources (sweepers)
HYPERPING_API_KEY=your_key go test -v -sweep=all -sweep-run=. ./internal/provider/

# Clean up specific resource type
HYPERPING_API_KEY=your_key go test -v -sweep=all -sweep-run=hyperping_monitor ./internal/provider/

# Linting
golangci-lint run

# Security scan
gosec ./...
```

#### Resource Sweepers

Resource sweepers automatically clean up orphaned test resources from your Hyperping account. Always run sweepers after acceptance testing to prevent resource accumulation.

**Important:** All acceptance test resources MUST be named with the prefix `tf-acc-test-` for sweepers to identify and delete them safely.

**Available sweepers:**
- `hyperping_monitor` - Cleans up test monitors
- `hyperping_incident` - Cleans up test incidents
- `hyperping_maintenance` - Cleans up test maintenance windows
- `hyperping_healthcheck` - Cleans up test healthchecks
- `hyperping_statuspage` - Cleans up test status pages
- `hyperping_outage` - Cleans up test outages (filtered by monitor name)

**Usage examples:**
```bash
# Run all sweepers (recommended after testing)
HYPERPING_API_KEY=your_key go test -v -sweep=all -sweep-run=. ./internal/provider/

# Run specific sweeper
HYPERPING_API_KEY=your_key go test -v -sweep=all -sweep-run=hyperping_monitor ./internal/provider/

# Dry run (see what would be deleted)
HYPERPING_API_KEY=your_key go test -v -sweep=all -sweep-run=hyperping_monitor -sweep-dry-run ./internal/provider/
```

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Keep functions under 50 lines where possible
- Write clear, descriptive variable names
- Add comments for complex logic
- Maintain test coverage above 50%

### Provider Development Guidelines

1. **Schema Design**
   - Follow Terraform naming conventions (snake_case)
   - Mark computed attributes as `Computed: true`
   - Use appropriate validators
   - Provide clear descriptions

2. **Resource Implementation**
   - Implement full CRUD operations
   - Add `ImportState` support
   - Handle API errors gracefully
   - Use context for cancellation

3. **Testing**
   - Write unit tests for utilities and helpers
   - Add acceptance tests for resources/data sources
   - Test error conditions
   - Verify state updates
   - **Test naming convention**: Prefix all test resources with `tf-acc-test-`
   - **Resource cleanup**: Use sweepers to clean up orphaned test resources

4. **Documentation**
   - Update resource/data source docs
   - Add examples in `examples/` directory
   - Document breaking changes in CHANGELOG

## Project Structure

```
├── internal/
│   ├── client/          # Hyperping API client
│   └── provider/        # Terraform provider implementation
├── examples/            # Usage examples
├── docs/               # Auto-generated documentation
├── .github/            # GitHub workflows and templates
└── tools/              # Development tools
```

## Release Process

Releases are automated via GitHub Actions when a version tag is pushed:

1. Maintainers create a new version tag (e.g., `v1.1.0`)
2. GitHub Actions builds multi-platform binaries
3. GoReleaser signs artifacts with GPG
4. Release is published to GitHub
5. Terraform Registry auto-detects the new version

## Questions?

- 💬 [Open a Discussion](https://github.com/develeap/terraform-provider-hyperping/discussions)
- 🐛 [Report a Bug](https://github.com/develeap/terraform-provider-hyperping/issues/new?template=bug_report.yml)
- ✨ [Request a Feature](https://github.com/develeap/terraform-provider-hyperping/issues/new?template=feature_request.yml)

---

**Thank you for contributing!** 🎉
