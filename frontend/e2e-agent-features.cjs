// @ts-check
/**
 * E2E Test Script for AI/Agent Feature Pages
 * Uses Puppeteer to verify all Agent-related pages load and function correctly.
 *
 * Run: node frontend/e2e-agent-features.cjs
 */
'use strict';
const puppeteer = require('puppeteer');
const fs = require('fs');
const path = require('path');

// ============================================================
// Configuration
// ============================================================
const BASE_URL = 'http://localhost:5174';
const API_URL = 'http://localhost:8000';
const SCREENSHOT_DIR = path.join(__dirname, 'e2e-agent-screenshots');

const NAV_TIMEOUT = 30000;
const WAIT_TIMEOUT = 8000;
const SHORT_WAIT = 1500;

// ============================================================
// Ensure screenshot directory
// ============================================================
if (!fs.existsSync(SCREENSHOT_DIR)) {
  fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
}

// ============================================================
// Results tracking
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
  await page.screenshot({ path: filePath, fullPage: true });
  log(`Screenshot: ${padded}-${name}.png`);
}

function delay(ms) {
  return new Promise(r => setTimeout(r, ms));
}

// ============================================================
// Test helpers
// ============================================================

/**
 * Navigate to a page, wait for content, and verify it loads.
 * Returns true if the page URL contains the expected path segment.
 */
async function navigateAndVerify(page, url, pathSegment, screenshotName) {
  await page.goto(url, { waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
  await delay(SHORT_WAIT);
  await screenshot(page, screenshotName);

  const urlMatch = page.url().includes(pathSegment);
  log(
    `${screenshotName} — 页面导航: ${urlMatch ? '成功' : 'URL不匹配'} (${page.url()})`,
    urlMatch ? 'pass' : 'fail'
  );
  return urlMatch;
}

/**
 * Check if the page has meaningful content (not just a blank/error page).
 */
async function verifyPageHasContent(page, description) {
  const hasContent = await page.evaluate(() => {
    const body = document.body;
    if (!body) return false;
    const text = (body.innerText || '').trim();
    return text.length > 50;
  });
  log(`${description} — 页面内容加载: ${hasContent ? '成功' : '内容过少'}`, hasContent ? 'pass' : 'fail');
  return hasContent;
}

/**
 * Find and click a button by text content.
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
 * Find tabs by text and click through them.
 */
async function findAndClickTabs(page, tabTexts, pageName) {
  for (const tabText of tabTexts) {
    const clicked = await page.evaluate((t) => {
      const elements = Array.from(document.querySelectorAll('button, [role="tab"], a'));
      for (const el of elements) {
        if (el.textContent && el.textContent.includes(t)) {
          el.click();
          return true;
        }
      }
      return false;
    }, tabText);
    log(`${pageName} — Tab "${tabText}": ${clicked ? '找到并点击' : '未找到'}`, clicked ? 'pass' : 'fail');
    await delay(SHORT_WAIT);
  }
}

/**
 * Check if specific text exists on the page.
 */
async function pageHasText(page, text) {
  return await page.evaluate((t) => {
    return (document.body.innerText || '').includes(t);
  }, text);
}

// ============================================================
// Main test runner
// ============================================================
async function run() {
  console.log('🚀 启动 Agent 功能 E2E 测试...\n');

  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-gpu'],
    defaultViewport: { width: 1920, height: 1080 },
  });

  const page = await browser.newPage();
  page.setDefaultTimeout(WAIT_TIMEOUT);

  let workspaceSlug = '1'; // default fallback

  try {
    // ============================================
    // Phase 0: API Login & Setup
    // ============================================
    log('═══════════════════════════════════════');
    log('Phase 0: 登录与环境准备');
    log('═══════════════════════════════════════');

    // First navigate to the frontend to establish the origin
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
    await delay(1000);

    // Login via API using page.evaluate to avoid CORS
    // Navigate to backend origin first to make API calls
    await page.goto(`${API_URL}/api/v1/auth/login`, {
      waitUntil: 'domcontentloaded',
      timeout: 10000,
    }).catch(() => {});

    const loginResult = await page.evaluate(async () => {
      try {
        const res = await fetch('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email: 'admin@reqmango.com', password: 'demo1234' }),
        });
        const data = await res.json();
        return { ok: res.ok, token: data.access_token || '', error: data.error || '' };
      } catch (e) {
        return { ok: false, token: '', error: e.message };
      }
    });

    if (!loginResult.ok || !loginResult.token) {
      log(`API 登录失败: ${loginResult.error || '无 token'}`, 'fail');
      log('尝试通过 UI 登录作为备用方案...');
    }

    let authToken = loginResult.token;

    if (authToken) {
      log('API 登录成功，获取到 token', 'pass');

      // Get workspace slug from workspaces API
      const wsResult = await page.evaluate(async (tkn) => {
        try {
          const res = await fetch('/api/v1/workspaces', {
            headers: { Authorization: `Bearer ${tkn}` },
          });
          const data = await res.json();
          const list = Array.isArray(data) ? data : (data.data || []);
          if (list.length > 0) {
            return { ok: true, slug: list[0].slug, id: list[0].id };
          }
          return { ok: false, slug: '', id: 0 };
        } catch (e) {
          return { ok: false, slug: '', id: 0, error: e.message };
        }
      }, authToken);

      if (wsResult.ok && wsResult.slug) {
        workspaceSlug = wsResult.slug;
        log(`工作空间: slug=${workspaceSlug}, id=${wsResult.id}`, 'pass');
      } else {
        log('未能获取工作空间 slug，使用默认值', 'fail');
      }

      // Navigate to frontend and set the token in localStorage
      await page.goto(BASE_URL, { waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
      await delay(1000);

      await page.evaluate((token) => {
        localStorage.setItem('token', token);
      }, authToken);

      log('Token 已设置到 localStorage', 'pass');
    } else {
      // Fallback: use UI login
      log('使用 UI 登录方式...', 'info');
      await page.goto(`${BASE_URL}/login`, { waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
      await delay(1000);

      const emailInput = await page.$('input[placeholder*="邮箱"], input[type="email"]');
      if (emailInput) {
        await emailInput.type('admin@reqmango.com');
        await page.type('input[type="password"]', 'demo1234');
        await clickButtonByText(page, '登录');
        await delay(3000);
        log('UI 登录完成', 'pass');
      } else {
        log('无法找到登录表单', 'fail');
      }
    }

    // Reload page to pick up the new token via router guard
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
    await delay(2000);
    await screenshot(page, 'login-complete');

    const afterLoginUrl = page.url();
    log(`登录后 URL: ${afterLoginUrl}`);

    // ============================================
    // Helper: build agent URL
    // ============================================
    const agentUrl = (subPath = '') => {
      return `${BASE_URL}/workspace/${workspaceSlug}/agents${subPath}`;
    };

    // ============================================
    // Test 1: Agent Dashboard (/agents)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 1: Agent Dashboard');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl(''), 'agents', '01-agent-dashboard');
    await verifyPageHasContent(page, 'Agent Dashboard');

    // Verify stats cards
    const statsCards = await page.evaluate(() => {
      const cards = document.querySelectorAll('.grid > div, [class*="stats"], [class*="card"]');
      let count = 0;
      for (const card of cards) {
        if (card.textContent.includes('Templates') || card.textContent.includes('Configs') ||
            card.textContent.includes('Skills') || card.textContent.includes('Tasks')) {
          count++;
        }
      }
      return count;
    });
    log(`Dashboard 统计卡片: 找到 ${statsCards} 个`, statsCards >= 3 ? 'pass' : 'fail');

    // Verify navigation grid (link cards)
    const navLinks = await page.evaluate(() => {
      const links = document.querySelectorAll('a[href*="/agents/"]');
      return links.length;
    });
    log(`Dashboard 导航链接: ${navLinks} 个`, navLinks >= 5 ? 'pass' : 'fail');

    // ============================================
    // Test 2: Agent Templates (/agents/templates)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 2: Agent Templates');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/templates'), 'templates', '02-agent-templates');
    await verifyPageHasContent(page, 'Agent Templates');

    // Click "Create" button and verify modal
    const createClicked = await clickButtonByText(page, 'Create');
    if (!createClicked) {
      await clickButtonByText(page, '创建');
    }
    await delay(1000);
    await screenshot(page, '02b-templates-create-modal');

    // Check if a modal/dialog appeared
    const modalVisible = await page.evaluate(() => {
      const modals = document.querySelectorAll('[class*="modal"], [class*="dialog"], [role="dialog"], .ant-modal, .ant-drawer');
      for (const m of modals) {
        if (m.offsetParent !== null || getComputedStyle(m).display !== 'none') {
          return true;
        }
      }
      return false;
    });
    log(`Templates 创建弹窗: ${modalVisible ? '已打开' : '未检测到弹窗'}`, modalVisible ? 'pass' : 'fail');

    // Close modal if open
    if (modalVisible) {
      await page.evaluate(() => {
        const closeButtons = document.querySelectorAll('[class*="close"], [aria-label="Close"], .ant-modal-close');
        for (const btn of closeButtons) {
          btn.click();
          break;
        }
      });
      await delay(500);
    }

    // ============================================
    // Test 3: Agent Configs (/agents/configs)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 3: Agent Configs');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/configs'), 'configs', '03-agent-configs');
    await verifyPageHasContent(page, 'Agent Configs');

    const configList = await page.evaluate(() => {
      const rows = document.querySelectorAll('tr, [class*="list"] > div, [class*="item"]');
      return rows.length;
    });
    log(`Config 列表项数: ${configList}`, 'info');

    // ============================================
    // Test 4: Skills (/agents/skills)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 4: Skills');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/skills'), 'skills', '04-agent-skills');
    await verifyPageHasContent(page, 'Skills');

    const skillList = await page.evaluate(() => {
      const rows = document.querySelectorAll('tr, [class*="list"] > div, [class*="card"]');
      return rows.length;
    });
    log(`Skills 列表项数: ${skillList}`, 'info');

    // ============================================
    // Test 5: Agent Tasks (/agents/tasks)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 5: Agent Tasks');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/tasks'), 'tasks', '05-agent-tasks');
    await verifyPageHasContent(page, 'Agent Tasks');

    // Check for task list or empty state
    const hasTaskContent = await page.evaluate(() => {
      const text = document.body.innerText || '';
      return text.includes('task') || text.includes('Task') || text.includes('任务') ||
             text.includes('No ') || text.includes('empty') || text.includes('Empty');
    });
    log(`Tasks 页面内容: ${hasTaskContent ? '有内容/空状态' : '未检测到内容'}`, hasTaskContent ? 'pass' : 'fail');

    // ============================================
    // Test 6: Runtimes (/agents/runtimes)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 6: Runtimes');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/runtimes'), 'runtimes', '06-agent-runtimes');
    await verifyPageHasContent(page, 'Runtimes');

    // ============================================
    // Test 7: Loops (/agents/loops)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 7: Loops');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/loops'), 'loops', '07-agent-loops');
    await verifyPageHasContent(page, 'Loops');

    // ============================================
    // Test 8: Sessions (/agents/sessions)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 8: Sessions');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/sessions'), 'sessions', '08-agent-sessions');
    await verifyPageHasContent(page, 'Sessions');

    // ============================================
    // Test 9: Memories (/agents/memories)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 9: Memories');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/memories'), 'memories', '09-agent-memories');
    await verifyPageHasContent(page, 'Memories');

    // Check for statistics cards (Total, Short-term, Medium-term, Long-term)
    const memoryStats = await page.evaluate(() => {
      const text = document.body.innerText || '';
      let count = 0;
      const keywords = ['Total', 'Short', 'Medium', 'Long', '总计', '短期', '中期', '长期'];
      for (const kw of keywords) {
        if (text.includes(kw)) count++;
      }
      return count;
    });
    log(`Memories 统计卡片关键词: ${memoryStats} 个`, memoryStats >= 2 ? 'pass' : 'fail');

    // ============================================
    // Test 10: Squads (/agents/squads) — with create & detail
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 10: Squads (含创建和详情)');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/squads'), 'squads', '10-agent-squads');
    await verifyPageHasContent(page, 'Squads');

    // Verify stats cards (Total, Active, Executions)
    const squadStats = await page.evaluate(() => {
      const text = document.body.innerText || '';
      let count = 0;
      const keywords = ['Total', 'Active', 'Execution', '总计', '活跃', '执行'];
      for (const kw of keywords) {
        if (text.includes(kw)) count++;
      }
      return count;
    });
    log(`Squads 统计卡片: ${squadStats} 个`, squadStats >= 2 ? 'pass' : 'fail');

    // Verify squad table exists
    const squadTableExists = await page.evaluate(() => {
      return document.querySelector('table') !== null;
    });
    log(`Squads 表格: ${squadTableExists ? '存在' : '不存在'}`, squadTableExists ? 'pass' : 'fail');

    // Click "Create" button
    log('尝试创建新的 Squad...');
    const squadCreated = await clickButtonByText(page, 'Create');
    if (!squadCreated) {
      await clickButtonByText(page, '创建');
    }
    await delay(1500);
    await screenshot(page, '10b-squads-create-modal');

    // Fill in the squad form
    const formFilled = await page.evaluate(() => {
      const inputs = document.querySelectorAll('input, textarea');
      let filled = 0;
      for (const input of inputs) {
        const placeholder = (input.placeholder || '').toLowerCase();
        const label = input.closest('label, [class*="form-item"]')?.textContent || '';
        if (placeholder.includes('name') || placeholder.includes('名称') || label.includes('Name') || label.includes('名称')) {
          input.value = 'E2E Test Squad';
          input.dispatchEvent(new Event('input', { bubbles: true }));
          filled++;
        }
        if (placeholder.includes('desc') || placeholder.includes('描述') || label.includes('Description') || label.includes('描述')) {
          input.value = 'Automated E2E test squad for agent features';
          input.dispatchEvent(new Event('input', { bubbles: true }));
          filled++;
        }
        if (placeholder.includes('goal') || placeholder.includes('目标') || label.includes('Goal') || label.includes('目标')) {
          input.value = 'E2E test goal for squad verification';
          input.dispatchEvent(new Event('input', { bubbles: true }));
          filled++;
        }
      }
      return filled;
    });
    log(`Squad 表单填充: ${formFilled} 个字段`, formFilled > 0 ? 'pass' : 'fail');

    // Submit form
    const submitClicked = await page.evaluate(() => {
      const buttons = Array.from(document.querySelectorAll('button'));
      for (const btn of buttons) {
        const text = (btn.textContent || '').trim();
        if (text === 'Save' || text === '保存' || text === 'Create' || text === '创建' ||
            text === '确认' || text === 'Confirm' || text === 'Submit' || text === '提交') {
          btn.click();
          return text;
        }
      }
      return null;
    });
    log(`Squad 提交按钮: ${submitClicked ? `已点击 "${submitClicked}"` : '未找到'}`, submitClicked ? 'pass' : 'fail');
    await delay(2000);
    await screenshot(page, '10c-squads-after-create');

    // Reload squads list to see if the new squad appears
    await page.goto(agentUrl('/squads'), { waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
    await delay(2000);
    await screenshot(page, '10d-squads-list-reload');

    const squadAppeared = await page.evaluate(() => {
      return (document.body.innerText || '').includes('E2E Test Squad');
    });
    log(`Squad 出现在列表中: ${squadAppeared ? '是' : '否'}`, squadAppeared ? 'pass' : 'fail');

    // Click into squad detail — try clicking the first view button
    log('进入 Squad 详情页...');
    const viewClicked = await page.evaluate(() => {
      // Look for view/eye icon buttons in table rows
      const buttons = document.querySelectorAll('button');
      for (const btn of buttons) {
        const svg = btn.querySelector('svg');
        const row = btn.closest('tr');
        if (svg && row) {
          btn.click();
          return true;
        }
      }
      return false;
    });

    if (viewClicked) {
      await delay(2000);
      await screenshot(page, '10e-squad-detail');
      log('Squad 详情页: 已进入', 'pass');
    } else {
      // Fallback: navigate directly to squad detail with ID 1
      log('未找到查看按钮，直接导航到 squad/1 ...');
      await page.goto(agentUrl('/squads/1'), { waitUntil: 'domcontentloaded', timeout: NAV_TIMEOUT });
      await delay(2000);
      await screenshot(page, '10e-squad-detail-direct');
      log('Squad 详情页: 直接导航', 'info');
    }

    // Verify squad detail page has tabs
    const squadDetailTabs = await page.evaluate(() => {
      const tabs = document.querySelectorAll('button[class*="border-b"], nav button, [role="tab"]');
      const tabTexts = [];
      for (const tab of tabs) {
        const text = (tab.textContent || '').trim();
        if (text.length > 0 && text.length < 30) {
          tabTexts.push(text);
        }
      }
      return tabTexts;
    });
    log(`Squad 详情页 Tabs: [${squadDetailTabs.join(', ')}]`, squadDetailTabs.length >= 2 ? 'pass' : 'fail');

    // Click through squad detail tabs
    await findAndClickTabs(page, squadDetailTabs, 'Squad Detail');

    // ============================================
    // Test 11: Autopilot (/agents/autopilot)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 11: Autopilot');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/autopilot'), 'autopilot', '11-agent-autopilot');
    await verifyPageHasContent(page, 'Autopilot');

    // ============================================
    // Test 12: Tools (/agents/tools)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 12: Tools (含4个Tab)');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/tools'), 'tools', '12-agent-tools');
    await verifyPageHasContent(page, 'Tools');

    // Find and click through the 4 tabs: Tools, Logs, Permissions, MCP
    const toolTabTexts = ['Tools', 'Logs', 'Permission', 'MCP'];
    for (const tabText of toolTabTexts) {
      const clicked = await page.evaluate((t) => {
        const tabs = document.querySelectorAll('button');
        for (const tab of tabs) {
          const text = (tab.textContent || '').trim();
          if (text.includes(t)) {
            tab.click();
            return text;
          }
        }
        return null;
      }, tabText);

      log(`Tools — Tab "${tabText}": ${clicked ? `找到 "${clicked}" 并点击` : '未找到'}`, clicked ? 'pass' : 'fail');
      await delay(SHORT_WAIT);
      await screenshot(page, `12b-tools-tab-${tabText.toLowerCase()}`);
    }

    // Verify table or content exists in Tools tab
    const toolsContent = await page.evaluate(() => {
      const table = document.querySelector('table');
      const text = document.body.innerText || '';
      return {
        hasTable: !!table,
        hasCreateButton: text.includes('Create') || text.includes('创建') || text.includes('Tool'),
      };
    });
    log(`Tools — 表格: ${toolsContent.hasTable ? '存在' : '不存在'}, 创建按钮: ${toolsContent.hasCreateButton ? '存在' : '不存在'}`,
      toolsContent.hasTable || toolsContent.hasCreateButton ? 'pass' : 'fail');

    // ============================================
    // Test 13: Monitor (/agents/monitor)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 13: Monitor');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/monitor'), 'monitor', '13-agent-monitor');
    await verifyPageHasContent(page, 'Monitor');

    // ============================================
    // Test 14: Performance (/agents/performance)
    // ============================================
    log('\n═══════════════════════════════════════');
    log('Test 14: Performance');
    log('═══════════════════════════════════════');

    await navigateAndVerify(page, agentUrl('/performance'), 'performance', '14-agent-performance');
    await verifyPageHasContent(page, 'Performance');

  } catch (error) {
    log(`❌ 未捕获异常: ${error.message}`, 'fail');
    console.error(error.stack);
    try {
      await screenshot(page, 'error-final');
    } catch (_) {
      // ignore screenshot errors
    }
  } finally {
    await browser.close();
  }

  // ============================================
  // Final Summary
  // ============================================
  console.log('\n' + '═'.repeat(60));
  console.log('  Agent 功能 E2E 测试报告');
  console.log('═'.repeat(60));

  const passed = results.filter(r => r.status === 'pass').length;
  const failed = results.filter(r => r.status === 'fail').length;
  const info = results.filter(r => r.status === 'info').length;
  const total = passed + failed;

  console.log(`  总计: ${total} | 通过: ${passed} | 失败: ${failed} | 信息: ${info}`);
  console.log(`  通过率: ${total > 0 ? ((passed / total) * 100).toFixed(1) : 0}%`);

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
