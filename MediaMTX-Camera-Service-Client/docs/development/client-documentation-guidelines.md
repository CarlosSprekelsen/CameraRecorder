---
title: "Client Documentation Guidelines"
description: "Markdown conventions, folder structure, and referencing existing server docs."
date: "2025-08-05"
---

# Documentation Guidelines

## Markdown conventions
- **Filenames**: use kebab-case, e.g. `client-documentation-guidelines.md`.
- **Front-matter**: each doc starts with YAML:
  ```yaml
  ---
  title: "Document Title"
  description: "Short description"
  date: "YYYY-MM-DD"
  ---
  ```
- **Headings**: ATX style (`#`, `##`, `###`).
- **Lists**: 2-space indent, unordered `-`.

## Where to put docs
- **Architecture**: `docs/architecture/`
- **API Reference**: `docs/api/`
- **Development**: `docs/development/`
- **Client-specific**: under `docs/development/client/` if needed.

## Referencing server docs
- **Never duplicate** server content. Instead link:
  > For full JSON-RPC API, see [Server API Reference](../api/json-rpc-methods.md).
- Use relative paths for internal links.
- If extending, add only client-specific examples.
