// Performance testing scenarios using k6
// Install: https://k6.io/docs/getting-started/installation/

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
export let errorRate = new Rate('errors');

// Test configuration
export let options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up
    { duration: '60s', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 100 },  // Ramp to 100 users
    { duration: '60s', target: 100 },  // Stay at 100 users
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<100'], // 95% of requests must complete below 100ms
    http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
    errors: ['rate<0.1'],              // Custom error rate must be below 10%
  },
};

const BASE_URL = 'http://localhost:8080';

export default function() {
  // Test 1: List radios
  let response = http.get(`${BASE_URL}/api/v1/radios`);
  check(response, {
    'list radios status is 200': (r) => r.status === 200,
    'list radios response time < 100ms': (r) => r.timings.duration < 100,
  });
  errorRate.add(response.status !== 200);

  sleep(0.1);

  // Test 2: Set power
  let powerPayload = JSON.stringify({ powerDbm: 10 + Math.random() * 20 });
  let powerHeaders = { 'Content-Type': 'application/json' };
  response = http.post(`${BASE_URL}/api/v1/radios/silvus-001/power`, powerPayload, { headers: powerHeaders });
  check(response, {
    'set power status is 200': (r) => r.status === 200,
    'set power response time < 100ms': (r) => r.timings.duration < 100,
  });
  errorRate.add(response.status !== 200);

  sleep(0.1);

  // Test 3: Set channel
  let channelPayload = JSON.stringify({ channelIndex: 1 + Math.floor(Math.random() * 3) });
  let channelHeaders = { 'Content-Type': 'application/json' };
  response = http.post(`${BASE_URL}/api/v1/radios/silvus-001/channel`, channelPayload, { headers: channelHeaders });
  check(response, {
    'set channel status is 200': (r) => r.status === 200,
    'set channel response time < 100ms': (r) => r.timings.duration < 100,
  });
  errorRate.add(response.status !== 200);

  sleep(0.1);

  // Test 4: Get power state
  response = http.get(`${BASE_URL}/api/v1/radios/silvus-001/power`);
  check(response, {
    'get power status is 200': (r) => r.status === 200,
    'get power response time < 100ms': (r) => r.timings.duration < 100,
  });
  errorRate.add(response.status !== 200);

  sleep(0.1);

  // Test 5: Get channel state
  response = http.get(`${BASE_URL}/api/v1/radios/silvus-001/channel`);
  check(response, {
    'get channel status is 200': (r) => r.status === 200,
    'get channel response time < 100ms': (r) => r.timings.duration < 100,
  });
  errorRate.add(response.status !== 200);

  sleep(0.1);
}

export function handleSummary(data) {
  return {
    'test-results.json': JSON.stringify(data, null, 2),
    stdout: `
🚀 Radio Control Container Performance Test Results
==================================================

📊 Test Summary:
- Total Requests: ${data.metrics.http_reqs.values.count}
- Failed Requests: ${data.metrics.http_req_failed.values.count}
- Error Rate: ${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%

⏱️  Response Times:
- Average: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms
- P95: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms
- P99: ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms

📈 Throughput:
- Requests/sec: ${data.metrics.http_reqs.values.rate.toFixed(2)}

✅ Performance Targets:
- P95 < 100ms: ${data.metrics.http_req_duration.values['p(95)'] < 100 ? 'PASS' : 'FAIL'}
- Error Rate < 10%: ${data.metrics.http_req_failed.values.rate < 0.1 ? 'PASS' : 'FAIL'}

==================================================
    `,
  };
}