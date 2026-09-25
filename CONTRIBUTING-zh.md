# 参与 Reqmango 贡献

感谢你愿意共建。Reqmango 是**自建项目管理**：新需求先分诊再进待办；AI 嵌在 Issue / Intake / Cycle 里，而不是默认做成独立 Agent IDE。

[English](CONTRIBUTING.md)

## 开始之前

1. 先读 [README-zh.md](README-zh.md) 的一句话定位。
2. **小步 PR**，避免一次性大重构。
3. 先看 [Issues](https://github.com/vinthuy/reqmango/issues) 和 [good first issues 清单](docs/dev/good-first-issues.md)，避免重复开工。

## 快速环境

### Docker（推荐）

```bash
git clone https://github.com/vinthuy/reqmango.git
cd reqmango
cp .env.example .env
docker compose up --build
```

打开 http://localhost — `demo@example.com` / `demo1234`。

可选 AI：在 `.env` 设置 `AI_API_KEY` 后重启。

### 本地开发

- Go **1.25+**、PostgreSQL **16+**、Node.js **20+**
- 后端：`cd backend && go run ./cmd/server/`（`:8000`）
- 前端：`cd frontend && npm install && npm run dev`（http://localhost:5173）

账号与环境变量见 README。

## 改哪里

| 方向 | 路径 |
|------|------|
| UI / 文案 | `frontend/src/`，语言包 `frontend/src/locales/{zh-CN,en-US}.json` |
| API / 业务 | `backend/internal/` |
| 部署 | `docker-compose.yml`、前后端 Dockerfile |
| 文档 | `docs/`、根目录 README |

新增界面文案时 **中英 locale 一起改**，用 `t('namespace.key')`，不要让页面露出 `common.refresh` 这种 key。

## 流程

1. Fork，从 `master` 拉分支。
2. 只做与 Issue 相关的改动。
3. 本地验证：
   - 前端：`cd frontend && npx vitest run`（动到类型时再跑 `npx vue-tsc --noEmit`）
   - 后端：`cd backend && go test ./internal/...`
   - 手动走一遍复现路径或主流程
4. 按 PR 模板提 PR，并关联 Issue。

CI 见 `.github/workflows/ci.yml`，请保持绿色。

## 欢迎的贡献

- 有复现步骤的缺陷修复
- i18n / 文档 / Docker 与 DX
- Issue、Intake、Cycle、Pages 上的小体验改进
- 补测试

## 请先讨论再做的

- 把 Agent 控制台 / 市场做成默认主叙事的大改
- 无说明的破坏性 API 变更
- 无功能需求的依赖大升级
- 提交密钥、真实 `.env`、大体积二进制

## License

贡献内容按 [MIT License](LICENSE) 授权。
