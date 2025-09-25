# CI/CD Duplication Detection

**Version:** 1.0  
**Date:** 2025-01-15  
**Status:** Active - Automated duplication prevention system  

## Overview

Automated duplication detection system for MediaMTX Camera Service Client that prevents violations of the "SINGLE mock implementation per API concern" rule through CI/CD integration.

## Ground Truth References

- **Testing Guidelines**: [client-testing-guidelines.md](./client-testing-guidelines.md)
- **Architecture**: [client-architecture.md](../architecture/client-architecture.md)
- **API Documentation**: [json-rpc-methods.md](../../mediamtx-camera-service-go/docs/api/json_rpc_methods.md)

## Requirements Coverage

- **REQ-CI-001**: Automated duplication detection
- **REQ-CI-002**: CI/CD integration
- **REQ-CI-003**: Pre-commit validation
- **REQ-CI-004**: Detailed reporting
- **REQ-CI-005**: Fast feedback loop

## System Components

### 1. Duplication Detection Script

**Location**: `scripts/detect-duplications.js`

**Purpose**: Comprehensive scanning of all test files for duplicate mock implementations.

**Features**:
- Scans all test directories (`tests/unit`, `tests/integration`, `tests/e2e`)
- Detects 8+ types of mock duplications
- Severity-based classification (CRITICAL, HIGH, MEDIUM, LOW)
- Configurable thresholds
- Detailed reporting with line numbers
- Exit codes for CI/CD integration

**Usage**:
```bash
npm run test:duplication-check
node scripts/detect-duplications.js
```

### 2. Pre-commit Hook Script

**Location**: `scripts/pre-commit-duplication-check.js`

**Purpose**: Fast duplication check on staged files before commit.

**Features**:
- Quick scan of only staged test files
- Focused on critical duplications
- Fast execution (< 1 second)
- Git integration
- Immediate feedback

**Usage**:
```bash
npm run test:duplication-check-quick
# Automatically runs on git commit via Husky
```

### 3. GitHub Actions Workflow

**Location**: `.github/workflows/duplication-detection.yml`

**Purpose**: Automated duplication detection in CI/CD pipeline.

**Triggers**:
- Push to `main`, `develop`, `feature/*` branches
- Pull requests to `main`, `develop`
- Manual workflow dispatch
- Changes to test files or scripts

**Jobs**:
1. **detect-duplications**: Main duplication detection
2. **test-consistency**: Ensures tests still pass after mock consolidation
3. **security-scan**: Security audit and secret detection
4. **notify-on-failure**: Failure notifications

### 4. Husky Pre-commit Hook

**Location**: `.husky/pre-commit`

**Purpose**: Prevent commits with duplicate mocks.

**Features**:
- Automatic execution on `git commit`
- Blocks commits with violations
- Quick feedback loop
- Integration with existing pre-commit hooks

## Configuration

### Duplication Patterns

The system detects the following patterns:

| Pattern | Severity | Rule |
|---------|----------|------|
| `mockWebSocketService` | CRITICAL | Use centralized MockDataFactory |
| `mockLoggerService` | CRITICAL | Use centralized MockDataFactory |
| `mockDeviceService` | CRITICAL | Use MockDataFactory.createMockDeviceService() |
| `mockFileService` | CRITICAL | Use MockDataFactory.createMockFileService() |
| `mockRecordingService` | CRITICAL | Use MockDataFactory.createMockRecordingService() |
| `jest.fn()` | MEDIUM | Consider centralized alternatives |
| Duplicate imports | LOW | Ensure centralized usage |

### Thresholds

| Severity | Threshold | Action |
|----------|-----------|---------|
| CRITICAL | 0 | Block CI/CD |
| HIGH | 2 | Warning |
| MEDIUM | 5 | Info |
| LOW | 10 | Info |

## Setup Instructions

### 1. Automated Setup

```bash
# Run the setup script
cd client
chmod +x scripts/setup-cicd.sh
./scripts/setup-cicd.sh
```

### 2. Manual Setup

```bash
# Install dependencies
npm install --save-dev husky@^8.0.3

# Initialize Husky
npx husky init

# Make scripts executable
chmod +x scripts/detect-duplications.js
chmod +x scripts/pre-commit-duplication-check.js

# Create pre-commit hook
echo 'node scripts/pre-commit-duplication-check.js' > .husky/pre-commit
chmod +x .husky/pre-commit
```

## Usage

### Local Development

```bash
# Full duplication scan
npm run test:duplication-check

# Quick pre-commit check
npm run test:duplication-check-quick

# All CI checks including duplication
npm run ci:all
```

### Git Workflow

```bash
# Pre-commit hook runs automatically
git add .
git commit -m "Add new test"  # Hook runs here

# Manual pre-commit check
npm run test:duplication-check-quick
```

### CI/CD Pipeline

The GitHub Actions workflow runs automatically on:
- Push to main/develop branches
- Pull requests
- Manual trigger

## Reporting

### Console Output

```
🔍 Starting automated duplication detection...

📊 DUPLICATION DETECTION REPORT
================================

📈 SUMMARY:
   Total files scanned: 15
   Total duplications found: 3
   Critical: 2
   High: 1
   Medium: 0
   Low: 0

🔍 DETAILED FINDINGS:

📄 tests/unit/services/test_file_service.ts:
   🚨 CRITICAL: WebSocket Service Mock
   Rule: SINGLE mock implementation per API concern
   Code: const mockWebSocketService = { sendRPC: jest.fn() }...

💡 RECOMMENDATIONS:
   1. Remove duplicate service mocks from individual test files
   2. Import centralized mocks from tests/utils/mocks.ts
   3. Use MockDataFactory.createMock*Service() methods
```

### GitHub Actions Output

- Detailed logs in Actions tab
- PR comments on violations
- Artifact uploads for reports
- Status checks on PRs

## Troubleshooting

### Common Issues

1. **Script Permission Denied**
   ```bash
   chmod +x scripts/detect-duplications.js
   chmod +x scripts/pre-commit-duplication-check.js
   ```

2. **Husky Hook Not Running**
   ```bash
   npx husky init
   chmod +x .husky/pre-commit
   ```

3. **GitHub Actions Not Triggering**
   - Check workflow file syntax
   - Verify branch protection rules
   - Ensure workflow file is in `.github/workflows/`

4. **False Positives**
   - Update patterns in `CONFIG.DUPLICATION_PATTERNS`
   - Adjust thresholds in `CONFIG.THRESHOLDS`
   - Add exclusions to `CONFIG.EXCLUDE_PATTERNS`

### Debug Mode

```bash
# Enable debug output
DEBUG=true node scripts/detect-duplications.js

# Verbose git output
GIT_TRACE=1 git commit -m "test"
```

## Integration with Existing Workflows

### ESLint Integration

```javascript
// .eslintrc.js
module.exports = {
  rules: {
    // Custom rule to detect mock duplications
    'no-duplicate-mocks': 'error'
  }
};
```

### Jest Integration

```javascript
// jest.config.js
module.exports = {
  setupFilesAfterEnv: ['<rootDir>/tests/setup.ts'],
  globalTeardown: '<rootDir>/scripts/teardown-duplication-check.js'
};
```

### Prettier Integration

```javascript
// .prettierrc.js
module.exports = {
  // Ensure consistent mock formatting
  plugins: ['prettier-plugin-duplication-check']
};
```

## Performance

### Benchmarks

| Operation | Time | Files Scanned |
|-----------|------|---------------|
| Full scan | ~2-3s | 50+ files |
| Pre-commit | ~0.5s | 5-10 files |
| CI/CD scan | ~5-8s | All files |

### Optimization

- Incremental scanning for large codebases
- Parallel processing for multiple files
- Caching of scan results
- Exclude patterns for irrelevant files

## Maintenance

### Regular Tasks

1. **Update Patterns** (monthly)
   - Review new mock patterns
   - Update detection rules
   - Adjust thresholds

2. **Performance Review** (quarterly)
   - Benchmark scan times
   - Optimize slow operations
   - Update dependencies

3. **Documentation Updates** (as needed)
   - Update setup instructions
   - Add new troubleshooting steps
   - Document new features

### Version Compatibility

| Node.js | NPM | Husky | Status |
|---------|-----|-------|---------|
| 20.19.0+ | 10.8.0+ | 8.0.3+ | ✅ Supported |
| 18.x | 9.x | 7.x | ⚠️ Limited |
| <18 | <9 | <7 | ❌ Not supported |

## Security Considerations

- No sensitive data in scan results
- Read-only file system access
- No network operations during scans
- Secure artifact handling in CI/CD

## Contributing

### Adding New Patterns

1. Update `CONFIG.DUPLICATION_PATTERNS` in `detect-duplications.js`
2. Add corresponding quick pattern in `pre-commit-duplication-check.js`
3. Update documentation
4. Test with existing codebase

### Reporting Issues

1. Check existing issues in repository
2. Provide detailed reproduction steps
3. Include relevant file paths and error messages
4. Test with latest version

---

**Maintenance**: This document should be updated whenever duplication detection patterns or CI/CD workflows are modified.
