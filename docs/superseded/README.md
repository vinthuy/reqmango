# Superseded Documents（历史归档）

本目录存放已被淘汰或取代的文档，仅供历史参考。

**这些文档不代表当前系统的状态。** 查看当前系统请阅读 [kb/](../kb/README.md) 中的知识库文档。

---

## 目录说明

### python-era/ — Python/FastAPI 时代 SDD 文档

这些文档编写于 2026-06-13 ~ 2026-06-14，基于 Python/FastAPI 后端的 SDD（Schema-Driven Development）流程。

**被取代原因**: 项目于 2026-06-19 决策将后端从 Python/FastAPI 重写为 Go/Gin。这些文档中描述的数据模型（SQLAlchemy）、Schema 定义（Pydantic）、API 路由（FastAPI）不再适用于当前的 Go 后端。

包含：
- `spec-old.md` — 旧版实现状态追踪（仅覆盖 Python 后端）
- `sdd-issue-*.md` — Issue 模块 SDD 文档三件套
- `sdd-cycle-*.md` — Cycle 模块 SDD 文档三件套
- `sdd-module-*.md` — Module 模块 SDD 文档三件套
- `sdd-issue-custom-fields.md` — 工作项与自定义字段功能文档
- `tech-architecture-old.md` — 旧版技术架构文档（Python 技术栈）

### pages-archive/ — 旧 pages 目录

`pages/` 目录中的设计文档已迁移到 [dev/features/](../dev/features/)，此目录为空归档占位。

---

## 关于 Tech Architecture

原始 `TECH_ARCHITECTURE.md`（~1500 行）被拆分为 `kb/architecture/` 下的多个专题文档：
- `tech-stack.md`
- `backend-go.md`
- `backend-python.md`
- `frontend.md`
- `data-model.md`
- `api-conventions.md`
- `project-layout.md`

原始文件保存在 `python-era/tech-architecture-old.md`，包含 Python/FastAPI 时代的完整代码示例（Pydantic Schema、SQLAlchemy Model、FastAPI Router 等）。
