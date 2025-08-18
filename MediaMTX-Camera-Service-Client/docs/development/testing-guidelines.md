---
title: "Testing Guidelines"
description: "Folder structure, naming, mocking, and coverage targets for tests."
date: "2025-08-05"
---

# Testing Guidelines

## Folder structure
```
tests/
  unit/         # isolated component and utility tests
  integration/  # API interaction tests (MSW, hooks)
  e2e/          # end-to-end flows (Cypress)
```

## Naming conventions
- **Unit / Integration**: `*.test.ts`, `*.test.tsx` or `*.spec.ts[x]`.
- **E2E**: `*.e2e.ts` or place under `tests/e2e/`.

## Mocking strategies
- **HTTP**: MSW (Mock Service Worker).
- **WebSockets**: use `socket.io-mock`, or Jest event-emitter stubs.
- **Other**: Supply fixtures in `tests/fixtures/`.

## Coverage targets & thresholds
- **Unit**: ≥ 80%
- **Integration**: ≥ 70%
- **E2E**: smoke tests covering critical flows
- Configure in `jest.config.js`:
  ```js
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80
    }
  }
  ```
