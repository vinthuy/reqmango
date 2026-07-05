const http = require('http');

function req(method, path, body, token) {
  return new Promise((resolve, reject) => {
    const opts = { hostname: 'localhost', port: 8000, path, method, headers: { 'Content-Type': 'application/json' } };
    if (token) opts.headers.Authorization = 'Bearer ' + token;
    const r = http.request(opts, res => {
      let d = '';
      res.on('data', c => d += c);
      res.on('end', () => {
        try { resolve({ status: res.statusCode, data: JSON.parse(d) }); }
        catch(e) { resolve({ status: res.statusCode, data: d }); }
      });
    });
    r.on('error', reject);
    if (body) r.write(JSON.stringify(body));
    r.end();
  });
}

(async () => {
  const ts = Date.now();

  // Register + Login
  await req('POST', '/api/v1/auth/register', { email: 'rpt' + ts + '@t.com', username: 'rpt' + ts, password: 'E2eTest123!', display_name: 'test' });
  const login = await req('POST', '/api/v1/auth/login', { email: 'rpt' + ts + '@t.com', password: 'E2eTest123!' });
  const token = login.data.access_token;
  console.log('Token OK');

  // Create workspace + project
  const ws = await req('POST', '/api/v1/workspaces', { name: 'RPT WS', slug: 'rpt-' + ts }, token);
  const wsId = ws.data.id;
  const proj = await req('POST', '/api/v1/projects?workspace_id=' + wsId, { name: 'RPT Proj', identifier: 'RPT', description: 'test' }, token);
  const pid = proj.data.id;
  console.log('Project ID:', pid);

  // Test 1: Distribution by state (no RQL)
  const t1 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar' }, token);
  console.log('\n1. Distribution by state:', t1.status, 'total:', t1.data.total);

  // Test 2: RQL = operator
  const t2 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'state = "Todo"' }, token);
  console.log('2. RQL =:', t2.status, t2.data.total !== undefined ? 'total:' + t2.data.total : JSON.stringify(t2.data));

  // Test 3: RQL != operator
  const t3 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'state != "Done"' }, token);
  console.log('3. RQL !=:', t3.status, t3.data.total !== undefined ? 'total:' + t3.data.total : JSON.stringify(t3.data));

  // Test 4: RQL IN operator
  const t4 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'priority', chart: 'pie', rql: 'priority IN ["urgent", "high"]' }, token);
  console.log('4. RQL IN:', t4.status, t4.data.total !== undefined ? 'total:' + t4.data.total : JSON.stringify(t4.data));

  // Test 5: RQL NOT IN operator
  const t5 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'state NOT IN ["Done"]' }, token);
  console.log('5. RQL NOT IN:', t5.status, t5.data.total !== undefined ? 'total:' + t5.data.total : JSON.stringify(t5.data));

  // Test 6: RQL LIKE operator
  const t6 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'name LIKE "test"' }, token);
  console.log('6. RQL LIKE:', t6.status, t6.data.total !== undefined ? 'total:' + t6.data.total : JSON.stringify(t6.data));

  // Test 7: RQL NOT LIKE operator
  const t7 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'name NOT LIKE "debug"' }, token);
  console.log('7. RQL NOT LIKE:', t7.status, t7.data.total !== undefined ? 'total:' + t7.data.total : JSON.stringify(t7.data));

  // Test 8: RQL >= operator
  const t8 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'start_date >= "2024-01-01"' }, token);
  console.log('8. RQL >=:', t8.status, t8.data.total !== undefined ? 'total:' + t8.data.total : JSON.stringify(t8.data));

  // Test 9: RQL <= operator
  const t9 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'target_date <= "2024-12-31"' }, token);
  console.log('9. RQL <=:', t9.status, t9.data.total !== undefined ? 'total:' + t9.data.total : JSON.stringify(t9.data));

  // Test 10: RQL IS NULL
  const t10 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'assignee IS NULL' }, token);
  console.log('10. RQL IS NULL:', t10.status, t10.data.total !== undefined ? 'total:' + t10.data.total : JSON.stringify(t10.data));

  // Test 11: RQL IS NOT NULL
  const t11 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'assignee IS NOT NULL' }, token);
  console.log('11. RQL IS NOT NULL:', t11.status, t11.data.total !== undefined ? 'total:' + t11.data.total : JSON.stringify(t11.data));

  // Test 12: RQL combined AND
  const t12 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'priority IN ["urgent"] AND state != "Done"' }, token);
  console.log('12. RQL AND:', t12.status, t12.data.total !== undefined ? 'total:' + t12.data.total : JSON.stringify(t12.data));

  // Test 13: Created trend
  const t13 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'created_trend', group_by: 'state', chart: 'line', interval: 'day' }, token);
  console.log('13. Created trend:', t13.status, t13.data.type);

  // Test 14: Created vs resolved
  const t14 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'created_vs_resolved', chart: 'bar', interval: 'week' }, token);
  console.log('14. Created vs resolved:', t14.status, t14.data.values2 !== undefined ? 'ok' : JSON.stringify(t14.data));

  // Test 15: Avg age
  const t15 = await req('POST', '/api/v1/projects/' + pid + '/reports', { report_type: 'avg_age', group_by: 'state', chart: 'bar' }, token);
  console.log('15. Avg age:', t15.status, t15.data.summary ? 'ok' : JSON.stringify(t15.data));

  // Test 16: Saved report CRUD
  const cr = await req('POST', '/api/v1/projects/' + pid + '/saved-reports', { name: 'Test Report', report_type: 'distribution', group_by: 'state', chart_type: 'bar', rql: 'priority = "high"' }, token);
  console.log('\n16. Create saved:', cr.status, cr.data.name);

  const lr = await req('GET', '/api/v1/projects/' + pid + '/saved-reports', null, token);
  console.log('17. List saved:', lr.status, 'count:', lr.data.length);

  const ur = await req('PATCH', '/api/v1/projects/' + pid + '/saved-reports/' + cr.data.id, { name: 'Updated' }, token);
  console.log('18. Update saved:', ur.status, ur.data.name);

  const dr = await req('DELETE', '/api/v1/projects/' + pid + '/saved-reports/' + cr.data.id, null, token);
  console.log('19. Delete saved:', dr.status);

  console.log('\n=== ALL TESTS DONE ===');
})().catch(e => console.error('FATAL:', e.message));
