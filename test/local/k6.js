import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: 30,
  duration: '10s',
  thresholds: {
    http_req_failed: ['rate==0.0']
  }
};

export default function() {
  const hosts = [
    {domainName: 'foo.domain.lb', functionName: 'golden-goose'},
    {domainName: 'bar.domain.lb', functionName: 'green-falcon'}
  ]
  const host = hosts[(new Date().getTime()) % 2]
  const res = http.get('http://localhost:41414/whatever', {
    headers: {
      'Host': host.domainName
    }
  });
  check(res, {
    'expected status 200': (r) => r.status === 200,
    'expected reply from selected lambda': (r) => {
      const data = JSON.parse(r.body)
      return data.env.FUNCTION_NAME == host.functionName
    }
  });
}
