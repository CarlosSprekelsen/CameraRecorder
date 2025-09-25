#!/usr/bin/env node

/**
 * Pre-commit Duplication Check Script
 * 
 * Purpose: Quick duplication check for staged files before commit
 * Prevents duplicate mock implementations from being committed
 * 
 * Ground Truth References:
 * - Client Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Architecture: ../docs/architecture/client-architecture.md
 * 
 * Requirements Coverage:
 * - REQ-CI-001: Automated duplication detection
 * - REQ-CI-003: Pre-commit validation
 * - REQ-CI-005: Fast feedback loop
 * 
 * Test Categories: CI/CD
 * API Documentation Reference: json-rpc-methods.md
 */

import { readFileSync } from 'fs';
import { execSync } from 'child_process';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = resolve(__filename, '..');

// Quick patterns for pre-commit check
const QUICK_PATTERNS = [
  {
    name: 'WebSocket Service Mock',
    pattern: /const\s+mockWebSocketService\s*=\s*\{[\s\S]*?sendRPC:\s*jest\.fn\(\)[\s\S]*?\}/,
    severity: 'CRITICAL',
    fix: 'Use centralized MockDataFactory.createMockWebSocketService()'
  },
  {
    name: 'Logger Service Mock', 
    pattern: /const\s+mockLoggerService\s*=\s*\{[\s\S]*?info:\s*jest\.fn\(\)[\s\S]*?\}/,
    severity: 'CRITICAL',
    fix: 'Use centralized MockDataFactory.createMockLoggerService()'
  },
  {
    name: 'Device Service Mock',
    pattern: /const\s+mockDeviceService\s*=\s*\{[\s\S]*?getCameraList:\s*jest\.fn\(\)[\s\S]*?\}/,
    severity: 'CRITICAL',
    fix: 'Use MockDataFactory.createMockDeviceService()'
  },
  {
    name: 'File Service Mock',
    pattern: /const\s+mockFileService\s*=\s*\{[\s\S]*?listRecordings:\s*jest\.fn\(\)[\s\S]*?\}/,
    severity: 'CRITICAL',
    fix: 'Use MockDataFactory.createMockFileService()'
  },
  {
    name: 'Recording Service Mock',
    pattern: /const\s+mockRecordingService\s*=\s*\{[\s\S]*?takeSnapshot:\s*jest\.fn\(\)[\s\S]*?\}/,
    severity: 'CRITICAL',
    fix: 'Use MockDataFactory.createMockRecordingService()'
  }
];

class PreCommitChecker {
  constructor() {
    this.violations = [];
    this.projectRoot = resolve(__dirname, '..');
  }

  /**
   * Main pre-commit check
   */
  async check() {
    console.log('🔍 Running pre-commit duplication check...\n');
    
    const stagedFiles = this.getStagedFiles();
    const testFiles = stagedFiles.filter(file => this.isTestFile(file));
    
    if (testFiles.length === 0) {
      console.log('✅ No test files in staging area\n');
      return 0;
    }
    
    console.log(`📁 Checking ${testFiles.length} staged test files\n`);
    
    for (const file of testFiles) {
      await this.checkFile(file);
    }
    
    if (this.violations.length > 0) {
      this.printViolations();
      return 1;
    }
    
    console.log('✅ No duplications found in staged files\n');
    return 0;
  }

  /**
   * Get staged files from git
   */
  getStagedFiles() {
    try {
      const output = execSync('git diff --cached --name-only', { 
        encoding: 'utf8',
        cwd: this.projectRoot 
      });
      return output.trim().split('\n').filter(file => file.length > 0);
    } catch (error) {
      console.warn('⚠️  Warning: Could not get staged files:', error.message);
      return [];
    }
  }

  /**
   * Check if file is a test file
   */
  isTestFile(filePath) {
    return filePath.includes('tests/') && 
           (filePath.includes('.test.') || filePath.includes('.spec.')) &&
           (filePath.endsWith('.ts') || filePath.endsWith('.js'));
  }

  /**
   * Check individual file for violations
   */
  async checkFile(filePath) {
    try {
      const fullPath = resolve(this.projectRoot, filePath);
      const content = readFileSync(fullPath, 'utf8');
      
      for (const pattern of QUICK_PATTERNS) {
        const matches = content.match(pattern.pattern);
        if (matches) {
          const line = this.getLineNumber(content, content.indexOf(matches[0]));
          
          this.violations.push({
            file: filePath,
            pattern: pattern.name,
            severity: pattern.severity,
            fix: pattern.fix,
            line: line,
            match: matches[0].substring(0, 80) + (matches[0].length > 80 ? '...' : '')
          });
        }
      }
    } catch (error) {
      console.warn(`⚠️  Warning: Could not check file ${filePath}: ${error.message}`);
    }
  }

  /**
   * Get line number from character index
   */
  getLineNumber(content, index) {
    return content.substring(0, index).split('\n').length;
  }

  /**
   * Print violations and recommendations
   */
  printViolations() {
    console.log('🚨 DUPLICATION VIOLATIONS FOUND:\n');
    
    for (const violation of this.violations) {
      console.log(`📄 ${violation.file}:${violation.line}`);
      console.log(`   🚨 ${violation.severity}: ${violation.pattern}`);
      console.log(`   💡 Fix: ${violation.fix}`);
      console.log(`   📝 Code: ${violation.match}\n`);
    }
    
    console.log('🛠️  QUICK FIXES:\n');
    console.log('   1. Remove duplicate mock declarations');
    console.log('   2. Import centralized mocks:');
    console.log('      import { MockDataFactory } from \'../utils/mocks\';');
    console.log('   3. Use factory methods:');
    console.log('      const mockService = MockDataFactory.createMockDeviceService();\n');
    
    console.log('📚 For detailed guidance, see:');
    console.log('   - docs/development/client-testing-guidelines.md');
    console.log('   - tests/utils/mocks.ts (centralized implementations)\n');
  }
}

// Main execution
async function main() {
  try {
    const checker = new PreCommitChecker();
    const exitCode = await checker.check();
    process.exit(exitCode);
  } catch (error) {
    console.error('❌ Error during pre-commit check:', error.message);
    process.exit(1);
  }
}

// Run if called directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}

export { PreCommitChecker };
