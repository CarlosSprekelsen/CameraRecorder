#!/usr/bin/env node

/**
 * Automated Duplication Detection Script
 * 
 * Purpose: Detect duplicate mock implementations across test files to prevent
 * violations of the "SINGLE mock implementation per API concern" rule.
 * 
 * Ground Truth References:
 * - Client Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Architecture: ../docs/architecture/client-architecture.md
 * 
 * Requirements Coverage:
 * - REQ-CI-001: Automated duplication detection
 * - REQ-CI-002: CI/CD integration
 * - REQ-CI-003: Pre-commit validation
 * - REQ-CI-004: Detailed reporting
 * 
 * Test Categories: CI/CD
 * API Documentation Reference: json-rpc-methods.md
 */

import { readFileSync, readdirSync, statSync } from 'fs';
import { join, relative, resolve } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = resolve(__filename, '..');

// Configuration
const CONFIG = {
  // Test directories to scan
  TEST_DIRS: [
    'tests/unit',
    'tests/integration', 
    'tests/e2e'
  ],
  
  // File patterns to include
  INCLUDE_PATTERNS: ['.test.ts', '.test.js', '.spec.ts', '.spec.js'],
  
  // File patterns to exclude
  EXCLUDE_PATTERNS: ['node_modules', '.git', 'coverage', 'dist'],
  
  // Duplication detection patterns
  DUPLICATION_PATTERNS: [
    // Service mock patterns
    {
      name: 'WebSocket Service Mock',
      pattern: /const\s+mockWebSocketService\s*=\s*\{[\s\S]*?sendRPC:\s*jest\.fn\(\)[\s\S]*?\}/g,
      severity: 'CRITICAL',
      rule: 'SINGLE mock implementation per API concern'
    },
    {
      name: 'Logger Service Mock',
      pattern: /const\s+mockLoggerService\s*=\s*\{[\s\S]*?info:\s*jest\.fn\(\)[\s\S]*?\}/g,
      severity: 'CRITICAL',
      rule: 'SINGLE mock implementation per API concern'
    },
    {
      name: 'Router Mock',
      pattern: /const\s+mockRouter\s*=\s*\{[\s\S]*?useNavigate:\s*jest\.fn\(\)[\s\S]*?\}/g,
      severity: 'HIGH',
      rule: 'SINGLE mock implementation per API concern'
    },
    {
      name: 'Auth Store Mock',
      pattern: /const\s+mockAuthStore\s*=\s*jest\.fn\(\)/g,
      severity: 'HIGH',
      rule: 'SINGLE mock implementation per API concern'
    },
    
    // Device service mock patterns
    {
      name: 'Device Service Mock',
      pattern: /const\s+mockDeviceService\s*=\s*\{[\s\S]*?getCameraList:\s*jest\.fn\(\)[\s\S]*?\}/g,
      severity: 'CRITICAL',
      rule: 'Use centralized MockDataFactory.createMockDeviceService()'
    },
    {
      name: 'File Service Mock',
      pattern: /const\s+mockFileService\s*=\s*\{[\s\S]*?listRecordings:\s*jest\.fn\(\)[\s\S]*?\}/g,
      severity: 'CRITICAL',
      rule: 'Use centralized MockDataFactory.createMockFileService()'
    },
    {
      name: 'Recording Service Mock',
      pattern: /const\s+mockRecordingService\s*=\s*\{[\s\S]*?takeSnapshot:\s*jest\.fn\(\)[\s\S]*?\}/g,
      severity: 'CRITICAL',
      rule: 'Use centralized MockDataFactory.createMockRecordingService()'
    },
    
    // Jest mock patterns
    {
      name: 'Jest Mock Function',
      pattern: /jest\.fn\(\s*\)/g,
      severity: 'MEDIUM',
      rule: 'Consider using centralized mocks when appropriate'
    },
    
    // Import duplication patterns
    {
      name: 'Duplicate Mock Import',
      pattern: /import.*from.*['"]\.\.\/.*\/mocks['"]/g,
      severity: 'LOW',
      rule: 'Ensure imports are using centralized mocks'
    }
  ],
  
  // Thresholds
  THRESHOLDS: {
    CRITICAL: 0,    // No critical duplications allowed
    HIGH: 2,        // Max 2 high severity duplications
    MEDIUM: 5,      // Max 5 medium severity duplications
    LOW: 10         // Max 10 low severity duplications
  }
};

class DuplicationDetector {
  constructor() {
    this.duplications = [];
    this.fileStats = new Map();
    this.severityCounts = { CRITICAL: 0, HIGH: 0, MEDIUM: 0, LOW: 0 };
  }

  /**
   * Main detection method
   */
  async detect() {
    console.log('🔍 Starting automated duplication detection...\n');
    
    const testFiles = this.findTestFiles();
    console.log(`📁 Found ${testFiles.length} test files to scan\n`);
    
    for (const filePath of testFiles) {
      await this.scanFile(filePath);
    }
    
    this.generateReport();
    return this.getExitCode();
  }

  /**
   * Find all test files to scan
   */
  findTestFiles() {
    const files = [];
    const projectRoot = resolve(__dirname, '..');
    
    for (const testDir of CONFIG.TEST_DIRS) {
      const fullPath = join(projectRoot, testDir);
      if (this.directoryExists(fullPath)) {
        files.push(...this.findFilesInDirectory(fullPath));
      }
    }
    
    return files;
  }

  /**
   * Recursively find files in directory
   */
  findFilesInDirectory(dirPath) {
    const files = [];
    
    try {
      const entries = readdirSync(dirPath);
      
      for (const entry of entries) {
        const fullPath = join(dirPath, entry);
        const stat = statSync(fullPath);
        
        if (stat.isDirectory()) {
          if (!CONFIG.EXCLUDE_PATTERNS.some(pattern => entry.includes(pattern))) {
            files.push(...this.findFilesInDirectory(fullPath));
          }
        } else if (stat.isFile()) {
          if (CONFIG.INCLUDE_PATTERNS.some(pattern => entry.includes(pattern))) {
            files.push(fullPath);
          }
        }
      }
    } catch (error) {
      console.warn(`⚠️  Warning: Could not read directory ${dirPath}: ${error.message}`);
    }
    
    return files;
  }

  /**
   * Check if directory exists
   */
  directoryExists(dirPath) {
    try {
      const stat = statSync(dirPath);
      return stat.isDirectory();
    } catch {
      return false;
    }
  }

  /**
   * Scan individual file for duplications
   */
  async scanFile(filePath) {
    try {
      const content = readFileSync(filePath, 'utf8');
      const relativePath = relative(resolve(__dirname, '..'), filePath);
      
      const fileDuplications = [];
      
      for (const pattern of CONFIG.DUPLICATION_PATTERNS) {
        const matches = [...content.matchAll(pattern.pattern)];
        
        if (matches.length > 0) {
          for (const match of matches) {
            const duplication = {
              file: relativePath,
              pattern: pattern.name,
              severity: pattern.severity,
              rule: pattern.rule,
              line: this.getLineNumber(content, match.index),
              match: match[0].substring(0, 100) + (match[0].length > 100 ? '...' : ''),
              position: match.index
            };
            
            fileDuplications.push(duplication);
            this.duplications.push(duplication);
            this.severityCounts[pattern.severity]++;
          }
        }
      }
      
      if (fileDuplications.length > 0) {
        this.fileStats.set(relativePath, fileDuplications);
      }
      
    } catch (error) {
      console.warn(`⚠️  Warning: Could not scan file ${filePath}: ${error.message}`);
    }
  }

  /**
   * Get line number from character index
   */
  getLineNumber(content, index) {
    return content.substring(0, index).split('\n').length;
  }

  /**
   * Generate detailed report
   */
  generateReport() {
    console.log('📊 DUPLICATION DETECTION REPORT');
    console.log('================================\n');
    
    // Summary
    this.printSummary();
    
    // Detailed findings
    if (this.duplications.length > 0) {
      this.printDetailedFindings();
      this.printRecommendations();
    } else {
      console.log('✅ No duplications found! All mocks are properly centralized.\n');
    }
    
    // Threshold violations
    this.printThresholdViolations();
  }

  /**
   * Print summary statistics
   */
  printSummary() {
    console.log('📈 SUMMARY:');
    console.log(`   Total files scanned: ${this.fileStats.size}`);
    console.log(`   Total duplications found: ${this.duplications.length}`);
    console.log(`   Critical: ${this.severityCounts.CRITICAL}`);
    console.log(`   High: ${this.severityCounts.HIGH}`);
    console.log(`   Medium: ${this.severityCounts.MEDIUM}`);
    console.log(`   Low: ${this.severityCounts.LOW}\n`);
  }

  /**
   * Print detailed findings by file
   */
  printDetailedFindings() {
    console.log('🔍 DETAILED FINDINGS:\n');
    
    for (const [file, duplications] of this.fileStats) {
      console.log(`📄 ${file}:`);
      
      // Group by severity
      const bySeverity = duplications.reduce((acc, dup) => {
        if (!acc[dup.severity]) acc[dup.severity] = [];
        acc[dup.severity].push(dup);
        return acc;
      }, {});
      
      for (const severity of ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW']) {
        if (bySeverity[severity]) {
          console.log(`   ${this.getSeverityIcon(severity)} ${severity}:`);
          for (const dup of bySeverity[severity]) {
            console.log(`      Line ${dup.line}: ${dup.pattern}`);
            console.log(`      Rule: ${dup.rule}`);
            console.log(`      Code: ${dup.match}\n`);
          }
        }
      }
      console.log('');
    }
  }

  /**
   * Print recommendations
   */
  printRecommendations() {
    console.log('💡 RECOMMENDATIONS:\n');
    
    const recommendations = [
      '1. Remove duplicate service mocks from individual test files',
      '2. Import centralized mocks from tests/utils/mocks.ts',
      '3. Use MockDataFactory.createMock*Service() methods',
      '4. Consolidate all jest.fn() calls into centralized implementations',
      '5. Run this script before committing changes'
    ];
    
    for (const rec of recommendations) {
      console.log(`   ${rec}`);
    }
    console.log('');
  }

  /**
   * Print threshold violations
   */
  printThresholdViolations() {
    console.log('⚠️  THRESHOLD VIOLATIONS:\n');
    
    let hasViolations = false;
    
    for (const [severity, count] of Object.entries(this.severityCounts)) {
      const threshold = CONFIG.THRESHOLDS[severity];
      if (count > threshold) {
        console.log(`   ${this.getSeverityIcon(severity)} ${severity}: ${count} found (threshold: ${threshold})`);
        hasViolations = true;
      }
    }
    
    if (!hasViolations) {
      console.log('   ✅ All thresholds within acceptable limits');
    }
    console.log('');
  }

  /**
   * Get severity icon
   */
  getSeverityIcon(severity) {
    const icons = {
      CRITICAL: '🚨',
      HIGH: '⚠️',
      MEDIUM: '🔶',
      LOW: 'ℹ️'
    };
    return icons[severity] || '📝';
  }

  /**
   * Get exit code based on violations
   */
  getExitCode() {
    for (const [severity, count] of Object.entries(this.severityCounts)) {
      const threshold = CONFIG.THRESHOLDS[severity];
      if (count > threshold) {
        return 1;
      }
    }
    return 0;
  }
}

// Main execution
async function main() {
  try {
    const detector = new DuplicationDetector();
    const exitCode = await detector.detect();
    process.exit(exitCode);
  } catch (error) {
    console.error('❌ Error during duplication detection:', error.message);
    process.exit(1);
  }
}

// Run if called directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}

export { DuplicationDetector, CONFIG };
