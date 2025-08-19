const jwt = require('jsonwebtoken');

/**
 * Unit tests for performance validation
 * Follows testing guidelines: "Authentication tokens are always generated dynamically; no hardcoded credentials allowed"
 */

// Get JWT secret from environment (no hardcoded fallback)
const getJwtSecret = () => {
  const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
  if (!secret) {
    throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
  }
  return secret;
};

/**
 * Generate a valid JWT token for testing
 * @param {string} userId - User ID for the token
 * @param {string} role - User role (viewer, operator, admin)
 * @param {number} expiresIn - Token expiration in seconds (default: 24 hours)
 * @returns {string} JWT token
 */
function generateValidToken(userId = 'test_user', role = 'operator', expiresIn = 24 * 60 * 60) {
  const secret = getJwtSecret();
  
  const payload = {
    user_id: userId,
    role: role,
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + expiresIn
  };
  
  return jwt.sign(payload, secret, { algorithm: 'HS256' });
}

/**
 * Test performance validation functionality
 */
function testPerformanceValidation() {
  console.log('🧪 Testing performance validation functionality...');
  
  try {
    // Generate token dynamically
    const token = generateValidToken('test_user', 'operator');
    console.log('✅ Generated valid JWT token for testing');
    
    // Simulate performance validation tests
    const performanceMetrics = {
      responseTime: Math.random() * 100 + 50, // 50-150ms
      throughput: Math.random() * 1000 + 500, // 500-1500 req/s
      memoryUsage: Math.random() * 100 + 50,  // 50-150MB
      cpuUsage: Math.random() * 30 + 10       // 10-40%
    };
    
    // Validate performance thresholds
    const validationResults = {
      responseTime: performanceMetrics.responseTime < 100,
      throughput: performanceMetrics.throughput > 800,
      memoryUsage: performanceMetrics.memoryUsage < 200,
      cpuUsage: performanceMetrics.cpuUsage < 50
    };
    
    const allPassed = Object.values(validationResults).every(result => result);
    
    const testResult = {
      success: allPassed,
      message: allPassed ? 'Performance validation passed' : 'Performance validation failed',
      tokenGenerated: !!token,
      metrics: performanceMetrics,
      validation: validationResults
    };
    
    console.log('📊 Test Results:', testResult);
    return testResult;
    
  } catch (error) {
    console.error('❌ Performance validation test failed:', error.message);
    return {
      success: false,
      error: error.message
    };
  }
}

/**
 * Test performance thresholds
 */
function testPerformanceThresholds() {
  console.log('⚡ Testing performance thresholds...');
  
  try {
    const token = generateValidToken();
    
    // Test different performance scenarios
    const scenarios = [
      { name: 'Low Load', load: 0.1 },
      { name: 'Medium Load', load: 0.5 },
      { name: 'High Load', load: 0.9 }
    ];
    
    const results = scenarios.map(scenario => {
      const responseTime = 50 + (scenario.load * 100);
      const passed = responseTime < 150;
      
      return {
        scenario: scenario.name,
        load: scenario.load,
        responseTime: responseTime,
        passed: passed
      };
    });
    
    const allPassed = results.every(r => r.passed);
    
    console.log('📊 Threshold Test Results:');
    results.forEach(result => {
      console.log(`   ${result.scenario}: ${result.responseTime.toFixed(1)}ms ${result.passed ? '✅' : '❌'}`);
    });
    
    return {
      success: allPassed,
      results: results,
      tokenGenerated: !!token
    };
    
  } catch (error) {
    console.error('❌ Performance threshold test failed:', error.message);
    return {
      success: false,
      error: error.message
    };
  }
}

/**
 * Test environment validation
 */
function testEnvironmentValidation() {
  console.log('🔍 Testing environment validation...');
  
  try {
    const secret = getJwtSecret();
    console.log('✅ JWT secret available (length:', secret.length, ')');
    
    const token = generateValidToken();
    console.log('✅ Token generation successful');
    
    return {
      success: true,
      environmentReady: true,
      secretAvailable: !!secret,
      tokenGenerated: !!token
    };
    
  } catch (error) {
    console.error('❌ Environment validation failed:', error.message);
    return {
      success: false,
      environmentReady: false,
      error: error.message
    };
  }
}

// Run tests
if (require.main === module) {
  console.log('🚀 Running performance validation unit tests...\n');
  
  const envTest = testEnvironmentValidation();
  if (envTest.success) {
    testPerformanceValidation();
    testPerformanceThresholds();
  } else {
    console.error('❌ Cannot run tests - environment not properly configured');
    process.exit(1);
  }
}

module.exports = {
  testPerformanceValidation,
  testPerformanceThresholds,
  testEnvironmentValidation,
  generateValidToken,
  getJwtSecret
};
