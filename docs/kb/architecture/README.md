# Architecture Overview

ReqManPy 采用前后端分离架构，当前技术体系如下。

---

## 技术栈总览

| 层级 | 当前技术 | 状态 | 旧技术（已淘汰） |
|------|----------|------|------------------|
| 后端框架 | Go + Gin 1.x | ✅ 主力 | Python + FastAPI |
| ORM | GORM 2.x | ✅ 主力 | SQLAlchemy 2.0 (async) |
| 数据库 | PostgreSQL | ✅ | SQLite (dev) |
| 认证 | JWT (golang-jwt/v5) | ✅ | python-jose |
| 前端框架 | Vue 3 + Composition API | ✅ | — |
| 前端构建 | Vite | ✅ | — |
| 状态管理 | Pinia | ✅ | — |
| CSS | Tailwind CSS | ✅ | — |
| 类型系统 | TypeScript | ✅ | — |

---

## 系统分层

```
┌──────────────────────────────────────────┐
│              前端 (Vue 3)                 │
│  Views → Components → Stores → API →     │
│         TypeScript Types                 │
└──────────────────┬───────────────────────┘
                   │ HTTP (JSON + JWT)
┌──────────────────▼───────────────────────┐
│            Go 后端 (Gin)                  │
│                                           │
│  Router → Middleware → Handler           │
│                          │                │
│                        Service            │
│                          │                │
│                    Model (GORM)           │
│                          │                │
│                      PostgreSQL           │
└──────────────────────────────────────────┘
```

## 后端分层职责

| 层 | 目录 | 职责 |
|----|------|------|
| Handler | `internal/handler/` | 绑定请求参数，调用 Service，写 HTTP 响应。不含业务逻辑 |
| Service | `internal/service/` | 纯业务逻辑，跨 Model 操作，返回 AppError |
| Model | `internal/model/` | GORM 模型定义，表结构映射 |
| DTO | `internal/dto/` | 请求/响应结构体，与 Model 解耦 |
| Middleware | `internal/middleware/` | JWT 认证、CORS、日志 |
| Router | `internal/router/` | 路由注册，挂载中间件和 Handler |
| Common | `internal/common/` | 错误类型、常量、分页工具 |

## 前端分层职责

| 层 | 目录 | 职责 |
|----|------|------|
| Views | `src/views/` | 页面级组件，对应路由 |
| Components | `src/components/` | 可复用组件 |
| Stores | `src/stores/` | Pinia 状态管理 |
| API | `src/api/` | Axios 封装，后端 API 调用 |
| Types | `src/types/` | TypeScript 类型定义 |
| Router | `src/router/` | Vue Router 路由配置 |

## 已实现模块一览

| 模块 | Go 后端 | 前端 | 状态 |
|------|---------|------|------|
| Auth（认证） | ✅ | ✅ | 完成 |
| Workspace（工作空间） | ✅ | ✅ | 完成 |
| Project（项目） | ✅ | ✅ | 完成 |
| Issue（工作项） | ✅ | ✅ | 完成 |
| State（状态） | ✅ | ✅ | 完成 |
| Label（标签） | ✅ | ✅ | 完成 |
| Cycle（周期） | ✅ | ✅ | 完成 |
| Module（模块） | ✅ | ✅ | 完成 |
| CustomField（自定义字段） | ✅ | ✅ | 完成 |
| IssueType（工作项类型） | ✅ | ✅ | 完成 |
| Workflow（工作流） | ✅ | ✅ | 完成 |
| Automation（自动化） | ✅ | ✅ | 完成 |
| Comments（评论） | ✅ | ✅ | 完成 |
| Notifications（通知） | ✅ | ✅ | 完成 |
| Saved Views（保存视图） | ✅ | ✅ | 完成 |
| Pages（页面文档） | ✅ | ✅ | 完成 |
| EstimatePoint（估算点） | ❌ | ✅ | 仅前端 |
| Attachments（附件） | ❌ | ✅ | 仅前端 |
| AI | ❌ | ❌ | 未实现 |

## 参考架构

- [plane-enterprise-info-architecture.md](plane-enterprise-info-architecture.md) — Plane Enterprise 信息架构（对标参考）

## 文档索引

- [tech-stack.md](tech-stack.md) — 技术栈详情
- [backend-go.md](backend-go.md) — Go 后端架构
- [backend-python.md](backend-python.md) — Python 后端（遗留）
- [frontend.md](frontend.md) — 前端架构
- [data-model.md](data-model.md) — 数据模型总览
- [api-conventions.md](api-conventions.md) — API 设计约定
- [project-layout.md](project-layout.md) — 项目目录结构
