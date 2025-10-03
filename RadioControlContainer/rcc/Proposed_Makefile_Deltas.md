# Proposed Makefile Deltas (Non-Binding)

**Date**: 2025-01-15  
**Purpose**: CI/Makefile gate improvements based on audit findings  
**Status**: PROPOSED - No direct edits made

---

## 1. Add Integration Coverage Gate

```diff
# Add after line 83 in Makefile
+integration-coverage-gate: ## Check integration coverage threshold
+	@echo "$(BLUE)Checking integration coverage (≥70%)...$(NC)"
+	@INTEGRATION_COVERAGE=$$(go tool cover -func=coverage/integration.out | grep total | awk '{print $$3}' | sed 's/%//'); \
+	if [ -z "$$INTEGRATION_COVERAGE" ]; then \
+		echo "$(RED)❌ Could not determine integration coverage$(NC)"; \
+		exit 1; \
+	fi; \
+	echo "Integration coverage: $$INTEGRATION_COVERAGE%"; \
+	if [ $$(echo "$$INTEGRATION_COVERAGE < 70" | bc -l) -eq 1 ]; then \
+		echo "$(RED)❌ Integration coverage $$INTEGRATION_COVERAGE% is below threshold 70%$(NC)"; \
+		exit 1; \
+	fi; \
+	echo "$(GREEN)✅ Integration coverage $$INTEGRATION_COVERAGE% meets threshold$(NC)"

# Update check-coverage target
check-coverage: ## Check coverage thresholds
	@echo "$(BLUE)Checking coverage thresholds...$(NC)"
	@$(MAKE) check-overall-coverage
	@$(MAKE) check-critical-packages-coverage
+	@$(MAKE) integration-coverage-gate
	@echo "$(GREEN)✅ Coverage thresholds met$(NC)"
```

## 2. Add E2E Contract Coverage Gate

```diff
# Add after integration-coverage-gate
+e2e-contract-gate: ## Check E2E contract coverage (100% endpoint coverage)
+	@echo "$(BLUE)Checking E2E contract coverage (100%)...$(NC)"
+	@E2E_COVERAGE=$$(go test ./test/e2e -v 2>&1 | grep "Coverage:" | awk '{print $$2}' | sed 's/%//' | head -1); \
+	if [ -z "$$E2E_COVERAGE" ]; then \
+		echo "$(YELLOW)⚠️  E2E coverage not available, running manifest-aware test$(NC)"; \
+		@go test ./test/e2e -run TestManifestAware_E2EExecution -v 2>&1 | grep "Coverage:" | awk '{print $$2}' | sed 's/%//' > e2e_coverage.tmp; \
+		E2E_COVERAGE=$$(cat e2e_coverage.tmp); \
+		rm -f e2e_coverage.tmp; \
+	fi; \
+	if [ -z "$$E2E_COVERAGE" ]; then \
+		echo "$(RED)❌ Could not determine E2E coverage$(NC)"; \
+		exit 1; \
+	fi; \
+	echo "E2E coverage: $$E2E_COVERAGE%"; \
+	if [ $$(echo "$$E2E_COVERAGE < 100" | bc -l) -eq 1 ]; then \
+		echo "$(RED)❌ E2E coverage $$E2E_COVERAGE% is below threshold 100%$(NC)"; \
+		exit 1; \
+	fi; \
+	echo "$(GREEN)✅ E2E coverage $$E2E_COVERAGE% meets threshold$(NC)"

# Update check-coverage target
check-coverage: ## Check coverage thresholds
	@echo "$(BLUE)Checking coverage thresholds...$(NC)"
	@$(MAKE) check-overall-coverage
	@$(MAKE) check-critical-packages-coverage
	@$(MAKE) integration-coverage-gate
+	@$(MAKE) e2e-contract-gate
	@echo "$(GREEN)✅ Coverage thresholds met$(NC)"
```

## 3. Fix Lint Configuration

```diff
# Update lint target to handle configuration issues
lint: ## Run golangci-lint with strict configuration
	@echo "$(BLUE)Running golangci-lint...$(NC)"
-	@golangci-lint run --enable=errcheck,staticcheck,gocritic,govet,ineffassign,misspell,unused --timeout=5m ./internal/auth/... ./internal/command/... ./internal/config/... ./internal/adapter/... ./internal/audit/...
+	@golangci-lint run --config=.golangci.yml --timeout=5m ./internal/auth/... ./internal/command/... ./internal/config/... ./internal/adapter/... ./internal/audit/... || \
+	(golangci-lint run --enable=errcheck,staticcheck,gocritic,govet,ineffassign,misspell,unused --timeout=5m ./internal/auth/... ./internal/command/... ./internal/config/... ./internal/adapter/... ./internal/audit/...)
	@echo "$(GREEN)✅ Lint passed with 0 warnings$(NC)"
```

## 4. Add Anti-Peek Enforcement

```diff
# Add new target
+anti-peek: ## Enforce E2E tests don't access internal packages
+	@echo "$(BLUE)Checking anti-peek enforcement...$(NC)"
+	@VIOLATIONS=$$(go list -f '{{.ImportPath}}' ./test/e2e/... | xargs -I {} sh -c 'go list -f "{{.Imports}}" {} | grep -q "github.com/radio-control/rcc/internal" && echo {}'); \
+	if [ -n "$$VIOLATIONS" ]; then \
+		echo "$(RED)❌ E2E tests accessing internal packages:$$VIOLATIONS$(NC)"; \
+		exit 1; \
+	fi; \
+	echo "$(GREEN)✅ Anti-peek enforcement passed$(NC)"

# Update test-all target
test-all: unit e2e race bench cover lint ## Run all tests and quality checks
+	@$(MAKE) anti-peek
	@echo "$(GREEN)✅ All quality gates passed$(NC)"
```

## 5. Add Performance Gate

```diff
# Add new target
+perf-gate: ## Run performance tests and validate thresholds
+	@echo "$(BLUE)Running performance tests...$(NC)"
+	@if command -v vegeta >/dev/null 2>&1; then \
+		bash test/perf/vegeta_scenarios.sh; \
+		P95_LATENCY=$$(grep "95th percentile" perf_results.txt | awk '{print $$3}' | sed 's/ms//'); \
+		ERROR_RATE=$$(grep "error rate" perf_results.txt | awk '{print $$3}' | sed 's/%//'); \
+		echo "P95 Latency: $$P95_LATENCY ms"; \
+		echo "Error Rate: $$ERROR_RATE%"; \
+		if [ $$(echo "$$P95_LATENCY > 100" | bc -l) -eq 1 ]; then \
+			echo "$(YELLOW)⚠️  P95 latency $$P95_LATENCY ms exceeds 100ms threshold$(NC)"; \
+		fi; \
+		if [ $$(echo "$$ERROR_RATE > 10" | bc -l) -eq 1 ]; then \
+			echo "$(YELLOW)⚠️  Error rate $$ERROR_RATE% exceeds 10% threshold$(NC)"; \
+		fi; \
+		echo "$(GREEN)✅ Performance gate completed$(NC)"; \
+	else \
+		echo "$(YELLOW)⚠️  Vegeta not installed, skipping performance tests$(NC)"; \
+		echo "$(YELLOW)Install with: go install github.com/tsenart/vegeta@latest$(NC)"; \
+	fi

# Update test-all target (optional, as perf-gate is soft)
test-all: unit e2e race bench cover lint ## Run all tests and quality checks
	@$(MAKE) anti-peek
+	@$(MAKE) perf-gate
	@echo "$(GREEN)✅ All quality gates passed$(NC)"
```

## 6. Improve Test Execution Order

```diff
# Update test-all target to run in optimal order
test-all: unit integration e2e race bench cover lint ## Run all tests and quality checks
-	@echo "$(GREEN)✅ All quality gates passed$(NC)"
+	@$(MAKE) anti-peek
+	@$(MAKE) perf-gate
+	@echo "$(GREEN)✅ All quality gates passed$(NC)"

# Add parallel execution for faster feedback
+test-fast: unit integration ## Run fast tests only
+	@echo "$(GREEN)✅ Fast tests completed$(NC)"

+test-slow: e2e race bench cover lint ## Run slow tests only
+	@$(MAKE) anti-peek
+	@$(MAKE) perf-gate
+	@echo "$(GREEN)✅ Slow tests completed$(NC)"
```

## 7. Add Coverage Reporting Improvements

```diff
# Enhance coverage-report target
coverage-report: cover ## Generate detailed coverage report
	@echo "$(BLUE)Generating detailed coverage report...$(NC)"
	@echo "$(YELLOW)Coverage Report$(NC)"
	@echo "=================="
	@go tool cover -func=$(COVERAGE_OUTPUT) | grep -E "(github.com/radio-control/rcc/internal/|total)"
+	@echo ""
+	@echo "$(BLUE)Integration Coverage:$(NC)"
+	@go tool cover -func=coverage/integration.out | grep total
+	@echo ""
+	@echo "$(BLUE)E2E Contract Coverage:$(NC)"
+	@go test ./test/e2e -run TestManifestAware_E2EExecution -v 2>&1 | grep "Coverage:" || echo "E2E coverage not available"
	@echo ""
	@echo "$(BLUE)HTML coverage report: $(COVERAGE_HTML)$(NC)"
	@echo "$(BLUE)Raw coverage data: $(COVERAGE_OUTPUT)$(NC)"
```

## 8. Add Quality Gate Summary

```diff
# Add new target for quality gate summary
+quality-summary: ## Show quality gate status summary
+	@echo "$(BLUE)Quality Gate Status Summary$(NC)"
+	@echo "=============================="
+	@echo "$(YELLOW)Unit Tests:$(NC)"
+	@$(MAKE) unit >/dev/null 2>&1 && echo "$(GREEN)✅ PASS$(NC)" || echo "$(RED)❌ FAIL$(NC)"
+	@echo "$(YELLOW)Integration Tests:$(NC)"
+	@$(MAKE) integration >/dev/null 2>&1 && echo "$(GREEN)✅ PASS$(NC)" || echo "$(RED)❌ FAIL$(NC)"
+	@echo "$(YELLOW)E2E Tests:$(NC)"
+	@$(MAKE) e2e >/dev/null 2>&1 && echo "$(GREEN)✅ PASS$(NC)" || echo "$(RED)❌ FAIL$(NC)"
+	@echo "$(YELLOW)Race Detection:$(NC)"
+	@$(MAKE) race >/dev/null 2>&1 && echo "$(GREEN)✅ PASS$(NC)" || echo "$(RED)❌ FAIL$(NC)"
+	@echo "$(YELLOW)Linting:$(NC)"
+	@$(MAKE) lint >/dev/null 2>&1 && echo "$(GREEN)✅ PASS$(NC)" || echo "$(RED)❌ FAIL$(NC)"
+	@echo "$(YELLOW)Coverage:$(NC)"
+	@$(MAKE) cover >/dev/null 2>&1 && echo "$(GREEN)✅ PASS$(NC)" || echo "$(RED)❌ FAIL$(NC)"
+	@echo "$(YELLOW)Anti-Peek:$(NC)"
+	@$(MAKE) anti-peek >/dev/null 2>&1 && echo "$(GREEN)✅ PASS$(NC)" || echo "$(RED)❌ FAIL$(NC)"
+	@echo "$(YELLOW)Performance:$(NC)"
+	@$(MAKE) perf-gate >/dev/null 2>&1 && echo "$(GREEN)✅ PASS$(NC)" || echo "$(YELLOW)⚠️  SKIP$(NC)"
```

## 9. Add Help Documentation

```diff
# Update help target
help: ## Show this help message
	@echo "$(BLUE)Radio Control Container - Build and Quality Gates$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)Coverage Requirements:$(NC)"
	@echo "  Overall: ≥$(COVERAGE_THRESHOLD)%"
	@echo "  Critical packages (auth/command/telemetry): ≥$(COVERAGE_THRESHOLD_CRITICAL)%"
+	@echo "  Integration: ≥70%"
+	@echo "  E2E Contract: 100%"
	@echo ""
	@echo "$(YELLOW)Lint Configuration:$(NC)"
	@echo "  Linters: errcheck, staticcheck, gocritic, stylecheck"
	@echo "  Warnings = Errors"
+	@echo ""
+	@echo "$(YELLOW)Performance Requirements:$(NC)"
+	@echo "  P95 Latency: <100ms"
+	@echo "  Error Rate: <10%"
+	@echo ""
+	@echo "$(YELLOW)Quality Gates:$(NC)"
+	@echo "  test-fast: Unit + Integration tests"
+	@echo "  test-slow: E2E + Race + Bench + Coverage + Lint"
+	@echo "  test-all: All quality gates"
+	@echo "  quality-summary: Show gate status"
```

---

## Implementation Notes

### Priority Order
1. **Critical**: Fix lint configuration, add integration coverage gate
2. **High**: Add E2E contract coverage gate, anti-peek enforcement  
3. **Medium**: Add performance gate, improve test execution order
4. **Low**: Add quality summary, enhance help documentation

### Dependencies
- `bc` command for floating-point comparisons
- `vegeta` for performance testing (optional)
- Updated `.golangci.yml` configuration file

### Testing
- All new targets should be tested in isolation
- Performance gate should be soft-fail to not block CI
- Anti-peek enforcement should be strict-fail

### Rollback Plan
- Keep original targets unchanged
- New targets are additive
- Can disable new gates by commenting out in `test-all`

---

*These proposed deltas are based on audit findings and represent suggested improvements to the Makefile. They are non-binding and should be reviewed by the development team before implementation.*
