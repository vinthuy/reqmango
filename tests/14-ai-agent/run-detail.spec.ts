import { test, expect, type APIRequestContext } from '../fixtures/auth';

/**
 * Coverage gap: the agent run-detail routes had no E2E coverage at all.
 *
 *   /workspaces/:wsParam/agents/loops/runs/:runId      -> views/agents/LoopRunDetail.vue
 *   /workspaces/:wsParam/agents/pipelines/runs/:runId  -> views/agents/PipelineRunDetail.vue
 *
 * Both components resolve the workspace from a NUMERIC id in the path
 * (`window.location.pathname.match(/\/workspaces\/(\d+)/)`), *not* from the
 * workspace slug, so these specs navigate with the real numeric workspace id.
 *
 * Run records are produced through the public API: `POST .../loops/{id}/start`
 * and `POST .../pipelines/{id}/run` both persist the run row synchronously and
 * only execute in a background goroutine, so the run is queryable immediately.
 * Budgets are deliberately tiny so the background execution finishes quickly.
 */

const API_BASE = process.env.E2E_API_BASE || 'http://localhost:8000/api/v1';
const WS_SLUG = process.env.E2E_WORKSPACE_SLUG || 'qa-test';
const EMAIL = process.env.TEST_EMAIL || 'qa_tester@reqmango.com';
const PASSWORD = process.env.TEST_PASSWORD || 'Test@12345';

let token = '';
let wsId = 0;
let loopRunId = 0;
let loopGoal = '';
let pipelineRunId = 0;

async function authHeaders(request: APIRequestContext): Promise<Record<string, string>> {
  if (!token) {
    const res = await request.post(`${API_BASE}/auth/login`, {
      data: { email: EMAIL, password: PASSWORD },
    });
    const body = await res.json();
    token = body.access_token;
  }
  return { Authorization: `Bearer ${token}` };
}

test.describe('Agent Run 详情页 (Loop / Pipeline)', () => {
  test.beforeAll(async ({ request }) => {
    const headers = await authHeaders(request);

    // Resolve the numeric workspace id from the slug the rest of the suite uses.
    const wsRes = await request.get(`${API_BASE}/workspaces`, { headers });
    const wsBody = await wsRes.json();
    const spaces = Array.isArray(wsBody) ? wsBody : wsBody?.data ?? [];
    const ws = spaces.find((w: any) => w.slug === WS_SLUG);
    if (!ws) throw new Error(`E2E workspace "${WS_SLUG}" not found`);
    wsId = ws.id;

    // --- Loop run -------------------------------------------------------
    loopGoal = `E2E loop goal ${Date.now()}`;
    const loopRes = await request.post(`${API_BASE}/workspaces/${wsId}/loops`, {
      headers,
      data: {
        name: `E2E Run Detail Loop ${Date.now()}`,
        description: 'created by run-detail e2e',
        loop_def: { goal: loopGoal, max_iterations: 1, max_tokens: 500, max_duration_sec: 30 },
      },
    });
    if (loopRes.ok()) {
      const loop = await loopRes.json();
      const startRes = await request.post(`${API_BASE}/workspaces/${wsId}/loops/${loop.id}/start`, {
        headers,
      });
      if (startRes.ok()) {
        loopRunId = (await startRes.json()).id;
      }
    }

    // --- Pipeline run ---------------------------------------------------
    const pipelineRes = await request.post(`${API_BASE}/workspaces/${wsId}/pipelines`, {
      headers,
      data: {
        name: `E2E Run Detail Pipeline ${Date.now()}`,
        description: 'created by run-detail e2e',
        pipeline_def: {
          name: 'e2e-run-detail',
          pipeline: {
            mode: 'sequential',
            stages: [{ name: 'E2E Stage', agent: 'planner', type: 'planner' }],
          },
          budget: { max_tokens: 500 },
        },
      },
    });
    if (pipelineRes.ok()) {
      const pipeline = await pipelineRes.json();
      const runRes = await request.post(`${API_BASE}/workspaces/${wsId}/pipelines/${pipeline.id}/run`, {
        headers,
      });
      if (runRes.ok()) {
        pipelineRunId = (await runRes.json()).id;
      }
    }
  });

  // === Loop run detail =================================================

  test('TC-RUN-001: Loop Run 详情页渲染运行概要', async ({ authedPage: page }) => {
    test.skip(!loopRunId, 'loop run was not created');
    await page.goto(`/workspaces/${wsId}/agents/loops/runs/${loopRunId}`);
    await page.waitForTimeout(2000);

    await expect(page.locator('h1')).toContainText(`Loop Run #${loopRunId}`);
    // The goal is rendered from the fetched run detail.
    await expect(page.getByText(loopGoal, { exact: false }).first()).toBeVisible();
    await expect(page.locator('text=Iterations').first()).toBeVisible();
  });

  test('TC-RUN-002: Loop Run 详情页显示预算用量', async ({ authedPage: page }) => {
    test.skip(!loopRunId, 'loop run was not created');
    await page.goto(`/workspaces/${wsId}/agents/loops/runs/${loopRunId}`);
    await page.waitForTimeout(2000);

    // BudgetGauge renders token/iteration progress for the run.
    await expect(
      page.locator('text=/token/i').or(page.locator('text=/iteration/i')).first()
    ).toBeVisible();
  });

  test('TC-RUN-003: Loop Run 详情页未知 run 优雅降级', async ({ authedPage: page }) => {
    await page.goto(`/workspaces/${wsId}/agents/loops/runs/99999999`);
    await page.waitForTimeout(2000);

    // Page shell must still render (route + component mount) instead of crashing.
    await expect(page.locator('h1')).toContainText('Loop Run #99999999');
    await expect(page.locator('text=← Back').first()).toBeVisible();
  });

  // === Pipeline run detail =============================================

  test('TC-RUN-004: Pipeline Run 详情页渲染运行概要', async ({ authedPage: page }) => {
    test.skip(!pipelineRunId, 'pipeline run was not created');
    await page.goto(`/workspaces/${wsId}/agents/pipelines/runs/${pipelineRunId}`);
    await page.waitForTimeout(2000);

    await expect(page.locator('h1')).toContainText(`Pipeline Run #${pipelineRunId}`);
    await expect(
      page.locator('text=/Tokens/i').or(page.locator('text=Stage Results')).first()
    ).toBeVisible();
  });

  test('TC-RUN-005: Pipeline Run 详情页阶段结果区域', async ({ authedPage: page }) => {
    test.skip(!pipelineRunId, 'pipeline run was not created');
    await page.goto(`/workspaces/${wsId}/agents/pipelines/runs/${pipelineRunId}`);
    await page.waitForTimeout(2000);

    // After the background execution settles the stage results section renders.
    await expect(page.locator('text=Stage Results').first()).toBeVisible();
  });

  test('TC-RUN-006: Pipeline Run 详情页未知 run 优雅降级', async ({ authedPage: page }) => {
    await page.goto(`/workspaces/${wsId}/agents/pipelines/runs/99999999`);
    await page.waitForTimeout(2000);

    await expect(page.locator('h1')).toContainText('Pipeline Run #99999999');
    await expect(page.locator('text=← Back').first()).toBeVisible();
  });

  // === Responsive ======================================================

  test('TC-RUN-007: Run 详情页响应式', async ({ authedPage: page }) => {
    await page.goto(`/workspaces/${wsId}/agents/loops/runs/99999999`);
    await page.waitForTimeout(1500);

    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('h1')).toContainText('Loop Run #');

    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('h1')).toContainText('Loop Run #');
  });
});
