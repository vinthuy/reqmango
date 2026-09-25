# AI 项目管理重新设计（对标 Plane AI）

> **Status (2026-09-26):** CANONICAL product direction — **implemented & accepted**.  
> Agent-platform / Harness / Loop / Multica docs moved to [docs/superseded/agent-platform/](../../superseded/agent-platform/README.md).  
> Core PM customization gate: [product-core-qa](../../dev/acceptance/2026-09-25-product-core-qa.md) §H PASS.


> **版本**: v1.0  
> **日期**: 2026-09-25  
> **状态**: Approved for planning（产品方向已确认；实施计划另文）  
> **对标**: [Plane AI](https://plane.so/ai) / [Plane Agents](https://plane.so/agents)  
> **替代叙事**: 不再以 Multica / MCP 分发 / Agent 控制台作为主产品故事

---

## 1. 背景与决策

### 1.1 为什么重做

Reqmango 已具备大量 AI/Agent 后端能力（Copilot、Ask/Build、创建预览、Intake 分诊 API、Issue↔Agent、Page AI、Harness/Loop 等），但产品叙事曾偏向：

- Multica 式「Agent 平台 / SDLC 自动化」
- 侧栏「AI Agents」超大控制台（Loop、Pipeline、Squad、Developer Agent…）
- MCP/CLI/SDK 分发作为增长主路径

这些与用户日常 **项目管理** 路径脱节。Plane AI 的卖点是：**在工作项生命周期里用 AI**（问、建、分诊、指派、总结），而不是另开一套 Agent IDE。

### 1.2 产品命题

> AI 是项目管理的一层能力，不是旁边的机器人站。  
> 问、建、分诊、指派、总结 —— 都发生在当前项目/Cycle/Issue 上下文中，可预览、可审计、可撤销边界清晰。

### 1.3 明确丢弃（产品主路径）

| 丢弃 / 降级 | 说明 |
|-------------|------|
| Multica 对标叙事 | `docs/superseded/agent-platform/AI_AGENT_PRD.md` 中 Multica「Agent as Teammates / 全 SDLC」主叙事 **废止**；该文归档为历史，不再指导排期 |
| MCP/CLI 一键分发当 Phase 0 | SDK/MCP 代码可保留，**不作为本设计的交付范围** |
| 「AI Agents」控制台当主入口 | Dashboard / Loop / Pipeline / Squad / Autopilot / Developer·Tester·CICD·SDLC Agent 等 **退出默认导航** |
| 以 Harness/Loop 为第一差异化 | 可作后续「超 Plane」能力；**本阶段不对用户讲编排平台** |

### 1.4 对标范围（Plane AI 主路径）

| Plane 能力 | 本设计是否纳入 Phase 1–2 |
|------------|-------------------------|
| Ask / Build（查 vs 改） | ✅ |
| Plan → Approve → Execute | ✅（Build / Create） |
| Context sidecar（当前视图上下文） | ✅ |
| Agent 指派 = 队友 | ✅ |
| @mention Agent | ✅ |
| 同一套审计 / 活动流 | ✅ |
| 创建前去重 | ✅ |
| Intake 自动分诊 | ✅ |
| AI → Page / 总结 Cycle | ✅ |
| Slack 内 AI | ❌ 本轮不做 |
| Agent Marketplace / ADK | ❌ 本轮不做 |
| OAuth GitHub App 深度 | ❌ 本轮不做（非 AI 主线） |

---

## 2. 目标用户与成功标准

### 2.1 用户

项目管理中的 **PM / 研发负责人 / 研發同学**：在 Issue、Sprint、Intake、Wiki 上完成工作，顺带使用 AI，而不是「先学 Agent 平台」。

### 2.2 Phase 1 成功标准（可验收）

1. 用户打开任意 Issue，侧栏可 **指派 Agent**（持久），状态可见；不必去 `/agent-issues`。  
2. 评论输入 `@` 可选 Agent；@ 后 Agent 回复出现在评论/活动流。  
3. Ctrl+J Copilot **自动带上**当前 project / issue / cycle（若在对应页）。  
4. Copilot 仅 **Ask | Build** 两主模式；Build/Create 必须先预览再确认。  
5. 手动创建 Issue 与 AI Create 在提交前展示 **疑似重复**。  
6. 项目内可打开 **Intake 分诊**（挂载现有 `TriagePanel`），AI 分析可用。  
7. Issue 活动流可见 Agent 操作（挂载/内嵌审计，不再只有孤儿组件）。  
8. 默认侧栏 **不再** 突出「AI Agents」控制台；高级入口可选保留。

---

## 3. 能力模型（五件事）

```
┌─────────────────────────────────────────────────────────┐
│                 AI Project Management                    │
├──────────┬──────────┬──────────┬──────────┬─────────────┤
│   Ask    │  Build   │  Triage  │  Assign  │   Capture   │
│ 查与洞察  │ 改与创建  │ 入口治理  │ Agent队友 │ 沉淀与总结  │
└──────────┴──────────┴──────────┴──────────┴─────────────┘
```

| 能力 | 用户一句话 | 主要交互 |
|------|------------|----------|
| **Ask** | 「这个 Sprint 谁阻塞了？」 | Sidecar / Ctrl+J，只读工具 |
| **Build** | 「按这段话建 5 个 Issue」 | 预览卡片 → 确认 → 执行 |
| **Triage** | 「新来的需求像不像已有的？」 | Intake 面板 + 创建前去重 |
| **Assign** | 「把这个交给分诊 Agent」 | 指派 / @ / 规则触发 |
| **Capture** | 「总结本 Cycle，存成 Page」 | 一键总结 / Save as Page |

---

## 4. 信息架构

### 4.1 默认导航

| 入口 | 行为 |
|------|------|
| Ctrl+J / FAB / 命令面板「AI」 | 打开 **项目 AI Sidecar**（Ask/Build） |
| Issue 详情 | 属性区 **Agent 指派**；活动流含 Agent；评论可 @Agent |
| 项目设置 / Issues 相关 | **Intake 分诊** 入口 |
| Pages 工具栏 | 保留 summarize / improve / generate |
| 工作区设置 | AI 模型与密钥（BYO） |

### 4.2 降级导航

| 原入口 | 处理 |
|--------|------|
| 侧栏「AI Agents」 | **从默认侧栏移除**；设置或「高级」中保留链接「Agent 高级（实验）」指向原 `/agents` 路由 |
| `/agents/*` 全套页面 | 路由保留，不删代码；文档标注 Experimental |
| Harness / Loop API | 保留后端；产品文案与主 UI 不引导 |

### 4.3 Copilot 模式收敛

| 现 Tab | 新模型 |
|--------|--------|
| Ask | **Ask** |
| Build | **Build**（含原 Build 的 plan-confirm） |
| Create | 并入 **Build** 的「创建工作项」技能（仍走预览） |
| Chart | 并入 **Ask** 的「可视化」或 Build 的「保存到仪表盘」结果动作 |
| Agent | 移除独立 Tab；改为 Issue 指派 + 「用当前 Agent 执行」快捷，或 Build 内选 Agent |

---

## 5. 关键用户流程

### 5.1 Context Sidecar（对齐 Plane）

打开 Sidecar 时组装上下文（前端传入 + 后端校验）：

```json
{
  "workspace_id": 1,
  "project_id": 12,
  "issue_id": 990,
  "cycle_id": 3,
  "page_id": null,
  "view": "issue_detail"
}
```

规则：

- 在 Issue 页：强制 `issue_id`，Ask/Build 默认围绕该 Issue。  
- 在 Cycle 页：带 `cycle_id`。  
- 在项目看板：带 `project_id`。  
- 上下文条常显：「当前：DEMO-12 · Sprint 24」（可一键清除到仅项目级）。

### 5.2 Build：Plan → Approve → Execute

1. 用户自然语言描述意图。  
2. 模型只输出 **结构化计划**（创建/更新字段列表），SSE 或 JSON。  
3. UI 展示可编辑预览（沿用/强化 Create 预览与 Result Actions）。  
4. 用户点确认后才调用写接口。  
5. 失败项逐条回报；成功写入 Issue 活动「AI 执行了计划 #…」。

Ask 模式禁止静默写操作；若模型想写，UI 提示「切换到 Build」。

### 5.3 Assign Agent（对齐 Plane Assign to ship）

- Issue 属性：`assignees`（人）与 **`agent_assignee`**（Agent）并列展示。  
- 指派后：调用已有 `assign-agent` API；侧栏显示状态（idle / running / done / needs_input）。  
- **Dispatch** 降级为「立即跑一次」次要按钮，主路径是指派。  
- `AgentIssuesPage` 保留为批量管理，不再是唯一入口。

### 5.4 @mention

- 评论与 Issue Chat 的 `@` 候选：**成员 + 本项目可用 Agent**。  
- 命中后走现有 `HandleMention` / chat mention；回复进同一线程。  
- 活动流记一条 Agent 响应摘要。

### 5.5 Triage 与创建前去重

- **Intake**：路由挂载 `TriagePanel`（项目设置或 Issues → 分诊队列）。  
- **创建 Issue**（表单 + AI Create）：标题/描述变更防抖后调用轻量查重 API（可复用 intake `ai-analyze` 的 duplicate 逻辑或 RQL/ILIKE + 可选 embedding 后续再加）。  
- UI：创建按钮旁展示「3 条相似」，可点开对比，仍允许强制创建。

### 5.6 Capture

- Copilot 结果动作保留：Save as Page、批量子工作项、图表入库。  
- Issue / Cycle 提供「一键总结」→ 调用已有 analyze / sprint-plan 或 chat Ask 模板，可存 Page。

### 5.7 审计

- 将 `AgentAuditLog`（或精简版）嵌进 Issue「活动」时间线过滤「仅 Agent」。  
- 与人类活动同一组件，不另开「神秘 Agent 后台」。

---

## 6. 开箱 Agent（Phase 2，对标 Plane 现成 Agent）

仅保留 **项目管理** 向，不做代码开发 Agent：

| Agent | 触发 | 产出 |
|-------|------|------|
| 请求分诊 | Intake 新建 / 手动 | 类型、优先级、标签建议、重复项、路由建议 |
| 交付风险 | Cycle 内定时或手动 | 阻塞、逾期、依赖风险评论 |
| Sprint 总结 | 手动 / Cycle 结束 | 总结文本 → 可选 Page |
| Spec 草稿 | @mention / 手动 | 需求提纲评论或 Page 草稿 |

实现优先：**配置 + 提示词 + 工具权限** 绑定现有 `agent` 模型与 automation `dispatch_agent`；不新开平台控制台。

---

## 7. 与现有代码的映射

### 7.1 保留并产品化（主路径）

| 能力 | 现有锚定 |
|------|----------|
| Sidecar | `AICopilot.vue`, `useAI.ts`, `POST …/ai/chat` |
| Create 预览 | Create 模式 + `POST …/ai/create` + `AIResultActions.vue` |
| Page AI | `ProjectPages.vue` + `POST …/pages/:id/ai` |
| 指派 API | `issue_agent_*` |
| @mention | `comment_service`, chat mention |
| Intake 分析 | `intake` + `TriagePanel.vue`（需挂载） |
| 审计数据 | Agent activity API + `AgentAuditLog.vue`（需挂载） |
| BYO Keys | `AISettingsPanel.vue` |

### 7.2 降级（保留代码，弱化产品）

`views/agents/*` 大部分、`ai/harness/*`、`ai/loop/*`、squad、autopilot、developer/tester/cicd/sdlc agent 视图与商店叙事。

### 7.3 废止文档指引

| 文档 | 处理 |
|------|------|
| `docs/superseded/agent-platform/AI_AGENT_PRD.md` | 文首标注 **Superseded**；Multica 方案不作排期依据 |
| `docs/superseded/agent-platform/2026-07-18-reqmango-agent-platform-design.md` | 标注为 **基础设施/远期**；产品主路径以本文为准 |
| 本文 | **现行 AI 产品设计** |

---

## 8. API / 数据（增量）

以接线为主，少造轮子：

| 增量 | 说明 |
|------|------|
| Chat/Create 请求强化 `context` | 必填 `project_id`；可选 `issue_id`/`cycle_id`/`page_id` |
| `GET/POST …/issues/duplicate-check` | 或扩展 create preview：返回 `duplicates: [{id, name, score}]` |
| Issue 详情响应 | 确保含 `agent_assignee` / 状态字段供侧栏绑定 |
| 活动流 | 聚合或过滤 Agent activity（可前端合并两个 API） |

无需为本阶段新建「Agent 平台」表结构。

---

## 9. 分阶段交付

### Phase 1 — Plane 日常面（优先，约 1–1.5 周）

1. 导航：隐藏默认「AI Agents」；高级入口可选。  
2. Copilot：Ask/Build 收敛 + 上下文条。  
3. Issue：指派 Agent + 状态；Dispatch 次要化。  
4. 评论/Chat：@ 含 Agent。  
5. 活动流：挂审计。  
6. 挂载 Intake Triage。  
7. 创建前去重。  
8. Build/Create 强制预览确认（加固现有行为与文案）。

### Phase 2 — 触发与开箱 Agent（约 1 周）

1. 自动化/事件：创建 → 分诊 Agent（产品化模板）。  
2. 四个开箱 Agent 配置与文案。  
3. Cycle「一键总结」入口。  
4. Issue AI Tab 真正调用 analyze / suggest-labels（不再只会打开空 Copilot）。

### 已取消作产品方向 — 原「Phase 3 超 Plane」

Harness、Loop、对抗评审、多 Agent 流水线：**不再作为下一阶段目标**（代码可保留降权，不排期、不讲产品故事）。

### 下一增量 — AI 自动化 + Capture（已定 B1 + C1）

| 代号 | 内容 | 计划 |
|------|------|------|
| **验收** | Phase 1–2 手测清单 | `docs/dev/acceptance/2026-09-25-ai-pm-phase1-2.md` |
| **B1** | 更多 AI 自动化模板（分诊 / 风险 / Spec 等）+ trigger 卫生 | `docs/superpowers/plans/2026-09-25-ai-automation-and-project-summary.md` |
| **C1** | 项目级 AI 总结 → 存 Page（Cycle 总结已有） | 同上 |

不做：B2 自然语言配规则、C2 Dashboard 定时 AI 报表（可另开）。

---

## 10. 非目标与风险

### 非目标

- 替换 Plane 的全部集成（Slack bot、Marketplace）。  
- 删除 Harness/Loop 代码（仅产品降权）。  
- **以 Harness/Loop/对抗评审/多 Agent 流水线作为下一产品阶段**（已取消）。  
- 以 IDE/MCP 安装转化率作为本阶段 KPI。

### 风险

| 风险 | 缓解 |
|------|------|
| 高级用户找不到原控制台 | 「高级 Agent（实验）」深链 + 发布说明 |
| 去重误报挡创建 | 只警告不拦截；可「仍要创建」 |
| Build 乱写生产数据 | 无确认不落库；Ask 禁写 |

---

## 11. 成功指标（产品）

- Issue 详情完成「指派 Agent」路径无需离开详情页。  
- 新建 Issue 流程中去重提示曝光率 > 0（有相似时展示）。  
- Intake 分诊页可从项目导航到达。  
- 默认新用户侧栏点击路径中，不再出现 Loop/Pipeline 作为第一屏。

---

## 12. 下一步

1. **执行** Phase 1–2 验收：`docs/dev/acceptance/2026-09-25-ai-pm-phase1-2.md`。  
2. **实施** B1+C1：`docs/superpowers/plans/2026-09-25-ai-automation-and-project-summary.md`。  
3. 原 Phase 3（Harness/Loop）**已取消**作产品方向。  
4. 管线：`docs/dev/pipeline-status.md` / `docs/dev/active/README.md`。

---

## 附录 A. Plane 对照速查

| Plane 文案 | Reqmango Phase 1–2 |
|------------|---------------------|
| Talk to build | Build + 预览 |
| Ask to know | Ask + 上下文 |
| Assign to ship | agent_assignee + @ |
| Duplicates before create | 创建前查重 |
| Incoming triage | Intake + TriagePanel |
| Same audit trail | 活动流嵌审计 |
| AI → Page | 已有 Result Actions，保持 |
| Agents templates | Phase 2 开箱四人组 |
| MCP / Marketplace | 非本设计范围 |
