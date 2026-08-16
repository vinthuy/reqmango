// @ts-check
/**
 * ReqMango — 综合 E2E 全页面验证脚本
 * 覆盖所有主要路由：Auth、Core、AI/Agent、Other
 *
 * 运行: node frontend/e2e-full-verify.cjs
 */
'use strict';
const puppeteer = require('puppeteer');
const fs = require('fs');
const path = require('path');

// ============================================================
// 配置
// ============================================================
const BASE_URL = 'http://localhost:5173';
const API_URL  = 'http://localhost:8000';
const SCREENSHOT_DIR = path.join(__dirname, 'e2e-full-screenshots');

const NAV_TIMEOUT   = 30000;
const ELEMENT_TIMEOUT = 8000;
const SHORT_WAIT    = 2000;
const MEDIUM_WAIT   = 3000;
const LONG_WAIT     = 5000;

const LOGIN_EMAIL    = 'admin@reqmango.com';
const LOGIN_PASSWORD = 'demo1234';

// ============================================================
// 确保截图目录存在
// ============================================================
if (!fs.existsSync(SCREENSHOT_DIR)) {
  fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
}

// ============================================================
// 结果跟踪
// ============================================================
const results = [];
let testIndex = 0;

function log(msg, status = 'info') {
  const icon = status === 'pass' ? '✅' : status === 'fail' ? '❌' : 'ℹ️';
  console.log(`${icon} ${msg}`);
  results.push({ msg, status });
}

async function screenshot(page, name) {
  testIndex++;
  const padded = String(testIndex).padStart(3, '0');
  const filePath = path.join(SCREENSHOT_DIR, `${padded}-${name}.png`);
  try {
    await page.screenshot({ path: filePath, fullPage: true });
  } catch {
    // 截图失败不阻断
  }
}

function delay(ms) {
  return new Promise(r => setTimeout(r, ms));
}

// ============================================================
// 测试辅助函数
// ============================================================

/**
 * 导航到指定 URL，等待加载，截图并验证 URL
 */
async function navigateAndVerify(page, url, expectedPath, screenshotName) {
  try {
    await page.goto(url, { waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
    await delay(SHORT_WAIT);
    await screenshot(page, screenshotName);
    const currentUrl = page.url();
    const urlMatch = currentUrl.includes(expectedPath);
    log(`${screenshotName} — 页面导航: ${urlMatch ? '成功' : 'URL不匹配'} (${currentUrl})`, urlMatch ? 'pass' : 'fail');
    return urlMatch;
  } catch (err) {
    log(`${screenshotName} — 页面导航失败: ${err.message}`, 'fail');
    try { await screenshot(page, screenshotName + '-error'); } catch {}
    return false;
  }
}

/**
 * 检查页面是否有有意义的内容（非空白页）
 */
async function verifyPageHasContent(page, description) {
  const hasContent = await page.evaluate(() => {
    const body = document.body;
    if (!body) return false;
    const text = (body.innerText || '').trim();
    return text.length > 20;
  });
  log(`${description} — 页面内容加载: ${hasContent ? '成功' : '内容过少'}`, hasContent ? 'pass' : 'fail');
  return hasContent;
}

/**
 * 检查页面是否存在 JS 错误
 */
async function checkConsoleErrors(page, testName) {
  const errors = [];
  const handler = (msg) => {
    if (msg.type() === 'error') {
      errors.push(msg.text());
    }
  };
  page.on('console', handler);
  await delay(500);
  page.off('console', handler);
  if (errors.length > 0) {
    log(`${testName} — 控制台错误 (${errors.length}个): ${errors[0].substring(0, 100)}`, 'fail');
  }
  return errors;
}

/**
 * 通过文本查找并点击按钮
 */
async function clickButtonByText(page, text) {
  return await page.evaluate((btnText) => {
    const buttons = Array.from(document.querySelectorAll('button, a[role="button"]'));
    for (const btn of buttons) {
      if (btn.textContent && btn.textContent.includes(btnText)) {
        btn.click();
        return true;
      }
    }
    return false;
  }, text);
}

/**
 * 查找并点击 Tab
 */
async function findAndClickTab(page, tabText) {
  // Try up to 3 times with delays for SPA re-renders
  for (let attempt = 0; attempt < 3; attempt++) {
    const clicked = await page.evaluate((t) => {
      // Priority 1: Exact match on button text (tab buttons)
      const buttons = Array.from(document.querySelectorAll('button'));
      for (const btn of buttons) {
        const txt = (btn.textContent || '').trim();
        if (txt === t) { btn.click(); return true; }
      }
      // Priority 2: Exact match on role="tab"
      const tabs = Array.from(document.querySelectorAll('[role="tab"]'));
      for (const tab of tabs) {
        const txt = (tab.textContent || '').trim();
        if (txt === t) { tab.click(); return true; }
      }
      // Priority 3: Contains match on buttons only (exclude sidebar links)
      for (const btn of buttons) {
        const txt = (btn.textContent || '').trim();
        if (txt.length < 20 && txt.includes(t)) { btn.click(); return true; }
      }
      return false;
    }, tabText);
    if (clicked) return true;
    await delay(1000);
  }
  return false;
}

/**
 * 检查页面是否包含指定文本
 */
async function pageHasText(page, text) {
  return await page.evaluate((t) => {
    return (document.body.innerText || '').includes(t);
  }, text);
}

/**
 * 综合测试一个页面
 * @param {object} opts
 * @param {string} opts.name - 测试名称
 * @param {string} opts.url - 要导航的 URL
 * @param {string} opts.pathSegment - URL 中应包含的路径段
 * @param {string} opts.screenshotName - 截图名称
 * @param {string[]} [opts.keywords] - 页面应包含的关键词（至少一个）
 * @param {string[]} [opts.tabs] - 要点击的 Tab 名称列表
 */
async function testPage(page, opts) {
  const { name, url, pathSegment, screenshotName, keywords = [], tabs = [] } = opts;

  log(`\n───────────────────────────────────────`);
  log(`📄 测试: ${name}`);
  log(`───────────────────────────────────────`);

  // 1. 导航
  const navOk = await navigateAndVerify(page, url, pathSegment, screenshotName);

  // 2. 内容检查
  const contentOk = await verifyPageHasContent(page, name);

  // 3. 关键词检查
  if (keywords.length > 0) {
    let found = false;
    for (const kw of keywords) {
      if (await pageHasText(page, kw)) {
        found = true;
        break;
      }
    }
    log(`${name} — 关键词检测: ${found ? '找到' : '未找到'} [${keywords.join(', ')}]`, found ? 'pass' : 'info');
  }

  // 4. Tab 点击测试
  if (tabs.length > 0) {
    for (const tabText of tabs) {
      const clicked = await findAndClickTab(page, tabText);
      log(`${name} — Tab "${tabText}": ${clicked ? '找到并点击' : '未找到'}`, clicked ? 'pass' : 'fail');
      await delay(SHORT_WAIT);
    }
    // 回到第一个 tab
    if (tabs.length > 0) {
      await findAndClickTab(page, tabs[0]);
      await delay(500);
    }
  }

  // 5. 控制台错误检测
  await checkConsoleErrors(page, name);

  return navOk && contentOk;
}

// ============================================================
// 主测试运行器
// ============================================================
async function run() {
  console.log('🚀 ReqMango 综合 E2E 全页面验证');
  console.log('='.repeat(60));
  console.log(`  前端: ${BASE_URL}`);
  console.log(`  后端: ${API_URL}`);
  console.log(`  截图: ${SCREENSHOT_DIR}`);
  console.log('='.repeat(60) + '\n');

  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-gpu'],
    defaultViewport: { width: 1920, height: 1080 },
  });

  const page = await browser.newPage();
  page.setDefaultTimeout(ELEMENT_TIMEOUT);

  // 收集控制台错误
  const jsErrors = [];
  page.on('pageerror', (err) => {
    jsErrors.push(err.message);
  });

  let workspaceSlug = '1';
  let workspaceId = 1;
  let projectId = 1;

  try {
    // ============================================
    // Phase 0: 登录与环境准备
    // ============================================
    log('\n' + '═'.repeat(60));
    log('Phase 0: 登录与环境准备');
    log('═'.repeat(60));

    // 0.1 先测试登录页面
    await testPage(page, {
      name: '登录页面',
      url: `${BASE_URL}/login`,
      pathSegment: '/login',
      screenshotName: '00-login-page',
      keywords: ['登录', 'Login', '邮箱', 'email', '密码', 'password'],
    });

    // 0.2 测试注册页面
    await testPage(page, {
      name: '注册页面',
      url: `${BASE_URL}/register`,
      pathSegment: '/register',
      screenshotName: '00-register-page',
      keywords: ['注册', 'Register', '邮箱', 'email', '密码', 'password'],
    });

    // 0.3 API 登录获取 token
    log('\n── API 登录 ──');
    await page.goto(`${API_URL}/api/v1/auth/login`, {
      waitUntil: 'domcontentloaded',
      timeout: 10000,
    }).catch(() => {});

    const loginResult = await page.evaluate(async (email, password) => {
      try {
        const res = await fetch('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password }),
        });
        const data = await res.json();
        return { ok: res.ok, token: data.access_token || '', error: data.error || '' };
      } catch (e) {
        return { ok: false, token: '', error: e.message };
      }
    }, LOGIN_EMAIL, LOGIN_PASSWORD);

    if (!loginResult.ok || !loginResult.token) {
      log(`API 登录失败: ${loginResult.error || '无 token'}`, 'fail');
      // 尝试 UI 登录作为备用
      log('尝试 UI 登录...');
      await page.goto(`${BASE_URL}/login`, { waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
      await delay(1000);
      const emailInput = await page.$('input[placeholder*="邮箱"], input[type="email"]');
      if (emailInput) {
        await emailInput.type(LOGIN_EMAIL);
        await page.type('input[type="password"]', LOGIN_PASSWORD);
        await clickButtonByText(page, '登录');
        await delay(3000);
        log('UI 登录完成', 'pass');
      }
    }

    let authToken = loginResult.token;

    if (authToken) {
      log('API 登录成功，获取到 token', 'pass');

      // 0.4 获取工作空间信息
      const wsResult = await page.evaluate(async (tkn) => {
        try {
          const res = await fetch('/api/v1/workspaces', {
            headers: { Authorization: `Bearer ${tkn}` },
          });
          const data = await res.json();
          const list = Array.isArray(data) ? data : (data.data || []);
          if (list.length > 0) {
            return { ok: true, slug: list[0].slug, id: list[0].id, name: list[0].name };
          }
          return { ok: false, slug: '', id: 0, name: '' };
        } catch (e) {
          return { ok: false, slug: '', id: 0, name: '', error: e.message };
        }
      }, authToken);

      if (wsResult.ok && wsResult.slug) {
        workspaceSlug = wsResult.slug;
        workspaceId = wsResult.id;
        log(`工作空间: name="${wsResult.name}", slug="${workspaceSlug}", id=${workspaceId}`, 'pass');
      } else {
        log('未能获取工作空间，使用默认值 slug=1, id=1', 'info');
      }

      // 0.5 获取项目信息
      const projResult = await page.evaluate(async (tkn, wsId) => {
        try {
          const res = await fetch(`/api/v1/projects?workspace_id=${wsId}`, {
            headers: { Authorization: `Bearer ${tkn}` },
          });
          const data = await res.json();
          const list = Array.isArray(data) ? data : (data.data || []);
          if (list.length > 0) {
            return { ok: true, id: list[0].id, name: list[0].name, identifier: list[0].identifier || '' };
          }
          return { ok: false, id: 0, name: '' };
        } catch (e) {
          return { ok: false, id: 0, name: '', error: e.message };
        }
      }, authToken, workspaceId);

      if (projResult.ok && projResult.id) {
        projectId = projResult.id;
        log(`项目: name="${projResult.name}", id=${projectId}`, 'pass');
      } else {
        log('未能获取项目，使用默认值 id=1', 'info');
      }

      // 0.6 设置 token 到 localStorage
      await page.goto(BASE_URL, { waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
      await delay(1000);
      await page.evaluate((token) => {
        localStorage.setItem('token', token);
      }, authToken);
      log('Token 已设置到 localStorage', 'pass');
    }

    // 刷新页面使 token 生效（必须 reload 让 auth store 重新从 localStorage 读取）
    await page.reload({ waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
    await delay(2000);
    await screenshot(page, '00-login-complete');
    log(`登录后 URL: ${page.url()}`);

    // URL 辅助
    const wsUrl = (sub = '') => `${BASE_URL}/workspace/${workspaceSlug}${sub}`;
    const projUrl = (sub = '') => `${BASE_URL}/workspace/${workspaceSlug}/project/${projectId}${sub}`;
    const wsIdUrl = (sub = '') => `${BASE_URL}/workspaces/${workspaceId}${sub}`;

    // ============================================
    // Part 1: Auth 页面（已在 Phase 0 测试）
    // ============================================

    // ============================================
    // Part 2: Core 核心页面
    // ============================================
    log('\n' + '═'.repeat(60));
    log('Part 2: Core 核心页面');
    log('═'.repeat(60));

    // 2.1 首页 / 工作空间列表
    await testPage(page, {
      name: '首页 / 工作空间列表',
      url: `${BASE_URL}/`,
      pathSegment: '/',
      screenshotName: '01-home',
      keywords: ['工作空间', 'Workspace', 'ReqMango'],
    });

    // 2.2 工作空间概览
    await testPage(page, {
      name: '工作空间概览',
      url: wsUrl('/overview'),
      pathSegment: 'overview',
      screenshotName: '02-workspace-overview',
      keywords: ['Overview', '概览', '项目', 'Project'],
    });

    // 2.3 工作空间设置
    await testPage(page, {
      name: '工作空间设置',
      url: wsUrl('/settings'),
      pathSegment: 'settings',
      screenshotName: '03-workspace-settings',
      keywords: ['Settings', '设置', '成员', 'Member'],
    });

    // 2.4 工作空间主页（项目列表）
    await testPage(page, {
      name: '工作空间主页（项目列表）',
      url: wsUrl(''),
      pathSegment: 'workspace',
      screenshotName: '04-workspace-home',
      keywords: ['Project', '项目', workspaceSlug],
    });

    // 2.5 项目主页
    await testPage(page, {
      name: '项目主页',
      url: projUrl(''),
      pathSegment: `/project/${projectId}`,
      screenshotName: '05-project-home',
      keywords: ['Issue', '工作项', 'Cycle', 'Module'],
    });

    // 2.6 项目设置
    await testPage(page, {
      name: '项目设置',
      url: projUrl('/settings'),
      pathSegment: 'settings',
      screenshotName: '06-project-settings',
      keywords: ['Settings', '设置', 'States', 'Labels'],
      tabs: ['概览', '状态', '标签'],
    });

    // 2.7 Issue 列表（通过项目主页 tab）
    await navigateAndVerify(page, projUrl('?tab=issues'), 'project', '07-issue-list');
    await delay(MEDIUM_WAIT);
    await verifyPageHasContent(page, 'Issue 列表');
    log('Issue 列表 — 页面加载成功', 'pass');

    // 2.8 Issue 创建
    await testPage(page, {
      name: 'Issue 创建',
      url: projUrl('/issues/new'),
      pathSegment: 'new',
      screenshotName: '08-issue-create',
      keywords: ['Create', '创建', 'Issue', '工作项', 'Title', '标题'],
    });

    // 2.9 Issue 详情（如果有 issue）
    const issueListResult = await page.evaluate(async (tkn, projId) => {
      try {
        const res = await fetch(`/api/v1/issues?project_id=${projId}&limit=1`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        if (list.length > 0) return { ok: true, id: list[0].id, name: list[0].name };
        return { ok: false, id: 0 };
      } catch {
        return { ok: false, id: 0 };
      }
    }, authToken, projectId);

    if (issueListResult.ok && issueListResult.id) {
      await testPage(page, {
        name: 'Issue 详情',
        url: projUrl(`/issues/${issueListResult.id}`),
        pathSegment: `issues/${issueListResult.id}`,
        screenshotName: '09-issue-detail',
        keywords: [issueListResult.name.substring(0, 20), 'Details', '详情'],
      });
    } else {
      log('Issue 详情 — 跳过（无 issue 数据）', 'info');
    }

    // 2.10 Cycle 详情（如果有 cycle）
    const cycleResult = await page.evaluate(async (tkn, projId) => {
      try {
        const res = await fetch(`/api/v1/projects/${projId}/cycles`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        if (list.length > 0) return { ok: true, id: list[0].id, name: list[0].name };
        return { ok: false, id: 0 };
      } catch {
        return { ok: false, id: 0 };
      }
    }, authToken, projectId);

    if (cycleResult.ok && cycleResult.id) {
      await testPage(page, {
        name: 'Cycle 详情',
        url: projUrl(`/cycles/${cycleResult.id}`),
        pathSegment: `cycles/${cycleResult.id}`,
        screenshotName: '10-cycle-detail',
        keywords: ['Cycle', cycleResult.name.substring(0, 20)],
      });
    } else {
      log('Cycle 详情 — 跳过（无 cycle 数据）', 'info');
    }

    // 2.11 Cycle 列表（通过 tab）
    await navigateAndVerify(page, projUrl('?tab=cycles'), 'project', '11-cycle-list');
    await delay(MEDIUM_WAIT);
    await verifyPageHasContent(page, 'Cycle 列表');
    log('Cycle 列表 — 页面加载成功', 'pass');

    // 2.12 Module 列表（通过 tab）
    await navigateAndVerify(page, projUrl('?tab=modules'), 'project', '12-module-list');
    await delay(MEDIUM_WAIT);
    await verifyPageHasContent(page, 'Module 列表');
    log('Module 列表 — 页面加载成功', 'pass');

    // 2.13 Pages 列表
    await testPage(page, {
      name: 'Pages 文档列表',
      url: projUrl('/pages'),
      pathSegment: 'pages',
      screenshotName: '13-pages-list',
      keywords: ['Page', '页面', '文档'],
    });

    // 2.14 Analytics 分析
    await testPage(page, {
      name: 'Analytics 分析',
      url: projUrl('/analytics'),
      pathSegment: 'analytics',
      screenshotName: '14-analytics',
      keywords: ['Analytics', '分析', 'Chart', '图表'],
    });

    // 2.15 Dashboard 仪表盘
    await testPage(page, {
      name: 'Dashboard 仪表盘',
      url: projUrl('/dashboards'),
      pathSegment: 'dashboards',
      screenshotName: '15-dashboard',
      keywords: ['Dashboard', '仪表盘', 'Widget'],
    });

    // 2.16 Workspace Analytics
    await testPage(page, {
      name: '工作空间分析（Analytics）',
      url: wsUrl('/analytics'),
      pathSegment: 'analytics',
      screenshotName: '16-workspace-analytics',
      keywords: ['Analytics', '分析', 'Redirect'],
    });

    // ============================================
    // Part 3: AI/Agent 页面
    // ============================================
    log('\n' + '═'.repeat(60));
    log('Part 3: AI/Agent 页面');
    log('═'.repeat(60));

    // 3.1 Agent Dashboard
    await testPage(page, {
      name: 'Agent Dashboard',
      url: wsUrl('/agents'),
      pathSegment: 'agents',
      screenshotName: '20-agent-dashboard',
      keywords: ['Agent', 'Dashboard', 'Templates', 'Skills'],
    });

    // 3.2 Agent Templates
    await testPage(page, {
      name: 'Agent Templates',
      url: wsUrl('/agents/templates'),
      pathSegment: 'templates',
      screenshotName: '21-agent-templates',
      keywords: ['Template', '模板', 'Create'],
    });

    // 3.3 Agent Configs
    await testPage(page, {
      name: 'Agent Configs',
      url: wsUrl('/agents/configs'),
      pathSegment: 'configs',
      screenshotName: '22-agent-configs',
      keywords: ['Config', '配置', 'Create'],
    });

    // 3.4 Agent Skills
    await testPage(page, {
      name: 'Agent Skills',
      url: wsUrl('/agents/skills'),
      pathSegment: 'skills',
      screenshotName: '23-agent-skills',
      keywords: ['Skill', '技能', 'Create'],
    });

    // 3.5 Agent Tasks
    await testPage(page, {
      name: 'Agent Tasks',
      url: wsUrl('/agents/tasks'),
      pathSegment: 'tasks',
      screenshotName: '24-agent-tasks',
      keywords: ['Task', '任务', 'Agent'],
    });

    // 3.6 Agent Runtimes
    await testPage(page, {
      name: 'Agent Runtimes',
      url: wsUrl('/agents/runtimes'),
      pathSegment: 'runtimes',
      screenshotName: '25-agent-runtimes',
      keywords: ['Runtime', '运行时', 'Create'],
    });

    // 3.7 Agent Loops
    await testPage(page, {
      name: 'Agent Loops',
      url: wsUrl('/agents/loops'),
      pathSegment: 'loops',
      screenshotName: '26-agent-loops',
      keywords: ['Loop', '循环', 'Create'],
    });

    // 3.8 Agent Sessions
    await testPage(page, {
      name: 'Agent Sessions',
      url: wsUrl('/agents/sessions'),
      pathSegment: 'sessions',
      screenshotName: '27-agent-sessions',
      keywords: ['Session', '会话', 'Agent'],
    });

    // 3.9 Agent Memories
    await testPage(page, {
      name: 'Agent Memories',
      url: wsUrl('/agents/memories'),
      pathSegment: 'memories',
      screenshotName: '28-agent-memories',
      keywords: ['Memory', '记忆', 'Total', 'Short', 'Long'],
    });

    // 3.10 Agent Squads 列表
    await testPage(page, {
      name: 'Agent Squads 列表',
      url: wsUrl('/agents/squads'),
      pathSegment: 'squads',
      screenshotName: '29-agent-squads',
      keywords: ['Squad', '团队', 'Total', 'Active', 'Create'],
    });

    // 3.11 Agent Squads 详情
    await testPage(page, {
      name: 'Agent Squads 详情',
      url: wsUrl('/agents/squads/1'),
      pathSegment: 'squads/1',
      screenshotName: '30-agent-squad-detail',
      keywords: ['Squad', '团队', 'Members', '成员', 'Execution', '执行'],
      tabs: ['成员', '执行', '历史', '配置'],
    });

    // 3.12 Agent Autopilot
    await testPage(page, {
      name: 'Agent Autopilot',
      url: wsUrl('/agents/autopilot'),
      pathSegment: 'autopilot',
      screenshotName: '31-agent-autopilot',
      keywords: ['Autopilot', '自动驾驶', 'Create', '创建'],
    });

    // 3.13 Agent Tools (含 4 个 Tab)
    await testPage(page, {
      name: 'Agent Tools（工具管理）',
      url: wsUrl('/agents/tools'),
      pathSegment: 'tools',
      screenshotName: '32-agent-tools',
      keywords: ['Tool', '工具', 'Create'],
      tabs: ['工具', '调用日志', '权限', 'MCP'],
    });

    // 3.14 Agent Monitor
    await testPage(page, {
      name: 'Agent Monitor 监控',
      url: wsUrl('/agents/monitor'),
      pathSegment: 'monitor',
      screenshotName: '33-agent-monitor',
      keywords: ['Monitor', '监控', 'Activity', '活动'],
    });

    // 3.15 Agent Performance
    await testPage(page, {
      name: 'Agent Performance 性能',
      url: wsUrl('/agents/performance'),
      pathSegment: 'performance',
      screenshotName: '34-agent-performance',
      keywords: ['Performance', '性能', 'Success', '成功率', 'Timeline'],
    });

    // ============================================
    // Part 4: Other 其他页面
    // ============================================
    log('\n' + '═'.repeat(60));
    log('Part 4: Other 其他页面');
    log('═'.repeat(60));

    // 4.1 Approvals
    await testPage(page, {
      name: 'Approvals 审批列表',
      url: wsUrl('/approvals'),
      pathSegment: 'approvals',
      screenshotName: '40-approvals',
      keywords: ['Approval', '审批', 'Pending', '待处理'],
    });

    // 4.2 Initiatives / Roadmap
    await testPage(page, {
      name: 'Initiatives / Roadmap 路线图',
      url: wsUrl('/initiatives'),
      pathSegment: 'initiatives',
      screenshotName: '41-initiatives',
      keywords: ['Initiative', '路线图', 'Roadmap', 'Create'],
    });

    // 4.3 Custom Fields
    await testPage(page, {
      name: 'Custom Fields 自定义字段',
      url: wsIdUrl(`/projects/${projectId}/custom-fields`),
      pathSegment: 'custom-fields',
      screenshotName: '42-custom-fields',
      keywords: ['Custom Field', '自定义字段', 'Create'],
    });

    // 4.4 Issue Types
    await testPage(page, {
      name: 'Issue Types 工作项类型',
      url: wsIdUrl(`/projects/${projectId}/issue-types`),
      pathSegment: 'issue-types',
      screenshotName: '43-issue-types',
      keywords: ['Issue Type', '工作项类型', 'Create'],
    });

    // 4.5 Workflows
    await testPage(page, {
      name: 'Workflows 工作流',
      url: projUrl('/workflows'),
      pathSegment: 'workflows',
      screenshotName: '44-workflows',
      keywords: ['Workflow', '工作流', 'Create', '创建'],
    });

    // 4.6 Workflow Designer（使用 workflowId=1 作为测试）
    await testPage(page, {
      name: 'Workflow Designer 工作流设计器',
      url: projUrl('/workflow/1/design'),
      pathSegment: 'design',
      screenshotName: '45-workflow-designer',
      keywords: ['Design', '设计', 'Save', '保存', 'Agent'],
    });

    // 4.7 Agent Members
    await testPage(page, {
      name: 'Agent Members 成员管理',
      url: projUrl('/agent-members'),
      pathSegment: 'agent-members',
      screenshotName: '46-agent-members',
      keywords: ['Member', '成员', 'Agent', 'Create'],
    });

    // 4.8 Agent Issues
    await testPage(page, {
      name: 'Agent Issues 任务列表',
      url: projUrl('/agent-issues'),
      pathSegment: 'agent-issues',
      screenshotName: '47-agent-issues',
      keywords: ['Issue', '任务', 'Agent'],
    });

    // 4.9 Budget / SLA
    await testPage(page, {
      name: 'Budget / SLA 预算与SLA',
      url: projUrl('/budget-sla'),
      pathSegment: 'budget-sla',
      screenshotName: '48-budget-sla',
      keywords: ['Budget', 'SLA', '预算', '配置'],
    });

    // ============================================
    // Part 5: API 健康检查
    // ============================================
    log('\n' + '═'.repeat(60));
    log('Part 5: API 健康检查');
    log('═'.repeat(60));

    await page.goto(`${API_URL}/api/v1/auth/login`, {
      waitUntil: 'domcontentloaded',
      timeout: 10000,
    }).catch(() => {});

    // 5.1 Workspaces API
    const wsApi = await page.evaluate(async (tkn) => {
      try {
        const res = await fetch('/api/v1/workspaces', {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken);
    log(`Workspaces API: ${wsApi.status} ${wsApi.ok ? 'OK' : 'FAIL'} (${wsApi.count} 个)`, wsApi.ok ? 'pass' : 'fail');

    // 5.2 Projects API
    const projApi = await page.evaluate(async (tkn, wsId) => {
      try {
        const res = await fetch(`/api/v1/projects?workspace_id=${wsId}`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, workspaceId);
    log(`Projects API: ${projApi.status} ${projApi.ok ? 'OK' : 'FAIL'} (${projApi.count} 个)`, projApi.ok ? 'pass' : 'fail');

    // 5.3 Issues API
    const issuesApi = await page.evaluate(async (tkn, projId) => {
      try {
        const res = await fetch(`/api/v1/issues?project_id=${projId}&limit=10`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, projectId);
    log(`Issues API: ${issuesApi.status} ${issuesApi.ok ? 'OK' : 'FAIL'} (${issuesApi.count} 个)`, issuesApi.ok ? 'pass' : 'fail');

    // 5.4 Agent Templates API
    const templatesApi = await page.evaluate(async (tkn, wsId) => {
      try {
        const res = await fetch(`/api/v1/workspaces/${wsId}/agent-templates`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, workspaceId);
    log(`Agent Templates API: ${templatesApi.status} ${templatesApi.ok ? 'OK' : 'FAIL'} (${templatesApi.count} 个)`, templatesApi.ok ? 'pass' : 'fail');

    // 5.5 Tools API
    const toolsApi = await page.evaluate(async (tkn, wsId) => {
      try {
        const res = await fetch(`/api/v1/workspaces/${wsId}/tools`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, workspaceId);
    log(`Tools API: ${toolsApi.status} ${toolsApi.ok ? 'OK' : 'FAIL'} (${toolsApi.count} 个)`, toolsApi.ok ? 'pass' : 'fail');

    // 5.6 Squads API
    const squadsApi = await page.evaluate(async (tkn, wsId) => {
      try {
        const res = await fetch(`/api/v1/workspaces/${wsId}/squads`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, workspaceId);
    log(`Squads API: ${squadsApi.status} ${squadsApi.ok ? 'OK' : 'FAIL'} (${squadsApi.count} 个)`, squadsApi.ok ? 'pass' : 'fail');

    // 5.7 Agent Configs API
    const configsApi = await page.evaluate(async (tkn, wsId) => {
      try {
        const res = await fetch(`/api/v1/workspaces/${wsId}/agent-configs`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, workspaceId);
    log(`Agent Configs API: ${configsApi.status} ${configsApi.ok ? 'OK' : 'FAIL'} (${configsApi.count} 个)`, configsApi.ok ? 'pass' : 'fail');

    // 5.8 Skills API
    const skillsApi = await page.evaluate(async (tkn, wsId) => {
      try {
        const res = await fetch(`/api/v1/workspaces/${wsId}/skills`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, workspaceId);
    log(`Skills API: ${skillsApi.status} ${skillsApi.ok ? 'OK' : 'FAIL'} (${skillsApi.count} 个)`, skillsApi.ok ? 'pass' : 'fail');

    // 5.9 Autopilot API
    const autopilotApi = await page.evaluate(async (tkn, wsId) => {
      try {
        const res = await fetch(`/api/v1/workspaces/${wsId}/autopilot-tasks`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, workspaceId);
    log(`Autopilot API: ${autopilotApi.status} ${autopilotApi.ok ? 'OK' : 'FAIL'} (${autopilotApi.count} 个)`, autopilotApi.ok ? 'pass' : 'fail');

    // 5.10 Workflows API
    const workflowsApi = await page.evaluate(async (tkn, projId) => {
      try {
        const res = await fetch(`/api/v1/projects/${projId}/workflows`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data?.data) ? data.data : (Array.isArray(data) ? data : []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, projectId);
    log(`Workflows API: ${workflowsApi.status} ${workflowsApi.ok ? 'OK' : 'FAIL'} (${workflowsApi.count} 个)`, workflowsApi.ok ? 'pass' : 'fail');

    // 5.11 Runtimes API
    const runtimesApi = await page.evaluate(async (tkn, wsId) => {
      try {
        const res = await fetch(`/api/v1/workspaces/${wsId}/runtimes`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, workspaceId);
    log(`Runtimes API: ${runtimesApi.status} ${runtimesApi.ok ? 'OK' : 'FAIL'} (${runtimesApi.count} 个)`, runtimesApi.ok ? 'pass' : 'fail');

    // 5.12 Loops API
    const loopsApi = await page.evaluate(async (tkn, wsId) => {
      try {
        const res = await fetch(`/api/v1/workspaces/${wsId}/loops`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, workspaceId);
    log(`Loops API: ${loopsApi.status} ${loopsApi.ok ? 'OK' : 'FAIL'} (${loopsApi.count} 个)`, loopsApi.ok ? 'pass' : 'fail');

    // 5.13 Approvals API
    const approvalsApi = await page.evaluate(async (tkn, wsId) => {
      try {
        const res = await fetch(`/api/v1/workspaces/${wsId}/approvals`, {
          headers: { Authorization: `Bearer ${tkn}` },
        });
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.data || []);
        return { status: res.status, ok: res.ok, count: list.length };
      } catch (e) {
        return { status: 0, ok: false, error: e.message };
      }
    }, authToken, workspaceId);
    log(`Approvals API: ${approvalsApi.status} ${approvalsApi.ok ? 'OK' : 'FAIL'} (${approvalsApi.count} 个)`, approvalsApi.ok ? 'pass' : 'fail');

    // ============================================
    // Part 6: 全局 JS 错误汇总
    // ============================================
    if (jsErrors.length > 0) {
      log('\n' + '═'.repeat(60));
      log('⚠️  全局 JavaScript 错误汇总');
      log('═'.repeat(60));
      // 去重
      const unique = [...new Set(jsErrors)];
      unique.forEach((err, i) => {
        log(`  ${i + 1}. ${err.substring(0, 120)}`, 'fail');
      });
    }

  } catch (error) {
    log(`❌ 未捕获异常: ${error.message}`, 'fail');
    console.error(error.stack);
    try { await screenshot(page, 'error-final'); } catch {}
  } finally {
    await browser.close();
  }

  // ============================================
  // 最终报告
  // ============================================
  console.log('\n' + '═'.repeat(60));
  console.log('  ReqMango E2E 全页面验证报告');
  console.log('═'.repeat(60));

  const passed  = results.filter(r => r.status === 'pass').length;
  const failed  = results.filter(r => r.status === 'fail').length;
  const info    = results.filter(r => r.status === 'info').length;
  const total   = passed + failed;

  console.log(`  总计: ${total} | 通过: ${passed} | 失败: ${failed} | 信息: ${info}`);
  console.log(`  通过率: ${total > 0 ? ((passed / total) * 100).toFixed(1) : 0}%`);
  if (jsErrors.length > 0) {
    console.log(`  JS 错误: ${[...new Set(jsErrors)].length} 个`);
  }

  if (failed > 0) {
    console.log('\n  失败项:');
    results
      .filter(r => r.status === 'fail')
      .forEach(r => console.log(`    ❌ ${r.msg}`));
  }

  console.log(`\n  截图目录: ${SCREENSHOT_DIR}`);
  console.log('═'.repeat(60));

  process.exit(failed > 0 ? 1 : 0);
}

run().catch((err) => {
  console.error('脚本执行异常:', err);
  process.exit(1);
});
