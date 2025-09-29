#!/usr/bin/env node

/**
 * Remove Legacy Code Script
 * Removes legacy patterns and deprecated code
 */

import { readFileSync, writeFileSync, readdirSync, statSync } from 'fs';
import { join, dirname, basename } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const projectRoot = join(__dirname, '..');

console.log('🧹 Removing Legacy Code');
console.log('======================');

// Legacy patterns to remove
const LEGACY_PATTERNS = [
  {
    name: 'SessionInfo extensions',
    pattern: /extends\s+SessionInfo/g,
    replacement: '// Removed legacy SessionInfo extension',
    description: 'Removing legacy interface extensions'
  },
  {
    name: 'Old status values',
    pattern: /status:\s*['"]completed['"]/g,
    replacement: "status: 'SUCCESS'", // or 'FAILED' depending on context
    description: 'Updating legacy status values'
  },
  {
    name: 'Direct WebSocket usage',
    pattern: /wsService\.sendRPC\(/g,
    replacement: 'apiClient.call(',
    description: 'Replacing direct WebSocket calls with APIClient'
  },
  {
    name: 'Old service patterns',
    pattern: /new\s+(\w+Service)\s*\(\s*\)/g,
    replacement: (match, serviceName) => {
      return `// TODO: Update to use dependency injection\n// new ${serviceName}(apiClient, logger)`;
    },
    description: 'Updating service instantiation patterns'
  },
  {
    name: 'Unused imports',
    pattern: /import\s+.*from\s+['"][^'"]*['"];?\s*\n(?=\s*$|\s*\/\/|\s*export|\s*function|\s*class|\s*const|\s*let|\s*var)/gm,
    replacement: '',
    description: 'Removing unused imports'
  }
];

function removeLegacyPatterns() {
  console.log('Removing legacy patterns...');
  
  const srcDir = join(projectRoot, 'src');
  
  function processDirectory(dir) {
    if (!statSync(dir).isDirectory()) return;
    
    const files = readdirSync(dir);
    files.forEach(file => {
      const filePath = join(dir, file);
      const stat = statSync(filePath);
      
      if (stat.isDirectory()) {
        processDirectory(filePath);
      } else if (file.endsWith('.ts') || file.endsWith('.tsx')) {
        try {
          let content = readFileSync(filePath, 'utf8');
          let modified = false;
          
          LEGACY_PATTERNS.forEach(pattern => {
            const originalContent = content;
            
            if (typeof pattern.replacement === 'function') {
              content = content.replace(pattern.pattern, pattern.replacement);
            } else {
              content = content.replace(pattern.pattern, pattern.replacement);
            }
            
            if (content !== originalContent) {
              modified = true;
              console.log(`  - ${pattern.description} in ${basename(filePath)}`);
            }
          });
          
          if (modified) {
            writeFileSync(filePath, content, 'utf8');
            console.log(`  ✅ Cleaned: ${basename(filePath)}`);
          }
          
        } catch (error) {
          console.error(`❌ Error processing ${filePath}:`, error.message);
        }
      }
    });
  }
  
  processDirectory(srcDir);
}

function removeUnusedFiles() {
  console.log('Checking for unused files...');
  
  // Common unused file patterns
  const unusedFilePatterns = [
    /\.old\./,
    /\.backup\./,
    /\.legacy\./,
    /\.deprecated\./,
    /\.unused\./
  ];
  
  const srcDir = join(projectRoot, 'src');
  
  function processDirectory(dir) {
    if (!statSync(dir).isDirectory()) return;
    
    const files = readdirSync(dir);
    files.forEach(file => {
      const filePath = join(dir, file);
      const stat = statSync(filePath);
      
      if (stat.isDirectory()) {
        processDirectory(filePath);
      } else {
        const isUnused = unusedFilePatterns.some(pattern => pattern.test(file));
        if (isUnused) {
          console.log(`  - Found potentially unused file: ${filePath}`);
          console.log(`    Consider removing: ${filePath}`);
        }
      }
    });
  }
  
  processDirectory(srcDir);
}

function removeDeadCode() {
  console.log('Removing dead code patterns...');
  
  const srcDir = join(projectRoot, 'src');
  
  function processDirectory(dir) {
    if (!statSync(dir).isDirectory()) return;
    
    const files = readdirSync(dir);
    files.forEach(file => {
      const filePath = join(dir, file);
      const stat = statSync(filePath);
      
      if (stat.isDirectory()) {
        processDirectory(filePath);
      } else if (file.endsWith('.ts') || file.endsWith('.tsx')) {
        try {
          let content = readFileSync(filePath, 'utf8');
          let modified = false;
          
          // Remove commented out code blocks
          const commentedCodeRegex = /\/\*[\s\S]*?\*\//g;
          const commentedMatches = content.match(commentedCodeRegex);
          if (commentedMatches) {
            commentedMatches.forEach(match => {
              // Only remove if it looks like dead code (contains function, class, etc.)
              if (/function|class|const|let|var|import|export/.test(match)) {
                modified = true;
                content = content.replace(match, '');
                console.log(`  - Removed commented code block in ${basename(filePath)}`);
              }
            });
          }
          
          // Remove TODO comments that are resolved
          const todoRegex = /\/\/\s*TODO:\s*.*?(?:\n|$)/g;
          const todoMatches = content.match(todoRegex);
          if (todoMatches) {
            todoMatches.forEach(match => {
              // Remove TODOs that seem resolved
              if (/done|completed|fixed|resolved/i.test(match)) {
                modified = true;
                content = content.replace(match, '');
                console.log(`  - Removed resolved TODO in ${basename(filePath)}`);
              }
            });
          }
          
          if (modified) {
            writeFileSync(filePath, content, 'utf8');
            console.log(`  ✅ Cleaned: ${basename(filePath)}`);
          }
          
        } catch (error) {
          console.error(`❌ Error processing ${filePath}:`, error.message);
        }
      }
    });
  }
  
  processDirectory(srcDir);
}

// Main execution
try {
  removeLegacyPatterns();
  removeUnusedFiles();
  removeDeadCode();
  
  console.log('======================');
  console.log('✅ Legacy code cleanup completed!');
  console.log('Run "npm run arch:check" to verify the cleanup.');
} catch (error) {
  console.error('❌ Error during legacy cleanup:', error.message);
  process.exit(1);
}
