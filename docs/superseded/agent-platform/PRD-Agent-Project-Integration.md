# ReqMango Agent-Project Integration 产品规格说明书

> 版本：v2.1  
> 日期：2026-07-26  
> 状态：✅ 评审通过

---

## 一、产品背景

### 1.1 现状

ReqMango 已具备完整的项目管理能力（Issue、Cycle、Module、Page、Dashboard、Workflow 等）和独立的 AI Agent 平台（Agent、Skill、Squad、Loop、Pipeline、Memory 等）。但两者之间**缺乏深度集成**，AI Agent 目前只能在独立的 `/agents` 页面操作，无法直接参与项目工作项的生命周期。

### 1.2 目标

将 AI Agent 从"独立工具"升级为"项目团队成员"，让 Agent 像人类成员一样参与工作项的全生命周期：**需求分析 → 概要设计 → 详细设计 → 开发 → 测试 → 运营**。

### 1.3 核心理念：AI 驱动的全生命周期协作

```
┌─────────────────────────────────────────────────────────────────┐
│                    ReqMango AI-Driven Lifecycle                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐ │
│  │ 需求分析  │ →  │ 概要设计  │ →  │ 详细设计  │ →  │   开发   │ │
│  │ PM Agent  │    │ 架构Agent │    │ Dev Agent │    │ CLI集成  │ │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘ │
│       ↓               ↓               ↓               ↓        │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐ │
│  │   测试   │ →  │   运营   │ →  │   反馈   │ →  │   迭代   │ │
│  │ QA Agent │    │ Ops Agent│    │ 数据分析 │    │ 持续改进 │ │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘ │
│                                                                 │
│  ═══════════════ Wiki 文档贯穿全程 ═══════════════════════════  │
│  需求规格书 → 概要设计 → 详细设计 → API文档 → 测试报告 → 运营手册 │
└─────────────────────────────────────────────────────────────────┘
```

### 1.4 竞品参考

以 **Linear for Agents** 为主要参考对象，同时参考 Jira Agent Sessions、CrewAI 多 Agent 协作框架。

---

## 二、Linear Agent 设计理念分析

### 2.1 Linear 的核心设计原则

| 原则 | 说明 |
|------|------|
| **Agent = 一等公民** | Agent 和人类用户一样，是工作区的正式成员，有完整 profile |
| **无缝分配** | Issue 可直接分配给 Agent，和分配给人类一样简单 |
| **进度透明** | Agent 在 Issue timeline 中流式更新进度，人类可实时追踪 |
| **人工兜底** | Agent 遇到问题时发起 Human escalation，按标准格式请求介入 |
| **批量并行** | 支持同时分配多个 Issue 给不同 Agent 并行执行 |
| **可配置** | 每个 Issue 可配置 Agent 的模型、分支、行为规则 |

### 2.2 Linear Agent 工作流

```
用户创建 Issue → 分配给 Agent → Agent 分析 Issue
    ↓
Agent 打开 draft PR → 流式更新进度 → 完成后请求 Review
    ↓
人工 Review → 合并/修改 → Issue 自动关闭
```

### 2.3 Linear vs ReqMango 功能对比

| 功能维度 | Linear | ReqMango 现状 | 差距分析 |
|----------|--------|---------------|----------|
| **Agent 身份** | Agent 是工作区一等公民，有用户 profile | Agent 独立管理，非项目成员 | 需要打通 |
| **Issue 分配** | 直接分配 Issue 给 Agent | AgentTask 与 Issue 仅通过 IssueID 弱关联 | 需要增强 |
| **进度回写** | Agent 在 timeline 流式更新 | AgentTask 状态独立，不回写 Issue | 需要新增 |
| **Human Escalation** | 标准化升级格式 | 无此机制 | 需要新增 |
| **批量执行** | 多 Issue 并行分配给 Agent | Squad 支持多 Agent，但无 Issue 批量分配 | 需要增强 |
| **成本追踪** | 按 PR 估算 Token 成本 | AgentTask 有 Token 追踪，但无项目级汇总 | 需要增强 |
| **Agent 配置** | 每个 Issue 可配置模型/分支 | AgentConfig 全局配置 | 已有，可沿用 |
| **Team 集成** | Agent 可添加到 Team | Squad 独立，未与 Project Team 打通 | 需要打通 |
| **Agent Sessions** | 统一视图追踪所有 Agent 进展 | AgentSession 已实现，但无项目级视图 | 需要增强 |

---

## 三、功能规格说明

### 3.1 Agent 作为项目虚拟成员

**用户故事：** 作为项目经理，我可以将 Agent 添加到项目团队中，像管理人类成员一样管理 AI 成员。

#### 3.1.1 数据模型

```go
// ProjectAgentMember 表示 Agent 在项目中的成员身份
type ProjectAgentMember struct {
    BaseModel
    ProjectID uint64      // 项目 ID
    AgentID   uint64      // Agent ID
    Role      string      // observer | member | admin
    IsActive  bool        // 是否活跃
}
```

#### 3.1.2 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/projects/:projectId/agent-members` | 获取项目 Agent 成员列表 |
| POST | `/api/v1/projects/:projectId/agent-members` | 添加 Agent 到项目 |
| PATCH | `/api/v1/projects/:projectId/agent-members/:agentId` | 更新 Agent 角色 |
| DELETE | `/api/v1/projects/:projectId/agent-members/:agentId` | 移除 Agent |

#### 3.1.3 前端交互

- 项目成员页面新增 **"AI 成员"** 标签页
- 支持从 Agent 列表中选择添加
- Agent 显示为虚拟成员卡片（带 AI 标识）
- 支持设置角色（observer/member/admin）

---

### 3.2 工作项 Agent 分配

**用户故事：** 作为团队成员，我可以将 Issue 直接分配给 Agent，Agent 会自动执行并更新状态。

#### 3.2.1 数据模型扩展

```go
// Issue 增加字段
type Issue struct {
    // ... 现有字段 ...
    AgentAssigneeID *uint64  // Agent 指派人
    AgentTaskID     *uint64  // 关联的 Agent 任务
}
```

#### 3.2.2 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/issues/:issueId/assign-agent` | 分配 Issue 给 Agent |
| DELETE | `/api/v1/issues/:issueId/unassign-agent` | 取消 Agent 分配 |
| GET | `/api/v1/issues/:issueId/agent-status` | 获取 Agent 执行状态 |

**请求体：**
```json
{
  "agent_id": 123,
  "task_type": "execute",
  "priority": "normal",
  "instructions": "可选的额外指令"
}
```

**响应：**
```json
{
  "agent_task_id": 456,
  "status": "enqueued",
  "estimated_duration": "10-30 minutes"
}
```

#### 3.2.3 执行流程

```
用户分配 Issue 给 Agent
    ↓
系统自动创建 AgentTask（关联 IssueID）
    ↓
Agent 认领任务 → 状态更新为 claimed
    ↓
Agent 开始执行 → 状态更新为 running
    ↓
Agent 在 Issue Activity 中记录进度
    ↓
完成：更新 Issue 状态为 Done，记录结果
失败：发起 Human Escalation，标记 Issue 为 Blocked
```

#### 3.2.4 前端交互

- Issue 详情页增加 **"分配给 AI"** 按钮
- 分配后显示 Agent 执行状态面板（实时进度）
- Issue Activity 时间线中显示 Agent 的操作记录
- Issue 属性栏显示 Agent 头像和状态

---

### 3.3 AI 辅助工作项分解

**用户故事：** 作为产品经理，我可以让 AI 分析一个复杂 Issue 并自动生成子任务建议。

#### 3.3.1 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/issues/:issueId/ai-decompose` | AI 分解工作项 |

**请求体：**
```json
{
  "depth": 2,
  "strategy": "balanced",
  "squad_id": null
}
```

**响应：**
```json
{
  "sub_issues": [
    {
      "name": "子任务1：...",
      "description": "...",
      "issue_type": "task",
      "priority": "high",
      "estimated_effort": "3-5h"
    }
  ],
  "decomposition_rationale": "...",
  "suggested_assignments": {
    "human": [...],
    "agent": [...]
  }
}
```

#### 3.3.2 前端交互

- Issue 详情页增加 **"AI 分解"** 按钮
- 弹出分解配置面板（深度、策略）
- 展示 AI 生成的子任务列表（可编辑）
- 支持一键创建为子 Issue
- 支持批量分配（人类/Agent）

---

### 3.4 Agent 进度回写

**用户故事：** 作为团队成员，我可以在 Issue 的 Activity 时间线中看到 Agent 的实时执行进度。

#### 3.4.1 实现方案

Agent 执行过程中，自动在关联 Issue 的 Activity 中记录：

| 事件类型 | 说明 |
|----------|------|
| `agent_assigned` | Agent 被分配到 Issue |
| `agent_started` | Agent 开始执行 |
| `agent_progress` | Agent 进度更新（每 N 分钟或关键节点） |
| `agent_completed` | Agent 完成执行 |
| `agent_failed` | Agent 执行失败 |
| `agent_escalation` | Agent 发起人工升级 |

#### 3.4.2 数据结构

```go
type IssueActivity struct {
    // ... 现有字段 ...
    AgentID    *uint64  // 关联的 Agent
    AgentTaskID *uint64 // 关联的 AgentTask
}
```

---

### 3.5 Human Escalation（人工升级）

**用户故事：** 当 Agent 遇到无法解决的问题时，它会按照标准格式请求人工介入。

#### 3.5.1 升级格式

```
## Human Escalation

**Agent**: [Agent Name]  
**Issue**: [Issue Link]  
**Blocked Since**: [Timestamp]  
**Reason**: [具体原因]

### Context
[Agent 已完成的工作和当前状态]

### Requested Action
[需要人类执行的具体操作]

### Suggested Resolution
[Agent 建议的解决方案（如有）]
```

#### 3.5.2 触发条件

1. Agent 执行超过最大重试次数
2. Agent 遇到权限不足
3. Agent 遇到需要人类判断的歧义
4. Agent 遇到外部依赖阻塞

---

### 3.6 项目 Squad 关联

**用户故事：** 作为项目经理，我可以将 Squad 关联到项目，作为项目的 AI 团队配置。

#### 3.6.1 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/projects/:projectId/squads` | 关联 Squad 到项目 |
| GET | `/api/v1/projects/:projectId/squads` | 获取项目关联的 Squad |
| DELETE | `/api/v1/projects/:projectId/squads/:squadId` | 解除关联 |

#### 3.6.2 前端交互

- 项目设置增加 **"AI 团队"** 配置
- 显示已关联的 Squad 列表
- 支持从 Squad 列表中选择添加
- 关联后可在 Issue 分配时选择 Squad 执行

---

### 3.7 项目级 Agent 监控

**用户故事：** 作为项目经理，我可以在项目仪表盘中看到 Agent 的工作统计。

#### 3.7.1 新增 Widget

| Widget | 说明 |
|--------|------|
| **Agent Task Summary** | Agent 任务完成数、失败数、进行中数 |
| **Agent Cost** | Agent Token 消耗和成本统计 |
| **Agent Timeline** | Agent 执行时间线（类似燃尽图） |
| **Agent Utilization** | Agent 工作负载分布 |

#### 3.7.2 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/projects/:projectId/ai/stats` | 项目级 AI 统计 |

---

### 3.8 Wiki 与 AI 深度集成

**用户故事：** 作为团队成员，我可以在 Wiki 文档中看到 AI 生成的需求分析、概要设计、详细设计，并且这些文档与工作项双向关联。

#### 3.8.1 设计理念

Wiki 文档贯穿项目全生命周期，每个阶段由对应的 Agent 角色协作完成：

```
Issue (工作项)
  │
  ├── 📄 需求规格书 (Wiki Page)
  │     └── PM Agent 生成 → 产品经理审核
  │
  ├── 📄 概要设计 (Wiki Page)
  │     └── 架构 Agent 生成 → 架构师审核
  │
  ├── 📄 详细设计 (Wiki Page)
  │     └── Dev Agent 生成 → 开发团队审核
  │
  ├── 📄 API 文档 (Wiki Page)
  │     └── Dev Agent 自动生成
  │
  ├── 📄 测试报告 (Wiki Page)
  │     └── QA Agent 生成 → 测试团队审核
  │
  └── 📄 运营手册 (Wiki Page)
        └── Ops Agent 生成 → 运营团队审核
```

#### 3.8.2 数据模型扩展

```go
// IssuePage 扩展（已有，增强）
type IssuePage struct {
    IssueID    uint64  // 关联的工作项
    PageID     uint64  // 关联的 Wiki 页面
    DocType    string  // requirement | hld | lld | api_doc | test_report | ops_manual
    AgentID    *uint64 // 生成该文档的 Agent
    Version    int     // 文档版本
    Status     string  // draft | review | approved | archived
}

// WikiPage 增加字段
type WikiPage struct {
    // ... 现有字段 ...
    AIGenerated  bool       // 是否 AI 生成
    AgentID      *uint64    // 生成的 Agent
    ReviewStatus string     // pending | approved | rejected
    ReviewerID   *uint64    // 审核人
}
```

#### 3.8.3 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/issues/:issueId/wiki/generate` | AI 生成 Wiki 文档 |
| GET | `/api/v1/issues/:issueId/wiki` | 获取工作项关联的 Wiki 文档 |
| POST | `/api/v1/pages/:pageId/ai/review` | AI 审核 Wiki 文档 |
| POST | `/api/v1/pages/:pageId/ai/improve` | AI 改进 Wiki 文档 |

**生成请求体：**
```json
{
  "doc_type": "requirement",
  "agent_id": 123,
  "template_id": null,
  "instructions": "可选的额外指令"
}
```

**生成响应：**
```json
{
  "page_id": 456,
  "title": "用户登录功能 - 需求规格书",
  "content": "...",
  "status": "draft",
  "ai_metadata": {
    "model": "deepseek-chat",
    "tokens_used": 2500,
    "generation_time": "15s"
  }
}
```

#### 3.8.4 前端交互

- Issue 详情页增加 **"AI 文档"** 标签页
- 显示文档类型卡片（需求/设计/测试/运营）
- 每个卡片显示状态（草稿/审核中/已批准）
- 支持一键生成、审核、改进
- Wiki 页面显示 AI 生成标识和审核状态

---

### 3.9 设计 Team（多 Agent 角色协作）

**用户故事：** 作为项目经理，我可以配置一个"设计团队"，由多个专业 Agent 角色协作完成工作项的全生命周期。

#### 3.9.1 设计 Team 角色定义

| 角色 | Agent 模板 | 职责 | 输出物 |
|------|-----------|------|--------|
| **产品经理** | PM Agent | 需求分析、用户故事编写、验收标准定义 | 需求规格书 (Wiki) |
| **架构师** | Architect Agent | 概要设计、技术选型、架构决策 | 概要设计文档 (Wiki) |
| **核心开发** | Dev Agent | 详细设计、API 设计、代码实现 | 详细设计 + 代码 |
| **测试工程师** | QA Agent | 测试用例设计、验收测试、Bug 提出 | 测试报告 + Bug Issue |
| **运营专员** | Ops Agent | 运营手册、工单转换、用户通知 | 运营手册 + 工单 |

#### 3.9.2 设计 Team 配置

```go
// DesignTeam 表示一个设计团队配置
type DesignTeam struct {
    BaseModel
    Name        string  // 团队名称
    ProjectID   uint64  // 所属项目
    WorkspaceID uint64  // 所属工作区
    Description string  // 团队描述
    IsActive    bool    // 是否启用
}

// DesignTeamMember 表示设计团队成员
type DesignTeamMember struct {
    BaseModel
    TeamID    uint64  // 团队 ID
    AgentID   uint64  // Agent ID
    Role      string  // pm | architect | developer | tester | ops
    IsLead    bool    // 是否为负责人
    SortOrder int     // 排序
}
```

#### 3.9.3 设计 Team 工作流

```
工作项创建
    ↓
PM Agent 自动生成需求规格书 (Wiki)
    ↓ 产品经理审核
架构师 Agent 自动生成概要设计 (Wiki)
    ↓ 架构师审核
Dev Agent 自动生成详细设计 (Wiki)
    ↓ 开发团队审核
Dev Agent 通过 CLI 执行开发（trae-cli/claude）
    ↓ 代码提交
QA Agent 执行验收测试
    ↓ 测试通过
    ├── 通过 → Ops Agent 生成运营手册
    └── 失败 → QA Agent 创建 Bug Issue → 回到 Dev Agent
```

#### 3.9.4 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/projects/:projectId/design-teams` | 获取设计团队列表 |
| POST | `/api/v1/projects/:projectId/design-teams` | 创建设计团队 |
| PUT | `/api/v1/projects/:projectId/design-teams/:teamId` | 更新设计团队 |
| DELETE | `/api/v1/projects/:projectId/design-teams/:teamId` | 删除设计团队 |
| POST | `/api/v1/projects/:projectId/design-teams/:teamId/members` | 添加成员 |
| DELETE | `/api/v1/projects/:projectId/design-teams/:teamId/members/:agentId` | 移除成员 |
| POST | `/api/v1/issues/:issueId/design-team/execute` | 执行设计团队流程 |

#### 3.9.5 前端交互

- 项目设置增加 **"设计团队"** 配置页面
- 拖拽排列 Agent 角色
- 每个角色可选择不同的 Agent 模板
- Issue 详情页显示设计团队执行状态
- 支持手动触发或自动触发

---

### 3.10 多 Agent 流程编排引擎

**用户故事：** 作为项目经理，我可以可视化地编排多 Agent 协作流程，灵活增减 Agent 节点，确保 Agent 之间的上下文交接清晰完整。

#### 3.10.1 设计理念

多 Agent 流程应该是**可编排、可配置、可观测**的：

```
┌─────────────────────────────────────────────────────────────────┐
│                   Agent Workflow Orchestration                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐ │
│  │  Trigger  │ →  │  Agent A │ →  │  Agent B │ →  │  Agent C │ │
│  │  (事件)   │    │  (PM)    │    │  (架构)   │    │  (开发)   │ │
│  └──────────┘    └────┬─────┘    └────┬─────┘    └────┬─────┘ │
│                       │               │               │        │
│                  ┌────▼─────┐    ┌────▼─────┐    ┌────▼─────┐ │
│                  │ Context  │    │ Context  │    │ Context  │ │
│                  │ Handoff  │    │ Handoff  │    │ Handoff  │ │
│                  └──────────┘    └──────────┘    └──────────┘ │
│                                                                 │
│  ════════════════ 流程编排层 ════════════════════════════════  │
│  可视化配置 | 条件分支 | 并行执行 | 循环重试 | 上下文传递       │
└─────────────────────────────────────────────────────────────────┘
```

#### 3.10.2 流程定义模型

```go
// AgentWorkflow 表示一个可编排的 Agent 工作流
type AgentWorkflow struct {
    BaseModel
    Name        string          // 工作流名称
    Description string          // 描述
    ProjectID   uint64          // 所属项目
    WorkspaceID uint64          // 所属工作区
    Version     int             // 版本号
    IsActive    bool            // 是否启用
    Config      json.RawMessage // 流程配置 JSON
    TriggerType string          // manual | event | cron | webhook
    TriggerConfig json.RawMessage // 触发配置
}

// WorkflowNode 表示流程中的一个节点（Agent 步骤）
type WorkflowNode struct {
    BaseModel
    WorkflowID    uint64          // 所属工作流
    AgentID       uint64          // 执行的 Agent
    NodeType      string          // agent | condition | parallel | loop | gate
    Name          string          // 节点名称
    Config        json.RawMessage // 节点配置
    ContextConfig json.RawMessage // 上下文交接配置
    SortOrder     int             // 排序
    Timeout       int             // 超时时间（秒）
    RetryPolicy   string          // retry | skip | abort
    MaxRetries    int             // 最大重试次数
}

// WorkflowEdge 表示节点之间的连接（流转规则）
type WorkflowEdge struct {
    BaseModel
    WorkflowID   uint64          // 所属工作流
    SourceNodeID uint64          // 源节点
    TargetNodeID uint64          // 目标节点
    Condition    string          // 条件表达式（可选）
    ContextMapping string        // 上下文字段映射 JSON
}

// WorkflowRun 表示一次工作流执行
type WorkflowRun struct {
    BaseModel
    WorkflowID   uint64          // 工作流 ID
    IssueID      *uint64         // 关联的 Issue（可选）
    Status       string          // pending | running | completed | failed | cancelled
    CurrentNode  *uint64         // 当前执行的节点
    Context      json.RawMessage // 全局上下文
    StartedAt    *time.Time      // 开始时间
    CompletedAt  *time.Time      // 完成时间
    TotalTokens  int             // 总 Token 消耗
    TotalCost    float64         // 总成本
}

// WorkflowNodeRun 表示一次节点执行
type WorkflowNodeRun struct {
    BaseModel
    WorkflowRunID  uint64          // 工作流运行 ID
    NodeID         uint64          // 节点 ID
    AgentID        uint64          // Agent ID
    Status         string          // pending | running | completed | failed | skipped
    InputContext   json.RawMessage // 输入上下文
    OutputContext  json.RawMessage // 输出上下文
    StartedAt      *time.Time
    CompletedAt    *time.Time
    TokensUsed     int
    Cost           float64
    ErrorInfo      string          // 错误信息
    RetryCount     int             // 重试次数
}
```

#### 3.10.3 上下文交接机制

**核心原则：** 每个 Agent 节点的输出成为下一个节点的输入，上下文通过标准化的 `ContextPayload` 传递。

```go
// ContextPayload 是 Agent 之间传递的标准化上下文
type ContextPayload struct {
    // 工作项信息
    Issue *IssueContext `json:"issue,omitempty"`
    
    // 文档信息
    Documents []DocumentContext `json:"documents,omitempty"`
    
    // 前置 Agent 的输出
    PreviousOutputs map[string]AgentOutput `json:"previous_outputs,omitempty"`
    
    // 全局共享数据
    SharedData map[string]interface{} `json:"shared_data,omitempty"`
    
    // 元数据
    Metadata *WorkflowMetadata `json:"metadata,omitempty"`
}

type IssueContext struct {
    ID          uint64   `json:"id"`
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Type        string   `json:"type"`
    Priority    string   `json:"priority"`
    Acceptance  []string `json:"acceptance_criteria"`
}

type DocumentContext struct {
    PageID   uint64 `json:"page_id"`
    DocType  string `json:"doc_type"`  // requirement | hld | lld | api_doc | test_report
    Title    string `json:"title"`
    Content  string `json:"content"`
    Version  int    `json:"version"`
}

type AgentOutput struct {
    AgentID     uint64                 `json:"agent_id"`
    AgentName   string                 `json:"agent_name"`
    NodeType    string                 `json:"node_type"`  // pm | architect | developer | tester | ops
    Output      map[string]interface{} `json:"output"`
    TokensUsed  int                    `json:"tokens_used"`
    Duration    string                 `json:"duration"`
    Status      string                 `json:"status"`
}

type WorkflowMetadata struct {
    WorkflowID    uint64 `json:"workflow_id"`
    WorkflowName  string `json:"workflow_name"`
    RunID         uint64 `json:"run_id"`
    NodeID        uint64 `json:"node_id"`
    NodeName      string `json:"node_name"`
    TriggerType   string `json:"trigger_type"`
    StartedAt     string `json:"started_at"`
}

// ParallelMergeStrategy 并行节点的上下文合并策略
type ParallelMergeStrategy struct {
    Strategy   string            // concat | merge | latest | custom
    FieldRules map[string]string // 字段级合并规则
}
```

**执行关系说明：**
```
WorkflowNodeRun → 调用 → AgentTask → 实际执行
```

#### 3.10.4 上下文映射配置

每个 Edge 可以配置上下文字段映射，控制数据如何从一个节点传递到下一个节点：

```json
{
  "source_node_id": 1,
  "target_node_id": 2,
  "context_mapping": {
    "issue": "$.issue",
    "documents": "$.documents[?(@.doc_type=='requirement')]",
    "agent_output": "$.previous_outputs.node_1",
    "custom_fields": {
      "architecture_decisions": "$.previous_outputs.node_1.output.decisions",
      "tech_stack": "$.previous_outputs.node_1.output.tech_stack"
    }
  }
}
```

#### 3.10.5 流程编排 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/projects/:projectId/workflows` | 获取工作流列表 |
| POST | `/api/v1/projects/:projectId/workflows` | 创建工作流 |
| PUT | `/api/v1/projects/:projectId/workflows/:workflowId` | 更新工作流 |
| DELETE | `/api/v1/projects/:projectId/workflows/:workflowId` | 删除工作流 |
| POST | `/api/v1/projects/:projectId/workflows/:workflowId/nodes` | 添加节点 |
| PUT | `/api/v1/projects/:projectId/workflows/:workflowId/nodes/:nodeId` | 更新节点 |
| DELETE | `/api/v1/projects/:projectId/workflows/:workflowId/nodes/:nodeId` | 删除节点 |
| POST | `/api/v1/projects/:projectId/workflows/:workflowId/edges` | 添加连接 |
| PUT | `/api/v1/projects/:projectId/workflows/:workflowId/edges/:edgeId` | 更新连接 |
| DELETE | `/api/v1/projects/:projectId/workflows/:workflowId/edges/:edgeId` | 删除连接 |
| POST | `/api/v1/projects/:projectId/workflows/:workflowId/execute` | 执行工作流 |
| GET | `/api/v1/projects/:projectId/workflows/:workflowId/runs` | 获取执行历史 |
| GET | `/api/v1/projects/:projectId/workflows/:workflowId/runs/:runId` | 获取执行详情 |
| GET | `/api/v1/projects/:projectId/workflows/:workflowId/runs/:runId/nodes` | 获取节点执行详情 |

#### 3.10.6 流程编排前端

**可视化流程设计器：**
- 拖拽式节点添加（从 Agent 列表拖入）
- 连线配置（点击节点间连线配置条件和上下文映射）
- 节点配置面板（Agent 选择、超时、重试策略）
- 条件分支配置（if/else 逻辑）
- 并行执行配置（fork/join）
- 实时预览和调试

**执行监控面板：**
- 流程图实时高亮当前执行节点
- 每个节点的输入/输出上下文可视化
- Token 消耗和成本追踪
- 错误日志和重试状态
- 支持暂停/恢复/取消

#### 3.10.7 预设工作流模板

| 模板名 | 节点 | 适用场景 |
|--------|------|----------|
| **需求到上线** | PM → 架构 → 开发 → 测试 → 运营 | 完整功能开发 |
| **快速修复** | 开发 → 测试 | Bug 修复 |
| **设计评审** | PM → 架构 | 技术方案评审 |
| **代码审查** | 开发 → 测试 → 开发 | 代码质量保障 |
| **发布流程** | 测试 → 运营 | 版本发布 |

#### 3.10.8 Workflow/Pipeline/Loop 功能定位

| 功能 | 定位 | 适用场景 |
|------|------|----------|
| **Workflow** | 跨Agent协作编排 | 需求→设计→开发→测试 完整流程 |
| **Pipeline** | 单Agent多步骤执行 | 代码审查→测试→部署 单一任务 |
| **Loop** | 单Agent循环迭代 | 持续优化、Sprint守护 |

---

### 3.11 开发 CLI 集成

**用户故事：** 作为开发者，我可以让 Agent 通过 trae-cli/claude 执行实际的代码开发任务。

#### 3.11.1 集成方式

```
ReqMango (Issue 分配)
    ↓
Agent Task 创建
    ↓
Agent 通过 CLI 执行开发
    ├── trae-cli: 代码生成、重构、优化
    ├── claude: 代码审查、问题解答
    └── github copilot: 代码补全
    ↓
代码提交 → PR 创建
    ↓
自动关联到 Issue
```

#### 3.11.2 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/agent-tasks/:taskId/cli/execute` | 通过 CLI 执行任务 |
| GET | `/api/v1/agent-tasks/:taskId/cli/status` | 获取 CLI 执行状态 |
| GET | `/api/v1/agent-tasks/:taskId/cli/logs` | 获取 CLI 执行日志 |

**执行请求体：**
```json
{
  "cli_tool": "trae-cli",
  "command": "generate",
  "params": {
    "file_path": "src/auth/login.ts",
    "description": "实现用户登录功能",
    "context": "..."
  }
}
```

#### 3.11.3 前端交互

- Issue 详情页显示 CLI 执行状态
- 实时日志输出
- 代码变更预览
- 一键提交/放弃

---

### 3.12 测试 Agent 与 Bug 自动提出

**用户故事：** 作为测试工程师，QA Agent 可以根据工作项（US/Feature）自动执行验收测试，并在发现问题时自动创建 Bug Issue。

#### 3.12.1 测试流程

```
US/Feature Issue 创建
    ↓
QA Agent 分析验收标准
    ↓
生成测试用例 (Wiki)
    ↓
执行测试（自动化/手动模拟）
    ├── 通过 → 标记 Issue 为 Done
    └── 失败 → 自动创建 Bug Issue
         ├── 关联到原 US/Feature
         ├── 包含失败截图/日志
         └── 分配给对应开发 Agent
```

#### 3.12.2 Bug Issue 自动创建

```go
// QA Agent 创建 Bug Issue
type BugFromAgent struct {
    Title           string   // Bug 标题
    Description     string   // Bug 描述（包含复现步骤）
    Priority        string   // 优先级
    AssigneeID      *uint64  // 分配给谁
    ParentIssueID   uint64   // 关联的 US/Feature
    TestEvidence    string   // 测试证据（截图/日志）
    ReproduceSteps  []string // 复现步骤
    ExpectedResult  string   // 期望结果
    ActualResult    string   // 实际结果
}
```

#### 3.11.3 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/issues/:issueId/qa/test` | 触发 QA Agent 测试 |
| GET | `/api/v1/issues/:issueId/qa/test-cases` | 获取测试用例 |
| GET | `/api/v1/issues/:issueId/qa/test-report` | 获取测试报告 |
| POST | `/api/v1/issues/:issueId/qa/report-bug` | 报告 Bug（自动创建 Issue） |

#### 3.12.4 前端交互

- Issue 详情页增加 **"测试"** 标签页
- 显示测试用例列表和执行状态
- Bug 自动关联到原 Issue
- 测试报告可视化（通过率、覆盖率）

---

### 3.13 运营 Agent 与工单转换

**用户故事：** 作为运营人员，Ops Agent 可以将产品需求和 Bug 自动转换为运营工单，并与开发团队集成。

#### 3.13.1 工单转换流程

```
需求/Bug Issue
    ↓
Ops Agent 分析内容
    ↓
判断类型
    ├── 用户反馈 → 创建客服工单
    ├── 功能需求 → 创建运营需求单
    ├── Bug 修复 → 创建紧急修复工单
    └── 运营活动 → 创建活动工单
    ↓
自动填充工单字段
    ├── 优先级（基于影响范围）
    ├── 分配团队（基于类型）
    ├── 截止时间（基于紧急程度）
    └── 通知相关方
```

#### 3.13.2 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/issues/:issueId/ops/convert` | 转换为运营工单 |
| GET | `/api/v1/ops/tickets` | 获取运营工单列表 |
| PUT | `/api/v1/ops/tickets/:ticketId` | 更新运营工单 |
| POST | `/api/v1/ops/tickets/:ticketId/escalate` | 升级工单 |

#### 3.13.3 前端交互

- Issue 详情页增加 **"运营"** 按钮
- 转换配置面板（工单类型、优先级、分配）
- 运营工单列表页面
- 工单状态追踪

---

## 四、功能点对比审核表

### 4.1 Linear vs ReqMango 功能点对比

| # | 功能点 | Linear | ReqMango 现状 | 优先级 | 实施难度 | 状态 |
|---|--------|--------|---------------|--------|----------|------|
| 1 | Agent 是项目成员 | Agent 是一等公民 | Agent 独立管理 | **P0** | 中 | 待开发 |
| 2 | Issue 直接分配给 Agent | 和分配给人类一样 | AgentTask 弱关联 | **P0** | 中 | 待开发 |
| 3 | Agent 进度回写 Issue | timeline 流式更新 | 无此机制 | **P0** | 中 | 待开发 |
| 4 | Human Escalation | 标准化升级格式 | 无此机制 | **P0** | 低 | 待开发 |
| 5 | 多 Agent 流程编排 | 无 | Loop/Pipeline 基础 | **P0** | 高 | 待开发 |
| 6 | 上下文交接机制 | 无 | 无 | **P0** | 高 | 待开发 |
| 7 | Wiki 与 AI 集成 | 无 | Page 已实现 | **P1** | 高 | 待增强 |
| 8 | 设计 Team 多角色协作 | 无 | Squad 已实现 | **P1** | 高 | 待开发 |
| 9 | 开发 CLI 集成 | 无 | 无 | **P1** | 高 | 待开发 |
| 10 | 测试 Agent 自动测 Bug | 无 | 无 | **P1** | 高 | 待开发 |
| 11 | 运营 Agent 工单转换 | 无 | 无 | **P1** | 中 | 待开发 |
| 12 | 批量 Issue 分配给 Agent | 多 Issue 并行 | Squad 多 Agent | **P1** | 高 | 待开发 |
| 13 | 项目级 Agent 监控 | Agent Sessions | AgentSession 已实现 | **P1** | 低 | 待增强 |
| 14 | 成本追踪 per Issue | 按 PR 估算 | Token 追踪已有 | **P2** | 低 | 已有基础 |
| 15 | Agent Session 项目视图 | 统一追踪 | AgentSession 已实现 | **P2** | 低 | 待增强 |
| 16 | Agent 自动分配 Issue | 基于规则 | auto-triage 已实现 | **P2** | 低 | 已有 |
| 17 | Agent 决策可解释性 | Linear记录思考过程 | 无 | **P1** | 中 | 待开发 |
| 18 | Agent 执行失败恢复 | 无 | 无 | **P1** | 中 | 待开发 |
| 19 | 批量 Issue 分配 | Linear支持 | 无 | **P1** | 低 | 待开发 |
| 20 | 新用户引导 | 各竞品均有 | 无 | **P1** | 低 | 待开发 |
| 21 | Agent 执行预览 | Devin有模拟执行 | 无 | **P1** | 中 | 待开发 |

### 4.2 功能完整性评估

| 维度 | Linear | ReqMango 现状 | ReqMango 目标 |
|------|--------|---------------|---------------|
| **Agent 身份** | 完整（一等公民） | 不足（独立管理） | 对齐 Linear |
| **Issue 集成** | 完整（直接分配） | 不足（弱关联） | 对齐 Linear |
| **进度可见性** | 完整（timeline 更新） | 不足（无回写） | 对齐 Linear |
| **人工兜底** | 完整（escalation） | 无 | 新增 |
| **多 Agent 协作** | 有限（Team 概念） | 完整（Squad） | 超越 Linear |
| **记忆系统** | 无 | 完整（Memory） | 保持优势 |
| **循环执行** | 无 | 完整（Loop） | 保持优势 |
| **流水线** | 无 | 完整（Pipeline） | 保持优势 |
| **自动驾驶** | 无 | 完整（Autopilot） | 保持优势 |
| **监控仪表盘** | 基础 | 完整（Monitor） | 保持优势 |

### 4.3 差异化优势

ReqMango 在以下方面**领先 Linear**：

1. **多 Agent 流程编排** - 可视化编排 Agent 协作流程，Linear 无此概念
2. **上下文交接机制** - Agent 间标准化上下文传递，Linear 无此概念
3. **Wiki 与 AI 深度集成** - 文档贯穿全生命周期，Linear 无此概念
4. **设计 Team 多角色协作** - PM/架构/开发/测试/运营 Agent 协作，Linear 无此概念
5. **测试 Agent 自动测 Bug** - QA Agent 自动执行测试并创建 Bug，Linear 无此概念
6. **运营 Agent 工单转换** - 需求/Bug 自动转运营工单，Linear 无此概念
7. **Squad 多 Agent 协作** - Linear 无此概念
8. **Memory 记忆系统** - Linear 无持久化记忆
9. **Loop 循环执行** - Linear 无自动化循环
10. **Pipeline 流水线** - Linear 无 DAG 编排
11. **Autopilot 自动驾驶** - Linear 无定时/触发执行
12. **Runtime 运行时管理** - Linear 无分布式执行

---

## 五、评审决策记录

> 日期：2026-07-26
> 评审人员：项目经理、测试经理、架构师、技术骨干

### 5.1 已确认决策

| 编号 | 决策项 | 选择 | 说明 |
|------|--------|------|------|
| PM-01 | 项目级成本预算控制 | **A: Phase 1增加** | 项目设置中增加AI预算配置，支持月度预算上限，超预算时告警或阻止执行 |
| PM-02 | Agent执行SLA | **A: 定义基础SLA** | 普通任务30分钟、复杂任务2小时，超时自动触发Human Escalation |
| QA-01 | Agent输出质量评估 | **B: 暂不增加** | 先由人工审核质量，后续迭代增加自动评估 |
| TECH-01 | 流程设计器实现策略 | **B: 一步到位** | Phase 1直接实现可视化拖拽设计器 |

### 5.2 已自动修复项

| 编号 | 问题 | 修复内容 |
|------|------|----------|
| ARCH-01 | WorkflowNode与AgentTask关系 | 增加执行关系说明：WorkflowNodeRun → AgentTask → 实际执行 |
| ARCH-03 | 并行执行上下文合并 | 增加ParallelMergeStrategy（concat/merge/latest/custom） |
| TECH-03 | Workflow/Pipeline/Loop定位 | 增加功能定位说明表 |

---

## 六、实施计划

### Phase 1: 核心打通（P0）
1. 新增 `ProjectAgentMember` 模型和 API
2. 修改 `Issue` 模型增加 `AgentAssigneeID` 字段
3. 新增"分配 Issue 给 Agent" API 和自动创建 AgentTask
4. 新增 Agent 进度回写 Issue Activity 机制
5. 新增 Human Escalation 标准化格式
6. 新增多 Agent 流程编排引擎（AgentWorkflow/WorkflowNode/WorkflowEdge）
7. 新增上下文交接机制（ContextPayload 标准化）
8. 新增项目级AI成本预算配置（月度预算上限、超预算告警）
9. 新增Agent执行SLA（普通30分钟、复杂2小时、超时Escalation）
10. 新增Agent决策可解释性记录（思考过程、决策依据）
11. 新增Agent执行失败恢复机制（重试、回滚、接管）
12. 新增批量操作支持（批量分配Issue给Agent）
13. 前端：项目成员页面增加 AI 成员管理
14. 前端：Issue 详情页增加 Agent 分配和执行面板
15. 前端：可视化流程设计器（拖拽式）
16. 前端：新用户引导流程（引导弹窗、快速开始模板）
17. 前端：Agent执行预览功能（模拟执行、预估时间和成本）

### Phase 2: Wiki 与 AI 集成（P1）
1. Wiki 文档与 Issue 双向关联增强
2. AI 生成需求规格书（PM Agent）
3. AI 生成概要设计（Architect Agent）
4. AI 生成详细设计（Dev Agent）
5. AI 审核与改进 Wiki 文档
6. 前端：Issue 详情页增加"AI 文档"标签页

### Phase 3: 设计 Team（P1）
1. 新增 `DesignTeam` 和 `DesignTeamMember` 模型
2. 设计团队配置 API
3. 设计团队工作流引擎
4. 前端：项目设置增加"设计团队"配置
5. 前端：Issue 详情页显示设计团队执行状态

### Phase 4: 开发 CLI 集成（P1）
1. trae-cli/claude 集成 API
2. CLI 执行状态和日志追踪
3. 代码变更预览
4. 前端：Issue 详情页显示 CLI 执行状态

### Phase 5: 测试 Agent（P2）
1. QA Agent 测试用例生成
2. QA Agent 验收测试执行
3. Bug 自动创建和关联
4. 前端：Issue 详情页增加"测试"标签页

### Phase 6: 运营 Agent（P2）
1. 运营工单转换 API
2. Ops Agent 工单自动填充
3. 工单状态追踪
4. 前端：运营工单列表页面

### Phase 7: 增强优化（P2）
1. 成本追踪 per Issue
2. Agent Session 项目级视图
3. 批量 Issue 分配给 Agent/Squad
4. Agent 自动分配规则增强

---

## 六、非功能需求

### 6.1 性能
- Agent 分配 Issue 响应时间 < 500ms
- Agent 进度回写延迟 < 2s
- 项目级 AI 统计查询 < 1s

### 6.2 安全
- Agent 权限继承项目成员权限
- Agent 不能访问未授权的 Issue
- Agent 操作记录完整审计日志

### 6.3 可观测性
- 所有 Agent 操作可追踪
- Issue Activity 中完整记录 Agent 行为
- 项目仪表盘实时显示 Agent 状态

---

## 七、审核结论

| 审核项 | 结论 |
|--------|------|
| 产品方向 | 与 Linear 理念对齐，同时保持差异化优势 |
| 功能完整性 | P0 功能覆盖核心场景，P1/P2 逐步增强 |
| 技术可行性 | 基于现有数据模型扩展，实现难度可控 |
| 开发建议 | 先完成 P0 核心打通，再逐步增强 |

**审核状态：待用户确认**
