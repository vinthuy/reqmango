# 工作项标签纯项目级（移除工作空间标签）实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 工作项标签回归纯项目级（对齐 Plane），移除工作空间标签的全部前后端实现与遗留数据，并顺带修复重名无约束、跨项目挂标签、报表标签路径三个缺陷。

**Architecture:** 单表 `labels` 保留，`project_id` 变 NOT NULL；删除 5 条工作空间标签路由/handler/service；项目标签查询去继承；启动时清理 `project_id IS NULL` 遗留行并建 `(project_id, name)` 部分唯一索引。前端删除 WorkspaceSettings 标签区块、ProjectSettings 去继承徽章。

**Tech Stack:** Go 1.22 + Gin + GORM + Postgres；Vue 3 + TS。仓库惯例：无标签单测，验证靠 build/vet + 双服务 live 冒烟（对照 spec `docs/superpowers/specs/2026-07-18-project-only-labels-design.md`）。

**约定:** 所有后端命令在 `D:\code\reqmango2\backend` 执行，agent-service 在 `D:\code\reqmango2\agent-service`，前端在 `D:\code\reqmango2\frontend`。提交信息末尾加 `Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>`（heredoc 方式）。

---

### Task 1: 后端移除工作空间标签 API（路由 + handler + service）

**Files:**
- Modify: `backend/internal/router/router.go:275-280`
- Modify: `backend/internal/handler/project_settings_handler.go:486-620`
- Modify: `backend/internal/service/project_settings_service.go:397-495`

- [ ] **Step 1: 删除路由**

删除 router.go 中这 6 行（275-280，含注释）：

```go
			// Workspace-level Settings: Labels
			workspaces.GET("/:wsParam/settings/labels", settingsH.ListWorkspaceLabels)
			workspaces.POST("/:wsParam/settings/labels", settingsH.CreateWorkspaceLabel)
			workspaces.GET("/:wsParam/settings/labels/:labelId", settingsH.GetWorkspaceLabel)
			workspaces.PUT("/:wsParam/settings/labels/:labelId", settingsH.UpdateWorkspaceLabel)
			workspaces.DELETE("/:wsParam/settings/labels/:labelId", settingsH.DeleteWorkspaceLabel)
```

保留其上的 States 路由块和其下的 `// Workspace-level Modules` 块之间的空行结构（两块之间留一个空行）。

- [ ] **Step 2: 删除 5 个 handler**

删除 `project_settings_handler.go` 从 `// ==================== Workspace Labels ====================`（:486）到文件中 `DeleteWorkspaceLabel` 函数结尾（:620，函数以 `c.JSON(http.StatusNoContent, nil)` + `}` 结束）的整块。被删函数：`ListWorkspaceLabels`、`CreateWorkspaceLabel`、`GetWorkspaceLabel`、`UpdateWorkspaceLabel`、`DeleteWorkspaceLabel`。注意 `getWorkspaceID`（:26）被 Workspace States handlers 共用，**保留**。

- [ ] **Step 3: 删除 5 个 service 方法**

删除 `project_settings_service.go` 中 `ListWorkspaceLabels`（:397-409）、`CreateWorkspaceLabel`（:411-431）、`GetWorkspaceLabel`（:433-443）、`UpdateWorkspaceLabel`（:445-473）、`DeleteWorkspaceLabel`（:475-495）五个函数（含各自的注释行）。保留其后的 `GetLabel`（:497 起）。

- [ ] **Step 4: 编译检查**

Run: `go build ./... && go vet ./...`
Expected: 无输出（成功）。若报未使用的 import，按编译器提示删除。

- [ ] **Step 5: Commit**

```bash
git add internal/router/router.go internal/handler/project_settings_handler.go internal/service/project_settings_service.go
git commit -m "refactor(labels): remove workspace-level label API"
```

---

### Task 2: 项目标签查询去继承 + 移除 is_inherited

**Files:**
- Modify: `backend/internal/service/project_settings_service.go`（ListLabels、SearchLabels、labelToResponse）
- Modify: `backend/internal/dto/response/label.go`

- [ ] **Step 1: ListLabels 去继承**

将（Task 1 完成后行号会前移，按内容定位）：

```go
// ListLabels returns all labels for a project including inherited workspace labels.
func (s *ProjectSettingsService) ListLabels(projectID uint64) ([]response.LabelResponse, error) {
	var project model.Project
	if err := s.db.Select("workspace_id").Where("id = ?", projectID).First(&project).Error; err != nil {
		return nil, common.Internal("Project not found")
	}
	var labels []model.Label
	if err := s.db.Where("(project_id = ? OR (project_id IS NULL AND workspace_id = ?))", projectID, project.WorkspaceID).Order("created_at ASC").Find(&labels).Error; err != nil {
		return nil, common.Internal("Database error")
	}
```

改为：

```go
// ListLabels returns all labels for a project.
func (s *ProjectSettingsService) ListLabels(projectID uint64) ([]response.LabelResponse, error) {
	var labels []model.Label
	if err := s.db.Where("project_id = ?", projectID).Order("created_at ASC").Find(&labels).Error; err != nil {
		return nil, common.Internal("Database error")
	}
```

- [ ] **Step 2: SearchLabels 去继承**

同样将 SearchLabels 开头的 project 查询删除，查询条件改为：

```go
// SearchLabels returns labels matching the query.
func (s *ProjectSettingsService) SearchLabels(projectID uint64, query string) ([]response.LabelResponse, error) {
	var labels []model.Label
	if err := s.db.Where("project_id = ? AND name ILIKE ?", projectID, "%"+query+"%").Order("created_at ASC").Find(&labels).Error; err != nil {
		return nil, common.Internal("Database error")
	}
```

- [ ] **Step 3: labelToResponse 删除 IsInherited**

在 `labelToResponse` 中删除这一行：

```go
		IsInherited: label.ProjectID == nil,
```

（`stateToResponse` 的 `IsInherited` 是 states 继承功能，**不动**。）

- [ ] **Step 4: LabelResponse 删除字段**

`backend/internal/dto/response/label.go` 删除：

```go
	IsInherited bool       `json:"is_inherited"`
```

- [ ] **Step 5: 编译检查 + Commit**

Run: `go build ./... && go vet ./...` → 成功。

```bash
git add internal/service/project_settings_service.go internal/dto/response/label.go
git commit -m "refactor(labels): project-scoped label queries, drop is_inherited from response"
```

---

### Task 3: ProjectID 值类型 + 重名 409 + 挂标签作用域校验

**Files:**
- Modify: `backend/internal/model/label.go:10`
- Modify: `backend/internal/dto/response/label.go:11`
- Modify: `backend/internal/service/project_settings_service.go`（CreateLabel、UpdateLabel、labelToResponse）
- Modify: `backend/internal/seed/seed.go:313-321, 931-934`
- Modify: `backend/internal/service/project_template_service.go:265-272`
- Modify: `backend/internal/service/issue_service.go:986-1006`

- [ ] **Step 1: 模型改值类型**

`model/label.go:10`：

```go
	ProjectID   uint64  `gorm:"not null;index" json:"project_id"`
```

- [ ] **Step 2: 响应改值类型**

`dto/response/label.go:11`：

```go
	ProjectID   uint64     `json:"project_id"`
```

- [ ] **Step 3: CreateLabel 改构造 + 重名 409**

`project_settings_service.go` CreateLabel 中：

```go
	projectIDPtr := projectID
	label := &model.Label{
		Name:        req.Name,
		Color:       color,
		Description: req.Description,
		ProjectID:   &projectIDPtr,
		WorkspaceID: project.WorkspaceID,
	}

	if err := s.db.Create(label).Error; err != nil {
		return nil, common.Internal("Failed to create label")
	}
```

改为：

```go
	label := &model.Label{
		Name:        req.Name,
		Color:       color,
		Description: req.Description,
		ProjectID:   projectID,
		WorkspaceID: project.WorkspaceID,
	}

	if err := s.db.Create(label).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, common.Conflict("Label name already exists in this project")
		}
		return nil, common.Internal("Failed to create label")
	}
```

CreateLabel 开头的 project 查询（取 WorkspaceID 用）**保留**。若文件未 import `"strings"` 则添加（参照 auth_service.go:61 的既有判重模式）。

- [ ] **Step 4: UpdateLabel 重名 409**

UpdateLabel 中的：

```go
		if err := s.db.Model(&label).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update label")
		}
```

改为：

```go
		if err := s.db.Model(&label).Updates(updates).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				return nil, common.Conflict("Label name already exists in this project")
			}
			return nil, common.Internal("Failed to update label")
		}
```

- [ ] **Step 5: labelToResponse 赋值适配**

`ProjectID: label.ProjectID,` 行保持不变（现在两侧都是 uint64，直接赋值成立）。确认无 `*label.ProjectID` 解引用残留。

- [ ] **Step 6: seed 两处改值构造**

`seed.go:313-321` 改为：

```go
		labels := make([]model.Label, len(labelTemplates))
		for i, l := range labelTemplates {
			labels[i] = model.Label{
				Name:        l.name, Color: l.color, ProjectID: proj.ID,
				WorkspaceID: ws.ID,
			}
			db.Create(&labels[i])
		}
```

`seed.go:931-934` 改为：

```go
		for _, l := range labelDefs {
			db.Create(&model.Label{Name: l.name, Color: l.color, ProjectID: proj.ID, WorkspaceID: ws.ID})
		}
```

（模块 modules 循环里的 `projectIDPtr` 是另一个变量，不动。）

- [ ] **Step 7: 项目模板复制标签改值构造**

`project_template_service.go:265-272` 改为：

```go
		for _, tl := range templateLabels {
			label := model.Label{
				Name:        tl.Name,
				Color:       tl.Color,
				ProjectID:   projectID,
				WorkspaceID: project.WorkspaceID,
			}
```

- [ ] **Step 8: AddLabel 作用域校验**

`issue_service.go` AddLabel（:987）在重复检查之前插入标签归属校验，函数开头部分改为：

```go
// AddLabel adds a label to an issue.
func (s *IssueService) AddLabel(issueID, labelID, actorID uint64) error {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	var label model.Label
	if err := s.db.First(&label, labelID).Error; err != nil {
		return common.NotFound("Label not found")
	}
	if label.ProjectID != issue.ProjectID {
		return common.BadRequest("Label does not belong to this project")
	}
```

其余（重复检查、Create、activity）不变。注：`model.Issue.ProjectID` 为 uint64 值类型，如实际为指针需解引用比较——以编译器为准。

- [ ] **Step 9: 全局搜残留 + 编译**

Run: `grep -rn "ProjectID: &" internal/ | grep -i label` → 应无结果。
Run: `go build ./... && go vet ./...` → 成功。

- [ ] **Step 10: Commit**

```bash
git add internal/model/label.go internal/dto/response/label.go internal/service/project_settings_service.go internal/seed/seed.go internal/service/project_template_service.go internal/service/issue_service.go
git commit -m "feat(labels): required project_id, duplicate-name 409, cross-project attach guard"
```

---

### Task 4: 启动数据清理 + 唯一索引 + 迁移文件

**Files:**
- Modify: `backend/cmd/server/main.go`
- Create: `backend/migrations/000009_project_only_labels.up.sql`
- Create: `backend/migrations/000009_project_only_labels.down.sql`

- [ ] **Step 1: main.go 清理（AutoMigrate 之前）**

在 `if err := db.AutoMigrate(` 语句**之前**（`fmt.Println("Database connected")` 之后）插入：

```go
	// Purge legacy workspace-level labels so project_id can become NOT NULL (project-only labels, aligned with Plane)
	db.Exec(`DELETE FROM issue_labels WHERE label_id IN (SELECT id FROM labels WHERE project_id IS NULL)`)
	db.Exec(`DELETE FROM labels WHERE project_id IS NULL`)
```

- [ ] **Step 2: main.go 唯一索引（AutoMigrate 之后）**

在全文索引 Exec（`CREATE INDEX IF NOT EXISTS idx_issues_search ...`）块之后插入：

```go
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_project_name ON labels (project_id, name) WHERE deleted_at IS NULL`).Error; err != nil {
		log.Printf("WARNING: failed to create labels unique index: %v", err)
	}
```

- [ ] **Step 3: 迁移文件**

`backend/migrations/000009_project_only_labels.up.sql`：

```sql
-- Work item labels become project-only (aligned with Plane): purge workspace-level labels
DELETE FROM issue_labels WHERE label_id IN (SELECT id FROM labels WHERE project_id IS NULL);
DELETE FROM labels WHERE project_id IS NULL;
ALTER TABLE labels ALTER COLUMN project_id SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_project_name ON labels (project_id, name) WHERE deleted_at IS NULL;
```

`backend/migrations/000009_project_only_labels.down.sql`：

```sql
DROP INDEX IF EXISTS idx_labels_project_name;
ALTER TABLE labels ALTER COLUMN project_id DROP NOT NULL;
```

- [ ] **Step 4: 编译 + Commit**

Run: `go build ./... && go vet ./...` → 成功。

```bash
git add cmd/server/main.go migrations/000009_project_only_labels.up.sql migrations/000009_project_only_labels.down.sql
git commit -m "feat(labels): purge workspace labels on startup, unique (project_id,name) index"
```

---

### Task 5: 前端删除 WorkspaceSettings 标签区块

**Files:**
- Modify: `frontend/src/views/WorkspaceSettings.vue`
- Modify: `frontend/src/locales/en-US.json`、`frontend/src/locales/zh-CN.json`

- [ ] **Step 1: 删除 ref 与导航项**

删除 :57 `const workspaceLabels = ref<any[]>([]);` 和 :64 导航数组中的 `{ id: 'labels', label: t('settings.labels'), icon: '🏷️', count: workspaceLabels.value.length },`。

- [ ] **Step 2: loadAllData 去掉标签请求并重排索引**

删除 :108 `api.get(\`/workspaces/${wid}/settings/labels\`).then(r => r.data),`（Promise.allSettled 数组第 13 项）。随后结果解构中删除 :127 的 workspaceLabels 行，并把 modules 的索引从 14 改 13：

```ts
    workspaceStates.value = results[12].status === 'fulfilled' ? (Array.isArray(results[12].value) ? results[12].value : []) : [];
    workspaceModules.value = results[13].status === 'fulfilled' ? (Array.isArray(results[13].value) ? results[13].value : []) : [];
```

- [ ] **Step 3: 删除 handlers 块**

删除 :239-275 整块（`// ===== Workspace Labels handlers =====` 至 `wsHandleDeleteLabel` 函数结束）：`wsShowLabelModal`、`wsEditingLabel`、`wsNewLabelForm`、`wsHandleAddLabel`、`wsHandleEditLabel`、`wsHandleSaveLabel`、`wsHandleDeleteLabel`。

- [ ] **Step 4: 删除模板区块与弹窗**

删除 `<!-- Workspace Labels Section -->` 区块（:387-400，`v-if="!loading && activeSection === 'labels'"` 的整个 div）和 `<!-- Workspace Label Modal -->`（:533-545，`v-if="wsShowLabelModal"` 的整个 div）。

- [ ] **Step 5: i18n 清理**

在 `en-US.json` 与 `zh-CN.json` 删除 `settings.workspaceLabelsDesc` 键（已确认仅 WorkspaceSettings.vue:390 使用）。`settings.labels`、`addLabel`、`editLabel`、`deleteLabel`、`confirmDeleteLabel`、`noLabels` 等键被 ProjectSettings 共用，**保留**。

- [ ] **Step 6: 构建检查**

Run: `cd D:\code\reqmango2\frontend && npm run build`
Expected: 通过（vue-tsc 无 wsHandle*/workspaceLabels 相关报错）。

- [ ] **Step 7: Commit**

```bash
git add src/views/WorkspaceSettings.vue src/locales/en-US.json src/locales/zh-CN.json
git commit -m "refactor(labels): remove workspace labels section from workspace settings"
```

---

### Task 6: 前端 ProjectSettings 去继承逻辑 + 辅助修复

**Files:**
- Modify: `frontend/src/views/ProjectSettings.vue:333, 735-739`
- Modify: `frontend/src/api/project-settings.ts:165-178`
- Modify: `frontend/src/components/ReportBuilder.vue:564`

- [ ] **Step 1: 创建请求去掉 workspace_id**

ProjectSettings.vue:333 改为：

```ts
      await api.post(`/projects/${projectId.value}/settings/labels`, {
```

- [ ] **Step 2: 移除继承徽章与只读逻辑**

:735-739 的标签渲染改为（去掉 is_inherited 三处）：

```html
              <div v-for="label in labels" :key="label.id" @click="handleEditLabel(label)" class="inline-flex items-center px-3 py-1.5 rounded-full cursor-pointer hover:opacity-80 transition-opacity" :style="{ backgroundColor: label.color + '20', borderColor: label.color }">
                <div class="w-2 h-2 rounded-full mr-2" :style="{ backgroundColor: label.color }"></div>
                <span class="text-sm font-medium" :style="{ color: label.color }">{{ label.name }}</span>
                <button @click.stop="handleDeleteLabel(label)" class="ml-2 text-gray-400 hover:text-red-500">✕</button>
```

（重名 409 的用户提示无需新增代码：`handleSaveLabel` 的 catch 已展示 `e?.response?.data?.message`，将直接显示后端 "Label name already exists in this project"。）

- [ ] **Step 3: createLabel 辅助函数简化**

`api/project-settings.ts` createLabel 改为（已确认无调用方，纯签名清理）：

```ts
export async function createLabel(projectId: number, data: LabelCreate): Promise<Label> {
  const response = await api.post(`/projects/${projectId}/settings/labels`, data)
  return response.data
}
```

- [ ] **Step 4: ReportBuilder 路径修复**

ReportBuilder.vue:564 改为：

```ts
      api.get(`/projects/${props.projectId}/settings/labels`).catch(() => ({ data: [] })),
```

（`types/project-settings.ts` 的 `Label.project_id` 已是必填 `number`，无需改动。）

- [ ] **Step 5: 构建 + Commit**

Run: `npm run build` → 通过。

```bash
git add src/views/ProjectSettings.vue src/api/project-settings.ts src/components/ReportBuilder.vue
git commit -m "fix(labels): drop inherited-label UI, clean create params, fix report builder labels path"
```

---

### Task 7: 删除工作空间标签 E2E 测试块

**Files:**
- Modify: `frontend/e2e/workspace-settings-e2e.spec.ts:542-627`

- [ ] **Step 1: 删除 describe 块**

删除从 :542 分隔注释 `// Workspace Labels Management (Inheritance Feature)` 到 :627 `})`（`test.describe('Workspace Settings - Labels Management', ...)` 整块，含 6 个用例；其中 "project labels include inherited workspace labels" 正是历史残留数据来源）。前面的 States 块（:540 结束）和后面的 Modules 块（:629 起）保留。

- [ ] **Step 2: 确认无其他继承断言**

Run: `grep -n "is_inherited" D:\code\reqmango2\frontend\e2e\*.spec.ts`
Expected: 仅剩 states/modules 相关（若有），无标签相关结果。`all-features.spec.ts:427` 的标签列表用例断言 `[200,404]`，改后仍通过，不动。

- [ ] **Step 3: Commit**

```bash
git add e2e/workspace-settings-e2e.spec.ts
git commit -m "test(labels): drop workspace label e2e block (feature removed)"
```

---

### Task 8: 全链路验证（live 冒烟）

**Files:** 无代码改动；发现 bug 则修复并追加提交。

- [ ] **Step 1: 双端构建测试二进制**

```bash
cd D:\code\reqmango2\backend && go build -o server_test.exe ./cmd/server
cd D:\code\reqmango2\agent-service && go build -o server_test.exe ./cmd/server
```

- [ ] **Step 2: 起服务（后台）**

backend（在 backend 目录）：`PORT=8010 AGENT_SERVICE_URL="http://localhost:8001" ./server_test.exe`
agent-service（在 agent-service 目录）：`DATABASE_URL="postgres://postgres:postgres@localhost:5432/reqmango?sslmode=disable&client_encoding=utf8&timezone=UTC" SECRET_KEY="change-me-in-production-use-a-long-random-string" AGENT_SERVICE_PORT=8001 MAIN_BACKEND_URL="http://localhost:8010" DEEPSEEK_API_KEY="$(grep DEEPSEEK_API_KEY /d/code/reqmango2/backend/.env | cut -d= -f2)" ./server_test.exe`
注意：8000 端口有无关旧实例，勿动。登录：`POST http://localhost:8010/api/v1/auth/login {"email":"admin@reqmango.com","password":"demo1234"}`。

- [ ] **Step 3: 数据迁移断言**

```bash
psql "postgres://postgres:postgres@localhost:5432/reqmango" \
  -c "SELECT count(*) FROM labels WHERE project_id IS NULL;" \
  -c "SELECT indexdef FROM pg_indexes WHERE indexname = 'idx_labels_project_name';" \
  -c "SELECT is_nullable FROM information_schema.columns WHERE table_name='labels' AND column_name='project_id';"
```

Expected: `0`；索引存在含 `WHERE (deleted_at IS NULL)`；`NO`。

- [ ] **Step 4: API 冒烟（依次断言）**

1. `GET /api/v1/workspaces/1/settings/labels` → **404**
2. `GET /api/v1/projects/4/settings/labels` → 200，所有元素无 `is_inherited` 字段且 `project_id` 均为 4
3. `POST /api/v1/projects/4/settings/labels` body `{"name":"E2E-唯一性测试","color":"#112233"}` → 201（记下 id=L1）
4. 同 body 再 POST → **409**，message 含 "already exists"
5. 在另一项目（如 project 5）同名 POST → 201（跨项目允许重名，记下 id=L2）
6. 建测试工作项：`POST /api/v1/issues?project_id=4&workspace_id=1` `{"name":"E2E label scope test"}` → 201（记下 issueId）
7. `POST /api/v1/issues/{issueId}/labels?label_id={L1}` → 200/201（本项目标签）
8. `POST /api/v1/issues/{issueId}/labels?label_id={L2}` → **400**，message 含 "does not belong"
9. `GET /api/v1/projects/4/settings/labels/search?q=唯一` → 200 且仅含本项目结果
（中文 payload 用 `printf` UTF-8 转义 + `--data-binary @file`，Git Bash 直接 -d 会发 GBK。）

- [ ] **Step 5: 浏览器/前端验证**

起前端 dev server（`npm run dev`），浏览器确认：工作空间设置侧边栏无"标签"项；项目设置标签页正常增删改、无绿色 ⚙️ 继承徽章、重名创建弹出后端 409 消息；报表构建器标签筛选下拉有数据。（若无浏览器自动化条件，用 Playwright `npx playwright test e2e/workspace-settings-e2e.spec.ts` 跑通剩余块，并明确告知用户 UI 部分的验证方式。）

- [ ] **Step 6: 清理测试数据与进程**

删除测试工作项与 L1/L2 标签（API 或 psql），停掉两个 server_test.exe 并删除二进制。psql 复核无 "E2E-唯一性测试" 残留。

- [ ] **Step 7: 收尾提交（如有修复）**

验证中发现的 bug 修复随发现随提交（`fix(labels): ...`）。全部通过后向用户汇报，询问是否推送双远程。

---

## Self-Review 结论

- Spec 覆盖：目标状态 6 条 ↔ Task 1(路由移除)、2(去继承/去 is_inherited)、3(NOT NULL/409/400)、4(数据清理/索引)、5-6(前端)、7(测试)、8(验证含全部断言) — 全覆盖。
- 占位符扫描：无 TBD/省略；所有代码块为真实现文件内容的完整替换段。
- 类型一致性：`ProjectID uint64` 贯穿 model/response/构造点；`label.ProjectID != issue.ProjectID` 比较在 Step 8 备注了指针兜底。
- 已知无需改动项已显式标注（types Label、all-features.spec.ts、agent-service suggest-labels、getWorkspaceID、states 继承）。
