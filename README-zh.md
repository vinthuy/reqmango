# Reqmango

**自建项目管理：新需求先分诊，再进待办。**  
提交后先给出类型、优先级与疑似重复建议，拿不准的才由人决定。

[English](README.md) · [贡献指南](CONTRIBUTING-zh.md) · [GitCode 国内主场](https://gitcode.com/yongfeng9m-/reqmanpy) · [GitHub 国际镜像](https://github.com/vinthuy/reqmango)

---

## 一条命令试用

```bash
# 国内推荐 GitCode 主场
git clone https://gitcode.com/yongfeng9m-/reqmanpy.git reqmango
# 国际镜像: git clone https://github.com/vinthuy/reqmango.git
cd reqmango
cp .env.example .env
docker compose up --build
```

**首次启动：** `docker compose up --build` 会在本地构建后端和前端镜像，
耗时可能达到几分钟，具体取决于网络和机器性能。请等待服务启动完成后再打开
页面；构建期间页面暂时无法打开或空白，并不代表启动失败。

浏览器打开 **http://localhost**，登录：

| | |
|---|---|
| 邮箱 | `demo@example.com` |
| 密码 | `demo1234` |

可选 AI（Intake 分诊 / 分析 / 标签建议）：在 `.env` 中设置 `AI_API_KEY` 后重启。

![Intake 提交 → 分诊队列](docs/assets/demo.gif)

### 一分钟走查

1. 使用演示账号登录。
2. 打开 **Demo** 项目 → **项目设置** → **请求分诊（Intake）**。
3. 打开 Intake 表单链接，提交一条简短需求（例如「手机端登录偶尔失败」）。
4. 回到分诊队列：接受 / 拒绝；配置了 `AI_API_KEY` 时可对工作项做 AI 分析与标签建议。

---

## 能力一览

| 方向 | 能力 |
|------|------|
| **工作流里的 AI** | Intake 分诊、Issue 分析、标签建议、Cycle / Sprint 总结、Ask / Build Copilot（写入前先预览） |
| 工作项 | 列表 + 看板、状态流转、最多 6 层层级 |
| 自定义字段与模板 | 文本 / 数字 / 下拉 / 布尔 / 日期 / 成员 / URL；工作区类型蓝图；项目模板 |
| 工作流与自动化 | 状态转换、审批、角色限制；触发器 → 条件 → 动作 |
| 关联与搜索 | Blocks / Relates / Duplicates；RQL 与多字段筛选 |
| 通知与安全 | 未读计数；XSS 清洗（bluemonday）；JWT + RBAC |
| API | 100+ REST 接口 |

AI 是项目管理上的一层能力，默认叙事不是独立的 Agent 控制台。代码里可能仍有高级 Agent 路由；日常路径是 Issue、Cycle、Intake 与 Pages。

---

## 环境要求

| 方式 | 需要 |
|------|------|
| **Docker（推荐）** | Docker + Docker Compose |
| 本地开发 | Go **1.25+**、PostgreSQL **16+**、Node.js **20+** |

版本与 CI、`docker-compose.yml` 一致（`postgres:16`，Go 1.25 构建镜像）。

---

## 本地开发（不用 Docker）

```bash
# 数据库
psql -U postgres -c "CREATE DATABASE reqmango;"

# 后端
cd backend
cat > .env << EOF
DATABASE_URL=postgres://postgres:postgres@localhost:5432/reqmango?sslmode=disable
SECRET_KEY=$(openssl rand -hex 32)
ACCESS_TOKEN_EXPIRE_MINUTES=10080
PORT=8000
DEBUG=true
SEED_DEMO=true
EOF
go run ./cmd/server/
```

服务启动时会自动迁移并写入种子数据，因此新建环境只需这一步。`SEED_DEMO=true` 会创建演示账号；不需要演示数据时删掉这一行，首次启动会创建一个管理员并把随机密码打印在日志里（见[生产部署](docs/deploy/production.md)）。

### 数据库迁移

若希望把建表作为一个独立、可审阅的步骤，用 `cmd/migrate`：

```bash
cd backend
go run ./cmd/migrate up      # 应用基线
go run ./cmd/migrate version # 1, dirty=false
```

`migrations/000001_baseline.up.sql` 是单一幂等基线，空库和「已由服务 AutoMigrate
建好的库」都能直接跑。压缩基线之前建好的旧库需要一次性
`go run ./cmd/migrate force 1`，详见[债务说明](docs/dev/debt/2026-10-18-sql-migration-chain.md)。

种子数据含 `demo@example.com` / `demo1234`、工作区 `reqmango-dev`（项目 `CORE` / `MOBILE` / `AI` / `OPENAPI`），以及示例 Sprint / 模块 / 工作项。演示账号的密码是公开的，**只适合本地试用**。

```bash
# 前端（API 在 :8000）
cd frontend
npm install
npm run dev
```

打开 **http://localhost:5173**。

### 本地生产构建

```bash
cd backend && go build -o server ./cmd/server/
cd frontend && npm run build
```

---

## 目录结构

```
reqmango/
├── backend/     # Go + Gin + GORM API
├── frontend/    # Vue 3 + TypeScript + Vite
├── sdk/         # MCP / CLI（可选，非默认产品叙事）
├── docs/        # API、架构、产品规格
└── docker-compose.yml
```

---

## 文档

- [API 参考](docs/API.md)
- [架构文档](docs/kb/architecture/) — [技术栈](docs/kb/architecture/tech-stack.md)、[Go 后端](docs/kb/architecture/backend-go.md)、[前端](docs/kb/architecture/frontend.md)、[数据模型](docs/kb/architecture/data-model.md)
- [生产部署](docs/deploy/production.md) — 密钥、关闭演示数据、TLS、备份

---

## 贡献

欢迎提交 Issue 与 Pull Request，尤其是小而清晰的改动。

- [贡献指南](CONTRIBUTING-zh.md) · [English](CONTRIBUTING.md)
- [适合新手的任务清单](docs/dev/good-first-issues.md)
- [安全策略](SECURITY.md) — 漏洞请走私密渠道，不要开公开 Issue
- **国内共建主场：** [GitCode](https://gitcode.com/yongfeng9m-/reqmanpy) · **国际镜像：** [GitHub](https://github.com/vinthuy/reqmango) — 同一套源码与产品叙事（[双远程说明](docs/dev/dual-remote.md)）

## License

[MIT](LICENSE)
