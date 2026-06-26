# Architecture Overview

ReqManPy 采用前后端分离架构，Go + Vue 3 全栈。

**最后更新**: 2026-06-27

---

## 技术栈总览

| 层级 | 技术 | 状态 |
|------|------|------|
| 后端框架 | Go + Gin 1.x | ✅ 主力 |
| ORM | GORM 2.x | ✅ |
| 数据库 | PostgreSQL 16+ | ✅ |
| 认证 | JWT (golang-jwt/v5) | ✅ |
| LLM 集成 | DeepSeek / Anthropic / OpenAI-compatible | ✅ |
| 前端框架 | Vue 3 + Composition API | ✅ |
| 前端构建 | Vite | ✅ |
| 状态管理 | Pinia | ✅ |
| CSS | Tailwind CSS | ✅ |
| 类型系统 | TypeScript | ✅ |

---

## 系统分层

```
┌──────────────────────────────────────────┐
│              前端 (Vue 3)                 │
│  Views → Components → Stores → API →     │
│         TypeScript Types                 │
└──────────────────┬───────────────────────┘
                   │ HTTP (JSON + JWT) + SSE (AI streaming)
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
| Handler | `internal/handler/` (27) | HTTP 请求绑定，调用 Service，写响应。不含业务逻辑 |
| Service | `internal/service/` (26) | 纯业务逻辑，跨 Model 操作，返回 AppError |
| Model | `internal/model/` (27) | GORM 模型定义，表结构映射 |
| DTO | `internal/dto/` | 请求/响应结构体，与 Model 解耦 |
| Middleware | `internal/middleware/` | JWT 认证、CORS、日志 |
| Router | `internal/router/` | 路由注册，挂载中间件和 Handler |
| Common | `internal/common/` | 错误类型、常量、分页工具 |
| RQL | `internal/rql/` | ReqMan Query Language (lexer/parser/executor) |

## 前端分层职责

| 层 | 目录 | 职责 |
|----|------|------|
| Views | `src/views/` (16) | 页面级组件，对应路由 |
| Components | `src/components/` (57) | 可复用组件 |
| Stores | `src/stores/` (3) | Pinia 状态管理 |
| API | `src/api/` (22) | Axios 封装，后端 API 调用 |
| Types | `src/types/` (19) | TypeScript 类型定义 |
| Composables | `src/composables/` (3) | 组合式函数 (useConfirm, useRQL, useAI) |
| Router | `src/router/` | Vue Router 配置 (16 条路由) |

---

## 已实现模块一览

| 模块 | Go 后端 | 前端 | 说明 |
|------|---------|------|------|
| Auth（认证） | ✅ | ✅ | JWT register/login/me |
| Workspace（工作空间） | ✅ | ✅ | CRUD + 成员管理 |
| Project（项目） | ✅ | ✅ | CRUD + 归档 + 统计 + Lead + Subscriber |
| Issue（工作项） | ✅ | ✅ | CRUD + 层级/归档/批量/导入导出/Issue-Page关联 |
| State（状态） | ✅ | ✅ | 5 固定分组，项目级 CRUD |
| Label（标签） | ✅ | ✅ | 项目级 CRUD |
| Cycle（周期） | ✅ | ✅ | CRUD + 开始/结束/取消 + 燃尽图 + 进度 |
| Module（模块） | ✅ | ✅ | 树形 CRUD + Issue 关联 + 进度统计 |
| CustomField（自定义字段） | ✅ | ✅ | 7 种类型 + 选项管理 + Issue 值绑定 |
| IssueType（工作项类型） | ✅ | ✅ | 工作空间级 + 项目级，层级绑定 |
| TypeTemplate | ✅ | ✅ | 工作空间级类型蓝图 |
| ProjectTemplate | ✅ | ✅ | 项目模板 + Apply |
| WorkItemTemplate | ✅ | ✅ | 工作项创建模板 |
| Workflow（工作流） | ✅ | ✅ | 状态流转 + 审批规则 |
| Automation（自动化） | ✅ | ✅ | Trigger-Condition-Action |
| Comment（评论） | ✅ | ✅ | 嵌套回复 + resolve |
| Notification（通知） | ✅ | ✅ | CRUD + 已读/未读 + 摘要统计 |
| Saved Views（保存视图） | ✅ | ✅ | JSONB 筛选/排序/列配置 + 默认视图 |
| Pages（页面文档） | ✅ | ✅ | 树形层级 + 归档/恢复 + Markdown |
| Release（发布管理） | ✅ | ✅ | CRUD + Issue 关联 + 进度 |
| Estimate（估算） | ✅ | ✅ | Points/Categories/Time 三种模式 |
| Attachment（附件） | ✅ | ✅ | 文件上传 + 元数据 |
| Relation（关联类型） | ✅ | ✅ | 自定义 in/out 命名 + Issue 关联 |
| RQL（查询语言） | ✅ | ✅ | 自定义 DSL 搜索 |
| AI（智能助手） | ✅ | ✅ | Chat SSE + NL Search + Smart Create + Analyze + Page AI + Triage |
| Time Tracking（工时） | ✅ | ✅ | Start/Stop/List/Summary |
| Recurring（重复工作项） | ✅ | ✅ | daily/weekly/monthly/cron + UI 配置 |
| Intake & Triage（接收分诊） | ✅ | ✅ | 公开提交 + Accept/Reject + AI 分析 |
| Conditional Fields（条件字段） | ✅ | ✅ | 字段显隐规则 |
| Cover Image（封面图） | ✅ | ✅ | 工作项封面 |
| Command Palette（⌘K） | ✅ | ✅ | 键盘快速导航 |
| AI Settings（AI 配置） | ✅ | ✅ | Provider/Model/APIKey UI 配置 |

---

## 未来可扩展

| 模块 | 说明 |
|------|------|
| Calendar/Gantt View | 日历/甘特图视图 |
| Webhook/Integrations | GitHub/GitLab/Slack |
| RBAC 自定义角色 | 基于 Permission 的权限方案 |
| Multi-language | 多语言国际化 |
