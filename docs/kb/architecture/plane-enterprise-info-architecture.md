# Plane Enterprise 信息架构（参考）

> **来源**: `D:\code\makeplane\docs` — Plane 官方 VitePress 文档仓库
> **用途**: 作为 ReqManPy 架构设计的对标参考，非实现规范
> **最后更新**: 2026-06-24

---

## 一、核心层级

```
Workspace（组织/公司）
│
├── Workspace Settings（工作空间管理员统一治理）
│   ├── General                    — 名称、URL、删除
│   ├── Members                    — 成员 + 角色管理
│   ├── Project States (Pro)       — 项目健康度状态（Draft→Planning→Execution→Monitoring→Completed→Cancelled）
│   ├── Projects → Project Labels (Business) — 项目分类标签，跨项目复用
│   ├── Work item Types (Enterprise Grid)    — Type + Properties + Hierarchy，项目按需导入
│   │   ├── Types Tab             — CRUD Type（名称、图标、颜色、描述）
│   │   ├── Properties Tab        — 自定义字段池（多对多绑定 Type）
│   │   └── Hierarchy Tab         — Type 父子嵌套规则（Epic→Story→Task）
│   ├── Automations (Enterprise Grid)        — 跨项目自动化，选择适用范围
│   ├── Plane Runner (Enterprise Grid)       — 脚本引擎（工作流条件 + 自动化动作）
│   ├── Relations (Enterprise Grid)          — 自定义关联类型（inward/outward 命名）
│   ├── Roles & Permissions        — RBAC + 自定义角色 + 权限方案
│   ├── Templates                  — 项目模板、工作项模板、页面模板
│   ├── Integrations               — GitHub/GitLab/Slack 等
│   └── Import/Export              — 数据迁移
│
└── Project Settings（项目管理员独立空间）
    ├── General                    — 名称、标识符、可见性、时区、归档/删除
    ├── Members                    — 项目成员 + 角色
    ├── Features                   — 功能开关（Cycles/Modules/Views/Pages/Intake/Time Tracking）
    ├── States                     — 工作项状态（5 组固定：Backlog/Unstarted/Started/Completed/Cancelled）
    ├── Labels                     — 工作项标签（项目内）
    ├── Estimates                  — 估算系统（Points/Categories/Time，每项目一个）
    ├── Automations                — 默认 + 自定义自动化（Business+）
    ├── Workflows (Business)       — 状态流转 + 审批门（Enterprise Grid 支持 Type-specific）
    ├── Work item Types (Pro)      — 项目级 Type；Enterprise Grid 显示"从工作空间导入"
    ├── Templates                  — 工作项模板
    └── Recurring Work Items       — 重复工作项
```

---

## 二、关键设计：两套 States 和两套 Labels

Plane 有两组不同层级的概念，不可混淆：

| 概念 | 层级 | 作用对象 | 配置位置 |
|------|------|---------|---------|
| **Project States** | 工作空间 | 项目本身（项目健康度阶段） | `Workspace Settings → Project States` |
| **Work Item States** | 项目 | 工作项（任务生命周期） | `Project Settings → States` |
| **Project Labels** | 工作空间 | 项目本身（项目分类标签） | `Workspace Settings → Projects → Project labels` |
| **Work Item Labels** | 项目 | 工作项（任务标签） | `Project Settings → Labels` |

> ReqManPy 的 States 和 Labels 是 Work Item 级别，对应 Plane 的 **Work Item States** 和 **Work Item Labels**。

---

## 三、Work Item States 的 5 个固定分组

来自 [Plane States 文档](D:\code\makeplane\docs\docs\core-concepts\issues\states.md)：

| 分组 | 默认状态 | 语义 |
|------|---------|------|
| **Backlog** | Backlog | 尚未准备好进入开发队列 |
| **Unstarted** | Todo | 已计划但未开始 |
| **Started** | In Progress | 正在执行中 |
| **Completed** | Done | 已成功完成 |
| **Cancelled** | Cancelled | 不再需要或不再适用 |

每个分组可自定义增删状态（如 Started 组内添加 In Review、Testing），但 5 个组本身是固定的。

---

## 四、Work Item Types 的层级设计

### 4.1 按计划版本

| 计划 | Type 归属 | 说明 |
|------|----------|------|
| Free | 无 | 仅内置 Task 类型 |
| Pro / Business | **项目级** | 在 `Project Settings → Work item Types` 中创建管理 |
| Enterprise Grid | **工作空间级** | 在 `Workspace Settings → Work item Types` 中定义，项目按需导入 |

### 4.2 Enterprise Grid 的 Type 架构

```
Workspace Settings → Work item Types
│
├── Types Tab（CRUD）
│   ├── 默认 Task（不可删除）
│   ├── 创建 Type：名称 + 图标 + 颜色 + 描述
│   └── 开关控制：启用/禁用（不影响已有工作项）
│
├── Properties Tab（字段池）
│   ├── 先创建 Property，再附加到 Type（多对多）
│   ├── 8 种类型：Text / Number / Dropdown / Boolean / Date / Member picker / Release picker / URL
│   ├── 可标记 Mandatory / Active
│   └── 一个 Property 可被多个 Type 共享（如 "Severity" 同时用于 Bug 和 Incident）
│
└── Hierarchy Tab（Type 层级）
    ├── 定义 Type 的父子嵌套规则
    ├── Level 0（叶子节点）→ Level 1 → Level 2（顶层）
    ├── 示例：Level 2: Epic → Level 1: Story, Bug → Level 0: Task, Spike
    └── 全局强制执行：无效嵌套在 UI/导入/批量操作中均被阻止
```

### 4.3 项目导入 Type

来自 [Workspace Work Item Types 文档](D:\code\makeplane\docs\docs\work-items\workspace-work-item-types.md)：

- 项目**不自动获得**所有工作空间 Type
- 项目管理员在 `Project Settings → Work item Types → Import from workspace` 中主动选择
- 导入的 Type 与工作空间定义保持链接，工作空间更新自动同步

---

## 五、Workflows 设计

来自 [Workflows 文档](D:\code\makeplane\docs\docs\workflows-and-approvals\workflows.md)：

### 5.1 基本概念

| 概念 | 说明 |
|------|------|
| **Default Workflow** | 每个项目一个，适用于所有 Type |
| **Type-specific Workflow**（Enterprise Grid） | 不同 Type 可有独立工作流（Bug 走审批流，Task 不走） |
| **Transition flow** | 状态 A → 状态 B，可指定 Who（谁能操作）、With（条件脚本） |
| **Approval flow** | 状态 A → 审批 → on approve 到 B / on reject 回退 |
| **Allow new work items** | 每个 State 可设定是否允许直接创建 |

### 5.2 Transition Conditions（Enterprise Grid）

- **Pre-validation**：流转前脚本校验（如"必须有 Assignee"），失败则阻塞流转
- **Post actions**：流转后脚本（如"发 Slack 通知"），不影响流转结果
- 使用 Plane Runner 脚本引擎

### 5.3 Workflows 与 States 的关系

- Workflow 创建后，从项目 States 池中**按需选择**参与该 Workflow 的 States
- 不同 Workflow 可包含不同的 States 子集
- Default Workflow 默认包含所有项目 States

---

## 六、Automations 设计

来自 [Custom Automations 文档](D:\code\makeplane\docs\docs\automations\custom-automations.md)：

### 6.1 两层架构

| 层级 | 配置位置 | 适用范围 | 计划要求 |
|------|---------|---------|---------|
| **Project Automation** | `Project Settings → Automations` | 单个项目 | Business |
| **Workspace Automation** | `Workspace Settings → Automations` | 所有项目或指定子集 | Enterprise Grid |

### 6.2 Trigger-Condition-Action 框架

```
[Trigger] → [Conditions] → [Actions]

Triggers:
  - Work item created / updated / state changed
  - Assignee changed / Comment added
  - Scheduled (Daily/Weekly/Monthly/Cron)

Conditions (AND/OR 逻辑):
  - 过滤字段：State, Priority, Assignees, Labels, Work item type, Created by
  - 运算符：is, is not, in, contains, >, >=, <, <=

Actions:
  - Change property (Priority/State/Assignees/Labels/Start date/Due date)
  - Add comment (支持模板变量 {{priority}})
  - Send webhook (Enterprise Grid，带 HMAC 签名)
  - Run script (Enterprise Grid，Plane Runner)
```

### 6.3 关键行为

- 每次变更事件触发一次检查
- Automation Bot 执行的变更**不会**再次触发自动化（防循环）
- Conditions 为空时对**所有**匹配 Trigger 的 Item 生效
- Scheduled Automation 仅支持 Run Script 动作

---

## 七、Estimates 设计

来自 [Estimates 文档](D:\code\makeplane\docs\docs\core-concepts\issues\estimates.md)：

| 类型 | 说明 | 示例 |
|------|------|------|
| **Points** | 数值型估算 | Linear (1,2,3,4,5)、Fibonacci (1,2,3,5,8,13)、Squares (1,4,9,16)、Custom |
| **Categories** | 文本型估算 | T-shirt sizes (XS,S,M,L,XL)、Easy/Medium/Hard、Custom |
| **Time** | 时长型估算 (Pro) | 1h, 2h, 3h, 4h, 5h30m, 6h30m、Custom |

- 每个项目只能有一个活跃的估算系统
- 在 `Project Settings → Estimates` 中配置

---

## 八、Custom Relations 设计

来自 [Custom Relations 文档](D:\code\makeplane\docs\docs\work-items\custom-relations.md)：

- **工作空间级配置**（Enterprise Grid）
- Plane 内置 3 种默认关系：Blocks、Related、Duplicate
- 自定义关系支持方向性命名：

| Title | Inward name | Outward name |
|-------|------------|--------------|
| Tests | tested by | tests |
| Caused by | caused | caused by |

- 配置位置：`Workspace Settings → Relations`

---

## 九、ReqManPy 对标映射

### 9.1 配置归属对照

| 功能 | Plane Enterprise 位置 | ReqManPy 当前 | ReqManPy 目标 |
|------|----------------------|--------------|--------------|
| Work Item Types | Workspace Settings | Workspace Settings（子组件） | Workspace Settings（主菜单） |
| Custom Fields（Properties） | Workspace Settings → Work item Types → Properties Tab | Workspace Settings | Workspace Settings（整合到 Type 内） |
| Hierarchy | Workspace Settings → Work item Types → Hierarchy Tab | 无 | 后续可选 |
| Type Templates | 不存在（被 Type+Properties 替代） | 独立组件 | **删除** |
| States | Project Settings | Workspace Settings（错位） | **移至 Project Settings** |
| Labels | Project Settings | Workspace Settings（错位） | **移至 Project Settings** |
| Workflows | Project Settings | Workspace Settings + 独立页面 | **移至 Project Settings** |
| Automations | Project + Workspace（分层） | Workspace Settings | **分层到 Project + Workspace** |
| Estimates | Project Settings | 无 | 后续可选 |
| Relations | Workspace Settings | Workspace Settings | ✅ 已对齐 |
| Project Features | Project Settings → Features | 无 | 后续可选 |

### 9.2 ReqManPy 差异化优势（高于 Plane）

| 能力 | ReqManPy | Plane |
|------|---------|-------|
| Workflow Transition 独立存储 | `state_transitions` 表，审批人数据结构化 | Workflow 内联定义 |
| Relation Types 自定义 | ✅ 工作空间级，当前已实现 | Enterprise Grid 才有 |
| Automation 条件 JSON | 自由 JSON 编辑 | 结构化 UI 选择器 |
| 审批流 | Transition + Approval，审批人可指定 | Approval flow，审批人从 Members 选 |

### 9.3 ReqManPy 待补全（Plane 有但我们没有）

| 能力 | 优先级 | 说明 |
|------|--------|------|
| Estimates 估算系统 | P2 | Points / Categories / Time 三种模式 |
| Project Features 功能开关 | P2 | 每个项目可独立开关 Cycles/Modules |
| Type Hierarchy（父子嵌套规则） | P3 | 工作空间级定义，全局强制执行 |
| Plane Runner 脚本引擎 | P3 | 工作流条件 + 自动化动作可编程 |

---

## 十、完整用户旅程

```
1. 创建工作空间
   注册 → 创建 Workspace → 自动成为 Owner
   
2. 工作空间管理员全局治理
   Workspace Settings
   ├── 邀请成员，分配角色
   ├── 定义 Work Item Types（Epic/Feature/Story/Bug/Task/Spike）
   ├── 创建 Properties（Priority/Severity/Version/Story Points…）
   ├── 绑定 Properties 到 Types（多对多勾选）
   ├── 配置 Hierarchy（可选：Epic→Story→Task）
   ├── 定义 Relation Types
   └── 配置跨项目 Automations（可选）

3. 创建项目
   Workspace → Projects → Create
   ├── 输入：名称、标识符、描述、可见性
   ├── 自动创建 5 组默认 States
   └── 创建者成为 Project Admin

4. 项目管理员配置项目
   Project Settings
   ├── Features：开关 Cycles/Modules
   ├── States：自定义每个分组内的状态
   ├── Labels：创建项目标签
   ├── Estimates：选择估算方式
   ├── Work item Types：从工作空间导入需要的 Type
   ├── Workflows：创建 Default Workflow + Type-specific Workflow
   │   ├── 选择参与 States → 定义 Transition/Approval flows
   │   └── 可附加 Pre-validation / Post-action 脚本
   └── Automations：创建项目级自动化规则

5. 日常使用：创建和管理工作项
   创建 Issue
   ├── ① 选择 Type → 决定可用 Properties 集合
   ├── ② 填写 Properties（Mandatory 必填）
   ├── ③ 选择 State（受 Workflow "Allow new work items" 约束）
   ├── ④ 选择 Labels（项目 Label 池）
   ├── ⑤ 设置 Estimate（来自项目 Estimates 配置）
   ├── ⑥ 关联 Cycle/Module（如 Features 启用）
   └── ⑦ 添加 Relations（使用工作空间 Relation Types）
   
   状态流转（受 Workflow 约束）
   ├── 只能按 Workflow 定义的 Transition 路径流转
   ├── Approval 流转需指定 Approver 批准
   └── Pre-validation 脚本校验 → Post-action 脚本自动执行
   
   Automation 自动触发
   └── Trigger 匹配 → Conditions 过滤 → Actions 执行
```
