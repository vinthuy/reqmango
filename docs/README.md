# reqmango Documentation

reqmango 项目管理平台的文档中心。

- **Knowledge Base**：系统**当前是什么**（唯一真相来源）
- **Development Pipeline**：正在构建 / 已验收什么
- **Superpowers**：设计与实现计划（含已完成、已废止）
- **Superseded**：历史归档（**不指导排期**）

---

## 现行产品方向（2026-09）

| 做 | 不做 |
|----|------|
| 原生 PM（Issue / Cycle / 类型·字段·工作流·自动化） | Harness / Loop / Multica Agent 控制台产品化 |
| Plane 路径 AI（Intake / Analyze / 自动化预览 / 仪表盘摘要） | 默认暴露 `/agents*`（需高级解锁） |

- PRD：[kb/PRD.md](kb/PRD.md)
- 设计真相：[superpowers/specs/2026-09-25-ai-project-management-redesign.md](superpowers/specs/2026-09-25-ai-project-management-redesign.md)
- 验收 PASS：[dev/acceptance/2026-09-25-product-core-qa.md](dev/acceptance/2026-09-25-product-core-qa.md)
- Agent 旧文档归档：[superseded/agent-platform/](superseded/agent-platform/README.md)

---

## 快速导航

| 想了解…… | 入口 |
|----------|------|
| 产品功能定义 | [kb/PRD.md](kb/PRD.md) |
| 架构总览 | [kb/architecture/README.md](kb/architecture/README.md) |
| Go 后端 | [kb/architecture/backend-go.md](kb/architecture/backend-go.md) |
| 前端 | [kb/architecture/frontend.md](kb/architecture/frontend.md) |
| 数据模型 | [kb/architecture/data-model.md](kb/architecture/data-model.md) |
| API 约定 | [kb/architecture/api-conventions.md](kb/architecture/api-conventions.md) |
| 管线状态 | [dev/pipeline-status.md](dev/pipeline-status.md) |
| 当前焦点 | [dev/active/](dev/active/) |
| 历史归档 | [superseded/README.md](superseded/README.md) |

---

## 如何使用

### 了解系统

从 [kb/README.md](kb/README.md) 开始。

### 开发新功能

1. [pipeline-status.md](dev/pipeline-status.md)
2. 读相关 KB
3. 模板：[dev/templates/](dev/templates/)
4. 完成后更新 KB；**不要**复活已废止的 Agent 平台叙事

### AI Agent

先读 pipeline + KB；废止文档仅在 `superseded/`，勿当作排期依据。

---

## 技术栈速览

| 层 | 当前 | 状态 |
|----|------|------|
| 后端 | Go + Gin + GORM + PostgreSQL 16 | 主力 |
| MCP | Go (stdio/SSE) | 独立模块 |
| 前端 | Vue 3 + TypeScript + Pinia + Tailwind | 主力 |
| 遗留后端 | Python + FastAPI | 已淘汰 → superseded/python-era |

---

## 目录总览

```
docs/
├── README.md
├── kb/                 # 全量知识库（真相）
├── dev/                # 管线 / 验收 / debt
├── superpowers/        # 设计与计划（含现行 redesign）
├── superseded/         # 历史：python-era + agent-platform
└── assets/             # README 媒体占位
```
