# Project Layout（项目目录结构）

**最后更新**: 2026-10-25

> 只写**目录职责与边界**。文件数量变化很快，具体清单以目录为准。

---

## 顶层结构

```
reqmango/
├── backend/              # Go 后端（主力）
│   ├── cmd/server/          # 入口 (main.go)
│   ├── cmd/route_audit/     # 路由审计工具
│   ├── internal/            # 内部代码（见下方模块边界）
│   │   ├── router/          # 路由注册（全部端点在这里）
│   │   ├── middleware/      # Auth / Authz / Access / Scope / CORS / Lang / Log / Metrics / Recovery / RateLimit
│   │   ├── handler/         # HTTP 处理器
│   │   ├── service/         # 业务逻辑
│   │   ├── model/           # GORM 数据模型
│   │   ├── dto/             # request / response
│   │   ├── ai/              # LLM 客户端、harness、Agent 编排
│   │   ├── rql/             # RQL 查询语言引擎
│   │   ├── workflow/        # 状态机规则
│   │   ├── plugin/          # 插件注册表 + 钩子
│   │   ├── security/        # HTML 清洗
│   │   ├── scheduler/       # 后台定时任务
│   │   ├── i18n/            # 后端国际化 (en/zh JSON)
│   │   ├── version/         # 构建信息（GET /version）
│   │   └── seed/            # 种子数据
│   ├── go.mod / go.sum      # Go 模块依赖
│   └── Dockerfile           # 多阶段构建：golang builder → Alpine 运行时
│
├── sdk/                     # MCP server + CLI 共享模块（github.com/reqmango/tools）
│   ├── client/              #   共享 API 客户端 + DTO + 错误映射
│   ├── mcp/                 #   MCP server（mcp-go，stdio + streamable HTTP）
│   ├── cli/                 #   reqmango CLI（cobra）
│   └── cmd/                 #   两个二进制入口
│
├── frontend/                # Vue 3 前端
│   ├── src/
│   │   ├── api/             # 一个后端资源一个 API 模块
│   │   ├── types/           # TypeScript 类型（含 filters.ts）
│   │   ├── stores/          # Pinia 状态管理
│   │   ├── composables/     # 组合式函数（含 useFilters.ts）
│   │   ├── views/           # 页面级路由组件（含 agent/、agents/）
│   │   ├── components/      # 可复用组件（含 FilterBar.vue）
│   │   ├── router/          # 路由定义 + minRoleLevel 守卫
│   │   └── locales/         # 前端 i18n (en-US / zh-CN)
│   ├── e2e/                 # Playwright E2E 测试
│   ├── tests/               # Vitest 独立测试目录
│   ├── src/__tests__/       # 与源码同放的单元测试
│   ├── index.html / vite.config.ts / tailwind.config.js
│   ├── nginx.conf / Dockerfile
│   └── package.json
│
├── tests/                   # 早期的 API / E2E 场景套件（按 00-home…21-legacy-routes 分组）
├── scripts/                 # 运维与开发辅助脚本（E2E、种子数据、报告生成）
├── docs/                    # 项目文档
│   ├── kb/                  # 知识库（架构 / PRD / changelog）
│   ├── dev/                 # 增量需求开发、发布流程、运维
│   ├── superpowers/         # AI 辅助设计文档（历史，不发布）
│   └── superseded/          # 历史归档（Python 时代 / 已废止的 Agent 平台）
│
├── site/                    # mkdocs 本地构建产物（生成物，勿手改）
├── docker-compose.yml       # 3 服务编排（db/backend/frontend）
├── Makefile                 # Docker 管理命令
├── mkdocs.yml               # 文档站配置
├── LICENSE                  # MIT 许可证
└── README.md                # 项目说明
```

---

## 关键入口文件

| 文件 | 用途 |
|------|------|
| `backend/cmd/server/main.go` | Go 服务启动：初始化 DB、注册路由、启动 HTTP |
| `backend/internal/router/router.go` | 所有 API 路由定义（唯一注册入口） |
| `backend/internal/middleware/auth.go` | JWT 认证中间件 |
| `backend/internal/middleware/authorization.go` | RBAC 鉴权中间件（RequirePermission + RequireRoleLevel） |
| `backend/internal/middleware/scope.go` | 租户收敛中间件（资源归属校验） |
| `backend/internal/model/role.go` | 角色模型（Role + Level 枚举） |
| `backend/internal/model/permission.go` | 权限模型（Permission + RolePermission 关联） |
| `backend/internal/service/role_service.go` | RBAC 角色服务 |
| `backend/internal/seed/seed_rbac.go` | RBAC 种子数据（55 权限 + 3 角色） |
| `backend/internal/config/config.go` | 配置默认值（以这里为准） |
| `frontend/src/main.ts` | Vue 应用入口 |
| `frontend/src/router/index.ts` | 前端路由定义 |
| `frontend/src/api/index.ts` | Axios 实例 + 拦截器 |
| `sdk/cmd/reqmango-mcp/main.go` | MCP Server 二进制入口（stdio + streamable HTTP） |

---

## 命名规范

### Go

- 包名：小写单数（`model`, `service`, `handler`）
- 文件名：`snake_case.go`（`issue_service.go`, `ai_handler.go`）
- 导出类型/函数：`PascalCase`（`IssueService`, `FindByID`）
- 私有类型/函数：`camelCase`（`parseQuery`, `buildWhere`）
- DTO 文件：`{resource}.go`（`issue.go`, `cycle.go`）

### Vue/TypeScript

- 文件名：`kebab-case.ts` / `PascalCase.vue`（`custom-field.ts`, `IssueCard.vue`）
- 组件：`PascalCase`
- 函数/变量：`camelCase`
- Pinia Store：`use{Name}Store`
- 类型文件：按模块 `{module}.ts`

---

## 模块边界

每个功能模块在后端通常涉及以下文件：

```
internal/
├── model/{module}.go                     # DB 模型
├── dto/request/{module}.go               # 请求结构体
├── dto/response/{module}.go              # 响应结构体
├── service/{module}_service.go           # 业务逻辑
├── handler/{module}_handler.go           # HTTP 处理
└── router/router.go                      # 路由注册（修改）
```

在前端涉及：

```
src/
├── types/{module}.ts                     # TS 类型
├── api/{module}.ts                       # API 调用
├── stores/{module}.ts                    # Pinia Store
├── components/{Module}*.vue              # 组件
└── views/{Module}*.vue                   # 视图（如有独立页面）
```

---

## 技术栈总览

| 层 | 技术 |
|----|------|
| 后端框架 | Go 1.25 + Gin |
| ORM | GORM (PostgreSQL 16) |
| 鉴权 | RBAC 自定义角色 + 细粒度权限（resource:action） |
| 认证 | JWT (golang-jwt/v5) + bcrypt |
| 配置 | Viper (`.env` + 环境变量) |
| 前端框架 | Vue 3 + TypeScript + Vite 5 |
| UI | Tailwind CSS + Headless UI |
| 状态管理 | Pinia |
| 图表 | Chart.js + ECharts |
| 富文本 | Tiptap |
| MCP | sdk/mcp 模块（mcp-go；stdio + streamable HTTP） |
| 容器化 | Docker Compose（3 服务） + Nginx 反向代理 |
| 测试 | Go testing + Vitest + Playwright |

版本细节见[技术栈](tech-stack.md)。
