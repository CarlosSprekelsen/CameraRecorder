#!/usr/bin/env node

const WebSocket = require('ws');

// Test the storage info API fix
async function testStorageInfoFix() {
    console.log('🔍 Testing Storage Info API Fix...');
    
    const ws = new WebSocket('ws://localhost:8002/ws');
    
    ws.on('open', () => {
        console.log('✅ Connected to WebSocket server');
        
        // Send get_storage_info request
        const request = {
            jsonrpc: "2.0",
            method: "get_storage_info",
            id: 1
        };
        
        console.log('📤 Sending get_storage_info request...');
        ws.send(JSON.stringify(request));
    });
    
    ws.on('message', (data) => {
        try {
            const response = JSON.parse(data);
            console.log('📥 Received response:', JSON.stringify(response, null, 2));
            
            // Check if response has the correct fields
            if (response.result) {
                const result = response.result;
                console.log('\n🔍 Validating response fields:');
                
                // Check for required fields
                const requiredFields = [
                    'total_space',
                    'used_space', 
                    'available_space',
                    'usage_percentage',  // ← Fixed field name
                    'recordings_size',
                    'snapshots_size',
                    'low_space_warning'   // ← Added field
                ];
                
                let allFieldsPresent = true;
                requiredFields.forEach(field => {
                    if (result.hasOwnProperty(field)) {
                        console.log(`✅ ${field}: ${result[field]} (${typeof result[field]})`);
                    } else {
                        console.log(`❌ ${field}: MISSING`);
                        allFieldsPresent = false;
                    }
                });
                
                // Check field types
                const typeChecks = [
                    { field: 'total_space', expected: 'number' },
                    { field: 'used_space', expected: 'number' },
                    { field: 'available_space', expected: 'number' },
                    { field: 'usage_percentage', expected: 'number' },
                    { field: 'recordings_size', expected: 'number' },
                    { field: 'snapshots_size', expected: 'number' },
                    { field: 'low_space_warning', expected: 'boolean' }
                ];
                
                console.log('\n🔍 Validating field types:');
                let allTypesCorrect = true;
                typeChecks.forEach(check => {
                    if (result.hasOwnProperty(check.field)) {
                        const actualType = typeof result[check.field];
                        if (actualType === check.expected) {
                            console.log(`✅ ${check.field}: ${actualType} (correct)`);
                        } else {
                            console.log(`❌ ${check.field}: ${actualType} (expected ${check.expected})`);
                            allTypesCorrect = false;
                        }
                    }
                });
                
                // Final validation
                if (allFieldsPresent && allTypesCorrect) {
                    console.log('\n🎉 SUCCESS: Storage info API is now compliant with documentation!');
                    console.log('✅ All required fields are present');
                    console.log('✅ All field types are correct');
                    console.log('✅ Field names match API documentation');
                } else {
                    console.log('\n❌ FAILURE: Storage info API still has issues');
                    if (!allFieldsPresent) {
                        console.log('❌ Missing required fields');
                    }
                    if (!allTypesCorrect) {
                        console.log('❌ Incorrect field types');
                    }
                }
            } else {
                console.log('❌ No result in response');
            }
            
        } catch (error) {
            console.error('❌ Error parsing response:', error);
        }
        
        ws.close();
    });
    
    ws.on('error', (error) => {
        console.error('❌ WebSocket error:', error);
    });
    
    ws.on('close', () => {
        console.log('🔌 WebSocket connection closed');
        process.exit(0);
    });
}

// Run the test
testStorageInfoFix().catch(console.error);
