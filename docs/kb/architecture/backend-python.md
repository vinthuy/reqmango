# Python Backend Architecture（Python 后端 — 遗留）

**最后更新**: 2026-06-21
**状态**: ⚠️ DEPRECATED — 正在被 Go 后端逐步替换。

---

> ⚠️ Python 后端代码保留在 `backend/` 目录中仅供部分功能参考（AI、Automation、CustomField 等尚未迁移到 Go）。新功能开发和 Bug 修复应针对 Go 后端进行。

## 项目路径

```
backend/
├── app/
│   ├── main.py                     # FastAPI 入口
│   ├── core/
│   │   ├── config.py              # Settings (pydantic-settings)
│   │   ├── security.py            # JWT + bcrypt
│   │   └── exceptions.py          # 自定义异常
│   ├── db/session.py              # 异步 SQLAlchemy session
│   ├── models/                     # SQLAlchemy ORM 模型（18 个文件）
│   ├── schemas/                    # Pydantic 验证模型（17 个文件）
│   ├── services/                   # 业务逻辑（16 个文件）
│   └── api/v1/endpoints/          # API 端点（16 个文件）
├── alembic/                        # 数据库迁移
└── requirements.txt
```

## 与 Go 后端的差异

Python 后端在以下方面比当前 Go 后端更完整：

| 功能 | Python | Go |
|------|--------|----|
| CustomField（自定义字段） | ✅ 完整 | ❌ |
| Workflow + Automation | ✅ 完整 | ❌ |
| IssueType（工作项类型） | ✅ 完整 | ❌ |
| EstimatePoint（估算点） | ✅ 完整 | ❌ |
| AI Assistant | ✅ 完整 | ❌ |
| NLP Parser | ✅ 完整 | ❌ |
| Comments（评论） | ✅ 完整 | ❌ |
| Notifications（通知） | ✅ 完整 | ❌ |
| Attachments（附件） | ✅ 完整 | ❌ |

这些功能将在后续增量开发中迁移到 Go 后端。

## 技术细节

- **框架**: FastAPI（异步，基于 Starlette + Pydantic）
- **ORM**: SQLAlchemy 2.0 异步模式
- **数据库**: 开发用 SQLite（aiosqlite），生产用 PostgreSQL（asyncpg）
- **认证**: JWT (python-jose) + bcrypt
- **API 文档**: 自动生成 OpenAPI (Swagger UI + ReDoc)

## 原始设计文档

详细架构设计见 [superseded/python-era/tech-architecture-old.md](../../superseded/python-era/tech-architecture-old.md)。
