# Project Layout（项目目录结构）

**最后更新**: 2026-06-21

---

## 顶层结构

```
reqmanpy/
├── backend-go/              # Go 后端（主力）
│   ├── cmd/server/          # 入口
│   ├── internal/            # 内部代码
│   │   ├── common/          # 公共工具（错误、常量、分页）
│   │   ├── config/          # 配置加载（Viper）
│   │   ├── model/           # GORM 数据模型
│   │   ├── dto/             # 请求/响应 DTO
│   │   ├── service/         # 业务逻辑
│   │   ├── handler/         # HTTP 处理器
│   │   ├── middleware/      # 中间件（Auth, CORS, Logger）
│   │   └── router/          # 路由注册
│   └── config/              # YAML 配置文件
│
├── backend/                 # Python 后端（遗留）
│   └── app/
│       ├── core/            # 配置、安全、异常
│       ├── db/              # 数据库 session
│       ├── models/          # SQLAlchemy 模型
│       ├── schemas/         # Pydantic 验证模型
│       ├── services/        # 业务逻辑
│       └── api/v1/endpoints/ # API 端点
│
├── frontend/                # Vue 3 前端
│   └── src/
│       ├── api/             # Axios API 调用
│       ├── types/           # TypeScript 类型
│       ├── stores/          # Pinia 状态管理
│       ├── views/           # 页面级组件
│       ├── components/      # 可复用组件
│       ├── router/          # 路由配置
│       └── composables/     # 组合式函数
│
├── docs/                    # 项目文档
│   ├── kb/                  # 全量知识库
│   ├── dev/                 # 增量需求开发
│   └── superseded/          # 历史归档
│
├── scripts/                 # 工具脚本（预留）
└── .claude/                 # Claude Code 配置
```

---

## 关键入口文件

| 文件 | 用途 |
|------|------|
| `backend-go/cmd/server/main.go` | Go 服务启动：初始化 DB、注册路由、启动 HTTP |
| `backend-go/internal/router/router.go` | 所有 API 路由定义 |
| `backend-go/internal/middleware/auth.go` | JWT 认证中间件 |
| `backend-go/config/config.yaml` | 默认配置 |
| `backend/app/main.py` | Python FastAPI 入口 |
| `frontend/src/main.ts` | Vue 应用入口 |
| `frontend/src/router/index.ts` | 前端路由定义 |
| `frontend/src/api/index.ts` | Axios 实例 + 拦截器 |

---

## 命名规范

### Go

- 包名：小写单数（`model`, `service`, `handler`）
- 文件名：`snake_case.go`（`auth_service.go`, `project_handler.go`）
- 导出类型/函数：`PascalCase`（`AuthService`, `FindByEmail`）
- 私有类型/函数：`camelCase`（`parsePagination`, `buildQuery`）
- DTO 文件：`{resource}.go`（`issue.go`）

### Vue/TypeScript

- 文件名：`kebab-case.ts` / `PascalCase.vue`（`custom-field.ts`, `IssueCard.vue`）
- 组件：`PascalCase`
- 函数/变量：`camelCase`
- Pinia Store：`use{Name}Store`
- 类型文件：按模块 `{module}.ts`

---

## 模块边界

每个功能模块在 Go 后端通常涉及以下文件：

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
