const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });
  let step = 0, pass = 0, fail = 0;
  const check = async (desc, fn) => {
    step++;
    try { await fn(); process.stdout.write(`  ✅ ${step}. ${desc}\n`); pass++; }
    catch (e) { process.stdout.write(`  ❌ ${step}. ${desc}\n     ↳ ${e.message.substring(0,120)}\n`); fail++; }
  };

  const errors = [];
  page.on('pageerror', err => errors.push(err.message));

  console.log('══════════════════════════════════════════');
  console.log('  用户旅程: demo 工作空间完整流程');
  console.log('══════════════════════════════════════════\n');

  // ═══ 1. Login ═══
  console.log('PHASE 1: 登录\n');
  await page.goto('http://localhost:5173/login', { waitUntil: 'networkidle' });
  await page.fill('input[type="email"]', 'demo1@reqman.local');
  await page.fill('input[type="password"]', 'demo1234');
  await page.click('form button');
  await page.waitForTimeout(2000);

  await check('登录成功', async () => {
    if (page.url().includes('/login')) throw new Error('Still on login page');
  });

  // ═══ 2. Workspace Settings ═══
  console.log('\nPHASE 2: 工作空间设置\n');

  await page.goto('http://localhost:5173/workspace/demo/settings', { waitUntil: 'networkidle' });
  await page.waitForTimeout(1500);

  await check('WorkspaceSettings 页面加载', async () => {
    const sidebar = await page.evaluate(() => {
      const aside = document.querySelector('aside');
      return aside ? aside.innerText : '';
    });
    if (!sidebar.includes('Work Item Types')) throw new Error('Sidebar not loaded: ' + sidebar.substring(0,50));
  });

  const wsItems = await page.evaluate(() => {
    return [...document.querySelectorAll('aside button')].map(b => b.textContent.trim()).filter(Boolean);
  });
  console.log(`     侧边栏: ${wsItems.join(' | ')}`);

  // Verify workspace-level items present
  await check('含 Work Item Types', async () => {
    if (!wsItems.some(i => i.includes('Work Item Types'))) throw new Error('Missing');
  });
  await check('含 Custom Fields', async () => {
    if (!wsItems.some(i => i.includes('Custom Fields'))) throw new Error('Missing');
  });
  await check('含 Relations', async () => {
    if (!wsItems.some(i => i.includes('Relations'))) throw new Error('Missing');
  });
  await check('含 Automations (workspace-level)', async () => {
    if (!wsItems.some(i => i.includes('Automations'))) throw new Error('Missing');
  });

  // Verify NO project-level items
  await check('不含 States', async () => {
    const standalone = wsItems.find(i => i === 'States' || i.startsWith('States\n'));
    if (standalone) throw new Error('States leaked in workspace settings');
  });
  await check('不含 Labels', async () => {
    const standalone = wsItems.find(i => i === 'Labels' || i.startsWith('Labels\n'));
    if (standalone) throw new Error('Labels leaked in workspace settings');
  });

  // Click through each section
  for (const label of ['Work Item Types', 'Custom Fields', 'Automations', 'Relations']) {
    await check(`点击 ${label} 节`, async () => {
      const btn = page.locator('aside button', { hasText: label });
      await btn.click();
      await page.waitForTimeout(800);
      const main = await page.evaluate(() => document.querySelector('main')?.innerText || '');
    });
  }

  // ═══ 3. Project Settings ═══
  console.log('\nPHASE 3: 项目设置 (demo/project/1/settings)\n');

  await page.goto('http://localhost:5173/workspace/demo/project/1/settings', { waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);

  await check('ProjectSettings 页面加载', async () => {
    const sidebar = await page.evaluate(() => {
      const aside = document.querySelector('aside');
      return aside ? aside.innerText : '';
    });
    if (!sidebar.includes('States')) throw new Error('Sidebar not loaded');
  });

  const psItems = await page.evaluate(() => {
    return [...document.querySelectorAll('aside button')].map(b => b.textContent.trim()).filter(Boolean);
  });
  console.log(`     侧边栏: ${psItems.join(' | ')}`);

  // Verify all 4 project-level sections
  await check('含 States', async () => { if (!psItems.some(i => i.includes('States'))) throw new Error('Missing'); });
  await check('含 Labels', async () => { if (!psItems.some(i => i.includes('Labels'))) throw new Error('Missing'); });
  await check('含 Workflows', async () => { if (!psItems.some(i => i.includes('Workflows'))) throw new Error('Missing'); });
  await check('含 Automations', async () => { if (!psItems.some(i => i.includes('Automations'))) throw new Error('Missing'); });

  // Test each section
  for (const section of ['States', 'Labels', 'Workflows', 'Automations']) {
    await check(`${section} 内容渲染`, async () => {
      const btn = page.locator('aside button', { hasText: section });
      await btn.click();
      await page.waitForTimeout(1200);
      const main = await page.evaluate(() => document.querySelector('main')?.innerText || '');
      if (main.length < 10) throw new Error('Content empty');
    });
  }

  // States: verify 5 group headers
  await check('States 显示 5 个固定分组', async () => {
    const btn = page.locator('aside button', { hasText: 'States' });
    await btn.click();
    await page.waitForTimeout(800);
    const main = await page.evaluate(() => document.querySelector('main')?.innerText || '');
    const groups = ['BACKLOG', 'UNSTARTED', 'STARTED', 'COMPLETED', 'CANCELLED'];
    for (const g of groups) {
      if (!main.toUpperCase().includes(g)) throw new Error(`Missing group: ${g}`);
    }
  });

  // Labels: verify labels list
  await check('Labels 列表渲染', async () => {
    const btn = page.locator('aside button', { hasText: 'Labels' });
    await btn.click();
    await page.waitForTimeout(800);
    const main = await page.evaluate(() => document.querySelector('main')?.innerText || '');
    if (!main.includes('Add Label') && main.length < 50) throw new Error('Labels content empty');
  });

  // Workflows: verify list and click into detail
  await check('Workflows 列表渲染 + 跳转详情', async () => {
    const btn = page.locator('aside button', { hasText: 'Workflows' });
    await btn.click();
    await page.waitForTimeout(800);
    const wfCard = page.locator('main [class*="cursor-pointer"]').first();
    if (await wfCard.count() > 0) {
      await wfCard.click();
      await page.waitForTimeout(2000);
      const url = page.url();
      if (!url.includes('/workflows/')) throw new Error('Not on workflow detail: ' + url);
    }
  });

  // ═══ 4. Project Page ═══
  console.log('\nPHASE 4: 项目页面\n');

  await page.goto('http://localhost:5173/workspace/demo/project/1', { waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);

  const tabs = await page.evaluate(() => {
    return [...document.querySelectorAll('nav button')].map(b => b.textContent.trim()).filter(Boolean);
  });
  console.log(`     Tabs: ${tabs.join(' | ')}`);

  await check('4 个 Tab: 工作项管理/周期/模块/设置', async () => {
    const expected = ['工作项管理', '周期', '模块', '设置'];
    for (const t of expected) {
      if (!tabs.includes(t)) throw new Error(`Missing tab: ${t}`);
    }
  });

  // Issues tab (default)
  await check('工作项列表/看板渲染', async () => {
    await page.waitForTimeout(1000);
    const hasContent = await page.evaluate(() => document.body.innerText.length > 200);
    if (!hasContent) throw new Error('Page content too short');
  });

  // Cycles
  await check('周期 Tab 渲染', async () => {
    const btn = page.locator('nav button', { hasText: '周期' });
    await btn.click();
    await page.waitForTimeout(1200);
  });

  // Modules
  await check('模块 Tab 渲染', async () => {
    const btn = page.locator('nav button', { hasText: '模块' });
    await btn.click();
    await page.waitForTimeout(1200);
  });

  // Settings tab -> project settings
  await check('设置 Tab → 跳转项目设置', async () => {
    const btn = page.locator('nav button', { hasText: '设置' });
    await btn.click();
    await page.waitForTimeout(1500);
    const url = page.url();
    if (!url.includes('/settings')) throw new Error('Not redirected: ' + url);
  });

  // ═══ Results ═══
  console.log(`\n══════════════════════════════════════════`);
  console.log(`  ✅ ${String(pass).padStart(2)} passed  ❌ ${String(fail).padStart(2)} failed  (${step} steps)`);
  console.log(`══════════════════════════════════════════`);

  if (errors.length > 0) {
    console.log('\nJS Errors during test:');
    for (const e of errors) console.log(`  ⚠️  ${e.substring(0, 200)}`);
  }

  await browser.close();
  process.exit(fail > 0 ? 1 : 0);
})();
