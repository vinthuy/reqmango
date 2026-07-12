# Sprint 1: 基础夯实 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 测试基础设施从 0 到 30% 覆盖率、CI/CD 跑通、UX 审计完成、Wiki 技术方案确定

**Architecture:** 三线并行——vinthuy 做架构审计+Wiki 方案+RQL 增强，A 做 UX 审计+i18n+筛选栏+状态组件+响应式，B 做测试框架+CI/CD+DB 迁移+测试第一批。B 的测试框架需在 W1 完成以便 A 后续写组件测试。三线在 W3 末交叉 review。

**Tech Stack:** Go testify + sqlmock + Vue Test Utils + GitHub Actions + golang-migrate + swaggo

**Design Spec:** `docs/superpowers/specs/2026-07-03-productization-design.md`

---

## Pre-flight: 团队环境搭建 (W1 Day 0)

### Task 0: 开发环境标准化

**Files:**
- Modify: `backend/go.mod`
- Modify: `frontend/package.json`
- Create: `.golangci.yml`
- Create: `.vscode/settings.json` (或 `.editorconfig`)

- [ ] **Step 1: 安装 Go 工具链依赖**

```bash
cd backend
go get github.com/stretchr/testify@latest
go get github.com/DATA-DOG/go-sqlmock@latest
go get github.com/golang-migrate/migrate/v4@latest
go get github.com/swaggo/swag@latest
go get github.com/swaggo/gin-swagger@latest
go get github.com/rs/zerolog@latest
go mod tidy
```

- [ ] **Step 2: 安装前端测试依赖**

```bash
cd frontend
npm install -D @vue/test-utils@latest happy-dom@latest @vitest/coverage-v8
npm install -D @playwright/test@latest
```

- [ ] **Step 3: 创建 golangci-lint 配置**

```yaml
# backend/.golangci.yml
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - gosec
    - sqlclosecheck
    - noctx
run:
  timeout: 5m
issues:
  exclude-use-default: false
```

- [ ] **Step 4: 更新 Makefile 添加开发命令**

```makefile
# 在现有 Makefile 末尾添加
.PHONY: lint lint-fix test test-backend test-frontend ci coverage dev dev-backend db-shell

lint:
	cd backend && golangci-lint run ./...
	cd frontend && npx eslint src/ --ext .ts,.vue

lint-fix:
	cd backend && golangci-lint run --fix ./...
	cd frontend && npx eslint src/ --ext .ts,.vue --fix

test-backend:
	cd backend && go test -race -coverprofile=coverage.out ./internal/...

test-frontend:
	cd frontend && npx vitest run --coverage

test: test-backend test-frontend

ci: lint test
	cd backend && go build ./cmd/server
	cd frontend && npx vite build

coverage:
	cd backend && go tool cover -html=coverage.out -o coverage.html

dev-backend:
	cd backend && go run ./cmd/server

db-shell:
	docker compose exec db psql -U reqmango -d reqmango
```

- [ ] **Step 5: 验证环境**

```bash
make lint        # 预期：有 lint 警告（现有代码问题），但 lint 工具本身正常运行
make test-backend  # 预期：现有 8 个测试通过（common/rql 包）
```

- [ ] **Step 6: Commit**

```bash
git add backend/go.mod backend/go.sum backend/.golangci.yml frontend/package.json frontend/package-lock.json Makefile
git commit -m "chore: set up dev toolchain (lint, test, coverage targets)"
```

---

## Track 1: vinthuy — 统筹 + 架构

### Task 1.1: 架构审计与债务梳理

**Files:**
- Review: `backend/internal/service/*.go` (40 files)
- Review: `backend/internal/handler/*.go` (32 files)
- Review: `backend/internal/model/*.go` (~50 files)
- Create: `docs/dev/architecture-audit-2026-07.md`

- [ ] **Step 1: 审计 service 层代码质量**

逐文件检查以下模式，记录问题：
```
检查清单（每个 service 文件）:
- [ ] 错误处理：是否吞掉错误？是否有 panic recover？
- [ ] N+1 查询：是否在循环内执行 DB 查询？
- [ ] 事务边界：多表操作是否在事务内？
- [ ] 输入校验：是否信任前端输入（应在 handler 层校验但 service 需防御）？
- [ ] 日志规范：错误是否记录了足够的上下文？
- [ ] 导出函数：是否有 godoc 注释？
```

优先审计这些高风险 service：
1. `issue_service.go` — 最复杂，涉及搜索/批量/导入导出
2. `workflow_service.go` — 审批逻辑
3. `automation_service.go` — 自动规则引擎
4. `ai_service.go` — AI 调用链路
5. `cycle_service.go` — 燃尽图计算

- [ ] **Step 2: 审计 handler 层**

检查：
- 请求参数校验是否完整（binding tags）
- 错误响应格式是否统一（错误码 + i18n key）
- 是否有 SQL 注入风险（动态 SQL 拼接）
- 鉴权中间件覆盖是否完整

- [ ] **Step 3: 审计 model 层**

检查：
- GORM tag 是否正确（外键、索引、约束）
- 是否有缺失的软删除（deleted_at）
- 是否有需要 unique index 的字段

- [ ] **Step 4: 输出审计报告**

```markdown
# 架构审计报告 2026-07

## 严重问题 (P0 - 阻塞发布)
1. [具体问题，文件:行号，修复建议]

## 重要问题 (P1 - 影响质量)
1. ...

## 改进建议 (P2 - 锦上添花)
1. ...

## 技术债务 Backlog
| ID | 问题 | 文件 | 优先级 | 预估工时 |
|----|------|------|--------|----------|
| ... |
```

- [ ] **Step 5: Commit**

```bash
git add docs/dev/architecture-audit-2026-07.md
git commit -m "docs: architecture audit report for Sprint 1 baseline"
```

---

### Task 1.2: 实时协作 Wiki 技术方案

**Files:**
- Create: `docs/dev/realtime-wiki-design.md`

- [ ] **Step 1: WebSocket 库选型验证**

创建 PoC 验证 gorilla/websocket：

```bash
mkdir -p backend/cmd/ws-poc
```

```go
// backend/cmd/ws-poc/main.go
package main

import (
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub maintains the set of active clients per room
type Hub struct {
    rooms map[string]map[*websocket.Conn]bool
    mu    sync.RWMutex
}

func (h *Hub) Join(room string, conn *websocket.Conn) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if h.rooms[room] == nil {
        h.rooms[room] = make(map[*websocket.Conn]bool)
    }
    h.rooms[room][conn] = true
}

func (h *Hub) Broadcast(room string, sender *websocket.Conn, msg []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    for conn := range h.rooms[room] {
        if conn != sender {
            conn.WriteMessage(websocket.TextMessage, msg)
        }
    }
}

func main() {
    hub := &Hub{rooms: make(map[string]map[*websocket.Conn]bool)}

    http.HandleFunc("/ws/page/", func(w http.ResponseWriter, r *http.Request) {
        conn, _ := upgrader.Upgrade(w, r, nil)
        roomID := r.URL.Path[len("/ws/page/"):]

        hub.Join(roomID, conn)
        defer conn.Close()

        for {
            _, msg, err := conn.ReadMessage()
            if err != nil {
                break
            }
            hub.Broadcast(roomID, conn, msg)
        }
    })

    log.Fatal(http.ListenAndServe(":8081", nil))
}
```

- [ ] **Step 2: 验证 PoC**

```bash
cd backend && go run ./cmd/ws-poc &
# 打开两个浏览器 tab 访问同一个 /ws/page/test，发送消息验证互传
```

- [ ] **Step 3: 编写技术方案文档**

```markdown
# 实时协作 Wiki 技术设计

## 选型决策
| 方案 | 优点 | 缺点 | 结论 |
|------|------|------|------|
| gorilla/websocket | Go 生态成熟 | 需自行管理连接 | ✅ 选择 |
| Centrifugo | 功能完整 | 额外服务依赖 | ❌ 过度 |
| SSE + POST | 简单 | 单向不适合协作 | ❌ |

## 架构设计
- Hub/Room 模式：每个 Page 一个 Room
- OT 算法最小实现：仅支持 insert/delete，不做完整 CRDT
- PageVersion 模型：content snapshot + operations diff

## 数据模型变更
Page 模型新增字段：
- version int (当前版本号)
- PageVersion 新模型：
  - id, page_id, version, content_snapshot (text), operations (jsonb), created_by, created_at

## API 设计
- GET /ws/page/:page_id — WebSocket 连接
- GET /api/pages/:id/versions — 版本历史列表
- GET /api/pages/:id/versions/:vid — 版本详情+diff
- POST /api/pages/:id/restore/:vid — 回滚到指定版本
```

- [ ] **Step 4: Commit**

```bash
git add docs/dev/realtime-wiki-design.md backend/cmd/ws-poc/
git commit -m "docs: real-time collaborative wiki technical design + PoC"
```

---

### Task 1.3: 代码规范 + Review 机制

**Files:**
- Create: `backend/.golangci.yml` (如 Task 0 已创建则修改)
- Create: `frontend/.eslintrc.cjs` (检查是否已存在)
- Create: `.github/pull_request_template.md`
- Create: `docs/dev/code-review-guide.md`

- [ ] **Step 1: 检查并完善 ESLint 配置**

```bash
ls frontend/.eslintrc* frontend/eslint.config.* 2>/dev/null
```

如果不存在或配置不完整：

```javascript
// frontend/.eslintrc.cjs
module.exports = {
  root: true,
  env: { browser: true, es2021: true, node: true },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:vue/vue3-recommended',
    'prettier',
  ],
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@typescript-eslint/parser',
    ecmaVersion: 'latest',
    sourceType: 'module',
  },
  rules: {
    'vue/multi-word-component-names': 'off',
    '@typescript-eslint/no-explicit-any': 'warn',
    '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
  },
};
```

- [ ] **Step 2: 配置 pre-commit hook**

```bash
# .git/hooks/pre-commit
#!/bin/sh
# Run lint before commit
echo "Running linters..."
cd backend && golangci-lint run --new-from-rev=HEAD~1 ./... 2>/dev/null || echo "⚠ Go lint warnings"
cd ../frontend && npx eslint src/ --ext .ts,.vue --max-warnings 10 2>/dev/null || echo "⚠ ESLint warnings"
```

- [ ] **Step 3: 创建 PR 模板**

```markdown
<!-- .github/pull_request_template.md -->
## 变更说明
<!-- 简述这个 PR 做了什么 -->

## 变更类型
- [ ] Bug 修复
- [ ] 新功能
- [ ] 重构
- [ ] 文档
- [ ] 测试

## 测试
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 手动测试通过

## Checklist
- [ ] 代码符合规范（lint 通过）
- [ ] 新增/修改的 API 已更新 Swagger 注释
- [ ] i18n 键已添加（如有新增文案）
```

- [ ] **Step 4: Commit**

```bash
git add .github/pull_request_template.md docs/dev/code-review-guide.md
git commit -m "docs: add code review guide and PR template"
```

---

### Task 1.4: RQL AI 搜索增强

**Files:**
- Modify: `backend/internal/rql/parser.go`
- Modify: `backend/internal/service/ai_service.go` (或相关 AI search handler)
- Modify: `frontend/src/composables/useRQL.ts`

- [ ] **Step 1: 分析当前 RQL 转译失败案例**

```bash
cd backend
grep -r "AI.*Search\|NL.*RQL\|NaturalLanguage" internal/service/ai_service.go
```

- [ ] **Step 2: 增强 RQL 解析器支持更多自然语言模式**

```go
// backend/internal/rql/parser.go — 新增支持的 NL 模式

// 新增模式匹配:
// "我的 issue" → assignee = me
// "上周创建的" → created_at > last_week
// "高优先级" → priority >= high
// "未开始的" → state_group = unstarted
// "最近更新的" → sort = -updated_at

func ParseNaturalLanguageHint(input string) *Query {
    hints := &Query{}
    lower := strings.ToLower(input)

    if strings.Contains(lower, "我的") || strings.Contains(lower, "my") {
        hints.Assignee = "__me__"
    }
    if strings.Contains(lower, "高优") || strings.Contains(lower, "high") {
        hints.Priority = "high,urgent"
    }
    if strings.Contains(lower, "未开始") || strings.Contains(lower, "unstarted") {
        hints.StateGroup = "unstarted"
    }
    if strings.Contains(lower, "上周") {
        hints.CreatedAfter = "last_week"
    }
    // ... 更多模式
    return hints
}
```

- [ ] **Step 3: 前端展示可编辑的 RQL**

```typescript
// frontend/src/composables/useRQL.ts — 新增函数

export function rqlToHumanReadable(rql: string): string {
  // 将 RQL 转为人类可读的筛选摘要
  // "priority>=high state_group=unstarted" → "高优先级 · 未开始"
  const parts: string[] = [];
  const parsed = parseRQL(rql);
  for (const [key, op, value] of parsed.conditions) {
    parts.push(`${FIELD_LABELS[key] || key} ${OP_LABELS[op]} ${value}`);
  }
  return parts.join(' · ') || '全部 Issue';
}
```

- [ ] **Step 4: 编写解析器测试**

```go
// backend/internal/rql/parser_test.go — 新增测试

func TestParseNaturalLanguageHint(t *testing.T) {
    tests := []struct {
        input    string
        expected *Query
    }{
        {"我的高优先级bug", &Query{Assignee: "__me__", Priority: "high,urgent"}},
        {"上周未开始的任务", &Query{CreatedAfter: "last_week", StateGroup: "unstarted"}},
    }
    for _, tt := range tests {
        got := ParseNaturalLanguageHint(tt.input)
        assert.Equal(t, tt.expected, got)
    }
}
```

- [ ] **Step 5: Run tests and commit**

```bash
cd backend && go test ./internal/rql/... -v
git add backend/internal/rql/ frontend/src/composables/useRQL.ts
git commit -m "feat: enhance RQL parser with natural language hint support"
```

---

### Task 1.5: Sprint 任务拆解 + 管理

**使用 reqmango 自身管理开发任务：**

- [ ] **Step 1: 在 reqmango 中创建 Sprint 1 Cycle**

```
创建 Cycle: "Sprint 1 - 基础夯实" (2026-07-06 ~ 2026-07-24)
```

- [ ] **Step 2: 创建 Issue 任务**

为每个子任务（T1.1-T1.6, A1.1-A1.5, B1.1-B1.6）创建 Issue，assignee 对应分配。

- [ ] **Step 3: 设置 Sprint 1 结束检查点**

```markdown
Sprint 1 验收标准:
- [ ] make ci 一键通过
- [ ] service 层覆盖率 ≥30%
- [ ] UX 审计报告完成
- [ ] CI badge: passing
- [ ] 所有 17 个子任务 Issue 已关闭
```

---

## Track 2: A — 产品体验

### Task A1.1: 全站 UX 审计

**Files:**
- Create: `docs/dev/ux-audit-2026-07.md`

- [ ] **Step 1: 逐页面截图审查**

审查这 23 个页面：
```
views/
├── IssueListView, IssueBoardView, IssueTreeView
├── IssueDetailView, IssueCreateView
├── CycleListView, CycleDetailView, CycleCreateView
├── ModuleListView, ModuleDetailView
├── ReleaseListView, ReleaseDetailView
├── PageListView, PageDetailView, PageCreateView
├── ProjectListView, ProjectDetailView, ProjectSettingsView
├── WorkspaceListView, WorkspaceSettingsView
├── DashboardView, ReportView
├── SettingsView (Profile, Notifications, API Keys)
└── IntakeFormView, TriageView
```

对每个页面标注：
- 布局/间距问题
- 空状态/加载态/错误态处理
- 交互反馈（点击、hover、过渡）
- 文案清晰度
- 性能问题（渲染闪烁、卡顿）

- [ ] **Step 2: 输出分级清单**

```markdown
# UX 审计报告

## P0 (阻塞体验)
| # | 页面 | 问题 | 截图 | 建议 |
|---|------|------|------|------|

## P1 (影响体验)  
| # | 页面 | 问题 | 截图 | 建议 |
|---|------|------|------|------|

## P2 (锦上添花)
| # | 页面 | 问题 | 截图 | 建议 |
|---|------|------|------|------|
```

- [ ] **Step 3: Commit**

```bash
git add docs/dev/ux-audit-2026-07.md
git commit -m "docs: UX audit report with P0/P1/P2 prioritization"
```

---

### Task A1.2: i18n 缺失扫描与修复

**Files:**
- Modify: `frontend/src/locales/zh.json`
- Modify: `frontend/src/locales/en.json`
- Modify: `frontend/src/views/*.vue` (23 files)
- Modify: `frontend/src/components/*.vue` (70+ files)

- [ ] **Step 1: 扫描硬编码中文**

```bash
cd frontend
# 扫描 Vue 模板中的硬编码中文
grep -rPn '>[\x{4e00}-\x{9fff}]+<' src/views/ src/components/ --include="*.vue" | head -100

# 扫描 TS 文件中的硬编码中文字符串
grep -rPn "['\"][\x{4e00}-\x{9fff}]+" src/ --include="*.ts" | grep -v "// " | grep -v "locales" | head -100
```

- [ ] **Step 2: 整理缺失的翻译键**

按页面分组输出缺失清单：

```json
// 示例输出格式
{
  "IssueListView": ["筛选条件", "共 X 条", "清除筛选"],
  "IssueBoardView": ["按状态分组", "拖拽移动", "无 Issue"],
  "CycleDetailView": ["燃尽图", "进度", "剩余工作量"],
  ...
}
```

- [ ] **Step 3: 补充 zh.json 和 en.json**

```json
// frontend/src/locales/zh.json — 补充条目
{
  "common": {
    "total_records": "共 {count} 条",
    "clear_filter": "清除筛选",
    "no_data": "暂无数据",
    "loading": "加载中...",
    "error_occurred": "出错了",
    "retry": "重试"
  },
  "filter": {
    "filter_by": "筛选条件",
    "add_filter": "添加筛选",
    "save_filter": "保存筛选",
    "my_issues": "我的 Issue",
    "assigned_to_me": "分配给我"
  },
  "board": {
    "group_by_status": "按状态分组",
    "drag_to_move": "拖拽移动",
    "no_issues": "无 Issue",
    "swimlane_by": "泳道分组"
  },
  "cycle": {
    "burndown_chart": "燃尽图",
    "progress": "进度",
    "remaining_work": "剩余工作量",
    "completed": "已完成"
  }
}
```

```json
// frontend/src/locales/en.json — 对应英文
{
  "common": {
    "total_records": "{count} records",
    "clear_filter": "Clear filters",
    "no_data": "No data",
    "loading": "Loading...",
    "error_occurred": "Something went wrong",
    "retry": "Retry"
  },
  ...
}
```

- [ ] **Step 4: 逐文件替换硬编码文本**

对每个 .vue 文件，将硬编码中文替换为 `{{ $t('key.path') }}`：

```vue
<!-- Before -->
<span>共 {{ total }} 条</span>

<!-- After -->
<span>{{ $t('common.total_records', { count: total }) }}</span>
```

- [ ] **Step 5: 验证**

```bash
cd frontend && npx vite build  # 确保构建通过
# 手动切换中/英文，检查 23 个页面
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/locales/ frontend/src/views/ frontend/src/components/
git commit -m "feat: complete i18n coverage for all 23 views and core components"
```

---

### Task A1.3: 筛选栏主流风格 Phase 2

**Files:**
- Modify: `frontend/src/components/FilterBar.vue` (843 lines)
- Create: `frontend/src/components/QuickFilterChips.vue`
- Modify: `frontend/src/composables/useRQL.ts`

- [ ] **Step 1: 创建 QuickFilterChips 组件**

```vue
<!-- frontend/src/components/QuickFilterChips.vue -->
<template>
  <div class="flex flex-wrap gap-2">
    <button
      v-for="preset in presets"
      :key="preset.key"
      @click="$emit('select', preset.key)"
      :class="[
        'px-3 py-1.5 rounded-full text-sm transition-colors',
        selected === preset.key
          ? 'bg-indigo-600 text-white'
          : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
      ]"
    >
      {{ preset.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  selected: string | null
}>()

defineEmits<{
  select: [key: string]
}>()

const presets = [
  { key: 'all', label: '全部' },
  { key: 'mine', label: '我的 Issue' },
  { key: 'assigned', label: '分配给我' },
  { key: 'recent', label: '最近更新' },
  { key: 'unstarted', label: '未开始' },
  { key: 'high_priority', label: '高优先级' },
]
</script>
```

- [ ] **Step 2: 在 FilterBar 中集成 QuickFilterChips**

在 FilterBar.vue 顶部插入：

```vue
<QuickFilterChips
  :selected="activeQuickFilter"
  @select="handleQuickFilter"
/>
```

添加逻辑：

```typescript
function handleQuickFilter(key: string) {
  const rqlMap: Record<string, string> = {
    all: '',
    mine: 'reporter=me',
    assigned: 'assignee=me',
    recent: 'sort=-updated_at',
    unstarted: 'state_group=unstarted',
    high_priority: 'priority>=high',
  }
  const rql = rqlMap[key] || ''
  activeQuickFilter.value = key
  emit('update:rql', rql)
}
```

- [ ] **Step 3: 筛选条件持久化到 URL**

```typescript
// 在 FilterBar 中 watch rql 变化 → 同步到 URL search params
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

watch(rql, (newRql) => {
  router.replace({ query: { ...route.query, rql: newRql || undefined } })
})

// 初始化时从 URL 恢复
onMounted(() => {
  if (route.query.rql) {
    rql.value = route.query.rql as string
  }
})
```

- [ ] **Step 4: 写组件测试**

```typescript
// frontend/src/components/__tests__/QuickFilterChips.spec.ts
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import QuickFilterChips from '../QuickFilterChips.vue'

describe('QuickFilterChips', () => {
  it('renders all preset filters', () => {
    const wrapper = mount(QuickFilterChips, { props: { selected: null } })
    expect(wrapper.text()).toContain('全部')
    expect(wrapper.text()).toContain('我的 Issue')
    expect(wrapper.text()).toContain('高优先级')
  })

  it('highlights selected filter', () => {
    const wrapper = mount(QuickFilterChips, { props: { selected: 'mine' } })
    const btn = wrapper.find('button.bg-indigo-600')
    expect(btn.exists()).toBe(true)
    expect(btn.text()).toContain('我的 Issue')
  })

  it('emits select event on click', async () => {
    const wrapper = mount(QuickFilterChips, { props: { selected: null } })
    await wrapper.findAll('button')[1].trigger('click')
    expect(wrapper.emitted('select')?.[0]).toEqual(['mine'])
  })
})
```

```bash
cd frontend && npx vitest run src/components/__tests__/QuickFilterChips.spec.ts
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/FilterBar.vue frontend/src/components/QuickFilterChips.vue frontend/src/components/__tests__/QuickFilterChips.spec.ts
git commit -m "feat: add QuickFilterChips + filter persistence in URL"
```

---

### Task A1.4: 状态组件标准化

**Files:**
- Create: `frontend/src/components/common/EmptyState.vue`
- Create: `frontend/src/components/common/ErrorState.vue`
- Create: `frontend/src/components/common/LoadingSkeleton.vue`
- Create: `frontend/src/components/common/__tests__/EmptyState.spec.ts`

- [ ] **Step 1: 编写 EmptyState 组件测试 (TDD)**

```typescript
// frontend/src/components/common/__tests__/EmptyState.spec.ts
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import EmptyState from '../EmptyState.vue'

describe('EmptyState', () => {
  it('renders icon, title and description', () => {
    const wrapper = mount(EmptyState, {
      props: { title: '暂无 Issue', description: '点击下方按钮创建第一个 Issue' }
    })
    expect(wrapper.text()).toContain('暂无 Issue')
    expect(wrapper.text()).toContain('点击下方按钮创建第一个 Issue')
  })

  it('renders action slot', () => {
    const wrapper = mount(EmptyState, {
      props: { title: '空' },
      slots: { action: '<button>创建</button>' }
    })
    expect(wrapper.text()).toContain('创建')
  })

  it('renders custom icon when provided', () => {
    const wrapper = mount(EmptyState, {
      props: { title: '空', icon: 'inbox' }
    })
    expect(wrapper.findComponent({ name: 'InboxIcon' }).exists()).toBe(true)
  })
})
```

- [ ] **Step 2: Run test, verify it fails**

```bash
cd frontend && npx vitest run src/components/common/__tests__/EmptyState.spec.ts
# Expected: FAIL
```

- [ ] **Step 3: 实现 EmptyState**

```vue
<!-- frontend/src/components/common/EmptyState.vue -->
<template>
  <div class="flex flex-col items-center justify-center py-16 px-4 text-center">
    <component :is="iconComponent" v-if="icon" class="w-16 h-16 text-gray-300 dark:text-gray-600 mb-4" />
    <div v-else class="w-16 h-16 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
      <svg class="w-8 h-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
      </svg>
    </div>
    <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-1">{{ title }}</h3>
    <p v-if="description" class="text-sm text-gray-500 dark:text-gray-400 max-w-sm">{{ description }}</p>
    <div v-if="$slots.action" class="mt-4">
      <slot name="action" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  title: string
  description?: string
  icon?: string
}>(), {
  icon: 'default'
})

const iconComponent = computed(() => {
  // Return icon component based on name
  return null // Simple default — extend as needed
})
</script>
```

- [ ] **Step 4: Run test, verify pass**

```bash
cd frontend && npx vitest run src/components/common/__tests__/EmptyState.spec.ts
# Expected: PASS
```

- [ ] **Step 5: 同样 TDD 实现 ErrorState 和 LoadingSkeleton**

```vue
<!-- frontend/src/components/common/ErrorState.vue -->
<template>
  <div class="flex flex-col items-center justify-center py-16 px-4 text-center">
    <div class="w-16 h-16 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center mb-4">
      <svg class="w-8 h-8 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
      </svg>
    </div>
    <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-1">{{ title }}</h3>
    <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{{ message }}</p>
    <button v-if="retryable" @click="$emit('retry')" class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700">
      {{ $t('common.retry') }}
    </button>
  </div>
</template>
```

```vue
<!-- frontend/src/components/common/LoadingSkeleton.vue -->
<template>
  <div :class="variantClass" class="animate-pulse">
    <!-- Table skeleton -->
    <template v-if="variant === 'table'">
      <div v-for="i in rows" :key="i" class="flex gap-4 py-3 border-b border-gray-100 dark:border-gray-800">
        <div v-for="j in cols" :key="j" class="h-4 bg-gray-200 dark:bg-gray-700 rounded" :style="{ width: randomWidth() }" />
      </div>
    </template>
    <!-- Card skeleton -->
    <template v-else-if="variant === 'card'">
      <div v-for="i in rows" :key="i" class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg mb-3">
        <div class="h-5 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-2" />
        <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-3" />
        <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-full" />
      </div>
    </template>
    <!-- Detail skeleton -->
    <template v-else-if="variant === 'detail'">
      <div class="h-8 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-6" />
      <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full mb-3" v-for="i in 5" :key="i" />
    </template>
  </div>
</template>
```

- [ ] **Step 6: 逐个替换高频页面的状态处理**

优先替换：
1. IssueListView — 空列表用 EmptyState 替换 "No issues found" div
2. IssueBoardView — 空列用 EmptyState
3. IssueDetailView — 加载中用 LoadingSkeleton(detail)
4. CycleDetailView — 错误时用 ErrorState

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/common/
git commit -m "feat: add standardized EmptyState, ErrorState, LoadingSkeleton components"
```

---

### Task A1.5: 响应式布局修复

**Files:**
- Modify: `frontend/src/App.vue` (或 layout 组件)
- Modify: `frontend/src/components/Sidebar.vue`
- Modify: `frontend/src/views/IssueBoardView.vue`
- Modify: `frontend/src/views/IssueListView.vue`

- [ ] **Step 1: 修复侧边栏折叠**

```vue
<!-- 在侧边栏组件中 -->
<script setup lang="ts">
import { useMediaQuery } from '@vueuse/core'

const isMobile = useMediaQuery('(max-width: 768px)')
const sidebarOpen = ref(!isMobile.value)

watch(isMobile, (v) => {
  sidebarOpen.value = !v
})
</script>

<template>
  <aside
    :class="[
      'transition-transform duration-300',
      isMobile ? 'fixed inset-y-0 left-0 z-50 w-64' : 'relative w-64',
      isMobile && !sidebarOpen ? '-translate-x-full' : 'translate-x-0'
    ]"
  >
    <!-- sidebar content -->
  </aside>
  <!-- Mobile overlay -->
  <div
    v-if="isMobile && sidebarOpen"
    class="fixed inset-0 bg-black/50 z-40"
    @click="sidebarOpen = false"
  />
</template>
```

- [ ] **Step 2: 表格横向滚动 + 固定首列**

```vue
<!-- IssueListView 表格部分 -->
<div class="overflow-x-auto -mx-4 sm:mx-0">
  <table class="min-w-[640px] w-full">
    <thead>
      <tr>
        <th class="sticky left-0 bg-white dark:bg-gray-900 z-10">ID</th>
        <th>标题</th>
        <th>状态</th>
        <!-- ... -->
      </tr>
    </thead>
  </table>
</div>
```

- [ ] **Step 3: 看板小屏单列**

```vue
<!-- IssueBoardView -->
<div :class="[
  'grid gap-4',
  isMobile ? 'grid-cols-1' : isTablet ? 'grid-cols-2' : 'grid-cols-[repeat(auto-fit,minmax(280px,1fr))]'
]">
```

- [ ] **Step 4: 日历视图移动端降级**

```vue
<!-- CalendarView — 移动端切换为列表模式 -->
<template>
  <ListView v-if="isMobile" :issues="issuesForDate" />
  <CalendarGrid v-else :issues="issues" />
</template>
```

- [ ] **Step 5: 多断点手工验证**

在浏览器 DevTools 中模拟：320px, 375px, 768px, 1024px, 1440px, 1920px，检查布局断裂点。

- [ ] **Step 6: Commit**

```bash
git add frontend/src/
git commit -m "fix: responsive layout for mobile/tablet/desktop breakpoints"
```

---

## Track 3: B — 工程质量

### Task B1.1: Go 测试基础设施

**Files:**
- Create: `backend/internal/testutil/db.go`
- Create: `backend/internal/testutil/auth.go`
- Create: `backend/internal/testutil/fixtures.go`
- Create: `docs/dev/go-testing-guide.md`

- [ ] **Step 1: 创建 testutil 包**

```go
// backend/internal/testutil/db.go
package testutil

import (
    "database/sql"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// NewMockDB creates a sqlmock database for unit testing
func NewMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
    t.Helper()

    sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
    if err != nil {
        t.Fatalf("failed to create sqlmock: %v", err)
    }

    dialector := postgres.New(postgres.Config{
        Conn: sqlDB,
    })
    db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
    if err != nil {
        t.Fatalf("failed to open gorm db: %v", err)
    }

    return db, mock, sqlDB
}
```

```go
// backend/internal/testutil/auth.go
package testutil

import (
    "net/http/httptest"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/reqmango/backend/internal/model"
)

// AuthUser injects a mock authenticated user into the gin context
func AuthUser(c *gin.Context, user *model.User) {
    c.Set("userID", user.ID)
    c.Set("user", user)
}

// NewTestContext creates a gin context with a test response recorder
func NewTestContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest(method, path, nil)
    return c, w
}
```

```go
// backend/internal/testutil/fixtures.go
package testutil

import "github.com/reqmango/backend/internal/model"

func NewTestUser() *model.User {
    return &model.User{
        ID:    1,
        Email: "test@example.com",
        Name:  "Test User",
    }
}

func NewTestWorkspace() *model.Workspace {
    return &model.Workspace{
        ID:   1,
        Name: "Test Workspace",
        Slug: "test-workspace",
    }
}

func NewTestProject(workspaceID uint) *model.Project {
    return &model.Project{
        ID:          1,
        Name:        "Test Project",
        Identifier:  "TEST",
        WorkspaceID: workspaceID,
    }
}
```

- [ ] **Step 2: 编写测试规范文档**

```markdown
# Go 测试规范

## 命名约定
- 测试文件: `<name>_test.go`，与被测文件同目录
- 测试函数: `Test<Function>_<Scenario>`
- Table-driven tests 优先

## Mock 原则
- Service 测试用 sqlmock 模拟数据库
- Handler 测试用 httptest 模拟 HTTP
- 不 mock 自己的 model（它们是纯数据结构）

## 覆盖率目标
- Sprint 1: service 层 ≥30%
- Sprint 2: service 层 ≥50%
- Sprint 3: service 层 ≥70%
```

- [ ] **Step 3: Run tests & commit**

```bash
cd backend && go test ./internal/testutil/... -v
git add backend/internal/testutil/ docs/dev/go-testing-guide.md
git commit -m "feat: add test utilities (mock DB, auth context, fixtures)"
```

---

### Task B1.2: 前端测试基础设施

**Files:**
- Create: `frontend/src/__tests__/setup.ts`
- Modify: `frontend/vitest.config.ts`

- [ ] **Step 1: 配置 Vitest 全局 setup**

```typescript
// frontend/vitest.config.ts
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'happy-dom',
    setupFiles: ['./src/__tests__/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['src/components/**/*.vue', 'src/composables/**/*.ts', 'src/stores/**/*.ts'],
    },
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
})
```

```typescript
// frontend/src/__tests__/setup.ts
import { config } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'

// Mock i18n
const i18n = createI18n({
  legacy: false,
  locale: 'zh',
  messages: {
    zh: {},
    en: {},
  },
})

config.global.plugins = [createPinia(), i18n]

// Mock IntersectionObserver
global.IntersectionObserver = class {
  observe() {}
  unobserve() {}
  disconnect() {}
} as any

// Mock ResizeObserver
global.ResizeObserver = class {
  observe() {}
  unobserve() {}
  disconnect() {}
} as any
```

- [ ] **Step 2: 编写第一个组件测试 (TDD 验证环境)**

```typescript
// frontend/src/components/__tests__/QuickFilterChips.spec.ts
// (已在 Task A1.3 Step 4 中编写)
```

- [ ] **Step 3: 验证测试环境**

```bash
cd frontend && npx vitest run
# Expected: 3 tests pass (QuickFilterChips)
```

- [ ] **Step 4: Commit**

```bash
git add frontend/vitest.config.ts frontend/src/__tests__/setup.ts
git commit -m "feat: configure vitest with Vue Test Utils, happy-dom, and global setup"
```

---

### Task B1.3: CI/CD Pipeline

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: 编写 CI workflow**

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [master]
  pull_request:
    branches: [master]

env:
  GO_VERSION: '1.22'

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with: { go-version: '${{ env.GO_VERSION }}' }

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          working-directory: backend
          args: --timeout=5m

      - uses: actions/setup-node@v4
        with: { node-version: '20' }

      - name: ESLint
        run: |
          cd frontend
          npm ci
          npx eslint src/ --ext .ts,.vue --max-warnings 10

  test:
    name: Test
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: reqmango
          POSTGRES_PASSWORD: reqmango
          POSTGRES_DB: reqmango_test
        ports: ['5432:5432']
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with: { go-version: '${{ env.GO_VERSION }}' }

      - name: Go tests
        run: |
          cd backend
          go test -race -coverprofile=coverage.out ./internal/...
          go tool cover -func=coverage.out | tail -1

      - uses: actions/setup-node@v4
        with: { node-version: '20' }

      - name: Frontend tests
        run: |
          cd frontend
          npm ci
          npx vitest run --coverage

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with: { go-version: '${{ env.GO_VERSION }}' }

      - name: Build backend
        run: cd backend && go build -ldflags="-s -w" ./cmd/server

      - uses: actions/setup-node@v4
        with: { node-version: '20' }

      - name: Build frontend
        run: |
          cd frontend
          npm ci
          npx vite build

  docker:
    name: Docker Build
    runs-on: ubuntu-latest
    needs: [build]
    if: startsWith(github.ref, 'refs/tags/v')
    steps:
      - uses: actions/checkout@v4

      - name: Build and push Docker image
        uses: docker/build-push-action@v6
        with:
          context: .
          push: false
          tags: reqmango:latest
```

- [ ] **Step 2: 添加 CI badge 到 README**

```markdown
<!-- README.md — 顶部添加 -->
![CI](https://github.com/reqmango/reqmango/actions/workflows/ci.yml/badge.svg)
```

- [ ] **Step 3: 推送并验证 CI 跑通**

```bash
git add .github/workflows/ci.yml README.md
git commit -m "ci: add GitHub Actions CI workflow (lint, test, build, docker)"
git push origin master
# 在 GitHub Actions tab 观察运行结果
```

---

### Task B1.4: DB Migration 版本化

**Files:**
- Create: `backend/migrations/000001_initial_schema.up.sql`
- Create: `backend/migrations/000001_initial_schema.down.sql`
- Create: `backend/cmd/migrate/main.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: 生成初始 schema dump**

```bash
# 启动数据库
docker compose up -d db

# 导出当前 schema
docker compose exec db pg_dump -U reqmango -d reqmango --schema-only --no-owner > backend/migrations/000001_initial_schema.up.sql

# 创建 down migration（清空所有表）
# 手动编写或使用: 按依赖顺序 DROP TABLE
```

```sql
-- backend/migrations/000001_initial_schema.down.sql
DROP TABLE IF EXISTS issue_activities CASCADE;
DROP TABLE IF EXISTS comments CASCADE;
DROP TABLE IF EXISTS attachments CASCADE;
DROP TABLE IF EXISTS issues CASCADE;
-- ... 按依赖顺序排列所有表
```

- [ ] **Step 2: 创建 migration CLI 工具**

```go
// backend/cmd/migrate/main.go
package main

import (
    "log"
    "os"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://reqmango:reqmango@localhost:5432/reqmango?sslmode=disable"
    }

    m, err := migrate.New("file://migrations", dbURL)
    if err != nil {
        log.Fatalf("migrate init failed: %v", err)
    }

    cmd := "up"
    if len(os.Args) > 1 {
        cmd = os.Args[1]
    }

    switch cmd {
    case "up":
        err = m.Up()
    case "down":
        err = m.Down()
    case "version":
        v, dirty, _ := m.Version()
        log.Printf("version=%d dirty=%v", v, dirty)
        return
    }

    if err != nil && err != migrate.ErrNoChange {
        log.Fatalf("migrate %s failed: %v", cmd, err)
    }
    log.Printf("migrate %s: OK", cmd)
}
```

- [ ] **Step 3: 修改 server/main.go 集成 migration**

```go
// 在 main.go 中，AutoMigrate 之前添加 migration 执行逻辑
// 保留 AutoMigrate 作为开发环境的便捷选项，生产用 migrate CLI
```

- [ ] **Step 4: CI 中添加 migration 检查**

```yaml
# 在 .github/workflows/ci.yml test job 中添加
- name: Validate migrations
  run: |
    cd backend
    go run ./cmd/migrate up
    go run ./cmd/migrate version
```

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/ backend/cmd/migrate/
git commit -m "feat: add versioned database migrations with golang-migrate"
```

---

### Task B1.5: Service 层测试 — 第 1 批 (8 个文件)

**Files:**
- Create: `backend/internal/service/issue_service_test.go`
- Create: `backend/internal/service/project_service_test.go`
- Create: `backend/internal/service/cycle_service_test.go`
- Create: `backend/internal/service/workspace_service_test.go`
- Create: `backend/internal/service/auth_service_test.go`
- Create: `backend/internal/service/workflow_service_test.go`
- Create: `backend/internal/service/notification_service_test.go`
- Create: `backend/internal/service/module_service_test.go`

- [ ] **Step 1: issue_service_test.go — CRUD + 搜索**

```go
// backend/internal/service/issue_service_test.go
package service

import (
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/reqmango/backend/internal/model"
    "github.com/reqmango/backend/internal/testutil"
)

func TestIssueService_Create(t *testing.T) {
    db, mock, sqlDB := testutil.NewMockDB(t)
    defer sqlDB.Close()

    svc := NewIssueService(db)

    issue := &model.Issue{
        Title:       "Test Issue",
        Description: "Test description",
        ProjectID:   1,
        CreatedBy:   1,
    }

    // Expect INSERT
    mock.ExpectBegin()
    mock.ExpectQuery(`INSERT INTO "issues"`).
        WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), issue.Title, issue.Description, issue.ProjectID, issue.CreatedBy).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
    mock.ExpectCommit()

    err := svc.Create(issue)
    require.NoError(t, err)
    assert.Equal(t, uint(1), issue.ID)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIssueService_GetByID(t *testing.T) {
    db, mock, sqlDB := testutil.NewMockDB(t)
    defer sqlDB.Close()

    svc := NewIssueService(db)

    rows := sqlmock.NewRows([]string{"id", "title", "project_id"}).
        AddRow(1, "Test Issue", 1)
    mock.ExpectQuery(`SELECT \* FROM "issues" WHERE "issues"\."id" = \$1`).
        WithArgs(uint(1)).
        WillReturnRows(rows)

    issue, err := svc.GetByID(1)
    require.NoError(t, err)
    assert.Equal(t, "Test Issue", issue.Title)
}

func TestIssueService_List(t *testing.T) {
    // Table-driven test for pagination and filtering
    tests := []struct {
        name   string
        offset int
        limit  int
        want   int
    }{
        {"first page", 0, 20, 2},
        {"empty page", 100, 20, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db, mock, sqlDB := testutil.NewMockDB(t)
            defer sqlDB.Close()
            svc := NewIssueService(db)

            rows := sqlmock.NewRows([]string{"id", "title", "project_id"})
            for i := 0; i < tt.want; i++ {
                rows.AddRow(i+1, "Issue", 1)
            }
            mock.ExpectQuery(`SELECT \* FROM "issues"`).
                WillReturnRows(rows)

            issues, err := svc.List(1, tt.offset, tt.limit)
            require.NoError(t, err)
            assert.Len(t, issues, tt.want)
        })
    }
}

func TestIssueService_Delete(t *testing.T) {
    db, mock, sqlDB := testutil.NewMockDB(t)
    defer sqlDB.Close()
    svc := NewIssueService(db)

    mock.ExpectBegin()
    mock.ExpectExec(`DELETE FROM "issues" WHERE "issues"\."id" = \$1`).
        WithArgs(uint(1)).
        WillReturnResult(sqlmock.NewResult(0, 1))
    mock.ExpectCommit()

    err := svc.Delete(1)
    require.NoError(t, err)
}
```

- [ ] **Step 2: project_service_test.go**

```go
// backend/internal/service/project_service_test.go
package service

func TestProjectService_Create(t *testing.T) { /* 同 pattern */ }
func TestProjectService_GetByID(t *testing.T) { /* ... */ }
func TestProjectService_ListByWorkspace(t *testing.T) { /* ... */ }
func TestProjectService_Archive(t *testing.T) { /* ... */ }
func TestProjectService_Unarchive(t *testing.T) { /* ... */ }
func TestProjectService_GetStats(t *testing.T) { /* ... */ }
```

- [ ] **Step 3: cycle_service_test.go**

```go
// 重点测试燃尽图计算逻辑
func TestCycleService_CalculateBurndown(t *testing.T) {
    db, mock, sqlDB := testutil.NewMockDB(t)
    defer sqlDB.Close()
    svc := NewCycleService(db)

    // Mock issues with completed dates
    // Verify burndown data points
}

func TestCycleService_StartCycle(t *testing.T) { /* 生命周期转换 */ }
func TestCycleService_CompleteCycle(t *testing.T) { /* ... */ }
```

- [ ] **Step 4-8: 其余 5 个 service 测试文件**

按相同 pattern（sqlmock + table-driven）编写：
- `workspace_service_test.go`: CRUD + 成员管理
- `auth_service_test.go`: 注册 + 登录 + JWT 生成/验证 + refresh
- `workflow_service_test.go`: 状态转换规则 + 审批逻辑
- `notification_service_test.go`: 9 种通知类型生成
- `module_service_test.go`: 层级树操作 + 进度计算

- [ ] **Step 9: Run all service tests**

```bash
cd backend
go test ./internal/service/... -v -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
# Expected: total coverage >= 30% for service package
```

- [ ] **Step 10: Commit**

```bash
git add backend/internal/service/*_test.go
git commit -m "test: add service layer tests (8 files) — issue, project, cycle, workspace, auth, workflow, notification, module"
```

---

### Task B1.6: Handler 集成测试框架

**Files:**
- Create: `backend/internal/handler/auth_handler_test.go`
- Create: `backend/internal/handler/workspace_handler_test.go`
- Modify: `backend/internal/testutil/http.go`

- [ ] **Step 1: 创建 HTTP 测试工具**

```go
// backend/internal/testutil/http.go
package testutil

import (
    "bytes"
    "encoding/json"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/reqmango/backend/internal/model"
)

type TestServer struct {
    Engine *gin.Engine
}

func NewTestServer(t *testing.T) *TestServer {
    t.Helper()
    gin.SetMode(gin.TestMode)
    r := gin.New()
    // 添加测试用中间件
    r.Use(func(c *gin.Context) {
        // 默认注入测试用户（可通过 header 覆盖）
        c.Set("userID", uint(1))
        c.Set("user", &model.User{ID: 1, Email: "test@test.com"})
        c.Next()
    })
    return &TestServer{Engine: r}
}

func (ts *TestServer) GET(t *testing.T, path string) *httptest.ResponseRecorder {
    t.Helper()
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", path, nil)
    ts.Engine.ServeHTTP(w, req)
    return w
}

func (ts *TestServer) POST(t *testing.T, path string, body any) *httptest.ResponseRecorder {
    t.Helper()
    jsonBody, _ := json.Marshal(body)
    w := httptest.NewRecorder()
    req := httptest.NewRequest("POST", path, bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    ts.Engine.ServeHTTP(w, req)
    return w
}

func (ts *TestServer) PUT(t *testing.T, path string, body any) *httptest.ResponseRecorder {
    t.Helper()
    jsonBody, _ := json.Marshal(body)
    w := httptest.NewRecorder()
    req := httptest.NewRequest("PUT", path, bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    ts.Engine.ServeHTTP(w, req)
    return w
}

func (ts *TestServer) DELETE(t *testing.T, path string) *httptest.ResponseRecorder {
    t.Helper()
    w := httptest.NewRecorder()
    req := httptest.NewRequest("DELETE", path, nil)
    ts.Engine.ServeHTTP(w, req)
    return w
}
```

- [ ] **Step 2: auth_handler_test.go — 注册→登录→Token 验证**

```go
// backend/internal/handler/auth_handler_test.go
package handler

import (
    "encoding/json"
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/reqmango/backend/internal/dto/request"
    "github.com/reqmango/backend/internal/testutil"
)

func TestAuthHandler_Register(t *testing.T) {
    ts := testutil.NewTestServer(t)
    // 注册 handler routes
    // TODO: 需要注入 mock service 或使用集成测试数据库

    w := ts.POST(t, "/api/auth/register", request.RegisterRequest{
        Email:    "newuser@test.com",
        Password: "SecurePass123!",
        Name:     "New User",
    })

    assert.Equal(t, http.StatusCreated, w.Code)
    var resp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.NotEmpty(t, resp["token"])
}

func TestAuthHandler_Login(t *testing.T) {
    ts := testutil.NewTestServer(t)

    w := ts.POST(t, "/api/auth/login", request.LoginRequest{
        Email:    "test@test.com",
        Password: "SecurePass123!",
    })

    assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_InvalidCredentials(t *testing.T) {
    ts := testutil.NewTestServer(t)

    w := ts.POST(t, "/api/auth/login", request.LoginRequest{
        Email:    "test@test.com",
        Password: "wrong",
    })

    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

- [ ] **Step 3: workspace_handler_test.go — CRUD 集成测试**

```go
// backend/internal/handler/workspace_handler_test.go
func TestWorkspaceHandler_Create(t *testing.T) { /* ... */ }
func TestWorkspaceHandler_List(t *testing.T) { /* ... */ }
func TestWorkspaceHandler_Update(t *testing.T) { /* ... */ }
func TestWorkspaceHandler_Delete(t *testing.T) { /* ... */ }
```

- [ ] **Step 4: Run handler tests**

```bash
cd backend && go test ./internal/handler/... -v
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/testutil/http.go backend/internal/handler/*_test.go
git commit -m "test: add handler integration test framework + auth and workspace tests"
```

---

## Sprint 1 结束检查点 (W3 末)

### 交叉 Review 流程

- [ ] **A review B's output:**
  - 检查 service 测试是否覆盖了正常路径 + 边界 + 错误路径
  - 检查 CI workflow 是否合理（lint → test → build 顺序）
  - 确认 testutil 是否易用（尝试自己写一个测试）

- [ ] **B review A's output:**
  - 检查 i18n 键是否有遗漏（跑 en locale 检查控制台错误）
  - 检查状态组件的 props 接口是否合理
  - 检查响应式变更有无破坏现有布局

- [ ] **vinthuy final review:**
  - 合并 A 和 B 的 review 意见
  - 确认 `make ci` 通过
  - 确认 service 测试覆盖率 ≥30%
  - 确认 UX 审计报告已分级

### Sprint 1 交付物

```bash
# 验证命令
make ci                    # 必须：lint + test + build 全通过
cd backend && go tool cover -func=coverage.out | grep total  # 必须：≥30%
make lint                  # 必须：无新增 lint error
ls docs/dev/               # 应有：architecture-audit, ux-audit, realtime-wiki-design, go-testing-guide
ls .github/workflows/      # 应有：ci.yml
```

**Sprint 1 结束 → 产出可演示版本：make ci 通过、测试覆盖率达标、UX 问题清单已分级。**
