# reqmango

现代化的项目管理平台，支持工作项管理、自定义字段、类型模板、工作流和自动化。

---

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.21+ + Gin + GORM |
| 数据库 | PostgreSQL 16+ |
| 前端 | Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS |
| 认证 | JWT (golang-jwt/v5) |

## 快速开始

### 前置条件

- Go 1.21+
- PostgreSQL 16+
- Node.js 18+

### 1. 克隆项目

```bash
git clone https://gitcode.com/yongfeng9m-/reqmango.git
cd reqmango
```

### 2. 配置数据库

```bash
# 创建数据库
psql -U postgres -c "CREATE DATABASE reqmango;"
```

### 3. 配置后端

```bash
cd backend

# 创建环境配置文件
cat > .env << EOF
DATABASE_URL=postgres://postgres:postgres@localhost:5432/reqmango?sslmode=disable
SECRET_KEY=change-me-in-production-use-a-long-random-string
ACCESS_TOKEN_EXPIRE_MINUTES=10080
PORT=8000
DEBUG=true
EOF

# 启动后端 (自动建表 + 种子数据)
go run ./cmd/server/
```

后端启动后自动创建数据库表并插入演示数据：
- 管理员账号: `demo@example.com` / `demo1234`
- 测试账号: `demo1@reqman.local` ~ `demo19@reqman.local` (密码同)
- Demo Workspace (slug: demo) + Demo Project (identifier: DEMO)
- 6 个默认状态, 4 个 Sprint, 5 个模块, 100 个工作项
- 3 个默认工作项类型 (Bug/Feature/Epic)
- 3 个自定义字段 (优先级/截止日期/版本号)

### 4. 配置前端

```bash
cd frontend
npm install
npm run dev
```

浏览器打开 `http://localhost:5173`，使用 `demo@example.com` / `demo1234` 登录。

### 5. 生产构建

```bash
# 后端
cd backend && go build -o server ./cmd/server/

# 前端
cd frontend && npm run build
```

---

## 项目结构

```
reqmango/
├── backend/              # Go 后端
│   ├── cmd/server/          # 入口
│   ├── internal/
│   │   ├── model/           # GORM 数据模型（34 文件）
│   │   ├── dto/             # 请求/响应 DTO（45 文件）
│   │   ├── service/         # 业务逻辑（35 文件）
│   │   ├── handler/         # HTTP 处理器（37 文件）
│   │   ├── rql/             # RQL 查询语言引擎
│   │   ├── middleware/      # 中间件（Auth/CORS/Lang/Log/RateLimit）
│   │   ├── i18n/            # 国际化 (en/zh)
│   │   ├── seed/            # 种子数据
│   │   ├── common/          # 公共工具
│   │   └── config/          # 配置加载
│   └── config/              # YAML 配置
├── mcp-server/              # MCP Server
├── frontend/                # Vue 3 前端
│   └── src/
│       ├── api/             # API 调用（35 模块）
│       ├── types/           # TypeScript 类型
│       ├── stores/          # Pinia 状态管理
│       ├── views/           # 页面
│       ├── components/      # 组件
│       └── router/          # 路由
└── docs/                    # 文档
    ├── kb/                  # 知识库 (架构文档)
    ├── dev/                 # 开发管线
    └── superseded/          # 历史归档
```

## 核心功能

| 功能 | 说明 |
|------|------|
| 工作项管理 | CRUD + 状态流转 + 列表/看板视图 |
| 自定义字段 | 7种类型 (text/number/dropdown/boolean/date/member/url) |
| 类型模板 | 工作空间级类型蓝图 + 层级定义 + 字段绑定 |
| 项目模板 | 打包类型模板，一键应用到项目 |
| 工作流 | 状态转换规则 + 审批 + 角色限制 |
| 自动化 | 触发器→条件→动作规则引擎 |
| 关联关系 | 自定义关系类型 (Blocks/Relates/Duplicates) |
| 层级系统 | 最多6层工作项层级 + 类型校验 |
| 高级搜索 | 多字段 AND 组合筛选 |
| API | 90+ RESTful 端点 |

## API 文档

详见 [docs/kb/architecture/](docs/kb/architecture/) 目录下的架构文档。

## 架构文档

- [技术栈](docs/kb/architecture/tech-stack.md)
- [Go 后端架构](docs/kb/architecture/backend.md)
- [前端架构](docs/kb/architecture/frontend.md)
- [数据模型](docs/kb/architecture/data-model.md)
- [API 约定](docs/kb/architecture/api-conventions.md)
- [类型层级 & 模板设计](docs/kb/architecture/type-hierarchy-template-design.md)
- [关联关系设计](docs/kb/architecture/relation-system-design.md)

## 贡献

欢迎提交 Issue 和 Pull Request。

## License

MIT
