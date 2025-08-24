---
title: "Client Coding Standards"
description: "ESLint/Prettier config, TypeScript best-practices, and React patterns for the client."
date: "2025-08-05"
---
STRICT RULES - HIGHEST PRIORITY
🚫 ABSOLUTE DO NOTS:
DO NOT COMPILE without explicit authorization
DO NOT FIX CODE without explicit authorization
DO NOT VIOLATE coding standards and naming strategy
DO NOT CREATE .md artifacts unless explicitly asked
DO NOT TAKE INITIATIVE - always ask for authorization first
DO NOT GET SIDE-TRACKED by error fixing or reactive work
✅ ABSOLUTE MUSTS:
ALWAYS ASK for authorization before any action
FOLLOW CODING STANDARDS strictly (ESLint/Prettier, TypeScript best-practices)
FOLLOW NAMING STRATEGY strictly (store prefixes, local prefixes, service prefixes)
FOLLOW ESTABLISHED PRIORITIES exactly as defined
FOCUS ON SYSTEMATIC APPROACH - not reactive error fixing
RESPECT GROUND TRUTH - architecture and requirements are authoritative

# Coding Standards

## ESLint + Prettier rules
- **ESLint**
  - Extend: `eslint:recommended`, `plugin:@typescript-eslint/recommended`, `plugin:react/recommended`, `plugin:react-hooks/recommended`
  - Parser: `@typescript-eslint/parser`
  - Plugins: `@typescript-eslint`, `react`, `react-hooks`, `prettier`
  - Rules:
    ```json
    {
      "rules": {
        "prettier/prettier": "error",
        "react/react-in-jsx-scope": "off",
        "@typescript-eslint/explicit-module-boundary-types": "warn"
      }
    }
    ```
- **Prettier** (`.prettierrc`)
  ```json
  {
    "singleQuote": true,
    "trailingComma": "all",
    "printWidth": 100,
    "endOfLine": "auto"
  }
  ```

## TypeScript best-practices
- **Strict mode** (`tsconfig.json`)
  ```json
  {
    "compilerOptions": {
      "strict": true,
      "noImplicitAny": true,
      "strictNullChecks": true,
      "noUncheckedIndexedAccess": true
    }
  }
  ```
- **No `any`**: Replace with `unknown`, `Record<string, unknown>`, or explicit interfaces.
- **Safe generics**:  
  - Always specify type parameters (`MyType<T>`).  
  - Constrain generics (`<T extends object>`).  
  - Avoid overly broad generics (`<T = any>`).

## React-specific patterns
- Use **functional components** only; **no class components**.
- Follow the [Rules of Hooks](https://reactjs.org/docs/hooks-rules.html):
  - Call hooks at the top level.
  - Only in React functions or custom hooks.
- Create **custom hooks** for shared logic (`useCamera`, `useWebSocket`).
- Manage side effects with `useEffect` and proper dependency arrays.
- Avoid inline functions in props; memoize with `useCallback` / `useMemo`.
- Use Context sparingly; prefer component-level state or Zustand.
