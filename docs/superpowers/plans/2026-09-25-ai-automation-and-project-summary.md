# AI 自动化模板 + 项目总结存 Page — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans`. Steps use checkbox (`- [ ]`) syntax.

**Goal:** 在项目管理路径上增强 **B1 AI 自动化模板** 与 **C1 项目级 AI 总结→Page**；不引入 Harness/Loop/多 Agent 流水线。

**Architecture:** 复用 `ensurePMAgents`、`dispatch_agent`、`automationTemplates`、`sprintPlan`/`analyzeWithAI`、`createPage`。以前端模板与入口为主，少造后端。

**Tech Stack:** Vue 3、现有 Automation + AI API。

**Spec / Direction:** `docs/superpowers/specs/2026-09-25-ai-project-management-redesign.md`（§9 已取消 Phase3 后续编排增强；本 plan = 下一增量）  
**Acceptance:** `docs/dev/acceptance/2026-09-25-ai-pm-phase1-2.md`（先跑通再合本 plan 大功能，或并行修验收缺陷）

## Global Constraints

- 禁止 独立 Agent 平台 / MCP / Agent 市场 / Loop 控制台叙事。  
- 自动化模板 trigger 必须点分：`issue.created` 等。  
- AI 写操作：模板动作以 `dispatch_agent` / `add_comment` 为主；无用户确认不静默改大量业务字段。  
- C1：总结可存 Page；不强制定时调度（定时留给以后）。  
- 中英文 i18n；小步 commit。

---

## 文件结构

| 区域 | 文件 | 动作 |
|------|------|------|
| B1 模板 | `ProjectSettings.vue`、locales | 新增 2–3 个 AI 模板；模板区常显（不限空列表） |
| B1 可选 | `AutomationManager.vue`（若仍用） | 点分 trigger 对齐 |
| C1 | `Project.vue` 或 Reports 区 | 「项目 AI 总结」+ 存 Page |
| C1 API | `ai.ts` | 必要时 `analyzeWithAI` 无 issue / 或 sprintPlan(project) |
| 文档 | pipeline / active / 本 plan | 进度 |

---

## Task 0（并行优先）: Phase 1–2 验收执行

**Files:** `docs/dev/acceptance/2026-09-25-ai-pm-phase1-2.md`

- [x] **Step 1:** 按清单手测，勾选结果。  
- [x] **Step 2:** P0/P1 缺陷在本增量分支修掉再合；P2 可记 backlog。  
- [x] **Step 3:** 结论写入清单「签字」；更新 `pipeline-status` Implement 状态。

---

## Task 1: B1 — 更多 AI 自动化模板

**Files:**
- Modify: `frontend/src/views/ProjectSettings.vue`
- Modify: `frontend/src/locales/zh-CN.json`、`en-US.json`

**模板（至少新增 2 个，均 `requiresPMAgents: true`）：**

| name key | trigger | agent | 动作 |
|----------|---------|-------|------|
| `riskOnCreate`（或 high priority） | `issue.created` + condition priority=urgent/high（若条件引擎支持；否则无条件） | `交付风险` | `dispatch_agent` |
| `specDraftOnCreate` | `issue.created` | `Spec 草稿` | `dispatch_agent`（prompt：写需求提纲评论） |

保留现有 `intakeTriage` → `请求分诊`。

- [x] **Step 1:** `applyTemplate` 已有 ensure + 按 name 解析 agent id；新模板复用。  
- [x] **Step 2:** 模板卡片在 **有/无已有规则时都可展示**（修「仅空列表显示」若仍存在）。  
- [x] **Step 3:** i18n 名称/描述。  
- [ ] **Step 4:** 手测：安装 Agent → 应用两模板 → 建 Issue 有对应 Agent 活动。  
- [ ] **Step 5:** Commit `feat(ai): add more AI automation templates`

---

## Task 2: B1 — 动作/触发卫生（小）

**Files:**
- Modify: `frontend/src/components/AutomationManager.vue`（若存在 underscore trigger）
- Grep: `issue_created|comment_added|state_changed` 在 frontend 自动化相关文件

- [x] **Step 1:** 凡用户可新建的规则 UI，写入点分 trigger。  
- [x] **Step 2:** 不强制迁移库内旧规则（文档一句说明即可）。  
- [ ] **Step 3:** Commit `fix(ai): align automation UI triggers with issue.created`

---

## Task 3: C1 — 项目 AI 总结 → Page

**Files:**
- Modify: `frontend/src/views/Project.vue`（Reports tab 或顶栏操作）**或** 独立轻量面板组件  
- Use: `analyzeWithAI(projectId, 0)` 若后端要求 issue_id — 改调 `sprintPlan(projectId)`（无 cycle）或现有项目级 analyze  
- Use: `createPage` 同 CycleDetail `saveSummaryAsPage`

**UI:**
- 入口文案：`项目 AI 总结` / `Project AI summary`  
- 结果面板 +「存为 Page」  
- 不替换现有 Reports/Metrics；只加一块 Capture

- [x] **Step 1:** 选 API：优先 `POST …/ai/sprint-plan`（无 cycle_id = 项目视角）或 `analyze`；与 Cycle 总结区分标题。  
- [x] **Step 2:** 实现按钮 + 面板 + 存 Page。  
- [x] **Step 3:** i18n。  
- [x] **Step 4:** 手测：项目页出总结 → Page 列表可见。  
- [x] **Step 5:** Commit `feat(ai): add project AI summary save-as-page`

---

## Task 4: 文档收尾

- [x] 更新 `docs/dev/pipeline-status.md`、`docs/dev/active/README.md`  
- [ ] Commit `docs: track AI automation templates and project summary plan`

---

## 验收（本增量）

| # | 标准 |
|---|------|
| 1 | 至少 3 个 AI 相关自动化模板可用（含原分诊） |
| 2 | 模板区不依赖「零规则」才显示 |
| 3 | 项目页可一键总结并存 Page |
| 4 | Phase 1–2 验收清单已执行（或明确遗留缺陷） |
| 5 | 无 Loop/Harness 产品入口回潮 |

---

## 明确不做

- Harness / Loop / 对抗评审 / Squad 流水线  
- 自然语言生成自动化规则（B2）  
- Dashboard AI widget / 定时出报（C2）  
- Slack / Marketplace  
