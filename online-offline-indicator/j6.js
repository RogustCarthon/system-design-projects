import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
  vus: 10,
  duration: '30s',
};

// GET "/status/:userId"
export function getEndpoint() {
  let userId = Math.floor(Math.random() * 100) + 1; // Generate a random user ID between 1 and 100
  let res = http.get(`http://localhost:8080/status/${userId}`);
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response body is not empty': (r) => r.body.length > 0,
  });
  sleep(1);
}

// POST "/status/:userId"
export function postEndpoint() {
  let userId = Math.floor(Math.random() * 100) + 1; // Generate a random user ID between 1 and 100
  let res = http.post(`http://localhost:8080/status/${userId}`, { status: 'online' });
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response body is not empty': (r) => r.body.length > 0,
  });
  sleep(1);
}

