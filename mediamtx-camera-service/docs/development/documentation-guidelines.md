# Documentation Guidelines

**Version:** 2.0  
**Authors:** Project Team  
**Date:** 2025-08-07  
**Status:** Approved  
**Related Epic/Story:** All development stories  

**Purpose:**  
Ensure all project documentation is consistent, discoverable, reviewable for IV&V, and aligned with the approved architecture and roadmap. This is the single source of truth for how to write, version, link, and evidence documentation across the repository.

---

## 1. Location & Scope

### Primary Structure
- **Primary guidelines file:** `docs/development/documentation-guidelines.md` (this file)
- **Topic areas (only create if multiple files are justified):**
  - Architecture: `docs/architecture/*.md`
  - API: `docs/api/*.md`
  - Deployment: `docs/deployment/*.md`
  - Development/process: `docs/development/*.md`
  - Examples: `docs/examples/*.md`
  - Decisions: `docs/decisions.md` (architecture decisions log)

### Test Documentation
- **Test validation docs:** Place lightweight README/validation notes adjacent to the test area
  - Example: `tests/unit/test_camera_discovery/README.md`
  - Refer back to this guidelines file for style and evidence conventions
- **Integration test docs:** `tests/ivv/*.md` for acceptance criteria and execution instructions

---

## 2. File Naming & Organization

### Naming Conventions
- Use **snake_case** (lowercase with underscores) for all documentation and code filenames
  - Examples: `camera_discovery_overview.md`, `capability_detection_validation.md`, `integration_acceptance_tests.md`
- **Exception:** Python test modules use standard `test_` prefix in snake_case (e.g., `test_capability_detection.py`, `test_udev_processing.py`) to align with pytest conventions

### Directory Structure Rules
- Each topic folder should only exist if it contains **two or more** related documents
- Avoid single-file subfolders unless grouping clearly anticipates growth
- Keep high-level entrypoints with stable names:
  - `docs/architecture/overview.md`
  - `docs/api/json_rpc_methods.md`
  - `docs/development/setup.md`
  - `docs/development/coding_standards.md`
  - `docs/development/principles.md`
  - `docs/development/documentation_guidelines.md`
  - `docs/decisions.md`
  - `docs/roadmap.md`

---

## 3. Document Structure Template

Every substantive `.md` file (architecture decision, test validation, feature spec, acceptance criteria) **must** follow this structure:

### Title and Metadata Block
```markdown
# <Human-readable title>
**Version:** x.y  
**Authors:** Name(s)  
**Date:** YYYY-MM-DD  
**Status:** draft | in review | approved  
**Related Epic/Story:** E1 / S3 / S2b etc.
```

### Required Sections
1. **Purpose/Overview:** Brief statement of the document's goal and scope
2. **Main Content:** Organized with descriptive headers (no generic "Details" or "Information")
3. **Evidence/References:** Links to related code, tests, or other documents where applicable
4. **Next Steps/Actions:** If the document drives implementation work

### Content Requirements
- Use clear, descriptive section headers (sentence case preferred)
- Include concrete examples where applicable
- Link to relevant code files, test cases, or other documentation
- Maintain professional tone (no emojis, informal language, or decorative elements)

---

## 4. Professional Standards

### Language and Tone
- **No emojis, ASCII art, or decorative elements** in any documentation
- Use clear, professional language appropriate for technical documentation
- Avoid informal expressions, slang, or conversational asides
- Write in active voice when possible

### Formatting Standards
- Use standard Markdown formatting consistently
- Code blocks must specify language for syntax highlighting
- Use bullet points sparingly; prefer numbered lists for sequences or prose paragraphs for explanations
- Bold key terms or important concepts, but avoid excessive formatting

---

## 5. Progress Monitoring and Task Centralization

### Single Source of Truth for Progress
- **ALL active TODOs, task lists, and implementation checklists** must be consolidated in `docs/roadmap.md`
- **Do NOT** append new TODOs or checklists to other documentation files (e.g., `overview.md`, `decisions.md`, architecture docs)
- **Do NOT** create separate sprint reports, progress files, or task tracking documents

### Preventing Documentation Creep
- **No standalone progress documents:** All task tracking happens in the centralized roadmap only
- **No outdated task lists:** If a task is moved or completed, update only the roadmap - do not leave copies elsewhere
- **No non-validated reports:** Progress claims must have evidence (code/tests/documentation) before being marked complete

### Task Management Rules
- New action items or decisions must be added to `docs/roadmap.md` only
- Architecture documents remain stable and do not contain evolving TODOs
- Any document claiming completion must reference the roadmap story and provide evidence

---

## 6. Professional Standards

## 6. Roadmap and IV&V Integration
- Every document that describes implementation work **must** reference the corresponding roadmap story (e.g., "Story: E1/S3")
- Documents describing completed work must include evidence section with specific file/line references

### TODO/STOP Standards in Documentation
Follow the same standards as code (from `docs/development/principles.md`):

```markdown
<!-- TODO: HIGH: Add API examples for snapshot capture [Story:E1/S3] -->
<!-- STOP: MEDIUM: Awaiting decision on authentication method [IV&V:S2] -->
```

### Evidence Requirements
For any document claiming implementation completion:
- **File references:** Specific files and line ranges where behavior is implemented
- **Test references:** Tests that validate the documented behavior
- **Configuration examples:** Sample configurations demonstrating usage
- **Commit/date stamps:** When available, reference specific commits or dates

---

### Linking to Roadmap

### JSON-RPC Methods
- All public API methods **must** be documented in `docs/api/json_rpc_methods.md`
- Include request/response schemas with examples
- Document error conditions and response codes
- Provide at least one complete usage example per method

### Configuration Documentation
- All configuration options must be documented with:
  - Purpose and default value
  - Valid ranges or options
  - Environment variable override (if applicable)
  - Example usage

## 7. API Documentation Standards

### JSON-RPC Methods
- All public API methods **must** be documented in `docs/api/json_rpc_methods.md`
- Include request/response schemas with examples
- Document error conditions and response codes
- Provide at least one complete usage example per method

### Configuration Documentation
- All configuration options must be documented with:
  - Purpose and default value
  - Valid ranges or options
  - Environment variable override (if applicable)
  - Example usage

---

## 8. Architecture Documentation

### Stability Requirements
- `docs/architecture/overview.md` must reflect **only approved, stable architecture**
- Do **not** include process checklists, evolving TODOs, or implementation steps in architecture docs
- Changes to architecture require explicit approval and entry in `docs/decisions.md`

### Decision Logging
- Architectural and technical decisions belong in `docs/decisions.md`
- Record **what was decided, when, and briefly why**
- Do not include ongoing tasks or implementation details in decisions

---

## 9. Test Documentation

### Unit Test Documentation
- Each test module should include a brief header comment explaining its scope
- Complex test scenarios should include setup/teardown documentation
- Test README files should explain the testing strategy for that component

### Integration Test Documentation
- Document test scenarios with clear success criteria
- Include setup instructions and prerequisites
- Specify expected outputs and failure modes
- Provide troubleshooting guidance for common issues

---

## 10. Maintenance and Updates

### Version Control
- Increment version number for significant changes to document structure or requirements
- Update date stamp when making substantial content changes
- Maintain change history for critical architecture or API documentation

### Review Process
- All documentation changes follow the same review process as code
- Architecture documentation requires explicit approval from project maintainer
- API documentation must be validated against actual implementation

### Consistency Checks
- Regularly audit documentation for compliance with these guidelines
- Ensure all TODO/STOP items are tracked in roadmap
- Verify all implementation claims have supporting evidence

---

## 11. Common Patterns and Examples

### Stability Requirements
- `docs/architecture/overview.md` must reflect **only approved, stable architecture**
- Do **not** include process checklists, evolving TODOs, or implementation steps in architecture docs
- Changes to architecture require explicit approval and entry in `docs/decisions.md`

### Decision Logging
- Architectural and technical decisions belong in `docs/decisions.md`
- Record **what was decided, when, and briefly why**
- Do not include ongoing tasks or implementation details in decisions

---

## 8. Test Documentation

### Unit Test Documentation
- Each test module should include a brief header comment explaining its scope
- Complex test scenarios should include setup/teardown documentation
- Test README files should explain the testing strategy for that component

### Integration Test Documentation
- Document test scenarios with clear success criteria
- Include setup instructions and prerequisites
- Specify expected outputs and failure modes
- Provide troubleshooting guidance for common issues

---

## 9. Maintenance and Updates

### Version Control
- Increment version number for significant changes to document structure or requirements
- Update date stamp when making substantial content changes
- Maintain change history for critical architecture or API documentation

### Review Process
- All documentation changes follow the same review process as code
- Architecture documentation requires explicit approval from project maintainer
- API documentation must be validated against actual implementation

### Consistency Checks
- Regularly audit documentation for compliance with these guidelines
- Ensure all TODO/STOP items are tracked in roadmap
- Verify all implementation claims have supporting evidence

---

## 10. Common Patterns and Examples

### Linking Between Documents
```markdown
See [Architecture Overview](../architecture/overview.md) for system design details.
Refer to [Coding Standards](./coding_standards.md) for implementation requirements.
```

### Code References
```markdown
Implementation in `src/camera_discovery/hybrid_monitor.py` lines 45-67.
Test coverage in `tests/unit/test_camera_discovery/test_hybrid_monitor.py`.
Configuration schema in `config/camera_service.yaml` section `discovery.monitors`.
```

### Evidence Examples
```markdown
**Evidence:**
- Feature implemented: `src/websocket_server/notification_handler.py:23-45`
- Tests passing: `tests/unit/test_websocket_server/test_notifications.py`
- API documented: `docs/api/json_rpc_methods.md#camera-status-notification`
- Config example: `config/examples/basic_setup.yaml`
```

---

## 12. Quality Checklist

Before finalizing any documentation:

- [ ] Follows required structure template with metadata block
- [ ] Uses professional language with no emojis or informal elements
- [ ] Includes concrete examples where applicable
- [ ] Links to related code/tests/configuration as appropriate
- [ ] References relevant roadmap stories or IV&V control points
- [ ] Passes spell check and grammar review
- [ ] Verified against actual implementation (for technical docs)

---

**Questions or Clarifications?**  
See `docs/development/principles.md` for project values and `docs/roadmap.md` for current development priorities.