import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export let options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const userId = Math.floor(Math.random() * 20) + 1;

  const res = http.get(
    `${BASE_URL}/users/${userId}/recommendations?limit=10`
  );

  const body = res.json();

  check(res, {
    'status is 200': (r) => r.status === 200,
    'has recommendations': () => body.recommendations?.length > 0,
  });

  sleep(0.1);
}