# reqmango Documentation

reqmango 项目管理平台的文档中心。本文档库分为四大区域：

- **Knowledge Base（全量知识库）**：描述系统**当前是什么**，是唯一真相来源
- **Development Pipeline（增量需求开发）**：管理**正在构建什么**，有完整的生命周期
- **Superpowers（AI 辅助设计）**：AI Agent 生成的设计文档和实现计划
- **Superseded（历史归档）**：已过时的旧文档，仅供历史参考

---

## 快速导航

| 想了解…… | 入口 |
|----------|------|
| 产品功能定义 | [kb/PRD.md](kb/PRD.md) |
| 项目目录结构 | [kb/architecture/project-layout.md](kb/architecture/project-layout.md) |
| Go 后端架构 | [kb/architecture/backend.md](kb/architecture/backend.md) |
| 前端架构 | [kb/architecture/frontend.md](kb/architecture/frontend.md) |
| 数据模型总览 | [kb/architecture/data-model.md](kb/architecture/data-model.md) |
| API 约定 | [kb/architecture/api-conventions.md](kb/architecture/api-conventions.md) |
| 系统技术栈 | [kb/architecture/tech-stack.md](kb/architecture/tech-stack.md) |
| 各功能开发状态 | [pipeline-status.md](dev/pipeline-status.md) |
| 当前正在开发的功能 | [dev/active/](dev/active/) |
| 如何开始一个新功能 | [dev/README.md](dev/README.md) |
| 功能开发模板 | [dev/templates/](dev/templates/) |
| 已过时的 Python 文档 | [superseded/README.md](superseded/README.md) |

---

## 如何使用本文档库

### 场景一：了解系统全貌

从 [kb/README.md](kb/README.md) 开始，按顺序阅读架构文档。

### 场景二：开发新功能

1. 查看 [dev/pipeline-status.md](dev/pipeline-status.md) 了解当前状态
2. 阅读相关 KB 架构文档了解现有实现
3. 从 [dev/templates/](dev/templates/) 复制模板
4. 按流程在 [dev/features/](dev/features/) 中创建功能文档
5. 完成后更新 KB 并归档

### 场景三：AI Agent 工作

如果你是 AI Agent：
1. 先读 [dev/pipeline-status.md](dev/pipeline-status.md) 了解上下文
2. 读相关 KB 文档了解现有模式
3. 按 [dev/README.md](dev/README.md) 中的流程创建/更新文档
4. 实现完成后填写 review 并触发 KB 更新

---

## 技术栈速览

| 层 | 当前 | 状态 |
|----|------|------|
| 后端 | Go + Gin + GORM + PostgreSQL 16 | 主力 |
| MCP Server | Go (stdio/SSE) + JSON-RPC 2.0 | 独立模块 |
| 前端 | Vue 3 + TypeScript + Pinia + Tailwind CSS | 主力 |
| 遗留后端 | Python + FastAPI + SQLAlchemy | 已淘汰 |

---

## 目录总览

```
docs/
├── README.md                         # 你在这里
├── reqmango-vs-planeai.md            # 竞品对标分析
├── kb/                               # 全量知识库
│   ├── PRD.md                        # 产品需求文档
│   ├── architecture/                 # 架构参考（9 文档）
│   └── changelog/                    # KB 变更日志
├── dev/                              # 增量需求开发
│   ├── features/                     # 功能设计文档
│   ├── templates/                    # 标准化模板
│   ├── active/                       # 当前活跃功能
│   └── archive/                      # 已完成归档
├── superpowers/                      # AI 辅助设计
│   ├── plans/                        # 实现计划
│   └── specs/                        # 设计规范
└── superseded/                       # 历史已淘汰文档
```
