# 工作项标签纯项目级（对齐 Plane）设计

日期：2026-07-18
状态：已批准（用户确认设计 + 修复范围"一起修"）

## 背景与动机

当前仓库的工作项标签是 GitLab 式两级模型：`labels.project_id IS NULL` 表示工作空间级标签，被该空间所有项目继承（项目设置里只读展示，`is_inherited=true`）；`project_id` 非空表示项目私有标签。

用户要求：**工作项标签回归纯项目级，与 Plane 的设计理念一致，移除工作空间标签。**（Plane 中 issue label 严格属于单个项目，工作空间层面没有可继承的 issue 标签。）

现存数据经查证：`project_id IS NULL` 的标签共 9 条，全部是 E2E 测试残留（名称形如 "E2E Inherit Label <timestamp>"），零工作项挂载，可直接硬删，不存在真实数据迁移问题。

本轮同时修复标签功能的既有缺陷（用户选择"一起修"）：同项目重名无约束、工作项可挂任意项目/空间的标签、报表构建器标签接口路径错误。

## 目标状态

- 标签必属一个项目：`labels.project_id NOT NULL`
- 工作空间设置不再有标签管理；`/workspaces/:id/settings/labels*` 路由全部移除
- 项目标签列表/搜索只返回本项目标签（无继承合并）
- 同一项目内标签名唯一（未删除的），冲突返回 409
- 工作项只能挂本项目的标签，跨项目挂载返回 400
- `is_inherited` 从标签响应中消失（**states 的工作空间继承机制不动**）

## 变更明细

### 1. 数据模型与数据清理（backend）

- `internal/model/label.go:10`：`ProjectID *uint64` → `ProjectID uint64`，tag 改为 `gorm:"not null;index"`。
- 唯一索引用原生 SQL 建（与 main.go 既有的全文索引 Exec 模式一致，GORM tag 不写 partial unique index）：
  ```sql
  CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_project_name ON labels (project_id, name) WHERE deleted_at IS NULL;
  ```
  已核实当前库中同项目无重名标签，索引可直接创建；该 Exec 的 error 需记录 warning 日志（不静默），以防其他环境存在脏数据时索引建失败无感知。
- `cmd/server/main.go`：在 `db.AutoMigrate(...)` **之前**执行清理（否则 NOT NULL 约束加不上）：
  ```sql
  DELETE FROM issue_labels WHERE label_id IN (SELECT id FROM labels WHERE project_id IS NULL);
  DELETE FROM labels WHERE project_id IS NULL;
  ```
  硬删（含软删行，NULL 行必须清零）。
- 新增 `backend/migrations/000009_project_only_labels.up.sql` / `.down.sql` 记录同等 SQL（清理 + NOT NULL + 唯一索引），与仓库迁移目录惯例保持一致（运行时仍靠 AutoMigrate + main.go Exec）。

### 2. 后端 API

移除（工作空间标签整套）：
- `internal/router/router.go:275-280`：5 条 `/workspaces/:wsParam/settings/labels*` 路由及注释
- `internal/handler/project_settings_handler.go:486-620` 附近：`ListWorkspaceLabels`、`CreateWorkspaceLabel`、`GetWorkspaceLabel`、`UpdateWorkspaceLabel`、`DeleteWorkspaceLabel` 5 个 handler 及 "Workspace Labels" 分节注释
- `internal/service/project_settings_service.go:397-495`：同名 5 个 service 方法

修改（继承逻辑与响应）：
- `ListLabels`（service:369）与 `SearchLabels`（service:387）：查询改为 `WHERE project_id = ?`（去掉 `OR (project_id IS NULL AND workspace_id = ?)`；随之两方法开头为取 workspace_id 而做的 Project 查询若仅此用途则一并删除）
- `labelToResponse`（service:588-609）：删除 `IsInherited` 赋值；`internal/dto/response/label.go`：`LabelResponse` 删除 `IsInherited` 字段，`ProjectID *uint64` → `uint64`
- `GetLabel`/`UpdateLabel`/`DeleteLabel`/CSV 导入的 `loadLabelMap`（issue_service.go:2150）已是 `project_id = ?` 语义，继承移除后自动一致，无需改动

新增校验：
- `CreateLabel`/`UpdateLabel`：撞唯一索引（Postgres 错误码 23505）时返回 409，消息含重名提示（如 "Label name already exists in this project"）
- 工作项挂标签 `AddLabel`（issue_service.go:987 附近）：先查标签，校验 `label.ProjectID == issue.ProjectID`，不符返回 400（"Label does not belong to this project"）

`ProjectID` 类型改动的编译波及点（全部改为值语义）：
- `internal/seed/seed.go:316`、`:933`（`ProjectID: &projectIDPtr` → `ProjectID: proj.ID` 等）
- `internal/service/project_template_service.go:267`
- `internal/service/project_settings_service.go:347`（CreateLabel 构造）

### 3. 前端

- `src/views/WorkspaceSettings.vue`：删除标签管理整块——数据加载（:108）、更新/创建（:259/:263）、删除（:274）、UI 区块（:388-400）、相关 ref/状态/函数、侧边栏导航项及对应 i18n 键（`src/locales/en-US.json`、`zh-CN.json` 中仅被该区块引用的键）
- `src/views/ProjectSettings.vue`：删除 `is_inherited` 只读逻辑（:735 附近的判断、绿色图标、禁点/禁删）；创建请求（:333）去掉 `?workspace_id=`；新增 409 错误的用户提示（重名）
- `src/api/project-settings.ts:174`：`createLabel` 去掉 `workspace_id` 参数
- `src/types/project-settings.ts`：`Label.project_id` 改必填 `number`
- `src/components/ReportBuilder.vue:564`：路径 `/projects/${id}/labels` → `/projects/${id}/settings/labels`（修复静默失败导致报表标签筛选恒为空）
- 其余标签消费方（IssueDetail、IssueDetailPanel、FilterBar、LabelSelector、AutomationForm）均走项目标签接口，无需改动

### 4. 测试

- 删除 `frontend/e2e/workspace-settings-e2e.spec.ts:543-620` 的 "Workspace Settings - Labels Management" describe 块及 "project labels include inherited workspace labels" 用例（该块正是 9 条残留数据的来源）
- `frontend/e2e/all-features.spec.ts:428` 若含继承断言则同步调整
- 后端如有引用被删方法的测试，一并删除/调整

## 验证方案

1. `cd backend && go build ./... && go vet ./...`；`cd agent-service && go build ./...`（agent-service 的 suggest-labels 已是 `project_id = ?`，只需确认编译）
2. `cd frontend && npm run build`（含 vue-tsc 类型检查）
3. 本地起双服务 live 冒烟：
   - 启动日志确认清理 SQL 执行、NOT NULL/唯一索引就位；`SELECT count(*) FROM labels WHERE project_id IS NULL` = 0
   - 项目标签 CRUD 正常；同项目重名创建 → 409；不同项目同名 → 允许
   - 工作项挂本项目标签 → 200；挂其他项目标签 → 400
   - `GET/POST /workspaces/1/settings/labels` → 404
   - 浏览器验证：WorkspaceSettings 无标签区块，ProjectSettings 标签管理正常、无继承徽章，报表构建器标签筛选出数据
4. 测试数据全部清理，测试二进制删除

## 不做的事（Out of Scope）

- states、issue types、custom fields 的工作空间级/继承机制一律不动
- 不引入 Plane 的 "Project Labels"（给项目本身打标签的工作空间级功能）
- 前端 `is_inherited` 相关的 Label 之外用法（states 仍在用）不动
