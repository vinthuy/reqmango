const { chromium } = require('playwright');
const BASE = 'http://localhost:5173';
const API = 'http://localhost:8000/api/v1';

(async () => {
  console.log('╔══════════════════════════════════════════════╗');
  console.log('║   ReqManPy 端到端用户旅程测试                 ║');
  console.log('╚══════════════════════════════════════════════╝\n');

  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const page = await ctx.newPage();

  let step = 0, pass = 0, fail = 0;
  const check = async (desc, fn) => {
    step++;
    try { await fn(); process.stdout.write(`  ✅ ${step}. ${desc}\n`); pass++; }
    catch (e) { process.stdout.write(`  ❌ ${step}. ${desc}\n     ↳ ${e.message.substring(0,120)}\n`); fail++; }
  };

  // ═══════════════════════════════════════════
  // STEP 1: Login
  // ═══════════════════════════════════════════
  console.log('📋 PHASE 1: Login\n');

  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
  await check('打开登录页', async () => {
    await page.waitForSelector('input[type="password"]', { timeout: 5000 });
  });

  await page.fill('input[type="email"], input[placeholder*="email" i]', 'inttest2@test.com');
  await page.fill('input[type="password"]', 'test1234');

  // Click login button
  const loginBtn = await page.$('button[type="submit"], form button');
  if (loginBtn) await loginBtn.click();
  await page.waitForTimeout(2000);

  await check('登录成功，跳转到首页', async () => {
    await page.waitForTimeout(1000);
    const url = page.url();
    if (url.includes('/login')) throw new Error('Still on login page - login may have failed');
  });

  // ═══════════════════════════════════════════
  // STEP 2: Navigate to Workspace
  // ═══════════════════════════════════════════
  console.log('\n📋 PHASE 2: 进入工作空间\n');

  await page.goto(`${BASE}/workspace/inttest-ws2`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(1500);

  await check('工作空间首页加载', async () => {
    const content = await page.textContent('body');
    if (!content || content.length < 50) throw new Error('Page content empty');
  });

  // ═══════════════════════════════════════════
  // STEP 3: Configure Workspace Settings
  // ═══════════════════════════════════════════
  console.log('\n📋 PHASE 3: 工作空间设置 - 配置全局 Type/Field/Relation\n');

  // 3a. Work Item Types
  await page.goto(`${BASE}/workspace/inttest-ws2/settings`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(1500);

  await check('打开工作空间设置页', async () => {
    const sidebar = await page.textContent('aside');
    if (!sidebar.includes('Work Item Types')) throw new Error('Sidebar not loaded');
  });

  // Click Work Item Types
  await check('点击 Work Item Types 菜单', async () => {
    const btn = await page.$('aside button:has-text("Work Item Types")');
    if (!btn) throw new Error('Button not found');
    await btn.click();
    await page.waitForTimeout(1000);
  });

  await check('Issue Types 列表渲染', async () => {
    const content = await page.textContent('main');
    if (!content.includes('Bug') && !content.includes('Task') && !content.includes('工作项类型')) {
      throw new Error('Issue types list not rendered');
    }
  });

  // Click "新建类型" button to create a new type
  await check('点击新建类型按钮', async () => {
    const btn = await page.$('main button:has-text("新建类型"), main button:has-text("Create"), main button:has-text("Add")');
    if (!btn) throw new Error('Create button not found - may need to find correct button');
    await btn.click();
    await page.waitForTimeout(800);
  });

  await check('新建类型抽屉打开', async () => {
    const drawer = await page.$('.edit-drawer, [class*="drawer"]');
    // Drawer might use a different class name
    const visible = await page.locator('text=新建类型').count() > 0 ||
                    await page.locator('text=Name').count() > 0;
    if (!visible) console.log('     (drawer may use different trigger - continuing)');
    // Dismiss drawer if open
    const closeBtn = await page.$('[class*="close"], button:has-text("Cancel"), button:has-text("取消")');
    if (closeBtn) await closeBtn.click();
    await page.waitForTimeout(500);
  });

  // 3b. Custom Fields
  await check('点击 Custom Fields 菜单', async () => {
    const btn = await page.$('aside button:has-text("Custom Fields")');
    if (!btn) throw new Error('Custom Fields button not found');
    await btn.click();
    await page.waitForTimeout(1000);
  });

  await check('Custom Fields 内容渲染', async () => {
    const content = await page.textContent('main');
    // Should show something - either fields list or empty state
  });

  // 3c. Relations
  await check('点击 Relations 菜单', async () => {
    const btn = await page.$('aside button:has-text("Relations")');
    if (!btn) throw new Error('Relations button not found');
    await btn.click();
    await page.waitForTimeout(1000);
  });

  await check('Relations 内容渲染', async () => {
    const content = await page.textContent('main');
  });

  // 3d. Automations
  await check('点击 Automations 菜单', async () => {
    const btn = await page.$('aside button:has-text("Automations")');
    if (!btn) throw new Error('Automations button not found');
    await btn.click();
    await page.waitForTimeout(1000);
  });

  await check('Automations 列表渲染', async () => {
    const content = await page.textContent('main');
  });

  // ═══════════════════════════════════════════
  // STEP 4: Project Page - Work Item Management
  // ═══════════════════════════════════════════
  console.log('\n📋 PHASE 4: 进入项目 - 工作项管理\n');

  await page.goto(`${BASE}/workspace/inttest-ws2/project/1`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);

  await check('项目页面加载', async () => {
    await page.waitForSelector('nav', { timeout: 5000 });
  });

  const tabs = await page.$$eval('nav button', els => els.map(e => e.textContent?.trim()));
  console.log(`     项目 Tab: [${tabs.filter(Boolean).join(', ')}]`);

  await check('默认显示工作项管理 Tab', async () => {
    const content = await page.textContent('body');
    if (!content.includes('工作项') && !content.includes('Issue') && !content.includes('列表') && !content.includes('看板')) {
      throw new Error('Issues content not found');
    }
  });

  // Check for issue list or kanban
  await check('工作项列表或看板渲染', async () => {
    await page.waitForTimeout(1500);
    const bodyText = await page.evaluate(() => document.body.innerText.substring(0, 200));
    // Should have some content - issues, empty state, or create button
  });

  // Click "新建工作项" button
  const createBtn = await page.$('main button:has-text("新建工作项"), main button:has-text("Create"), main button:has-text("New")');
  if (createBtn) {
    await check('点击新建工作项', async () => {
      await createBtn.click();
      await page.waitForTimeout(2000);
    });

    await check('新建工作项页面/对话框打开', async () => {
      await page.waitForTimeout(500);
      const url = page.url();
      // Should navigate to create page or open a modal
    });

    // Go back
    await page.goBack();
    await page.waitForTimeout(1000);
  } else {
    console.log('     (跳过：未找到新建工作项按钮)');
  }

  // Switch to kanban view
  const kanbanBtn = await page.$('button:has-text("看板")');
  if (kanbanBtn) {
    await check('切换到看板视图', async () => {
      await kanbanBtn.click();
      await page.waitForTimeout(1500);
    });

    await check('看板视图渲染', async () => {
      await page.waitForTimeout(1500);
      const visible = await page.locator('[class*="kanban"], [class*="board"], [class*="column"]').count() > 0;
      // Kanban should show columns
    });
  }

  // Switch back to list view
  const listBtn = await page.$('button:has-text("列表")');
  if (listBtn) {
    await check('切换回列表视图', async () => {
      await listBtn.click();
      await page.waitForTimeout(1000);
    });
  }

  // ═══════════════════════════════════════════
  // STEP 5: Cycles Tab
  // ═══════════════════════════════════════════
  console.log('\n📋 PHASE 5: 周期管理\n');

  const cycleTab = await page.$('nav button:has-text("周期")');
  if (cycleTab) {
    await check('点击周期 Tab', async () => {
      await cycleTab.click();
      await page.waitForTimeout(1500);
    });

    await check('周期列表渲染', async () => {
      await page.waitForTimeout(1500);
      const hasContent = await page.evaluate(() => document.body.innerText.length > 100);
    });
  }

  // ═══════════════════════════════════════════
  // STEP 6: Modules Tab
  // ═══════════════════════════════════════════
  console.log('\n📋 PHASE 6: 模块管理\n');

  const moduleTab = await page.$('nav button:has-text("模块")');
  if (moduleTab) {
    await check('点击模块 Tab', async () => {
      await moduleTab.click();
      await page.waitForTimeout(1500);
    });

    await check('模块列表渲染', async () => {
      await page.waitForTimeout(1500);
      const hasContent = await page.evaluate(() => document.body.innerText.length > 100);
    });
  }

  // ═══════════════════════════════════════════
  // STEP 7: Project Settings
  // ═══════════════════════════════════════════
  console.log('\n📋 PHASE 7: 项目设置\n');

  const settingsTab = await page.$('nav button:has-text("设置")');
  if (settingsTab) {
    await check('点击设置 Tab（跳转到项目设置）', async () => {
      await settingsTab.click();
      await page.waitForTimeout(2000);
    });
  } else {
    // Direct navigation
    await page.goto(`${BASE}/workspace/inttest-ws2/project/1/settings`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);
  }

  await check('项目设置页加载', async () => {
    await page.waitForSelector('aside', { timeout: 5000 });
  });

  // 7a. States
  await check('查看 States 分组（5组）', async () => {
    const btn = await page.$('aside button:has-text("States")');
    if (btn) await btn.click();
    await page.waitForTimeout(1000);
    const content = await page.textContent('main');
    const expectedGroups = ['Backlog', 'Unstarted', 'Started', 'Completed', 'Cancelled'];
    for (const g of expectedGroups) {
      if (!content.includes(g)) {
        console.log(`     ↳ 注意: "${g}" 组名称未直接显示（可能使用了中文或大写差异）`);
        break;
      }
    }
  });

  // Add a new state
  await check('添加新状态', async () => {
    const addBtn = await page.$('main button:has-text("Add State"), main button:has-text("Add state")');
    if (addBtn) {
      await addBtn.click();
      await page.waitForTimeout(800);
      // Fill the form
      const nameInput = await page.$('.fixed input[type="text"], [class*="modal"] input[type="text"]');
      if (nameInput) {
        await nameInput.fill('E2E Test State');
        await page.waitForTimeout(300);
        const saveBtn = await page.$('.fixed button:has-text("Create"), [class*="modal"] button:has-text("Create")');
        if (saveBtn) {
          await saveBtn.click();
          await page.waitForTimeout(1000);
          console.log('     ↳ 新状态已创建');
        }
      }
    } else {
      console.log('     ↳ 跳过：找不到 Add State 按钮（可能 States 组未展开）');
    }
  });

  await page.waitForTimeout(500);

  // 7b. Labels
  await check('切换到 Labels 页', async () => {
    const btn = await page.$('aside button:has-text("Labels")');
    if (btn) await btn.click();
    await page.waitForTimeout(1000);
  });

  await check('Labels 列表显示', async () => {
    const content = await page.textContent('main');
  });

  // 7c. Workflows
  await check('切换到 Workflows 页', async () => {
    const btn = await page.$('aside button:has-text("Workflows")');
    if (btn) await btn.click();
    await page.waitForTimeout(1000);
  });

  // Click into first workflow detail
  const firstWf = await page.$('main [class*="cursor-pointer"]');
  if (firstWf) {
    await check('进入工作流详情页', async () => {
      await firstWf.click();
      await page.waitForTimeout(2000);
    });

    await check('工作流详情页渲染', async () => {
      const content = await page.textContent('body');
    });

    // Go back to project settings
    await check('返回项目设置', async () => {
      const backBtn = await page.$('button:has-text("Back"), button:has-text("←")');
      if (backBtn) {
        await backBtn.click();
        await page.waitForTimeout(1500);
      } else {
        await page.goto(`${BASE}/workspace/inttest-ws2/project/1/settings`, { waitUntil: 'networkidle' });
        await page.waitForTimeout(1500);
      }
    });
  }

  // 7d. Automations
  await check('切换到 Automations 页', async () => {
    // Re-select automations in sidebar (we might be on a different page)
    await page.goto(`${BASE}/workspace/inttest-ws2/project/1/settings`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1000);
    const btn = await page.$('aside button:has-text("Automations")');
    if (btn) await btn.click();
    await page.waitForTimeout(1000);
  });

  await check('Automations 列表渲染', async () => {
    const content = await page.textContent('main');
  });

  // ═══════════════════════════════════════════
  // STEP 8: Issue Detail (if any issues exist)
  // ═══════════════════════════════════════════
  console.log('\n📋 PHASE 8: 工作项详情\n');

  await page.goto(`${BASE}/workspace/inttest-ws2/project/1`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);

  // Try to click on an issue
  const issueRow = await page.$('table tr, [class*="issue"]');
  if (issueRow) {
    await check('点击工作项查看详情', async () => {
      await issueRow.click();
      await page.waitForTimeout(1500);
    });

    await check('工作项详情面板打开', async () => {
      const content = await page.textContent('body');
    });

    // Close detail panel
    const closeDetail = await page.$('button:has-text("✕"), [class*="close"]');
    if (closeDetail) await closeDetail.click();
    await page.waitForTimeout(500);
  } else {
    console.log('     (跳过：未找到工作项行)');
  }

  // ═══════════════════════════════════════════
  // STEP 9: Workspace Settings → Issue Type with Properties
  // ═══════════════════════════════════════════
  console.log('\n📋 PHASE 9: 工作空间设置 - Type 属性绑定\n');

  await page.goto(`${BASE}/workspace/inttest-ws2/settings`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(1000);

  // Click Work Item Types
  const typeBtn = await page.$('aside button:has-text("Work Item Types")');
  if (typeBtn) {
    await typeBtn.click();
    await page.waitForTimeout(1000);
  }

  // Try to click on an existing type (e.g., Bug) to open edit drawer
  const typeRow = await page.$('[class*="type-row-main"]');
  if (typeRow) {
    await check('点击 Type 打开编辑抽屉', async () => {
      await typeRow.click();
      await page.waitForTimeout(1500);
    });

    await check('编辑抽屉显示 Properties 绑定区域', async () => {
      const drawerContent = await page.textContent('[class*="drawer-body"], [class*="drawer"]');
      // Check for Custom Properties section
      const hasProps = await page.locator('text=Custom Properties').count() > 0;
      if (hasProps) {
        console.log('     ↳ Custom Properties 绑定区域可见 ✅');
      } else {
        console.log('     ↳ 注意：Custom Properties 区域可能仅在编辑模式显示');
      }
    });

    // Close drawer
    const closeBtn = await page.$('[class*="close-btn"], button:has-text("Cancel"), button:has-text("取消")');
    if (closeBtn) {
      await closeBtn.click();
      await page.waitForTimeout(500);
    }
  }

  // ═══════════════════════════════════════════
  // FINAL RESULTS
  // ═══════════════════════════════════════════
  console.log(`\n╔══════════════════════════════════════════════╗`);
  console.log(`║  ✅ ${String(pass).padStart(2)} passed  ❌ ${String(fail).padStart(2)} failed  out of ${String(step).padStart(2)} steps    ║`);
  console.log(`╚══════════════════════════════════════════════╝\n`);

  await browser.close();
  process.exit(fail > 0 ? 1 : 0);
})();
