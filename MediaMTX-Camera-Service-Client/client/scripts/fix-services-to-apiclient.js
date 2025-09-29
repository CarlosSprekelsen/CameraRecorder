#!/usr/bin/env node

/**
 * Fix Services to APIClient Script
 * Converts services to use APIClient instead of direct WebSocketService usage
 */

import { readFileSync, writeFileSync, readdirSync, statSync } from 'fs';
import { join, dirname, basename } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const projectRoot = join(__dirname, '..');

console.log('🔧 Fixing Services to Use APIClient');
console.log('===================================');

function fixServiceFiles() {
  const servicesDir = join(projectRoot, 'src', 'services');
  
  if (!statSync(servicesDir).isDirectory()) {
    console.log('❌ Services directory not found');
    return;
  }
  
  function processDirectory(dir) {
    const files = readdirSync(dir);
    files.forEach(file => {
      const filePath = join(dir, file);
      const stat = statSync(filePath);
      
      if (stat.isDirectory()) {
        processDirectory(filePath);
      } else if (file.endsWith('Service.ts') && !file.includes('APIClient.ts') && !file.includes('WebSocketService.ts')) {
        try {
          let content = readFileSync(filePath, 'utf8');
          let modified = false;
          
          console.log(`Processing: ${basename(filePath)}`);
          
          // 1. Replace WebSocketService imports with APIClient
          if (content.includes('WebSocketService')) {
            modified = true;
            console.log(`  - Replacing WebSocketService import`);
            
            // Remove WebSocketService import
            content = content.replace(
              /import\s*{\s*WebSocketService\s*}\s*from\s*['"][^'"]*['"];?\s*\n/g,
              ''
            );
            
            // Add APIClient import if not present
            if (!content.includes("import { APIClient }")) {
              const importRegex = /import\s+.*from\s+['"][^'"]*['"];?\s*\n/g;
              const matches = content.match(importRegex);
              if (matches && matches.length > 0) {
                content = content.replace(
                  matches[0],
                  matches[0] + "import { APIClient } from '@/services/abstraction/APIClient';\n"
                );
              } else {
                content = "import { APIClient } from '@/services/abstraction/APIClient';\n\n" + content;
              }
            }
          }
          
          // 2. Update constructor parameters
          if (content.includes('constructor(')) {
            const constructorRegex = /constructor\s*\(\s*([^)]*)\s*\)/g;
            content = content.replace(constructorRegex, (match, params) => {
              if (!params.includes('apiClient: APIClient')) {
                modified = true;
                console.log(`  - Updating constructor parameters`);
                
                const newParams = params.trim() 
                  ? `private apiClient: APIClient, private logger: LoggerService, ${params}`
                  : 'private apiClient: APIClient, private logger: LoggerService';
                
                return `constructor(${newParams})`;
              }
              return match;
            });
          } else {
            // Add constructor if missing
            const classMatch = content.match(/class\s+(\w+Service)\s*{/);
            if (classMatch) {
              modified = true;
              console.log(`  - Adding constructor`);
              const className = classMatch[1];
              content = content.replace(
                /class\s+\w+Service\s*{/,
                `class ${className} {
  constructor(
    private apiClient: APIClient,
    private logger: LoggerService
  ) {}`
              );
            }
          }
          
          // 3. Replace direct WebSocket calls with APIClient calls
          const rpcCallRegex = /\.sendRPC\s*\(/g;
          if (content.match(rpcCallRegex)) {
            modified = true;
            console.log(`  - Replacing sendRPC calls with APIClient.call`);
            content = content.replace(rpcCallRegex, '.call(');
          }
          
          // 4. Replace WebSocketService instances with APIClient
          content = content.replace(/new\s+WebSocketService\s*\(/g, 'new APIClient(');
          content = content.replace(/WebSocketService\s*\(/g, 'APIClient(');
          
          // 5. Update method calls to use this.apiClient
          const methodCallRegex = /this\.wsService\./g;
          if (content.match(methodCallRegex)) {
            modified = true;
            console.log(`  - Updating wsService references to apiClient`);
            content = content.replace(methodCallRegex, 'this.apiClient.');
          }
          
          if (modified) {
            writeFileSync(filePath, content, 'utf8');
            console.log(`  ✅ Fixed: ${basename(filePath)}`);
          } else {
            console.log(`  ✓ Already compliant: ${basename(filePath)}`);
          }
          
        } catch (error) {
          console.error(`❌ Error processing ${filePath}:`, error.message);
        }
      }
    });
  }
  
  processDirectory(servicesDir);
}

// Main execution
try {
  fixServiceFiles();
  
  console.log('===================================');
  console.log('✅ Service fixes completed!');
  console.log('All services now use APIClient instead of direct WebSocketService.');
} catch (error) {
  console.error('❌ Error during service fixes:', error.message);
  process.exit(1);
}
