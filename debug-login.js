const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1920, height: 1080 } });

  // Check login page structure
  await page.goto('http://localhost:5175/login', { waitUntil: 'networkidle', timeout: 15000 });
  await page.waitForTimeout(2000);

  const bodyText = await page.textContent('body');
  console.log('Body text:', bodyText?.substring(0, 500));

  // Find all input elements
  const inputs = await page.locator('input').all();
  console.log('\nInput count:', inputs.length);
  for (const inp of inputs) {
    const type = await inp.getAttribute('type');
    const placeholder = await inp.getAttribute('placeholder');
    console.log('  Input:', type, placeholder);
  }

  // Find all buttons
  const buttons = await page.locator('button').all();
  console.log('\nButton count:', buttons.length);
  for (const btn of buttons) {
    const text = await btn.textContent();
    console.log('  Button:', text?.trim()?.substring(0, 40));
  }

  // Try to fill in and submit
  const emailInput = page.locator('input[type="email"]');
  const emailCount = await emailInput.count();
  console.log('\nEmail inputs:', emailCount);

  if (emailCount > 0) {
    await emailInput.fill('test@test.com');
    const pwInput = page.locator('input[type="password"]');
    await pwInput.fill('test1234');

    // Find submit button
    const submitBtn = page.locator('button[type="submit"]');
    await submitBtn.click();
    await page.waitForTimeout(3000);
    console.log('After submit URL:', page.url());
  }

  // Navigate to project directly
  console.log('\n--- Direct project navigation ---');
  await page.goto('http://localhost:5175/workspace/test-workspace/project/1', { waitUntil: 'networkidle', timeout: 15000 });
  await page.waitForTimeout(3000);
  console.log('Project page URL:', page.url());
  const h1 = page.locator('h1');
  if (await h1.count() > 0) {
    console.log('H1 text:', await h1.first().textContent());
  }

  // Check for .filter-bar
  const filterBar = page.locator('.filter-bar');
  const fbCount = await filterBar.count();
  console.log('Filter bar count:', fbCount);

  // Check view buttons in filter bar
  if (fbCount > 0) {
    const fbButtons = filterBar.locator('button');
    const fbBtnCount = await fbButtons.count();
    console.log('Filter bar buttons:', fbBtnCount);
    for (let i = 0; i < fbBtnCount && i < 10; i++) {
      const text = await fbButtons.nth(i).textContent();
      console.log('  FB btn:', text?.trim()?.substring(0, 30));
    }
  }

  // Check saved-view-selector
  const sv = page.locator('.saved-view-selector');
  console.log('SavedView selector count:', await sv.count());

  // Check for issue-list/kanban/tree views
  console.log('IssueList:', await page.locator('.issue-list').count());
  console.log('IssueKanban:', await page.locator('.issue-kanban').count());
  console.log('IssueTreeView:', await page.locator('.issue-tree-view').count());
  console.log('IssueCalendar:', await page.locator('.issue-calendar').count());
  console.log('IssueGantt:', await page.locator('.issue-gantt').count());

  await browser.close();
})();
