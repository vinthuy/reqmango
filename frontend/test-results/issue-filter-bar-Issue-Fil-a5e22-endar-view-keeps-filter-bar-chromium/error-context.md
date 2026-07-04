# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: issue-filter-bar.spec.ts >> Issue Filter Bar >> switch to calendar view keeps filter bar
- Location: e2e\issue-filter-bar.spec.ts:159:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.issue-filter-bar').first()
Expected: visible
Timeout: 10000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 10000ms
  - waiting for locator('.issue-filter-bar').first()

```

```yaml
- main
```

# Test source

```ts
  66  | 
  67  |   test('quick filter chips are visible', async ({ page }) => {
  68  |     await goToProject(page)
  69  |     const chips = page.locator('.issue-filter-bar button:has-text("我的"), .issue-filter-bar button:has-text("Mine"), .issue-filter-bar button:has-text("未分配"), .issue-filter-bar button:has-text("高优先级")')
  70  |     const count = await chips.count()
  71  |     expect(count).toBeGreaterThanOrEqual(1)
  72  |   })
  73  | 
  74  |   // ============================================================
  75  |   // 2. Quick Filter Toggle
  76  |   // ============================================================
  77  |   test('quick filter toggles to active state', async ({ page }) => {
  78  |     await goToProject(page)
  79  |     const chip = page.locator('.issue-filter-bar button:has-text("高优先级"), .issue-filter-bar button:has-text("High Priority")').first()
  80  |     if (await chip.isVisible({ timeout: 3000 }).catch(() => false)) {
  81  |       // Check initial state (should be inactive)
  82  |       const initialClass = await chip.getAttribute('class')
  83  |       await chip.click()
  84  |       await page.waitForTimeout(500)
  85  |       // Should have changed appearance
  86  |       const newClass = await chip.getAttribute('class')
  87  |       expect(newClass).not.toBe(initialClass)
  88  |     }
  89  |   })
  90  | 
  91  |   test('quick filter toggles back off', async ({ page }) => {
  92  |     await goToProject(page)
  93  |     const chip = page.locator('.issue-filter-bar button:has-text("高优先级"), .issue-filter-bar button:has-text("High Priority")').first()
  94  |     if (await chip.isVisible({ timeout: 3000 }).catch(() => false)) {
  95  |       await chip.click()
  96  |       await page.waitForTimeout(300)
  97  |       await chip.click()
  98  |       await page.waitForTimeout(300)
  99  |     }
  100 |     // Page should still be functional
  101 |     await expect(page.locator('body')).toBeVisible()
  102 |   })
  103 | 
  104 |   // ============================================================
  105 |   // 3. Add Filter dropdown
  106 |   // ============================================================
  107 |   test('add filter button is clickable', async ({ page }) => {
  108 |     await goToProject(page)
  109 |     // Find the + filter button and click it
  110 |     const addBtn = page.locator('.issue-filter-bar button:has(svg)').last()
  111 |     if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  112 |       await addBtn.click()
  113 |       await page.waitForTimeout(500)
  114 |     }
  115 |     // Page should still be functional
  116 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 5000 })
  117 |   })
  118 | 
  119 |   test('filter dropdown closes on second click', async ({ page }) => {
  120 |     await goToProject(page)
  121 |     const addBtn = page.locator('.issue-filter-bar button').filter({ has: page.locator('svg') }).last()
  122 |     if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  123 |       await addBtn.click()
  124 |       await page.waitForTimeout(300)
  125 |       await addBtn.click()
  126 |       await page.waitForTimeout(300)
  127 |     }
  128 |     await expect(page.locator('body')).toBeVisible()
  129 |   })
  130 | 
  131 |   // ============================================================
  132 |   // 4. View Switching preserves filter bar
  133 |   // ============================================================
  134 |   test('filter bar visible in list view', async ({ page }) => {
  135 |     await goToProject(page)
  136 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  137 |   })
  138 | 
  139 |   test('switch to kanban view keeps filter bar', async ({ page }) => {
  140 |     await goToProject(page)
  141 |     const kanbanBtn = page.locator('.issue-filter-bar button[title="看板"], .issue-filter-bar button[title*="kanban"], .issue-filter-bar button:has-text("📌")').first()
  142 |     if (await kanbanBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  143 |       await kanbanBtn.click()
  144 |       await page.waitForTimeout(800)
  145 |     }
  146 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  147 |   })
  148 | 
  149 |   test('switch to tree view keeps filter bar', async ({ page }) => {
  150 |     await goToProject(page)
  151 |     const treeBtn = page.locator('.issue-filter-bar button[title="树形"], .issue-filter-bar button[title*="tree"], .issue-filter-bar button:has-text("🌳")').first()
  152 |     if (await treeBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  153 |       await treeBtn.click()
  154 |       await page.waitForTimeout(800)
  155 |     }
  156 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  157 |   })
  158 | 
  159 |   test('switch to calendar view keeps filter bar', async ({ page }) => {
  160 |     await goToProject(page)
  161 |     const calBtn = page.locator('.issue-filter-bar button[title="日历"], .issue-filter-bar button[title*="calendar"], .issue-filter-bar button:has-text("📅")').first()
  162 |     if (await calBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  163 |       await calBtn.click()
  164 |       await page.waitForTimeout(800)
  165 |     }
> 166 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
      |                                                             ^ Error: expect(locator).toBeVisible() failed
  167 |   })
  168 | 
  169 |   test('switch to gantt view keeps filter bar', async ({ page }) => {
  170 |     await goToProject(page)
  171 |     const ganttBtn = page.locator('.issue-filter-bar button[title="甘特"], .issue-filter-bar button[title*="gantt"], .issue-filter-bar button:has-text("📊")').first()
  172 |     if (await ganttBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  173 |       await ganttBtn.click()
  174 |       await page.waitForTimeout(800)
  175 |     }
  176 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  177 |   })
  178 | 
  179 |   test('switch back to list from kanban', async ({ page }) => {
  180 |     await goToProject(page)
  181 |     // Go to kanban first
  182 |     const kanbanBtn = page.locator('.issue-filter-bar button[title="看板"], .issue-filter-bar button:has-text("📌")').first()
  183 |     if (await kanbanBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  184 |       await kanbanBtn.click()
  185 |       await page.waitForTimeout(500)
  186 |     }
  187 |     // Go back to list
  188 |     const listBtn = page.locator('.issue-filter-bar button[title="列表"], .issue-filter-bar button:has-text("📋")').first()
  189 |     if (await listBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  190 |       await listBtn.click()
  191 |       await page.waitForTimeout(500)
  192 |     }
  193 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  194 |   })
  195 | 
  196 |   // ============================================================
  197 |   // 5. Search
  198 |   // ============================================================
  199 |   test('search input accepts text', async ({ page }) => {
  200 |     await goToProject(page)
  201 |     const input = page.locator('.issue-filter-bar input[type="text"]').first()
  202 |     await input.fill('test bug')
  203 |     await expect(input).toHaveValue('test bug')
  204 |   })
  205 | 
  206 |   test('search triggers on Enter', async ({ page }) => {
  207 |     await goToProject(page)
  208 |     const input = page.locator('.issue-filter-bar input[type="text"]').first()
  209 |     await input.fill('login')
  210 |     await input.press('Enter')
  211 |     await page.waitForTimeout(500)
  212 |     await expect(page.locator('body')).toBeVisible()
  213 |   })
  214 | 
  215 |   // ============================================================
  216 |   // 6. i18n - Switch language, filter bar updates
  217 |   // ============================================================
  218 |   test('filter bar labels switch to English', async ({ page }) => {
  219 |     await goToProject(page)
  220 |     // Switch to English
  221 |     const langBtn = page.locator('button:has-text("中"), button:has-text("EN")').first()
  222 |     if (await langBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  223 |       await langBtn.click()
  224 |       await page.waitForTimeout(500)
  225 |     }
  226 |     // Filter bar should still be functional
  227 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 5000 })
  228 |   })
  229 | })
  230 | 
```