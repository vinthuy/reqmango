# {Feature Name} — Implementation Plan

> **面向 AI Agent**: 本计划设计为 AI 可理解和执行的格式。使用 Skill: `superpowers:subagent-driven-development` 或 `superpowers:executing-plans`。

**Spec/Design 参考**: `dev/features/{date}-{slug}/design.md`
**总步骤数**: {N}
**目标后端**: Go/Gin

---

## 文件结构总览

### Go Backend

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/dto/request/{name}.go` | Create | Request structs |
| `internal/dto/response/{name}.go` | Create | Response structs |
| `internal/model/{name}.go` | Create/Modify | DB model |
| `internal/service/{name}_service.go` | Create/Modify | Business logic |
| `internal/handler/{name}_handler.go` | Create/Modify | HTTP handlers |
| `internal/router/router.go` | Modify | Route registration |

### Vue Frontend

| File | Action | Responsibility |
|------|--------|---------------|
| `src/types/{name}.ts` | Create/Modify | TypeScript types |
| `src/api/{name}.ts` | Create/Modify | API client |
| `src/stores/{name}.ts` | Create/Modify | Pinia store |
| `src/components/{Name}*.vue` | Create/Modify | Components |
| `src/views/{Name}*.vue` | Create/Modify | Views |
| `src/router/index.ts` | Modify | Routes (if new pages) |

---

## Task 1: {Task Name}

**Files:**
- Create: `...`
- Modify: `...`

- [ ] **Step 1: {Step description}**

{具体代码/变更说明}

- [ ] **Step 2: {Step description}**

{具体代码/变更说明}

---

## Task 2: {Task Name}

**Files:**
- Create: `...`
- Modify: `...`

- [ ] **Step 1: {Step description}**

---

## Task N: Test + Verify

- [ ] **curl test**
  ```bash
  curl -X POST http://localhost:8080/api/v1/... -H "Authorization: Bearer $TOKEN" ...
  ```

- [ ] **Frontend smoke test**
  - 打开 {页面}
  - 验证 {功能}

- [ ] **KB update**
  - 参考 `review.md` 中的 KB 更新清单
