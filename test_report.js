const BASE = 'http://localhost:8000/api/v1';

async function main() {
  const login = await fetch(BASE + '/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: 'admin@reqmango.com', password: 'demo1234' })
  });
  const { access_token: t } = await login.json();
  const h = { 'Authorization': 'Bearer ' + t, 'Content-Type': 'application/json' };

  const tests = ['created_vs_resolved', 'avg_age', 'current_age', 'created_trend'];
  for (const type of tests) {
    const r = await fetch(BASE + '/projects/1/reports', {
      method: 'POST',
      headers: h,
      body: JSON.stringify({ report_type: type, interval: type === 'created_trend' ? 'week' : undefined })
    });
    const body = await r.text();
    console.log(`${type}: ${r.status} - ${body.substring(0, 300)}`);
  }
}

main().catch(e => console.log('Error:', e.message));
