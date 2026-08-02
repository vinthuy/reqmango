# Agent-Project Integration 架构设计评审

> 日期：2026-07-26
> 评审人员：架构师、产品经理、技术骨干、测试经理、用户代表、竞品分析员
> 评审对象：ARCH-Agent-Project-Integration.md v1.0

---

## 一、评审总结

| 评审角色 | 结论 | 问题数 | 已修复 |
|----------|------|--------|--------|
| 架构师 | ✅ 通过 | 2 | 2 |
| 产品经理 | ✅ 通过 | 3 | 2 |
| 技术骨干 | ✅ 通过 | 4 | 3 |
| 测试经理 | ✅ 通过 | 3 | 2 |
| 用户代表 | ✅ 通过 | 2 | 1 |
| 竞品分析员 | ✅ 通过 | 2 | 1 |

**总体结论：设计评审通过，可进入编码实现阶段。**

---

## 二、各角色评审意见

### 2.1 架构师评审

#### ✅ 通过项
1. **分层架构清晰** - Model → Service → Handler → Router 职责明确
2. **数据模型设计合理** - 扩展现有模型，兼容性好
3. **Workflow 执行引擎设计完善** - 拓扑排序、串行/并行执行、上下文传递
4. **依赖注入模式正确** - 通过接口打破循环依赖

#### ⚠️ 问题与建议

**ARCH-D-01: WorkflowNode 与 AgentTask 的关系需要更清晰** [已修复]
- **问题**：WorkflowNodeRun 如何调用 AgentTask 的流程未明确
- **影响**：开发时可能产生歧义
- **修复**：在设计文档中增加执行关系说明

**ARCH-D-02: ContextPayload 压缩策略未定义** [已修复]
- **问题**：文档提到上下文压缩，但具体策略未定义
- **影响**：可能导致 Token 消耗过大
- **修复**：增加压缩策略说明（摘要替换、字段裁剪、分页加载）

---

### 2.2 产品经理评审

#### ✅ 通过项
1. **功能覆盖完整** - 24项功能点全部覆盖
2. **用户故事清晰** - 每个模块都有明确的用户故事
3. **API 设计规范** - RESTful 风格，路径清晰

#### ⚠️ 问题与建议

**PM-D-01: 缺少 Workflow 预设模板的详细定义** [已修复]
- **问题**：设计文档提到5种预设模板，但模板的具体节点配置未定义
- **影响**：用户无法快速开始使用
- **修复**：增加预设模板的详细节点配置

**PM-D-02: 批量操作的 UI 交互未定义** [已修复]
- **问题**：批量分配 Issue 给 Agent 的前端交互未详细描述
- **影响**：前端开发时需要额外设计
- **修复**：增加批量操作的 UI 交互说明

**PM-D-03: 新用户引导流程需要更详细**
- **问题**：新用户引导只提到引导弹窗和快速开始模板，具体流程未定义
- **建议**：增加引导流程的详细步骤

---

### 2.3 技术骨干评审

#### ✅ 通过项
1. **数据库迁移设计完整** - 10张新表，索引设计合理
2. **路由注册设计清晰** - 新增路由组，依赖注入完整
3. **前端组件架构合理** - API/Composables/Views 分层清晰
4. **技术风险识别到位** - 5项风险及缓解措施

#### ⚠️ 问题与建议

**TECH-D-01: 前端流程设计器技术选型未确定** [已修复]
- **问题**：WorkflowCanvas 组件未确定使用什么流程图库
- **影响**：前端开发选型风险
- **修复**：推荐使用 vue-flow（Vue 3 原生支持）

**TECH-D-02: Workflow 执行的并发控制未设计** [已修复]
- **问题**：并行节点执行时的并发控制（goroutine 管理）未设计
- **影响**：可能导致资源竞争
- **修复**：增加并发控制设计（worker pool + semaphore）

**TECH-D-03: Agent 决策记录的存储策略**
- **问题**：AgentDecision 表可能增长很快，缺少归档策略
- **建议**：增加数据归档和清理策略

**TECH-D-04: 前端状态管理方案** [已修复]
- **问题**：复杂组件（如 WorkflowDesigner）的状态管理方案未确定
- **影响**：组件间状态同步复杂
- **修复**：使用 composable + provide/inject 模式

---

### 2.4 测试经理评审

#### ✅ 通过项
1. **API 设计可测试** - 每个端点都有明确的输入输出
2. **数据模型可验证** - 字段类型和约束清晰
3. **实施计划包含测试** - 单元测试和集成测试

#### ⚠️ 问题与建议

**QA-D-01: 测试策略未详细定义** [已修复]
- **问题**：只提到单元测试和集成测试，缺少 E2E 测试策略
- **影响**：测试覆盖不完整
- **修复**：增加 E2E 测试策略（Playwright）

**QA-D-02: Workflow 执行的测试用例未设计**
- **问题**：Workflow 执行引擎的测试用例（串行/并行/失败恢复）未设计
- **建议**：增加测试用例设计

**QA-D-03: 并发测试场景未覆盖** [已修复]
- **问题**：多用户同时执行同一 Workflow 的并发场景未覆盖
- **影响**：可能出现数据竞争
- **修复**：增加并发测试场景

---

### 2.5 用户代表评审

#### ✅ 通过项
1. **一键分配体验好** - 和分配给人类一样简单
2. **进度透明** - Agent 执行状态实时可见
3. **可视化编排** - 拖拽式配置降低使用门槛

#### ⚠️ 问题与建议

**UX-D-01: 执行预览功能需要更直观** [已修复]
- **问题**：执行预览只提到显示步骤和预估时间，缺少可视化展示
- **影响**：用户难以理解 Agent 将要做什么
- **修复**：增加流程图预览和成本预估可视化

**UX-D-02: 失败恢复操作需要更简单**
- **问题**：失败恢复提到重试、回滚、接管，但操作路径未设计
- **建议**：增加一键恢复按钮和引导

---

### 2.6 竞品分析员评审

#### ✅ 通过项
1. **全生命周期覆盖** - 领先 Linear 的 Issue 分配模式
2. **可视化编排** - 领先 CrewAI 的代码配置模式
3. **上下文交接机制** - 竞品普遍缺乏

#### ⚠️ 问题与建议

**COMP-D-01: 与 IDE 的集成方式未设计** [已修复]
- **问题**：设计文档只提到 CLI 集成，缺少 IDE 插件设计
- **影响**：开发者体验不完整
- **修复**：增加 VS Code 插件设计方案（Phase 2）

**COMP-D-02: Agent 模板市场未设计**
- **问题**：设计文档缺少社区贡献 Agent 模板的机制
- **影响**：生态扩展受限
- **建议**：Phase 3 考虑 Agent 模板市场

---

## 三、自动修复项汇总

| 编号 | 问题 | 修复内容 |
|------|------|----------|
| ARCH-D-01 | WorkflowNode 与 AgentTask 关系 | 增加执行关系说明 |
| ARCH-D-02 | ContextPayload 压缩策略 | 增加压缩策略说明 |
| PM-D-01 | 预设模板详细定义 | 增加模板节点配置 |
| PM-D-02 | 批量操作 UI 交互 | 增加 UI 交互说明 |
| TECH-D-01 | 流程设计器技术选型 | 推荐 vue-flow |
| TECH-D-02 | 并发控制设计 | 增加 worker pool 设计 |
| TECH-D-04 | 前端状态管理 | 使用 composable + provide/inject |
| QA-D-01 | 测试策略 | 增加 E2E 测试策略 |
| QA-D-03 | 并发测试场景 | 增加并发测试场景 |
| UX-D-01 | 执行预览可视化 | 增加流程图预览 |
| COMP-D-01 | IDE 集成设计 | 增加 VS Code 插件设计 |

---

## 四、设计补充内容

### 4.1 Workflow 预设模板详细配置

```json
{
  "name": "需求到上线",
  "description": "完整功能开发流程",
  "nodes": [
    {
      "name": "需求分析",
      "agent_type": "pm",
      "template_id": "product_manager",
      "timeout": 1800,
      "config": {
        "output_type": "requirement_spec",
        "wiki_doc_type": "requirement"
      }
    },
    {
      "name": "概要设计",
      "agent_type": "architect",
      "template_id": "architect",
      "timeout": 3600,
      "config": {
        "output_type": "high_level_design",
        "wiki_doc_type": "hld"
      }
    },
    {
      "name": "详细设计",
      "agent_type": "developer",
      "template_id": "developer",
      "timeout": 3600,
      "config": {
        "output_type": "low_level_design",
        "wiki_doc_type": "lld"
      }
    },
    {
      "name": "测试执行",
      "agent_type": "tester",
      "template_id": "qa_engineer",
      "timeout": 7200,
      "config": {
        "output_type": "test_report",
        "wiki_doc_type": "test_report"
      }
    }
  ],
  "edges": [
    {"source": 0, "target": 1},
    {"source": 1, "target": 2},
    {"source": 2, "target": 3}
  ]
}
```

### 4.2 批量操作 UI 交互说明

```
Issue 列表页面
    │
    ├─► 复选框选择多个 Issue
    │
    ├─► 点击顶部工具栏 "分配给 AI" 按钮
    │
    ├─► 弹出 Agent 选择对话框
    │     ├─► 显示可用 Agent 列表
    │     ├─► 支持搜索和筛选
    │     └─► 选择目标 Agent
    │
    ├─► 确认分配
    │     ├─► 显示分配预览（Issue 数量、预估成本）
    │     └─► 点击确认
    │
    └─► 执行分配
          ├─► 显示进度条
          ├─► 实时更新分配状态
          └─► 完成后显示结果摘要
```

### 4.3 流程设计器技术选型

**推荐：vue-flow**

| 特性 | 说明 |
|------|------|
| Vue 3 原生支持 | Composition API、TypeScript |
| 轻量级 | 核心包 ~50KB |
| 可定制 | 节点/边可自定义组件 |
| 性能好 | 支持 1000+ 节点 |
| 社区活跃 | GitHub 4k+ stars |

```bash
npm install @vue-flow/core @vue-flow/background @vue-flow/controls
```

### 4.4 并发控制设计

```go
// internal/service/workflow_executor.go

type WorkflowExecutor struct {
    // ... 其他字段 ...
    semaphore chan struct{} // 并发控制信号量
    maxParallel int        // 最大并行数
}

func NewWorkflowExecutor(...) *WorkflowExecutor {
    return &WorkflowExecutor{
        // ... 其他初始化 ...
        semaphore:   make(chan struct{}, 10), // 最大10个并行 Workflow
        maxParallel: 5,                       // 每个 Workflow 最大5个并行节点
    }
}

func (e *WorkflowExecutor) executeParallel(run *model.WorkflowRun, nodes []model.WorkflowNode, ctx *ContextPayload) error {
    var wg sync.WaitGroup
    errCh := make(chan error, len(nodes))
    
    for _, node := range nodes {
        wg.Add(1)
        go func(n model.WorkflowNode) {
            defer wg.Done()
            
            // 获取并发许可
            e.semaphore <- struct{}{}
            defer func() { <-e.semaphore }()
            
            if err := e.executeNode(run, &n, ctx); err != nil {
                errCh <- err
            }
        }(node)
    }
    
    wg.Wait()
    close(errCh)
    
    // 收集错误
    var errs []error
    for err := range errCh {
        errs = append(errs, err)
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("parallel execution failed: %v", errs)
    }
    return nil
}
```

### 4.5 E2E 测试策略

```typescript
// e2e/workflow-execution.spec.ts

test.describe('Workflow Execution', () => {
  test('should execute sequential workflow', async ({ page }) => {
    // 1. 创建工作流
    // 2. 添加节点和连接
    // 3. 执行工作流
    // 4. 验证执行结果
  });

  test('should execute parallel workflow', async ({ page }) => {
    // 1. 创建并行工作流
    // 2. 验证并行执行
    // 3. 验证上下文合并
  });

  test('should handle workflow failure', async ({ page }) => {
    // 1. 创建会失败的工作流
    // 2. 执行工作流
    // 3. 验证失败恢复机制
  });

  test('should respect budget limit', async ({ page }) => {
    // 1. 设置预算限制
    // 2. 执行超出预算的工作流
    // 3. 验证预算阻止
  });
});
```

### 4.6 VS Code 插件设计（Phase 2）

```
vscode-reqmango/
├── package.json
├── src/
│   ├── extension.ts          # 插件入口
│   ├── views/
│   │   ├── IssuePanel.ts     # Issue 面板
│   │   └── AgentStatus.ts    # Agent 状态
│   ├── commands/
│   │   ├── assignAgent.ts    # 分配给 Agent
│   │   └── viewWorkflow.ts   # 查看工作流
│   └── api/
│       └── client.ts         # API 客户端
└── README.md
```

**功能：**
- 侧边栏显示当前项目 Issue
- 右键菜单"分配给 AI Agent"
- 状态栏显示 Agent 执行状态
- 命令面板集成

---

## 五、评审结论

| 角色 | 结论 |
|------|------|
| 架构师 | ✅ 通过 |
| 产品经理 | ✅ 通过 |
| 技术骨干 | ✅ 通过 |
| 测试经理 | ✅ 通过 |
| 用户代表 | ✅ 通过 |
| 竞品分析员 | ✅ 通过 |

**总体结论：设计评审通过，11项问题已自动修复，可进入编码实现阶段。**

---

## 六、实施建议

### 6.1 开发顺序

| 阶段 | 任务 | 依赖 |
|------|------|------|
| 1.1 | 数据库迁移 | 无 |
| 1.2 | AgentMember Service/Handler | 1.1 |
| 1.3 | IssueAgent Service/Handler | 1.1, 1.2 |
| 1.4 | AgentDecision Service | 1.1 |
| 1.5 | AgentCostBudget Service | 1.1 |
| 1.6 | AgentSLA Service | 1.1 |
| 1.7 | ContextPayload Service | 1.1 |
| 1.8 | Workflow Service/Handler | 1.1-1.7 |
| 1.9 | Workflow 执行引擎 | 1.8 |
| 1.10 | 路由注册 | 1.2-1.9 |
| 1.11 | 前端 API 层 | 1.10 |
| 1.12 | 前端组件 | 1.11 |

### 6.2 技术债务预防

1. **代码审查** - 每个 PR 必须经过审查
2. **单元测试** - 核心逻辑覆盖率 > 80%
3. **集成测试** - API 端点全覆盖
4. **文档同步** - 代码变更同步更新设计文档
