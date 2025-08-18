---
title: "CI/CD Guide"
description: "Outline of GitHub Actions workflow and how to extend it."
date: "2025-08-05"
---

# CI/CD Guide

## Workflow outline (GitHub Actions)
```yaml
name: CI

on: [pull_request, push]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: npm ci
      - run: npm run lint

  build:
    needs: lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: npm ci
      - run: npm run build

  test:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: npm ci
      - run: npm run test

  deploy:
    if: github.ref == 'refs/heads/main'
    needs: [lint, build, test]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: npm ci
      - run: npm run deploy
```

## Adding new steps
1. Edit `.github/workflows/ci.yml`.
2. Under `jobs:`, add a new job or step:
   ```yaml
   my-new-step:
     runs-on: ubuntu-latest
     steps:
       - run: npm run my-script
   ```
3. Reference outputs of previous jobs if needed.
