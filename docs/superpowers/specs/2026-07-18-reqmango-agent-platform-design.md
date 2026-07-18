# Reqmango Agent Platform — 系统架构 & 详细设计说明书

> **版本**: v1.0
> **日期**: 2026-07-18
> **作者**: vinthuy
> **状态**: Draft — 待评审

---

## 目录

1. [执行摘要](#1-执行摘要)
2. [竞品全景分析](#2-竞品全景分析)
3. [战略定位](#3-战略定位)
4. [系统架构概览](#4-系统架构概览)
5. [Harness Engine 详细设计](#5-harness-engine-详细设计)
6. [Loop Engine 详细设计](#6-loop-engine-详细设计)
7. [Phase 路线图 & 里程碑](#7-phase-路线图--里程碑)
8. [杀手场景设计: Sprint Agent Loop](#8-杀手场景设计-sprint-agent-loop)
9. [API & 集成设计](#9-api--集成设计)
10. [数据模型](#10-数据模型)
11. [安全与治理](#11-安全与治理)
12. [实施优先级](#12-实施优先级)
13. [风险评估与缓解](#13-风险评估与缓解)
14. [成功指标](#14-成功指标)

---

## 1. 执行摘要

### 1.1 背景

Reqmango 当前拥有 22 项 AI 能力和较为完整的项目管理功能体系。但在 AI Agent 领域，Atlassian Jira Agent (Rovo) 已于 2026 年 5 月 GA，每日处理 500 万+ MCP 工具调用；Plane AI 以开源 AI-Native 定位快速追赶。市场正从"AI 辅助功能"向"AI Agent 平台"范式迁移。

### 1.2 核心洞察

| 维度 | Jira Agent | Plane AI | **Reqmango 机会** |
|------|-----------|----------|-------------------|
| Agent 协作 | 单Agent调度 | 单Agent事件触发 | **多Agent编排 (Harness)** |
| 自主执行 | 无闭环 | 无闭环 | **自主Loop工程** |
| 质量保证 | 自审自评 | 自审自评 | **对抗性验证** |
| 部署 | Cloud only | 自托管 | **自托管 + BYO Keys** |
| 开源 | 闭源 | 开源 | **MIT 协议** |
| Agent扩展 | Rovo Studio (无代码) | ADK + Marketplace | **Agent SDK + 开放市场** |

### 1.3 产品愿景

> **打造项目管理领域最智能的开源 Agent 编排平台** — 让 Reqmango 不只是"又一个 Jira 替代品"，而是"用 AI Agent 重新定义项目管理的平台"。

### 1.4 核心差异化主张

1. **Harness 多Agent编排** — Planner→Executor→Reviewer 流水线，对抗性验证
2. **Loop 自主闭环** — Act→Observe→Reason→Repeat，设目标而非步骤
3. **开源 + 数据主权** — MIT 协议 + 自托管 + BYO Keys，Jira 无法复制的壁垒
4. **MCP 开放生态** — 双向 MCP，让任何 AI 工具都能接入 Reqmango

---

## 2. 竞品全景分析

### 2.1 Jira Agent (Atlassian Rovo) — 深度分析

**当前状态**: GA (2026年5月)，Jira Cloud Premium/Enterprise 独占

**Agent 矩阵**:

| Agent | 定位 | 成熟度 |
|-------|------|--------|
| Rovo Dev | AI编码+自动提PR | GA |
| Jira Delivery Agent | 项目健康监控+日报 | GA |
| Rovo Ops Agent | 事件管理+事后复盘 | GA |
| Rovo Chat (Max) | 多步推理复杂任务 | GA |
| 9个第三方Agent | Figma/Canva/Box等 | GA (via MCP) |

**核心能力**:
- Agent 作为指派人出现在 Board/Sprint/Release 中
- @mention Agent 在评论中上下文协作
- Agent 嵌入 Jira Automation 规则 (状态转换触发)
- 完整审计链 — Agent 行为与人类同权限同日志
- 双向 MCP — Rovo MCP Server (给外部Agent) + MCP Skills (接外部App)
- Rovo Studio 无代码 Agent 构建平台
- Teamwork Graph — 跨 Jira/Confluence/Code 的知识图谱

**关键数据**:
- 500万+ MCP 工具调用/工作日
- 近1/3 为写操作 (创建Issue/评论/状态更新)
- 100万+ 月活用户，93% 付费版
- >50% 企业客户采用 (6个月内)

**短板 (Reqmango 可攻击的窗口)**:
1. **Cloud only** — 无可自托管AI，数据必须上 Atlassian 云
2. **单Agent调度** — Rovo Swarm 仅是并行执行，非编排流水线
3. **无自主Loop** — Automation Mode 仅增强 Prompt，非 Act→Observe→Reason→Repeat
4. **按次付费** — $0.30/次，高频使用成本不可控
5. **外部Agent集成不成熟** — Remote Agent 仍在 EAP，实验性、不推荐生产
6. **Agent不能完全自主** — Assign Issue 需用户确认 (ROVO-115)
7. **JSM不支持** — Agent 分配和 @mention 在 Jira Service Management 不工作 (JSDCLOUD-18577)

### 2.2 Plane AI — 深度分析

**当前状态**: Beta，自托管对等

**核心能力**:
- Agent 作为指派人 + @mention
- 事件驱动 Agent: On create (auto-triage) + On assign
- Auto-triage: 自动分类/标签/分配/路由 intake
- Context-aware AI sidecar: 锚定当前视图 (工作项/Cycle/Initiative)
- MCP Server (MIT 开源, 55+ 工具, 8 类别)
- ADK (Agent Development Kit) + Marketplace
- BYO Keys + 完全自托管 AI 功能对等
- Agent Run 生命周期追踪 (created/in_progress/completed/failed/stopped/stale)

**短板**:
1. **Agent 仍是 Beta** — 无 GA 时间表，有 stale (5分钟无活动) 等问题
2. **无多Agent编排** — 没有 Harness/Pipeline 概念
3. **无自主Loop** — 仅事件触发→执行，无闭环验证
4. **无第三方Agent生态** — 不如 Jira 的 9 个 OOTB Agent
5. **无独立查询语言** — 不如 Reqmango 的 RQL
6. **无 AI Sprint Planning** — Reqmango 独有
7. **无公开采用数据** — 无法验证生产成熟度

### 2.3 竞品态势图

```
高 ┤                          ┌──────────┐
    │                          │  Jira    │ ← GA, 500万+调用/日
A   │                          │  Agent   │   企业级生态
g   │      ┌──────────┐        └──────────┘
e   │      │ Reqmango │
n   │      │  (目标)   │  ← Harness+Loop+开源
t   │      └──────────┘    自托管Agent编排
    │  ┌──────────┐
深   │  │ Reqmango │  ← 22项AI能力,Agent CRUD
度   │  │  (当前)   │    RQL,AI Sprint Planning
    │  └──────────┘
    │      ┌──────────┐
低  ┤      │  Plane   │  ← Beta, Auto-triage
    │      │   AI     │    开源MCP Server
    │      └──────────┘
    └──────────────────────────────────────
       实验期     成长期     成熟期

              Agent 成熟度 →
```

---

## 3. 战略定位

### 3.1 一句话定位

> **Reqmango = 开源自托管的项目管理 + Harness多Agent编排引擎 + Loop自主闭环**

### 3.2 目标用户画像

| 画像 | 痛点 | Reqmango 价值主张 |
|------|------|-------------------|
| 数据合规敏感企业 | 不能用 Jira Cloud | 自托管 + 数据不出域 |
| 高AI使用量团队 | Jira $0.30/次成本不可控 | BYO Keys, 零边际成本 |
| 追求深度的工程团队 | Jira Agent 太浅 (单Agent) | Harness 多Agent编排 |
| 开源社区/ISV | 想基于PM工具构建Agent应用 | MIT + Agent SDK + 市场 |
| 中国/亚洲团队 | 海外工具访问慢/不可用 | 本土部署 + 中英双语 |

### 3.3 竞争护城河

**短期护城河 (Phase 1-2)**:
- Harness 多Agent编排 (Jira 和 Plane 都没有)
- Loop 自主闭环 (Jira 和 Plane 都没有)
- 对抗性验证 (行业首创)

**中期护城河 (Phase 3-4)**:
- Agent SDK + 开源市场 (网络效应)
- Agent Graph (开源版 Teamwork Graph)
- 双向 MCP Hub (生态锁定)

**不可复制护城河**:
- MIT 开源协议 (Jira 无法转向)
- 自托管数据主权 (Jira Cloud only 无法竞争的合规场景)

### 3.4 不做什么

| 不做 | 原因 |
|------|------|
| AI 代码生成 (Rovo Dev 竞品) | 不是项目管理核心；Jira 已占位 |
| 项目管理之外的 Agent | 聚焦 PM 垂直领域深度 |
| 闭源商业版 | MIT 是核心壁垒 |
| 自建 LLM | 专注编排层，模型层用 BYO Keys |

---

## 4. 系统架构概览

### 4.1 双引擎架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    Reqmango Agent Platform                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────┐  ┌────────────────────────────┐   │
│  │     🔧 Harness Engine     │  │      🔄 Loop Engine         │   │
│  │                          │  │                            │   │
│  │  ┌──────┐ ┌──────────┐  │  │  ┌──────────────────────┐  │   │
│  │  │Planner│→│ Executor │  │  │  │   Trigger System     │  │   │
│  │  └──────┘ └────┬─────┘  │  │  │  (Event/Cron/Webhook) │  │   │
│  │                ↓         │  │  └──────────┬───────────┘  │   │
│  │  ┌──────────────┐       │  │             ↓              │   │
│  │  │   Reviewer    │       │  │  ┌──────────────────────┐  │   │
│  │  │ (对抗性验证)   │       │  │  │   Execution Context   │  │   │
│  │  └──────────────┘       │  │  │  (Worktree隔离+Budget) │  │   │
│  │                          │  │  └──────────┬───────────┘  │   │
│  │  • Pipeline 编排         │  │             ↓              │   │
│  │  • Fan-out 并行          │  │  ┌──────────────────────┐  │   │
│  │  • 动态模型路由          │  │  │   Observe & Reason    │  │   │
│  │  • Agent 间消息传递      │  │  │  (结果检查+目标判定)   │  │   │
│  └───────────┬──────────────┘  │  └──────────┬───────────┘  │   │
│              │                  │             ↓              │   │
│              │                  │  ┌──────────────────────┐  │   │
│              │                  │  │   Repeat or Stop     │  │   │
│              │                  │  │  (继续/停止/升级人工) │  │   │
│              │                  │  └──────────────────────┘  │   │
│              │                  └────────────┬───────────────┘   │
│              └───────────────┬──────────────┘                   │
│                              ↓                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              🧩 Agent SDK & Registry                      │   │
│  │  Agent定义 → 注册 → 发现 → 权限 → 审计 → 版本管理          │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐             │
│         ↓                    ↓                     ↓             │
│  ┌─────────────┐  ┌──────────────────┐  ┌─────────────────┐    │
│  │ MCP Server  │  │   Automation     │  │   RQL Engine    │    │
│  │ (工具暴露)   │  │ (事件→触发Agent) │  │ (Agent查询语言)  │    │
│  │             │  │                  │  │                 │    │
│  │ SSE/STDIO   │  │ 5种触发器类型     │  │ 词法/语法/执行器 │    │
│  │ 60+ Tools   │  │ dispatch_agent   │  │ Agent可调用     │    │
│  └─────────────┘  └──────────────────┘  └─────────────────┘    │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                   📊 Agent Observability                   │   │
│  │  Session回放 → 执行链路追踪 → 成本分析 → 质量仪表盘          │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 技术选型

| 层 | 选型 | 理由 |
|----|------|------|
| Harness 编排引擎 | Go (goroutines + channels) | 与后端一致；天然并发 |
| Loop 状态机 | Go (FSM pattern) | 确定性状态转换 |
| Worktree 隔离 | git worktree | 成熟稳定；零额外依赖 |
| Agent 间通信 | Go channels + Redis (可选) | 进程内用 channel；跨进程用 Redis |
| Agent 定义 | YAML/JSON DSL | 可读、可版本控制 |
| 前端编排可视化 | Vue 3 + Cytoscape.js / DAG 图 | 流程图展示 Pipeline |
| LLM 调用 | 现有 AI Service 复用 | DeepSeek/Anthropic/OpenAI |
| 存储 | PostgreSQL (现有) + JSONB | Agent 配置和结果灵活存储 |

### 4.3 与现有系统的关系

```
现有系统                          新增 Agent Platform
─────────                        ────────────────
AI Chat Service  ─────────────→  Agent 调用 LLM
AgentService     ─────────────→  Agent Registry 基座
AutomationRule   ─────────────→  Loop Trigger 源
MCP Server       ─────────────→  Agent Tool 暴露
RQL Engine       ─────────────→  Agent 查询语言
EventBus         ─────────────→  Harness 事件驱动
AIChatSidebar    ─────────────→  Agent Copilot 面板 (FR-02)
AICreateDialog   ─────────────→  Agent Copilot 面板 (FR-02)
```

设计原则: **不重写，只扩展**。现有 AI Service/AgentService/MCP Server 作为基座，Harness+Loop 作为编排层叠加其上。

---

## 5. Harness Engine 详细设计

### 5.1 核心概念

```
Harness = 多Agent编排框架

核心三要素:
  1. Pipeline  — 定义 Agent 执行顺序和数据流
  2. Router    — 根据任务类型/复杂度动态路由到不同Agent+模型
  3. Verifier  — 对抗性验证，独立的 Review Agent 评审输出
```

### 5.2 Pipeline 编排模型

#### 5.2.1 基础 Pipeline: Planner → Executor → Reviewer

```
                     ┌─────────────┐
                     │   Trigger   │
                     └──────┬──────┘
                            │
                     ┌──────▼──────┐
                     │   Planner   │  ← 分析任务、产出 Spec
                     │  (Opus/XHigh)│
                     └──────┬──────┘
                            │ Spec
                     ┌──────▼──────┐
                     │  Executor   │  ← 执行产出 (代码/工单/报告)
                     │  (Sonnet)   │
                     └──────┬──────┘
                            │ Artifact
                     ┌──────▼──────┐
                     │  Reviewer   │  ← 对抗性验证
                     │  (Sonnet)   │
                     └──────┬──────┘
                            │
              ┌─────────────┼─────────────┐
              │ Pass        │ Fail         │ Max Retries
         ┌────▼───┐   ┌─────▼──────┐  ┌────▼─────┐
         │ Deliver│   │ Re-Execute │  │ Escalate │
         └────────┘   │ +Feedback  │  │ to Human │
                      └────────────┘  └──────────┘
```

#### 5.2.2 高级模式

**Fan-out (并行分发)**:
```
Planner
  ├→ Executor A (前端任务)
  ├→ Executor B (后端任务)
  └→ Executor C (测试任务)
       │
  Reviewer (汇总评审)
```

**Tournament (竞赛模式)**:
```
Planner
  ├→ Executor A (方案1)
  ├→ Executor B (方案2)
  └→ Executor C (方案3)
       │
  Judge Agent (两两对比→选出最优)
```

**Classify-and-Act (分类路由)**:
```
Classifier Agent (判断任务类型)
  ├→ 前端专家Agent
  ├→ 后端专家Agent
  └→ DevOps Agent
```

### 5.3 对抗性验证 (Adversarial Verification)

#### 核心原则

> **写代码的Agent永远不能评审自己的代码。**

#### 验证协议

```
对于每个 Executor 输出:
  1. 启动 3 个独立的 Reviewer Agent
  2. 每个 Reviewer 的 Prompt 以 "Try to REFUTE this output" 开头
  3. 每个 Reviewer 单独产出 Verdict: { pass: bool, issues: [], confidence: 0-1 }
  4. 投票: ≥2 个 Reviewers 通过 → 交付; <2 → 驳回+反馈→重新执行
  5. 最多 N 次重试 (默认3次)，超出升级人工
```

#### Reviewer Agent Prompt 模板

```
You are a QA Reviewer. Your job is to find problems.

CRITICAL INSTRUCTION: Assume the following output contains errors.
Your task is to FIND them. Default to REFUTED if you are uncertain.

Output to review:
{executor_output}

Original goal:
{planner_spec}

Checklist:
1. Does the output fully satisfy the spec?
2. Are there logical errors or omissions?
3. Are there any edge cases not handled?
4. Is the quality acceptable for production?
5. Are there security or performance concerns?

Return JSON:
{
  "pass": true/false,
  "issues": [{"severity": "critical|major|minor", "description": "..."}],
  "confidence": 0-1
}
```

### 5.4 Worktree 并行隔离

```
         ┌────────────────────────────┐
         │     Agent Worktree Pool    │
         ├────────────────────────────┤
         │                            │
Trigger ─→ Planner Agent              │
              │ (main branch)         │
              │ produces spec         │
              ▼                       │
         ┌─────────────────────┐      │
         │ Worktree A          │      │
         │ Executor Agent 1    │      │
         │ (前端代码修改)       │      │
         └─────────────────────┘      │
         ┌─────────────────────┐      │
         │ Worktree B          │      │
         │ Executor Agent 2    │      │
         │ (后端代码修改)       │      │
         └─────────────────────┘      │
         ┌─────────────────────┐      │
         │ Worktree C          │      │
         │ Reviewer Agent 1    │      │
         │ (代码审查)           │      │
         └─────────────────────┘      │
                            │         │
         自动清理 (无变更则删除)        │
                            │         │
         └────────────────────────────┘
```

**隔离策略**:
- 只读 Agent: 无需 Worktree，直接在当前仓库搜索
- 只写 Agent: 创建独立 Worktree，完成后合入或丢弃
- 读写 Agent: Worktree + 变更对比 → 人工确认 → 合入

### 5.5 动态模型路由

```
任务类型          模型选择            Reasoning Effort
────────          ────────            ────────────────
Planner           Opus/Fable          xhigh/max
复杂Executor      Opus/Sonnet         high
简单Executor      Sonnet/Haiku        medium/low
Reviewer          Sonnet              high
Classifier        Haiku/Sonnet        low
Judge             Opus/Sonnet         high
```

**路由策略**:
1. 根据任务复杂度自动选择模型 (基于 Planner 输出的复杂度评分)
2. 根据 Token Budget 动态降级 (预算紧张时降级 Executor)
3. 用户可覆盖 (Pipeline 配置中指定模型)

### 5.6 Harness DSL (Pipeline 定义语言)

```yaml
# .reqmango/agents/pipelines/sprint-review.yml
name: sprint-review
description: "Sprint 结束自动复盘 Pipeline"
trigger:
  type: cron
  schedule: "0 9 * * 5"  # 每周五 9:00
  # 或
  # type: event
  # event: cycle.completed

pipeline:
  stages:
    - name: analyze
      agent: sprint-analyzer
      model: opus
      effort: high
      input:
        cycle_id: "{{trigger.cycle_id}}"
      output: analysis

    - name: review
      agent: sprint-reviewer
      model: sonnet
      effort: high
      mode: adversarial  # 启用3票制对抗性验证
      input:
        analysis: "{{stages.analyze.output}}"
      output: review

    - name: report
      agent: report-generator
      model: sonnet
      effort: medium
      input:
        review: "{{stages.review.output}}"
      output: report
      on_complete:
        - action: create_page
          params:
            title: "Sprint Review {{trigger.cycle_name}}"
            content: "{{stages.report.output}}"
        - action: notify
          params:
            channel: slack
            message: "Sprint review completed"

  retry:
    max_attempts: 3
    backoff: exponential

  budget:
    max_tokens: 100000
    on_exhausted: escalate_to_human
```

### 5.7 Agent Registry 数据流

```
Agent 定义 (YAML/Markdown)
  │
  ▼
AgentRegistry.Register(definition)
  │
  ├→ 校验 (Schema + 权限模型 + 工具声明)
  ├→ 索引 (按能力标签: triage/assign/code/review/report)
  └→ 存储 (PostgreSQL JSONB)
       │
       ▼
AgentDiscovery.Find(capabilities: ["triage", "review"])
  │
  ▼
返回匹配的 Agent 列表 (按评分排序)
```

### 5.8 错误处理策略

| 场景 | 策略 |
|------|------|
| Planner 失败 | Pipeline 终止，通知用户 |
| Executor 失败 | 自动重试 (max 3次)，失败则跳过该分支 |
| Reviewer 全部否决 | 反馈 Executor 重试，超限则降级为人工 |
| Token Budget 耗尽 | 优雅降级 — 用更便宜模型完成或暂停等人工 |
| Worktree 冲突 | 等待锁释放 (timeout 60s)，超时则新建 worktree |
| LLM API 错误 | 指数退避重试 (1s→2s→4s→8s)，超限终止 |

---

## 6. Loop Engine 详细设计

### 6.1 核心概念

```
Loop = Act → Observe → Reason → Repeat 自主闭环

区别:
  Harness = 有始有终的工作流 (一次性Pipeline)
  Loop    = 持续运行的闭环 (直到目标达成或条件终止)
```

### 6.2 Loop 状态机

```
                    ┌──────────┐
                    │  IDLE    │
                    └────┬─────┘
                         │ Trigger 激活
                    ┌────▼─────┐
                    │ PLANNING │ ← 分析目标，制定初始策略
                    └────┬─────┘
                         │
                    ┌────▼─────┐
              ┌─────│  ACTING  │←──────────────┐
              │     └────┬─────┘               │
              │          │ 执行动作              │
              │     ┌────▼─────┐               │
              │     │OBSERVING │ 收集结果/反馈   │
              │     └────┬─────┘               │
              │          │                     │
              │     ┌────▼─────┐               │
              │     │REASONING │ 判断是否达标   │
              │     └────┬─────┘               │
              │          │                     │
              │   ┌──────┼──────┐              │
              │   │      │      │              │
              │ 达标  未达标  需等待            │
              │   │      │      │              │
              │   │      └──────┼──────────────┘
              │   │             │ 调整策略，重新 ACTING
              │   │        ┌────▼─────┐
              │   │        │ WAITING  │  等待外部事件
              │   │        └────┬─────┘
              │   │             │ 事件到达
              │   │             └──────────────┐
              │   │                            │
         ┌────▼───▼──┐                    ┌────▼─────┐
         │ COMPLETED │                    │  FAILED  │
         └───────────┘                    └──────────┘
```

### 6.3 Loop 执行模型

#### 6.3.1 Goal-based Loop (目标驱动)

```go
// 伪代码
type GoalLoop struct {
    Goal        string   // "Sprint 进度偏差 < 10%"
    Metrics     []Metric // 当前进度: 偏差15%
    MaxIterations int    // 最大迭代次数: 10
    Budget      Budget   // Token 预算: 50000
}

func (l *GoalLoop) Run() {
    for i := 0; i < l.MaxIterations; i++ {
        // Act: 分析+采取行动
        action := l.PlanAction()
        result := l.ExecuteAction(action)

        // Observe: 收集指标
        metrics := l.CollectMetrics()

        // Reason: 判断目标
        if l.GoalAchieved(metrics) {
            l.Status = COMPLETED
            return
        }

        // 未达标→调整策略继续
        if l.BudgetExhausted() {
            l.Escalate("预算耗尽，需要人工介入")
            return
        }
    }
}
```

#### 6.3.2 Loop-until-dry 模式 (穷举发现)

用于未知数量的发现任务 (找 Bug / 去重检测 / 风险识别):

```
round = 0
dry_rounds = 0
seen = Set()

while dry_rounds < 2:
    found = Agent.FindIssues()  # 本轮发现
    fresh = found - seen        # 去重
    if fresh.isEmpty():
        dry_rounds += 1
        continue
    dry_rounds = 0
    seen.addAll(fresh)
    verified = AdversarialVerify(fresh)  # 对抗性验证
    confirmed.addAll(verified)
```

#### 6.3.3 Scheduled Loop (定时巡检)

```yaml
loop:
  name: daily-triage
  trigger:
    type: cron
    schedule: "0 8 * * 1-5"  # 工作日早8点
  pipeline:
    - scan_new_items:
        query: "created > -24h"
    - classify_and_assign:
        mode: auto_triage
    - detect_duplicates:
        against: backlog
    - flag_urgent:
        rules: [priority=urgent, mention=@lead]
    - generate_digest:
        format: markdown
        action: post_to_slack
```

### 6.4 Trigger 系统

| Trigger 类型 | 描述 | 示例 |
|-------------|------|------|
| Event | 项目事件触发 | issue.created, state.changed, comment.added |
| Cron | 定时触发 | "0 9 * * 1" (每周一早9点) |
| Webhook | 外部Webhook触发 | GitLab MR opened, Sentry error spike |
| Manual | 手动触发 | 用户在UI点击"运行Agent Loop" |
| Chained | 另一个Loop完成触发 | Loop A 完成 → 自动启动 Loop B |

### 6.5 Budget 控制系统

```go
type BudgetController struct {
    MaxTokens      int       // 总Token预算
    UsedTokens     int       // 已消耗
    MaxCost        float64   // 总费用预算 (USD)
    UsedCost       float64   // 已花费
    MaxIterations  int       // 最大迭代次数
    Iteration      int       // 当前迭代
    MaxDuration    Duration  // 最大运行时间
    StartTime      Time
}

func (b *BudgetController) CanContinue() bool {
    if b.MaxTokens > 0 && b.UsedTokens >= b.MaxTokens {
        return false  // Token预算耗尽
    }
    if b.MaxCost > 0 && b.UsedCost >= b.MaxCost {
        return false  // 费用预算耗尽
    }
    if b.MaxIterations > 0 && b.Iteration >= b.MaxIterations {
        return false  // 迭代次数达上限
    }
    if b.MaxDuration > 0 && time.Since(b.StartTime) >= b.MaxDuration {
        return false  // 超时
    }
    return true
}
```

### 6.6 State/Memory 持久化

Loop 的上下文需要在迭代之间持久化 (跨越 LLM 上下文窗口重置):

```
┌────────────────────────────────────────┐
│            Loop State Store            │
├────────────────────────────────────────┤
│                                        │
│  Loop ID: loop_abc123                  │
│  Status: ACTING                        │
│  Goal: "Sprint偏差 < 10%"              │
│  Current Iteration: 3/10               │
│  Tokens Used: 12,450/50,000            │
│  History:                              │
│    Iter 1: Action=X, Result=Y          │
│    Iter 2: Action=Z, Result=W          │
│  Working Memory:                       │
│    {                                   │
│      "bottleneck_issue": "PROJ-42",    │
│      "blocked_assignee": "alice",      │
│      "strategy": "reallocate_tasks"    │
│    }                                   │
│                                        │
└────────────────────────────────────────┘
```

**存储方案**:
- PostgreSQL `agent_loop_runs` 表 (JSONB working_memory)
- 每次迭代后 update (不保存完整LLM上下文，只保存结构化摘要)
- 新迭代开始时注入 "Previous Loop State" 到 Agent 上下文

### 6.7 停止条件

```
停止类型:
  1. Goal Achieved  — 目标指标达标
  2. Budget Exhausted — Token/费用/时间/迭代次数耗尽
  3. No Progress    — 连续N次迭代无改善 (stuck detection)
  4. Escalated      — 外部事件 (人工干预/优先级变更)
  5. Error          — 不可恢复错误 (LLM API 持续失败)
```

**Stuck Detection**:
```
if 连续3次迭代后 metrics 无改善:
    → 自动升级人工: "Loop stuck — 过去3次迭代未观察到进展。
       当前状态: {...}。建议: 调整目标或提供额外上下文。"
```

---

## 7. Phase 路线图 & 里程碑

### 7.1 Phase 1: Agent Loop MVP (M1-M3)

**目标**: 追平 Jira Agent 的核心交互模式 + 首个自主Loop展示差异化

| 模块 | 交付物 | 优先级 | 依赖 |
|------|--------|--------|------|
| **Agent 自动化触发** | FR-01 完成: `dispatch_agent` 动作注册到 AutomationRule | P0 | 无 |
| **基础 Loop 引擎** | Act→Observe→Reason→Repeat 核心状态机 | P0 | Agent自动化 |
| **Budget 控制** | Token/Cost/Iteration 预算控制 | P0 | Loop引擎 |
| **Agent Session 面板** | Agent 运行历史、状态、日志可视 | P0 | 无 |
| **Sprint Agent Loop (旗舰场景)** | Agent 自动分析 Sprint 进度→识别阻塞→建议调整→执行→复检 | P0 | Loop引擎 |
| **Triage Loop** | 新工单自动去重→分类→分配→紧急标记 | P1 | Loop引擎 |
| **AI Copilot 面板** | FR-02: 统一AI入口 (Ask/Build/Create/Agent) | P1 | 无 |
| **Copilot 操作化** | FR-03: AI 结果→一键创建工作项/添加到仪表盘 | P1 | Copilot面板 |

**里程碑**: 
- M1末: Loop 引擎核心可用，能运行一个简单的 Goal-based Loop
- M2末: Sprint Agent Loop 端到端可用，内部 dogfooding
- M3末: Phase 1 发布，Sprint Loop + Triage Loop 两个场景上线

### 7.2 Phase 2: Harness Engine (M4-M6)

**目标**: 多Agent编排引擎上线，对抗性验证建立质量壁垒

| 模块 | 交付物 | 优先级 |
|------|--------|--------|
| **Harness 编排引擎** | Pipeline/Parallel/Fan-out/Tournament 四种编排模式 | P0 |
| **对抗性验证** | 3票制 Reviewer 协议 + Refute-first Prompt 模板 | P0 |
| **Worktree 隔离** | Git Worktree Pool + 并行Agent执行环境 | P0 |
| **动态模型路由** | 任务复杂度→模型+Effort自动匹配 | P1 |
| **Harness DSL** | YAML Pipeline 定义语言 + 解析器 + 校验器 | P0 |
| **Agent Registry** | Agent 注册/发现/权限/版本管理 | P1 |
| **Pipeline 可视化** | 前端 DAG 流程图展示 Pipeline 执行状态 | P1 |
| **Sprint + Harness 联动** | Sprint Agent Loop 升级为 Harness Pipeline | P1 |

**里程碑**:
- M4末: Harness 引擎 alpha，支持基础 P→E→R Pipeline
- M5末: 对抗性验证上线，Worktree 并行隔离可用
- M6末: Phase 2 发布，Pipeline 模板库首批 5+ 模板

### 7.3 Phase 3: Agent Studio (M7-M9)

**目标**: Agent 可视化编排 + 自主Loop无人值守

| 模块 | 交付物 | 优先级 |
|------|--------|--------|
| **Agent Studio 前端** | 可视化 Pipeline Builder (拖拽式) | P0 |
| **Pipeline 模板市场** | 社区贡献+官方模板 (Sprint/Triage/Review/Report) | P0 |
| **Agent 调试器** | Step-through 执行，每步查看中间结果 | P1 |
| **Agent SDK (ADK)** | 自定义Agent开发工具包 + 文档 | P0 |
| **Autonomous Loop** | 无人值守模式 (Cron/Webhook自动触发+执行+报告) | P0 |
| **跨工作空间 Loop** | Agent在工作空间间协调 (多项目管理) | P1 |
| **Loop 性能分析** | 每次Loop执行的效率报告 + 优化建议 | P1 |
| **学习型 Loop** | 从历史Loop中学习，优化后续执行策略 | P2 |

**里程碑**:
- M7末: Agent Studio 前端可用，Pipeline Builder MVP
- M8末: ADK v1 发布，模板市场上线
- M9末: Phase 3 发布，Autonomous Loop 生产可用

### 7.4 Phase 4: Agent OS (M10-M12)

**目标**: 平台化 — Agent 成为 Reqmango 的一等公民

| 模块 | 交付物 | 优先级 |
|------|--------|--------|
| **Agent 市场** | 社区Agent/Pipeline发现+安装+评价 | P0 |
| **双向 MCP Hub** | Reqmango 作为 MCP Client 调用外部 Agent | P0 |
| **外部Agent集成** | Claude/Cursor/Codex Agent 接入 Reqmango | P1 |
| **Agent Graph** | 项目管理知识图谱 (开源版 Teamwork Graph) | P0 |
| **预测型 Loop** | Agent 主动预测问题 (基于历史+图谱) | P1 |
| **多Agent协作Loop** | 多个 Harness Pipeline 协同工作 | P1 |
| **企业治理** | SSO/SAML + 完整审计链 + 合规报告 | P1 |

**里程碑**:
- M10末: Agent 市场 + MCP Hub 上线
- M11末: Agent Graph v1, 预测型 Loop beta
- M12末: Phase 4 发布，全平台 GA

---

## 8. 杀手场景设计: Sprint Agent Loop

### 8.1 为什么选 Sprint Loop

1. **复用现有能力**: AI Sprint Planning (容量建议+风险分析) 已实现
2. **复用触发器**: AutomationRule 5种触发器已可用
3. **差异化最强**: Jira/Plane 都没有自主Sprint管理Agent
4. **价值最直观**: 从"人工盯Sprint"到"Agent自主盯Sprint"
5. **Loop 最自然**: Sprint 的自然周期 (每日→检查→调整) 就是 Loop

### 8.2 用户故事

> **作为** Sprint Master
> **我想要** Agent 在 Sprint 期间自动检查进度、识别阻塞、调整分配
> **以便** Sprint 偏差 < 10%，且我不需要每天手动盯看板

### 8.3 Loop 设计

```
Trigger: 每个工作日 9:00 自动触发 (Cron) + Sprint 状态变更时触发

Loop 执行流程:
  ┌─────────────────────────────────────────────┐
  │ ACTING:                                      │
  │  Agent 获取 Sprint 数据:                      │
  │  - 当前进度 (燃尽图)                          │
  │  - 各Issue状态/指派人                         │
  │  - 阻塞项 (长时间无进展的Issue)               │
  │  - 成员工作负载                               │
  └───────────────────┬─────────────────────────┘
                      │
  ┌───────────────────▼─────────────────────────┐
  │ OBSERVING:                                   │
  │  计算关键指标:                                │
  │  - 进度偏差 = 预期进度% - 实际进度%            │
  │  - 负载不均 = max(负载) - min(负载)           │
  │  - 阻塞率 = 阻塞Issue数 / 总Issue数           │
  └───────────────────┬─────────────────────────┘
                      │
  ┌───────────────────▼─────────────────────────┐
  │ REASONING:                                   │
  │  IF 进度偏差 < 10% AND 负载不均 < 30%:        │
  │    → COMPLETED (本次检查达标)                  │
  │  IF 进度偏差 >= 10%:                          │
  │    → 识别top 3 阻塞原因                       │
  │    → 建议调整方案 (重新分配/调整范围/延期)      │
  │    → 如果有高风险项 → 标记+通知                │
  │  IF 连续3次检查偏差恶化:                       │
  │    → 升级人工 + 生成详细分析报告               │
  └───────────────────┬─────────────────────────┘
                      │
  ┌───────────────────▼─────────────────────────┐
  │ REPEAT:                                      │
  │  每天自动运行直到:                             │
  │  - Sprint 完成 (所有Issue done)               │
  │  - Sprint 到期                                │
  │  - 人工终止                                   │
  │                                              │
  │  每个Sprint周期结束时:                         │
  │  → 自动生成 Sprint Review 报告                │
  │  → 沉淀 Lessons Learned 到 Loop Memory        │
  └─────────────────────────────────────────────┘
```

### 8.4 Agent Action 清单

| Action | 触发条件 | 执行 |
|--------|---------|------|
| 进度检查 | 每日定时 | 计算燃尽图偏差 |
| 阻塞检测 | 进度检查时 | 识别 >3天无进展的Issue |
| 负载分析 | 进度检查时 | 统计每人负载，标记过载/空闲 |
| 调整建议 | 偏差 >10% | 生成调整方案 (重新分配/调整范围) |
| 自动调整 | 偏差 >10% 且置信度 >80% | 自动执行调整 (如：将未开始Issue移到下个Sprint) |
| 风险通知 | 检测到高风险 | @Sprint Master + Slack 通知 |
| 日报生成 | 每次检查后 | 生成Sprint健康摘要 |
| 复盘报告 | Sprint完成时 | 生成完整复盘报告 (完成率/阻塞分析/改进建议) |

### 8.5 Sprint Loop 配置示例

```yaml
# .reqmango/loops/sprint-guardian.yml
name: sprint-guardian
description: "Sprint 自主守护Agent — 每日检查、自动调整、风险预警"
version: "1.0"

trigger:
  - type: cron
    schedule: "0 9 * * 1-5"  # 工作日早9点
  - type: event
    event: cycle.state_changed

loop:
  type: goal_based
  goal: "Sprint进度偏差 < 10% AND 无人过载(>150%容量)"
  max_iterations_per_check: 3
  max_daily_tokens: 30000

agent:
  model:
    planner: opus
    executor: sonnet
    effort: high

actions:
  - analyze_progress
  - detect_blockers
  - check_workload
  - suggest_rebalance
  - auto_rebalance (requires: confidence >= 0.8)
  - notify_stakeholders
  - generate_daily_digest
  - generate_sprint_review

notifications:
  - channel: slack
    on: [blocker_detected, risk_escalated, sprint_completed]
  - channel: in_app
    on: [daily_digest, auto_rebalance_executed]

budget:
  max_tokens_per_day: 50000
  max_cost_per_sprint: 2.00  # USD
  on_budget_critical: notify_admin
```

---

## 9. API & 集成设计

### 9.1 新增 REST API

```
POST   /api/v1/workspaces/{ws}/agents/pipelines          # 创建Pipeline
GET    /api/v1/workspaces/{ws}/agents/pipelines          # 列出Pipelines
GET    /api/v1/workspaces/{ws}/agents/pipelines/{id}     # 获取Pipeline详情
PUT    /api/v1/workspaces/{ws}/agents/pipelines/{id}     # 更新Pipeline
DELETE /api/v1/workspaces/{ws}/agents/pipelines/{id}     # 删除Pipeline
POST   /api/v1/workspaces/{ws}/agents/pipelines/{id}/run # 手动触发Pipeline

POST   /api/v1/workspaces/{ws}/agents/loops               # 创建Loop
GET    /api/v1/workspaces/{ws}/agents/loops               # 列出Loops
GET    /api/v1/workspaces/{ws}/agents/loops/{id}          # 获取Loop详情+执行历史
POST   /api/v1/workspaces/{ws}/agents/loops/{id}/start    # 启动Loop
POST   /api/v1/workspaces/{ws}/agents/loops/{id}/stop     # 停止Loop
GET    /api/v1/workspaces/{ws}/agents/loops/{id}/runs     # Loop运行历史

GET    /api/v1/workspaces/{ws}/agents/registry             # 列出已注册Agent
POST   /api/v1/workspaces/{ws}/agents/registry             # 注册新Agent
GET    /api/v1/workspaces/{ws}/agents/sessions             # Agent执行Session列表
GET    /api/v1/workspaces/{ws}/agents/sessions/{id}        # Session详情 (含回放)

POST   /api/v1/workspaces/{ws}/agents/marketplace/search   # 搜索模板市场
POST   /api/v1/workspaces/{ws}/agents/marketplace/install  # 安装模板
```

### 9.2 MCP Server 扩展 (Phase 1)

新增 Tool 定义:

```json
{
  "tools": [
    {
      "name": "dispatch_agent_pipeline",
      "description": "触发一个Agent Pipeline执行",
      "parameters": {
        "pipeline_id": "string",
        "context": "object"
      }
    },
    {
      "name": "start_agent_loop",
      "description": "启动一个Agent Loop",
      "parameters": {
        "loop_id": "string",
        "goal_override": "string (optional)"
      }
    },
    {
      "name": "get_agent_session",
      "description": "获取Agent会话状态和结果",
      "parameters": {
        "session_id": "string"
      }
    },
    {
      "name": "query_agent_graph",
      "description": "查询Agent知识图谱",
      "parameters": {
        "query": "string (RQL or natural language)",
        "entity_type": "issue|cycle|module|member"
      }
    }
  ]
}
```

### 9.3 Webhook 扩展

```
新增事件:
  agent.pipeline.started
  agent.pipeline.stage.completed
  agent.pipeline.completed
  agent.pipeline.failed
  agent.loop.iteration
  agent.loop.completed
  agent.loop.escalated
  agent.review.failed (对抗性验证未通过)
```

---

## 10. 数据模型

### 10.1 新增核心表

```sql
-- Pipeline 定义表
CREATE TABLE agent_pipelines (
    id              SERIAL PRIMARY KEY,
    workspace_id    INTEGER NOT NULL REFERENCES workspaces(id),
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    pipeline_def    JSONB NOT NULL,        -- Pipeline DSL (YAML→JSON)
    version         VARCHAR(50) DEFAULT '1.0',
    status          VARCHAR(50) DEFAULT 'active', -- active/draft/archived
    created_by      INTEGER REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Pipeline 执行记录
CREATE TABLE agent_pipeline_runs (
    id              SERIAL PRIMARY KEY,
    pipeline_id     INTEGER NOT NULL REFERENCES agent_pipelines(id),
    trigger_type    VARCHAR(50),            -- manual/event/cron/webhook
    trigger_context JSONB,                  -- 触发时的上下文数据
    status          VARCHAR(50) DEFAULT 'pending', -- pending/running/completed/failed
    stages_result   JSONB,                  -- 每个Stage的执行结果
    tokens_used     INTEGER DEFAULT 0,
    cost_usd        DECIMAL(10,4) DEFAULT 0,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Loop 定义表
CREATE TABLE agent_loops (
    id              SERIAL PRIMARY KEY,
    workspace_id    INTEGER NOT NULL REFERENCES workspaces(id),
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    loop_def        JSONB NOT NULL,         -- Loop DSL (YAML→JSON)
    version         VARCHAR(50) DEFAULT '1.0',
    status          VARCHAR(50) DEFAULT 'active',
    created_by      INTEGER REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Loop 运行状态 (持久化跨迭代Memory)
CREATE TABLE agent_loop_runs (
    id              SERIAL PRIMARY KEY,
    loop_id         INTEGER NOT NULL REFERENCES agent_loops(id),
    status          VARCHAR(50) DEFAULT 'running', -- running/completed/failed/escalated
    current_iteration INTEGER DEFAULT 0,
    max_iterations  INTEGER DEFAULT 100,
    goal            TEXT NOT NULL,
    goal_metrics    JSONB,                  -- 目标指标当前值
    working_memory  JSONB DEFAULT '{}',     -- Loop 工作记忆 (跨迭代持久化)
    tokens_used     INTEGER DEFAULT 0,
    cost_usd        DECIMAL(10,4) DEFAULT 0,
    stopped_reason  VARCHAR(100),           -- goal_achieved/budget_exhausted/stuck/escalated/manual
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Loop 迭代记录
CREATE TABLE agent_loop_iterations (
    id              SERIAL PRIMARY KEY,
    loop_run_id     INTEGER NOT NULL REFERENCES agent_loop_runs(id),
    iteration_num   INTEGER NOT NULL,
    action_taken    JSONB NOT NULL,         -- 执行的动作描述+参数
    result_observed JSONB NOT NULL,         -- 观察到的结果+指标
    reasoning       TEXT,                   -- AI 推理过程
    decision        VARCHAR(50) NOT NULL,   -- continue/stop/escalate
    tokens_used     INTEGER DEFAULT 0,
    duration_ms     INTEGER,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Agent Session (统一的可观测性记录)
CREATE TABLE agent_sessions (
    id              VARCHAR(64) PRIMARY KEY,  -- UUID
    workspace_id    INTEGER NOT NULL REFERENCES workspaces(id),
    agent_type      VARCHAR(50) NOT NULL,     -- pipeline_stage/loop_iteration/standalone
    agent_ref       VARCHAR(255),             -- pipeline_id:stage 或 loop_id
    status          VARCHAR(50) DEFAULT 'running',
    model_used      VARCHAR(100),
    input_summary   TEXT,                     -- 输入摘要 (不存完整上下文)
    output_summary  TEXT,                     -- 输出摘要
    tokens_input    INTEGER DEFAULT 0,
    tokens_output   INTEGER DEFAULT 0,
    cost_usd        DECIMAL(10,4) DEFAULT 0,
    tools_called    JSONB DEFAULT '[]',       -- [{tool_name, count}]
    error_message   TEXT,
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    metadata        JSONB DEFAULT '{}'
);

-- Agent Registry
CREATE TABLE agent_registry (
    id              SERIAL PRIMARY KEY,
    workspace_id    INTEGER REFERENCES workspaces(id), -- NULL = 全局Agent
    name            VARCHAR(255) NOT NULL UNIQUE,
    display_name    VARCHAR(255),
    description     TEXT,
    capabilities    TEXT[] NOT NULL,          -- {triage, code_review, planning, reporting, ...}
    agent_def       JSONB NOT NULL,           -- Agent定义 (model/prompt/tools/...)
    version         VARCHAR(50) DEFAULT '1.0.0',
    author          VARCHAR(255),
    is_verified     BOOLEAN DEFAULT FALSE,    -- 官方验证?
    installs_count  INTEGER DEFAULT 0,
    rating          DECIMAL(3,2),
    source          VARCHAR(50) DEFAULT 'local', -- local/marketplace
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Agent Graph 节点 (Phase 4)
CREATE TABLE agent_graph_nodes (
    id              SERIAL PRIMARY KEY,
    entity_type     VARCHAR(50) NOT NULL,     -- issue/cycle/module/member/page
    entity_id       INTEGER NOT NULL,
    properties      JSONB DEFAULT '{}',
    embeddings      vector(1536),            -- pgvector 扩展
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Agent Graph 边 (Phase 4)
CREATE TABLE agent_graph_edges (
    id              SERIAL PRIMARY KEY,
    source_type     VARCHAR(50) NOT NULL,
    source_id       INTEGER NOT NULL,
    target_type     VARCHAR(50) NOT NULL,
    target_id       INTEGER NOT NULL,
    relation_type   VARCHAR(100) NOT NULL,    -- blocks/relates/depends_on/assigned_to/...
    weight          DECIMAL(5,4) DEFAULT 1.0,
    properties      JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_pipelines_workspace ON agent_pipelines(workspace_id);
CREATE INDEX idx_pipeline_runs_pipeline ON agent_pipeline_runs(pipeline_id);
CREATE INDEX idx_pipeline_runs_status ON agent_pipeline_runs(status);
CREATE INDEX idx_loops_workspace ON agent_loops(workspace_id);
CREATE INDEX idx_loop_runs_loop ON agent_loop_runs(loop_id);
CREATE INDEX idx_loop_runs_status ON agent_loop_runs(status);
CREATE INDEX idx_loop_iterations_run ON agent_loop_iterations(loop_run_id);
CREATE INDEX idx_agent_sessions_workspace ON agent_sessions(workspace_id);
CREATE INDEX idx_agent_sessions_status ON agent_sessions(status);
CREATE INDEX idx_agent_registry_capabilities ON agent_registry USING GIN(capabilities);
CREATE INDEX idx_agent_registry_workspace ON agent_registry(workspace_id);
```

### 10.2 现有表扩展

```sql
-- 扩展 Issue 表
ALTER TABLE issues ADD COLUMN agent_session_id VARCHAR(64);
ALTER TABLE issues ADD COLUMN agent_assignee_id INTEGER REFERENCES agent_registry(id);

-- 扩展 AutomationRule 的动作支持
-- (FR-01 已规划: dispatch_agent 动作类型)
-- 无需 DDL 变更 — AutomationRule.actions JSONB 已支持扩展

-- 扩展 IssueActivity 表
-- 新增 activity_type: 'agent_dispatched', 'agent_completed', 'loop_iteration'
-- issue_activities.activity_type 是 VARCHAR, 无需 DDL
```

---

## 11. 安全与治理

### 11.1 Agent 权限模型

```
Agent 权限 = min(Agent 配置的权限, 触发用户的权限)

示例:
  - 用户 Alice (Admin) 触发 Agent → Agent 拥有 Admin 权限
  - 用户 Bob (Member) 触发 Agent → Agent 拥有 Member 权限
  - Cron 触发 → Agent 使用配置的 Service Account 权限
```

### 11.2 审计链

```
每次 Agent 动作记录:
  - 谁触发 (用户/系统)
  - 哪个 Agent
  - 什么动作 (创建/修改/删除/读取)
  - 什么对象 (Issue/Cycle/Page)
  - 什么结果 (成功/失败/人工确认)
  - Token 消耗
  - 时间戳

查询接口:
  GET /api/v1/workspaces/{ws}/agents/audit?date_range=...
```

### 11.3 安全约束

| 约束 | 描述 |
|------|------|
| Agent 不能提升权限 | Agent 无法执行超出触发用户权限的操作 |
| Agent 不能修改自身定义 | Pipeline/Loop 配置只能通过 API 修改 |
| Agent 不能旁路审批流 | 受审批工作流约束 |
| Loop 必须有预算上限 | 创建 Loop 时必须设置 max_tokens 或 max_cost |
| 敏感操作需人工确认 | 删除/归档/发布 操作默认需要用户确认 |
| API Key 加密存储 | BYO Keys 使用 AES-256-GCM 加密存储 |

---

## 12. 实施优先级

### 12.1 Phase 1 任务分解

```
优先级 P0 (必须):
  □ 1.1 Agent 自动化触发器 (FR-01 完成)
       - ActionExecutor 注册 dispatch_agent handler
       - AutomationRule 表单增加 dispatch_agent 选项
       - AgentService 依赖注入到 AutomationService
       - 测试: 集成测试 EventBus→Agent 调度链路

  □ 1.2 基础 Loop 引擎
       - LoopStateMachine: IDLE→PLANNING→ACTING→OBSERVING→REASONING
       - LoopRunner: 执行 Loop 迭代的主循环
       - BudgetController: Token/Cost/Iteration 控制
       - 数据库: agent_loops, agent_loop_runs, agent_loop_iterations 表

  □ 1.3 Sprint Agent Loop (旗舰场景)
       - SprintAnalyzer Agent: 获取 Sprint 数据+计算指标
       - BlockerDetector Agent: 识别阻塞项
       - WorkloadAnalyzer Agent: 负载分析
       - SprintGuardian Loop 配置 + 测试

  □ 1.4 Agent Session 可观测面板
       - 前端: Agent Session 列表 + 详情页
       - 后端: agent_sessions 表 + CRUD API
       - 实时状态轮询 (SSE)

优先级 P1 (应该):
  □ 1.5 Triage Loop
       - IntakeTriage Agent: 去重+分类+分配
       - 复用现有 AITriage 能力

  □ 1.6 AI Copilot 统一面板 (FR-02)
       - AICopilot.vue: Tab切换 (Ask/Build/Create/Agent)
       - 整合现有 AIChatSidebar + AICreateDialog

  □ 1.7 AI 结果操作化 (FR-03)
       - AIResultActions.vue: 操作按钮组
       - AIQuickCreateDialog.vue: 预填充+确认

优先级 P2 (可以):
  □ 1.8 MCP Server 扩展 (dispatch_agent_pipeline / start_agent_loop 工具)
  □ 1.9 Loop 配置 UI (前端 YAML 编辑器+预设模板)
```

### 12.2 后端模块组织

```
backend/internal/
├── agent/                    # 新: Agent Platform 核心
│   ├── harness/              # Harness 编排引擎
│   │   ├── pipeline.go       # Pipeline 定义+执行器
│   │   ├── planner.go        # Planner Agent 角色
│   │   ├── executor.go       # Executor Agent 角色
│   │   ├── reviewer.go       # Reviewer Agent (对抗性验证)
│   │   ├── router.go         # 模型路由
│   │   └── worktree.go       # Worktree 隔离管理
│   ├── loop/                 # Loop 引擎
│   │   ├── state_machine.go  # Loop 状态机
│   │   ├── runner.go         # Loop 执行器
│   │   ├── trigger.go        # Trigger 系统 (Event/Cron/Webhook)
│   │   ├── budget.go         # Budget 控制器
│   │   └── observer.go       # 指标收集+目标判断
│   ├── registry/             # Agent 注册中心
│   │   ├── registry.go       # 注册/发现/版本管理
│   │   └── capabilities.go   # 能力标签匹配引擎
│   ├── session/              # Session 可观测
│   │   ├── session.go        # Session 记录+回放
│   │   └── telemetry.go      # 遥测数据收集
│   └── model/                # 数据模型
│       ├── pipeline.go       # Pipeline/Run 模型
│       ├── loop.go           # Loop/Run/Iteration 模型
│       ├── session.go        # Session 模型
│       └── registry.go       # Registry 模型
├── handler/
│   ├── agent_pipeline.go     # Pipeline API Handlers
│   ├── agent_loop.go         # Loop API Handlers
│   ├── agent_registry.go     # Registry API Handlers
│   └── agent_session.go      # Session API Handlers
└── service/
    └── agent_service.go      # 扩展: 编排入口服务
```

### 12.3 前端模块组织

```
frontend/src/
├── views/
│   └── agents/               # 新: Agent 管理页面
│       ├── AgentDashboard.vue     # Agent 仪表盘
│       ├── PipelineBuilder.vue    # 可视化Pipeline编辑
│       ├── PipelineRunDetail.vue  # Pipeline 运行详情
│       ├── LoopBuilder.vue        # Loop 配置编辑
│       ├── LoopRunDetail.vue      # Loop 运行详情+历史
│       └── AgentRegistry.vue      # Agent 注册中心
├── components/
│   └── agents/                # 新: Agent 组件
│       ├── AICopilot.vue          # FR-02: 统一AI面板
│       ├── AgentSessionCard.vue   # Session 卡片
│       ├── AgentSessionTimeline.vue # Session 时间线
│       ├── PipelineDAG.vue        # Pipeline DAG 可视化
│       ├── LoopStateView.vue      # Loop 状态展示
│       ├── AIResultActions.vue    # FR-03: 结果操作化
│       └── AgentSelector.vue      # Agent 选择器
├── api/
│   ├── agent-pipeline.ts     # Pipeline API
│   ├── agent-loop.ts         # Loop API
│   ├── agent-registry.ts     # Registry API
│   └── agent-session.ts      # Session API
└── stores/
    ├── agentPipeline.ts      # Pipeline Store
    ├── agentLoop.ts          # Loop Store
    └── agentSession.ts       # Session Store
```

---

## 13. 风险评估与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| LLM API 不稳定导致 Loop 中断 | 中 | 高 | 指数退避重试 + 降级为人工 |
| Token 成本失控 | 中 | 高 | Budget 硬限制 + 动态模型降级 + 卡死检测 |
| Agent 产生错误结果无人发现 | 中 | 高 | 对抗性验证 (3票制) + 关键操作人工确认 |
| Worktree 并发冲突 | 低 | 中 | 锁机制 + 超时处理 + 自动重试 |
| Loop 陷入死循环 | 低 | 高 | Stuck Detection + 迭代上限 + Budget 上限 |
| Agent 代码与现有系统耦合过重 | 中 | 中 | Agent Platform 作为独立模块，通过接口桥接 |
| 社区采用度低 | 中 | 高 | Phase 1 先做出杀手场景再推广 SDK |
| Jira 快速跟进多Agent编排 | 高 | 中 | 开源 + 自托管壁垒 + 持续差异化 Loop 深度 |
| 人才/技能缺口 | 中 | 中 | Phase 1 从现有 Go 后端团队可完成的范围开始 |

---

## 14. 成功指标

### 14.1 产品指标

| 指标 | Phase 1 目标 | Phase 2 目标 | Phase 4 目标 |
|------|------------|------------|------------|
| Agent Loop 周活用户 | 50 | 200 | 500+ |
| Sprint Agent Loop 采纳率 | 15% | 35% | 50%+ |
| 对抗性验证拦截率 | — | >30% | >50% |
| Agent Pipeline 模板数 | 3 | 15 | 50+ |
| MCP 工具调用/日 | 1000 | 10000 | 50000+ |
| 社区 Agent 贡献数 | — | 5 | 30+ |

### 14.2 技术指标

| 指标 | 目标 |
|------|------|
| Pipeline 执行延迟 (不含LLM) | < 500ms |
| Loop 迭代间隔 | < 3s 启动时间 |
| Agent Session 记录写入 | < 100ms |
| Worktree 创建/清理 | < 2s |
| Loop 可用性 (Cron触发) | 99.5% |
| Token Budget 控制精度 | 误差 < 5% |

### 14.3 北极星指标

> **Phase 4 末: Agent 自主完成的工作占所有工作项更新的 20%+**
> (即: 每5次Issue更新中有1次是由Agent自主发起的)

---

## 附录

### A. 术语表

| 术语 | 定义 |
|------|------|
| Harness | 多Agent编排框架 — Planner→Executor→Reviewer 流水线 |
| Loop | 自主闭环 — Act→Observe→Reason→Repeat 持续执行 |
| Pipeline | 一次性有始有终的Agent工作流 |
| Adversarial Verification | 对抗性验证 — 独立Review Agent评审Executor输出 |
| Worktree | Git Worktree隔离 — 并行Agent在独立分支执行 |
| Agent Registry | Agent注册中心 — 注册/发现/权限/版本管理 |
| Agent Graph | 项目管理知识图谱 — 实体关系网络 |
| RQL | Reqmango Query Language — 自研查询语言 |
| BYO Keys | Bring Your Own Keys — 用户自带LLM API Key |
| MCP | Model Context Protocol — AI工具互联协议 |

### B. 参考文档

- [Reqmango Competitive Analysis](docs/reqmango-vs-competitor.md)
- [AI Phase 1 Requirements](docs/specs/ai-phase1-requirements.md)
- [Reqmango PRD](docs/kb/PRD.en.md)
- [Atlassian Agents in Jira](https://www.atlassian.com/blog/rovo/ai-agents-in-jira)
- [Plane AI](https://plane.so/ai)
- [Claude Code Dynamic Workflows](https://claude.com/blog/a-harness-for-every-task-dynamic-workflows-in-claude-code)

---

> **下一步**: 本文档评审通过后，进入 [writing-plans 技能] 生成可执行的 Phase 1 实施计划。
