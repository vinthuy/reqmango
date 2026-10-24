# 配置项

所有配置都来自环境变量。用 Compose 时写在仓库根目录的 `.env`（由 `docker compose` 自动读取），裸跑二进制时就是进程环境变量。默认值以代码为准：`backend/internal/config/config.go`。

!!! tip "最省事的开始"
    `cp .env.example .env` 已经能跑起本地演示。要上服务器，先看[生产部署](deploy/production.md)的清单，再回来调这里的值。

---

## 数据库

| 变量 | 默认 | 说明 |
|------|------|------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/reqmango?sslmode=disable` | 完整 DSN。设置后优先级高于下面的 `DB_*` 组合 |
| `DB_USER` | `reqmango` | Compose 数据库用户名 |
| `DB_PASSWORD` | `reqmango` | Compose 数据库密码。**公网实例必须改** |
| `DB_NAME` | `reqmango` | 数据库名 |
| `DB_PORT` | `5432` | 宿主机映射端口 |

Compose 会用 `DB_*` 拼出容器内的 `DATABASE_URL`（主机名为 `db`），所以不必手写 DSN；裸部署则需要自己给 `DATABASE_URL`。

---

## 安全与认证

| 变量 | 默认 | 说明 |
|------|------|------|
| `SECRET_KEY` | 未设置 | JWT 签名密钥。未设置（或仍是 `change-me-in-production`）时，进程启动会生成一个随机密钥并打印警告：重启即失效，必然要显式设置 |
| `ACCESS_TOKEN_EXPIRE_MINUTES` | `10080`（7 天） | 访问令牌有效期 |
| `ADMIN_EMAIL` | `admin@reqmango.com` | 空库首次启动时创建超级管理员的邮箱 |
| `ADMIN_PASSWORD` | 空 | 留空则生成 24 位随机密码，**只在日志里打印一次** |
| `SEED_DEMO` | `false` | 是否写入演示数据与演示账号（`demo@example.com` / `demo1234`） |
| `SEED_E2E` | `false` | 是否写入 E2E 测试夹具账号（仅供 CI） |

`SEED_DEMO` 和 `SEED_E2E` 的账号密码写在公开仓库里，默认关闭是有意为之。`.env.example` 里打开 `SEED_DEMO=true` 只是为了让 `docker compose up` 第一次就能登录。

引导管理员只在“库里还没有超级管理员”时运行，且不会把已注册用户提升为管理员。详见[生产部署](deploy/production.md#2-turn-the-demo-data-off)。

---

## 服务与日志

| 变量 | 默认 | 说明 |
|------|------|------|
| `PORT` | `8000` | 后端监听端口（容器内固定 8000，一般不用改） |
| `API_PORT` | `8000` | 宿主机映射给后端的端口 |
| `WEB_PORT` | `80` | 宿主机映射给前端的端口 |
| `DEBUG` | `true` | Gin 调试模式；也会影响默认 SQL 日志级别 |
| `DB_LOG_LEVEL` | 跟随 `DEBUG` | `silent` / `error` / `warn` / `info`。`info` 会把每条 SQL **连同参数**打出来，生产环境别用 |
| `SLOW_REQUEST_MS` | `1000` | 超过该耗时的请求以 `[SLOW]` 前缀记录路由模板；`0` 关闭。`/health`、`/readyz` 不记 |

日志里不会出现完整 DSN：启动时只打印前 30 个字符。

---

## 限流

| 变量 | 默认 | 说明 |
|------|------|------|
| `RATE_LIMIT_REQUESTS` | `500` | 每个客户端 IP 在窗口内允许的请求数；`0` 关闭限流 |
| `RATE_LIMIT_WINDOW_SEC` | `60` | 窗口长度（秒） |

限流键是客户端 IP（`c.ClientIP()`）。Gin 默认信任所有代理转发头，所以：
反向代理之后能正确按真实 IP 计，但**后端端口绝不能直接暴露在公网**，否则任何人伪造 `X-Forwarded-For` 就能绕过限流。生产部署请把 `8000` 绑到回环地址。

超限时返回 `429`，响应头带 `X-RateLimit-Limit` / `X-RateLimit-Remaining` / `X-RateLimit-Reset`。

---

## AI（可选）

不填 `AI_API_KEY` 时，AI 相关按钮会隐藏或给出提示，其余功能不受影响。

| 变量 | 默认 | 说明 |
|------|------|------|
| `AI_API_KEY` | 空 | 兼容 OpenAI 的服务密钥；也会读 `DEEPSEEK_API_KEY` |
| `AI_PROVIDER` | `deepseek` | 提供方标识 |
| `AI_MODEL` | `deepseek-chat` | 模型名 |
| `AI_BASE_URL` | `https://api.deepseek.com/v1` | OpenAI 兼容端点，指向自建网关也可以 |

API Key 只留在后端；浏览器不会拿到它。

---

## 镜像与发布（仅 Compose）

| 变量 | 默认 | 说明 |
|------|------|------|
| `REQMANGO_TAG` | `latest` | 镜像标签。生产环境**钉版本号**，不要漂在 `latest` |
| `REQMANGO_BACKEND_IMAGE` | `ghcr.io/vinthuy/reqmango-backend:<tag>` | 覆盖后端镜像地址（内网仓库、离线环境） |
| `REQMANGO_FRONTEND_IMAGE` | `ghcr.io/vinthuy/reqmango-frontend:<tag>` | 覆盖前端镜像地址 |

`REQMANGO_COMMIT` / `REQMANGO_BUILD_DATE` 只在从源码构建时作为镜像元数据使用，会体现在 `GET /version` 里。

---

## 一份够用的生产 `.env`

```dotenv
SEED_DEMO=false
SEED_E2E=false

SECRET_KEY=<openssl rand -hex 32 的输出>
ADMIN_EMAIL=you@example.com
ADMIN_PASSWORD=<强密码，或者留空去日志里取一次性随机密码>

DB_USER=reqmango
DB_PASSWORD=<不是 reqmango>
DB_NAME=reqmango

DEBUG=false
DB_LOG_LEVEL=warn

REQMANGO_TAG=0.1.0
```

改完后 `docker compose up -d`，用 `curl -s localhost/version` 确认跑的是哪个构建。
