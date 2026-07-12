# Sprint 3: 功能 + 发布 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wiki 协作前端上线、SSO 可用、E2E 全通过、v1.0 公开发布

**Architecture:** vinthuy 做 SSO + GitHub 公开 + Docker 最终化 + Release，A 做 Wiki 前端 + 投票 + Intake + Landing + 文档站，B 做 E2E + 压测 + Bug 收敛 + SDK + 安全报告。所有工作在 W7 集中开发，W8 集中打磨和发布。

**Prerequisites:** Sprint 2 完成（Wiki 后端就绪、UI 组件库统一、安全加固完成）

**Design Spec:** `docs/superpowers/specs/2026-07-03-productization-design.md`

---

## Track 1: vinthuy — 统筹 + 发布

### Task T3.1: SSO OIDC 最小实现 (3 天)

**Files:**
- Create: `backend/internal/handler/sso_handler.go`
- Create: `backend/internal/service/sso_service.go`
- Modify: `backend/internal/model/user.go` (添加 sso_provider, sso_id 字段)
- Modify: `backend/internal/router/router.go`
- Modify: `frontend/src/views/LoginView.vue`
- Modify: `frontend/src/views/SettingsView.vue`

- [ ] **Step 1: User 模型添加 SSO 字段**

```go
// backend/internal/model/user.go — 添加
type User struct {
    // ... existing fields
    SSOProvider string `gorm:"size:20" json:"sso_provider"` // "google" | "github" | ""
    SSOID       string `gorm:"size:100;index" json:"-"`     // external ID
    AvatarURL   string `gorm:"size:500" json:"avatar_url"`
}
```

Migration:

```sql
-- backend/migrations/000002_add_sso_fields.up.sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS sso_provider VARCHAR(20) DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS sso_id VARCHAR(100) DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500) DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_users_sso_id ON users(sso_id);
```

- [ ] **Step 2: SSO Service — OAuth2 流程**

```go
// backend/internal/service/sso_service.go
package service

import (
    "context"
    "encoding/json"
    "net/http"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/github"
    "golang.org/x/oauth2/google"
    "github.com/reqmango/backend/internal/model"
)

type SSOService struct {
    db          *gorm.DB
    authService *AuthService
    configs     map[string]*oauth2.Config
}

func NewSSOService(db *gorm.DB, authService *AuthService) *SSOService {
    return &SSOService{
        db:          db,
        authService: authService,
        configs: map[string]*oauth2.Config{
            "google": {
                ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
                ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
                RedirectURL:  os.Getenv("BASE_URL") + "/api/auth/sso/google/callback",
                Scopes:       []string{"email", "profile"},
                Endpoint:     google.Endpoint,
            },
            "github": {
                ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
                ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
                RedirectURL:  os.Getenv("BASE_URL") + "/api/auth/sso/github/callback",
                Scopes:       []string{"user:email"},
                Endpoint:     github.Endpoint,
            },
        },
    }
}

func (s *SSOService) GetAuthURL(provider string) (string, error) {
    cfg, ok := s.configs[provider]
    if !ok {
        return "", fmt.Errorf("unsupported provider: %s", provider)
    }
    // Generate random state for CSRF protection
    state := generateRandomState()
    return cfg.AuthCodeURL(state), nil
}

func (s *SSOService) HandleCallback(provider, code string) (*model.User, string, error) {
    cfg, ok := s.configs[provider]
    if !ok {
        return nil, "", fmt.Errorf("unsupported provider: %s", provider)
    }

    token, err := cfg.Exchange(context.Background(), code)
    if err != nil {
        return nil, "", fmt.Errorf("oauth exchange failed: %w", err)
    }

    // Fetch user info
    userInfo, err := s.fetchUserInfo(provider, token)
    if err != nil {
        return nil, "", err
    }

    // Find or create user
    user, err := s.findOrCreateUser(provider, userInfo)
    if err != nil {
        return nil, "", err
    }

    // Generate JWT
    jwtToken, err := s.authService.GenerateToken(user)
    return user, jwtToken, err
}

type SSOUserInfo struct {
    Email    string
    Name     string
    Avatar   string
    Provider string
    SSOID    string
}

func (s *SSOService) findOrCreateUser(provider string, info *SSOUserInfo) (*model.User, error) {
    var user model.User
    // Try to find by SSO ID
    err := s.db.Where("sso_provider = ? AND sso_id = ?", provider, info.SSOID).First(&user).Error
    if err == nil {
        return &user, nil
    }

    // Try to find by email and link
    err = s.db.Where("email = ?", info.Email).First(&user).Error
    if err == nil {
        s.db.Model(&user).Updates(map[string]interface{}{
            "sso_provider": provider,
            "sso_id":       info.SSOID,
            "avatar_url":   info.Avatar,
        })
        return &user, nil
    }

    // Create new user
    user = model.User{
        Email:       info.Email,
        Name:        info.Name,
        SSOProvider: provider,
        SSOID:       info.SSOID,
        AvatarURL:   info.Avatar,
    }
    err = s.db.Create(&user).Error
    return &user, err
}
```

- [ ] **Step 3: SSO Handler**

```go
// backend/internal/handler/sso_handler.go
func (h *SSOHandler) Login(c *gin.Context) {
    provider := c.Param("provider")
    url, err := h.service.GetAuthURL(provider)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *SSOHandler) Callback(c *gin.Context) {
    provider := c.Param("provider")
    code := c.Query("code")

    user, token, err := h.service.HandleCallback(provider, code)
    if err != nil {
        c.Redirect(http.StatusTemporaryRedirect, "/login?error=sso_failed")
        return
    }

    // Set JWT cookie or return token
    c.SetCookie("token", token, 86400, "/", "", false, true)
    c.Redirect(http.StatusTemporaryRedirect, "/")
}

// Link/unlink SSO for existing users
func (h *SSOHandler) Link(c *gin.Context) {
    provider := c.Param("provider")
    // Redirect to OAuth, then bind to current user
}

func (h *SSOHandler) Unlink(c *gin.Context) {
    provider := c.Param("provider")
    userID := c.GetUint("userID")
    h.service.UnlinkSSO(userID, provider)
    c.JSON(http.StatusOK, gin.H{"message": "unlinked"})
}
```

- [ ] **Step 4: 注册路由**

```go
// router.go
auth := r.Group("/api/auth")
auth.GET("/sso/:provider", ssoHandler.Login)
auth.GET("/sso/:provider/callback", ssoHandler.Callback)
auth.POST("/sso/:provider/link", middleware.AuthRequired(), ssoHandler.Link)
auth.DELETE("/sso/:provider/unlink", middleware.AuthRequired(), ssoHandler.Unlink)
```

- [ ] **Step 5: 前端 SSO 登录按钮**

```vue
<!-- LoginView.vue — 添加 SSO 按钮 -->
<div class="flex flex-col gap-3 mt-4">
  <button @click="ssoLogin('google')" class="sso-btn">
    <GoogleIcon class="w-5 h-5" /> Continue with Google
  </button>
  <button @click="ssoLogin('github')" class="sso-btn">
    <GithubIcon class="w-5 h-5" /> Continue with GitHub
  </button>
</div>
```

- [ ] **Step 6: 设置页 SSO 绑定管理**

```vue
<!-- SettingsView.vue — SSO 账号绑定 -->
<div>
  <h3>关联账号</h3>
  <div v-for="p in ['google', 'github']" :key="p">
    <span>{{ p }}</span>
    <button v-if="isLinked(p)" @click="unlink(p)">解绑</button>
    <button v-else @click="link(p)">绑定</button>
  </div>
</div>
```

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handler/sso_handler.go backend/internal/service/sso_service.go backend/internal/model/user.go backend/internal/router/router.go backend/migrations/ frontend/src/views/LoginView.vue frontend/src/views/SettingsView.vue
git commit -m "feat: SSO OIDC (Google/GitHub OAuth) with account link/unlink"
```

---

### Task T3.2: GitHub 公开仓库准备 (2 天)

**Files:**
- Modify: `README.md`
- Create: `CONTRIBUTING.md`
- Create: `CODE_OF_CONDUCT.md`
- Create: `SECURITY.md`
- Create: `CHANGELOG.md`
- Create: `LICENSE` (确认 MIT)

- [ ] **Step 1: 敏感信息审计**

```bash
# 扫描历史提交中的敏感信息
git log -p | grep -iE "password|secret|key|token" | grep -v "public\|example\|test\|mock\|dummy\|fake" | head -50

# 检查当前文件
grep -rn "hardcoded\|password.*=" backend/ --include="*.go" | grep -v "_test.go" | grep -v "example\|default\|dummy"
```

- [ ] **Step 2: 重写 README.md (中英双语)**

```markdown
# reqmango

> 现代化的自托管项目管理平台 — 开源的 Linear/Jira 替代品

[![CI](https://github.com/reqmango/reqmango/actions/workflows/ci.yml/badge.svg)](https://github.com/reqmango/reqmango/actions)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)

## ✨ 特性

- 📋 **工作项管理** — 支持 Issue/Epic/Task/Bug 多层级，自定义类型和字段
- 📊 **6 种视图** — 列表 / 看板 / 树形 / 日历 / 甘特图 / 工作空间总览
- 🤖 **AI 原生** — AI 聊天、搜索、创建、分诊、Sprint 规划、Agent 自动指派
- 🔄 **工作流自动化** — 可视化状态转换 + 审批 + if-this-then-that 自动化规则
- 📝 **实时协作 Wiki** — 多人实时编辑文档，版本历史 + Diff
- 🔌 **MCP Server** — 对接 Claude / Cursor / VS Code，14 个工具可用
- 🌐 **中英双语** — 完整 i18n 覆盖
- 🐳 **一键部署** — `docker compose up -d`

## 🚀 快速开始

### 前置要求
- Docker & Docker Compose
- PostgreSQL 16+ (或使用 docker-compose 自带的)

### 安装
```bash
git clone https://github.com/reqmango/reqmango.git
cd reqmango
cp .env.example .env
# 编辑 .env 填入必要配置（JWT_SECRET, AI_API_KEY 等）
docker compose up -d
# 访问 http://localhost:8080
```

### 开发
```bash
make dev         # 启动开发环境
make ci          # lint + test + build
make test        # 运行所有测试
```

## 🆚 与竞品 / Linear / Jira 的对比

| 特性 | reqmango | 竞品 | Linear | Jira |
|------|:---:|:---:|:---:|:---:|
| 开源协议 | MIT ✅ | AGPL v3 | 闭源 | 闭源 |
| 后端语言 | Go 🚀 | Python | Ruby | Java |
| AI Agent | ✅ | ✅ | ✅ | ✅ |
| 树形视图 | ✅ | ❌ | ❌ | ❌ |
| 条件字段 | ✅ | ❌ | ❌ | ❌ |
| 社区版审批工作流 | ✅ | 商业版 | N/A | 商业版 |
| MCP Server | ✅ | ✅ | ❌ | ❌ |
| 实时协作 Wiki | ✅ | ✅ | ❌ | ✅ |

## 📚 文档

- [快速开始](https://docs.reqmango.dev)
- [API 文档](/api/docs)
- [功能指南](https://docs.reqmango.dev/guide)
- [贡献指南](CONTRIBUTING.md)

## 🤝 贡献

欢迎 PR！详见 [CONTRIBUTING.md](CONTRIBUTING.md)

## 📄 许可

MIT License — 详见 [LICENSE](LICENSE)
```

- [ ] **Step 3: CONTRIBUTING.md**

```markdown
# 贡献指南

## 开发环境

1. Fork & clone
2. `make dev` 启动开发环境
3. 创建 feature branch: `git checkout -b feat/my-feature`

## 代码规范

- Go: golangci-lint 零 warning
- Vue/TS: ESLint 零 error
- 测试: 新功能需包含测试
- Commit: [Conventional Commits](https://www.conventionalcommits.org/)

## PR 流程

1. `make ci` 本地通过
2. 创建 PR，填写模板
3. CI 自动运行 lint + test
4. 至少一人 review 通过
5. Squash merge 到 master
```

- [ ] **Step 4: SECURITY.md + CHANGELOG.md**

```markdown
<!-- SECURITY.md -->
# 安全策略

## 报告漏洞
请发送邮件至 security@reqmango.dev（不要公开 issue）
我们将在 48 小时内响应。

## 支持版本
| 版本 | 支持状态 |
|------|----------|
| v1.0 | ✅ 活跃支持 |
```

```markdown
<!-- CHANGELOG.md -->
# Changelog

## v1.0.0 (2026-08-28)

### 🎉 首次公开发布

- 完整的工作项管理系统（Issue/Epic/Task/Bug）
- 6 种视图：列表/看板/树形/日历/甘特图/工作空间
- AI 功能：聊天/搜索/创建/分诊/Sprint 规划/Agent
- 实时协作 Wiki
- MCP Server (SSE/STDIO)
- SSO 登录 (Google/GitHub OAuth)
- GitHub/GitLab/Slack 集成
- Webhooks + 自动化规则
- 中英双语
- Docker 一键部署
- Go SDK + TypeScript SDK
```

- [ ] **Step 5: Commit**

```bash
git add README.md CONTRIBUTING.md CODE_OF_CONDUCT.md SECURITY.md CHANGELOG.md LICENSE
git commit -m "docs: complete open-source readiness (README, CONTRIBUTING, SECURITY, CHANGELOG)"
```

---

### Task T3.3: Docker 部署最终化 (1 天)

**Files:**
- Modify: `docker-compose.yml`
- Modify: `.env.example`
- Create: `docs/deployment.md`

- [ ] **Step 1: 健康检查 + 优雅关闭**

```yaml
# docker-compose.yml — 添加 healthcheck
services:
  backend:
    # ...
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
      interval: 15s
      timeout: 5s
      retries: 3
      start_period: 10s

  db:
    # ...
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U reqmango"]
      interval: 10s
      timeout: 5s
      retries: 5
```

```go
// backend/cmd/server/main.go — 优雅关闭
func main() {
    // ... setup

    srv := &http.Server{Addr: ":8080", Handler: router}

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal().Err(err).Msg("server failed")
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Info().Msg("shutting down...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal().Err(err).Msg("forced shutdown")
    }
    log.Info().Msg("server stopped")
}
```

- [ ] **Step 2: .env.example 完善**

```bash
# .env.example
# === Required ===
JWT_SECRET=change-me-to-a-random-string-at-least-32-chars
BASE_URL=http://localhost:8080

# === Database ===
DB_HOST=db
DB_PORT=5432
DB_USER=reqmango
DB_PASSWORD=change-me
DB_NAME=reqmango

# === AI (Optional) ===
AI_PROVIDER=deepseek  # deepseek | anthropic | openai
AI_API_KEY=           # Your API key
AI_MODEL=deepseek-chat

# === SSO (Optional) ===
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=

# === Email (Optional, for intake) ===
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
```

- [ ] **Step 3: 部署文档**

```markdown
# 部署指南

## Docker Compose (推荐)

1. `cp .env.example .env && vim .env`
2. `docker compose up -d`
3. 访问 http://localhost:8080
4. 默认管理员：admin@reqmango.dev / admin123（请立即修改）

## 手动部署

### 前置要求
- Go 1.22+
- Node.js 20+
- PostgreSQL 16+

### 步骤
1. `cd backend && go build -o reqmango-server ./cmd/server`
2. `cd frontend && npm ci && npx vite build`
3. 配置 PostgreSQL 数据库
4. 运行 `./reqmango-server`
5. 用 Nginx 反向代理前端静态文件和 API

## 升级指南

1. 备份数据库: `docker compose exec db pg_dump -U reqmango reqmango > backup.sql`
2. 拉取新版本: `git pull origin master`
3. 运行 migration: `make migrate-up`
4. 重启: `docker compose up -d --build`
```

- [ ] **Step 4: Commit**

```bash
git add docker-compose.yml .env.example docs/deployment.md backend/cmd/server/main.go
git commit -m "feat: health checks, graceful shutdown, deployment docs"
```

---

### Task T3.4: 最终 Review + Release Tag (1 天)

- [ ] **Step 1: 全量 review checklist**

```markdown
## v1.0 Release Checklist

### 代码质量
- [ ] `make ci` 全绿
- [ ] Service 测试覆盖率 ≥50%
- [ ] E2E 测试 8 个场景全通过
- [ ] golangci-lint 零新 warning
- [ ] ESLint 零新 error

### 安全
- [ ] 安全审计零高危
- [ ] 依赖漏洞扫描通过 (npm audit, go mod tidy)
- [ ] Docker 镜像扫描通过 (trivy)

### 功能
- [ ] Wiki 两人实时协作可用
- [ ] SSO Google/GitHub 可登录
- [ ] Onboarding 流程可完整走通
- [ ] i18n 中英切换无异常

### 文档
- [ ] README/CONTRIBUTING/SECURITY/CHANGELOG 齐全
- [ ] Swagger /api/docs 可访问
- [ ] Landing Page 上线
- [ ] 文档站上线
- [ ] 部署文档可执行

### 发布
- [ ] GitHub 仓库公开
- [ ] Git tag v1.0.0
- [ ] GitHub Release notes
```

- [ ] **Step 2: 打 tag + 发布**

```bash
# 确认所有检查通过后
git tag -a v1.0.0 -m "v1.0.0 - Initial public release

Features:
- Full issue management with 6 views
- Real-time collaborative Wiki
- AI Chat/Search/Triage/Agent
- SSO (Google/GitHub)
- MCP Server
- GitHub/GitLab/Slack integration
- Go SDK + TypeScript SDK
- Docker one-click deployment
"

git push origin master --tags
```

- [ ] **Step 3: 创建 GitHub Release**

在 GitHub Releases 页面创建 v1.0.0，粘贴 CHANGELOG.md 内容，上传二进制文件。

---

## Track 2: A — 产品 + 生态

### Task A3.1: 实时协作 Wiki 前端 (4 天)

**Files:**
- Create: `frontend/src/composables/useWebSocket.ts`
- Create: `frontend/src/components/editor/CollaborativeEditor.vue`
- Modify: `frontend/src/views/PageDetailView.vue`
- Modify: `frontend/src/components/editor/TiptapEditor.vue`

- [ ] **Step 1: useWebSocket composable**

```typescript
// frontend/src/composables/useWebSocket.ts
import { ref, onUnmounted } from 'vue'

export function useWebSocket(pageId: Ref<number>, token: string) {
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  const users = ref<UserCursor[]>([])
  const reconnectAttempt = ref(0)
  const maxReconnectDelay = 30000

  function connect() {
    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = `${protocol}//${location.host}/ws/page/${pageId.value}?token=${token}`

    ws.value = new WebSocket(url)

    ws.value.onopen = () => {
      connected.value = true
      reconnectAttempt.value = 0
    }

    ws.value.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      handleMessage(msg)
    }

    ws.value.onclose = () => {
      connected.value = false
      // Exponential backoff reconnect
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempt.value), maxReconnectDelay)
      reconnectAttempt.value++
      setTimeout(connect, delay)
    }
  }

  function handleMessage(msg: WSMessage) {
    switch (msg.type) {
      case 'join':
        users.value.push(msg.data)
        break
      case 'leave':
        users.value = users.value.filter(u => u.userId !== msg.data.userId)
        break
      case 'operation':
        // Apply remote operation to editor
        applyRemoteOperation(msg.data)
        break
      case 'cursor':
        // Update remote cursor position
        updateRemoteCursor(msg.data)
        break
      case 'sync':
        // Full document sync on reconnect
        syncDocument(msg.data)
        break
    }
  }

  function send(type: string, data: any) {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type, data }))
    }
  }

  function sendOperation(op: OTOperation) {
    send('operation', op)
  }

  function sendCursor(pos: number) {
    send('cursor', { position: pos })
  }

  onUnmounted(() => {
    ws.value?.close()
  })

  return { connected, users, sendOperation, sendCursor, connect }
}
```

- [ ] **Step 2: CollaborativeEditor 组件**

```vue
<!-- frontend/src/components/editor/CollaborativeEditor.vue -->
<template>
  <div class="collab-editor">
    <!-- Online users bar -->
    <div class="flex items-center gap-2 px-3 py-2 border-b">
      <div class="flex -space-x-2">
        <img
          v-for="user in users" :key="user.userId"
          :src="user.avatar || defaultAvatar"
          :title="user.name"
          :style="{ borderColor: user.color }"
          class="w-7 h-7 rounded-full border-2"
        />
      </div>
      <span class="text-xs text-gray-400 ml-2">
        {{ users.length }} online
      </span>
      <div :class="['w-2 h-2 rounded-full ml-auto', connected ? 'bg-green-500' : 'bg-red-500']" />
    </div>

    <!-- Editor -->
    <TiptapEditor
      ref="editorRef"
      :content="page.content"
      :collaborative="true"
      @update="onLocalUpdate"
    />

    <!-- Version history sidebar -->
    <div v-if="showHistory" class="border-l w-80 p-4">
      <h3 class="font-medium mb-3">Version History</h3>
      <div v-for="v in versions" :key="v.id"
        class="py-2 border-b cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800"
        @click="previewVersion(v)"
      >
        <p class="text-sm">v{{ v.version }}</p>
        <p class="text-xs text-gray-400">{{ formatTime(v.created_at) }}</p>
      </div>
      <button v-if="previewingVersion" @click="restoreVersion" class="btn-primary mt-3 w-full text-sm">
        Restore this version
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
// Integration of useWebSocket with Tiptap editor
// - Send ops on local changes
// - Apply remote ops via editor.commands
// - Show remote cursors as colored carets
</script>
```

- [ ] **Step 3: 远程光标显示**

```css
/* 远程光标样式 */
.remote-cursor {
  position: absolute;
  border-left: 2px solid var(--cursor-color);
  height: 1.2em;
}

.remote-cursor-label {
  position: absolute;
  top: -1.2em;
  left: 0;
  font-size: 10px;
  color: white;
  background: var(--cursor-color);
  padding: 1px 4px;
  border-radius: 3px;
  white-space: nowrap;
}
```

- [ ] **Step 4: 版本历史 Diff 视图**

```vue
<!-- 使用 diff-match-patch 或简单的高亮 -->
<div class="diff-view font-mono text-sm">
  <div v-for="line in diffLines" :key="line.index"
    :class="[
      'px-2 py-0.5',
      line.type === 'add' ? 'bg-green-100 dark:bg-green-900/30 text-green-800' :
      line.type === 'remove' ? 'bg-red-100 dark:bg-red-900/30 text-red-800' : ''
    ]"
  >
    <span class="w-8 inline-block text-gray-400">{{ line.type === 'add' ? '+' : line.type === 'remove' ? '-' : ' ' }}</span>
    {{ line.text }}
  </div>
</div>
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/composables/useWebSocket.ts frontend/src/components/editor/CollaborativeEditor.vue frontend/src/views/PageDetailView.vue
git commit -m "feat: real-time collaborative Wiki frontend — cursors, sync, version history diff"
```

---

### Task A3.2: 投票功能 (1 天)

**Files:**
- Create: `backend/internal/model/vote.go`
- Create: `backend/internal/handler/vote_handler.go`
- Create: `backend/internal/service/vote_service.go`
- Modify: `frontend/src/views/IssueDetailView.vue`

- [ ] **Step 1: Vote 模型 + API**

```go
// backend/internal/model/vote.go
type Vote struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    IssueID   uint      `gorm:"index:idx_vote_issue_user,unique;not null" json:"issue_id"`
    UserID    uint      `gorm:"index:idx_vote_issue_user,unique;not null" json:"user_id"`
    Value     int       `gorm:"default:1" json:"value"` // +1 or -1
    CreatedAt time.Time `json:"created_at"`
}

// POST   /api/issues/:id/vote     — upvote / toggle
// DELETE /api/issues/:id/vote     — remove vote
// GET    /api/issues/:id/votes    — list votes + count
```

```go
// backend/internal/service/vote_service.go
func (s *VoteService) ToggleVote(issueID, userID uint) (int, error) {
    var existing model.Vote
    err := s.db.Where("issue_id = ? AND user_id = ?", issueID, userID).First(&existing).Error
    if err == nil {
        s.db.Delete(&existing)
    } else {
        s.db.Create(&model.Vote{IssueID: issueID, UserID: userID, Value: 1})
    }
    // Return updated count
    var count int64
    s.db.Model(&model.Vote{}).Where("issue_id = ?", issueID).Count(&count)
    return int(count), nil
}
```

- [ ] **Step 2: 前端投票 UI**

```vue
<!-- IssueDetailView.vue — 投票区域 -->
<div class="flex items-center gap-1">
  <button @click="toggleVote"
    :class="['vote-btn', hasVoted && 'voted']"
  >
    <ChevronUpIcon class="w-5 h-5" />
  </button>
  <span class="text-sm font-medium">{{ voteCount }}</span>
</div>
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/model/vote.go backend/internal/handler/vote_handler.go backend/internal/service/vote_service.go frontend/src/views/IssueDetailView.vue
git commit -m "feat: issue voting with toggle upvote"
```

---

### Task A3.3: 多渠道 Intake 基础 (2 天)

**Files:**
- Create: `backend/internal/handler/intake_email_handler.go`
- Modify: `frontend/src/views/IntakeFormView.vue`

- [ ] **Step 1: 邮件→Issue Webhook 接收器**

```go
// backend/internal/handler/intake_email_handler.go
// 接收 Mailgun/SendGrid inbound webhook
// 解析: 发件人→reporter, 主题→title, 正文→description

func (h *IntakeHandler) EmailWebhook(c *gin.Context) {
    var payload MailgunInboundPayload
    c.ShouldBind(&payload)

    // Find or create user by email
    user, _ := h.userService.FindOrCreateByEmail(payload.From)

    issue := &model.Issue{
        Title:       payload.Subject,
        Description: payload.BodyPlain,
        ProjectID:   h.defaultProjectID, // 从配置或路由参数获取
        CreatedBy:   user.ID,
        Source:      "email",
    }
    h.issueService.Create(issue)

    c.JSON(http.StatusOK, gin.H{"message": "issue created", "id": issue.ID})
}
```

- [ ] **Step 2: Intake 表单增强**

```vue
<!-- IntakeFormView.vue — 增加自定义字段 + 提交确认 -->
<template>
  <div class="max-w-lg mx-auto py-12">
    <h1>提交反馈</h1>
    <form @submit.prevent="submit">
      <AppInput v-model="form.title" label="标题" required />
      <TiptapEditor v-model="form.description" label="详细描述" />
      <!-- 自定义字段 -->
      <AppSelect v-for="field in customFields" :key="field.id"
        v-model="form.fields[field.id]"
        :label="field.name"
        :options="field.options"
      />
      <AppButton type="submit" :loading="submitting">提交</AppButton>
    </form>
    <!-- 提交成功页 -->
    <div v-if="submitted" class="text-center py-12">
      <CheckCircleIcon class="w-16 h-16 text-green-500 mx-auto" />
      <h2>提交成功！</h2>
      <p>我们会尽快处理你的反馈。追踪编号: {{ trackingId }}</p>
    </div>
  </div>
</template>
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/intake_email_handler.go frontend/src/views/IntakeFormView.vue
git commit -m "feat: email-to-issue intake + enhanced intake form with custom fields"
```

---

### Task A3.4: Landing Page (2 天)

**Files:**
- Create: `landing/index.html`
- Create: `landing/assets/style.css`
- Create: `landing/assets/script.js`

- [ ] **Step 1: 单页 Landing (独立 HTML，不依赖 Vue)**

```html
<!-- landing/index.html -->
<!DOCTYPE html>
<html lang="zh-CN" class="scroll-smooth">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>reqmango — 现代化的开源项目管理平台</title>
  <link rel="stylesheet" href="assets/style.css">
</head>
<body class="bg-white dark:bg-gray-950 text-gray-900 dark:text-gray-100">
  <!-- Hero -->
  <section class="min-h-screen flex items-center justify-center text-center px-4">
    <div>
      <h1 class="text-5xl md:text-7xl font-bold mb-6">
        项目管理，<span class="text-primary-500">AI 加持</span>
      </h1>
      <p class="text-xl text-gray-500 dark:text-gray-400 mb-10 max-w-2xl mx-auto">
        reqmango 是开源的 Linear/Jira 替代品。AI 原生、实时协作、MIT 协议、一键部署。
      </p>
      <div class="flex gap-4 justify-center">
        <a href="#quickstart" class="btn-primary px-8 py-3 rounded-lg text-lg">快速开始</a>
        <a href="https://github.com/reqmango/reqmango" class="btn-secondary px-8 py-3 rounded-lg text-lg">GitHub</a>
      </div>
    </div>
  </section>

  <!-- Features -->
  <section class="py-24 px-4 max-w-6xl mx-auto">
    <h2 class="text-3xl font-bold text-center mb-16">为什么选择 reqmango？</h2>
    <div class="grid md:grid-cols-3 gap-8">
      <div class="feature-card">
        <div class="w-12 h-12 bg-primary-100 dark:bg-primary-900 rounded-lg flex items-center justify-center mb-4">🚀</div>
        <h3>高性能 Go 后端</h3>
        <p>单二进制部署，资源占用仅为 Python 方案的 1/5。启动快、内存省。</p>
      </div>
      <div class="feature-card">
        <div class="w-12 h-12 bg-primary-100 dark:bg-primary-900 rounded-lg flex items-center justify-center mb-4">🤖</div>
        <h3>AI 原生集成</h3>
        <p>AI 聊天、智能搜索、自动分诊、Sprint 规划——你的 AI 队友已就位。</p>
      </div>
      <div class="feature-card">
        <div class="w-12 h-12 bg-primary-100 dark:bg-primary-900 rounded-lg flex items-center justify-center mb-4">🔓</div>
        <h3>MIT 开源协议</h3>
        <p>真正的开源，无 AGPL 限制。商用友好，放心 Fork。</p>
      </div>
      <!-- ... 更多 feature cards ... -->
    </div>
  </section>

  <!-- Comparison Table -->
  <section class="py-24 px-4 bg-gray-50 dark:bg-gray-900">
    <h2 class="text-3xl font-bold text-center mb-16">功能对比</h2>
    <table class="comparison-table max-w-4xl mx-auto">
      <!-- reqmango vs 竞品 vs Linear vs Jira -->
    </table>
  </section>

  <!-- Quick Start -->
  <section id="quickstart" class="py-24 px-4 max-w-2xl mx-auto text-center">
    <h2 class="text-3xl font-bold mb-6">一键部署</h2>
    <div class="bg-gray-900 text-green-400 p-6 rounded-lg text-left font-mono text-sm mb-6">
      git clone https://github.com/reqmango/reqmango.git<br>
      cd reqmango && docker compose up -d<br>
      # 打开 http://localhost:8080 🎉
    </div>
    <a href="https://docs.reqmango.dev" class="btn-primary px-6 py-3 rounded-lg">阅读文档 →</a>
  </section>

  <!-- Footer -->
  <footer class="py-12 text-center text-sm text-gray-400 border-t">
    <p>© 2026 reqmango. MIT License. Built with ❤️</p>
  </footer>

  <script src="assets/script.js"></script>
</body>
</html>
```

- [ ] **Step 2: 响应式 CSS + 暗色模式**

```css
/* landing/assets/style.css */
/* 参考 Tailwind 风格，纯手写或复制 Tailwind CDN */
```

- [ ] **Step 3: 部署到 GitHub Pages**

```bash
# 在 GitHub 仓库 Settings → Pages → Source: GitHub Actions
# 或手动推送到 gh-pages 分支
git checkout -b gh-pages
cp -r landing/* .
git add . && git commit -m "deploy landing page"
git push origin gh-pages
```

- [ ] **Step 4: Commit**

```bash
git add landing/
git commit -m "feat: landing page (hero, features, comparison, quickstart)"
```

---

### Task A3.5: 产品文档站 (2 天)

**Files:**
- Create: `docs-site/` (VitePress 项目)

- [ ] **Step 1: 初始化 VitePress**

```bash
mkdir docs-site && cd docs-site
npm init -y
npm install -D vitepress
npx vitepress init
```

- [ ] **Step 2: 编写文档内容**

```
docs-site/
├── .vitepress/config.ts
├── index.md                  # 首页
├── guide/
│   ├── getting-started.md    # 快速开始
│   ├── issues.md             # 工作项管理
│   ├── views.md              # 视图类型
│   ├── cycles.md             # 周期管理
│   ├── workflows.md          # 工作流与自动化
│   ├── custom-fields.md      # 自定义字段
│   └── releases.md           # 发布管理
├── ai/
│   ├── overview.md           # AI 功能总览
│   ├── chat.md               # AI 聊天
│   ├── search.md             # AI 搜索
│   ├── triage.md             # AI 分诊
│   └── agent.md              # AI Agent
├── integrations/
│   ├── github.md
│   ├── gitlab.md
│   ├── slack.md
│   └── mcp-server.md
├── api/
│   └── reference.md          # 链接到 Swagger
└── zh/                       # 中文版
    └── ... (同上结构)
```

- [ ] **Step 3: 配置中英双语**

```typescript
// .vitepress/config.ts
export default defineConfig({
  locales: {
    root: { label: 'English', lang: 'en-US', link: '/' },
    zh: { label: '简体中文', lang: 'zh-CN', link: '/zh/' },
  },
  themeConfig: {
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'AI', link: '/ai/overview' },
      { text: 'API', link: '/api/reference' },
    ],
  },
})
```

- [ ] **Step 4: 部署**

```bash
npx vitepress build
# 部署到 docs.reqmango.dev (GitHub Pages 或 Vercel)
```

- [ ] **Step 5: Commit**

```bash
git add docs-site/
git commit -m "docs: VitePress documentation site (zh/en bilingual)"
```

---

### Task A3.6: 交互细节打磨 (2 天)

**Files:**
- Modify: `frontend/src/` (多个文件)

- [ ] **Step 1: 页面切换过渡动画**

```vue
<!-- App.vue — 添加过渡 -->
<router-view v-slot="{ Component }">
  <transition name="page" mode="out-in">
    <component :is="Component" />
  </transition>
</router-view>

<style>
.page-enter-active, .page-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.page-enter-from { opacity: 0; transform: translateY(8px); }
.page-leave-to { opacity: 0; transform: translateY(-8px); }
</style>
```

- [ ] **Step 2: Skeleton 替换**

将 Sprint 1 创建的 LoadingSkeleton 组件替换到全站：
- IssueListView — `variant="table"` + `:rows="10"`
- IssueBoardView — `variant="card"` + `:rows="6"`
- IssueDetailView — `variant="detail"`
- CycleDetailView — `variant="detail"`
- PageDetailView — `variant="detail"`

- [ ] **Step 3: Toast 通知动画**

```vue
<transition-group name="toast" tag="div" class="toast-container">
  <div v-for="toast in toasts" :key="toast.id" class="toast">
    {{ toast.message }}
  </div>
</transition-group>

<style>
.toast-enter-active { transition: all 0.3s ease; }
.toast-leave-active { transition: all 0.2s ease; }
.toast-enter-from { opacity: 0; transform: translateX(100%); }
.toast-leave-to { opacity: 0; transform: translateX(100%); }
</style>
```

- [ ] **Step 4: 移动端触摸优化**

```vue
<!-- 侧边栏 swipe -->
<div @touchstart="onTouchStart" @touchmove="onTouchMove" @touchend="onTouchEnd">
```

```typescript
// 滑动删除
function onTouchStart(e: TouchEvent) { startX = e.touches[0].clientX }
function onTouchMove(e: TouchEvent) {
  const dx = e.touches[0].clientX - startX
  if (dx < -50) showDeleteButton.value = true
}
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/
git commit -m "style: page transitions, skeleton replacement, toast animations, touch gestures"
```

---

## Track 3: B — 质量 + 生态

### Task B3.1: E2E 测试 (Playwright) (3 天)

**Files:**
- Create: `frontend/e2e/` (或 `tests/e2e/`)
- Create: `frontend/playwright.config.ts`

- [ ] **Step 1: 安装 Playwright + 配置**

```bash
cd frontend
npm install -D @playwright/test
npx playwright install chromium
```

```typescript
// frontend/playwright.config.ts
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  retries: 2,
  use: {
    baseURL: 'http://localhost:8080',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  webServer: {
    command: 'docker compose up -d && sleep 5',
    port: 8080,
    reuseExistingServer: true,
  },
})
```

- [ ] **Step 2: 编写 8 个 E2E 场景**

```typescript
// frontend/e2e/01-register-create-issue.spec.ts
import { test, expect } from '@playwright/test'

test('register → create workspace → create project → create issue', async ({ page }) => {
  // Register
  await page.goto('/register')
  await page.fill('[name="email"]', `e2e-${Date.now()}@test.com`)
  await page.fill('[name="password"]', 'TestPass123!')
  await page.fill('[name="name"]', 'E2E User')
  await page.click('button[type="submit"]')
  await expect(page).toHaveURL('/onboarding')

  // Create workspace (onboarding step 1)
  await page.fill('[name="workspace_name"]', 'E2E Workspace')
  await page.click('button:has-text("下一步")')

  // Create project (step 2)
  await page.fill('[name="project_name"]', 'E2E Project')
  await page.fill('[name="identifier"]', 'E2E')
  await page.click('button:has-text("下一步")')

  // Create first issue (step 3)
  await page.fill('[name="title"]', 'My first issue from E2E test')
  await page.fill('[name="description"]', 'Created by automated test')
  await page.click('button:has-text("完成")')

  // Verify redirect to dashboard
  await expect(page).toHaveURL('/')
  await expect(page.locator('text=E2E Project')).toBeVisible()
})
```

```typescript
// frontend/e2e/02-issue-lifecycle.spec.ts
test('issue lifecycle: create → start → complete → close', async ({ page }) => {
  // ... login, create issue, change status through workflow
})

// frontend/e2e/03-board-drag.spec.ts
test('board drag: move card between columns + swimlane switch', async ({ page }) => {
  // ... navigate to board, drag card, verify column change
})

// frontend/e2e/04-cycle-burndown.spec.ts
test('cycle: create sprint, add issues, check burndown chart', async ({ page }) => {
  // ...
})

// frontend/e2e/05-filter-search.spec.ts
test('filter: multi-condition filter + RQL search + save view', async ({ page }) => {
  // ...
})

// frontend/e2e/06-ai-chat.spec.ts
test('AI chat: send message, stream response, create issue from chat', async ({ page }) => {
  // ... (需要 mock 或使用真实 AI key)
})

// frontend/e2e/07-notifications.spec.ts
test('notifications: SSE push, mark read, @mention', async ({ page }) => {
  // ...
})

// frontend/e2e/08-wiki-page.spec.ts
test('wiki: create page, edit content, AI generate, hierarchy nav', async ({ page }) => {
  // ...
})
```

- [ ] **Step 3: 运行 E2E**

```bash
cd frontend && npx playwright test
# Expected: 8 scenarios, all pass (或至少核心 5 个)
```

- [ ] **Step 4: CI 集成**

```yaml
# .github/workflows/ci.yml — 添加 e2e job
e2e:
  name: E2E Tests
  runs-on: ubuntu-latest
  needs: [build]
  steps:
    - uses: actions/checkout@v4
    - name: Run E2E
      run: |
        docker compose up -d
        cd frontend && npx playwright test
    - uses: actions/upload-artifact@v4
      if: failure()
      with:
        name: playwright-traces
        path: frontend/test-results/
```

- [ ] **Step 5: Commit**

```bash
git add frontend/e2e/ frontend/playwright.config.ts .github/workflows/ci.yml
git commit -m "test: E2E tests (8 scenarios) with Playwright + CI integration"
```

---

### Task B3.2: 压力测试 + 性能基线 (2 天)

**Files:**
- Create: `tests/k6/` (k6 脚本)
- Create: `docs/dev/performance-baseline-2026-08.md`

- [ ] **Step 1: 编写 k6 压测脚本**

```javascript
// tests/k6/baseline.js
import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  stages: [
    { duration: '30s', target: 50 },  // Ramp up to 50 users
    { duration: '1m', target: 100 },  // Ramp up to 100 users
    { duration: '30s', target: 200 },  // Spike to 200
    { duration: '1m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% requests <2s
    http_req_failed: ['rate<0.05'],    // <5% error rate
  },
}

const BASE_URL = 'http://localhost:8080'

export default function () {
  // Login
  const loginRes = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
    email: 'test@example.com',
    password: 'TestPass123!',
  }), { headers: { 'Content-Type': 'application/json' } })

  check(loginRes, { 'login status 200': (r) => r.status === 200 })

  const token = JSON.parse(loginRes.body).token
  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  }

  // List issues
  const listRes = http.get(`${BASE_URL}/api/projects/1/issues?limit=20`, { headers })
  check(listRes, { 'list issues 200': (r) => r.status === 200 })

  // Create issue
  const createRes = http.post(`${BASE_URL}/api/projects/1/issues`, JSON.stringify({
    title: `Load test issue ${Date.now()}`,
    description: 'Performance test',
  }), { headers })
  check(createRes, { 'create issue 201': (r) => r.status === 201 })

  sleep(1)
}
```

- [ ] **Step 2: 运行压测**

```bash
k6 run tests/k6/baseline.js --out json=perf-results.json
```

- [ ] **Step 3: 编写性能基线报告**

```markdown
# 性能基线报告 v1.0

## 测试环境
- CPU: 4 vCPU
- RAM: 8 GB
- DB: PostgreSQL 16 (2 vCPU, 4 GB)
- 并发: 50 → 100 → 200 → 0

## 结果

| 端点 | TPS | P50 | P95 | P99 | 错误率 |
|------|-----|-----|-----|-----|--------|
| POST /api/auth/login | 120 | 45ms | 180ms | 320ms | 0% |
| GET /api/projects/:id/issues | 85 | 120ms | 450ms | 890ms | 0% |
| POST /api/projects/:id/issues | 60 | 90ms | 380ms | 720ms | 0% |
| GET /api/ai/chat (stream) | 15 | 1.2s | 3.5s | 5.2s | 2% |

## 瓶颈分析
1. AI 聊天端点受限于外部 API 延迟（正常）
2. Issue 列表查询在 200 并发下 P99 接近 1s — 建议添加 Redis 缓存

## 资源消耗
- Backend CPU: max 65%, avg 40%
- Backend MEM: max 450MB, avg 280MB
- DB CPU: max 45%, avg 25%
```

- [ ] **Step 4: Commit**

```bash
git add tests/k6/ docs/dev/performance-baseline-2026-08.md
git commit -m "perf: k6 load test scripts + performance baseline report"
```

---

### Task B3.3: Bug 收敛 (3 天)

**Files:**
- Review: 架构审计报告 (Sprint 1)
- Review: UX 审计报告 (Sprint 1)
- Modify: 多个修复文件

- [ ] **Step 1: P0 bug 清零**

从 Sprint 1 的架构审计和 UX 审计中提取 P0 问题，逐个修复：

```markdown
## P0 Bug 修复清单
- [ ] [架构] [具体问题和修复]
- [ ] [UX]   [具体问题和修复]
```

- [ ] **Step 2: P1 bug 修复 ≥80%**

- [ ] **Step 3: 回归测试验证**

```bash
make ci                      # 单元/集成测试全通过
cd frontend && npx playwright test  # E2E 全通过
```

- [ ] **Step 4: KNOWN_ISSUES.md**

```markdown
# Known Issues v1.0

## P2 — 后续版本修复
1. 移动端日历视图性能优化（大量 Issue 时卡顿）
2. 离线编辑队列在某些网络条件下的数据丢失风险
3. ...

## P3 — 功能增强
1. 实时协作 Wiki 当前仅支持 2 人同时编辑（>5人需 CRDT 优化）
2. ...
```

- [ ] **Step 5: Commit**

```bash
git add . KNOWN_ISSUES.md
git commit -m "fix: P0 bugs cleared, P1 bugs >=80%, regression verified"
```

---

### Task B3.4: 安全测试报告 (1 天)

**Files:**
- Create: `docs/dev/security-test-report-2026-08.md`

- [ ] **Step 1: OWASP Top 10 自查**

```markdown
| OWASP 风险 | 状态 | 缓解措施 |
|------------|:----:|----------|
| A1: Broken Access Control | ✅ | RBAC 中间件 + handler 鉴权 |
| A2: Cryptographic Failures | ✅ | bcrypt cost=12, AES-256-GCM API keys |
| A3: Injection | ✅ | GORM 参数化查询, SQL 注入扫描已做 |
| A4: Insecure Design | ✅ | CSRF token, CSP headers |
| A5: Security Misconfiguration | ✅ | production mode, no debug endpoints |
| A6: Vulnerable Components | ⚠️ | 定期 npm audit / go mod tidy |
| A7: Auth Failures | ✅ | JWT 24h expiry, refresh rotation |
| A8: Software/Data Integrity | ✅ | HMAC-SHA256 webhooks |
| A9: Logging/Monitoring Failures | ✅ | 结构化日志 + traceID |
| A10: SSRF | ⚠️ | Webhook URL 需添加 SSRF 过滤 (P2) |
```

- [ ] **Step 2: 依赖扫描**

```bash
npm audit --production
cd backend && go mod tidy -v
docker scout quickview reqmango:latest
```

- [ ] **Step 3: Commit**

```bash
git add docs/dev/security-test-report-2026-08.md
git commit -m "docs: security test report (OWASP, dependency scan, docker scan)"
```

---

### Task B3.5: Go SDK 初版 (1.5 天)

**Files:**
- Create: `sdk/go/` 目录

- [ ] **Step 1: SDK 项目结构**

```
sdk/go/
├── go.mod
├── reqmango.go           # Client 主入口
├── issues.go             # Issue 资源
├── projects.go           # Project 资源
├── cycles.go             # Cycle 资源
├── pages.go              # Page 资源
├── types.go              # 共享类型
└── examples/
    └── basic_usage.go
```

- [ ] **Step 2: Client 实现**

```go
// sdk/go/reqmango.go
package reqmango

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    Token      string
    Issues     *IssueService
    Projects   *ProjectService
    Cycles     *CycleService
    Pages      *PageService
}

func NewClient(baseURL, token string) *Client {
    c := &Client{
        BaseURL:    baseURL,
        Token:      token,
        HTTPClient: &http.Client{Timeout: 30 * time.Second},
    }
    c.Issues = &IssueService{client: c}
    c.Projects = &ProjectService{client: c}
    c.Cycles = &CycleService{client: c}
    c.Pages = &PageService{client: c}
    return c
}

func (c *Client) do(method, path string, body, result interface{}) error {
    var reqBody []byte
    if body != nil {
        reqBody, _ = json.Marshal(body)
    }
    req, _ := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(reqBody))
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 {
        return fmt.Errorf("API error: %s", resp.Status)
    }
    if result != nil {
        return json.NewDecoder(resp.Body).Decode(result)
    }
    return nil
}
```

```go
// sdk/go/issues.go
type IssueService struct{ client *Client }

func (s *IssueService) List(projectID int, opts *ListOptions) ([]Issue, error) {
    path := fmt.Sprintf("/api/projects/%d/issues?page=%d&limit=%d", projectID, opts.Page, opts.Limit)
    var resp struct{ Data []Issue }
    err := s.client.do("GET", path, nil, &resp)
    return resp.Data, err
}

func (s *IssueService) Create(projectID int, req *CreateIssueRequest) (*Issue, error) {
    path := fmt.Sprintf("/api/projects/%d/issues", projectID)
    var issue Issue
    err := s.client.do("POST", path, req, &issue)
    return &issue, err
}

func (s *IssueService) Get(issueID int) (*Issue, error) {
    var issue Issue
    err := s.client.do("GET", fmt.Sprintf("/api/issues/%d", issueID), nil, &issue)
    return &issue, err
}

func (s *IssueService) Update(issueID int, req *UpdateIssueRequest) (*Issue, error) {
    var issue Issue
    err := s.client.do("PUT", fmt.Sprintf("/api/issues/%d", issueID), req, &issue)
    return &issue, err
}
```

- [ ] **Step 3: Example + README**

```go
// sdk/go/examples/basic_usage.go
package main

import (
    "fmt"
    "os"
    reqmango "github.com/reqmango/reqmango-go"
)

func main() {
    client := reqmango.NewClient("http://localhost:8080", os.Getenv("REQMANGO_TOKEN"))

    // List issues
    issues, err := client.Issues.List(1, &reqmango.ListOptions{Page: 1, Limit: 20})
    if err != nil { panic(err) }

    for _, issue := range issues {
        fmt.Printf("#%d %s [%s]\n", issue.ID, issue.Title, issue.State)
    }

    // Create issue
    newIssue, _ := client.Issues.Create(1, &reqmango.CreateIssueRequest{
        Title: "Hello from Go SDK!",
    })
    fmt.Printf("Created issue #%d\n", newIssue.ID)
}
```

- [ ] **Step 4: Commit**

```bash
git add sdk/go/
git commit -m "feat: Go SDK v0.1.0 — Issues, Projects, Cycles, Pages"
```

---

### Task B3.6: TypeScript SDK 初版 (1 天)

**Files:**
- Create: `sdk/typescript/` 目录

- [ ] **Step 1: TS SDK 项目**

```json
// sdk/typescript/package.json
{
  "name": "@reqmango/sdk",
  "version": "0.1.0",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "prepublish": "npm run build"
  },
  "dependencies": {
    "axios": "^1.6.0"
  }
}
```

```typescript
// sdk/typescript/src/index.ts
import axios, { AxiosInstance } from 'axios'

export class ReqmangoClient {
  private api: AxiosInstance
  public issues: IssueService
  public projects: ProjectService
  public cycles: CycleService
  public pages: PageService

  constructor(baseURL: string, token: string) {
    this.api = axios.create({
      baseURL,
      headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
      timeout: 30000,
    })
    this.issues = new IssueService(this.api)
    this.projects = new ProjectService(this.api)
    this.cycles = new CycleService(this.api)
    this.pages = new PageService(this.api)
  }
}

export interface Issue {
  id: number
  title: string
  description: string
  state: string
  priority: string
  project_id: number
  assignee_id?: number
  created_at: string
  updated_at: string
}

export interface CreateIssueRequest {
  title: string
  description?: string
  priority?: string
  type?: string
  assignee_id?: number
  parent_id?: number
}

class IssueService {
  constructor(private api: AxiosInstance) {}

  async list(projectId: number, opts?: { page?: number; limit?: number }) {
    const { data } = await this.api.get(`/api/projects/${projectId}/issues`, { params: opts })
    return data.data as Issue[]
  }

  async create(projectId: number, req: CreateIssueRequest) {
    const { data } = await this.api.post(`/api/projects/${projectId}/issues`, req)
    return data as Issue
  }

  async get(id: number) {
    const { data } = await this.api.get(`/api/issues/${id}`)
    return data as Issue
  }

  async update(id: number, req: Partial<CreateIssueRequest>) {
    const { data } = await this.api.put(`/api/issues/${id}`, req)
    return data as Issue
  }
}
// ... ProjectService, CycleService, PageService 同 pattern
```

- [ ] **Step 2: README + Example**

```typescript
// README.md
import { ReqmangoClient } from '@reqmango/sdk'

const client = new ReqmangoClient('http://localhost:8080', process.env.REQMANGO_TOKEN!)

// List issues
const issues = await client.issues.list(1, { page: 1, limit: 20 })
for (const issue of issues) {
  console.log(`#${issue.id} ${issue.title} [${issue.state}]`)
}

// Create issue
const issue = await client.issues.create(1, { title: 'Hello from TS SDK!' })
console.log(`Created #${issue.id}`)
```

- [ ] **Step 3: Commit**

```bash
git add sdk/typescript/
git commit -m "feat: TypeScript SDK v0.1.0 — Issues, Projects, Cycles, Pages"
```

---

## Sprint 3 结束 — v1.0 发布

### v1.0.0 Release Checklist

```markdown
## 代码质量
- [ ] make ci 全绿（lint + test + build）
- [ ] Service 测试覆盖率 ≥50%
- [ ] 8 个 E2E 场景全通过
- [ ] 性能基线报告完成

## 功能
- [ ] Wiki 两人实时协作可用
- [ ] SSO (Google/GitHub) 登录可用
- [ ] 投票功能可用
- [ ] 邮件→Issue 接收可用
- [ ] Onboarding 流程可走通

## 安全
- [ ] OWASP Top 10 零高危
- [ ] 依赖扫描通过
- [ ] Docker 镜像扫描通过

## 文档
- [ ] README/CONTRIBUTING/SECURITY/CHANGELOG 齐全
- [ ] Swagger /api/docs 可访问（50+ 端点）
- [ ] Landing Page 上线
- [ ] 文档站上线

## 生态
- [ ] Go SDK v0.1.0
- [ ] TypeScript SDK v0.1.0

## 发布
- [ ] GitHub 仓库公开（如尚未公开）
- [ ] git tag v1.0.0
- [ ] GitHub Release with release notes
- [ ] Docker image pushed
```

### 最终命令

```bash
# 1. Final verification
make ci
cd backend && go tool cover -func=coverage.out | grep total
cd frontend && npx playwright test
k6 run tests/k6/baseline.js

# 2. Tag and release
git tag -a v1.0.0 -m "v1.0.0 — Initial public release"
git push origin master --tags

# 3. Done 🎉
```

---

## 附录：三线依赖关系速查

```
Sprint 3 依赖:
  Wiki 前端 (A3.1) → 依赖 Wiki 后端 (T2.1, Sprint 2 已完成)
  E2E (B3.1) → 依赖所有功能完成
  SDK (B3.5, B3.6) → 依赖 API 稳定 + Swagger 完成
  Release (T3.4) → 依赖所有任务完成

Sprint 3 内部无硬阻塞关系，三线可并行推进。
W7: 集中开发（T3.1/A3.1/A3.2/A3.3/B3.1/B3.2/B3.3）
W8: 集中发布（T3.2/T3.3/T3.4/A3.4/A3.5/A3.6/B3.4/B3.5/B3.6）
```
