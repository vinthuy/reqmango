# 自定义关联关系架构设计

> 参考 PlaneAI Enterprise Grid Relations 设计

---

## 1. PlaneAI 参考模型

PlaneAI 支持三种关联关系：

### 1.1 内置关系
| 类型 | 触发 | 说明 |
|------|------|------|
| Parent-Child | 层级结构 | 父子工作项，深度最多6层 |
| Dependency | 排程约束 | Blocked by/Blocking, Starts Before/After, Finishes Before/After |
| Logical | 逻辑关联 | Relates To, Duplicate, Implements |

### 1.2 自定义关系 (Enterprise Grid)
Workspace Admin 可创建自定义关系类型，指定：
- **Title**: 关系名称 (e.g., "Tests")
- **Inward name**: 从源工作项读取的名称 (e.g., "tested by")  
- **Outward name**: 从目标工作项读取的名称 (e.g., "tests")

所有项目自动可用。

---

## 2. 当前项目实现

### 2.1 数据模型

```
┌──────────────────────────────────────────┐
│           WORKSPACE LEVEL                │
│                                          │
│  ┌──────────────────────┐               │
│  │ RelationType         │               │
│  │ - name               │  e.g., "Blocks"
│  │ - inward_name        │  "blocked by" │
│  │ - outward_name       │  "blocks"     │
│  │ - workspace_id       │               │
│  └──────────────────────┘               │
└──────────────────────┬───────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────┐
│           ISSUE LEVEL                    │
│                                          │
│  ┌──────────────────────┐               │
│  │ IssueRelation         │               │
│  │ - issue_id            │  源工作项     │
│  │ - related_issue_id    │  目标工作项   │
│  │ - relation_type_id    │  关系类型     │
│  │ - comment             │  备注         │
│  └──────────────────────┘               │
└──────────────────────────────────────────┘
```

### 2.2 API

```
# 关系类型 (Workspace 级别)
POST   /relations/types?workspace_id=N   创建关系类型
GET    /relations/types?workspace_id=N   列表
PUT    /relations/types/:id              更新
DELETE /relations/types/:id              删除

# 工作项关联
POST   /issues/:id/relations             创建关联
GET    /issues/:id/relations             查看关联
DELETE /relations/:id                    删除关联
```

### 2.3 默认关系类型

| 名称 | Inward | Outward | 用途 |
|------|--------|---------|------|
| Blocks | blocked by | blocks | 阻塞关系 |
| Relates To | related to | relates to | 一般关联 |
| Duplicates | duplicated by | duplicates | 重复标记 |

### 2.4 架构定位

```
工作项关系体系
├── 层级关系 (Parent-Child)     ✅ 已实现 (parent_id + depth)
├── 自定义关联 (Relations)      ✅ 已实现 (本系统)
├── 依赖关系 (Dependencies)     ⏳ 后续 (排程约束)
└── 子工作项 (Sub-issues)       ✅ 已有 (parent_id)
```
