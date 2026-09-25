# Superseded — Agent 平台叙事（已废止）

本目录存放 **Harness / Loop / 独立 Agent 控制台式 Agent 控制台** 相关设计与 PRD。  
它们描述的是历史或远期基础设施方向，**不再指导产品排期**。

## 当前产品真相

| 主题 | 文档 |
|------|------|
| 产品方向（原生 AI 项目管理路径 AI PM） | [kb/PRD.md](../../kb/PRD.md) · [2026-09-25 redesign](../../superpowers/specs/2026-09-25-ai-project-management-redesign.md) |
| 原生 PM + 核心定制验收 | [product-core-qa](../../dev/acceptance/2026-09-25-product-core-qa.md) |
| 降级 backlog（代码清理） | [agent-platform-demotion](../../dev/debt/2026-09-25-agent-platform-demotion.md) |

## 本目录清单

| 文件 | 原路径 | 废止原因 |
|------|--------|----------|
| `AI_AGENT_PRD.md` | `docs/AI_AGENT_PRD.md` | 独立 Agent 平台 / 全 SDLC Agent 主叙事废止 |
| `PRD-Agent-Project-Integration.md` | `docs/PRD-Agent-Project-Integration.md` | Agent-as-teammate 产品化已降级 |
| `ARCH-Agent-Project-Integration.md` | `docs/ARCH-Agent-Project-Integration.md` | 同上 |
| `agent-project-integration-plan.md` | `docs/agent-project-integration-plan.md` | 同上 |
| `2026-07-18-reqmango-agent-platform-design.md` | `superpowers/specs/` | Harness/Loop 不作产品主路径 |
| `2026-07-18-phase1-agent-loop-mvp.md` | `superpowers/plans/` | Loop MVP 计划归档 |
| `2026-07-18-phase2-harness-engine.md` | `superpowers/plans/` | Harness 计划归档 |

## 仍保留在代码中、但非产品主叙事

- 开箱 PM Agent（ensure-pm）、Issue 指派 / `@Agent`、`dispatch_agent`
- Intake / Analyze / automation-preview / Dashboard `ai_summary`
- `/agents/*` 路由默认重定向（需 `localStorage.rm_advanced_agents=1` 解锁）

归档日期：2026-09-26
