# Principles & Guidelines

These principles must be followed by all contributors and maintainers of the MediaMTX Camera Service project.

- [ ] **Test & Document Before Coding**  
      Write or update documentation and tests before (and during) implementing any new feature or change.
- [ ] **Single Source of Configuration**  
      All runtime parameters (ports, IPs, paths, etc.) must be read from a single, well-documented configuration file or environment variable loader. Never hard-code values.
- [ ] **Keep Code, Logs, and Docs Professional**  
      Do **not** use emojis, color formatting, or unprofessional decorations in code, documentation, logs, or messages.
- [ ] **Architecture Consistency**  
      All changes must respect the [architecture overview](../architecture/overview.md) and its component boundaries.
- [ ] **Strict Linting & Style Compliance**  
      All code must pass linting and auto-formatting before merge.
- [ ] **All Features Documented**  
      Every feature, public method, or config must be documented in `/docs` and have at least one usage/test example.

---

## Why These Principles?

- **Quality:** Well-tested and documented features are robust and maintainable.
- **Maintainability:** Consistent config and structure make it easy to adapt and scale.
- **Professionalism:** Clear, professional communication is critical for open-source and production environments.
- **Clarity:** Traceability and documentation prevent future errors and misunderstandings.

For more details, see the [project README](../../README.md) and [architecture overview](../architecture/overview.md).
