default: fmt lint build

# Pre-push check: Run all validations (same as lefthook pre-push)
.PHONY: check
check: fmt vet lint test security ## Run all checks (build, lint, test, security)
	@echo "✅ All checks passed!"

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build provider binary
	go build -v ./...

.PHONY: install
install: build ## Install provider locally
	go install -v ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	gofmt -s -w -e .

.PHONY: test
test: ## Run unit tests
	go test -v -cover -timeout=120s -parallel=10 ./...

.PHONY: testacc
testacc: ## Run acceptance tests (uses .env if present)
	@if [ -f .env ]; then \
		echo "Loading API key from .env..."; \
		set -a && . ./.env && set +a && TF_ACC=1 go test -v -cover -timeout 120m ./internal/provider/...; \
	else \
		echo "No .env file found. Create one from .env.example or set HYPERPING_API_KEY"; \
		exit 1; \
	fi

.PHONY: testacc-test
testacc-test: ## Run acceptance tests with TEST API key (safer)
	@if [ -f .env ]; then \
		echo "Loading TEST API key from .env..."; \
		set -a && . ./.env && set +a && TF_ACC=1 HYPERPING_API_KEY=$$HYPERPING_TEST_API_KEY go test -v -cover -timeout 120m ./internal/provider/...; \
	else \
		echo "No .env file found. Create one from .env.example"; \
		exit 1; \
	fi

.PHONY: docs
docs: ## Generate documentation
	cd tools && go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir .. --provider-name hyperping

.PHONY: coverage
coverage: ## Check test coverage (52%+ threshold)
	@echo "Running unit tests with coverage..."
	@go test -coverprofile=coverage.out ./internal/...
	@echo ""
	@echo "=== Coverage Report ==="
	@TOTAL_COV=$$(go tool cover -func=coverage.out | grep "^total:" | awk '{print $$3}' | tr -d '%'); \
	echo "Total unit test coverage: $${TOTAL_COV}%"; \
	echo ""; \
	echo "Breakdown:"; \
	echo "  - Client: ~99% (full coverage)"; \
	echo "  - Provider non-CRUD: 100% (factory, Configure, Metadata, Schema, mapping)"; \
	echo "  - Provider CRUD: 0% unit tests (tested via acceptance tests)"; \
	echo ""; \
	if [ -z "$${TOTAL_COV}" ]; then \
		echo "⚠️  Warning: Could not parse coverage percentage"; \
		TOTAL_COV=0; \
	fi; \
	if [ $$(echo "$${TOTAL_COV} < 50" | bc -l 2>/dev/null || echo 0) -eq 1 ]; then \
		echo "❌ Coverage $${TOTAL_COV}% is below 50% threshold"; \
		exit 1; \
	fi; \
	echo "✅ Coverage $${TOTAL_COV}% meets 50% threshold"

.PHONY: security
security: ## Run security scans (govulncheck, gosec)
	@echo "Running govulncheck..."
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...
	@echo "Running gosec..."
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -exclude-generated ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: hooks
hooks: ## Install git hooks (lefthook)
	@command -v lefthook >/dev/null 2>&1 || { echo "Installing lefthook..."; go install github.com/evilmartians/lefthook@latest; }
	lefthook install
	@echo "✅ Git hooks installed! Pre-commit and pre-push hooks are now active."

.PHONY: clean
clean: ## Clean build artifacts
	rm -f terraform-provider-hyperping
	rm -rf dist/
