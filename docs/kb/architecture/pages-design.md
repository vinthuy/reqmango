# Pages（页面文档/Wiki）设计

**最后更新**: 2026-10-25

## 概述

Pages 是项目内的文档/Wiki 系统，支持富文本编辑、层级组织和归档管理。专为需求管理场景中的 PRD、技术方案、会议记录等文档需求设计。

## 数据模型

```go
type Page struct {
    BaseModel
    Title       string    // 页面标题
    Content     string    // Markdown/HTML 内容
    ContentJSON *string   // JSONB: TipTap 编辑器 JSON 格式（已启用）
    Published   bool      // 是否发布
    ArchivedAt  *time.Time // 归档时间
    Sequence    int       // 同级排序
    ParentID    *uint64   // 父页面（树形层级）
    Depth       int       // 根为 0。创建时 depth = 父级 + 1，大于 5 则拒绝
    LockedByID  *uint64   // 编辑锁
    LockedAt    *time.Time
    ProjectID   uint64
    WorkspaceID uint64
    SourceType  string    // 来源，如 cycle、issue；可空
    SourceID    *uint64
}
```

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/projects/:id/pages` | 列出所有页面（可含已归档） |
| POST | `/projects/:id/pages` | 创建页面（支持 parent_id） |
| GET | `/projects/:id/pages/tree` | 获取层级树结构 |
| GET | `/projects/:id/pages/:pageId` | 获取单个页面 |
| PUT | `/projects/:id/pages/:pageId` | 更新页面内容/标题 |
| DELETE | `/projects/:id/pages/:pageId` | 软删除 |
| POST | `/projects/:id/pages/:pageId/archive` | 归档 |
| POST | `/projects/:id/pages/:pageId/restore` | 恢复 |
| POST | `/projects/:id/pages/:pageId/move` | 移动（重新设置父级和排序） |
| GET | `/projects/:projectId/pages/search` | 搜索，参数 `q` |
| GET | `/projects/:projectId/pages/by-source` | 按 `source_type`、`source_id` 列出 |
| POST | `/projects/:projectId/pages/:pageId/lock` | 加锁 |
| POST | `/projects/:projectId/pages/:pageId/unlock` | 解锁 |
| GET | `/projects/:projectId/pages/:pageId/export` | 导出 |
| POST | `/projects/:projectId/pages/:pageId/convert-to-issue` | 转成工作项 |
| GET | `/projects/:projectId/pages/:pageId/children` | 直接子页面 |

## 层级规则

- 创建时子页面 `depth = 父级 + 1`，`depth > 5` 拒绝（错误文案是 “Maximum page nesting depth is 5”）。根页面 depth 为 0，因此存库的 depth 最大是 5
- 页面不能是自己的父级（防止循环引用）
- 移动时自动重新计算 `depth` 字段
- 同级页面按 `sequence` 字段排序

## 归档机制

- 归档通过设置 `archived_at` 时间戳实现
- 归档的页面不在 tree 中显示，但可通过 `include_archived=true` 参数查询
- 子页面不受父页面归档影响

## 前端

- `ProjectPages.vue` — 主视图（左侧页面树 + 右侧编辑器）
- `PageTree.vue` — 递归层级树组件，支持 hover 操作（新建子页面、删除）
- 编辑器使用 TipTap 富文本编辑器（content_json 字段已启用），通过 `TipTapEditor.vue` 组件提供所见即所得编辑体验
- 通过路由 `/workspace/:slug/project/:id/pages` 访问

## 设计决策

1. **深度限制**：depth ≤ 5 防止过深的嵌套导致性能问题和糟糕的 UX
2. **归档而非删除**：默认流程是归档（可恢复），删除是软删除（GORM DeletedAt）。「从工作空间复制类型」属于类型模板，不在页面模型里

## 页面版本管理

页面支持版本历史追踪，由以下模块实现：

**后端：**
- `page_version.go` (model) — PageVersion 模型，存储页面版本历史
- `page_version_service.go` — 版本管理业务逻辑
- `page_version_handler.go` — HTTP 处理器

**API 端点：**

版本在更新页面时由 `PageVersionService.SaveVersion` 写入，没有单独的「创建版本」接口。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/projects/:projectId/pages/:pageId/versions` | 版本列表 |
| GET | `/projects/:projectId/pages/:pageId/versions/:versionId` | 指定版本 |
| POST | `/projects/:projectId/pages/:pageId/versions/:versionId/restore` | 恢复该版本 |

**前端：**
- `PageVersionPanel.vue` — 版本历史面板，展示版本列表和恢复操作
- `PageVersionDiff.vue` — 版本对比视图，可视化两个版本之间的差异

## 页面模板

页面支持模板功能，便于复用常见文档结构：

**后端：**
- `page_template.go` (model) — PageTemplate 模型
- `page_template_service.go` — 模板业务逻辑
- `page_template_handler.go` — HTTP 处理器

**API 端点：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/projects/:projectId/page-templates` | 列出 |
| POST | `/projects/:projectId/page-templates` | 创建 |
| GET | `/projects/:projectId/page-templates/:templateId` | 单条 |
| PUT | `/projects/:projectId/page-templates/:templateId` | 更新 |
| DELETE | `/projects/:projectId/page-templates/:templateId` | 删除 |

模板可以挂在工作空间（`project_id` 为空）或某个项目上。字段是 name、description、content、content_json、is_default。

**前端：**
- `PageTemplateSelector.vue` — 模板选择器，在创建页面时可从模板库中选择预设模板
