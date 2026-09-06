# Tech Stack（技术栈）

**最后更新**: 2026-08-30

---

## 当前技术栈

### Go 后端（主力）

| 组件 | 选型 | 版本 | 用途 |
|------|------|------|------|
| HTTP 框架 | Gin | v1.x | 路由、中间件、请求处理 |
| ORM | GORM | v1.30.0 | 数据库映射、迁移、查询 |
| 数据库驱动 | pgx (postgres) | — | PostgreSQL 驱动 |
| 数据库 | PostgreSQL | 16+ | 主数据库 |
| 认证 | golang-jwt/jwt | v5.2.1 | JWT token 签发和验证 |
| 密码 | bcrypt | — | 密码哈希 |
| 配置 | Viper | v1.19.0 | 多源配置（环境变量、YAML） |
| HTML 清洗 | bluemonday | v1.0.27 | HTML sanitization |
| JSON 序列化 | bytedance/sonic | — | 高性能 JSON 编解码 |

### 前端

| 组件 | 选型 | 版本 | 用途 |
|------|------|------|------|
| 框架 | Vue 3 | ^3.4.21 | Composition API |
| 构建 | Vite | — | 开发服务器 + 打包 |
| 语言 | TypeScript | — | 类型安全 |
| 状态管理 | Pinia | — | 全局状态 |
| 路由 | Vue Router 4 | — | SPA 路由 |
| CSS | Tailwind CSS | — | 原子化样式 |
| HTTP | Axios | — | API 调用 |
| 富文本编辑器 | TipTap (vue-3) | ^3.27.1 | 页面文档/Wiki 编辑 |
| 拖拽 | vue-draggable-plus | ^0.6.1 | 拖拽排序 |

### Python 后端（⚠️ 已完全弃用）

> ⚠️ 该后端已完全弃用，代码保留在 superseded/ 目录中作为参考

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
- [English](tech-stack.en.md)
