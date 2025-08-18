#!/usr/bin/env node

import WebSocket from 'ws';

async function testWebSocketIntegration() {
    console.log('=== Camera Service WebSocket Integration Test ===');
    console.log('Date:', new Date().toISOString());
    console.log('Server: ws://localhost:8002/ws');
    console.log('');

    const tests = [];
    let passed = 0;
    let failed = 0;

    // Test 1: WebSocket Connection
    try {
        console.log('Test 1: WebSocket Connection...');
        const ws = new WebSocket('ws://localhost:8002/ws');
        
        await new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error('Connection timeout'));
            }, 5000);

            ws.on('open', () => {
                clearTimeout(timeout);
                console.log('✅ WebSocket connection established');
                tests.push({ name: 'WebSocket Connection', status: 'PASSED' });
                passed++;
                ws.close();
                resolve();
            });

            ws.on('error', (error) => {
                clearTimeout(timeout);
                console.log('❌ WebSocket connection failed:', error.message);
                tests.push({ name: 'WebSocket Connection', status: 'FAILED', error: error.message });
                failed++;
                reject(error);
            });
        });
    } catch (error) {
        console.log('❌ Test 1 failed:', error.message);
    }

    // Test 2: JSON-RPC get_camera_list
    try {
        console.log('\nTest 2: get_camera_list API...');
        const ws = new WebSocket('ws://localhost:8002/ws');
        
        await new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error('Request timeout'));
            }, 10000);

            ws.on('open', () => {
                const request = {
                    jsonrpc: '2.0',
                    id: 1,
                    method: 'get_camera_list'
                };
                ws.send(JSON.stringify(request));
            });

            ws.on('message', (data) => {
                try {
                    const response = JSON.parse(data.toString());
                    clearTimeout(timeout);
                    
                    if (response.jsonrpc === '2.0' && response.id === 1) {
                        console.log('✅ get_camera_list response received');
                        console.log('   Result:', JSON.stringify(response.result, null, 2));
                        tests.push({ name: 'get_camera_list API', status: 'PASSED' });
                        passed++;
                    } else {
                        console.log('❌ Invalid response format');
                        tests.push({ name: 'get_camera_list API', status: 'FAILED', error: 'Invalid response format' });
                        failed++;
                    }
                    ws.close();
                    resolve();
                } catch (error) {
                    clearTimeout(timeout);
                    console.log('❌ Failed to parse response:', error.message);
                    tests.push({ name: 'get_camera_list API', status: 'FAILED', error: error.message });
                    failed++;
                    reject(error);
                }
            });

            ws.on('error', (error) => {
                clearTimeout(timeout);
                console.log('❌ WebSocket error:', error.message);
                tests.push({ name: 'get_camera_list API', status: 'FAILED', error: error.message });
                failed++;
                reject(error);
            });
        });
    } catch (error) {
        console.log('❌ Test 2 failed:', error.message);
    }

    // Test 3: JSON-RPC list_snapshots
    try {
        console.log('\nTest 3: list_snapshots API...');
        const ws = new WebSocket('ws://localhost:8002/ws');
        
        await new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error('Request timeout'));
            }, 10000);

            ws.on('open', () => {
                const request = {
                    jsonrpc: '2.0',
                    id: 2,
                    method: 'list_snapshots'
                };
                ws.send(JSON.stringify(request));
            });

            ws.on('message', (data) => {
                try {
                    const response = JSON.parse(data.toString());
                    clearTimeout(timeout);
                    
                    if (response.jsonrpc === '2.0' && response.id === 2) {
                        console.log('✅ list_snapshots response received');
                        console.log('   Result:', JSON.stringify(response.result, null, 2));
                        tests.push({ name: 'list_snapshots API', status: 'PASSED' });
                        passed++;
                    } else {
                        console.log('❌ Invalid response format');
                        tests.push({ name: 'list_snapshots API', status: 'FAILED', error: 'Invalid response format' });
                        failed++;
                    }
                    ws.close();
                    resolve();
                } catch (error) {
                    clearTimeout(timeout);
                    console.log('❌ Failed to parse response:', error.message);
                    tests.push({ name: 'list_snapshots API', status: 'FAILED', error: error.message });
                    failed++;
                    reject(error);
                }
            });

            ws.on('error', (error) => {
                clearTimeout(timeout);
                console.log('❌ WebSocket error:', error.message);
                tests.push({ name: 'list_snapshots API', status: 'FAILED', error: error.message });
                failed++;
                reject(error);
            });
        });
    } catch (error) {
        console.log('❌ Test 3 failed:', error.message);
    }

    // Test 4: JSON-RPC list_recordings
    try {
        console.log('\nTest 4: list_recordings API...');
        const ws = new WebSocket('ws://localhost:8002/ws');
        
        await new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error('Request timeout'));
            }, 10000);

            ws.on('open', () => {
                const request = {
                    jsonrpc: '2.0',
                    id: 3,
                    method: 'list_recordings'
                };
                ws.send(JSON.stringify(request));
            });

            ws.on('message', (data) => {
                try {
                    const response = JSON.parse(data.toString());
                    clearTimeout(timeout);
                    
                    if (response.jsonrpc === '2.0' && response.id === 3) {
                        console.log('✅ list_recordings response received');
                        console.log('   Result:', JSON.stringify(response.result, null, 2));
                        tests.push({ name: 'list_recordings API', status: 'PASSED' });
                        passed++;
                    } else {
                        console.log('❌ Invalid response format');
                        tests.push({ name: 'list_recordings API', status: 'FAILED', error: 'Invalid response format' });
                        failed++;
                    }
                    ws.close();
                    resolve();
                } catch (error) {
                    clearTimeout(timeout);
                    console.log('❌ Failed to parse response:', error.message);
                    tests.push({ name: 'list_recordings API', status: 'FAILED', error: error.message });
                    failed++;
                    reject(error);
                }
            });

            ws.on('error', (error) => {
                clearTimeout(timeout);
                console.log('❌ WebSocket error:', error.message);
                tests.push({ name: 'list_recordings API', status: 'FAILED', error: error.message });
                failed++;
                reject(error);
            });
        });
    } catch (error) {
        console.log('❌ Test 4 failed:', error.message);
    }

    // Summary
    console.log('\n=== Test Summary ===');
    console.log(`Total Tests: ${tests.length}`);
    console.log(`Passed: ${passed}`);
    console.log(`Failed: ${failed}`);
    console.log(`Success Rate: ${((passed / tests.length) * 100).toFixed(1)}%`);
    
    console.log('\n=== Detailed Results ===');
    tests.forEach(test => {
        const status = test.status === 'PASSED' ? '✅' : '❌';
        console.log(`${status} ${test.name}: ${test.status}`);
        if (test.error) {
            console.log(`   Error: ${test.error}`);
        }
    });

    return { tests, passed, failed, successRate: (passed / tests.length) * 100 };
}

// Run the test
testWebSocketIntegration().catch(console.error);
