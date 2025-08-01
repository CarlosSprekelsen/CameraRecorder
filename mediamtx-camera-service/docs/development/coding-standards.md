# Coding Standards – MediaMTX Camera Service

**Version:** 1.0  
**Status:** Approved  
**Applies To:** All code and scripts in this repository

---

## 1. General Principles

- **Follow the approved architecture:**  
  All modules, classes, and functions must align with the structure and interfaces described in `docs/architecture/overview.md`.
- **Document before you code:**  
  Every public module, class, and method must include a clear, concise docstring describing its purpose, parameters, and return values (if any).
- **Single source of truth for configuration:**  
  All runtime parameters (e.g., paths, ports, environment settings) must be read from config files or environment variables—not hard-coded.
- **Professionalism:**  
  No emojis, ASCII art, colored log output, or non-standard formatting in any code, logs, or documentation.

---

## 2. Style Guide

- **Python version:**  
  Code must be compatible with Python 3.10+.
- **Linting and formatting:**  
  All code must pass `flake8` and be auto-formatted with `black`.
- **Naming conventions:**  
  - Modules: `lowercase_with_underscores.py`
  - Classes: `CamelCase`
  - Functions/Methods: `snake_case`
  - Constants: `UPPERCASE_WITH_UNDERSCORES`
- **Imports:**  
  - Group imports: standard library, then 3rd-party, then project modules.
  - Use explicit relative imports within a package (e.g., `from .foo import Bar`).
- **Type annotations:**  
  All public functions/methods must include type annotations for parameters and return values.
- **Error handling:**  
  Use explicit exception classes, log errors at the appropriate level, and avoid bare `except:` clauses.

---

## 3. Logging

- **Structured JSON logs:**  
  All production logs must be in structured JSON format (see `logging_config.py`).  
  Human-readable logs are permitted in development mode only.
- **Correlation/request IDs:**  
  All log records must include a correlation or request ID when available.
- **Log levels:**  
  Use `DEBUG` for dev-only details, `INFO` for normal ops, `WARNING` for recoverable issues, `ERROR` for failures, and `CRITICAL` for fatal conditions.

---

## 4. Documentation

- **Docstrings:**  
  Use standard triple-quoted ("""...""") docstrings for all public classes, methods, and functions.  
  - The first line is a short summary.  
  - Include descriptions for all parameters and return types (Google or NumPy style accepted).
- **API documentation:**  
  All public APIs must be documented in `docs/api/json-rpc-methods.md`.
- **README and module docs:**  
  Keep `README.md` focused on usage and pointers to docs—do not duplicate detailed instructions.

---

## 5. Testing

- **Test coverage:**  
  All modules must have corresponding unit tests in `tests/unit/`.
- **Test names:**  
  Name test files as `test_<module>.py` and use clear, descriptive test method names.
- **Test structure:**  
  Use pytest-style or unittest classes as appropriate.
- **Stubs/Scaffolds:**  
  New modules must include minimal test scaffolds (even if only `pass` or `TODO`) before implementation.

---

## 6. Code Review & CI

- **All code must be reviewed before merge.**
- **Continuous Integration:**  
  Code must pass all linting, formatting, and test checks in CI before being merged to main.

---

## 7. Security & Privacy

- **Never log secrets, passwords, or sensitive data.**
- **Use least-privilege for all system/service access.**

---

## 8. Updates & Violations

- Propose changes to this standard via pull request and review.
- Major violations are grounds for code review rejection.

---

**Questions?**  
Refer to `docs/development/principles.md` for the “why”—this document is the “how.”

