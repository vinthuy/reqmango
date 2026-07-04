# ReqMango AI Phase 1 需求规格说明书

## 版本历史

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|---------|
| v1.0 | 2026-07-04 | vinthuy | 初始版本 |

---

## 1. 概述

### 1.1 背景

当前 ReqMango 拥有 22 项 AI 能力，但存在以下问题：
1. Agent 只能手动触发，未与自动化工作流集成
2. AI 功能入口分散（Chat/Create/Triage 各自独立）
3. AI 分析结果只能查看，无法转化为可执行的工作项

### 1.2 目标

Phase 1（Quick Wins）聚焦 3 个低投入高影响改进，使 AI 从"被动工具"进化为"主动协作者"。

---

## 2. 功能需求

---

### FR-01: Agent 自动化触发器

#### 需求描述

允许用户在自动化规则中选择 `dispatch_agent` 动作类型，将 Agent 任务调度绑定到工作流事件。

当指定的触发器事件发生时（如 issue_created / state_changed），系统自动调用 Agent 执行预定义的分析/分类/分配任务。

#### 验收标准

| ID | 验收标准 | 优先级 |
|----|---------|--------|
| AC-01.1 | 自动化规则创建/编辑表单中，动作类型下拉增加 `dispatch_agent` 选项 | P0 |
| AC-01.2 | `dispatch_agent` 动作可配置：目标 Agent ID、任务描述 | P0 |
| AC-01.3 | 当自动化规则被触发且条件匹配时，系统自动调用 AgentService.DispatchAgent() | P0 |
| AC-01.4 | Agent 调度结果记录在 AgentActivity 表中 | P1 |
| AC-01.5 | Agent 调度失败不应阻塞其他自动化规则的执行 | P0 |
| AC-01.6 | 支持在现有 5 种触发器类型（issue_created/updated/state_changed/assignee_changed/comment_added）上绑定 Agent | P0 |
| AC-01.7 | 后端 ActionExecutor 注册 `dispatch_agent` 处理器 | P0 |

#### 方案设计

**架构决策**：在 `DefaultActionExecutor` 中注册新的 `dispatch_agent` 动作处理器，通过依赖注入将 `AgentService` 传递给 `AutomationService`。

```
事件流:
  Issue 创建/更新 → EventBus.Publish(event)
    → AutomationService.handleAutomationEvent()
      → 匹配规则，评估条件
      → ActionExecutor.Execute(actions)
        → dispatch_agent handler
          → AgentService.DispatchAgent(task, context)
            → LLM 调用，执行工具
            → 记录 AgentActivity
```

**后端改动**：
1. `automation_service.go`: `DefaultActionExecutor` 添加 `agentSvc` 字段
2. `automation_service.go`: 注册 `dispatch_agent` handler
3. `router.go`: `NewAutomationService` 传入 `agentSvc` 引用

**前端改动**：
1. `ProjectSettings.vue` 自动化创建表单的动作类型增加 `dispatch_agent` 选项

---

### FR-02: 统一 AI Copilot 面板

#### 需求描述

将当前分散的 AI 入口（浮动 FAB 按钮 → AIChatSidebar、工具栏按钮 → AICreateDialog）整合为一个统一的 `AICopilot` 侧边面板。

#### 验收标准

| ID | 验收标准 | 优先级 |
|----|---------|--------|
| AC-02.1 | FAB 按钮 `🤖` 点击打开统一的 AI Copilot 面板 | P0 |
| AC-02.2 | 面板顶部包含模式切换 Tab：Ask / Build / Create / Chart | P0 |
| AC-02.3 | Ask 模式：与现有 AI Chat 功能一致（SSE 流式对话+工具调用） | P0 |
| AC-02.4 | Build 模式：与现有 Build 模式一致（预览→确认→执行） | P0 |
| AC-02.5 | Create 模式：内嵌原有的 AICreateDialog 表单（NL→预览→创建） | P0 |
| AC-02.6 | Chart 模式：与现有图表生成功能一致 | P0 |
| AC-02.7 | 面板底部保留快捷操作按钮（分析项目/规划 Sprint/分类 Intake） | P1 |
| AC-02.8 | Ctrl+J 快捷键切换面板显示/隐藏 | P0 |
| AC-02.9 | 面板宽度 480px，响应式适配 | P1 |

#### 方案设计

**架构决策**：创建新的 `AICopilot.vue` 组件，整合 `AIChatSidebar.vue` 和 `AICreateDialog.vue` 的功能。保留原有组件代码，新组件通过模式切换复用现有逻辑。

```
组件结构:
  AICopilot.vue (新)
  ├── TabBar (Ask | Build | Create | Chart)
  ├── AskMode    ← 复用 AIChatSidebar 的聊天逻辑
  ├── BuildMode  ← 复用 AIChatSidebar 的 Build 模式逻辑
  ├── CreateMode ← 复用 AICreateDialog 的逻辑
  └── ChartMode  ← 复用 AIChatSidebar 的 Chart 模式逻辑
```

**前端改动**：
1. 新建 `AICopilot.vue` 组件
2. 更新 `Project.vue` 将 `AIChatSidebar` 替换为 `AICopilot`
3. 移除 `AICreateDialog` 的独立入口（功能合并到 Copilot）
4. 保留 `AIChatSidebar.vue` 和 `AICreateDialog.vue` 文件以备回退

---

### FR-03: AI 结果操作化

#### 需求描述

AI 对话的每次有意义输出（分析结果、搜索列表、图表、Agent 执行结果）都应附带可操作的按钮，让用户一键将洞察转化为行动。

#### 验收标准

| ID | 验收标准 | 优先级 |
|----|---------|--------|
| AC-03.1 | AI 分析结果下方显示 `📋 生成工作项` 按钮 | P0 |
| AC-03.2 | 点击 `生成工作项` 打开创建表单，预填充 AI 分析出的关键信息 | P0 |
| AC-03.3 | AI 图表结果下方显示 `📊 添加到仪表盘` 按钮 | P1 |
| AC-03.4 | AI 搜索结果支持 `全选 → 批量创建子任务` | P1 |
| AC-03.5 | Agent 执行结果支持 `📝 保存为页面` 按钮 | P2 |
| AC-03.6 | 操作按钮仅在 AI 返回结构化结果时显示（非普通聊天文本） | P0 |
| AC-03.7 | 生成的预填充表单支持用户编辑后再确认创建 | P0 |

#### 方案设计

**架构决策**：在 AI Chat 消息渲染中，检测结构化数据（tool_result / chart / analysis），自动渲染对应的操作按钮组。按钮触发对应的创建流程。

```
AI 消息类型 → 操作按钮映射:
  tool_result(search_issues)   → [📋 批量创建子任务] [📊 导出CSV]
  tool_result(get_project_stats) → [📊 添加到仪表盘] [📝 生成报告]
  tool_result(get_issue)       → [📋 创建关联任务] [📝 生成摘要]
  chart (AIChartResponse)      → [📊 保存到仪表盘] [📝 添加到页面]
  agent_dispatch_result        → [📋 创建跟进任务] [📝 保存结果]
  analyze (AIAnalyzeResponse)  → [📋 生成改进任务] [📊 创建仪表盘]
```

**前端改动**：
1. `AIChatSidebar.vue` 消息气泡增加操作按钮区域
2. 新建 `AIResultActions.vue` 子组件
3. 新建 `AIQuickCreateDialog.vue` 弹窗组件（预填充+编辑+确认）

**后端改动**：
无需后端改动。操作按钮通过现有的 Issue/Create、Page/Create、Dashboard API 实现。

---

## 3. 非功能需求

| ID | 需求 | 描述 |
|----|------|------|
| NFR-01 | 性能 | Agent 自动化触发延迟 < 3 秒（含 LLM 调用） |
| NFR-02 | 可靠性 | Agent 调度失败不影响原工作流（隔离执行） |
| NFR-03 | 兼容性 | 所有新功能兼容 DeepSeek/Anthropic/OpenAI 三种 Provider |
| NFR-04 | 可用性 | AI Copilot 面板首次打开渲染时间 < 500ms |

---

## 4. 测试策略

| 层级 | 类型 | 覆盖 |
|------|------|------|
| 单元测试 | Go test | ActionExecutor dispatch_agent handler |
| 集成测试 | Go test | EventBus → Agent 调度链路 |
| API 测试 | curl / Postman | 自动化规则 CRUD（含 dispatch_agent 动作） |
| E2E 测试 | Playwright | AI Copilot 面板模式切换、操作按钮生成工作项 |
| 人工验收 | 手动 | Agent 自动触发效果、AI 结果操作化体验 |
