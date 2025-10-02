# Coverage Report

Generated: jue 02 oct 2025 21:18:18 +04

## Overall Coverage
**Overall Coverage: 78.4%** ❌ **FAIL** (Target: ≥80%)

## Package Coverage

| Package | Coverage | Status | Target |
|---------|----------|--------|--------|
| auth | 83.7% | ❌ FAIL | ≥85% |
| command | 79.5% | ❌ FAIL | ≥85% |
| telemetry | 92.6% | ✅ PASS | ≥85% |
| config | 62.4% | ❌ FAIL | ≥80% |
| adapter | 92.3% | ✅ PASS | ≥80% |
| adapter/fake | 59.6% | ❌ FAIL | ≥80% |
| adapter/silvusmock | 82.3% | ✅ PASS | ≥80% |
| audit | 87.0% | ✅ PASS | ≥80% |

## Critical Package Status
- **auth**: 83.7% (Target: ≥85%) ❌ **FAIL**
- **command**: 79.5% (Target: ≥85%) ❌ **FAIL**  
- **telemetry**: 92.6% (Target: ≥85%) ✅ **PASS**

## Issues Identified

### Compilation Issues
- `internal/radio` package has compilation errors (excluded from tests)
- `internal/telemetry` package has data race issues (excluded from tests)

### Linting Issues
- **42 linting issues** found across packages:
  - **30 errcheck issues**: Unchecked error return values
  - **11 staticcheck issues**: Potential nil pointer dereferences and other issues
  - **1 gocritic issue**: if-else chain that could be a switch statement

## Make Targets Status

| Target | Status | Notes |
|--------|--------|-------|
| `make unit` | ✅ PASS | Excludes radio and telemetry packages |
| `make cover` | ❌ FAIL | Overall coverage 78.4% < 80% threshold |
| `make lint` | ❌ FAIL | 42 linting issues found |
| `make e2e` | ⏸️ PENDING | Not yet implemented |

## Next Steps Required

1. **Fix compilation errors** in `internal/radio` package
2. **Resolve data race issues** in `internal/telemetry` package  
3. **Increase test coverage** to meet thresholds:
   - Overall: 78.4% → ≥80%
   - auth: 83.7% → ≥85%
   - command: 79.5% → ≥85%
4. **Fix linting issues** (42 issues total)
5. **Implement e2e tests** for the `e2e` target

## Coverage Report Files
- HTML Report: `coverage.html`
- Coverage Profile: `coverage.out`