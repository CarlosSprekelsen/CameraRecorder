---------------------------------------------------------------------------
Prompt 5 [Complete]: src/camera_service/main.py

Context:
I am a solo engineer. Ground truth is in:
- docs/architecture/overview.md
- docs/development/principles.md
- docs/development/documentation-guidelines.md

Goal: Audit and improve `src/camera_service/main.py` so that:
- Startup sequence correctly wires configuration, logging, service manager, and handles dependency failures with clear errors.
- Shutdown (signal handling) is graceful on typical termination signals; all running components are torn down cleanly.
- Errors during initialization bubble with sufficient context and don’t leave partially-initialized state.
- TODO/STOP comments are canonical or explicitly deferred.

Additional Goal: Verify that a test scaffold exists under `tests/unit/test_camera_service/` (e.g., `test_main_startup.py`). If missing, emit a minimal pytest stub covering:
  * Successful startup and teardown,
  * Signal-triggered shutdown,
  * Initialization failure paths (e.g., config load failure) and their observable behavior.

Scope: Only modify `src/camera_service/main.py` and create the test stub if needed.

Instructions:
1. Audit startup/shutdown logic for robustness and clarity of failure modes.
2. Normalize any TODO/STOP comments.
3. Ensure signal handling works and is documented/observable.
4. Check or emit test stub with clear expected assertions.

Output:
- Summary of findings and changes.
- Updated `main.py`.
- Starter test file if missing.
- Evidence (line references).
- Suggested further tests.
- Any open questions.

---------------------------------------------------------------------------
Prompt 6 [Complete]: src/camera_service/config.py

Context:
I am a solo engineer. Ground truth is in:
- docs/architecture/overview.md
- docs/development/principles.md
- docs/development/documentation-guidelines.md

Goal: Audit and harden `src/camera_service/config.py` so that:
- Configuration loading handles missing, malformed, or partially invalid YAML gracefully with fallbacks.
- Environment variable overrides are validated; invalid overrides do not crash the service but log appropriate errors.
- Hot reload mechanism triggers updates safely and does not leave inconsistent state.
- Schema validation is comprehensive (types, ranges, required fields) and error accumulation is clear.
- TODO/STOP comments follow canonical format.

Additional Goal: Verify presence of test scaffold `tests/unit/test_camera_service/test_config_manager.py`. If missing, generate a stub covering:
  * Loading default config when none exists,
  * Env var overrides (valid and invalid),
  * Malformed config detection,
  * Hot reload simulation.

Scope: Only modify `src/camera_service/config.py` and create the test stub if needed.

Instructions:
1. Audit fallback behavior, override handling, and hot reload safety.
2. Normalize TODO/STOP comments.
3. Ensure errors are logged without crashing.
4. Check for or emit test stub with fixtures.

Output:
- Audit summary and fixes.
- Updated `config.py`.
- Test scaffold stub if missing.
- Evidence of behaviors/cases covered.
- Suggested additional tests.
- Any open clarifications.

---------------------------------------------------------------------------
Prompt 7 [Complete]: src/camera_service/logging_config.py

Context:
I am a solo engineer. Ground truth is in:
- docs/architecture/overview.md
- docs/development/principles.md
- docs/development/documentation-guidelines.md

Goal: Audit and complete `src/camera_service/logging_config.py` so that:
- Log rotation is either implemented per configuration (`max_file_size`, `backup_count`) or explicitly deferred with a canonical STOP-style decision note including rationale and plan.
- Correlation ID propagation is reliable and present in all relevant log emitters.
- Formatter selection (console vs structured/JSON) behaves per config and degrades gracefully on misconfiguration.
- TODO/STOP comments are canonical.

Additional Goal: Verify test scaffold `tests/unit/test_camera_service/test_logging_config.py` exists. If missing, emit a stub covering:
  * Formatter behavior,
  * Correlation ID presence,
  * Deferred rotation logic (if not implemented) vs active rotation.

Scope: Only modify `logging_config.py` and possibly create the test stub.

Instructions:
1. Audit current implementation for rotation, formatting, and correlation usage.
2. Implement rotation or add clear deferment note.
3. Normalize comments.
4. Check/create test stub.

Output:
- Summary of audit findings and corrections.
- Updated `logging_config.py`.
- Starter test stub if absent.
- Evidence (line numbers).
- Suggested test expansions.
- Any open questions.

---------------------------------------------------------------------------
Prompt 8: src/common/ (seed shared utilities)

Context:
I am a solo engineer. Ground truth is in:
- docs/architecture/overview.md
- docs/development/principles.md
- docs/development/documentation-guidelines.md

Goal: Populate `src/common/` with at least one practical shared utility and refactor an existing consumer to use it, to give the package immediate value and reduce duplication. Candidate modules:
  * `retry.py`: exponential backoff with jitter helper.
  * `types.py`: shared enums/constants (e.g., camera status).
  * `logging_helpers.py`: correlation ID injection/retrieval helper.

Additional Goal: Verify tests or usage exist that exercise the new common utility (e.g., health monitor, hybrid_monitor, or controller uses `common/retry.py`).

Scope: Create new module(s) under `src/common/` and modify one existing consumer to leverage it (small refactor). Do not add unrelated features.

Instructions:
1. Create at least one utility (e.g., backoff helper) with minimal interface and defaults.
2. Refactor a consumer (pick one: health monitor in controller, retry logic in hybrid_monitor, etc.) to use it.
3. Write or update a minimal docstring/instruction in `src/common/__init__.py` or the module.
4. If no existing test covers this yet, emit a stub test under `tests/unit/test_common/` for the utility.

Output:
- New `src/common/` module (e.g., `retry.py`) and refactored consumer.
- Test stub if needed.
- Summary of what was added/refactored.
- Evidence (file/line references).
- Any open clarifications.

---------------------------------------------------------------------------
Prompt 1 [Complete]: Expand udev testing and metadata reconciliation (S3)

Ground truth sources (authoritative): 
- docs/architecture/overview.md
- docs/development/principles.md
- docs/api/json-rpc-methods.md
- docs/development/documentation-guidelines.md
- Condensed roadmap/backlog (priority focused on S3 hardening).
Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.

Directive: ONLY use the above documents as the source of truth. Do not invent new features or make architectural assumptions beyond what is specified. If anything is ambiguous or missing for a required behavior, stop and ask one precise clarifying question before proceeding.

Goal:
1. Expand and tighten test coverage for `src/camera_discovery/hybrid_monitor.py` to cover:
   - udev add/remove/change events including race conditions and invalid device nodes.
   - Polling fallback when udev events are missed or stale.
   - Capability parsing variations (multiple frame rates, malformed output).
2. Validate reconciliation between the effective capability output from hybrid_monitor (frequency-weighted provisional/confirmed merge) and the metadata consumed by `src/camera_service/service_manager.py`. Ensure provisional vs confirmed semantics propagate without drift or silent mismatch.

Scope:
- Create/extend pytest files under `tests/unit/test_camera_discovery/`.
- Write a lightweight integration test or verification helper that feeds hybrid_monitor’s merged capability result to service_manager’s metadata path and asserts consistency.
- Do not add new architectural components or unrelated features.

Acceptance Criteria:
- Tests exist for udev event variants: add/remove/change, invalid node, racing sequences.  
- Tests simulate missing udev events and verify polling fallback triggers and recovers.  
- Capability parsing tests cover multiple fps formats and gracefully handle malformed outputs.  
- Reconciliation check/assertion shows the effective capability (after confirmation logic) used by service_manager matches expected provisional/confirmed state; any divergence is surfaced clearly.  
- No undefined behavior silently swallowed; mismatches are enumerated in the summary.

Test Requirements:
- Files like `test_hybrid_monitor_udev_fallback.py` and `test_hybrid_monitor_capability_parsing.py` with concrete fixture-driven scenarios.
- A reconciliation test (could be in same or separate file) that mocks or drives hybrid_monitor output into service_manager and verifies metadata alignment.

Output:
- Bullet summary of findings and enhancements (with file/line references where applicable).  
- New or updated test code.  
- Reconciliation validation code/results and any adjustments needed.  
- Suggested additional edge-case tests.  
- Any open clarifying question (only if absolutely required).

---------------------------------------------------------------------------
Prompt 2 [Complete] : Harden and validate service manager lifecycle and observability (S3)

Ground truth sources (authoritative): 
- docs/architecture/overview.md
- docs/development/principles.md
- docs/api/json-rpc-methods.md
- docs/development/documentation-guidelines.md
- Condensed roadmap/backlog emphasizing S3 lifecycle/observability.
Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.

Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.


Directive: Only rely on those documents. Do not add new features or speculative behavior. If a required decision (e.g., fallback priority or metadata merging ambiguity) is unclear, stop and ask exactly one focused question.

Goal:
- Audit and improve `src/camera_service/service_manager.py` so that:
  * Camera lifecycle (connect → MediaMTX stream actions → notification → disconnect) sequencing is deterministic and resistant to partial failures.
  * Capability metadata is annotated with provisional vs confirmed state and propagated in notifications.
  * Correlation IDs are consistently included in lifecycle logs and notifications for traceability.
  * Failures in dependent subsystems (MediaMTX/controller, capability retrieval) are handled defensively with clear fallback or error signaling; no silent inconsistency.
  * All TODO/STOP comments follow canonical format or are explicit deferred decisions with rationale and date.

Scope:
- Modify `service_manager.py`.
- Create or extend test stub under `tests/unit/test_camera_service/test_service_manager_lifecycle.py`.

Acceptance Criteria:
- Lifecycle flow is covered by at least one test that injects success and failure modes and asserts order and side effects.  
- Notifications include explicit flags/fields for provisional vs confirmed metadata.  
- Logs include correlation IDs at key transition points, including error paths.  
- Errors from MediaMTX or missing capabilities do not crash silently; fallback logic is observable.  
- All comment placeholders are canonicalized.

Test Requirements:
- `test_service_manager_lifecycle.py` simulating connect/disconnect with injected MediaMTX failure and verifying proper recovery or fallback.
- Assertions about metadata stability (e.g., provisional stays until confirmed).

Output:
- Summary of audit findings and applied changes with line references.  
- Updated `service_manager.py`.  
- Lifecycle test(s) stub or implementation.  
- Any remaining ambiguity documented (only one question if needed).


---------------------------------------------------------------------------
Prompt 3 [Complete]: Add MediaMTX edge-case health monitor tests (S4)
Ground truth sources (authoritative): 
- docs/architecture/overview.md
- docs/development/principles.md
- docs/api/json-rpc-methods.md
- docs/development/documentation-guidelines.md
- Backlog prioritizing closure of S4 edge-case behavior.
Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.

Directive: Base everything strictly on the existing implementation and documentation; do not extrapolate new recovery policies unless grounded. If the exact expected behavior under combinations (e.g., flapping) is unclear, ask one precise question.

Goal:
- Build tests for the health monitoring logic in `src/mediamtx_wrapper/controller.py` covering:
  * Circuit breaker opening after the configured failure threshold.  
  * Recovery requiring N consecutive successful checks before closing (confirmation logic).  
  * Stability under flapping (alternating success/failure) to avoid oscillation.  
  * Backoff/jitter behavior deterministically validated (or controlled for test reliability).

Scope:
- Add tests under `tests/unit/test_mediamtx_wrapper/` (new files if necessary).

Acceptance Criteria:
- Tests exist for:
  * Failure sequence triggering open state.  
  * Recovery only after the necessary consecutive successes.  
  * Flapping scenario does not prematurely reset or reopen unexpectedly.  
  * Controlled simulation of backoff behavior (e.g., mocking time or overriding jitter) to assert bounds.  
- Any deviation between implementation and expected circuit breaker state machine is documented.

Test Requirements:
- Files like `test_health_monitor_circuit_breaker_flapping.py` and `test_health_monitor_recovery_confirmation.py`.  
- Mocks to simulate health check results and control timing.

Output:
- Test implementations.  
- Summary of any behavioral gaps found and recommendations/fixes.  
- Suggested follow-up stress or integration tests.


---------------------------------------------------------------------------
Prompt 4: Document closure of resolved partials (S4)
Ground truth sources (authoritative): 
- docs/architecture/overview.md
- docs/development/principles.md
- docs/api/json-rpc-methods.md
- docs/development/documentation-guidelines.md

Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.

Directive: Only document what was actually resolved per the audit artifacts. No embellishment. If the linkage (evidence) for any purported closure is missing, note it explicitly.

Goal:
- Create and insert concise closure log entries for the following resolved partials:
  * Snapshot capture implementation completeness.  
  * Recording duration accuracy and error handling.  
  * Versioning/deprecation deferral (canonical STOP decision).  
  * Capability merging policy stabilization (weighted merge + confirmation).  
  * Health monitor recovery refinement (consecutive-success confirmation).  

Scope:
- Documentation update (e.g., extend `docs/architecture/overview.md` or a dedicated decision log file).

Acceptance Criteria:
- Each closure entry includes: original deficiency, change applied, date (YYYY-MM-DD), and direct evidence references (file name + line or test).  
- Format matches existing documentation style (title/metadata/related story).  
- Summary snippet suitable for copying into `roadmap.md` under S4 to reflect “closed” partials.

Output:
- Markdown entries for closure (complete snippet).  
- Suggested update text for roadmap to mark those partials resolved.


---------------------------------------------------------------------------
Prompt 5: Draft S5 acceptance test plan and implement core integration smoke test

Ground truth sources (authoritative): 
- docs/architecture/overview.md
- docs/development/principles.md
- docs/api/json-rpc-methods.md
- docs/development/documentation-guidelines.md

Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.
- Condensed roadmap/backlog focusing on S5 validation.

Directive: Do not add functionality beyond what's needed to exercise the existing implementation. If any end-to-end dependency is ambiguous (e.g., exact notification schema), refer to API doc; if still unclear, ask one precise question.

Goal:
- Draft a detailed acceptance test plan for S5 end-to-end flows (camera discovery → MediaMTX stream/record/snapshot → WebSocket notification → shutdown/error recovery).  
- Implement a core “happy path” integration smoke test that exercises the full flow and validates key state transitions.

Scope:
- New acceptance test plan document (e.g., `tests/ivv/acceptance-plan.md`).  
- One integration test/harness under `tests/integration/` or `tests/ivv/` implementing the smoke path.

Acceptance Criteria:
- Plan enumerates scenarios with clear success criteria, including recovery/error injection paths.  
- Smoke test performs: camera connect, capability detection, stream creation, start/stop recording, snapshot capture, and notification receipt with expected metadata.  
- Smoke test verifies end-to-end orchestration and surface failure if a critical step fails.

Output:
- Acceptance test plan file.  
- Working smoke test code.  
- Instructions for running it.  
- Any discovered gaps that block full flow (with evidence).


---------------------------------------------------------------------------
Prompt 6: Create missing camera_service support module test stubs (S14)

Ground truth sources (authoritative): 
- docs/architecture/overview.md
- docs/development/principles.md
- docs/api/json-rpc-methods.md
- docs/development/documentation-guidelines.md
- Current backlog emphasizes test readiness for support modules.
Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.

Directive: Only create structured stubs; do not implement full business logic in tests. Use canonical TODO/STOP formatting for placeholders.

Goal:
- Provide minimal pytest test stubs for:
  * `tests/unit/test_camera_service/test_main_startup.py`  
  * `tests/unit/test_camera_service/test_config_manager.py`  
  * `tests/unit/test_camera_service/test_logging_config.py`

Scope:
- Create the three test stub files.

Acceptance Criteria:
- Each stub contains:  
  * At least one placeholder test with a descriptive docstring.  
  * Canonical TODO comments detailing expected behavior (startup/shutdown, config fallback, formatter/correlation ID).  
  * Import of target module and setup scaffolding.  
- No attempt to assume future changes—tests describe acceptance without hard implementation.

Output:
- Three test stub files with skeleton test functions and TODOs.


---------------------------------------------------------------------------
Prompt 7: Add tests README and conventions doc (S14)
Ground truth sources (authoritative): 
- docs/architecture/overview.md
- docs/development/principles.md
- docs/api/json-rpc-methods.md
- docs/development/documentation-guidelines.md
- Existing test scaffolds and backlog identifying proliferation of test areas.
Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.

Directive: Do not introduce new test paradigms outside current structure. Clarify, don’t speculate.

Goal:
- Create `tests/unit/README.md` that explains:  
  * Purpose of each subdirectory (websocket_server, mediamtx_wrapper, camera_service, camera_discovery, common).  
  * Naming conventions (snake_case, `test_<behavior>.py`).  
  * Story-to-test mapping (S3/S4/S5/S14) with minimal checklist.  
  * How to add new tests and mark completion.  
  * How to run tests and interpret results.

Scope:
- Documentation only (single file).

Acceptance Criteria:
- README includes sections: Overview, Directory mapping, Naming conventions, Story/test mapping, Adding tests, Running tests, Contribution guidelines.  
- References the roadmap stories and expected acceptance criteria.  

Output:
- `tests/unit/README.md` content.


---------------------------------------------------------------------------
Prompt 8: Improve deployment/install script (S5)

Ground truth sources (authoritative): 
- docs/architecture/overview.md
- docs/development/principles.md
- docs/api/json-rpc-methods.md
- docs/development/documentation-guidelines.md
- Existing backlog wants repeatable environment for S5 validation.
Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.

Directive: Only enhance the existing `deployment/scripts/install.sh` (or equivalent) to bootstrap the current code. Do not redesign deployment architecture.

Goal:
- Complete the install script so that on a clean target it:  
  * Installs system dependencies required by the project.  
  * Sets up configuration (templates or example).  
  * Installs Python dependencies.  
  * Enables/starts the service (systemd or documented fallback).  
  * Is idempotent and safe to re-run.  
  * Provides a verification/smoke check at end.

Scope:
- Modify the install script and add minimal inline usage documentation.

Acceptance Criteria:
- Script can be run on a fresh environment (note assumed OS) and results in a running service.  
- Includes failure-safe re-execution logic.  
- Ends with a self-check (e.g., ping local API or check service health).  
- Comments document assumptions and usage.

Output:
- Updated `install.sh`.  
- Verification snippet/instructions.  
- Any remaining manual prerequisites clearly listed.


---------------------------------------------------------------------------
Prompt 9: Enable CI to enforce tests, linting, and type checking (S14)

Ground truth sources (authoritative): 
- docs/architecture/overview.md
- docs/development/principles.md
- docs/api/json-rpc-methods.md
- docs/development/documentation-guidelines.md
- Backlog prioritizing automated quality gates for S14.
Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.

Directive: Build a minimal pipeline; do not over-engineer (start with essential checks). All behavior must be explicit; if tool versions or defaults are ambiguous, state assumptions.

Goal:
- Create a CI workflow (e.g., GitHub Actions) that on push/PR:  
  * Installs dependencies.  
  * Runs formatting check (black in check mode).  
  * Runs linter/type-checker (flake8, mypy).  
  * Executes unit tests.  
  * Fails on any violation.

Scope:
- CI config file (e.g., `.github/workflows/ci.yml`) and any small helpers.

Acceptance Criteria:
- Workflow YAML exists and is functional.  
- Output shows clear separation of steps and fails visibly on errors.  
- README snippet included explaining how new tests or directories are picked up.  
- Badges/status suggestion included (optional).

Output:
- CI workflow config.  
- Documentation for extending it.


---------------------------------------------------------------------------
Prompt 10: Begin security feature implementation groundwork (E2)

Ground truth sources (authoritative): 
- docs/architecture/overview.md
- docs/development/principles.md
- docs/api/json-rpc-methods.md
- docs/development/documentation-guidelines.md
- Backlog beginning E2 groundwork.
Hard constraints:
- Only implement what is required by those ground truth documents and the current roadmap/backlog. Do not add, change, or invent any feature not present there.
- For any TODO/STOP or placeholder encountered, either replace it with real working code (if the behavior is defined) or normalize it to the canonical deferred format with rationale, date, and related story. 
- Update ONLY the file(s) explicitly named in the prompt; do not touch unrelated modules.
- For every change, include evidence: filename, specific section or line range, date (use current date), and ideally a commit reference if applicable.
- If any requirement is ambiguous or a needed detail is missing, STOP and list the ambiguity as one precise clarifying question instead of guessing.


Directive: Do not implement full security features prematurely. Focus on design, risk modeling, and scaffolding interfaces only. Stop and ask one focused question if an intended boundary or integration point is unclear.

Goal:
- Produce a lightweight security design and scaffold for:  
  * Authentication (JWT/API key) interface for WebSocket server.  
  * Health-check endpoint specification (contract only).  
  * Rate limiting / connection management sketch.  
  * TLS/SSL support checklist.  
  * Configuration schema for secrets, keys, certificates.

Scope:
- Security plan document.  
- Placeholder code stubs/interfaces (e.g., auth hook in WebSocket server, health-check handler skeleton).  
- Configuration definition (YAML/ENV) for security parameters.

Acceptance Criteria:
- Document includes threat model bullets, feature breakdown, configuration schema, and IV&V verification points.  
- Code stubs exist with canonical TODOs linking to plan items.  
- Clear next-step actionable list for the first security sprint.

Output:
- Security plan Markdown.  
- Code stub files or snippets.  
- Config schema sketch.  
- List of prioritized follow-up tasks.

----------------------------------------------------------------
Audit & Align: Test vs Implementation Sweep

Context: Solo-engineered camera service project. Ground truth is in:

    docs/architecture/overview.md

    docs/development/principles.md

    docs/development/documentation-guidelines.md

Scope: Focus on the camera_discovery module and its tests. Later reuse for other modules.

Goals:

    Identify mismatches between tests and code (e.g., tests mocking the wrong function, expecting missing helpers, wrong assumptions about return types).

    Detect and classify failing test patterns:

        TypeErrors from using objects like dicts (e.g., subscripting non-dict results).

        Assertion mismatches due to bounds/policy drift (e.g., frame rate thresholds).

        Missing utility methods the tests rely on.

        Error message content expectations not matching implementation (e.g., test expects “timeout” substring).

    Propose minimal, ground-truth-compliant fixes: change tests to align with implementation or adjust implementation when test expectations reflect intended behavior documented in architecture/principles.

Instructions:

    For each failing test, output: file/name, failure reason, current code snippet, proposed patch (diff-style), and which document justifies the intended behavior.

    Normalize any TODO/STOP comments encountered in both tests and source to canonical format if a decision is deferred.

    Do not invent new features; if test intent is unclear (e.g., what constitutes a “valid” frame rate), ask exactly one clarification question specifying the ambiguity.

    Group findings into: Fixed, Deferred (with canonical TODO), and Needs Clarification.

Example targets for this run:

    Add __getitem__ to CapabilityDetectionResult or adjust tests.

    Adjust frame rate parsing bounds so that 300 fps passes but 500 fps fails per test expectations.

    Ensure timeout handling includes “timeout” in error messages.

    Align failure simulation in tests with the asynchronous subprocess API (patch asyncio.create_subprocess_exec, not subprocess.run).

    Implement missing _extract_resolutions_from_output helper to satisfy test calls.

Output:

    Concrete patch snippets for each fix.

    Summary table of all current mismatches and their resolution state.

    Suggested next three tests to validate the regression surface after these changes.
-----------
**Audit Task – Static/Runtime Issue Sweep**
**Context:** Solo-engineered Python project with ground truth in:
* `docs/architecture/overview.md`
* `docs/development/principles.md`
* `docs/development/documentation-guidelines.md`
**Goal:** Perform a focused audit across the repository for the following classes of issues beyond long-line formatting:
1. Bare `except:` usage (flake8 E722) – replace with `except Exception:` or a more specific exception; add canonical TODO if the exact type is deferred.
2. Shadowed imports or names (F402) – identify where imported symbols (e.g., `field`) are reused as loop variables or overwritten; rename locals or remove unused imports.
3. Undefined names (F821) – locate uses of names that aren’t defined or imported (e.g., `get_current_config`) and either implement thin helpers in their intended modules or adjust callers to use existing APIs.
4. Non-dict objects being subscripted – detect usages like `obj["x"]` where `obj` is a custom object without `__getitem__`; either update the test/consumer to use attribute access or add a safe `__getitem__` adapter in the class with fallback to KeyError.
5. Any other non-formatting lint errors that imply potential logic bugs or omissions (e.g., unused imports, missing exception handling context).
**Instructions:**
* Only modify the minimal needed code to fix each issue; do not invent new features.
* For each finding, produce a table: file, line number(s), issue type, existing code snippet, proposed fix (diff-style), and rationale referencing the relevant story/standard (e.g., `[IV&V:S14]`).
* If a fix requires a decision (e.g., what exception to catch), insert a canonical TODO with rationale and stop further automated changes for that case.
* Do not aggregate multiple responsibilities into one vague change; keep one fix per issue item.
**Output:**
* Patch suggestions (diffs or updated snippets).
* Summary table of all issues found and their statuses (fixed / deferred with TODO / ambiguous needing clarification).
* A short list of top 3 blockers remaining after this sweep (if any), with precise questions if clarification is needed.
**Example fixes to include:**
* Replace bare `except:` with `except Exception:` and log.
* Add `__getitem__` to `CapabilityDetectionResult` to support legacy subscripting or update tests.
* Implement missing helper `get_current_config` or adjust calling tests.
**Acceptance:** All F722/F402/F821/non-subscriptable usage issues are either resolved or have explicit deferred decisions annotated; summary produced for inclusion in the roadmap/audit log.



Test Suite Execution & Failure Resolution

Context
Solo-engineered MediaMTX Camera Service project with established ground truth:
- docs/architecture/overview.md - System design and component interfaces
- docs/development/principles.md - Project values and TODO/STOP standards  
- docs/development/coding-standards.md - Style guide and technical requirements
- docs/development/documentation-guidelines.md - Documentation standards

Task Scope
Execute the COMPLETE test suite and generate minimal, targeted fixes for any failures across ALL modules. Focus ONLY on making tests pass while maintaining architectural compliance.

CRITICAL: Run ALL tests to 100% completion before analyzing any failures. Do not stop at early failures.

Execution Steps

1. Test Execution - MUST RUN TO COMPLETION
Run the complete test suite across all modules with these flags:
python3 -m pytest tests/ -v --tb=short --continue-on-collection-errors --maxfail=0

The --maxfail=0 flag means do not stop on any number of failures - run everything to completion.

If run_all_tests.py exists, use:
python3 run_all_tests.py --verbose --coverage --continue-on-errors

This MUST cover all modules:
- src/camera_discovery/
- src/camera_service/
- src/mediamtx_wrapper/
- src/websocket_server/
- src/common/

WAIT FOR 100% COMPLETION - Do not analyze failures until all tests have run.

2. Failure Analysis Protocol
Only after seeing the complete failure report, categorize each failure:
- Import/Missing Method: Method called in test doesn't exist in implementation
- Mock Mismatch: Test mocks wrong API (e.g. subprocess.run vs asyncio.create_subprocess_exec)
- Type Mismatch: Test expects dict but gets dataclass (or vice versa)
- Bounds/Validation: Test data exceeds implementation constraints
- Error Message: Test expects specific error text not present in implementation
- Async/Sync: Test doesn't properly handle async methods
- Configuration: Missing config values or environment setup
- Network/HTTP: Incorrect request/response mocking for API calls
- WebSocket: Improper WebSocket connection or message handling mocks

3. Fix Generation Rules

STRICT CONSTRAINTS:
- Fix ONLY what's needed to make the test pass
- NO new features or architectural changes
- NO scope creep beyond test failures
- Preserve all existing functionality
- Maintain type annotations and error handling
- Follow canonical TODO format if deferral needed:
  # TODO: PRIORITY: description [IV&V:ControlPoint|Story:Reference]

Fix Preference Order:
1. Adjust test if it tests unintended behavior
2. Add missing method if test assumes it exists and it should
3. Modify implementation only if test reflects documented requirements
4. Add compatibility layer for legacy test patterns

Output Format

For each fix, provide:

Fix Header
Fix #N: [Test Name] - [Issue Type]
File: path/to/file.py
Lines: X-Y  
Issue: Brief description
Justification: [Architecture doc reference]

Code Patch
# BEFORE (current code)
def existing_method():
    current_implementation()

# AFTER (fixed code)  
def existing_method():
    fixed_implementation()
    # Added: specific change made

Validation
✅ Test passes: test_specific_name
✅ No regressions: Related functionality unchanged
✅ Architecture compliance: [Reference to relevant doc section]

Module-Specific Areas to Check

Based on system architecture, pay attention to:

Camera Discovery Module:
- Missing extraction methods (_extract_resolutions_from_output, _extract_formats_from_output)
- Subprocess mocking mismatches (subprocess.run vs asyncio.create_subprocess_exec)
- Frame rate bounds validation (1-240 vs 1-400 range)
- Udev event handling
- CapabilityDetectionResult subscripting support

Camera Service Module:
- Configuration loading and validation
- Service manager lifecycle
- Environment variable overrides
- Config hot-reload functionality
- Missing get_current_config function

MediaMTX Wrapper Module:
- HTTP client mocking for MediaMTX API calls
- Configuration validation
- Health monitoring and circuit breaker logic
- Stream management operations
- Recording and snapshot operations

WebSocket Server Module:
- WebSocket connection handling
- JSON-RPC 2.0 message formatting
- Client authentication and subscription management
- Notification broadcasting
- API specification compliance

Common Module:
- Type definitions and data structures
- Shared utilities and helpers

Success Criteria

- All tests across ALL modules pass
- No new failures introduced in any module
- Coverage maintains or improves across entire codebase
- All fixes have architectural justification
- No scope creep beyond test failures
- Code follows project coding standards
- Changes documented with rationale

Constraints Summary

DO:
- Run complete test suite to 100% before making any changes
- Fix specific test failures with minimal changes across all modules
- Reference architecture docs for justification
- Maintain existing API contracts throughout system
- Add missing utility methods if tests expect them
- Use proper async/await patterns for all async code
- Ensure WebSocket and HTTP mocking is correct
- Validate configuration handling across all components

DON'T:
- Stop test execution early due to failures
- Add new features not required by tests
- Change architectural decisions
- Remove working functionality from any module
- Introduce breaking changes to any API
- Add dependencies without justification
- Make cosmetic changes unrelated to test failures
- Modify cross-module interfaces without considering impacts

IMPORTANT: First run all tests to completion, then provide a summary of all failures, then generate fixes. Do not make any code changes until you have the complete failure picture.

Deliverable: Working test suite with minimal, justified patches across ALL modules that maintain architectural integrity while ensuring complete test coverage passes.
