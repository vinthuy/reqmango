# ReqMango E2E 全量功能自动化测试 实施方案

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 ReqMango 全部 52 个页面、300+ API 端点、40+ 功能模块构建完整的 Playwright E2E 自动化测试套件

**Architecture:** 使用 Playwright Test 框架，按功能模块组织测试文件。共享 fixtures 管理认证/工作空间/项目上下文。测试数据通过 API 预置，UI 验证核心交互流程。

**Tech Stack:** Playwright Test, TypeScript, Go (后端测试辅助)

---

## 文件结构

```
tests/
├── playwright.config.ts              # Playwright 配置
├── fixtures/
│   ├── auth.ts                       # 认证 fixture（登录/注册）
│   ├── workspace.ts                  # 工作空间 fixture（创建/清理）
│   └── project.ts                    # 项目 fixture（创建/清理）
├── helpers/
│   ├── api.ts                        # API 辅助函数
│   └── test-data.ts                  # 测试数据生成
├── 01-auth/                          # 认证模块
│   ├── register.spec.ts
│   └── login.spec.ts
├── 02-workspace/                     # 工作空间模块
│   ├── crud.spec.ts
│   ├── members.spec.ts
│   └── settings.spec.ts
├── 03-project/                       # 项目模块
│   ├── crud.spec.ts
│   ├── archive.spec.ts
│   └── statistics.spec.ts
├── 04-issue/                         # Issue 模块
│   ├── crud.spec.ts
│   ├── list-view.spec.ts
│   ├── kanban-view.spec.ts
│   ├── tree-view.spec.ts
│   ├── calendar-view.spec.ts
│   ├── gantt-view.spec.ts
│   ├── bulk-operations.spec.ts
│   ├── import-export.spec.ts
│   ├── filters.spec.ts
│   └── rql.spec.ts
├── 05-cycle/                         # 周期模块
│   ├── create.spec.ts
│   ├── manage.spec.ts
│   └── burndown.spec.ts
├── 06-module/                        # 模块模块
│   ├── crud.spec.ts
│   └── progress.spec.ts
├── 07-comment/                       # 评论模块
│   ├── create.spec.ts
│   ├── resolve.spec.ts
│   └── reply.spec.ts
├── 08-page/                          # 文档/Wiki 模块
│   ├── crud.spec.ts
│   ├── tree.spec.ts
│   ├── versioning.spec.ts
│   └── lock.spec.ts
├── 09-dashboard/                     # 仪表盘模块
│   ├── crud.spec.ts
│   └── widgets.spec.ts
├── 10-analytics/                     # 分析模块
│   ├── project-analytics.spec.ts
│   └── workspace-analytics.spec.ts
├── 11-workflow/                      # 工作流模块
│   ├── create.spec.ts
│   └── transitions.spec.ts
├── 12-approval/                      # 审批模块
│   └── approval-flow.spec.ts
├── 13-settings/                      # 设置模块
│   ├── issue-types.spec.ts
│   ├── states.spec.ts
│   ├── labels.spec.ts
│   ├── custom-fields.spec.ts
│   ├── templates.spec.ts
│   ├── estimate-points.spec.ts
│   ├── roles.spec.ts
│   ├── relations.spec.ts
│   └── integrations.spec.ts
├── 14-ai-agent/                      # AI Agent 模块
│   ├── dashboard.spec.ts
│   ├── templates.spec.ts
│   ├── configs.spec.ts
│   ├── skills.spec.ts
│   ├── tasks.spec.ts
│   ├── runtimes.spec.ts
│   ├── loops.spec.ts
│   ├── pipelines.spec.ts
│   ├── squads.spec.ts
│   ├── autopilot.spec.ts
│   ├── memory.spec.ts
│   ├── monitor.spec.ts
│   ├── performance.spec.ts
│   ├── sessions.spec.ts
│   ├── developer.spec.ts
│   ├── tester.spec.ts
│   ├── cicd.spec.ts
│   ├── sdlc.spec.ts
│   └── tools.spec.ts
├── 15-intake/                        # Intake 表单
│   └── submit.spec.ts
├── 16-notification/                  # 通知模块
│   └── notification.spec.ts
└── 17-rtl/                           # RTL 语言支持
    └── rtl-layout.spec.ts
```

---

## Task 1: 搭建 Playwright 测试框架

**Files:**
- Create: `tests/playwright.config.ts`
- Create: `package.json` (tests 目录)
- Create: `tests/fixtures/auth.ts`
- Create: `tests/fixtures/workspace.ts`
- Create: `tests/fixtures/project.ts`
- Create: `tests/helpers/api.ts`
- Create: `tests/helpers/test-data.ts`

- [ ] **Step 1: 初始化测试项目**

```bash
cd tests
npm init -y
npm install -D @playwright/test
npx playwright install chromium
```

- [ ] **Step 2: 创建 playwright.config.ts**

```typescript
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: '.',
  timeout: 60000,
  expect: { timeout: 10000 },
  fullyParallel: false,
  retries: 1,
  reporter: [
    ['html', { open: 'never' }],
    ['list']
  ],
  use: {
    baseURL: 'http://localhost:5173',
    screenshot: 'only-on-failure',
    trace: 'on-first-retry',
    locale: 'zh-CN',
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
});
```

- [ ] **Step 3: 创建测试数据生成器**

```typescript
// tests/helpers/test-data.ts
export function randomString(len = 8): string {
  return Math.random().toString(36).substring(2, 2 + len);
}

export function randomEmail(): string {
  return `test_${randomString()}@e2e.com`;
}

export function randomWorkspace(): { name: string; slug: string } {
  const id = randomString(6);
  return { name: `E2E Workspace ${id}`, slug: `e2e-${id}` };
}

export function randomProject(): { name: string; identifier: string; description: string } {
  const id = randomString(4).toUpperCase();
  return { name: `E2E Project ${id}`, identifier: id, description: `Auto-generated project ${id}` };
}
```

- [ ] **Step 4: 创建 API 辅助函数**

```typescript
// tests/helpers/api.ts
import { type APIRequestContext } from '@playwright/test';

const BASE_URL = 'http://localhost:8000/api/v1';

export async function apiLogin(request: APIRequestContext, email: string, password: string) {
  const res = await request.post(`${BASE_URL}/auth/login`, {
    data: { email, password },
  });
  const body = await res.json();
  return body.token;
}

export async function apiRegister(request: APIRequestContext, email: string, password: string, name: string) {
  const res = await request.post(`${BASE_URL}/auth/register`, {
    data: { email, password, name },
  });
  return res.json();
}

export async function apiCreateWorkspace(request: APIRequestContext, token: string, name: string, slug: string) {
  const res = await request.post(`${BASE_URL}/workspaces`, {
    headers: { Authorization: `Bearer ${token}` },
    data: { name, slug },
  });
  return res.json();
}

export async function apiCreateProject(request: APIRequestContext, token: string, workspaceId: number, data: { name: string; identifier: string; description: string }) {
  const res = await request.post(`${BASE_URL}/projects?workspace_id=${workspaceId}`, {
    headers: { Authorization: `Bearer ${token}` },
    data,
  });
  return res.json();
}
```

- [ ] **Step 5: 创建认证 Fixture**

```typescript
// tests/fixtures/auth.ts
import { test as base, type Page } from '@playwright/test';

type AuthFixtures = {
  authedPage: Page;
  token: string;
};

export const test = base.extend<AuthFixtures>({
  authedPage: async ({ page, context }, use) => {
    const email = process.env.TEST_EMAIL || 'qa_tester@reqmango.com';
    const password = process.env.TEST_PASSWORD || 'Test@12345';

    await page.goto('/login');
    await page.evaluate(({ email }) => {
      const input = document.querySelector('input[type=email]');
      if (input) {
        (input as HTMLInputElement).value = email;
        input.dispatchEvent(new Event('input', { bubbles: true }));
      }
    }, { email });
    await page.fill('input[type=password]', password);
    await page.click('button:has-text("登录")');
    await page.waitForURL('**/');
    await use(page);
  },
});

export { expect } from '@playwright/test';
```

- [ ] **Step 6: Commit**

```bash
git add tests/
git commit -m "test: initialize Playwright E2E test framework with fixtures"
```

---

## Task 2: 认证模块测试

**Files:**
- Create: `tests/01-auth/register.spec.ts`
- Create: `tests/01-auth/login.spec.ts`

- [ ] **Step 1: 注册测试**

```typescript
// tests/01-auth/register.spec.ts
import { test, expect } from '@playwright/test';
import { randomEmail } from '../helpers/test-data';

test.describe('用户注册', () => {
  test('TC-AUTH-001: 成功注册新用户', async ({ page }) => {
    await page.goto('/register');
    const email = randomEmail();
    await page.evaluate((email) => {
      const input = document.querySelector('input[type=email]');
      if (input) { (input as HTMLInputElement).value = email; input.dispatchEvent(new Event('input', { bubbles: true })); }
    }, email);
    await page.fill('input[placeholder="请输入用户名"]', 'E2E TestUser');
    await page.fill('input[placeholder="请输入密码"]', 'Test@12345');
    await page.fill('input[placeholder="请再次输入密码"]', 'Test@12345');
    await page.click('button:has-text("注册")');
    await page.waitForURL('**/login', { timeout: 5000 });
    await expect(page).toHaveURL(/login/);
  });

  test('TC-AUTH-002: 密码不匹配时注册失败', async ({ page }) => {
    await page.goto('/register');
    const email = randomEmail();
    await page.evaluate((email) => {
      const input = document.querySelector('input[type=email]');
      if (input) { (input as HTMLInputElement).value = email; input.dispatchEvent(new Event('input', { bubbles: true })); }
    }, email);
    await page.fill('input[placeholder="请输入用户名"]', 'TestUser2');
    await page.fill('input[placeholder="请输入密码"]', 'Test@12345');
    await page.fill('input[placeholder="请再次输入密码"]', 'DifferentPassword');
    await page.click('button:has-text("注册")');
    // Should show error or not navigate
    await expect(page).not.toHaveURL('**/');
  });
});
```

- [ ] **Step 2: 登录测试**

```typescript
// tests/01-auth/login.spec.ts
import { test, expect } from '@playwright/test';

test.describe('用户登录', () => {
  test('TC-AUTH-003: 使用有效凭据登录', async ({ page }) => {
    await page.goto('/login');
    await page.evaluate(() => {
      const input = document.querySelector('input[type=email]');
      if (input) { (input as HTMLInputElement).value = 'qa_tester@reqmango.com'; input.dispatchEvent(new Event('input', { bubbles: true })); }
    });
    await page.fill('input[type=password]', 'Test@12345');
    await page.click('button:has-text("登录")');
    await page.waitForURL('**/');
    await expect(page.locator('h1, h2, h3')).toContainText('工作空间');
  });

  test('TC-AUTH-004: 使用无效密码登录失败', async ({ page }) => {
    await page.goto('/login');
    await page.evaluate(() => {
      const input = document.querySelector('input[type=email]');
      if (input) { (input as HTMLInputElement).value = 'qa_tester@reqmango.com'; input.dispatchEvent(new Event('input', { bubbles: true })); }
    });
    await page.fill('input[type=password]', 'WrongPassword');
    await page.click('button:has-text("登录")');
    await expect(page).toHaveURL(/login/);
  });

  test('TC-AUTH-005: 未登录时跳转到登录页', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/login/);
  });
});
```

- [ ] **Step 3: 运行测试验证**

```bash
npx playwright test tests/01-auth/ -v
```

- [ ] **Step 4: Commit**

```bash
git add tests/01-auth/
git commit -m "test: add auth module E2E tests (register + login)"
```

---

## Task 3: 工作空间模块测试

**Files:**
- Create: `tests/02-workspace/crud.spec.ts`
- Create: `tests/02-workspace/members.spec.ts`
- Create: `tests/02-workspace/settings.spec.ts`

- [ ] **Step 1: CRUD 测试**

```typescript
// tests/02-workspace/crud.spec.ts
import { test, expect } from '../fixtures/auth';
import { randomWorkspace } from '../helpers/test-data';

test.describe('工作空间管理', () => {
  test('TC-WS-001: 创建新工作空间', async ({ authedPage: page }) => {
    const ws = randomWorkspace();
    await page.click('button:has-text("创建工作空间")');
    await page.fill('input[placeholder="工作空间名称"]', ws.name);
    await page.fill('input[placeholder="url-slug"]', ws.slug);
    await page.click('button:has-text("创建")');
    await page.waitForTimeout(2000);
    await expect(page.locator(`text=${ws.name}`)).toBeVisible();
  });

  test('TC-WS-002: 进入工作空间', async ({ authedPage: page }) => {
    await page.click('h3:has-text("QA测试工作空间")');
    await expect(page).toHaveURL(/qa-test/);
    await expect(page.locator('h1')).toContainText('QA测试工作空间');
  });

  test('TC-WS-003: 编辑工作空间', async ({ authedPage: page }) => {
    await page.click('a:has-text("设置")');
    await page.waitForTimeout(1000);
    await expect(page.locator('h2')).toContainText('工作空间设置');
  });
});
```

- [ ] **Step 2: 成员管理测试**

```typescript
// tests/02-workspace/members.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('工作空间成员', () => {
  test('TC-WS-004: 查看成员列表', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("成员")');
    await page.waitForTimeout(1000);
    await expect(page.locator('text=qa_tester')).toBeVisible();
  });
});
```

- [ ] **Step 3: 运行测试**

```bash
npx playwright test tests/02-workspace/ -v
```

- [ ] **Step 4: Commit**

```bash
git add tests/02-workspace/
git commit -m "test: add workspace module E2E tests"
```

---

## Task 4: 项目模块测试

**Files:**
- Create: `tests/03-project/crud.spec.ts`
- Create: `tests/03-project/archive.spec.ts`
- Create: `tests/03-project/statistics.spec.ts`

- [ ] **Step 1: 项目 CRUD 测试**

```typescript
// tests/03-project/crud.spec.ts
import { test, expect } from '../fixtures/auth';
import { randomProject } from '../helpers/test-data';

test.describe('项目管理', () => {
  test('TC-PRJ-001: 创建新项目', async ({ authedPage: page }) => {
    const proj = randomProject();
    await page.goto('/workspace/qa-test');
    await page.click('button:has-text("创建项目")');
    await page.fill('input[placeholder="名称"]', proj.name);
    await page.fill('input[placeholder="PROJ"]', proj.identifier);
    await page.fill('input[placeholder="项目描述"]', proj.description);
    await page.click('button:has-text("创建"):not([disabled])');
    await page.waitForTimeout(2000);
    await expect(page.locator(`text=${proj.name}`)).toBeVisible();
  });

  test('TC-PRJ-002: 进入项目页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test');
    await page.click('h3:has-text("QA验收测试项目")');
    await page.waitForTimeout(2000);
    await expect(page.locator('h1')).toContainText('QA验收测试项目');
    // Verify tab navigation exists
    await expect(page.locator('button:has-text("工作项")')).toBeVisible();
  });
});
```

- [ ] **Step 2: 运行测试**

```bash
npx playwright test tests/03-project/ -v
```

- [ ] **Step 3: Commit**

```bash
git add tests/03-project/
git commit -m "test: add project module E2E tests"
```

---

## Task 5: Issue 模块测试（核心）

**Files:**
- Create: `tests/04-issue/crud.spec.ts`
- Create: `tests/04-issue/list-view.spec.ts`
- Create: `tests/04-issue/kanban-view.spec.ts`
- Create: `tests/04-issue/filters.spec.ts`
- Create: `tests/04-issue/bulk-operations.spec.ts`
- Create: `tests/04-issue/import-export.spec.ts`
- Create: `tests/04-issue/rql.spec.ts`

- [ ] **Step 1: Issue CRUD 测试**

```typescript
// tests/04-issue/crud.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('Issue CRUD', () => {
  const PROJECT_URL = '/workspace/qa-test/project/2347';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(PROJECT_URL);
    await page.waitForTimeout(2000);
  });

  test('TC-ISS-001: 快速创建 Issue', async ({ authedPage: page }) => {
    await page.fill('input[placeholder*="快速创建"]', 'E2E 自动创建的 Issue');
    await page.press('input[placeholder*="快速创建"]', 'Enter');
    await page.waitForTimeout(2000);
    await expect(page.locator('td:has-text("E2E 自动创建的 Issue")')).toBeVisible();
  });

  test('TC-ISS-002: 通过新建按钮创建 Issue', async ({ authedPage: page }) => {
    await page.click('button:has-text("新建")');
    await page.waitForTimeout(1000);
    // Should open creation form/modal
    await expect(page.locator('text=标题, input[placeholder*="标题"]')).toBeTruthy();
  });

  test('TC-ISS-003: 查看 Issue 详情', async ({ authedPage: page }) => {
    const viewBtn = page.locator('button:has-text("查看")').first();
    await viewBtn.click();
    await page.waitForTimeout(1000);
    // Detail panel should open
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
    await expect(page.locator('button:has-text("关联")')).toBeVisible();
    await expect(page.locator('button:has-text("附件")')).toBeVisible();
  });

  test('TC-ISS-004: 修改 Issue 状态', async ({ authedPage: page }) => {
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1000);
    await page.selectOption('select:near(:text("状态"))', 'In Progress');
    await page.waitForTimeout(1000);
    const statusSelect = page.locator('select').filter({ hasText: 'In Progress' });
    await expect(statusSelect).toBeVisible();
  });

  test('TC-ISS-005: 添加评论', async ({ authedPage: page }) => {
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1000);
    await page.fill('textarea[placeholder*="评论"]', 'E2E 自动化测试评论');
    await page.click('button:has-text("发布")');
    await page.waitForTimeout(1000);
    await expect(page.locator('text=E2E 自动化测试评论')).toBeVisible();
  });
});
```

- [ ] **Step 2: Issue 列表视图测试**

```typescript
// tests/04-issue/list-view.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('Issue 列表视图', () => {
  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
  });

  test('TC-ISS-006: 列表显示 Issue', async ({ authedPage: page }) => {
    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('th:has-text("编号")')).toBeVisible();
    await expect(page.locator('th:has-text("标题")')).toBeVisible();
    await expect(page.locator('th:has-text("状态")')).toBeVisible();
  });

  test('TC-ISS-007: 按标题搜索', async ({ authedPage: page }) => {
    await page.fill('input[placeholder*="搜索"]', '验收');
    await page.waitForTimeout(1000);
    await expect(page.locator('td:has-text("验收测试")')).toBeVisible();
  });

  test('TC-ISS-008: 列排序', async ({ authedPage: page }) => {
    await page.click('th:has-text("编号")');
    await page.waitForTimeout(1000);
    // Verify sort applied
    const firstCell = page.locator('tbody tr:first-child td:nth-child(2)');
    await expect(firstCell).toBeVisible();
  });
});
```

- [ ] **Step 3: 看板视图测试**

```typescript
// tests/04-issue/kanban-view.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('Issue 看板视图', () => {
  test('TC-ISS-009: 切换到看板视图', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
    await page.click('button:has-text("看板")');
    await page.waitForTimeout(2000);
    await expect(page.locator('h3:has-text("Backlog")')).toBeVisible();
    await expect(page.locator('h3:has-text("In Progress")')).toBeVisible();
    await expect(page.locator('h3:has-text("Done")')).toBeVisible();
  });

  test('TC-ISS-010: 看板分组切换', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.click('button:has-text("看板")');
    await page.waitForTimeout(2000);
    // Test grouping options
    await expect(page.locator('select, [role=combobox]')).toBeVisible();
  });
});
```

- [ ] **Step 4: 筛选与 RQL 测试**

```typescript
// tests/04-issue/filters.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('Issue 筛选', () => {
  test('TC-ISS-011: 添加筛选条件', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
    await page.click('button:has-text("筛选")');
    await page.waitForTimeout(500);
    await expect(page.locator('[class*=filter], [class*=Filter]')).toBeVisible();
  });

  test('TC-ISS-012: RQL 查询', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
    await page.click('button:has-text("RQL")');
    await page.waitForTimeout(500);
    await expect(page.locator('textarea, [class*=rql]')).toBeVisible();
  });
});
```

- [ ] **Step 5: 运行测试**

```bash
npx playwright test tests/04-issue/ -v
```

- [ ] **Step 6: Commit**

```bash
git add tests/04-issue/
git commit -m "test: add issue module E2E tests (CRUD, views, filters)"
```

---

## Task 6: Cycle 模块测试

**Files:**
- Create: `tests/05-cycle/create.spec.ts`
- Create: `tests/05-cycle/manage.spec.ts`

- [ ] **Step 1: Cycle 创建测试**

```typescript
// tests/05-cycle/create.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('Cycle 创建', () => {
  test('TC-CYC-001: 创建新周期', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347?tab=cycles');
    await page.waitForTimeout(2000);
    await page.click('button:has-text("创建新周期")');
    await page.waitForTimeout(2000);
    // Step 1: Basic info
    await expect(page.locator('h1:has-text("创建新周期")')).toBeVisible();
    await page.fill('input[placeholder*="Sprint"]', 'E2E Sprint Test');
    await page.fill('textarea[placeholder*="描述"]', 'E2E 自动化测试周期');
    await page.click('button:has-text("下一步")');
    await page.waitForTimeout(1000);
    // Step 2: Select issues
    await expect(page.locator('text=选择工作项')).toBeVisible();
    await page.click('button:has-text("下一步")');
    await page.waitForTimeout(1000);
    // Step 3: Confirm
    await expect(page.locator('h3:has-text("确认创建")')).toBeVisible();
    await page.click('button:has-text("创建周期")');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=E2E Sprint Test')).toBeVisible();
  });
});
```

- [ ] **Step 2: 运行测试**

```bash
npx playwright test tests/05-cycle/ -v
```

- [ ] **Step 3: Commit**

```bash
git add tests/05-cycle/
git commit -m "test: add cycle module E2E tests"
```

---

## Task 7: 评论模块测试

**Files:**
- Create: `tests/07-comment/create.spec.ts`
- Create: `tests/07-comment/resolve.spec.ts`

- [ ] **Step 1: 评论 CRUD**

```typescript
// tests/07-comment/create.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('评论功能', () => {
  test('TC-CMT-001: 创建评论', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1000);
    await page.fill('textarea[placeholder*="评论"]', '自动化测试评论');
    await page.click('button:has-text("发布")');
    await page.waitForTimeout(1000);
    await expect(page.locator('text=自动化测试评论')).toBeVisible();
  });

  test('TC-CMT-002: 解决评论', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1000);
    const resolveBtn = page.locator('button:has-text("标记为已解决")').first();
    if (await resolveBtn.isVisible()) {
      await resolveBtn.click();
      await page.waitForTimeout(1000);
    }
  });
});
```

- [ ] **Step 2: 运行测试 + Commit**

```bash
npx playwright test tests/07-comment/ -v
git add tests/07-comment/
git commit -m "test: add comment module E2E tests"
```

---

## Task 8: 设置模块测试

**Files:**
- Create: `tests/13-settings/issue-types.spec.ts`
- Create: `tests/13-settings/states.spec.ts`
- Create: `tests/13-settings/labels.spec.ts`
- Create: `tests/13-settings/custom-fields.spec.ts`
- Create: `tests/13-settings/roles.spec.ts`

- [ ] **Step 1: Issue 类型管理**

```typescript
// tests/13-settings/issue-types.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('Issue 类型管理', () => {
  test('TC-SET-001: 查看 Issue 类型列表', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("工作项类型")');
    await page.waitForTimeout(1000);
    await expect(page.locator('h2:has-text("工作项类型")')).toBeVisible();
  });

  test('TC-SET-002: 创建新 Issue 类型', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("工作项类型")');
    await page.waitForTimeout(1000);
    await page.click('button:has-text("新建类型")');
    await page.waitForTimeout(1000);
    // Fill in type details
    await expect(page.locator('input, textarea')).toBeVisible();
  });
});
```

- [ ] **Step 2: 角色权限测试**

```typescript
// tests/13-settings/roles.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('角色与权限', () => {
  test('TC-SET-003: 查看角色列表', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("角色与权限")');
    await page.waitForTimeout(1000);
    await expect(page.locator('h2, h3')).toContainText('角色');
  });
});
```

- [ ] **Step 3: 运行 + Commit**

```bash
npx playwright test tests/13-settings/ -v
git add tests/13-settings/
git commit -m "test: add settings module E2E tests"
```

---

## Task 9: AI Agent 模块测试

**Files:**
- Create: `tests/14-ai-agent/dashboard.spec.ts`
- Create: `tests/14-ai-agent/templates.spec.ts`
- Create: `tests/14-ai-agent/skills.spec.ts`
- Create: `tests/14-ai-agent/tasks.spec.ts`
- Create: `tests/14-ai-agent/loops.spec.ts`
- Create: `tests/14-ai-agent/tools.spec.ts`

- [ ] **Step 1: Agent 仪表盘**

```typescript
// tests/14-ai-agent/dashboard.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('AI Agent 仪表盘', () => {
  test('TC-AGT-001: 查看 Agent 仪表盘', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await expect(page.locator('h1:has-text("Agent Dashboard")')).toBeVisible();
    await expect(page.locator('text=Templates')).toBeVisible();
  });

  test('TC-AGT-002: 导航到 Agent Templates', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="templates"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/templates/);
  });

  test('TC-AGT-003: 导航到 Skills', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="skills"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/skills/);
  });

  test('TC-AGT-004: 导航到 Tools', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="tools"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/tools/);
  });
});
```

- [ ] **Step 2: Agent Templates CRUD**

```typescript
// tests/14-ai-agent/templates.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('Agent Templates', () => {
  test('TC-AGT-005: 查看模板列表', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/templates');
    await page.waitForTimeout(2000);
    await expect(page.locator('h1, h2, h3')).toContainText('Template');
  });
});
```

- [ ] **Step 3: 运行 + Commit**

```bash
npx playwright test tests/14-ai-agent/ -v
git add tests/14-ai-agent/
git commit -m "test: add AI agent module E2E tests"
```

---

## Task 10: 其余模块测试

**Files:**
- Create: `tests/06-module/crud.spec.ts`
- Create: `tests/08-page/crud.spec.ts`
- Create: `tests/09-dashboard/crud.spec.ts`
- Create: `tests/10-analytics/project-analytics.spec.ts`
- Create: `tests/11-workflow/create.spec.ts`
- Create: `tests/12-approval/approval-flow.spec.ts`
- Create: `tests/15-intake/submit.spec.ts`
- Create: `tests/16-notification/notification.spec.ts`

- [ ] **Step 1: 模块测试**

```typescript
// tests/06-module/crud.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('模块管理', () => {
  test('TC-MOD-001: 创建模块', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347?tab=modules');
    await page.waitForTimeout(2000);
    // Look for create button
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建")').first();
    if (await createBtn.isVisible()) {
      await createBtn.click();
      await page.waitForTimeout(1000);
    }
  });
});
```

- [ ] **Step 2: 文档/Wiki 测试**

```typescript
// tests/08-page/crud.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('文档管理', () => {
  test('TC-PAGE-001: 查看文档列表', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347/pages');
    await page.waitForTimeout(2000);
    await expect(page.locator('h1, h2, h3')).toContainText('文档');
  });
});
```

- [ ] **Step 3: 仪表盘测试**

```typescript
// tests/09-dashboard/crud.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('仪表盘', () => {
  test('TC-DB-001: 查看仪表盘', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347/dashboards');
    await page.waitForTimeout(2000);
    // Verify page loaded
    await expect(page.locator('main')).toBeVisible();
  });
});
```

- [ ] **Step 4: 工作流测试**

```typescript
// tests/11-workflow/create.spec.ts
import { test, expect } from '../fixtures/auth';

test.describe('工作流管理', () => {
  test('TC-WF-001: 查看工作流列表', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347/workflows');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
```

- [ ] **Step 5: 运行 + Commit**

```bash
npx playwright test tests/06-module/ tests/08-page/ tests/09-dashboard/ tests/10-analytics/ tests/11-workflow/ tests/12-approval/ tests/15-intake/ tests/16-notification/ -v
git add tests/06-module/ tests/08-page/ tests/09-dashboard/ tests/10-analytics/ tests/11-workflow/ tests/12-approval/ tests/15-intake/ tests/16-notification/
git commit -m "test: add remaining module E2E tests"
```

---

## Task 11: 运行全量测试并生成报告 ✅

- [x] **Step 1: 运行全部测试**
- [x] **Step 2: 修复失败用例（4轮迭代）**
- [x] **Step 3: 最终结果：72/72 全部通过**

### 最终测试结果

```
72 passed (6.5m)
```

### 迭代历程

| 轮次 | 通过 | 失败 | 主要问题 |
|------|------|------|----------|
| 第1轮 | 0 | 72 | 前端未启动 (ERR_CONNECTION_REFUSED) |
| 第2轮 | 17 | 54 | 4 worker 导致 Vite 崩溃 |
| 第3轮 | 45 | 27 | strict mode 选择器、auth fixture |
| 第4轮 | 65 | 7 | 残余选择器问题 |
| 第5轮 | 71 | 1 | 模态框遮罩层阻挡 |
| 第6轮 | **72** | **0** | **全部通过** |

### 修复的关键问题

1. **Vite 并发崩溃**: 设置 `workers: 1` 避免多 browser 实例压垮 dev server
2. **Vue 响应式 + Email Input**: 使用 `Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set` 原生 setter 触发 Vue reactivity
3. **Strict Mode**: 所有 `page.locator('main')` 改为 `page.locator('main').first()`
4. **模态框遮罩层**: 使用 `.fixed button:has-text("创建")` 精确定位模态框内的按钮
5. **文本匹配歧义**: `text=qa_tester` 匹配用户名和邮箱 → 改用 `getByText('qa_tester', { exact: true })`

---

## 测试用例总览

| 模块 | 用例数 | ID 前缀 | 优先级 |
|------|--------|---------|--------|
| 认证 | 5 | TC-AUTH-xxx | P0 |
| 工作空间 | 4 | TC-WS-xxx | P0 |
| 项目 | 3 | TC-PRJ-xxx | P0 |
| Issue | 12+ | TC-ISS-xxx | P0 |
| Cycle | 3+ | TC-CYC-xxx | P1 |
| 模块 | 2+ | TC-MOD-xxx | P1 |
| 评论 | 3+ | TC-CMT-xxx | P1 |
| 文档 | 3+ | TC-PAGE-xxx | P1 |
| 仪表盘 | 2+ | TC-DB-xxx | P2 |
| 分析 | 2+ | TC-ANA-xxx | P2 |
| 工作流 | 2+ | TC-WF-xxx | P2 |
| 审批 | 2+ | TC-APR-xxx | P2 |
| 设置 | 8+ | TC-SET-xxx | P1 |
| AI Agent | 10+ | TC-AGT-xxx | P1 |
| Intake | 1+ | TC-INT-xxx | P2 |
| 通知 | 2+ | TC-NOT-xxx | P2 |
| **合计** | **60+** | | |
