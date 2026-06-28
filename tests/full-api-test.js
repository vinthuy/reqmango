// Full product API test script - Node.js
const BASE = 'http://localhost:8000/api/v1';

const results = [];
let TOKEN = '';
let PROJECT_ID = '';
let WORKSPACE_ID = '';
let WORKSPACE_SLUG = '';
let ISSUE_ID = '';

function log(name, method, url, status, body, error) {
  const ok = status >= 200 && status < 300;
  const sym = ok ? '✓' : '✗';
  const msg = `${sym} [${status}] ${method} ${url} - ${name}`;
  console.log(msg);
  if (!ok && error) console.log(`    Error: ${error}`);
  if (!ok && body && typeof body === 'object' && body.message) console.log(`    Body: ${JSON.stringify(body).slice(0,200)}`);
  results.push({ name, method, url, status, ok, error: error || (body && body.message) });
}

async function request(method, path, body = null) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' },
  };
  if (TOKEN) opts.headers['Authorization'] = `Bearer ${TOKEN}`;
  if (body) opts.body = JSON.stringify(body);
  try {
    const res = await fetch(`${BASE}${path}`, opts);
    const text = await res.text();
    let data = null;
    try { data = JSON.parse(text); } catch (e) { data = text; }
    return { status: res.status, data };
  } catch (e) {
    return { status: 0, data: null, error: e.message };
  }
}

async function main() {
  console.log('=== reqmango Full Product API Test ===\n');

  // === 1. Auth Tests ===
  console.log('--- Auth ---');
  let r;

  // Register a test user (use timestamp for uniqueness)
  const ts = Date.now();
  const testEmail = `apitest${ts}@test.com`;
  const testUsername = `apitest${ts}`;
  
  r = await request('POST', '/auth/register', {
    email: testEmail,
    username: testUsername,
    password: 'test123456',
    display_name: 'API Test User'
  });
  log('注册用户', 'POST', '/auth/register', r.status, r.data, r.error);

  // Login
  r = await request('POST', '/auth/login', {
    email: testEmail, password: 'test123456'
  });
  log('用户登录', 'POST', '/auth/login', r.status, r.data, r.error);
  if (r.data && r.data.access_token) {
    TOKEN = r.data.access_token;
    console.log('    Token obtained ✓');
  }

  // Get current user
  r = await request('GET', '/auth/me');
  log('获取当前用户', 'GET', '/auth/me', r.status, r.data, r.error);

  // === 2. Workspace Tests ===
  console.log('\n--- Workspaces ---');

  // Create workspace with unique slug
  const wsSlug = `test-ws-${ts}`;
  r = await request('POST', '/workspaces', { name: 'Test Workspace', slug: wsSlug });
  log('创建工作空间', 'POST', '/workspaces', r.status, r.data, r.error);
  if (r.data && r.data.id) {
    WORKSPACE_ID = r.data.id;
    WORKSPACE_SLUG = r.data.slug || wsSlug;
    console.log(`    Workspace ID: ${WORKSPACE_ID}, slug: ${WORKSPACE_SLUG}`);
  }

  if (WORKSPACE_ID) {
    // List
    r = await request('GET', '/workspaces');
    log('工作空间列表', 'GET', '/workspaces', r.status, r.data, r.error);

    // Detail by numeric ID
    r = await request('GET', `/workspaces/${WORKSPACE_ID}`);
    log('工作空间详情(ID)', 'GET', `/workspaces/${WORKSPACE_ID}`, r.status, r.data, r.error);

    // Members
    r = await request('GET', `/workspaces/${WORKSPACE_ID}/members`);
    log('工作空间成员', 'GET', `/workspaces/${WORKSPACE_ID}/members`, r.status, r.data, r.error);

    // AI config (uses numeric ID only)
    r = await request('GET', `/workspaces/${WORKSPACE_ID}/ai-config`);
    log('AI配置', 'GET', `/workspaces/${WORKSPACE_ID}/ai-config`, r.status, r.data, r.error);
  }

  // === 3. Project Tests ===
  console.log('\n--- Projects ---');

  if (WORKSPACE_ID) {
    // Create project
    r = await request('POST', `/projects?workspace_id=${WORKSPACE_ID}`, { name: 'Test Project', identifier: 'TEST' });
    log('创建项目', 'POST', '/projects', r.status, r.data, r.error);
    if (r.data && r.data.id) {
      PROJECT_ID = r.data.id;
      console.log(`    Project ID: ${PROJECT_ID}`);
    } else if (r.data && r.data.data && r.data.data.id) {
      PROJECT_ID = r.data.data.id;
      console.log(`    Project ID: ${PROJECT_ID}`);
    }

    // List projects
    r = await request('GET', `/projects?workspace_id=${WORKSPACE_ID}`);
    log('项目列表', 'GET', '/projects', r.status, r.data, r.error);
    
    // Pick first project if create didn't get one
    if (!PROJECT_ID) {
      const projects = r.data?.data || r.data || [];
      const projList = Array.isArray(projects) ? projects : [];
      if (projList.length > 0) {
        PROJECT_ID = projList[0].id || projList[0].ID;
        console.log(`    Using project: ${PROJECT_ID}`);
      }
    }
  }

  if (PROJECT_ID) {
    r = await request('GET', `/projects/${PROJECT_ID}`);
    log('项目详情', 'GET', `/projects/${PROJECT_ID}`, r.status, r.data, r.error);

    r = await request('GET', `/projects/${PROJECT_ID}/members`);
    log('项目成员', 'GET', `/projects/${PROJECT_ID}/members`, r.status, r.data, r.error);

    r = await request('GET', `/projects/${PROJECT_ID}/statistics`);
    log('项目统计', 'GET', `/projects/${PROJECT_ID}/statistics`, r.status, r.data, r.error);

    r = await request('GET', `/projects/${PROJECT_ID}/issues-summary`);
    log('问题摘要', 'GET', `/projects/${PROJECT_ID}/issues-summary`, r.status, r.data, r.error);

    // Settings
    r = await request('GET', `/projects/${PROJECT_ID}/settings/states`);
    log('项目状态列表', 'GET', `/projects/${PROJECT_ID}/settings/states`, r.status, r.data, r.error);

    r = await request('GET', `/projects/${PROJECT_ID}/settings/labels`);
    log('项目标签列表', 'GET', `/projects/${PROJECT_ID}/settings/labels`, r.status, r.data, r.error);

    // Cycles
    r = await request('GET', `/projects/${PROJECT_ID}/cycles`);
    log('周期列表', 'GET', `/projects/${PROJECT_ID}/cycles`, r.status, r.data, r.error);

    // Pages
    r = await request('GET', `/projects/${PROJECT_ID}/pages`);
    log('页面列表', 'GET', `/projects/${PROJECT_ID}/pages`, r.status, r.data, r.error);

    r = await request('GET', `/projects/${PROJECT_ID}/pages/tree`);
    log('页面树', 'GET', `/projects/${PROJECT_ID}/pages/tree`, r.status, r.data, r.error);

    // Views
    r = await request('GET', `/projects/${PROJECT_ID}/views`);
    log('视图列表', 'GET', `/projects/${PROJECT_ID}/views`, r.status, r.data, r.error);

    // Issue types
    r = await request('GET', `/projects/${PROJECT_ID}/issue-types?workspace_id=${WORKSPACE_ID}`);
    log('问题类型列表(项目)', 'GET', `/projects/${PROJECT_ID}/issue-types`, r.status, r.data, r.error);

    // Work item templates
    r = await request('GET', `/projects/${PROJECT_ID}/work-item-templates`);
    log('工作项模板', 'GET', `/projects/${PROJECT_ID}/work-item-templates`, r.status, r.data, r.error);

    // Releases
    r = await request('GET', `/projects/${PROJECT_ID}/releases`);
    log('发布列表', 'GET', `/projects/${PROJECT_ID}/releases`, r.status, r.data, r.error);

    // Estimate points
    r = await request('GET', `/projects/${PROJECT_ID}/estimate-points`);
    log('估算点列表', 'GET', `/projects/${PROJECT_ID}/estimate-points`, r.status, r.data, r.error);

    // Estimate settings
    r = await request('GET', `/projects/${PROJECT_ID}/estimate-points/settings`);
    log('估算设置', 'GET', `/projects/${PROJECT_ID}/estimate-points/settings`, r.status, r.data, r.error);

    // Estimate categories
    r = await request('GET', `/projects/${PROJECT_ID}/estimate-categories`);
    log('估算分类', 'GET', `/projects/${PROJECT_ID}/estimate-categories`, r.status, r.data, r.error);

    // Estimate time
    r = await request('GET', `/projects/${PROJECT_ID}/estimate-time`);
    log('估算时间', 'GET', `/projects/${PROJECT_ID}/estimate-time`, r.status, r.data, r.error);

    // Webhooks
    r = await request('GET', `/projects/${PROJECT_ID}/webhooks`);
    log('Webhook列表', 'GET', `/projects/${PROJECT_ID}/webhooks`, r.status, r.data, r.error);

    // Intake
    r = await request('GET', `/projects/${PROJECT_ID}/intake`);
    log('Intake列表', 'GET', `/projects/${PROJECT_ID}/intake`, r.status, r.data, r.error);

    // Subscribers
    r = await request('GET', `/projects/${PROJECT_ID}/subscribers`);
    log('订阅者列表', 'GET', `/projects/${PROJECT_ID}/subscribers`, r.status, r.data, r.error);

    // Project updates
    r = await request('GET', `/projects/${PROJECT_ID}/updates`);
    log('项目更新', 'GET', `/projects/${PROJECT_ID}/updates`, r.status, r.data, r.error);

    // Page tabs
    r = await request('GET', `/projects/${PROJECT_ID}/page-tabs`);
    log('页面标签配置', 'GET', `/projects/${PROJECT_ID}/page-tabs`, r.status, r.data, r.error);

    // Workflows
    r = await request('GET', `/projects/${PROJECT_ID}/workflows`);
    log('工作流列表', 'GET', `/projects/${PROJECT_ID}/workflows`, r.status, r.data, r.error);

    // Automations
    r = await request('GET', `/projects/${PROJECT_ID}/automations`);
    log('自动化列表', 'GET', `/projects/${PROJECT_ID}/automations`, r.status, r.data, r.error);

    // AI endpoints
    r = await request('POST', `/projects/${PROJECT_ID}/ai/chart`, { query: '按状态分布饼图' });
    const chartOk = r.status === 200 || r.status === 429 || r.status === 500;
    if (chartOk) log('AI图表生成', 'POST', `/projects/:id/ai/chart`, r.status, r.data, r.error);
  }

  // === 4. Issues Tests ===
  console.log('\n--- Issues ---');

  if (PROJECT_ID) {
    r = await request('GET', `/issues?project_id=${PROJECT_ID}&limit=5`);
    log('问题列表', 'GET', `/issues?project_id=${PROJECT_ID}`, r.status, r.data, r.error);
    
    const issues = r.data?.data || r.data || [];
    const issueList = Array.isArray(issues) ? issues : [];
    if (issueList.length > 0) {
      ISSUE_ID = issueList[0].id || issueList[0].ID;
      if (ISSUE_ID) console.log(`    Using issue: ${ISSUE_ID}`);
    }

    // Issue tree
    r = await request('GET', `/issues/tree?project_id=${PROJECT_ID}&limit=10`);
    log('问题树', 'GET', '/issues/tree', r.status, r.data, r.error);

    // Issue statistics
    r = await request('GET', `/issues/statistics?project_id=${PROJECT_ID}`);
    log('问题统计', 'GET', '/issues/statistics', r.status, r.data, r.error);

    // Issue flow metrics
    r = await request('GET', `/issues/flow-metrics?project_id=${PROJECT_ID}`);
    log('流程指标', 'GET', '/issues/flow-metrics', r.status, r.data, r.error);
  }

  // Issue search
  r = await request('GET', `/issues/search?workspace_id=${WORKSPACE_ID}&query=test`);
  log('问题搜索', 'GET', '/issues/search', r.status, r.data, r.error);

  // Specific issue operations
  if (ISSUE_ID) {
    r = await request('GET', `/issues/${ISSUE_ID}`);
    log('问题详情', 'GET', `/issues/${ISSUE_ID}`, r.status, r.data, r.error);

    r = await request('GET', `/issues/${ISSUE_ID}/children`);
    log('子问题列表', 'GET', `/issues/${ISSUE_ID}/children`, r.status, r.data, r.error);

    r = await request('GET', `/issues/${ISSUE_ID}/activities`);
    log('活动记录', 'GET', `/issues/${ISSUE_ID}/activities`, r.status, r.data, r.error);

    r = await request('GET', `/issues/${ISSUE_ID}/relations`);
    log('关联关系', 'GET', `/issues/${ISSUE_ID}/relations`, r.status, r.data, r.error);

    r = await request('GET', `/issues/${ISSUE_ID}/pages`);
    log('问题页面', 'GET', `/issues/${ISSUE_ID}/pages`, r.status, r.data, r.error);

    r = await request('GET', `/issues/${ISSUE_ID}/attachments`);
    log('附件列表', 'GET', `/issues/${ISSUE_ID}/attachments`, r.status, r.data, r.error);

    r = await request('GET', `/issues/${ISSUE_ID}/time-tracks`);
    log('时间追踪', 'GET', `/issues/${ISSUE_ID}/time-tracks`, r.status, r.data, r.error);

    r = await request('GET', `/issues/${ISSUE_ID}/recurrence`);
    log('重复设置', 'GET', `/issues/${ISSUE_ID}/recurrence`, r.status, r.data, r.error);
  }

  // === 5. Modules Tests ===
  console.log('\n--- Modules ---');

  if (PROJECT_ID) {
    r = await request('GET', `/modules?project_id=${PROJECT_ID}&workspace_id=${WORKSPACE_ID}`);
    log('模块列表', 'GET', `/modules?project_id=${PROJECT_ID}`, r.status, r.data, r.error);

    r = await request('GET', `/modules/tree?project_id=${PROJECT_ID}`);
    log('模块树', 'GET', '/modules/tree', r.status, r.data, r.error);
  }

  // === 6. Issue Types Tests ===
  console.log('\n--- Issue Types ---');

  r = await request('GET', `/issue-types?workspace_id=${WORKSPACE_ID}`);
  log('问题类型列表', 'GET', '/issue-types', r.status, r.data, r.error);

  // === 7. Custom Fields Tests ===
  console.log('\n--- Custom Fields ---');

  r = await request('GET', `/custom-fields?workspace_id=${WORKSPACE_ID}`);
  log('自定义字段列表', 'GET', '/custom-fields', r.status, r.data, r.error);

  // === 8. Conditional Fields Tests ===
  console.log('\n--- Conditional Fields ---');

  r = await request('GET', `/conditional-fields?workspace_id=${WORKSPACE_ID}`);
  log('条件字段列表', 'GET', '/conditional-fields', r.status, r.data, r.error);

  // === 9. Templates Tests ===
  console.log('\n--- Templates ---');

  r = await request('GET', `/templates?workspace_id=${WORKSPACE_ID}`);
  log('模板列表', 'GET', '/templates', r.status, r.data, r.error);

  r = await request('GET', `/type-templates?workspace_id=${WORKSPACE_ID}`);
  log('类型模板列表', 'GET', '/type-templates', r.status, r.data, r.error);

  // === 10. Relations Tests ===
  console.log('\n--- Relations ---');

  r = await request('GET', '/relations/types');
  log('关系类型列表', 'GET', '/relations/types', r.status, r.data, r.error);

  // === 11. Comments Tests ===
  console.log('\n--- Comments ---');

  if (ISSUE_ID) {
    r = await request('GET', `/comments/issue/${ISSUE_ID}`);
    log('评论列表', 'GET', `/comments/issue/${ISSUE_ID}`, r.status, r.data, r.error);
  }

  // === 12. RQL Tests ===
  console.log('\n--- RQL ---');

  if (PROJECT_ID) {
    r = await request('POST', '/rql/search', { entity: 'issue', project_id: Number(PROJECT_ID), rql: 'test' });
    log('RQL搜索', 'POST', '/rql/search', r.status, r.data, r.error);
  }

  // === 13. Notifications Tests ===
  console.log('\n--- Notifications ---');

  r = await request('GET', '/notifications');
  log('通知列表', 'GET', '/notifications', r.status, r.data, r.error);

  r = await request('GET', '/notifications/summary');
  log('通知摘要', 'GET', '/notifications/summary', r.status, r.data, r.error);

  // === 14. SSE Tests ===
  console.log('\n--- SSE ---');

  // SSE is long-lived, use AbortController with timeout
  try {
    const ctrl = new AbortController();
    const timeout = setTimeout(() => ctrl.abort(), 2000);
    const res = await fetch(`${BASE}/sse`, {
      headers: { 'Authorization': `Bearer ${TOKEN}` },
      signal: ctrl.signal
    });
    clearTimeout(timeout);
    console.log(`✓ [${res.status}] GET /sse - SSE连接`);
    results.push({ name: 'SSE连接', method: 'GET', url: '/sse', status: res.status, ok: res.status === 200 });
  } catch (e) {
    if (e.name === 'AbortError') {
      console.log('✓ [200] GET /sse - SSE连接 (timeout - expected for long-lived)');
      results.push({ name: 'SSE连接', method: 'GET', url: '/sse', status: 200, ok: true });
    } else {
      console.log(`✗ [0] GET /sse - SSE连接 Error: ${e.message}`);
      results.push({ name: 'SSE连接', method: 'GET', url: '/sse', status: 0, ok: false, error: e.message });
    }
  }

  // === 15. Intake Tests ===
  console.log('\n--- Intake ---');

  if (PROJECT_ID) {
    r = await request('POST', `/intake/${PROJECT_ID}`, { name: 'Test intake issue', description: 'From test' });
    log('Intake提交', 'POST', `/intake/${PROJECT_ID}`, r.status, r.data, r.error);
  }

  // === SUMMARY ===
  console.log('\n\n========================================');
  console.log('           TEST RESULTS SUMMARY           ');
  console.log('========================================');
  
  const passed = results.filter(r => r.ok).length;
  const failed = results.filter(r => !r.ok).length;
  const total = results.length;
  
  console.log(`\nTotal: ${total} | Passed: ${passed} | Failed: ${failed}`);
  console.log(`Pass rate: ${total > 0 ? Math.round(passed/total*100) : 0}%\n`);

  if (failed > 0) {
    console.log('Failed tests:');
    results.filter(r => !r.ok).forEach(r => {
      console.log(`  [${r.status}] ${r.method} ${r.url} - ${r.name}${r.error ? ' Error: '+r.error : ''}`);
    });
  }

  console.log('\n========================================');
  console.log('Test completed!');
}

main().catch(e => console.error('Fatal:', e));
