# Development Pipeline（增量需求开发）

本文档描述 ReqManPy 的增量功能开发流程。每个功能从需求提出到完成归档，遵循标准化的生命周期。

---

## 功能生命周期

```
BACKLOG ──→ SPEC ──→ DESIGN ──→ PLAN ──→ IMPLEMENT ──→ REVIEW ──→ KB UPDATE ──→ ARCHIVE
```

### 阶段 1: Backlog（需求池）

功能想法记录在 [pipeline-status.md](pipeline-status.md) 的 Backlog 行中。仅需一行描述 + 优先级标记，不创建独立目录。

### 阶段 2: Spec（需求规格）

创建 `dev/features/{YYYY-MM-DD}-{slug}/design.md`，包含：
- 用户故事和场景
- 功能需求列表（R1, R2, ...）
- 验收标准（AC1, AC2, ...）
- 非功能性约束

**产出**: 可供评审的需求文档

### 阶段 3: Design（技术设计）

在同一 `design.md` 中扩展以下章节：
- API 路由设计（方法、路径、请求/响应格式）
- 数据模型变更（新表、字段变更）
- 服务层逻辑要点
- 前端组件树和路由
- 完整文件变更清单

**关键约束**: 设计必须引用 KB 中现有模块的实现模式，保持一致性。参考 [kb/architecture/](../kb/architecture/) 中的文档。

**产出**: 可供开发人员执行的技术设计

### 阶段 4: Plan（实施计划）

创建 `plan.md`，将设计分解为可执行的步骤：
- 每步有明确的文件列表和变更描述
- 步骤之间有明确的依赖关系
- 每步有 checkbox 供跟踪
- 面向 AI Agent：格式设计为 AI 可理解和执行

**产出**: 可供 AI Agent 或开发者顺序执行的计划

### 阶段 5: Implement（实施）

- 将 [active/README.md](active/README.md) 指向当前功能
- 按 plan 逐步实施
- 更新 `plan.md` 中的 checkbox 状态
- 实施过程中发现的设计偏差记录在 `review.md` 中

### 阶段 6: Review（审查）

创建 `review.md`，包含：
- 验收标准 checklist（全部通过？）
- 与原始设计的偏离记录
- 测试结果（后端 + 前端）
- KB 文档更新清单

### 阶段 7: KB Update + Archive（知识库更新 + 归档）

- 按 review 中的清单更新 KB 文档
- 在 [kb/changelog/README.md](../kb/changelog/README.md) 中添加条目
- 更新 [pipeline-status.md](pipeline-status.md) 状态
- 将功能目录从 `features/` 移动到 `archive/`

---

## 目录结构

```
dev/
├── README.md                # 你在这里
├── pipeline-status.md       # 所有功能的状态一览
├── features/                # 开发中的功能
│   └── {date}-{slug}/
│       ├── design.md        # Spec + Design
│       ├── plan.md          # Implementation plan
│       └── review.md        # Completion review
├── active/                  # 当前活跃功能指针
├── archive/                 # 已完成归档
└── templates/               # 可复用的模板
```

## 与 KB 的关系

```
dev/features/        kb/
    │                   │
    │  Spec/Design      │  参考现有实现
    │ ──────────────→   │
    │                    │
    │  Implement         │
    │  (编码实现)        │
    │                    │
    │  Review + KB更新   │  写入最终状态
    │ ──────────────→   │
    │                    │
    │  Archive           │
```

功能文档是**过程性**的（记录怎么做的），KB 是**结果性**的（记录做成什么样）。
功能完成后，关键知识沉淀到 KB，过程文档归档。

## 对于 AI Agent

1. 从 [pipeline-status.md](pipeline-status.md) 了解当前状态
2. 对 New Feature 任务：先创建 `design.md`（Spec + Design），再创建 `plan.md`
3. 对已有 Plan 的任务：直接执行，更新 checkbox
4. 完成后：填写 `review.md`，触发 KB 更新
