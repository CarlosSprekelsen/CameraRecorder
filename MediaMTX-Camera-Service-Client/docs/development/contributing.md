---
title: "Contributing Guide"
description: "How to branch, commit, and review for the client repo."
date: "2025-08-05"
---

# Contributing

## Branching model
- Base from `main`.
- Feature branches: `feature/short-description`.
- Hotfix branches: `hotfix/issue-number`.

## Pull Request Process
1. Open PR against `main`.
2. Link issue: `#123 <short description>`.
3. Assign reviewers and label accordingly.

## Semantic commit messages
- Format: `<type>(<scope>): <description>`
- Types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`.
- Example: `feat(camera): add snapshot history hook`.

## PR checklist
- [ ] Code has been linted (`npm run lint`).
- [ ] Tests are passing (`npm run test`).
- [ ] Documentation updated (if applicable).
- [ ] CI/CD pipeline green.
