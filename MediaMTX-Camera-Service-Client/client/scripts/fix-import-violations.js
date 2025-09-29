#!/usr/bin/env node

/**
 * Fix Import Violations Script
 * Automatically fixes import boundary violations
 */

import { readFileSync, writeFileSync, readdirSync, statSync } from 'fs';
import { join, dirname, basename } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const projectRoot = join(__dirname, '..');

console.log('🔧 Fixing Import Violations');
console.log('===========================');

// Service to Store mapping
const SERVICE_TO_STORE_MAP = {
  'CameraService': 'cameraStore',
  'RecordingService': 'recordingStore',
  'StreamService': 'streamStore',
  'AuthService': 'authStore',
  'ConfigService': 'configStore',
  'LoggerService': 'loggerStore'
};

function fixComponentImports() {
  console.log('Fixing component imports...');
  
  const componentsDir = join(projectRoot, 'src', 'components');
  const pagesDir = join(projectRoot, 'src', 'pages');
  
  function processDirectory(dir, type) {
    if (!statSync(dir).isDirectory()) return;
    
    const files = readdirSync(dir);
    files.forEach(file => {
      const filePath = join(dir, file);
      const stat = statSync(filePath);
      
      if (stat.isDirectory()) {
        processDirectory(filePath, type);
      } else if (file.endsWith('.tsx') || file.endsWith('.ts')) {
        try {
          let content = readFileSync(filePath, 'utf8');
          let modified = false;
          
          console.log(`Processing ${type}: ${basename(filePath)}`);
          
          // 1. Replace service imports with store imports
          const serviceImportRegex = /import\s*{\s*([^}]+)\s*}\s*from\s*['"]@?\/?services\/([^'"]+)['"];?\s*\n/g;
          content = content.replace(serviceImportRegex, (match, imports, servicePath) => {
            // Skip LoggerService as it's allowed
            if (servicePath.includes('LoggerService')) {
              return match;
            }
            
            modified = true;
            console.log(`  - Replacing service import: ${servicePath}`);
            
            // Extract service name and convert to store
            const serviceName = servicePath.split('/').pop().replace(/\.ts$/, '').replace('Service', '');
            const storeName = serviceName.charAt(0).toLowerCase() + serviceName.slice(1) + 'Store';
            
            return `import { use${serviceName}Store } from '@/stores/${storeName}';`;
          });
          
          // 2. Replace service usage with store usage
          Object.keys(SERVICE_TO_STORE_MAP).forEach(serviceName => {
            const storeName = SERVICE_TO_STORE_MAP[serviceName];
            const storeHook = `use${serviceName.replace('Service', 'Store')}`;
            
            // Replace service instantiation with store hook
            const serviceUsageRegex = new RegExp(`new\\s+${serviceName}\\s*\\(`, 'g');
            if (content.match(serviceUsageRegex)) {
              modified = true;
              console.log(`  - Replacing ${serviceName} usage with store hook`);
              
              // Add store hook import if not present
              if (!content.includes(storeHook)) {
                const serviceName = storeName.replace('Store', '');
                const importLine = `import { ${storeHook} } from '@/stores/${storeName}';`;
                
                // Find a good place to insert the import
                const lastImport = content.lastIndexOf('import ');
                if (lastImport !== -1) {
                  const nextLine = content.indexOf('\n', lastImport);
                  content = content.slice(0, nextLine + 1) + importLine + '\n' + content.slice(nextLine + 1);
                }
              }
              
              // Replace service usage with store hook
              content = content.replace(serviceUsageRegex, `${storeHook}()`);
            }
          });
          
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
  
  if (statSync(componentsDir).isDirectory()) {
    processDirectory(componentsDir, 'component');
  }
  if (statSync(pagesDir).isDirectory()) {
    processDirectory(pagesDir, 'page');
  }
}

function fixStoreImports() {
  console.log('Fixing store imports...');
  
  const storesDir = join(projectRoot, 'src', 'stores');
  
  if (!statSync(storesDir).isDirectory()) {
    console.log('❌ Stores directory not found');
    return;
  }
  
  const files = readdirSync(storesDir);
  files.forEach(file => {
    if (file.endsWith('.ts')) {
      const filePath = join(storesDir, file);
      try {
        let content = readFileSync(filePath, 'utf8');
        let modified = false;
        
        console.log(`Processing store: ${basename(filePath)}`);
        
        // Remove component imports
        const componentImportRegex = /import\s+.*from\s+['"]@?\/?components\/[^'"]*['"];?\s*\n/g;
        const componentMatches = content.match(componentImportRegex);
        if (componentMatches) {
          modified = true;
          console.log(`  - Removing component imports`);
          content = content.replace(componentImportRegex, '');
        }
        
        // Remove page imports
        const pageImportRegex = /import\s+.*from\s+['"]@?\/?pages\/[^'"]*['"];?\s*\n/g;
        const pageMatches = content.match(pageImportRegex);
        if (pageMatches) {
          modified = true;
          console.log(`  - Removing page imports`);
          content = content.replace(pageImportRegex, '');
        }
        
        // Remove service imports (except APIClient and LoggerService)
        const serviceImportRegex = /import\s+.*from\s+['"]@?\/?services\/(?!abstraction\/APIClient|logger\/LoggerService)[^'"]*['"];?\s*\n/g;
        const serviceMatches = content.match(serviceImportRegex);
        if (serviceMatches) {
          modified = true;
          console.log(`  - Removing service imports (except APIClient and LoggerService)`);
          content = content.replace(serviceImportRegex, '');
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

function fixServiceImports() {
  console.log('Fixing service imports...');
  
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
      } else if (file.endsWith('.ts') && !file.includes('APIClient.ts')) {
        try {
          let content = readFileSync(filePath, 'utf8');
          let modified = false;
          
          console.log(`Processing service: ${basename(filePath)}`);
          
          // Remove store imports
          const storeImportRegex = /import\s+.*from\s+['"]@?\/?stores\/[^'"]*['"];?\s*\n/g;
          const storeMatches = content.match(storeImportRegex);
          if (storeMatches) {
            modified = true;
            console.log(`  - Removing store imports`);
            content = content.replace(storeImportRegex, '');
          }
          
          // Remove component imports
          const componentImportRegex = /import\s+.*from\s+['"]@?\/?components\/[^'"]*['"];?\s*\n/g;
          const componentMatches = content.match(componentImportRegex);
          if (componentMatches) {
            modified = true;
            console.log(`  - Removing component imports`);
            content = content.replace(componentImportRegex, '');
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
  fixComponentImports();
  fixStoreImports();
  fixServiceImports();
  
  console.log('===========================');
  console.log('✅ Import violations fixed!');
  console.log('Run "npm run arch:check" to verify the fixes.');
} catch (error) {
  console.error('❌ Error during import fixes:', error.message);
  process.exit(1);
}
