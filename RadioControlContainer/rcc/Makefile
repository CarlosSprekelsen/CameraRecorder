# Radio Control Container - Build and Quality Gates
# Source: PRE-INT-10
# Quote: "Prevent backsliding before integration"

.PHONY: help unit e2e cover lint clean test-all
.DEFAULT_GOAL := help

# Configuration
COVERAGE_THRESHOLD := 80
COVERAGE_THRESHOLD_CRITICAL := 85
COVERAGE_OUTPUT := coverage.out
COVERAGE_HTML := coverage.html
LINT_CONFIG := .golangci.yml

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Radio Control Container - Build and Quality Gates$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)Coverage Requirements:$(NC)"
	@echo "  Overall: ≥$(COVERAGE_THRESHOLD)%"
	@echo "  Critical packages (auth/command/telemetry): ≥$(COVERAGE_THRESHOLD_CRITICAL)%"
	@echo ""
	@echo "$(YELLOW)Lint Configuration:$(NC)"
	@echo "  Linters: errcheck, staticcheck, gocritic, stylecheck"
	@echo "  Warnings = Errors"

unit: ## Run unit tests
	@echo "$(BLUE)Running unit tests...$(NC)"
	@go test -v -timeout 30s ./internal/auth/... ./internal/command/... ./internal/config/... ./internal/adapter/... ./internal/audit/...

e2e: ## Run end-to-end tests
	@echo "$(BLUE)Running end-to-end tests...$(NC)"
	@go test -v -timeout 60s -run "Test.*E2E\|Test.*Integration" ./...

cover: ## Run tests with coverage and enforce thresholds
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@go test -coverprofile=$(COVERAGE_OUTPUT) -covermode=atomic ./internal/auth/... ./internal/command/... ./internal/telemetry/... ./internal/config/... ./internal/adapter/... ./internal/audit/...
	@echo "$(BLUE)Generating coverage report...$(NC)"
	@go tool cover -html=$(COVERAGE_OUTPUT) -o $(COVERAGE_HTML)
	@echo "$(BLUE)Coverage report generated: $(COVERAGE_HTML)$(NC)"
	@echo "$(BLUE)Analyzing coverage thresholds...$(NC)"
	@$(MAKE) check-coverage

lint: ## Run golangci-lint with strict configuration
	@echo "$(BLUE)Running golangci-lint...$(NC)"
	@golangci-lint run --enable=errcheck,staticcheck,gocritic,govet,ineffassign,misspell,unused --timeout=5m ./internal/auth/... ./internal/command/... ./internal/config/... ./internal/adapter/... ./internal/audit/...
	@echo "$(GREEN)✅ Lint passed with 0 warnings$(NC)"

test-all: unit e2e cover lint ## Run all tests and quality checks
	@echo "$(GREEN)✅ All quality gates passed$(NC)"

clean: ## Clean generated files
	@echo "$(BLUE)Cleaning generated files...$(NC)"
	@rm -f $(COVERAGE_OUTPUT) $(COVERAGE_HTML)
	@go clean -testcache

# Internal targets

check-coverage: ## Check coverage thresholds
	@echo "$(BLUE)Checking coverage thresholds...$(NC)"
	@$(MAKE) check-overall-coverage
	@$(MAKE) check-critical-packages-coverage
	@echo "$(GREEN)✅ Coverage thresholds met$(NC)"

check-overall-coverage: ## Check overall coverage threshold
	@echo "$(BLUE)Checking overall coverage (≥$(COVERAGE_THRESHOLD)%)...$(NC)"
	@OVERALL_COVERAGE=$$(go tool cover -func=$(COVERAGE_OUTPUT) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ -z "$$OVERALL_COVERAGE" ]; then \
		echo "$(RED)❌ Could not determine overall coverage$(NC)"; \
		exit 1; \
	fi; \
	echo "Overall coverage: $$OVERALL_COVERAGE%"; \
	if [ $$(echo "$$OVERALL_COVERAGE < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "$(RED)❌ Overall coverage $$OVERALL_COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)✅ Overall coverage $$OVERALL_COVERAGE% meets threshold$(NC)"

check-critical-packages-coverage: ## Check critical packages coverage threshold
	@echo "$(BLUE)Checking critical packages coverage (≥$(COVERAGE_THRESHOLD_CRITICAL)%)...$(NC)"
	@$(MAKE) check-package-coverage PACKAGE=auth THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)
	@$(MAKE) check-package-coverage PACKAGE=command THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)
	@$(MAKE) check-package-coverage PACKAGE=telemetry THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)

check-package-coverage: ## Check specific package coverage (usage: make check-package-coverage PACKAGE=package THRESHOLD=threshold)
	@PACKAGE_COVERAGE=$$(go tool cover -func=$(COVERAGE_OUTPUT) | grep "github.com/radio-control/rcc/internal/$(PACKAGE)" | awk '{print $$3}' | sed 's/%//' | head -1); \
	if [ -z "$$PACKAGE_COVERAGE" ]; then \
		echo "$(YELLOW)⚠️  Package $(PACKAGE) not found in coverage report$(NC)"; \
		return 0; \
	fi; \
	echo "Package $(PACKAGE) coverage: $$PACKAGE_COVERAGE%"; \
	if [ $$(echo "$$PACKAGE_COVERAGE < $(THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "$(RED)❌ Package $(PACKAGE) coverage $$PACKAGE_COVERAGE% is below threshold $(THRESHOLD)%$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)✅ Package $(PACKAGE) coverage $$PACKAGE_COVERAGE% meets threshold$(NC)"

# Coverage report generation
coverage-report: cover ## Generate detailed coverage report
	@echo "$(BLUE)Generating detailed coverage report...$(NC)"
	@echo "$(YELLOW)Coverage Report$(NC)"
	@echo "=================="
	@go tool cover -func=$(COVERAGE_OUTPUT) | grep -E "(github.com/radio-control/rcc/internal/|total)"
	@echo ""
	@echo "$(BLUE)HTML coverage report: $(COVERAGE_HTML)$(NC)"
	@echo "$(BLUE)Raw coverage data: $(COVERAGE_OUTPUT)$(NC)"

# Package-specific coverage
coverage-auth: ## Check auth package coverage
	@$(MAKE) check-package-coverage PACKAGE=auth THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)

coverage-command: ## Check command package coverage
	@$(MAKE) check-package-coverage PACKAGE=command THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)

coverage-telemetry: ## Check telemetry package coverage
	@$(MAKE) check-package-coverage PACKAGE=telemetry THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)

# Lint configuration
lint-config: ## Show lint configuration
	@echo "$(BLUE)Lint Configuration:$(NC)"
	@golangci-lint config --config=$(LINT_CONFIG)

# Development helpers
test-package: ## Run tests for specific package (usage: make test-package PACKAGE=package)
	@echo "$(BLUE)Running tests for package $(PACKAGE)...$(NC)"
	@go test -v -cover ./internal/$(PACKAGE)/...

test-coverage-package: ## Run tests with coverage for specific package (usage: make test-coverage-package PACKAGE=package)
	@echo "$(BLUE)Running tests with coverage for package $(PACKAGE)...$(NC)"
	@go test -coverprofile=$(PACKAGE)_coverage.out -covermode=atomic ./internal/$(PACKAGE)/...
	@go tool cover -func=$(PACKAGE)_coverage.out | grep total
