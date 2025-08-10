# Development Principles

These principles must be followed by all contributors and maintainers of the project. They exist to enforce quality, traceability, and professional consistency from architecture through implementation, testing, and operations.

## Core Principles

- **Test & Document Before Coding**  
  Write or update documentation and tests before (and during) implementing any new feature or change. No behavior is considered complete until it has corresponding documentation and test coverage.

- **Single Source of Configuration**  
  All runtime parameters (ports, IPs, paths, feature flags, etc.) must be read from a single, well-documented configuration system with clear precedence (defaults < YAML/config file < environment overrides). Hard-coded values are forbidden in production logic.

- **Keep Code, Logs, and Docs Professional**  
  Do **not** use emojis, unstructured color formatting, or informal decorations in code, documentation, logs, or messages. Clarity and consistency take precedence.

- **Architecture Consistency**  
  All changes must conform to the approved architecture. Any deviation requires an explicit architectural decision, documented in the Architecture Decisions section, and must pass IV&V before merging.

- **Strict Linting & Style Compliance**  
  All code must pass automated linting and formatting checks before merge. Style guidelines, naming conventions, and readability standards must be adhered to uniformly.

- **All Features Documented**  
  Every public API, configuration option, method, and behavioral change must be documented in `/docs` with at least one usage example or test demonstrating its intended use.

- **Traceability & Control Points**  
  Work flows through defined IV&V control points (architecture/scaffolding, implementation/integration, testing/verification, release/operations). No phase may advance without passing the prior gate with evidence: code, docs, tests, and explicit reviewer sign-off.

- **SDR Implementation Readiness**  
  Architecture must demonstrate implementation feasibility through proof-of-concept validation before detailed implementation begins. No implementation phase may commence without validated architecture components, integration patterns, and performance characteristics.

---

## TODO and STOP Comment Formatting Standard

All in-code TODO and STOP comments **must follow the exact format below**. This enables automated and human traceability, linking work to roadmap stories and IV&V control points. Comments must not be buried in prose.

### TODO Format

```
# TODO: <PRIORITY>: <short description> [IV&V:<ControlPointRef>|Story:<StoryRef>]
```

- `<PRIORITY>` is one of **CRITICAL**, **HIGH**, **MEDIUM**, **LOW**.
- `<short description>` is a concise statement of what needs implementing or fixing.
- `[IV&V:...]` or `[Story:...]` must include the exact reference from `roadmap.md`, e.g., `IV&V:S1b`, `Story:E1/S3`.
- At least one roadmap entry must exist for every TODO; the TODO must appear in the corresponding roadmap item.

Examples (copy-paste these lines as-is into code):
```python
# TODO: CRITICAL: Replace placeholder return with actual API call [IV&V:S1b]
# TODO: MEDIUM: Add schema validation to configuration loader [Story:E1/S1b]
```

### STOP Format (Blocked Work)

```
# STOP: <PRIORITY>: <reason> [IV&V:<ControlPointRef>]
```

- Used when progress is intentionally halted pending clarification, decision, or external dependency.
- Must have a corresponding blocker entry in `roadmap.md` under "STOP BLOCKAGES" with the same reference.

Example:
```python
# STOP: HIGH: Awaiting decision on including "metrics" field in camera status response [IV&V:S2]
```

### Rules for TODO/STOP Comments

- **Every** TODO/STOP must be tracked in `roadmap.md` with matching priority and reference. 
- **No TODO or STOP may remain unresolved** at the time of an IV&V control point sign-off or release unless explicitly deferred with a documented decision and reflected in the roadmap as a deferred item.
- **Once resolved**, the code must either remove the TODO/STOP or replace it with an audit note (e.g., `# DONE:` with a brief resolution summary), and the corresponding roadmap item must be updated to `[x]` with evidence.
- If the same underlying issue spans multiple files, one canonical TODO with cross-references is preferred; others may reference it (e.g., `[Story:E1/S1b] see core TODO in service_manager.py`).

---

## IV&V Alignment Requirements

- All work must map **forward** (roadmap → implementation/documentation) and **reverse** (implementation/documentation → roadmap).
- No feature, API method, configuration option, or component is "phantom": every implemented piece must have a corresponding roadmap item; every roadmap completion claim must be supported by tangible evidence (file/line/commit/test).
- Before any task is marked complete:
  - Code exists and exercises real behavior (no unimplemented stubs, `pass`, or placeholder responses).
  - Tests cover the behavior and pass.
  - Documentation reflects the change (API docs, architecture, README if applicable).
  - Reviewer has validated and signed off, with dates and references logged.

---

## Why These Principles?

- **Quality:** Well-tested, traceable, and documented changes reduce regressions and onboarding cost.
- **Maintainability:** Consistent conventions and single sources of truth make evolution safer.
- **Professionalism:** Clear communication and structured work reflect reliability to users and contributors.
- **Clarity:** Explicit linking between intent, design, implementation, and verification prevents drift and hidden debt.

For more details, see your project's README and architecture documentation.