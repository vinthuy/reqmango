# Tech Stack（技术栈）

**最后更新**: 2026-10-25

> 本文只记录**选型**和**大版本**。具体补丁版本以 `backend/go.mod` 和 `frontend/package.json` 为准，
> 不在此处抄写——手抄的版本号是文档过期最快的部分。

---

## 运行时

| 运行时 | 版本 | 依据 |
|--------|------|------|
| Go | 1.25 | `backend/go.mod` 的 `go` 指令 |
| Node.js | 24（LTS） | `frontend/package.json` 的 `engines`、`.nvmrc` |
| PostgreSQL | 16+ | 生产与 CI 均使用 16 |
| 浏览器 | 支持 ES2022 的现代浏览器 | Vite 构建目标 |

---

## Go 后端（主力）

| 组件 | 选型 | 用途 |
|------|------|------|
| HTTP 框架 | Gin | 路由、中间件、请求处理 |
| ORM | GORM | 数据库映射、迁移、查询 |
| 数据库驱动 | pgx (postgres) | PostgreSQL 驱动 |
| 数据库 | PostgreSQL | 主数据库 |
| 认证 | golang-jwt/jwt v5 | JWT token 签发和验证 |
| 密码 | bcrypt | 密码哈希 |
| 配置 | Viper | 多源配置（环境变量、YAML） |
| HTML 清洗 | bluemonday | HTML sanitization |
| JSON 序列化 | bytedance/sonic | 高性能 JSON 编解码 |

---

## 前端

| 组件 | 选型 | 用途 |
|------|------|------|
| 框架 | Vue 3 | Composition API |
| 构建 | Vite 5 | 开发服务器 + 打包 |
| 语言 | TypeScript | 类型安全 |
| 状态管理 | Pinia | 全局状态 |
| 路由 | Vue Router 4 | SPA 路由 |
| CSS | Tailwind CSS | 原子化样式 |
| HTTP | Axios | API 调用 |
| 富文本编辑器 | TipTap (vue-3) | 页面文档/Wiki 编辑 |
| 拖拽 | vue-draggable-plus | 拖拽排序 |

---

## 测试与工具链

| 层 | 选型 |
|----|------|
| Go 单测 | 标准库 `testing` + `go-sqlmock` |
| 前端单测 | Vitest |
| E2E | Playwright（`frontend/e2e`） |
| Lint | golangci-lint、ESLint、`vue-tsc` |
| 镜像构建 | 多阶段 Docker（Go/Node builder + Alpine 运行时） |

---

## Python 后端（⚠️ 已完全弃用）

> ⚠️ 该后端已完全弃用，代码保留在 `superseded/` 目录中作为参考

| 组件 | 选型 | 用途 |
|------|------|------|
| 框架 | FastAPI | 异步 Web 框架 |
| ORM | SQLAlchemy 2.0 | 异步数据库操作 |
| 验证 | Pydantic V2 | 数据验证和序列化 |
| 数据库 | SQLite / PostgreSQL | 开发/生产数据库 |
| 认证 | python-jose | JWT |

---

## 选型理由

### 为什么从 Python 迁移到 Go？

1. **性能**: Go 编译为原生二进制，单机吞吐量远超 Python 异步
2. **部署**: 单一静态二进制，无需虚拟环境
3. **类型安全**: 编译时类型检查，减少运行时错误
4. **并发**: goroutine 模型天然适合 IO 密集型 API

### 为什么保留 Vue 3 前端？

前后端通过 REST API 解耦，前端框架无需随后端语言变更而替换。

---

## 🌐 语言

- **中文** (本文档)
- [English](https://vinthuy.github.io/reqmango/en/kb/architecture/tech-stack/)
