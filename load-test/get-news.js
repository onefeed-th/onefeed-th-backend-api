import http from 'k6/http';

export const options = {
  vus: 800,
  duration: '5m',
  thresholds: {
    http_req_failed: ['rate<0.01'], // error rate < 1% (=> success > 99%)
    http_req_duration: ['p(95)<5000'], // optional: 95% request < 5sec
  },
}

export default function () {
  const url = 'https://onefeed-th-api.artzakub.com/api/v1/news';
  const payload = JSON.stringify({
    source: [
      "MacThai",
      "DroidSans",
      "เกมถูกบอกด้วย"
    ],
    page: 1,
    limit: 20
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  http.post(url, payload, params);

}