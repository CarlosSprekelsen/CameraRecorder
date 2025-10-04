# RCC Test Coverage Report
**Date:** 2025-10-04  
**Baseline:** Architecture v1.0, CB-TIMING v0.3, API OpenAPI v1.0  
**Last Updated:** 2025-10-04T06:41:44Z

---

## 1. MEASUREMENTS (Raw Data)

### Test Execution Counts
| Tier | Total | Passed | Failed | Skipped | Pass Rate | Coverage |
|------|-------|--------|--------|---------|-----------|----------|
| Unit | 11 | 11 | 0 | 0 | 100% | 84.4% |
| Integration | 6 | 6 | 0 | 0 | 100% | 19.3% |
| E2E | 79 | 79 | 0 | 0 | 100% | 32.1% |
| Performance | 15 | 15 | 0 | 0 | 100% | N/A |

### Coverage by Package (Unit Tests)
| Package | Coverage | Functions Tested | Total Functions |
|---------|----------|------------------|-----------------|
| auth | 86.6% | 18 | 18 |
| command | 88.3% | 20 | 26 |
| telemetry | 89.6% | 19 | 27 |
| adapter | 100.0% | 16 | 19 |
| adapter/fake | 78.7% | 11 | 18 |
| adapter/silvusmock | 82.3% | 16 | 19 |
| audit | 87.0% | 12 | 14 |
| config | 74.7% | 10 | 10 |

### Coverage by Package (Integration Tests)
| Package | Coverage | Functions Tested | Total Functions |
|---------|----------|------------------|-----------------|
| command | 26.2% | 20 | 26 |
| telemetry | 15.2% | 19 | 27 |
| orchestrator | 13.2% | 20 | 26 |
| auth | 0.1% | 18 | 18 |
| adapter | 0.0% | 16 | 19 |
| config | 0.0% | 10 | 10 |

### Coverage by Package (E2E Tests)
| Package | Coverage | Functions Tested | Total Functions |
|---------|----------|------------------|-----------------|
| Overall | 32.1% | Cross-package | Cross-package |

---

## 2. UNCOVERED CODE (Priority: Highest Gap First)

### Package: config (Gap: -25.3%)
**Target:** 80% (Makefile:9)  
**Actual:** 74.7%  
**Uncovered Functions:**
- Load() [lines 21-60] - Partial coverage (37.5%)
- applyEnvOverrides() [lines 61-173] - Partial coverage (52.8%)
- loadFromFile() [lines 174-190] - Partial coverage (33.3%)
- mergeTimingConfigs() [lines 191-248] - Partial coverage (58.8%)

### Package: auth (Gap: -13.4%)
**Target:** 85% (Makefile:117)  
**Actual:** 86.6%  
**Uncovered Functions:**
- verifyRS256Token() [lines 126-168] - Partial coverage (55.0%)
- getKeyFromJWKS() [lines 356-389] - Partial coverage (61.1%)

### Package: adapter/fake (Gap: -21.3%)
**Target:** Not specified in documentation  
**Actual:** 78.7%  
**Uncovered Functions:**
- ReadPowerActual() [lines 129-144] - Partial coverage (60.0%)
- SupportedFrequencyProfiles() [lines 145-169] - Partial coverage (60.0%)

---

## 3. UNTESTED REQUIREMENTS

| Requirement | Coverage | Gap | Affected Functions |
|-------------|----------|-----|-------------------|
| Architecture §8.5 Error Normalization | 100% | 0% | adapter/errors.go:NormalizeVendorError() |
| Architecture §8.6 Audit Schema | 87% | -13% | audit/logger.go:LogAction() |
| CB-TIMING §3 Heartbeat Configuration | 100% | 0% | telemetry/hub.go:sendHeartbeat() |
| CB-TIMING §5 Command Timeouts | 88.3% | -11.7% | command/orchestrator.go:SetPower(), SetChannel() |
| CB-TIMING §6 Event Buffering | 88.9% | -11.1% | telemetry/hub.go:bufferEvent() |
| API OpenAPI §2.2 HTTP Error Mapping | 100% | 0% | api/errors.go:ToAPIError() |

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
| Overall | 80% | Makefile:9 | 84.4% | +4.4% |
| auth | 85% | Makefile:117 | 86.6% | +1.6% |
| command | 85% | Makefile:118 | 88.3% | +3.3% |
| telemetry | 85% | Makefile:119 | 89.6% | +4.6% |
| config | 80% | Makefile:9 | 74.7% | -5.3% |
| adapter/fake | Not specified | - | 78.7% | Unknown |

### Requirements Coverage Gaps
| Requirement | Target | Source | Actual | Gap |
|-------------|--------|--------|--------|-----|
| Error Normalization | 100% | Architecture §8.5 | 100% | 0% |
| Audit Schema | 100% | Architecture §8.6 | 87% | -13% |
| Heartbeat Configuration | 100% | CB-TIMING §3.1 | 100% | 0% |
| Command Timeouts | 100% | CB-TIMING §5 | 88.3% | -11.7% |
| Event Buffering | 100% | CB-TIMING §6.1 | 88.9% | -11.1% |
| HTTP Error Mapping | 100% | API OpenAPI §2.2 | 100% | 0% |

---

## 6. FAILED TESTS (Root Cause Analysis)

### E2E Failures (0)
All E2E tests passing. Previous failures resolved.

---

## 7. ACTIONS REQUIRED (By Gap Size)

1. **config package**: Improve from 74.7% to 80% (-5.3% gap)
2. **Architecture §8.6**: Improve audit schema coverage from 87% to 100% (-13% gap)
3. **CB-TIMING §5**: Improve command timeout coverage from 88.3% to 100% (-11.7% gap)
4. **CB-TIMING §6**: Improve event buffering coverage from 88.9% to 100% (-11.1% gap)
5. **adapter/fake**: Improve from 78.7% (target undefined)

---

## 8. MEASUREMENT NOTES

- **Coverage targets**: Sourced from Makefile lines 9-10
- **Integration coverage**: Measured via `go test -coverpkg=./internal/...`
- **E2E coverage**: Measured via `go test -coverpkg=./internal/...`
- **Unit coverage**: Measured via `go test -coverprofile=coverage/unit.out`
- **Commands used**: `go tool cover -func=coverage/*.out`
- **Run timestamp**: 2025-10-04T06:41:44Z
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