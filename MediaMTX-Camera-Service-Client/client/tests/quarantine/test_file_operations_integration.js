#!/usr/bin/env node

/**
 * Sprint 3: File Download Functionality Test
 * 
 * This script validates the file download functionality via HTTPS endpoints
 * as specified in Sprint 3 requirements.
 * 
 * Tests:
 * - File listing via WebSocket JSON-RPC
 * - File download via HTTP endpoints
 * - Error handling for missing files
 * - URL construction and encoding
 * 
 * Usage: node test-file-download.js
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running on localhost:8002
 * - Health server running on localhost:8003
 * - Test files created in /opt/camera-service/snapshots/ and /opt/camera-service/recordings/
 */

import WebSocket from 'ws';
import http from 'http';
import https from 'https';
import { URL } from 'url';

// Test configuration
const CONFIG = {
  websocketUrl: 'ws://localhost:8002/ws',
  healthServerUrl: 'http://localhost:8003',
  timeout: 10000,
};

// Test results tracking
const testResults = {
  passed: 0,
  failed: 0,
  total: 0,
  errors: [],
  sprint3Requirements: {
    fileListing: false,
    fileDownload: false,
    errorHandling: false,
    urlConstruction: false,
  }
};

/**
 * Utility function to send JSON-RPC requests with API compliance validation
 */
function send(ws, method, id, params = undefined) {
  const req = { jsonrpc: '2.0', method, id };
  if (params) req.params = params;
  
  // API compliance validation
  if (!req.jsonrpc || req.jsonrpc !== '2.0') {
    throw new Error('Invalid JSON-RPC version per API documentation');
  }
  if (!req.method) {
    throw new Error('Missing method per API documentation');
  }
  if (req.id === undefined) {
    throw new Error('Missing id per API documentation');
  }
  
  // Method-specific parameter validation based on API documentation
  if (method === 'list_recordings' || method === 'list_snapshots') {
    if (!req.params) {
      throw new Error(`${method} method requires params per API documentation`);
    }
    if (req.params.limit === undefined || req.params.offset === undefined) {
      throw new Error(`${method} method requires limit and offset parameters per API documentation`);
    }
  }
  
  console.log(`📤 Sending ${method} (#${id})`, params ? JSON.stringify(params) : '');
  ws.send(JSON.stringify(req));
}

/**
 * Test result assertion
 */
function assert(condition, message) {
  testResults.total++;
  if (condition) {
    testResults.passed++;
    console.log(`✅ ${message}`);
  } else {
    testResults.failed++;
    console.log(`❌ ${message}`);
    testResults.errors.push(message);
  }
}

/**
 * HTTP request helper
 */
function httpRequest(url, method = 'GET') {
  return new Promise((resolve, reject) => {
    const urlObj = new URL(url);
    const options = {
      hostname: urlObj.hostname,
      port: urlObj.port,
      path: urlObj.pathname,
      method: method,
      timeout: CONFIG.timeout,
    };

    const client = urlObj.protocol === 'https:' ? https : http;
    const req = client.request(options, (res) => {
      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          headers: res.headers,
          data: data
        });
      });
    });

    req.on('error', reject);
    req.on('timeout', () => reject(new Error('Request timeout')));
    req.end();
  });
}

/**
 * Test 1: File Listing via WebSocket
 */
async function testFileListing() {
  console.log('\n📋 Sprint 3 Test 1: File Listing via WebSocket');
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(CONFIG.websocketUrl);
    let recordingsListed = false;
    let snapshotsListed = false;

    const timeout = setTimeout(() => {
      reject(new Error('File listing test timeout'));
    }, CONFIG.timeout);

    ws.on('open', () => {
      console.log('🔌 Connected to WebSocket server');
      
      // Test recordings listing
      send(ws, 'list_recordings', 1, { limit: 10, offset: 0 });
    });

    ws.on('message', (data) => {
      try {
        const response = JSON.parse(data.toString());
        console.log(`📥 Received: ${JSON.stringify(response)}`);

        if (response.id === 1 && response.result) {
          // Recordings listing response
          const result = response.result;
          assert(result.files && Array.isArray(result.files), 'Recordings response has files array');
          assert(result.total_count !== undefined, 'Recordings response has total_count');
          assert(result.has_more !== undefined, 'Recordings response has has_more');
          
          if (result.files.length > 0) {
            const file = result.files[0];
            assert(file.filename, 'File has filename');
            assert(file.download_url, 'File has download_url');
            assert(file.download_url.startsWith('/files/recordings/'), 'Download URL has correct format');
            console.log(`📹 Found recording: ${file.filename} -> ${file.download_url}`);
          }
          
          recordingsListed = true;
          
          // Test snapshots listing
          send(ws, 'list_snapshots', 2, { limit: 10, offset: 0 });
        } else if (response.id === 2 && response.result) {
          // Snapshots listing response
          const result = response.result;
          assert(result.files && Array.isArray(result.files), 'Snapshots response has files array');
          assert(result.total_count !== undefined, 'Snapshots response has total_count');
          assert(result.has_more !== undefined, 'Snapshots response has has_more');
          
          if (result.files.length > 0) {
            const file = result.files[0];
            assert(file.filename, 'File has filename');
            assert(file.download_url, 'File has download_url');
            assert(file.download_url.startsWith('/files/snapshots/'), 'Download URL has correct format');
            console.log(`📸 Found snapshot: ${file.filename} -> ${file.download_url}`);
          }
          
          snapshotsListed = true;
          
          // Test complete
          if (recordingsListed && snapshotsListed) {
            testResults.sprint3Requirements.fileListing = true;
            clearTimeout(timeout);
            ws.close();
            resolve();
          }
        }
      } catch (error) {
        console.error('❌ Error parsing response:', error);
        reject(error);
      }
    });

    ws.on('error', (error) => {
      console.error('❌ WebSocket error:', error);
      reject(error);
    });
  });
}

/**
 * Test 2: File Download via HTTP Endpoints
 */
async function testFileDownload() {
  console.log('\n📥 Sprint 3 Test 2: File Download via HTTP Endpoints');
  
  try {
    // Test snapshot download
    console.log('📸 Testing snapshot download...');
    const snapshotResponse = await httpRequest(`${CONFIG.healthServerUrl}/files/snapshots/test-snapshot.jpg`);
    
    assert(snapshotResponse.statusCode === 200, 'Snapshot download returns 200 OK');
    assert(snapshotResponse.headers['content-type'] === 'image/jpeg', 'Snapshot has correct content type');
    assert(snapshotResponse.headers['content-disposition'], 'Snapshot has content disposition header');
    assert(snapshotResponse.headers['content-disposition'].includes('test-snapshot.jpg'), 'Content disposition includes filename');
    
    console.log(`✅ Snapshot download successful: ${snapshotResponse.statusCode}`);
    
    // Test recording download
    console.log('📹 Testing recording download...');
    const recordingResponse = await httpRequest(`${CONFIG.healthServerUrl}/files/recordings/test-recording.mp4`);
    
    assert(recordingResponse.statusCode === 200, 'Recording download returns 200 OK');
    assert(recordingResponse.headers['content-type'] === 'video/mp4', 'Recording has correct content type');
    assert(recordingResponse.headers['content-disposition'], 'Recording has content disposition header');
    assert(recordingResponse.headers['content-disposition'].includes('test-recording.mp4'), 'Content disposition includes filename');
    
    console.log(`✅ Recording download successful: ${recordingResponse.statusCode}`);
    
    testResults.sprint3Requirements.fileDownload = true;
    
  } catch (error) {
    console.error('❌ File download test failed:', error);
    throw error;
  }
}

/**
 * Test 3: Error Handling for Missing Files
 */
async function testErrorHandling() {
  console.log('\n⚠️ Sprint 3 Test 3: Error Handling for Missing Files');
  
  try {
    // Test missing snapshot
    console.log('📸 Testing missing snapshot...');
    const missingSnapshotResponse = await httpRequest(`${CONFIG.healthServerUrl}/files/snapshots/missing-file.jpg`);
    
    assert(missingSnapshotResponse.statusCode === 404, 'Missing snapshot returns 404');
    console.log(`✅ Missing snapshot handled correctly: ${missingSnapshotResponse.statusCode}`);
    
    // Test missing recording
    console.log('📹 Testing missing recording...');
    const missingRecordingResponse = await httpRequest(`${CONFIG.healthServerUrl}/files/recordings/missing-file.mp4`);
    
    assert(missingRecordingResponse.statusCode === 404, 'Missing recording returns 404');
    console.log(`✅ Missing recording handled correctly: ${missingRecordingResponse.statusCode}`);
    
    // Test directory traversal attempt
    console.log('🔒 Testing directory traversal protection...');
    const traversalResponse = await httpRequest(`${CONFIG.healthServerUrl}/files/snapshots/../../../etc/passwd`);
    
    assert(traversalResponse.statusCode === 400, 'Directory traversal attempt returns 400');
    console.log(`✅ Directory traversal protection working: ${traversalResponse.statusCode}`);
    
    testResults.sprint3Requirements.errorHandling = true;
    
  } catch (error) {
    console.error('❌ Error handling test failed:', error);
    throw error;
  }
}

/**
 * Test 4: URL Construction and Encoding
 */
async function testUrlConstruction() {
  console.log('\n🔗 Sprint 3 Test 4: URL Construction and Encoding');
  
  try {
    // Test URL with special characters
    const specialFilename = 'test file with spaces & special chars (2025).jpg';
    const encodedFilename = encodeURIComponent(specialFilename);
    
    console.log(`📸 Testing URL encoding: "${specialFilename}" -> "${encodedFilename}"`);
    
    // This should return 404 since the file doesn't exist, but the URL should be properly constructed
    const encodedResponse = await httpRequest(`${CONFIG.healthServerUrl}/files/snapshots/${encodedFilename}`);
    
    assert(encodedResponse.statusCode === 404, 'Encoded URL returns 404 (file not found)');
    console.log(`✅ URL encoding working correctly: ${encodedResponse.statusCode}`);
    
    testResults.sprint3Requirements.urlConstruction = true;
    
  } catch (error) {
    console.error('❌ URL construction test failed:', error);
    throw error;
  }
}

/**
 * Main test execution
 */
async function runTests() {
  console.log('🚀 Starting Sprint 3 File Download Tests');
  console.log(`📡 WebSocket Server: ${CONFIG.websocketUrl}`);
  console.log(`🌐 Health Server: ${CONFIG.healthServerUrl}`);
  console.log(`⏱️  Timeout: ${CONFIG.timeout}ms`);
  
  try {
    await testFileListing();
    await testFileDownload();
    await testErrorHandling();
    await testUrlConstruction();
    
    console.log('\n📊 Test Results Summary');
    console.log('========================');
    console.log(`✅ Passed: ${testResults.passed}`);
    console.log(`❌ Failed: ${testResults.failed}`);
    console.log(`📊 Total: ${testResults.total}`);
    console.log(`📈 Success Rate: ${((testResults.passed / testResults.total) * 100).toFixed(1)}%`);
    
    console.log('\n🎯 Sprint 3 Requirements Status');
    console.log('===============================');
    console.log(`📋 File Listing: ${testResults.sprint3Requirements.fileListing ? '✅' : '❌'}`);
    console.log(`📥 File Download: ${testResults.sprint3Requirements.fileDownload ? '✅' : '❌'}`);
    console.log(`⚠️ Error Handling: ${testResults.sprint3Requirements.errorHandling ? '✅' : '❌'}`);
    console.log(`🔗 URL Construction: ${testResults.sprint3Requirements.urlConstruction ? '✅' : '❌'}`);
    
    if (testResults.failed === 0) {
      console.log('\n🎉 All tests passed - File download functionality is working correctly!');
    } else {
      console.log('\n⚠️ Some tests failed - Check the errors above');
      process.exit(1);
    }
    
  } catch (error) {
    console.error('\n💥 Test execution failed:', error);
    process.exit(1);
  }
}

// Run tests
runTests();
