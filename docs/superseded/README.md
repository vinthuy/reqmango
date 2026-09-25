# Superseded Documents（历史归档）

本目录存放已淘汰或被取代的文档，**仅供历史参考**。  
**它们不代表当前系统状态。** 查看现行系统请阅读 [kb/](../kb/README.md)。

---

## 目录说明

### agent-platform/ — Agent 平台 / Harness·Loop 叙事（2026-09-26 归档）

产品方向改为「原生 PM + 原生 AI 项目管理路径 AI」后，下列文档**废止排期指导**：

- 独立 Agent 平台 / 全 SDLC Agent PRD
- Agent-Project Integration PRD/ARCH/Plan
- Harness / Loop 设计与实施计划

入口：[agent-platform/README.md](agent-platform/README.md)  
现行方向：[kb/PRD.md](../kb/PRD.md) · [AI PM redesign](../superpowers/specs/2026-09-25-ai-project-management-redesign.md)

### python-era/ — Python/FastAPI 时代 SDD 文档

编写于 2026-06-13 ~ 2026-06-14，基于 Python/FastAPI 后端的 SDD 流程。

**被取代原因**: 2026-06-19 起后端重写为 Go/Gin。其中的 SQLAlchemy / Pydantic / FastAPI 描述不再适用。

包含：spec-old.md、sdd-issue-*、sdd-cycle-*、sdd-module-*、sdd-issue-custom-fields.md、	ech-architecture-old.md

### pages-archive/

原 pages/ 设计已迁至 [dev/features/](../dev/features/)，此处为空占位。

---

## 关于 Tech Architecture

原始 TECH_ARCHITECTURE.md 已拆分为 kb/architecture/ 下专题文档（	ech-stack、ackend-go、rontend、data-model 等）。完整旧稿见 python-era/tech-architecture-old.md。
