---
title: "Release & Deployment Guide"
description: "How to build, host the PWA, and manage env vars & secrets."
date: "2025-08-05"
---

# Release & Deployment Guide

## Building the PWA
- Run `npm run build` – outputs to `dist/`.
- Verify service worker registration in `vite.config.ts`.

## Hosting
- Serve static `dist/` via CDN or static host (Netlify, Vercel, S3+CloudFront).
- Ensure `index.html` fallback routing for SPA.

## Service Worker
- Register in your entrypoint (`src/main.tsx`):
  ```ts
  if ('serviceWorker' in navigator) {
    window.addEventListener('load', () =>
      navigator.serviceWorker.register('/service-worker.js')
    );
  }
  ```

## Environment variables & secret management
- **Local**: `.env.development`, `.env.production` (never commit secrets).
- **Prefix**: only `VITE_` variables are exposed to client.
- **CI/CD**: store secrets in GitHub Actions Secrets; inject via:
  ```yaml
  - name: Build with env
    run: npm run build
    env:
      VITE_API_URL: ${{ secrets.VITE_API_URL }}
  ```
- **Never** commit `.env.*` to source control.
