import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users over 30s
    { duration: '1m', target: 10 },   // Stay at 10 users for 1 minute
    { duration: '30s', target: 50 },  // Ramp up to 50 users over 30s
    { duration: '2m', target: 50 },   // Stay at 50 users for 2 minutes
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    // Success criteria: 100% success rate (2xx responses)
    'http_req_failed': ['rate<0.01'], // Less than 1% failed requests
    // Success criteria: p95 < 100ms
    'http_req_duration': ['p(95)<100'],
    // Additional thresholds for monitoring
    'http_req_duration': ['p(50)<50', 'p(99)<200'],
    'errors': ['rate<0.01'],
  },
};

// Base URL - can be overridden with K6_BASE_URL environment variable
const BASE_URL = __ENV.K6_BASE_URL || 'http://localhost:8080';

export default function () {
  // Test read-only endpoints
  const endpoints = [
    '/version',
    '/monitoring',
    '/photos',
    '/configuration',
    '/logs',
  ];

  // Randomly select an endpoint to test (simulates real-world usage patterns)
  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  
  const response = http.get(`${BASE_URL}${endpoint}`);
  
  // Check that the request succeeded
  const success = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time OK': (r) => r.timings.duration < 100,
  });

  // Track error rate
  errorRate.add(!success);

  // Small sleep to simulate think time between requests
  sleep(1);
}
