#!/usr/bin/env node

/**
 * Architecture Fix Script
 * Automatically fixes common architecture violations
 */

import { readFileSync, writeFileSync, readdirSync, statSync } from 'fs';
import { join, dirname, basename } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const projectRoot = join(__dirname, '..');

console.log('🔧 Architecture Fix Script');
console.log('=========================');

// RULE 1: Fix service imports in components to use stores
function fixServiceImportsInComponents() {
  console.log('Fixing service imports in components...');
  
  const componentsDir = join(projectRoot, 'src', 'components');
  const pagesDir = join(projectRoot, 'src', 'pages');
  
  function processDirectory(dir) {
    if (!statSync(dir).isDirectory()) return;
    
    const files = readdirSync(dir);
    files.forEach(file => {
      const filePath = join(dir, file);
      const stat = statSync(filePath);
      
      if (stat.isDirectory()) {
        processDirectory(filePath);
      } else if (file.endsWith('.tsx') || file.endsWith('.ts')) {
        try {
          let content = readFileSync(filePath, 'utf8');
          let modified = false;
          
          // Replace service imports with store imports
          const serviceImportRegex = /import\s*{\s*([^}]+)\s*}\s*from\s*['"]@?\/?services\/([^'"]+)['"]/g;
          content = content.replace(serviceImportRegex, (match, imports, servicePath) => {
            modified = true;
            console.log(`  - Replacing service import in ${basename(filePath)}`);
            
            // Map service to store
            const serviceName = servicePath.split('/').pop().replace(/\.ts$/, '');
            const storeName = serviceName.replace(/Service$/, 'Store');
            
            return `import { use${storeName} } from '@/stores/${storeName.toLowerCase()}'`;
          });
          
          if (modified) {
            writeFileSync(filePath, content, 'utf8');
          }
        } catch (error) {
          console.error(`Error processing ${filePath}:`, error.message);
        }
      }
    });
  }
  
  if (statSync(componentsDir).isDirectory()) {
    processDirectory(componentsDir);
  }
  if (statSync(pagesDir).isDirectory()) {
    processDirectory(pagesDir);
  }
}

// RULE 2: Fix WebSocketService usage in services to use APIClient
function fixWebSocketServiceUsage() {
  console.log('Fixing WebSocketService usage in services...');
  
  const servicesDir = join(projectRoot, 'src', 'services');
  
  function processDirectory(dir) {
    if (!statSync(dir).isDirectory()) return;
    
    const files = readdirSync(dir);
    files.forEach(file => {
      const filePath = join(dir, file);
      const stat = statSync(filePath);
      
      if (stat.isDirectory()) {
        processDirectory(filePath);
      } else if (file.endsWith('.ts') && !file.includes('APIClient.ts')) {
        try {
          let content = readFileSync(filePath, 'utf8');
          let modified = false;
          
          // Replace WebSocketService imports with APIClient
          if (content.includes('WebSocketService')) {
            modified = true;
            console.log(`  - Fixing WebSocketService usage in ${basename(filePath)}`);
            
            content = content.replace(
              /import\s*{\s*WebSocketService\s*}\s*from\s*['"][^'"]*['"]/g,
              "import { APIClient } from '@/services/abstraction/APIClient'"
            );
            
            content = content.replace(
              /new\s+WebSocketService\(/g,
              'new APIClient('
            );
            
            content = content.replace(
              /\.sendRPC\(/g,
              '.call('
            );
          }
          
          if (modified) {
            writeFileSync(filePath, content, 'utf8');
          }
        } catch (error) {
          console.error(`Error processing ${filePath}:`, error.message);
        }
      }
    });
  }
  
  if (statSync(servicesDir).isDirectory()) {
    processDirectory(servicesDir);
  }
}

// RULE 3: Fix service constructor patterns
function fixServiceConstructors() {
  console.log('Fixing service constructor patterns...');
  
  const servicesDir = join(projectRoot, 'src', 'services');
  
  function processDirectory(dir) {
    if (!statSync(dir).isDirectory()) return;
    
    const files = readdirSync(dir);
    files.forEach(file => {
      const filePath = join(dir, file);
      const stat = statSync(filePath);
      
      if (stat.isDirectory()) {
        processDirectory(filePath);
      } else if (file.endsWith('Service.ts') && !file.includes('APIClient.ts')) {
        try {
          let content = readFileSync(filePath, 'utf8');
          let modified = false;
          
          // Check if constructor needs APIClient and Logger
          if (content.includes('class') && content.includes('Service') && !content.includes('apiClient: APIClient')) {
            modified = true;
            console.log(`  - Adding proper constructor to ${basename(filePath)}`);
            
            // Add imports
            if (!content.includes("import { APIClient }")) {
              content = content.replace(
                /import\s+.*from\s+['"][^'"]*['"];?\s*\n/g,
                (match) => match + "import { APIClient } from '@/services/abstraction/APIClient';\nimport { LoggerService } from '@/services/logger/LoggerService';\n"
              );
            }
            
            // Add constructor if missing
            if (!content.includes('constructor(')) {
              const classMatch = content.match(/class\s+(\w+Service)\s*{/);
              if (classMatch) {
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
          }
          
          if (modified) {
            writeFileSync(filePath, content, 'utf8');
          }
        } catch (error) {
          console.error(`Error processing ${filePath}:`, error.message);
        }
      }
    });
  }
  
  if (statSync(servicesDir).isDirectory()) {
    processDirectory(servicesDir);
  }
}

// RULE 4: Remove component imports from stores
function fixStoreImports() {
  console.log('Fixing store imports...');
  
  const storesDir = join(projectRoot, 'src', 'stores');
  
  if (!statSync(storesDir).isDirectory()) return;
  
  const files = readdirSync(storesDir);
  files.forEach(file => {
    if (file.endsWith('.ts')) {
      const filePath = join(storesDir, file);
      try {
        let content = readFileSync(filePath, 'utf8');
        let modified = false;
        
        // Remove component imports
        const componentImportRegex = /import\s+.*from\s+['"]@?\/?components\/[^'"]*['"];?\s*\n/g;
        const matches = content.match(componentImportRegex);
        if (matches) {
          modified = true;
          console.log(`  - Removing component imports from ${basename(filePath)}`);
          content = content.replace(componentImportRegex, '');
        }
        
        // Remove page imports
        const pageImportRegex = /import\s+.*from\s+['"]@?\/?pages\/[^'"]*['"];?\s*\n/g;
        const pageMatches = content.match(pageImportRegex);
        if (pageMatches) {
          modified = true;
          console.log(`  - Removing page imports from ${basename(filePath)}`);
          content = content.replace(pageImportRegex, '');
        }
        
        if (modified) {
          writeFileSync(filePath, content, 'utf8');
        }
      } catch (error) {
        console.error(`Error processing ${filePath}:`, error.message);
      }
    }
  });
}

// Main execution
try {
  fixServiceImportsInComponents();
  fixWebSocketServiceUsage();
  fixServiceConstructors();
  fixStoreImports();
  
  console.log('=========================');
  console.log('✅ Architecture fixes completed!');
  console.log('Run "npm run arch:check" to verify the fixes.');
} catch (error) {
  console.error('❌ Error during architecture fixes:', error.message);
  process.exit(1);
}
