#!/usr/bin/env node

import fetch from 'node-fetch';

async function testFileDownload() {
    console.log('=== Camera Service File Download Test ===');
    console.log('Date:', new Date().toISOString());
    console.log('Server: http://localhost:8003');
    console.log('');

    const tests = [];
    let passed = 0;
    let failed = 0;

    // Test 1: Snapshot file download
    try {
        console.log('Test 1: Snapshot File Download...');
        const response = await fetch('http://localhost:8003/files/snapshots/test-snapshot.jpg');
        
        if (response.ok) {
            const content = await response.text();
            console.log('✅ Snapshot download successful');
            console.log(`   Status: ${response.status}`);
            console.log(`   Content-Length: ${response.headers.get('content-length')}`);
            console.log(`   Content-Type: ${response.headers.get('content-type')}`);
            console.log(`   Content Size: ${content.length} bytes`);
            tests.push({ name: 'Snapshot File Download', status: 'PASSED' });
            passed++;
        } else {
            console.log('❌ Snapshot download failed');
            console.log(`   Status: ${response.status}`);
            console.log(`   Error: ${response.statusText}`);
            tests.push({ name: 'Snapshot File Download', status: 'FAILED', error: `${response.status} ${response.statusText}` });
            failed++;
        }
    } catch (error) {
        console.log('❌ Test 1 failed:', error.message);
        tests.push({ name: 'Snapshot File Download', status: 'FAILED', error: error.message });
        failed++;
    }

    // Test 2: Recording file download
    try {
        console.log('\nTest 2: Recording File Download...');
        const response = await fetch('http://localhost:8003/files/recordings/test-recording.mp4');
        
        if (response.ok) {
            const content = await response.text();
            console.log('✅ Recording download successful');
            console.log(`   Status: ${response.status}`);
            console.log(`   Content-Length: ${response.headers.get('content-length')}`);
            console.log(`   Content-Type: ${response.headers.get('content-type')}`);
            console.log(`   Content Size: ${content.length} bytes`);
            tests.push({ name: 'Recording File Download', status: 'PASSED' });
            passed++;
        } else {
            console.log('❌ Recording download failed');
            console.log(`   Status: ${response.status}`);
            console.log(`   Error: ${response.statusText}`);
            tests.push({ name: 'Recording File Download', status: 'FAILED', error: `${response.status} ${response.statusText}` });
            failed++;
        }
    } catch (error) {
        console.log('❌ Test 2 failed:', error.message);
        tests.push({ name: 'Recording File Download', status: 'FAILED', error: error.message });
        failed++;
    }

    // Test 3: Missing file handling
    try {
        console.log('\nTest 3: Missing File Handling...');
        const response = await fetch('http://localhost:8003/files/snapshots/missing-file.jpg');
        
        if (response.status === 404) {
            console.log('✅ Missing file handled correctly');
            console.log(`   Status: ${response.status}`);
            console.log(`   Error: ${response.statusText}`);
            tests.push({ name: 'Missing File Handling', status: 'PASSED' });
            passed++;
        } else {
            console.log('❌ Missing file not handled correctly');
            console.log(`   Status: ${response.status}`);
            console.log(`   Expected: 404`);
            tests.push({ name: 'Missing File Handling', status: 'FAILED', error: `Expected 404, got ${response.status}` });
            failed++;
        }
    } catch (error) {
        console.log('❌ Test 3 failed:', error.message);
        tests.push({ name: 'Missing File Handling', status: 'FAILED', error: error.message });
        failed++;
    }

    // Test 4: Directory traversal protection
    try {
        console.log('\nTest 4: Directory Traversal Protection...');
        const response = await fetch('http://localhost:8003/files/snapshots/../../../etc/passwd');
        
        if (response.status === 404) {
            console.log('✅ Directory traversal blocked correctly');
            console.log(`   Status: ${response.status}`);
            console.log(`   Error: ${response.statusText}`);
            tests.push({ name: 'Directory Traversal Protection', status: 'PASSED' });
            passed++;
        } else {
            console.log('❌ Directory traversal not blocked');
            console.log(`   Status: ${response.status}`);
            console.log(`   Expected: 404`);
            tests.push({ name: 'Directory Traversal Protection', status: 'FAILED', error: `Expected 404, got ${response.status}` });
            failed++;
        }
    } catch (error) {
        console.log('❌ Test 4 failed:', error.message);
        tests.push({ name: 'Directory Traversal Protection', status: 'FAILED', error: error.message });
        failed++;
    }

    // Test 5: URL encoding handling
    try {
        console.log('\nTest 5: URL Encoding Handling...');
        const filename = encodeURIComponent('test file with spaces & special chars (2025).jpg');
        const response = await fetch(`http://localhost:8003/files/snapshots/${filename}`);
        
        if (response.status === 404) {
            console.log('✅ URL encoding handled correctly');
            console.log(`   Status: ${response.status}`);
            console.log(`   Filename: ${filename}`);
            tests.push({ name: 'URL Encoding Handling', status: 'PASSED' });
            passed++;
        } else {
            console.log('❌ URL encoding not handled correctly');
            console.log(`   Status: ${response.status}`);
            console.log(`   Expected: 404`);
            tests.push({ name: 'URL Encoding Handling', status: 'FAILED', error: `Expected 404, got ${response.status}` });
            failed++;
        }
    } catch (error) {
        console.log('❌ Test 5 failed:', error.message);
        tests.push({ name: 'URL Encoding Handling', status: 'FAILED', error: error.message });
        failed++;
    }

    // Test 6: Content-Type headers
    try {
        console.log('\nTest 6: Content-Type Headers...');
        const response = await fetch('http://localhost:8003/files/snapshots/test-snapshot.jpg');
        
        if (response.ok) {
            const contentType = response.headers.get('content-type');
            console.log('✅ Content-Type header present');
            console.log(`   Content-Type: ${contentType}`);
            
            if (contentType && contentType.includes('image/')) {
                console.log('✅ Correct image content type');
                tests.push({ name: 'Content-Type Headers', status: 'PASSED' });
                passed++;
            } else {
                console.log('❌ Incorrect content type');
                tests.push({ name: 'Content-Type Headers', status: 'FAILED', error: `Expected image/*, got ${contentType}` });
                failed++;
            }
        } else {
            console.log('❌ File download failed');
            tests.push({ name: 'Content-Type Headers', status: 'FAILED', error: `${response.status} ${response.statusText}` });
            failed++;
        }
    } catch (error) {
        console.log('❌ Test 6 failed:', error.message);
        tests.push({ name: 'Content-Type Headers', status: 'FAILED', error: error.message });
        failed++;
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
testFileDownload().catch(console.error);
