# AI 项目管理 Phase 2 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 Phase 1 日常面升级为「触发 + 开箱 Agent」：创建/Intake 可自动分诊、四个项目管理 Agent 可一键安装、Cycle 一键总结、Issue AI Tab 真正调用 analyze / suggest-labels。

**Architecture:** 复用现有 `agents` CRUD、`automation` 事件总线（`issue.created` + `dispatch_agent`）、`ai/analyze|suggest-labels|sprint-plan`、`TriagePanel`。Phase 2 以种子配置、模板接线与前端入口为主；不新建 Agent 控制台，不引入 Agent 平台/MCP 叙事。

**Tech Stack:** Vue 3 + TypeScript、Go/Gin、现有 Automation EventBus、AgentService、AIService。

**Spec:** `docs/superpowers/specs/2026-09-25-ai-project-management-redesign.md`（§6 开箱 Agent、§9 Phase 2）

## Global Constraints

- 面向日常项目管理的 AI 项目管理；禁止 独立 Agent 平台 / MCP 分发 / Agent 市场文案进入默认 UI。
- 开箱 Agent **仅项目管理向**（分诊 / 风险 / Sprint 总结 / Spec 草稿）；不做代码开发 Agent。
- 自动化：`dispatch_agent` 触发后仍须可审计（走现有 Agent activity）；**Intake 提交也要能进事件总线**（今日缺口）。
- Build/写操作：自动化分诊以「建议评论 / 标签建议 / 指派 Agent」为主；**禁止**自动化静默改业务字段除非用户已启用对应动作模板且文案标明。
- 触发类型必须与后端一致：`issue.created`（点分），**禁止**再写入 `issue_created`（现有模板 bug，本 Phase 一并修）。
- 中英文 i18n 同步；每 Task 有可运行验证；小步 commit（用户未要求则不 push）。

---

## 文件结构总览

| 区域 | 文件 | 动作 |
|------|------|------|
| 开箱 Agent 种子 | `backend/internal/ai/service/pm_agents.go`（新）、`agent_service.go`、workspace 创建钩子或 seed | Ensure 四人组 |
| Intake → 自动化 | `backend/internal/handler/intake_handler.go`、可选 `issue_service.runAutomations` | Submit 后发 `issue.created` |
| 自动化模板 | `frontend/src/views/ProjectSettings.vue`、locales | 分诊模板 + 修正 trigger |
| Cycle 总结 | `frontend/src/views/CycleDetail.vue`、`frontend/src/api/ai.ts` | 一键总结 + 可选存 Page |
| Issue AI Tab | `frontend/src/views/IssueDetail.vue` | analyze / suggest-labels 真调用 |
| i18n / 管线 | locales、`docs/dev/pipeline-status.md`、`active/README.md` | Phase 2 进度 |

---

## Task 1: 开箱四人组 Agent — Ensure API / 种子

**Files:**
- Create: `backend/internal/ai/service/pm_agents.go`
- Create: `backend/internal/ai/service/pm_agents_test.go`
- Modify: `backend/internal/ai/handler/agent.go`（可选 `POST …/agents/ensure-pm`）
- Modify: `backend/internal/router/router.go`
- Modify: workspace 创建成功路径（搜 `CreateWorkspace` / `WorkspaceService.Create`）或 `seed.go`：工作区创建后调用 Ensure

**Interfaces:**
- Produces: `EnsurePMAgents(workspaceID, userID uint64) ([]*model.Agent, error)` — idempotent by stable `name`（或 slug tag）
- Produces agent names (exact, bilingual display via i18n on FE; DB name use Chinese primary matching product locale default, or English keys — **use these exact `name` values**):
  1. `请求分诊`
  2. `交付风险`
  3. `Sprint 总结`
  4. `Spec 草稿`

- [x] **Step 1: 写失败测试（幂等 Ensure）**

```go
// pm_agents_test.go
func TestPMAgentDefs_HasFour(t *testing.T) {
	assert.Len(t, PMAgentDefs(), 4)
	names := map[string]bool{}
	for _, d := range PMAgentDefs() {
		names[d.Name] = true
		assert.NotEmpty(t, d.SystemPrompt)
	}
	assert.True(t, names["请求分诊"])
	assert.True(t, names["交付风险"])
	assert.True(t, names["Sprint 总结"])
	assert.True(t, names["Spec 草稿"])
}
```

- [x] **Step 2: 实现 `PMAgentDefs` + `EnsurePMAgents`**

```go
// pm_agents.go — 核心结构（提示词可加长，但必须覆盖 Spec §6 产出）
type PMAgentDef struct {
	Name         string
	Avatar       string
	Capabilities []string
	SystemPrompt string
}

func PMAgentDefs() []PMAgentDef {
	return []PMAgentDef{
		{
			Name: "请求分诊", Avatar: "🏥",
			Capabilities: []string{"triage", "search_issues", "suggest_labels"},
			SystemPrompt: `你是请求分诊 Agent。对新建/Intake 工作项：建议类型与优先级、可能重复项、标签与路由。只输出结构化建议；不擅自改库。`,
		},
		{
			Name: "交付风险", Avatar: "⚠️",
			Capabilities: []string{"search_issues", "get_cycle_progress", "add_comment"},
			SystemPrompt: `你是交付风险 Agent。分析 Cycle/Issue 的阻塞、逾期与依赖风险，用评论给出可执行建议。`,
		},
		{
			Name: "Sprint 总结", Avatar: "📊",
			Capabilities: []string{"get_cycle_progress", "search_issues", "list_cycles"},
			SystemPrompt: `你是 Sprint 总结 Agent。根据 Cycle 进度与完成情况写中文总结，可建议存为 Page。`,
		},
		{
			Name: "Spec 草稿", Avatar: "📝",
			Capabilities: []string{"get_issue", "add_comment"},
			SystemPrompt: `你是 Spec 草稿 Agent。根据 Issue 写需求提纲（背景/目标/范围/验收），输出评论或 Page 草稿。`,
		},
	}
}

func (s *AgentService) EnsurePMAgents(workspaceID, userID uint64) ([]*model.Agent, error) {
	var out []*model.Agent
	for _, def := range PMAgentDefs() {
		var existing model.Agent
		err := s.db.Where("workspace_id = ? AND name = ? AND deleted_at IS NULL", workspaceID, def.Name).First(&existing).Error
		if err == nil {
			out = append(out, &existing)
			continue
		}
		prompt := def.SystemPrompt
		ag, err := s.Create(workspaceID, userID, &AgentCreateRequest{
			Name: def.Name, Avatar: def.Avatar, AgentType: "builtin",
			Capabilities: def.Capabilities, Status: "active",
			SystemPrompt: &prompt,
			PermissionMode: strPtr("public"), // 或项目可见；与现有枚举对齐
		})
		if err != nil {
			return nil, err
		}
		out = append(out, ag)
	}
	return out, nil
}
```

权限枚举以 `model.AgentPermissionMode` 为准；若无 `public`，用工作区成员可调用的现有模式，并在实现时查 `model/agent.go`。

- [x] **Step 3: 暴露 `POST /workspaces/:wsParam/agents/ensure-pm`（需登录，workspace 成员）**

```go
// handler: 调 EnsurePMAgents；返回 { agents: [...] }
```

- [x] **Step 4: 工作区创建后调用 Ensure（搜 `WorkspaceService.Create` 成功分支）**

对**已有**工作区：前端「安装开箱 Agent」按钮可调 ensure-pm（Task 2 UI）；后端创建钩子覆盖新工作区。

- [x] **Step 5: 跑测**

```bash
cd backend && go test ./internal/ai/service/ -run "PMAgent" -count=1 -v
```

Expected: PASS

- [x] **Step 6: Commit**

```bash
git add backend/internal/ai/service/pm_agents.go backend/internal/ai/service/pm_agents_test.go backend/internal/ai/handler/agent.go backend/internal/router/router.go
# + workspace create hook file(s)
git commit -m "feat(ai): ensure four out-of-box PM agents per workspace"
```

---

## Task 2: 前端 — 安装开箱 Agent + 自动化「创建→分诊」模板

**Files:**
- Modify: `frontend/src/api/agent.ts` — `ensurePMAgents(workspaceId)`
- Modify: `frontend/src/views/ProjectSettings.vue` — automations 区模板 + 可选 AI/Agent 安装入口
- Modify: `frontend/src/locales/zh-CN.json`、`en-US.json`
- Fix: 现有 `automationTemplates` 的 `trigger: 'issue_created'` → `'issue.created'`（及 `comment_added` → `comment.added`，`state_changed` → `issue.state_changed`）

**Interfaces:**
- Consumes: `POST /workspaces/:id/agents/ensure-pm`
- Consumes: 分诊 Agent id（Ensure 返回中 name=`请求分诊`）
- Produces: 一键创建自动化规则：`trigger issue.created` + `dispatch_agent`

- [x] **Step 1: API**

```ts
// agent.ts
export async function ensurePMAgents(workspaceId: number): Promise<Agent[]> {
  const res = await api.post(`/workspaces/${workspaceId}/agents/ensure-pm`)
  return res.data?.agents ?? res.data ?? []
}
```

- [x] **Step 2: 修正全部模板 trigger 为点分形式**

对照后端 `registerEventHandlers` 列表：`issue.created` | `issue.updated` | `issue.state_changed` | `issue.assigned` | `comment.added` | `scheduled`。

- [x] **Step 3: 增加模板 `intakeTriage`**

```ts
{
  name: 'intakeTriage',
  icon: '🏥',
  bgClass: 'bg-teal-100',
  trigger: 'issue.created',
  conditions: [], // 可选：intake_status equals pending（若条件引擎支持）
  actions: [{
    type: 'dispatch_agent',
    value: triageAgentId, // apply 时解析：先 ensurePMAgents，找到「请求分诊」
    field: '请对此工作项做分诊：建议类型/优先级/标签，并指出可能重复项。只输出建议。'
  }],
  requiresPMAgents: true,
}
```

`applyTemplate` 若 `requiresPMAgents`：先 `ensurePMAgents(workspaceId)`，再填 `value`。

- [x] **Step 4: i18n**

```json
// zh-CN automationTemplates
"intakeTriage": "新建即分诊",
"intakeTriageDesc": "Issue/Intake 创建后派给「请求分诊」Agent",
"installPMAgents": "安装开箱 Agent",
"installPMAgentsDone": "已安装/已存在四个项目管理 Agent"
```

```json
// en-US
"intakeTriage": "Triage on create",
"intakeTriageDesc": "Dispatch Request Triage agent when an issue is created",
"installPMAgents": "Install starter agents",
"installPMAgentsDone": "Four PM agents installed or already present"
```

- [ ] **Step 5: 手测**

1. 项目设置 → 自动化 → 安装开箱 Agent → 列表出现四人组。  
2. 点「新建即分诊」→ 规则创建且 `trigger_type` JSON 内为 `issue.created`。  
3. 普通创建 Issue → Agent activity / 评论出现分诊建议（需 LLM 配置）。

- [x] **Step 6: Commit**

```bash
git commit -m "feat(ai): add PM agent install and create-to-triage automation template"
```

---

## Task 3: Intake Submit 进入自动化事件总线

**Files:**
- Modify: `backend/internal/handler/intake_handler.go`
- Modify: 注入 `IssueService` 或 EventBus（看现有 DI；优先复用 `IssueService.runAutomations` / 公开的 `PublishIssueCreated`）
- Test: `backend/internal/handler/intake_handler_test.go` 或 service 级（若无 DB，用接口 mock / 抽 `AutomationPublisher`）

**背景：** `IntakeHandler.Submit` 今日仅 `db.Create(issue)`，**不**发 `issue.created`，导致 Task 2 模板对 Intake 无效。

- [x] **Step 1: 定位 Issue 创建后如何发事件**

在 `issue_service.go` 搜 `runAutomations` / `issue.created`；抽出可复用函数，例如：

```go
func (s *IssueService) NotifyIssueCreated(issue *model.Issue) {
    s.runAutomations(issue.ProjectID, "issue.created", map[string]interface{}{
        "issue_id": issue.ID, "project_id": issue.ProjectID,
        "workspace_id": issue.WorkspaceID, "name": issue.Name,
        // 若有 intake_status 一并传入供条件匹配
    })
}
```

（参数以现有 `runAutomations` 签名为准，勿发明不一致字段名。）

- [x] **Step 2: Intake Submit 在 Create 成功后调用 Notify**

```go
if err := h.db.Create(issue).Error; err != nil { ... }
if h.issueSvc != nil {
    h.issueSvc.NotifyIssueCreated(issue)
}
c.JSON(201, issue)
```

构造函数 `NewIntakeHandler` 增加依赖；`router.go` 接线。

- [x] **Step 3: 单测或集成断言**

至少：Create 后 publisher 被调用 1 次（mock）；或文档化手测步骤若 DI 难测。

- [ ] **Step 4: 手测**

启用「新建即分诊」→ 公网 Intake 表单提交 → pending Issue + Agent 活动。

- [x] **Step 5: Commit**

```bash
git commit -m "feat(ai): fire issue.created automations from intake submit"
```

---

## Task 4: Cycle「一键总结」

**Files:**
- Modify: `frontend/src/views/CycleDetail.vue` — 标题栏按钮
- Modify: `frontend/src/api/ai.ts` — `sprintPlan` 可选 `cycle_id` query（若后端已忽略未知字段则先传）
- Modify: `backend/internal/ai/handler/ai.go` + `ai_service.go` `SprintPlan` — 可选按 `cycle_id` 收窄上下文（增量；无则先用项目级 sprint-plan + 前端把 cycle 名塞进展示）
- Modify: locales

**Interfaces:**
- Consumes: `POST /projects/:projectId/ai/sprint-plan`
- Produces: 页内结果面板 +「存为 Page」按钮（复用 `createPage` 或跳转 `pages?content=`，对齐 `Project.vue` `handleAISaveAsPage`）

- [x] **Step 1: UI 按钮**

在 `CycleDetail.vue` Start/End 旁：

```vue
<button
  type="button"
  class="..."
  :disabled="summarizing"
  @click="runCycleSummary"
>{{ summarizing ? t('common.loading') : t('cycle.aiSummary') }}</button>
```

- [x] **Step 2: 调用与展示**

```ts
async function runCycleSummary() {
  summarizing.value = true
  summaryError.value = ''
  try {
    summary.value = await sprintPlan(projectId) // 后续带 cycleId
  } catch (e: any) {
    summaryError.value = e?.response?.data?.message || e.message
  } finally {
    summarizing.value = false
  }
}
```

渲染 `summary` 文本/结构化字段（按 API 实际响应：读 `AISprintPlanResponse`）。

- [x] **Step 3（可选加固）: SprintPlan 接受 cycle_id**

```go
// query cycle_id > 0 时：只汇总该 cycle 的 issues / 进度，写入 prompt
```

- [x] **Step 4: 「存为 Page」**

```ts
async function saveSummaryAsPage() {
  await createPage(projectId, workspaceId, {
    title: `${cycle.value?.name || 'Cycle'} — ${t('cycle.aiSummary')}`,
    content: summaryMarkdown.value,
  })
}
```

- [ ] **Step 5: 手测** — 打开 Cycle → 一键总结有返回 → 可存 Page。

- [x] **Step 6: Commit**

```bash
git commit -m "feat(ai): add one-click cycle summary on cycle detail"
```

---

## Task 5: Issue AI Tab — 真实 analyze / suggest-labels

**Files:**
- Modify: `frontend/src/views/IssueDetail.vue`（`executeAIAction`、AI Tab 模板）
- Use: `analyzeWithAI`、`suggestLabels` from `@/api/ai`
- Modify: locales（加载中 / 空标签 / 错误）

**现状：** `executeAIAction` 对所有 key 仅 `showAICopilot = true`。

- [x] **Step 1: 状态**

```ts
const aiTabLoading = ref(false)
const aiTabError = ref('')
const aiAnalyzeResult = ref<any>(null)
const aiLabelSuggestions = ref<any[]>([])
```

- [x] **Step 2: 改 `executeAIAction`**

```ts
async function executeAIAction(action: string) {
  if (action === 'copilot') {
    showAICopilot.value = true
    return
  }
  if (!projectId.value || !issueId) return
  aiTabLoading.value = true
  aiTabError.value = ''
  try {
    if (action === 'summarize' || action === 'risk') {
      aiAnalyzeResult.value = await analyzeWithAI(projectId.value, issueId)
      aiLabelSuggestions.value = []
    } else if (action === 'suggest') {
      const res = await suggestLabels(projectId.value, issueId)
      aiLabelSuggestions.value = res?.labels ?? res ?? []
    }
  } catch (e: any) {
    aiTabError.value = e?.response?.data?.message || e.message || t('ai.connectionFailed')
  } finally {
    aiTabLoading.value = false
  }
}
```

- [x] **Step 3: AI Tab 面板**

在 AI Tab 按钮列表下方渲染：

- loading 指示
- `aiAnalyzeResult` 关键字段（对照 `AIAnalyzeResponse`：bottlenecks / summary 等，按实际 JSON）
- 标签建议列表 + 「应用到 Issue」可选（调用现有 `addIssueLabel`；无则仅展示）
- 保留「打开 Copilot」

- [x] **Step 4: 前端单测（可选）**

若有 IssueDetail 测试基建则加；否则手测：AI Tab → 总结有数据；建议标签有列表；Copilot 仍可开。

- [x] **Step 5: Commit**

```bash
git commit -m "feat(ai): wire issue AI tab to analyze and suggest-labels"
```

---

## Task 6: 文档与验收收尾

**Files:**
- Modify: `docs/dev/pipeline-status.md` — Phase 2 行
- Modify: `docs/dev/active/README.md` — 指向本 plan
- Modify: Spec §12 下一步改为 Phase 2 计划已就绪 / 实施中

- [x] **Step 1: 更新管线**

| 功能 | Implement |
|------|-----------|
| AI PM Phase 2（触发+开箱） | 🔄 Tasks 1–5 已接线；手测验收待完成 |

- [x] **Step 2: 自检本 plan checkbox；缺项记 Phase 3 或 bug list**

- [x] **Step 3: Commit**

```bash
git commit -m "docs: mark AI PM Phase 2 implementation progress"
```

---

## Phase 2 实施进度备注（2026-09-25）

- Tasks 1–5 代码已合入 `feature/ai-pm-phase2`（commits `14f3f28`…`f0d5343`）。计划内**手测**步骤未勾选。
- Deferred（非阻塞；可 Phase 3 / chore）:
  - Ensure：无 DB 级幂等集成测；工作区创建种子失败吞错误；无 `(workspace_id, name)` unique
  - `AutomationManager` / 部分 e2e 仍用 underscore trigger（本 plan 仅强制 ProjectSettings）
  - Analyze API 仍为项目级健康；标签建议 Apply 需 `label_id` 或名称匹配
  - 下方验收清单待人工跑通后再将管线 Implement 标为完成

## 验收清单（对照 Spec §9 Phase 2）

| # | 标准 | Task |
|---|------|------|
| 1 | 可安装四个开箱 PM Agent（幂等） | 1–2 |
| 2 | 「创建→分诊」自动化模板可用，trigger=`issue.created` | 2 |
| 3 | Intake 提交也会触发自动化 | 3 |
| 4 | Cycle 详情有一键总结并可存 Page | 4 |
| 5 | Issue AI Tab 调用 analyze / suggest-labels，不只开空 Copilot | 5 |
| 6 | 无 Agent 平台/MCP 主叙事回归 | 全局 |

---

## 不在本 Plan

- Phase 3 Harness / Loop / 对抗评审 / 多 Agent 流水线  
- Slack / Marketplace / GitHub App OAuth 深度  
- `registry.SeedDefaults`（`sprint-analyzer` 等）与产品 `agents` 表合并（保持分离，避免混淆）  
- 修复全部历史自动化模板 trigger（本 plan **只强制**改 ProjectSettings 内模板；库内旧规则可另开 chore）

---

## 执行建议

使用 `superpowers:subagent-driven-development` 按 Task 1→6 顺序执行；每 Task 验证通过后再开下一 Task。
