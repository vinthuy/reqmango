# Project Layout（项目目录结构）

**最后更新**: 2026-06-28

---

## 顶层结构

```
reqmanpy/
├── backend/              # Go 后端（主力）
│   ├── cmd/server/          # 入口 (main.go)
│   ├── internal/            # 内部代码
│   │   ├── common/          # 公共工具（错误码、常量、分页）
│   │   ├── config/          # 配置加载（Viper + 环境变量）
│   │   ├── model/           # GORM 数据模型（36 文件）
│   │   ├── dto/             # 请求/响应 DTO（47 文件）
│   │   ├── service/         # 业务逻辑（36 文件）
│   │   ├── handler/         # HTTP 处理器（38 文件）
│   │   ├── rql/             # RQL 查询语言引擎（9 文件）
│   │   ├── middleware/      # 中间件（Auth/Authz/CORS/Lang/Log/RateLimit）
│   │   ├── i18n/            # 后端国际化 (en/zh JSON)
│   │   ├── seed/            # 种子数据
│   │   └── router/          # 路由注册（80+ 端点）
│   ├── config/              # YAML 配置文件
│   ├── go.mod / go.sum      # Go 模块依赖
│   ├── Dockerfile            # 多阶段构建（distroless）
│   └── test_api.sh           # API 测试脚本
│
├── mcp-server/              # 独立 MCP Server（Go module）
│   ├── main.go              # 入口（stdio MCP）
│   ├── server.go            # JSON-RPC 2.0 协议
│   ├── client.go            # ReqManPy REST API 客户端
│   ├── tools.go             # 34 个 MCP 工具定义
│   └── go.mod               # 独立模块
│
├── frontend/                # Vue 3 前端
│   ├── src/
│   │   ├── api/             # 36 个 API 模块
│   │   ├── types/           # 22 个 TypeScript 类型文件
│   │   ├── stores/          # Pinia 状态管理（Auth/Cycle/Module）
│   │   ├── composables/     # 6 个组合式函数
│   │   ├── views/           # 18 个页面级路由组件
│   │   ├── components/      # 78 个可复用组件
│   │   ├── router/          # 20 个路由定义
│   │   └── locales/         # 前端 i18n (en-US / zh-CN)
│   ├── e2e/                 # Playwright E2E 测试（5 文件）
│   ├── tests/               # Vitest 单元测试（6 文件）
│   ├── index.html / vite.config.ts / tailwind.config.js
│   ├── nginx.conf / Dockerfile
│   └── package.json
│
├── docs/                    # 项目文档
│   ├── kb/                  # 全量知识库
│   ├── dev/                 # 增量需求开发
│   ├── superpowers/         # AI 辅助设计文档
│   └── superseded/          # 历史归档（Python 时代）
│
├── .private/                # 私有文档（不入 git）
├── docker-compose.yml       # 3 服务编排（db/backend/frontend）
├── Makefile                 # Docker 管理命令
├── LICENSE                  # MIT 许可证
└── README.md                # 项目说明
```

---

## 关键入口文件

| 文件 | 用途 |
|------|------|
| `backend/cmd/server/main.go` | Go 服务启动：初始化 DB、注册路由、启动 HTTP |
| `backend/internal/router/router.go` | 所有 API 路由定义（80+ 端点） |
| `backend/internal/middleware/auth.go` | JWT 认证中间件 |
| `backend/internal/middleware/authorization.go` | RBAC 鉴权中间件（RequirePermission + RequireRoleLevel） |
| `backend/internal/model/role.go` | 角色模型（Role + Level 枚举） |
| `backend/internal/model/permission.go` | 权限模型（Permission + RolePermission 关联） |
| `backend/internal/service/role_service.go` | RBAC 角色服务（8 方法） |
| `backend/internal/seed/seed_rbac.go` | RBAC 种子数据（55 权限 + 3 角色） |
| `backend/config/config.yaml` | 默认配置 |
| `frontend/src/main.ts` | Vue 应用入口 |
| `frontend/src/router/index.ts` | 前端 20 个路由定义 |
| `frontend/src/api/index.ts` | Axios 实例 + 拦截器 |
| `mcp-server/main.go` | MCP Server 入口（stdio 传输） |

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
| 后端框架 | Go 1.22 + Gin |
| ORM | GORM (PostgreSQL 16) |
| 鉴权 | RBAC 自定义角色 + 细粒度权限（resource:action） |
| 认证 | JWT (golang-jwt/v5) + bcrypt |
| 配置 | Viper (YAML + 环境变量) |
| 前端框架 | Vue 3 + TypeScript + Vite |
| UI | Tailwind CSS + Headless UI |
| 状态管理 | Pinia |
| 图表 | Chart.js + ECharts |
| 富文本 | Tiptap |
| MCP | 独立 Go 模块 (JSON-RPC 2.0, stdio/SSE) |
| 容器化 | Docker Compose（3 服务） + Nginx 反向代理 |
| 测试 | Go testing + Vitest + Playwright |
