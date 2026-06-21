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
