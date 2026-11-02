import http from 'k6/http';
import { check, group } from 'k6';

// Smoke test: validates all endpoints work correctly with minimal load
export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    'http_req_failed': ['rate<0.01'],
    'http_req_duration': ['p(95)<100'],
  },
};

const BASE_URL = __ENV.K6_BASE_URL || 'http://localhost:8080';

export default function () {
  group('API Smoke Tests', () => {
    group('Version Endpoint', () => {
      const res = http.get(`${BASE_URL}/version`);
      check(res, {
        'status is 200': (r) => r.status === 200,
        'returns version info': (r) => r.body.includes('version'),
      });
    });

    group('Monitoring Endpoint', () => {
      const res = http.get(`${BASE_URL}/monitoring`);
      check(res, {
        'status is 200': (r) => r.status === 200,
        'returns JSON': (r) => r.headers['Content-Type'].includes('application/json'),
      });
    });

    group('Photos Endpoint', () => {
      const res = http.get(`${BASE_URL}/photos`);
      check(res, {
        'status is 200': (r) => r.status === 200,
        'returns JSON': (r) => r.headers['Content-Type'].includes('application/json'),
      });
    });

    group('Configuration Endpoint', () => {
      const res = http.get(`${BASE_URL}/configuration`);
      check(res, {
        'status is 200': (r) => r.status === 200,
        'returns JSON': (r) => r.headers['Content-Type'].includes('application/json'),
      });
    });

    group('Logs Endpoint', () => {
      const res = http.get(`${BASE_URL}/logs`);
      check(res, {
        'status is 200': (r) => r.status === 200,
        'returns JSON': (r) => r.headers['Content-Type'].includes('application/json'),
      });
    });

    group('Pprof Endpoint (when enabled)', () => {
      const res = http.get(`${BASE_URL}/debug/pprof/`);
      check(res, {
        'status is 200 when pprof enabled': (r) => r.status === 200 || r.status === 404,
      });
    });
  });
}
