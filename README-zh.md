# Reqmango

**自建项目管理：新需求先分诊，再进待办。**  
提交后先给出类型、优先级与疑似重复建议，拿不准的才由人决定。

[English](README.md)

---

## 一条命令试用

```bash
git clone https://github.com/vinthuy/reqmango.git
cd reqmango
cp .env.example .env
docker compose up --build
```

浏览器打开 **http://localhost**，登录：

| | |
|---|---|
| 邮箱 | `demo@example.com` |
| 密码 | `demo1234` |

可选 AI（Intake 分诊 / 分析 / 标签建议）：在 `.env` 中设置 `AI_API_KEY` 后重启。

> **演示动图：** 录完 Intake 分诊后，把 GIF 放到 [`docs/assets/demo.gif`](docs/assets/demo.gif)。在此之前请按下面步骤手动走一遍。

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
EOF
go run ./cmd/server/
```

种子数据含 `demo@example.com` / `demo1234`、工作区 `demo`、项目 `DEMO`，以及示例 Sprint / 模块 / 工作项。

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

---

## 贡献

欢迎提交 Issue 与 Pull Request。

## License

[MIT](LICENSE)
