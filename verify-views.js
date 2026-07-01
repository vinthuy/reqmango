const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1920, height: 1080 } });
  const BASE = 'http://localhost:5175';

  try {
    // 1. Visit login page
    console.log('=== 1. Login Page ===');
    await page.goto(BASE + '/login', { waitUntil: 'networkidle', timeout: 15000 });
    const loginText = await page.textContent('body');
    console.log('Has login form:', loginText.includes('Login') || loginText.includes('login'));

    // 2. Login
    await page.fill('input[type="email"]', 'test@test.com');
    await page.fill('input[type="password"]', 'test1234');
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);
    console.log('After login URL:', page.url());
    console.log('Login success:', page.url().includes('workspace') || page.url().includes('home'));

    // 3. Navigate to project
    console.log('\n=== 2. Project Page ===');
    await page.goto(BASE + '/workspace/test-workspace/project/1', { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(3000);
    const h1 = await page.$('h1');
    const projectTitle = h1 ? await h1.textContent() : 'NOT FOUND';
    console.log('Project title:', projectTitle?.substring(0, 50));

    // 4. Check SavedView selector
    console.log('\n=== 3. Saved View Selector ===');
    const svVisible = await page.$('.saved-view-selector');
    console.log('SavedView selector visible:', !!svVisible);
    if (svVisible) {
      const svText = await svVisible.textContent();
      console.log('SV text:', svText?.substring(0, 80));
    }

    // 5. Check for console errors
    console.log('\n=== 4. Console Errors ===');
    const errors = [];
    page.on('console', msg => { if (msg.type() === 'error') errors.push(msg.text()); });
    await page.waitForTimeout(2000);
    console.log('Console errors:', errors.length);
    if (errors.length > 0) {
      errors.slice(0, 5).forEach(e => console.log('  -', e.substring(0, 120)));
    }

    // 6. Check issue-list view loads
    console.log('\n=== 5. List View ===');
    const listView = await page.$('.issue-list');
    console.log('List view visible:', !!listView);

    // 7. Check kanban view
    console.log('\n=== 6. Switch to Kanban ===');
    const allButtons = await page.$$('button');
    let clicked = false;
    for (const btn of allButtons) {
      const text = await btn.textContent();
      if (text && (text.trim() === 'Kanban' || text.trim() === '看板')) {
        await btn.click();
        clicked = true;
        break;
      }
    }
    if (clicked) {
      await page.waitForTimeout(2000);
      const kanban = await page.$('.issue-kanban');
      console.log('Kanban view visible:', !!kanban);
    } else {
      console.log('Kanban button not found');
    }

    // 8. Check tree view
    console.log('\n=== 7. Switch to Tree ===');
    const allButtons2 = await page.$$('button');
    let treeClicked = false;
    for (const btn of allButtons2) {
      const text = await btn.textContent();
      if (text && (text.trim() === 'Tree' || text.trim() === '树形')) {
        await btn.click();
        treeClicked = true;
        break;
      }
    }
    if (treeClicked) {
      await page.waitForTimeout(2000);
      const tree = await page.$('.issue-tree-view');
      console.log('Tree view visible:', !!tree);
    } else {
      console.log('Tree button not found');
    }

    // 9. Check calendar view
    console.log('\n=== 8. Switch to Calendar ===');
    const allButtons3 = await page.$$('button');
    let calClicked = false;
    for (const btn of allButtons3) {
      const text = await btn.textContent();
      if (text && (text.trim() === 'Calendar' || text.trim() === '日历')) {
        await btn.click();
        calClicked = true;
        break;
      }
    }
    if (calClicked) {
      await page.waitForTimeout(2000);
      const cal = await page.$('.issue-calendar');
      console.log('Calendar view visible:', !!cal);
    } else {
      console.log('Calendar button not found');
    }

    // 10. Check gantt view
    console.log('\n=== 9. Switch to Gantt ===');
    const allButtons4 = await page.$$('button');
    let ganttClicked = false;
    for (const btn of allButtons4) {
      const text = await btn.textContent();
      if (text && (text.trim() === 'Gantt' || text.trim() === '甘特')) {
        await btn.click();
        ganttClicked = true;
        break;
      }
    }
    if (ganttClicked) {
      await page.waitForTimeout(2000);
      const gantt = await page.$('.issue-gantt');
      console.log('Gantt view visible:', !!gantt);
    } else {
      console.log('Gantt button not found');
    }

    console.log('\n=== DONE ===');
  } catch (e) {
    console.error('ERROR:', e.message);
  } finally {
    await browser.close();
  }
})();
