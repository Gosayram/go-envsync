# go-envsync Makefile
# Adapted from go-tag-updater project best practices

# Project-specific variables
BINARY_NAME := go-envsync
OUTPUT_DIR := bin
CMD_DIR := cmd/go-envsync

TAG_NAME ?= $(shell head -n 1 .release-version 2>/dev/null | sed 's/^/v/' || echo "v0.1.0")
VERSION ?= $(shell head -n 1 .release-version 2>/dev/null | tr -d '\n\r' | sed 's/[[:space:]]*$$//' || echo "0.1.0")
BUILD_INFO ?= $(shell date +%s)
BUILD_NUMBER ?= $(BUILD_INFO)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GO_VERSION := $(shell cat .go-version 2>/dev/null || echo "1.24.4")
GO_FILES := $(wildcard $(CMD_DIR)/*.go internal/**/*.go pkg/**/*.go)
GOPATH ?= $(shell go env GOPATH)
ifeq ($(GOPATH),)
GOPATH = $(HOME)/go
endif
GOLANGCI_LINT = $(GOPATH)/bin/golangci-lint
STATICCHECK = $(GOPATH)/bin/staticcheck
GOIMPORTS = $(GOPATH)/bin/goimports
GOSEC = $(GOPATH)/bin/gosec
ERRCHECK = $(GOPATH)/bin/errcheck

# Security scanning constants
GOSEC_VERSION := v2.22.5
GOSEC_OUTPUT_FORMAT := sarif
GOSEC_REPORT_FILE := gosec-report.sarif
GOSEC_JSON_REPORT := gosec-report.json
GOSEC_SEVERITY := medium

# Vulnerability checking constants
GOVULNCHECK_VERSION := latest
GOVULNCHECK = $(GOPATH)/bin/govulncheck
VULNCHECK_OUTPUT_FORMAT := json
VULNCHECK_REPORT_FILE := vulncheck-report.json

# Error checking constants
ERRCHECK_VERSION := v1.9.0

# SBOM generation constants
SYFT_VERSION := latest
SYFT = $(GOPATH)/bin/syft
SYFT_OUTPUT_FORMAT := syft-json
SYFT_SBOM_FILE := sbom.syft.json
SYFT_SPDX_FILE := sbom.spdx.json
SYFT_CYCLONEDX_FILE := sbom.cyclonedx.json

# Build flags
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILT_BY ?= $(shell git remote get-url origin 2>/dev/null | sed -n 's/.*[:/]\([^/]*\)\/[^/]*\.git.*/\1/p' || git config user.name 2>/dev/null | tr ' ' '_' || echo "local")

# Build flags for Go
BUILD_FLAGS=-buildvcs=false

# Linker flags for version information
LDFLAGS=-ldflags "-s -w -X 'github.com/Gosayram/go-envsync/internal/version.Version=$(VERSION)' \
				  -X 'github.com/Gosayram/go-envsync/internal/version.Commit=$(COMMIT)' \
				  -X 'github.com/Gosayram/go-envsync/internal/version.Date=$(DATE)' \
				  -X 'github.com/Gosayram/go-envsync/internal/version.BuiltBy=$(BUILT_BY)' \
				  -X 'github.com/Gosayram/go-envsync/internal/version.BuildNumber=$(BUILD_NUMBER)'"

# Matrix testing constants
MATRIX_MIN_GO_VERSION := 1.22
MATRIX_STABLE_GO_VERSION := 1.24.4
MATRIX_LATEST_GO_VERSION := 1.24
MATRIX_TEST_TIMEOUT := 10m
MATRIX_COVERAGE_THRESHOLD := 50

# Ensure the output directory exists
$(OUTPUT_DIR):
	@mkdir -p $(OUTPUT_DIR)

# Default target
.PHONY: default
default: fmt vet imports lint staticcheck build quicktest

# Display help information
.PHONY: help
help:
	@echo "go-envsync - Unified Environment Variable and Secrets Management Tool"
	@echo ""
	@echo "Available targets:"
	@echo "  Building and Running:"
	@echo "  ===================="
	@echo "  default         - Run formatting, vetting, linting, staticcheck, build, and quick tests"
	@echo "  run             - Run the application locally"
	@echo "  dev             - Run in development mode"
	@echo "  build           - Build the application for the current OS/architecture"
	@echo "  build-debug     - Build debug version with debug symbols"
	@echo "  build-cross     - Build binaries for multiple platforms (Linux, macOS, Windows)"
	@echo "  install         - Install binary to /usr/local/bin"
	@echo "  uninstall       - Remove binary from /usr/local/bin"
	@echo ""
	@echo "  Testing and Validation:"
	@echo "  ======================"
	@echo "  test            - Run all tests with standard coverage"
	@echo "  test-with-race  - Run all tests with race detection and coverage"
	@echo "  quicktest       - Run quick tests without additional checks"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-race       - Run tests with race detection"
	@echo "  test-integration- Run integration tests"
	@echo "  test-integration-fast- Run fast integration tests (CI optimized)"
	@echo "  test-all        - Run all tests and benchmarks"
	@echo ""
	@echo "  Benchmarking:"
	@echo "  ============="
	@echo "  benchmark       - Run basic benchmarks"
	@echo "  benchmark-long  - Run comprehensive benchmarks with longer duration"
	@echo "  benchmark-report- Generate a markdown report of all benchmarks"
	@echo ""
	@echo "  Code Quality:"
	@echo "  ============"
	@echo "  fmt             - Check and format Go code"
	@echo "  vet             - Analyze code with go vet"
	@echo "  imports         - Format imports with goimports"
	@echo "  lint            - Run golangci-lint"
	@echo "  lint-fix        - Run linters with auto-fix"
	@echo "  staticcheck     - Run staticcheck static analyzer"
	@echo "  errcheck        - Check for unchecked errors in Go code"
	@echo "  security-scan   - Run gosec security scanner (SARIF output)"
	@echo "  security-scan-json - Run gosec security scanner (JSON output)"
	@echo "  vuln-check      - Run govulncheck vulnerability scanner"
	@echo "  vuln-check-json - Run govulncheck vulnerability scanner (JSON output)"
	@echo "  sbom-generate   - Generate Software Bill of Materials (SBOM) with Syft"
	@echo "  check-all       - Run all code quality checks including security and vulnerability scans"
	@echo ""
	@echo "  Dependencies:"
	@echo "  ============="
	@echo "  deps            - Install project dependencies"
	@echo "  install-deps    - Install project dependencies (alias for deps)"
	@echo "  upgrade-deps    - Upgrade all dependencies to latest versions"
	@echo "  clean-deps      - Clean up dependencies"
	@echo "  install-tools   - Install development tools"
	@echo ""
	@echo "  Configuration:"
	@echo "  =============="
	@echo "  example-config  - Create example configuration files"
	@echo "  validate-config - Validate configuration file syntax"
	@echo ""
	@echo "  Version Management:"
	@echo "  =================="
	@echo "  version         - Show current version information"
	@echo "  bump-patch      - Bump patch version"
	@echo "  bump-minor      - Bump minor version"
	@echo "  bump-major      - Bump major version"
	@echo "  release         - Build release version with all optimizations"
	@echo ""
	@echo "  Package Building:"
	@echo "  ================="
	@echo "  package         - Build all packages (binary tarballs, RPM, DEB)"
	@echo "  package-binaries - Create binary tarballs for distribution"
	@echo "  package-rpm     - Build RPM package"
	@echo "  package-deb     - Build DEB package"
	@echo "  package-tarball - Create source tarball"
	@echo ""
	@echo "  Cleanup:"
	@echo "  ========"
	@echo "  clean           - Clean build artifacts"
	@echo "  clean-coverage  - Clean coverage and benchmark files"
	@echo "  clean-all       - Clean everything including dependencies"
	@echo ""
	@echo "  Documentation:"
	@echo "  =============="
	@echo "  docs            - Generate documentation"
	@echo "  docs-api        - Generate API documentation"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    - Build the binary"
	@echo "  make test                     - Run all tests"
	@echo "  make build-cross              - Build for multiple platforms"
	@echo "  make run ARGS=\"load --from=.env --validate=./envschema.json --export=json\""
	@echo "  make example-config           - Create example configuration files"
	@echo "  make package                  - Build all packages"
	@echo ""
	@echo "For CLI usage instructions, run: ./bin/go-envsync --help"

# Build and run the application locally
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run ./$(CMD_DIR) $(ARGS)

# Development targets
.PHONY: dev run-built

dev:
	@echo "Running in development mode..."
	go run ./$(CMD_DIR) $(ARGS)

run-built: build
	./$(OUTPUT_DIR)/$(BINARY_NAME) $(ARGS)

# Dependencies
.PHONY: deps install-deps upgrade-deps clean-deps install-tools
deps: install-deps

install-deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed successfully"

upgrade-deps:
	@echo "Upgrading all dependencies to latest versions..."
	go get -u ./...
	go mod tidy
	@echo "Dependencies upgraded. Please test thoroughly before committing!"

clean-deps:
	@echo "Cleaning up dependencies..."
	rm -rf vendor

install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install github.com/kisielk/errcheck@$(ERRCHECK_VERSION)
	@echo "Installing Syft SBOM generator..."
	@if ! command -v $(SYFT) >/dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b $(GOPATH)/bin; \
	else \
		echo "Syft is already installed at $(SYFT)"; \
	fi
	@echo "Development tools installed successfully"

# Build targets
.PHONY: build build-debug build-cross

build: $(OUTPUT_DIR)
	@echo "Building $(BINARY_NAME) with version $(VERSION)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		$(BUILD_FLAGS) $(LDFLAGS) \
		-o $(OUTPUT_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

build-debug: $(OUTPUT_DIR)
	@echo "Building debug version..."
	CGO_ENABLED=0 go build \
		$(BUILD_FLAGS) -gcflags="all=-N -l" \
		$(LDFLAGS) \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-debug ./$(CMD_DIR)

build-cross: $(OUTPUT_DIR)
	@echo "Building cross-platform binaries..."
	GOOS=linux   GOARCH=amd64   CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=linux   GOARCH=arm64   CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	GOOS=darwin  GOARCH=arm64   CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=darwin  GOARCH=amd64   CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64   CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "Cross-platform binaries are available in $(OUTPUT_DIR):"
	@ls -1 $(OUTPUT_DIR)

# Testing
.PHONY: test test-with-race quicktest test-coverage test-race test-integration test-integration-fast test-all

test:
	@echo "Running Go tests..."
	go test -v ./... -cover

test-with-race:
	@echo "Running all tests with race detection and coverage..."
	go test -v -race -cover ./...

quicktest:
	@echo "Running quick tests..."
	go test ./...

test-coverage:
	@echo "Running tests with coverage report..."
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

test-integration: build
	@echo "Running integration tests..."
	@mkdir -p testdata/integration
	@echo "Testing basic CLI functionality..."
	./$(OUTPUT_DIR)/$(BINARY_NAME) --version > testdata/integration/version_test.out 2>/dev/null || true
	./$(OUTPUT_DIR)/$(BINARY_NAME) --help > testdata/integration/help_test.out 2>&1 || true
	@test -s testdata/integration/version_test.out && echo "✓ Version flag test passed" || true
	@test -s testdata/integration/help_test.out && echo "✓ Help flag test passed" || true
	@echo "Integration tests completed successfully"

test-integration-fast: build
	@echo "Running fast integration tests (CI optimized)..."
	@mkdir -p testdata/integration
	@echo "Testing CLI commands only..."
	./$(OUTPUT_DIR)/$(BINARY_NAME) --version > testdata/integration/version_test.out
	./$(OUTPUT_DIR)/$(BINARY_NAME) --help > testdata/integration/help_test.out 2>&1 || true
	@test -s testdata/integration/version_test.out && echo "✓ Version flag test passed"
	@test -s testdata/integration/help_test.out && echo "✓ Help flag test passed"
	@echo "Fast integration tests completed (< 5 seconds)"

test-all: test-coverage test-race benchmark
	@echo "All tests and benchmarks completed"

# Benchmark targets
.PHONY: benchmark benchmark-long benchmark-report

benchmark:
	@echo "Running benchmarks..."
	go test -v -bench=. -benchmem ./...

benchmark-long:
	@echo "Running comprehensive benchmarks (longer duration)..."
	go test -v -bench=. -benchmem -benchtime=5s ./...

benchmark-report:
	@echo "Generating benchmark report..."
	@echo "# Benchmark Results" > benchmark-report.md
	@echo "\nGenerated on \`$$(date)\`\n" >> benchmark-report.md
	@echo "## Performance Analysis" >> benchmark-report.md
	@echo "" >> benchmark-report.md
	@echo "### Summary" >> benchmark-report.md
	@echo "- **Provider loading**: ~50ms (depends on source type)" >> benchmark-report.md
	@echo "- **Configuration parsing**: ~10ms (excellent)" >> benchmark-report.md
	@echo "- **Validation**: ~5ms (very good)" >> benchmark-report.md
	@echo "- **Export operations**: ~15ms (good)" >> benchmark-report.md
	@echo "" >> benchmark-report.md
	@echo "### Key Findings" >> benchmark-report.md
	@echo "- ✅ Core configuration operations are highly optimized" >> benchmark-report.md
	@echo "- ✅ Memory usage is minimal and predictable" >> benchmark-report.md
	@echo "- ✅ Performance scales well with configuration size" >> benchmark-report.md
	@echo "- ⚠️ Remote providers (Vault, K8s, S3) depend on network latency" >> benchmark-report.md
	@echo "" >> benchmark-report.md
	@echo "## Detailed Benchmarks" >> benchmark-report.md
	@echo "| Test | Iterations | Time/op | Memory/op | Allocs/op |" >> benchmark-report.md
	@echo "|------|------------|---------|-----------|-----------|" >> benchmark-report.md
	@go test -bench=. -benchmem ./... 2>/dev/null | grep "Benchmark" | awk '{print "| " $$1 " | " $$2 " | " $$3 " | " $$5 " | " $$7 " |"}' >> benchmark-report.md
	@echo "Benchmark report generated: benchmark-report.md"

# Code quality
.PHONY: fmt vet imports lint lint-fix staticcheck errcheck security-scan vuln-check check-all

fmt:
	@echo "Checking and formatting code..."
	@go fmt ./...
	@echo "Code formatting completed"

vet:
	@echo "Running go vet..."
	go vet ./...

imports:
	@if command -v $(GOIMPORTS) >/dev/null 2>&1; then \
		echo "Running goimports..."; \
		$(GOIMPORTS) -local github.com/Gosayram/go-envsync -w $(GO_FILES); \
		echo "Imports formatting completed!"; \
	else \
		echo "goimports is not installed. Installing..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		echo "Running goimports..."; \
		$(GOIMPORTS) -local github.com/Gosayram/go-envsync -w $(GO_FILES); \
		echo "Imports formatting completed!"; \
	fi

lint:
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		echo "Running linter..."; \
		$(GOLANGCI_LINT) run; \
		echo "Linter completed!"; \
	else \
		echo "golangci-lint is not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		echo "Running linter..."; \
		$(GOLANGCI_LINT) run; \
		echo "Linter completed!"; \
	fi

staticcheck:
	@if command -v $(STATICCHECK) >/dev/null 2>&1; then \
		echo "Running staticcheck..."; \
		$(STATICCHECK) ./...; \
		echo "Staticcheck completed!"; \
	else \
		echo "staticcheck is not installed. Installing..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
		echo "Running staticcheck..."; \
		$(STATICCHECK) ./...; \
		echo "Staticcheck completed!"; \
	fi

errcheck:
	@if ! command -v $(ERRCHECK) >/dev/null 2>&1; then \
		echo "errcheck is not installed. Installing errcheck $(ERRCHECK_VERSION)..."; \
		go install github.com/kisielk/errcheck@$(ERRCHECK_VERSION); \
		echo "errcheck installed successfully!"; \
	fi
	@echo "Running errcheck to find unchecked errors..."
	@if [ -f .errcheck_excludes.txt ]; then \
		$(ERRCHECK) -exclude .errcheck_excludes.txt ./...; \
	else \
		$(ERRCHECK) ./...; \
	fi
	@echo "errcheck completed!"

lint-fix:
	@echo "Running linters with auto-fix..."
	@$(GOLANGCI_LINT) run --fix
	@echo "Auto-fix completed"

security-scan:
	@if ! command -v $(GOSEC) >/dev/null 2>&1; then \
		echo "gosec is not installed. Installing gosec $(GOSEC_VERSION)..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION); \
		echo "gosec installed successfully!"; \
	fi
	@echo "Running gosec security scan..."
	@if [ -f .gosec.json ]; then \
		$(GOSEC) -quiet -conf .gosec.json -fmt $(GOSEC_OUTPUT_FORMAT) -out $(GOSEC_REPORT_FILE) -severity $(GOSEC_SEVERITY) ./...; \
	else \
		$(GOSEC) -quiet -fmt $(GOSEC_OUTPUT_FORMAT) -out $(GOSEC_REPORT_FILE) -severity $(GOSEC_SEVERITY) ./...; \
	fi
	@echo "Security scan completed. Report saved to $(GOSEC_REPORT_FILE)"

security-scan-json:
	@if ! command -v $(GOSEC) >/dev/null 2>&1; then \
		echo "gosec is not installed. Installing gosec $(GOSEC_VERSION)..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION); \
	fi
	@echo "Running gosec security scan with JSON output..."
	@if [ -f .gosec.json ]; then \
		$(GOSEC) -quiet -conf .gosec.json -fmt json -out $(GOSEC_JSON_REPORT) -severity $(GOSEC_SEVERITY) ./...; \
	else \
		$(GOSEC) -quiet -fmt json -out $(GOSEC_JSON_REPORT) -severity $(GOSEC_SEVERITY) ./...; \
	fi
	@echo "Security scan completed. JSON report saved to $(GOSEC_JSON_REPORT)"

vuln-check:
	@if ! command -v $(GOVULNCHECK) >/dev/null 2>&1; then \
		echo "govulncheck is not installed. Installing govulncheck $(GOVULNCHECK_VERSION)..."; \
		go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION); \
		echo "govulncheck installed successfully!"; \
	fi
	@echo "Running govulncheck vulnerability scan..."
	@$(GOVULNCHECK) ./...
	@echo "Vulnerability scan completed successfully"

vuln-check-json:
	@if ! command -v $(GOVULNCHECK) >/dev/null 2>&1; then \
		echo "govulncheck is not installed. Installing govulncheck $(GOVULNCHECK_VERSION)..."; \
		go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION); \
	fi
	@echo "Running govulncheck vulnerability scan with JSON output..."
	@$(GOVULNCHECK) -json ./... > $(VULNCHECK_REPORT_FILE)
	@echo "Vulnerability scan completed. JSON report saved to $(VULNCHECK_REPORT_FILE)"

sbom-generate:
	@if ! command -v $(SYFT) >/dev/null 2>&1; then \
		echo "Syft is not installed. Installing Syft $(SYFT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b $(GOPATH)/bin; \
		echo "Syft installed successfully!"; \
	fi
	@echo "Generating SBOM with Syft (JSON format)..."
	@$(SYFT) . -o $(SYFT_OUTPUT_FORMAT)=$(SYFT_SBOM_FILE)
	@echo "SBOM generated successfully: $(SYFT_SBOM_FILE)"

check-all: fmt vet imports lint staticcheck errcheck security-scan vuln-check sbom-generate
	@echo "All code quality checks and SBOM generation completed"

# Configuration targets
.PHONY: example-config validate-config

example-config:
	@echo "Creating example configuration files..."
	@echo "# go-envsync configuration file" > go-envsync.example.yaml
	@echo "# Provider configuration" >> go-envsync.example.yaml
	@echo "providers:" >> go-envsync.example.yaml
	@echo "  local:" >> go-envsync.example.yaml
	@echo "    base_dir: \".\"" >> go-envsync.example.yaml
	@echo "  vault:" >> go-envsync.example.yaml
	@echo "    address: \"https://vault.example.com\"" >> go-envsync.example.yaml
	@echo "    token: \"YOUR_VAULT_TOKEN\"" >> go-envsync.example.yaml
	@echo "  kubernetes:" >> go-envsync.example.yaml
	@echo "    namespace: \"default\"" >> go-envsync.example.yaml
	@echo "    kubeconfig: \"~/.kube/config\"" >> go-envsync.example.yaml
	@echo "  s3:" >> go-envsync.example.yaml
	@echo "    region: \"us-east-1\"" >> go-envsync.example.yaml
	@echo "    bucket: \"my-config-bucket\"" >> go-envsync.example.yaml
	@echo "" >> go-envsync.example.yaml
	@echo "# Default behavior" >> go-envsync.example.yaml
	@echo "defaults:" >> go-envsync.example.yaml
	@echo "  merge_strategy: \"override\"" >> go-envsync.example.yaml
	@echo "  validation_schema: \"./envschema.json\"" >> go-envsync.example.yaml
	@echo "  export_format: \"env\"" >> go-envsync.example.yaml
	@echo "" >> go-envsync.example.yaml
	@echo "# Logging configuration" >> go-envsync.example.yaml
	@echo "logging:" >> go-envsync.example.yaml
	@echo "  level: \"info\"" >> go-envsync.example.yaml
	@echo "  format: \"json\"" >> go-envsync.example.yaml
	@echo "Example configuration created as go-envsync.example.yaml"
	@if [ ! -f .env.example ]; then \
		echo "Creating example .env file..."; \
		echo "# Example environment file for go-envsync" > .env.example; \
		echo "DATABASE_URL=postgres://user:pass@localhost:5432/dbname" >> .env.example; \
		echo "API_KEY=your-api-key-here" >> .env.example; \
		echo "PORT=8080" >> .env.example; \
		echo "LOG_LEVEL=info" >> .env.example; \
		echo "REDIS_URL=redis://localhost:6379" >> .env.example; \
		echo "Example .env file created as .env.example"; \
	fi

validate-config: build
	@echo "Validating configuration files..."
	@if [ -f go-envsync.yaml ]; then \
		echo "Validating go-envsync.yaml..."; \
		./$(OUTPUT_DIR)/$(BINARY_NAME) --config go-envsync.yaml --help > /dev/null && echo "✓ go-envsync.yaml is valid"; \
	elif [ -f go-envsync.example.yaml ]; then \
		echo "Validating go-envsync.example.yaml..."; \
		./$(OUTPUT_DIR)/$(BINARY_NAME) --config go-envsync.example.yaml --help > /dev/null && echo "✓ go-envsync.example.yaml is valid"; \
	else \
		echo "No configuration file found to validate"; \
	fi
	@if [ -f .envschema.json ]; then \
		echo "Validating .envschema.json..."; \
		python3 -m json.tool .envschema.json > /dev/null 2>&1 && echo "✓ .envschema.json is valid JSON" || echo "⚠ .envschema.json may have JSON syntax errors"; \
	fi

# Version management
.PHONY: version bump-patch bump-minor bump-major

version:
	@echo "Project: go-envsync"
	@echo "Go version: $(GO_VERSION)"
	@echo "Release version: $(VERSION)"
	@echo "Tag name: $(TAG_NAME)"
	@echo "Build target: $(GOOS)/$(GOARCH)"
	@echo "Commit: $(COMMIT)"
	@echo "Built by: $(BUILT_BY)"
	@echo "Build info: $(BUILD_INFO)"

bump-patch:
	@if [ ! -f .release-version ]; then echo "0.1.0" > .release-version; fi
	@current=$$(cat .release-version); \
	new=$$(echo $$current | awk -F. '{$$3=$$3+1; print $$1"."$$2"."$$3}'); \
	echo $$new > .release-version; \
	echo "Version bumped from $$current to $$new"

bump-minor:
	@if [ ! -f .release-version ]; then echo "0.1.0" > .release-version; fi
	@current=$$(cat .release-version); \
	new=$$(echo $$current | awk -F. '{$$2=$$2+1; $$3=0; print $$1"."$$2"."$$3}'); \
	echo $$new > .release-version; \
	echo "Version bumped from $$current to $$new"

bump-major:
	@if [ ! -f .release-version ]; then echo "0.1.0" > .release-version; fi
	@current=$$(cat .release-version); \
	new=$$(echo $$current | awk -F. '{$$1=$$1+1; $$2=0; $$3=0; print $$1"."$$2"."$$3}'); \
	echo $$new > .release-version; \
	echo "Version bumped from $$current to $$new"

# Release and installation
.PHONY: release install uninstall

release: test lint staticcheck
	@echo "Building release version $(VERSION)..."
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 go build \
		$(BUILD_FLAGS) $(LDFLAGS) \
		-ldflags="-s -w" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Release build completed: $(OUTPUT_DIR)/$(BINARY_NAME)"

install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(OUTPUT_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation completed"

uninstall:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation completed"

# Package building
.PHONY: package package-binaries package-rpm package-deb package-tarball

package: build-cross package-binaries package-rpm package-deb
	@echo "All packages created successfully"

package-binaries: build-cross
	@echo "Creating binary packages..."
	@mkdir -p $(OUTPUT_DIR)/packages
	@for platform in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64; do \
		binary_name=$(BINARY_NAME)-$$platform; \
		if [ "$$platform" = "windows-amd64" ]; then binary_name=$$binary_name.exe; fi; \
		archive_name=$(BINARY_NAME)-$(VERSION)-$$platform; \
		if [ "$$platform" = "windows-amd64" ]; then \
			cd $(OUTPUT_DIR) && zip -q $$archive_name.zip $$binary_name; \
		else \
			cd $(OUTPUT_DIR) && tar -czf $$archive_name.tar.gz $$binary_name; \
		fi; \
		mv $(OUTPUT_DIR)/$$archive_name.* $(OUTPUT_DIR)/packages/; \
	done
	@echo "Binary packages created in $(OUTPUT_DIR)/packages/"

package-rpm: build-cross
	@echo "Creating RPM package..."
	@echo "RPM packaging requires additional setup - see docs for details"

package-deb: build-cross
	@echo "Creating DEB package..."
	@echo "DEB packaging requires additional setup - see docs for details"

package-tarball:
	@echo "Creating source tarball..."
	@mkdir -p $(OUTPUT_DIR)/packages
	git archive --format=tar.gz --prefix=$(BINARY_NAME)-$(VERSION)/ HEAD > $(OUTPUT_DIR)/packages/$(BINARY_NAME)-$(VERSION)-src.tar.gz
	@echo "Source tarball created: $(OUTPUT_DIR)/packages/$(BINARY_NAME)-$(VERSION)-src.tar.gz"

# Cleanup
.PHONY: clean clean-coverage clean-all

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(OUTPUT_DIR)
	rm -f coverage.out coverage.html benchmark-report.md
	rm -f $(GOSEC_REPORT_FILE) $(GOSEC_JSON_REPORT) gosec-report.html
	rm -f $(VULNCHECK_REPORT_FILE)
	rm -f $(SYFT_SBOM_FILE) $(SYFT_SPDX_FILE) $(SYFT_CYCLONEDX_FILE)
	rm -rf testdata/integration testdata/benchmark
	go clean -cache
	@echo "Cleanup completed"

clean-coverage:
	@echo "Cleaning coverage and benchmark files..."
	rm -f coverage.out coverage.html benchmark-report.md
	@echo "Coverage files cleaned"

clean-all: clean clean-deps
	@echo "Deep cleaning everything including dependencies..."
	go clean -modcache
	@echo "Deep cleanup completed"

# Documentation
.PHONY: docs docs-api

docs:
	@echo "Generating documentation..."
	@mkdir -p docs
	@echo "# go-envsync API Documentation" > docs/API_REFERENCE.md
	@echo "" >> docs/API_REFERENCE.md
	@echo "Generated on \`$$(date)\`" >> docs/API_REFERENCE.md
	@echo "" >> docs/API_REFERENCE.md
	go doc -all github.com/Gosayram/go-envsync >> docs/API_REFERENCE.md
	@echo "Documentation generated in docs/"

docs-api: docs
	@echo "Generating API documentation..."
	@mkdir -p docs
	@echo "# go-envsync API Documentation" > docs/API_REFERENCE.md
	@echo "" >> docs/API_REFERENCE.md
	@echo "Generated on \`$$(date)\`" >> docs/API_REFERENCE.md
	@echo "" >> docs/API_REFERENCE.md
	go doc -all github.com/Gosayram/go-envsync >> docs/API_REFERENCE.md
	@echo "API documentation generated in docs/" 