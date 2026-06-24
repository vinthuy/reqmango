# KB Changelog（知识库变更日志）

记录每次知识库更新，按时间倒序。

---

## 2026-06-21 — 文档体系重组

**类型**: 文档重组

**变更内容**:
- 建立 `kb/`（全量知识库）+ `dev/`（增量需求开发）+ `superseded/`（历史归档）三层文档架构
- 从 `TECH_ARCHITECTURE.md` 拆分出 7 个专题架构文档
- 将 Python 时代 SDD 文档（12 个文件）归档到 `superseded/python-era/`
- 将 Cycle/Module 功能设计文档迁移到 `dev/features/`
- 创建 4 个标准化模板（design, plan, tasks, review）
- 创建 `pipeline-status.md` 替代旧的 `spec.md` 状态追踪

**影响文件**:
- `kb/architecture/tech-stack.md` — 新建
- `kb/architecture/backend-go.md` — 新建
- `kb/architecture/backend-python.md` — 新建
- `kb/architecture/frontend.md` — 新建
- `kb/architecture/data-model.md` — 新建
- `kb/architecture/api-conventions.md` — 新建
- `kb/architecture/project-layout.md` — 新建
- `kb/PRD.md` — 从 docs/PRD.md 迁移

---

## 2026-06-25 — 补充工作项管理关键缺失功能 (5 Features)

**类型**: Feature

**变更内容**:
- Feature G: 统一 Project Settings 页面（4→12 tabs），合并新旧两套实现
- Feature F: 实现通知系统后端（model + service + handler，8 个 API 端点）
- Feature A: 实现 Saved Views 保存视图（JSONB 存储筛选/排序/列配置，7 个 API 端点 + 前端选择器）
- Feature C: 项目级 Work Item Type 配置（CopyFromWorkspace + Reorder，4 个 API 端点 + 前端管理组件）
- Feature B: 实现 Pages/Wiki 文档系统（层级树 depth≤5，archive/restore/move，10 个 API 端点 + 前端页面编辑器）
- Bug 修复: 4 个 handler 的 getUserID 从 `c.Get("user_id")` 修正为 `c.Get("currentUser")` 获取 `*model.User.ID`

**影响文件** (33 files, +3137/-399):
- 新增 13 个 Go 后端文件（3 models, 3 request DTOs, 3 response DTOs, 3 services, 1 handler + 2 额外 handlers）
- 新增 8 个前端文件（2 types, 2 APIs, 3 components, 1 view）
- 修改 8 个现有文件（main.go, router.go, seed.go, issue_type_service.go, Project.vue, ProjectSettings.vue, router/index.ts, types/project.ts）
- 删除 1 个文件（components/ProjectSettings.vue）

**KB 更新**:
- `architecture/README.md` — 模块状态表更新（新增 6 个完成模块）
- `architecture/data-model.md` — 新增 3 张表（notifications, saved_views, pages），移除已迁移表
- `architecture/frontend.md` — 更新文件数量（API 15→17, Types 12→14, Views 14→15, Components 32→35, Routes 10→13）

**关联提交**: `da8ba26`

---

## 使用说明

后续每次功能完成并更新 KB 时，在此追加一条记录：

```markdown
## {YYYY-MM-DD} — {简短描述}

**类型**: Feature | Architecture | Fix

**变更内容**:
- {变更点 1}
- {变更点 2}

**关联功能**: {dev/archive/{date}-{slug}/ | 或关联的 PR/Branch}
```
