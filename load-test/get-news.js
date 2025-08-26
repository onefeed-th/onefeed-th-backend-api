import http from 'k6/http';
import { check } from 'k6';


export const options = {
  vus: 1000,
  duration: '5m',
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

  const res = http.post(url, payload, params);

  check(res, {
    'is status 200': (r) => r.status === 200,
  });
}