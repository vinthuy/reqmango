# KB Changelog（知识库变更日志）

---

## 2026-07-04 — 搜索API全面补全 + IssueType管理增强 + N+1查询优化

**类型**: Feature + Fix + Enhancement

**变更内容**:

### 搜索API全面补全
- 新增 `Initiative` 搜索API：`GET /workspaces/:id/initiatives/search?q=` (按名称模糊搜索)
- 新增 `Release` 搜索API：`GET /projects/:id/releases/search?q=` (按名称/版本模糊搜索)
- 新增 `Cycle` 搜索API：`GET /projects/:id/cycles/search?q=` (按名称模糊搜索)
- 新增 `Module` 搜索API：`GET /modules/search?project_id=&workspace_id=&q=` (按名称模糊搜索)
- 新增 `Label` 搜索API：`GET /projects/:id/settings/labels/search?q=` (按名称模糊搜索)
- 新增 `Page` 搜索API：`GET /projects/:id/pages/search?q=` (按标题/内容模糊搜索)

### IssueType管理增强
- IssueTypeList页面新增自定义字段绑定UI（显示已绑定字段、添加/删除字段、切换必填状态）
- 修复IssueType模板应用逻辑：`CopyFromWorkspace` 方法现在会正确复制字段关联
- 新增工作空间级别排序API：`PATCH /issue-types/reorder-workspace?workspace_id=`
- 新增国际化翻译：`issueTypePage.customFields`, `required`, `removeField` 等

### N+1查询优化
- 优化 `CycleService.ListByProject`：批量加载Project和User数据，避免N+1查询
- 新增 `buildResponseWithData` 方法支持预加载数据
- `CycleService.Search` 同样使用批量加载

### API响应格式统一
- `Release` API响应格式统一为 `{"data": release}` 格式（与Initiative和Issue保持一致）

### 其他修复
- 修复 `release_service.go` 中IssueID字符串转换bug：`string(rune(issueID+'0'))` → `fmt.Sprintf("%d", issueID)`
- 修复 `ai_service.go` 中switch-case语法错误
- 修复 `CycleCreate.vue` 路由参数NaN问题（支持两种路由风格）
- 修复 `issue_service.go` 中SetCycle验证cycle存在性
- 修复 `cycle_service.go` 中Delete事务保护

**影响文件**:
- 后端新增：cycle_service.go (Search方法, buildResponseWithData)
- 后端修改：module_service.go, project_settings_service.go, page_service.go, issue_type_service.go, issue_type_handler.go, router.go
- 前端修改：IssueTypeList.vue (字段绑定UI), CycleCreate.vue (路由修复)
- 新增迁移：000004_initiative_soft_delete.up.sql

---

## 2026-07-03 — FilterBar 统一筛选栏重构 + SavedView 增强

**类型**: Feature + Documentation

**变更内容**:
- 新增 `FilterBar.vue` 统一筛选栏组件，替代旧分散筛选组件（IssueFilterBar/QuickFilterChips/RQLInput）
- 新增 `useFilters.ts` composable（Provide/Inject 架构，RQL 双向同步）
- 新增 `types/filters.ts`（FilterCondition/FilterField 类型 + FILTER_FIELDS 12 字段 + buildRQL/parseRQL）
- SavedView 扩展：新增 `sort_config`, `columns`, `group_by`, `rql` 字段
- RQL executor 增强：新增 `state_group` 字段映射、日期 BETWEEN 语法支持
- 语义操作符系统：is/is not/contains/is any of/between 等 14 种操作符
- 筛选字段扩充至 12 个，排序选项 5 种，分组选项 8 种
- i18n 新增 `filter.*` 命名空间（zh-CN + en-US）
- KB 同步更新：README.md, architecture/README.md, data-model.md, frontend.md, saved-views-design.md, backend-go.md, project-layout.md

**影响文件**:
- 前端新增：FilterBar.vue, useFilters.ts, types/filters.ts
- 后端修改：saved_view.go (model/dto), executor.go (state_group), issue_handler.go (sort)
- 修改文件：Project.vue, IssueList.vue, SavedViewSelector.vue, locales/*.json

---

## 2026-06-27 — 差距补全计划收官：28/28 Tasks + KB 最终同步

**类型**: Feature + Documentation

**变更内容**:
- Task 13-14,19,24-28 完成 (Time Tracking, Recurring, Intake & Triage, AI 全家桶)
- 全部 28/28 Tasks 完成 🎉
- KB 全量同步: 架构文档、数据模型、前端架构更新至最终状态
- 新增 TimeTrack, RecurrenceRule, ConditionalField 表
- 新增 IntakeForm, TriagePanel, CommandPalette, TimeTrackPanel, RecurrenceConfig, AISettingsPanel 组件
- 代码规模: 27 models, 26 services, 27 handlers, 57 components, 16 views

---

## 2026-06-26 — AI 智能助手实现 + 知识库全量同步

**类型**: Feature + Documentation

**变更内容**:
- Task 20-23: AI Infrastructure (LLM Client + AIService + AIHandler + 前端)
- LLM Client 双协议支持 (Anthropic + OpenAI-compatible/DeepSeek)
- 17 个 AI Tool Functions 映射到已有 API
- AIChatSidebar (SSE 流式对话) + AICreateDialog (NL→工作项)
- DeepSeek 为默认 Provider
- 知识库全量同步：架构文档、数据模型、前端架构更新至实际状态
- 清理已过时的 "Python 后端独有模型" 列表
- 清理已过时的模块状态表

**影响文件**:
- 新增 10 个 AI 文件 (model/ai_config.go, service/llm_client.go, service/ai_service.go, handler/ai_handler.go, types/ai.ts, api/ai.ts, composables/useAI.ts, components/AIChatSidebar.vue, components/AICreateDialog.vue)
- 修改 4 个现有文件 (config.go, router.go, main.go, Project.vue)
- KB 全面更新: architecture/README.md, data-model.md, frontend.md

---

## 2026-06-25 — 工作项管理关键功能补全 (5 Features)

**类型**: Feature

**变更内容**:
- Feature G: 统一 Project Settings (4→12 tabs)
- Feature F: Notification 后端
- Feature A: Saved Views 保存视图
- Feature C: 项目级 Issue Type 配置
- Feature B: Pages/Wiki 文档系统

**KB 更新**: architecture/README.md, data-model.md, frontend.md + 3 新设计文档

---

## 2026-06-25 — ProjectTemplateManager 集成修复

**类型**: Fix

将孤儿组件 ProjectTemplateManager.vue 集成到 WorkspaceSettings。

---

## 2026-06-21 — 文档体系重组

**类型**: 文档重组

建立 kb/dev/superseded 三层架构，从 TECH_ARCHITECTURE.md 拆分 7 个专题文档。
