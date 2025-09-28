#!/bin/bash

# Fix test imports and service initialization for architecture compliance

echo "🔧 Fixing test files for architecture compliance..."

# List of test files to fix
TEST_FILES=(
    "tests/integration/server_connectivity.test.ts"
    "tests/integration/contract_validation.test.ts" 
    "tests/integration/security.test.ts"
    "tests/integration/performance.test.ts"
)

for file in "${TEST_FILES[@]}"; do
    echo "Fixing $file..."
    
    # Add APIClient import
    sed -i '/import.*WebSocketService/a import { APIClient } from '\''../../src/services/abstraction/APIClient'\'';' "$file"
    
    # Replace service initialization patterns
    sed -i 's/new AuthService(webSocketService)/new AuthService(apiClient, loggerService)/g' "$file"
    sed -i 's/new DeviceService(webSocketService, loggerService)/new DeviceService(apiClient, loggerService)/g' "$file"
    sed -i 's/new FileService(webSocketService, loggerService)/new FileService(apiClient, loggerService)/g' "$file"
    sed -i 's/new RecordingService(webSocketService, loggerService)/new RecordingService(apiClient, loggerService)/g' "$file"
    sed -i 's/new ServerService(webSocketService, loggerService)/new ServerService(apiClient, loggerService)/g' "$file"
    
    # Add APIClient creation before service initialization
    sed -i '/await webSocketService.connect()/a\    \n    // Create APIClient for services\n    const apiClient = new APIClient(webSocketService, loggerService);' "$file"
    
    echo "✅ Fixed $file"
done

echo "🎉 All test files fixed for architecture compliance!"
