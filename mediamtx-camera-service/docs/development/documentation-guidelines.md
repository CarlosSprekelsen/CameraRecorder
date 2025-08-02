# Documentation Guidelines

**Purpose:**  
Ensure all project documentation is consistent, discoverable, reviewable for IV&V, and aligned with the architecture and roadmap. This is the single source of truth for how to write, version, link, and evidence documentation across the repository.

## 1. Location & Scope
- **Primary guidelines file:** `docs/development/documentation-guidelines.md` (this file).  
- **Topic areas (keep only if multiple files are justified):**  
  - Architecture: `docs/architecture/*.md`  
  - API: `docs/api/*.md`  
  - Deployment: `docs/deployment/*.md`  
  - Development/process: `docs/development/*.md`  
  - Examples: `docs/examples/*.md`  
- **Test validation docs:** Place lightweight README/validation notes adjacent to the test area, e.g., `tests/unit/test_camera_discovery/README.md`. Refer back to this guidelines file for style and evidence conventions.

## 2. File Naming & Organization
- Use **snake_case** (lowercase with underscores) for documentation and code filenames: `camera_discovery_overview.md`, `capability_detection_validation.md`, `integration_acceptance_tests.md`.  
- **Exception:** Python test modules and validation scripts use the standard `test_` prefix in snake_case (e.g., `test_capability_detection.py`, `test_udev_processing.py`) to align with pytest conventions.  
- Each topic folder should only exist if it contains **two or more** related documents; avoid single-file subfolders unless grouping clearly anticipates growth.  
- Keep high-level entrypoints with stable names, for example:
  - `docs/architecture/overview.md`
  - `docs/api/json_rpc_methods.md`
  - `docs/development/setup.md`
  - `docs/development/coding_standards.md`
  - `docs/development/principles.md`
  - `docs/development/documentation_guidelines.md`


## 3. Document Structure Template
Every substantive `.md` (architecture decision, test validation, feature spec, acceptance criteria) should follow this minimal structure:

### Title and Metadata
```markdown
# <Human-readable title>
**Version:** x.y  
**Authors:** Name(s)  
**Date:** YYYY-MM-DD  
**Status:** draft | in review | approved  
**Related Epic/Story:** E1 / S3 / S2b etc.
