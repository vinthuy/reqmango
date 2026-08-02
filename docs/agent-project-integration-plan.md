# Agent 与项目管理打通方案

## 一、核心设计思路

**Agent 作为项目的虚拟成员**，参与工作项的分解、执行和协作。

```
Project (项目)
  ├── ProjectMember (项目成员)
  │     ├── Human Member (人类成员)
  │     └── AI Member (Agent 虚拟成员) ← 新增
  ├── Issue (工作项)
  │     ├── Assignee → User (人类指派)
  │     └── AgentAssignee → Agent (AI 指派) ← 新增
  └── Squad (项目团队)
        └── SquadMember → Agent (团队成员)
```

## 二、需要新增的功能

### 2.1 Agent 作为项目虚拟成员
- 新增 `ProjectAgentMember` 模型：关联 Project 和 Agent
- Agent 可以被添加到项目团队中，显示在成员列表中
- Agent 有角色（observer/member/admin），可以参与工作项分配

### 2.2 工作项与 Agent 任务打通
- Issue 增加 `AgentAssigneeID` 字段：直接分配给 Agent
- 工作项详情页显示 Agent 执行状态
- 一键将工作项转为 Agent 任务

### 2.3 工作项分解与 Agent 协作
- 工作项分解时可以选择"AI 辅助分解"
- Agent 根据工作项描述自动生成子任务建议
- 子任务可以分配给其他 Agent 或人类成员

### 2.4 设计方案与 Agent 协作
- 设计方案可以让 Agent 参与评审
- Agent 自动生成设计方案建议
- 设计方案评审结果关联到 Agent 会话

### 2.5 项目团队与 Squad 打通
- 项目可以关联 Squad 作为团队配置
- Squad 执行结果可以更新到项目工作项
- 项目成员可以触发 Squad 执行

## 三、实施计划

### Phase 1: 数据模型打通 (后端)
1. 新增 `ProjectAgentMember` 模型
2. 修改 `Issue` 模型增加 Agent 分配字段
3. 修改 `Squad` 模型增加 ProjectID 关联
4. 新增对应的 Handler 和 Service

### Phase 2: API 打通 (后端)
1. 项目成员管理 API 增加 Agent 成员
2. 工作项 API 增加 Agent 分配和任务创建
3. Squad API 增加项目关联
4. 新增"AI 辅助分解"API

### Phase 3: 前端页面打通
1. 项目成员页面增加 Agent 成员管理
2. 工作项详情页增加 Agent 执行面板
3. 工作项分解增加 AI 辅助选项
4. 项目设置增加 Squad 关联

### Phase 4: 交互流程打通
1. 工作项分配给 Agent 时自动创建 AgentTask
2. AgentTask 完成时自动更新工作项状态
3. Squad 执行结果同步到工作项评论
4. 项目仪表盘显示 Agent 工作统计

## 四、关键 API 设计

### 4.1 项目 Agent 成员管理
```
GET    /api/v1/projects/:projectId/agent-members
POST   /api/v1/projects/:projectId/agent-members
DELETE /api/v1/projects/:projectId/agent-members/:agentId
```

### 4.2 工作项 Agent 分配
```
POST   /api/v1/issues/:issueId/assign-agent
Body: { "agent_id": 123, "task_type": "decompose|execute|review" }
Response: { "agent_task_id": 456 }
```

### 4.3 AI 辅助工作项分解
```
POST   /api/v1/issues/:issueId/ai-decompose
Body: { "depth": 2, "squad_id": null }
Response: { "sub_issues": [...], "suggestion": "..." }
```

### 4.4 项目 Squad 关联
```
POST   /api/v1/projects/:projectId/squads
Body: { "squad_id": 789 }
GET    /api/v1/projects/:projectId/squads
```

## 五、前端页面改动

### 5.1 项目成员页面
- 增加"AI 成员"标签页
- 显示已添加的 Agent 列表
- 支持从 Agent 列表中选择添加

### 5.2 工作项详情页
- 增加"AI 执行"面板
- 显示 Agent 任务状态和进度
- 支持一键分配给 Agent

### 5.3 工作项看板
- 卡片显示 Agent 头像标识
- 支持拖拽分配给 Agent

### 5.4 项目设置
- 增加"AI 团队"配置
- 关联 Squad 和配置 Agent 参数
