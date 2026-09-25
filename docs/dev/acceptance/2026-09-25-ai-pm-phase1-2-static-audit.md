# Phase 1–2 Static Acceptance Audit

> **Date**: 2026-09-25  
> **Branch**: `master`  
> **Method**: Code wiring only — no UI hand-test results invented.  
> **Source checklist**: `docs/dev/acceptance/2026-09-25-ai-pm-phase1-2.md`  
> **Verdicts**: `WIRED` = path exists end-to-end in code · `PARTIAL` = core present but entry/error/edge gap · `MISSING` = no concrete wiring found

---

## Cross-cutting: automation trigger notation

| Check | Evidence | Verdict |
|-------|----------|---------|
| Templates use dotted triggers | `ProjectSettings.vue` `automationTemplates[].trigger` = `'issue.created'` / `'comment.added'` / `'issue.state_changed'` | **WIRED** |
| Persist as dotted type | `applyTemplate` → `trigger_type: JSON.stringify({ type: template.trigger })` | **WIRED** |
| Backend event name | `IssueService.NotifyIssueCreated` → `runAutomations(..., "issue.created", ...)` | **WIRED** |

**Conclusion**: Automation template triggers are dotted (`issue.created`), not underscored.

---

## Phase 1

### P1-1 — 侧栏无 Agents 主入口 → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| Sidebar nav | `frontend/src/components/AppSidebar.vue` — `navItems` (projects / initiatives / analytics / settings only; no agents) |
| TopBar | `frontend/src/components/TopBar.vue` — settings paths only; no Agents item |
| Advanced entry | `frontend/src/components/AISettingsPanel.vue` — `advancedAgentsPath` → `/workspace/:slug/agents` or `/workspaces/:id/agents`; link `t('ai.advancedAgentsLink')` |
| Host | `frontend/src/views/WorkspaceSettings.vue` mounts `AISettingsPanel` |
| Routes | `frontend/src/router/index.ts` — `/workspace/:slug/agents`, `/workspaces/:wsParam/agents` |

**Hand-test risk**: Testers may miss the advanced link under Workspace Settings → AI; sidebar absence alone is not enough to find `/agents`.

---

### P1-2 — Copilot Ask\|Build + 上下文 → **PARTIAL**

| Layer | File + symbol |
|-------|----------------|
| Modes | `frontend/src/components/AICopilot.vue` — tabs Ask\|Build only (`CopilotMode = 'ask' \| 'build'`); comment `Tab Bar: Ask \| Build only` |
| Context bar | same — `contextLabel`, `clearIssueContext` / project-only via `t('ai.contextProjectOnly')` |
| Issue wiring | `frontend/src/views/IssueDetail.vue` — `<AICopilot :issue-id="issueId" :issue-label="issueCopilotLabel" …>` |
| Ctrl+J / FAB on **Project** | `frontend/src/views/Project.vue` — keydown Ctrl/Meta+J + FAB button |
| Ctrl+J / FAB on **Issue** | **Not found** in `IssueDetail.vue` — open only via AI tab (`showAICopilot = true`) |

**Hand-test risk**: On Issue detail route, Ctrl+J / FAB will not open Copilot; must use AI tab “Copilot” button. Context bar only shows current Issue after that open path.

---

### P1-3 — Ask 只读提示 → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| UI | `AICopilot.vue` — `v-if="mode === 'ask'"` → `{{ t('ai.askReadOnlyHint') }}` |
| i18n | `frontend/src/locales/zh-CN.json` / `en-US.json` — `ai.askReadOnlyHint` |

**Hand-test risk**: Hint is copy-only; live Ask may still call backend tools depending on chat mode enforcement (not verified as hard client lock).

---

### P1-4 — 指派 Agent 持久 → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| Sidebar UI | `IssuePropertySidebar.vue` — `AgentSelector`, primary `emitAssign` / Assign button, secondary outline Dispatch (`dispatch-agent`) |
| Status display | same — `agentStatus.agent_name` + `task_status` |
| Parent handlers | `IssueDetail.vue` — `assignAgent` / `unassignAgent` / `dispatchAgent` / `loadAgentStatus` |
| API | `frontend/src/api/issue-agent.ts` — `assign`, `getStatus` |
| Backend | `backend/internal/router/router.go` — `POST …/assign-agent`, `GET …/agent-status` |

**Hand-test risk**: Frontend `unassign` calls `DELETE /issues/:id/unassign-agent` but router exposes `DELETE /:issueId/assign-agent` — unassign may 404; assign+refresh path should still work.

---

### P1-5 — 评论 @Agent → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| Picker | `CommentList.vue` — `allMentionCandidates` = members + agents; `kind === 'agent'` → AI badge; `agentApi.list` |
| Backend mention | `comment_service.go` — `parseMentions` + `agentClient.HandleMention` |
| Agent handler | `backend/internal/ai/service/agent_service.go` — `HandleMention` |

**Hand-test risk**: Reply/activity needs LLM + matching agent `name`; without Key, mention may log activity but no useful reply.

---

### P1-6 — 活动流 Agent 审计 → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| Activity tab | `IssueTabActivity.vue` — Agent section + `<AgentAuditLog :workspace-id :issue-id>` |
| Audit UI | `AgentAuditLog.vue` — filters by agent/action; filters list by `act.issue_id === props.issueId` |
| API | `agentApi.listWorkspaceActivity` |

**Hand-test risk**: Empty until assign/dispatch/mention/automation has written activity rows.

---

### P1-7 — Intake 分诊入口 → **PARTIAL**

| Layer | File + symbol |
|-------|----------------|
| Settings nav | `ProjectSettings.vue` — section `triage` + `<TriagePanel :project-id>` |
| Panel | `TriagePanel.vue` — list pending + `analyzeAI` → `POST …/intake/:id/ai-analyze` |
| Gap | `analyzeAI` uses `catch (_) {}` — failures leave no visible error |

**Hand-test risk**: Failed AI analyze looks like a no-op (no toast/error), so “明确错误” expectation may fail even though the button is wired.

---

### P1-8 — 创建前去重 → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| Quick create | `QuickCreateInput.vue` — `runDuplicateCheck` → `issueApi.checkDuplicates`; button `createAnyway` when `duplicates.length > 0` |
| Full create | `IssueCreate.vue` — same pattern |
| API | `frontend/src/api/issue.ts` — `checkDuplicates` |
| Backend | `router.go` — `POST|GET …/issues/duplicate-check`; `issue_handler.CheckDuplicates` |

**Hand-test risk**: Debounce 400ms; very short/unique titles may show no warning (expected). Create is never blocked by duplicates.

---

## Phase 2

### P2-1 — 安装开箱 Agent → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| UI | `ProjectSettings.vue` — `installPMAgents()` → `ensurePMAgents(workspaceId)` |
| Client | `frontend/src/api/agent.ts` — `ensurePMAgents` → `POST …/agents/ensure-pm` |
| Backend | `agent.EnsurePM` → `AgentService.EnsurePMAgents` |
| Defs (idempotent by name) | `backend/internal/ai/service/pm_agents.go` — `PMAgentDefs`: 请求分诊 / 交付风险 / Sprint 总结 / Spec 草稿 |

**Hand-test risk**: Button lives under project Automation settings; needs workspace membership and working API.

---

### P2-2 — 创建→分诊模板 → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| Template | `ProjectSettings.vue` — `automationTemplates` entry `intakeTriage`: `trigger: 'issue.created'`, action `dispatch_agent` |
| Resolve agent | `applyTemplate` — `ensurePMAgents` then find `TRIAGE_AGENT_NAME = '请求分诊'`, set `action.value = triage.id` |
| i18n label | `automationTemplates.intakeTriage` = 「新建即分诊」 |

**Hand-test risk**: Apply fails if ensure-pm does not return an agent named exactly `请求分诊`.

---

### P2-3 — 普通创建触发 → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| Create path | `IssueService.Create` → `NotifyIssueCreated` → `runAutomations(..., "issue.created", …)` |
| Action | `automation_service.go` — `RegisterAction("dispatch_agent", …)` |

**Hand-test risk**: Rule must be enabled; triage dispatch needs LLM Key — without Key expect empty/failed activity (checklist allows “需 LLM”).

---

### P2-4 — Intake 触发 → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| Public submit | `backend/internal/handler/intake_handler.go` — `Submit` creates issue with `IntakeStatus: pending`, then `h.issueSvc.NotifyIssueCreated(issue)` |
| Same automation bus | `IssueService.NotifyIssueCreated` — `issue.created` |

**Hand-test risk**: Intake handler must be constructed with non-nil `issueSvc`; same LLM dependency as P2-3 for visible triage output.

---

### P2-5 — Cycle 一键总结 → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| UI | `CycleDetail.vue` — button `t('cycle.aiSummary')` → `runCycleSummary` → `sprintPlan(project_id, cycleId)` |
| Result panel | same — summary / risks / `saveSummaryAsPage` → `t('cycle.saveAsPage')` |
| API | `frontend/src/api/ai.ts` — sprint plan helper; backend `AIHandler.SprintPlan` |

**Hand-test risk**: Needs AI config; save-as-page needs page create permissions.

---

### P2-6 — Issue AI Tab → **WIRED**

| Layer | File + symbol |
|-------|----------------|
| Tab actions | `IssueDetail.vue` — `aiActions` + `executeAIAction` |
| Summarize / risk | `analyzeWithAI(projectId, issueId)` → in-page `[data-test="ai-analyze-result"]` |
| Suggest labels | `suggestLabels(...)` → `[data-test="ai-label-suggestions"]` |
| Copilot still available | `showAICopilot = true` via header button / `action === 'copilot'` |

**Hand-test risk**: `suggest` maps to label suggestions (not free-form “steps”); LLM required for non-empty panels.

---

## Summary table

| Item | Verdict |
|------|---------|
| P1-1 | WIRED |
| P1-2 | PARTIAL |
| P1-3 | WIRED |
| P1-4 | WIRED |
| P1-5 | WIRED |
| P1-6 | WIRED |
| P1-7 | PARTIAL |
| P1-8 | WIRED |
| P2-1 | WIRED |
| P2-2 | WIRED |
| P2-3 | WIRED |
| P2-4 | WIRED |
| P2-5 | WIRED |
| P2-6 | WIRED |

**Counts**: 12 WIRED · 2 PARTIAL · 0 MISSING

**Static blockers for hand-test prep** (not live results):
1. P1-2 — Issue page lacks Ctrl+J / FAB (open via AI tab only).
2. P1-7 — Triage AI analyze swallows errors.
3. P1-4 — possible unassign URL mismatch (`unassign-agent` vs `assign-agent`).
4. Triggers confirmed dotted: `issue.created`.
