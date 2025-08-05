## 🌱 Cross-Epic Stories

### S14: Automated Testing & Continuous Integration - SUBSTANTIALLY COMPLETE
- Status: ✅ Substantially Complete  
- Summary: Test suite execution and failure resolution completed. Core testing infrastructure functional. Type checking errors reduced from 95 to 29. Remaining errors are non-blocking polish items.
- Evidence: Test execution artifacts (2025-08-05), functional test suite with `python3 run_all_tests.py`, error reduction across 7 files

### S15: Documentation & Developer Onboarding - PARTIALLY COMPLETE
- Status: 🟡 In progress  
- Summary: Core principles, coding standards, and architectural overview exist. Need to (a) capture resolved partials and decisions, (b) sync API docs, (c) provide a concise test/acceptance guide for upcoming IV&V work.  
- Key Deliverables:  
    - Update API docs to reflect actual implemented fields and behaviors.  
    - Document capability confirmation and health recovery policies.  
    - Provide a lightweight acceptance test plan for S5.  

---

## Backlog (Prioritized)

1. [DONE] Expand udev testing and metadata reconciliation (S3)  
2. [DONE] Harden and validate service manager lifecycle and observability (S3)  
3. [DONE] Add MediaMTX edge-case health monitor tests (S4)  
4. [DONE] Document closure of resolved partials (S4)
5. [DONE] Test Suite Execution & Failure Resolution
   - Completed: 2025-08-05
   - Evidence: Test execution, error reduction 95→29, functional pipeline
6. Draft S5 acceptance test plan and implement core integration smoke test  
7. Create missing camera_service support module test stubs (S14)  
8. Add tests README and conventions doc (S14)  
9. Improve deployment/install script (S5)  
10. Enable CI to enforce tests, linting, and type checking (S14)  
11. Begin security feature implementation groundwork (E2)
12. Polish remaining type checking errors (29 remaining - low priority)
13. Add comprehensive type annotations to untyped functions
14. Investigate WebSocket API compatibility updates
15. Fine-tune coverage thresholds per module criticality

---

## Status Summary

- **Architecture & Scaffolding (S1a/S2):** ✅ Complete  
- **Fast-track Audit (S2b):** ✅ Baseline captured and folded into stories  
- **Camera Discovery & Monitoring (S3):** ✅ Complete  
- **MediaMTX Integration (S4):** ✅ Complete — all partials resolved and documented (SC-1 through SC-5)
- **Core Integration IV&V (S5):** 🔴 Pending  
- **Testing & CI (S14):** ✅ Substantially Complete  
- **Documentation & Onboarding (S15):** 🟡 In progress  
- **Security (E2):** ⬜ Pending  
- **Client APIs/SDK (E3):** ⬜ Pending  
- **Extensibility (E4):** ⬜ Planning  
- **Deployment & Ops (E5):** ⬜ Pending