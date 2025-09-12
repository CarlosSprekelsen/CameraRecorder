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
