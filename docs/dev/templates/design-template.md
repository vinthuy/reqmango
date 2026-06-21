# {Feature Name} — Feature Design

**创建日期**: {YYYY-MM-DD}
**状态**: Draft | Approved
**后端目标**: Go/Gin
**前端**: Vue 3

---

## 1. Spec（需求规格）

### 1.1 用户故事

> 作为 {角色}，我想要 {动作}，以便 {收益}。

### 1.2 功能需求

- **R1**: {需求描述}
- **R2**: {需求描述}

### 1.3 验收标准

- [ ] **AC1**: {验收条件}
- [ ] **AC2**: {验收条件}

### 1.4 不包含（Out of Scope）

- {明确排除的内容}

### 1.5 依赖

- KB 参考文档: [kb/architecture/xxx.md](../kb/architecture/xxx.md)
- 前置功能: {如果有}

---

## 2. Design（技术设计）

### 2.1 API 设计

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/v1/{resource} | {描述} | JWT |
| POST | /api/v1/{resource} | {描述} | JWT |

#### 请求示例

```json
{
  "field": "value"
}
```

#### 响应示例

```json
{
  "id": 1,
  "field": "value"
}
```

### 2.2 数据模型

#### 新表: `{table_name}`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 主键 |
| name | string | 名称 |
| created_at | time.Time | 创建时间 |

#### 关联表: `{join_table}`

| 字段 | 类型 | 说明 |
|------|------|------|
| a_id | uint64 | 外键 A |
| b_id | uint64 | 外键 B |

### 2.3 服务层

```
{ServiceName}:
  - {Method}(...): {简要逻辑描述}
```

### 2.4 前端组件

| 组件 | 用途 | 类型 |
|------|------|------|
| {ComponentName}.vue | {用途} | Create/Modify |
| {ComponentName}.vue | {用途} | Create/Modify |

### 2.5 Store 接口

```typescript
// stores/{name}.ts
export const use{Name}Store = defineStore('{name}', () => {
  // state + actions
})
```

### 2.6 路由（如新增页面）

| Route | View | Description |
|-------|------|-------------|
| /workspace/:slug/project/:id/{path} | {ViewName}.vue | {描述} |

### 2.7 文件变更清单

| File | Action | Description |
|------|--------|-------------|
| `backend-go/internal/model/{name}.go` | Create | DB 模型 |
| `backend-go/internal/dto/request/{name}.go` | Create | 请求 DTO |
| `backend-go/internal/dto/response/{name}.go` | Create | 响应 DTO |
| `backend-go/internal/service/{name}_service.go` | Create | 业务逻辑 |
| `backend-go/internal/handler/{name}_handler.go` | Create | HTTP handler |
| `backend-go/internal/router/router.go` | Modify | 注册路由 |
| `frontend/src/api/{name}.ts` | Create/Modify | API 调用 |
| `frontend/src/types/{name}.ts` | Create/Modify | TypeScript 类型 |
| `frontend/src/stores/{name}.ts` | Create/Modify | Pinia store |
| `frontend/src/components/{Name}*.vue` | Create/Modify | 组件 |
| `frontend/src/views/{Name}*.vue` | Create/Modify | 视图 |
