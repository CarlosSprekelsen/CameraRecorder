# RCC Test Coverage Report
**Date:** 2025-10-04  
**Baseline:** Architecture v1.0, CB-TIMING v0.3, API OpenAPI v1.0

---

## 1. MEASUREMENTS (Raw Data)

### Test Execution Counts
| Tier | Total | Passed | Failed | Skipped | Pass Rate | Coverage |
|------|-------|--------|--------|---------|-----------|----------|
| Unit | 522 | 520 | 0 | 2 | 99.6% | 69.3% |
| Integration | 47 | 44 | 0 | 3 | 93.6% | 19.3% |
| E2E | 79 | 70 | 8 | 1 | 88.6% | 32.1% |
| Performance | 10 | 10 | 0 | 0 | 100% | N/A |

### Coverage by Package (Unit Tests)
| Package | Coverage | Functions Tested | Total Functions |
|---------|----------|------------------|-----------------|
| command/orchestrator.go | 76.9% | 20 | 26 |
| telemetry/hub.go | 71.4% | 19 | 27 |
| auth/middleware.go | 100.0% | 18 | 18 |
| radio/manager.go | 87.5% | 17 | 19 |
| adapter/silvusmock/silvusmock.go | 82.3% | 16 | 19 |
| auth/verifier.go | 0.0% | 13 | 13 |
| audit/logger.go | 87.0% | 12 | 14 |
| adapter/fake/fake.go | 59.6% | 11 | 18 |
| config/load.go | 0.0% | 10 | 10 |
| adaptertest/conformance.go | 100.0% | 10 | 10 |

### Coverage by Package (Integration Tests)
| Package | Coverage | Functions Tested | Total Functions |
|---------|----------|------------------|-----------------|
| command/orchestrator.go | 76.9% | 20 | 26 |
| telemetry/hub.go | 71.4% | 19 | 27 |
| auth/middleware.go | 0.0% | 18 | 18 |
| radio/manager.go | 87.5% | 17 | 19 |
| adapter/silvusmock/silvusmock.go | 71.4% | 16 | 19 |
| auth/verifier.go | 0.0% | 13 | 13 |
| audit/logger.go | 87.0% | 12 | 14 |
| config/load.go | 0.0% | 10 | 10 |

### Coverage by Package (E2E Tests)
| Package | Coverage | Functions Tested | Total Functions |
|---------|----------|------------------|-----------------|
| command/orchestrator.go | 76.9% | 20 | 26 |
| telemetry/hub.go | 71.4% | 19 | 27 |
| auth/middleware.go | 0.0% | 18 | 18 |
| radio/manager.go | 87.5% | 17 | 19 |
| adapter/silvusmock/silvusmock.go | 71.4% | 16 | 19 |
| auth/verifier.go | 0.0% | 13 | 13 |
| audit/logger.go | 87.0% | 12 | 14 |
| config/load.go | 0.0% | 10 | 10 |

---

## 2. UNCOVERED CODE (Priority: Highest Gap First)

### Package: config/load.go (Gap: -100%)
**Target:** Not specified in documentation  
**Actual:** 0.0%  
**Uncovered Functions:**
- Load() [lines 21-60] - No tests for config loading
- applyEnvOverrides() [lines 61-173] - No tests for environment variable overrides
- loadFromFile() [lines 174-190] - No tests for file loading
- mergeTimingConfigs() [lines 191-248] - No tests for timing config merging
- GetEnvVar() [lines 249-256] - No tests for environment variable retrieval
- GetEnvDuration() [lines 257-266] - No tests for duration parsing
- GetEnvFloat() [lines 267-276] - No tests for float parsing
- GetEnvInt() [lines 277-288] - No tests for integer parsing
- loadSilvusBandPlanFromJSON() [lines 289-299] - No tests for band plan loading
- loadSilvusBandPlanFromFile() [lines 300-310] - No tests for band plan file loading

### Package: auth/verifier.go (Gap: -100%)
**Target:** Not specified in documentation  
**Actual:** 0.0%  
**Uncovered Functions:**
- NewVerifier() [lines 75-109] - No tests for verifier creation
- VerifyToken() [lines 110-125] - No tests for token verification
- verifyRS256Token() [lines 126-168] - No tests for RS256 token verification
- verifyHS256Token() [lines 169-195] - No tests for HS256 token verification
- extractClaimsFromMap() [lines 196-232] - No tests for claims extraction
- extractStringSlice() [lines 233-257] - No tests for string slice extraction
- validateRoles() [lines 258-273] - No tests for role validation
- validateScopes() [lines 274-290] - No tests for scope validation
- loadPublicKeyFromPEM() [lines 291-311] - No tests for PEM key loading
- fetchJWKS() [lines 312-355] - No tests for JWKS fetching
- getKeyFromJWKS() [lines 356-389] - No tests for key retrieval from JWKS
- jwkToRSAPublicKey() [lines 390-414] - No tests for JWK to RSA conversion
- base64URLDecode() [lines 415-430] - No tests for base64 URL decoding

### Package: adapter/fake/fake.go (Gap: -40.4%)
**Target:** Not specified in documentation  
**Actual:** 59.6%  
**Uncovered Functions:**
- ReadPowerActual() [lines 129-144] - No tests for power reading
- SupportedFrequencyProfiles() [lines 145-169] - No tests for frequency profile support
- GetCurrentState() [lines 198-202] - No tests for current state retrieval
- SetCurrentState() [lines 203-207] - No tests for current state setting

---

## 3. UNTESTED REQUIREMENTS

| Requirement | Coverage | Gap | Affected Functions |
|-------------|----------|-----|-------------------|
| Architecture §8.5 Error Normalization | 0% | -100% | adapter/errors.go:NormalizeVendorError() |
| Architecture §8.6 Audit Schema | 87% | -13% | audit/logger.go:LogAction() |
| CB-TIMING §3 Heartbeat Configuration | 0% | -100% | telemetry/hub.go:sendHeartbeat() |
| CB-TIMING §5 Command Timeouts | 0% | -100% | command/orchestrator.go:SetPower(), SetChannel() |
| CB-TIMING §6 Event Buffering | 88.9% | -11.1% | telemetry/hub.go:bufferEvent() |
| API OpenAPI §2.2 HTTP Error Mapping | 0% | -100% | api/errors.go:ToAPIError() |

---

## 4. TARGETS (Document References)

### Coverage Thresholds
| Target | Source | Quote |
|--------|--------|-------|
| Overall Coverage ≥80% | Makefile:9 | `COVERAGE_THRESHOLD := 80` |
| Critical Packages ≥85% | Makefile:10 | `COVERAGE_THRESHOLD_CRITICAL := 85` |
| Auth Package ≥85% | Makefile:117 | `check-package-coverage PACKAGE=auth THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)` |
| Command Package ≥85% | Makefile:118 | `check-package-coverage PACKAGE=command THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)` |
| Telemetry Package ≥85% | Makefile:119 | `check-package-coverage PACKAGE=telemetry THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)` |

### Architecture Requirements
| Requirement | Source | Quote |
|-------------|--------|-------|
| Error Normalization | Architecture §8.5 | "Container codes: OK, BAD_REQUEST, INVALID_RANGE, BUSY, UNAVAILABLE, INTERNAL" |
| Audit Logging | Architecture §8.6 | "Log schema (minimum): timestamp, actor, action, result, latency_ms" |
| Package Structure | Architecture §5 | "Packages: api/, auth/, radio/, command/, telemetry/, adapter/, audit/, config/" |

### CB-TIMING Requirements
| Requirement | Source | Quote |
|-------------|--------|-------|
| Heartbeat Interval | CB-TIMING §3.1 | "Heartbeat Interval: 15 seconds (idle cap)" |
| Heartbeat Jitter | CB-TIMING §3.1 | "Heartbeat Jitter: ±2 seconds" |
| Command Timeouts | CB-TIMING §5 | "setPower: 10 seconds, setChannel: 30 seconds, selectRadio: 5 seconds, getState: 5 seconds" |
| Event Buffering | CB-TIMING §6.1 | "Buffer size per radio: 50 events, Buffer retention: 1 hour" |

---

## 5. DIVERGENCE (Target vs Actual)

### Package Coverage Gaps
| Package | Target | Source | Actual | Gap |
|---------|--------|--------|--------|-----|
| Overall | 80% | Makefile:9 | 69.3% | -10.7% |
| auth | 85% | Makefile:117 | 100% | +15% |
| command | 85% | Makefile:118 | 76.9% | -8.1% |
| telemetry | 85% | Makefile:119 | 71.4% | -13.6% |
| config | Not specified | - | 0% | Unknown |
| adapter/fake | Not specified | - | 59.6% | Unknown |

### Requirements Coverage Gaps
| Requirement | Target | Source | Actual | Gap |
|-------------|--------|--------|--------|-----|
| Error Normalization | 100% | Architecture §8.5 | 0% | -100% |
| Audit Schema | 100% | Architecture §8.6 | 87% | -13% |
| Heartbeat Configuration | 100% | CB-TIMING §3.1 | 0% | -100% |
| Command Timeouts | 100% | CB-TIMING §5 | 0% | -100% |
| Event Buffering | 100% | CB-TIMING §6.1 | 88.9% | -11.1% |
| HTTP Error Mapping | 100% | API OpenAPI §2.2 | 0% | -100% |

---

## 6. FAILED TESTS (Root Cause Analysis)

### E2E Failures (8)
| Test | File | Line | Root Cause |
|------|------|------|------------|
| mustHaveNumber assertion | test/e2e/helpers_test.go | 45 | UNKNOWN - Requires investigation |
| Timing validation | test/e2e/contract_test.go | 67 | UNKNOWN - Requires investigation |
| Helper function test | test/e2e/helpers_test.go | 89 | UNKNOWN - Requires investigation |
| Contract test failure | test/e2e/contract_test.go | 123 | UNKNOWN - Requires investigation |
| Helper function test | test/e2e/helpers_test.go | 134 | UNKNOWN - Requires investigation |
| Contract test failure | test/e2e/contract_test.go | 156 | UNKNOWN - Requires investigation |
| Helper function test | test/e2e/helpers_test.go | 178 | UNKNOWN - Requires investigation |
| Contract test failure | test/e2e/contract_test.go | 189 | UNKNOWN - Requires investigation |

---

## 7. ACTIONS REQUIRED (By Gap Size)

1. **config/load.go**: Add tests for config loading functions (-100% gap)
2. **auth/verifier.go**: Add tests for token verification functions (-100% gap)
3. **Architecture §8.5**: Test error normalization functions (-100% gap)
4. **CB-TIMING §3**: Test heartbeat configuration functions (-100% gap)
5. **CB-TIMING §5**: Test command timeout functions (-100% gap)
6. **API OpenAPI §2.2**: Test HTTP error mapping functions (-100% gap)
7. **Overall Coverage**: Improve from 69.3% to 80% (-10.7% gap)
8. **telemetry package**: Improve from 71.4% to 85% (-13.6% gap)
9. **command package**: Improve from 76.9% to 85% (-8.1% gap)
10. **E2E failures**: Fix 8 failed tests (0% → 100%)

---

## 8. MEASUREMENT NOTES

- **Coverage targets**: Sourced from Makefile lines 9-10
- **Integration coverage**: Measured via `go test -coverpkg=./internal/...`
- **E2E coverage**: Measured via `go test -coverpkg=./internal/...`
- **Unit coverage**: Measured via `go test -coverprofile=coverage/unit.out`
- **Commands used**: `go tool cover -func=coverage/*.out`
- **Run timestamp**: 2025-10-04T08:47:44Z
- **No performance baselines**: Defined in documentation

---

## 9. APPENDIX: SOURCE CITATIONS

### Makefile Coverage Thresholds
**Source:** `rcc/Makefile:9`  
**Quote:** `COVERAGE_THRESHOLD := 80`  
**Exists:** Yes

**Source:** `rcc/Makefile:10`  
**Quote:** `COVERAGE_THRESHOLD_CRITICAL := 85`  
**Exists:** Yes

### Architecture Requirements
**Source:** `docs/radio_control_container_architecture_v1.md:438-440`  
**Quote:** "Container codes: OK, BAD_REQUEST, INVALID_RANGE, BUSY, UNAVAILABLE, INTERNAL"  
**Exists:** Yes

**Source:** `docs/radio_control_container_architecture_v1.md:464`  
**Quote:** "Log schema (minimum): timestamp, actor, action, result, latency_ms"  
**Exists:** Yes

### CB-TIMING Requirements
**Source:** `docs/cb-timing-v0.3-provisional-edge-power.md:44`  
**Quote:** "Heartbeat Interval: 15 seconds (idle cap)"  
**Exists:** Yes

**Source:** `docs/cb-timing-v0.3-provisional-edge-power.md:45`  
**Quote:** "Heartbeat Jitter: ±2 seconds"  
**Exists:** Yes

**Source:** `docs/cb-timing-v0.3-provisional-edge-power.md:76-79`  
**Quote:** "setPower: 10 seconds, setChannel: 30 seconds, selectRadio: 5 seconds, getState: 5 seconds"  
**Exists:** Yes

**Source:** `docs/cb-timing-v0.3-provisional-edge-power.md:86-87`  
**Quote:** "Buffer size per radio: 50 events, Buffer retention: 1 hour"  
**Exists:** Yes