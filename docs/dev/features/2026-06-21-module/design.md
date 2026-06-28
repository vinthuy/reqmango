# Module（模块）管理功能设计文档

**版本**: v1.0  
**创建日期**: 2026-06-21  
**状态**: 已确认  

---

## 1. 概述

参考 Plane 模块管理特性，完善本项目（reqmango）的 Module 管理功能。Module 是项目内的工作项分组单元，支持父子层级（树形结构），一个 Issue 可同时属于多个 Module。

### 1.1 功能范围

- Module CRUD（创建/编辑/删除/列表）
- 模块树形结构（父子层级）
- Issue-Module 多对多关联（添加/移除/列表）
- 模块进度统计
- 侧滑详情面板（Project Tab 集成）

### 1.2 不包含

- 模块独立详情页（路由）
- 模块创建向导
- 模块归档/取消归档功能优化

---

## 2. 后端 API 设计

### 2.1 路由清单

```
# 现有（保留）
POST   /api/v1/modules?workspace_id=N             创建模块
GET    /api/v1/modules?project_id=&workspace_id=   列表（含 include_archived）
GET    /api/v1/modules/:moduleId                   详情
PUT    /api/v1/modules/:moduleId                   更新
DELETE /api/v1/modules/:moduleId                   软删除

# 新增
POST   /api/v1/modules/:moduleId/issues?issue_id=N 添加Issue到模块
DELETE /api/v1/modules/:moduleId/issues/:issueId    从模块移除Issue
GET    /api/v1/modules/:moduleId/issues             模块内Issue列表
GET    /api/v1/modules/:moduleId/progress           模块进度
GET    /api/v1/modules/:moduleId/statistics         模块统计
GET    /api/v1/modules/tree?project_id=N            模块树形结构
```

### 2.2 关键设计决策

- **多对多关联**：新增 `module_issues` 关联表，一个 Issue 可属于多个 Module
- **树形构建**：后端递归构建 `ModuleTreeNode`，子模块按 `order` 排序
- **软删除**：通过 `BaseModel.DeletedAt` 实现
- **错误处理**：统一使用 `common.AppError` 模式

---

## 3. 数据模型

### 3.1 Module 模型（微调）

```go
type Module struct {
    BaseModel
    Name        string     `gorm:"type:varchar(100);not null" json:"name"`
    Description string     `gorm:"type:text" json:"description"`
    ProjectID   uint64     `gorm:"not null;index" json:"project_id"`
    WorkspaceID uint64     `gorm:"not null;index" json:"workspace_id"`
    ParentID    *uint64    `gorm:"index" json:"parent_id"`
    Order       int        `gorm:"default:0" json:"order"`
    ArchivedAt  *time.Time `json:"archived_at"`
    IsArchived  bool       `gorm:"default:false" json:"is_archived"`

    // 关系
    Project  Project       `gorm:"foreignKey:ProjectID" json:"-"`
    IssueLinks []ModuleIssue `gorm:"foreignKey:ModuleID" json:"-"`
}
```

### 3.2 ModuleIssue 关联表（新增）

```go
type ModuleIssue struct {
    ModuleID uint64 `gorm:"primaryKey;autoIncrement:false"`
    IssueID  uint64 `gorm:"primaryKey;autoIncrement:false"`
    Module   Module `gorm:"foreignKey:ModuleID;constraint:OnDelete:CASCADE"`
    Issue    Issue  `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE"`
}
```

---

## 4. Service 层设计

### 4.1 新增方法

```go
// Issue 关联
func AddIssue(moduleID, issueID uint64) error
func RemoveIssue(moduleID, issueID uint64) error
func ListIssues(moduleID uint64, stateID *uint64, priority string, limit, offset int) ([]response.IssueResponse, int64, error)

// 分析
func GetProgress(moduleID uint64) (*response.ModuleProgress, error)
func GetStatistics(moduleID uint64) (*response.ModuleStatistics, error)

// 树形
func BuildTree(projectID uint64) ([]*response.ModuleTreeNode, error)
```

### 4.2 现有方法改动

- `Create`：增加 `userID` 参数，校验 workspace/project 存在
- `Delete`：改为软删除（GORM 自动处理）
- `List`：返回 `(items, total, error)` 三元组
- 全部改为 `*common.AppError` 错误模式

### 4.3 关键业务逻辑

- **AddIssue**：校验 Module/Issue 存在且属于同一 project，防重复添加
- **BuildTree**：查询项目所有模块 → 按 `parent_id` 分组 → 递归构建树，子节点按 `order` 排序
- **GetProgress**：JOIN `module_issues` + `states` 统计完成数和状态分布
- **软删除**：GORM 自动设置 `deleted_at`，CASCADE 清理 `module_issues`

---

## 5. 前端设计

### 5.1 组件架构

```
Project.vue Tab "模块"
  ├─ ModuleList（store 驱动，卡片/树形视图）
  │   ├─ ModuleCard → click → ModuleDetailPanel
  │   └─ ModuleTree → click → ModuleDetailPanel
  └─ ModuleDetailPanel（侧滑面板）
      ├─ 头部：名称 + 编辑/删除按钮
      ├─ 进度卡片（总数/完成/进度%）
      └─ Issue 列表 + 添加/移除
```

ModuleFormModal 内联弹窗用于创建/编辑，不独立路由。

### 5.2 Pinia Store

```typescript
stores/module.ts:
  State: modules, moduleTree, currentModule, progress, moduleIssues, isLoading, error
  Actions:
    fetchModules(projectId, workspaceId)
    fetchModuleTree(projectId)
    createModule(data) → created
    updateModule(id, data) → updated
    deleteModule(id)
    addIssueToModule(moduleId, issueId)
    removeIssueFromModule(moduleId, issueId)
    fetchModuleIssues(moduleId, filters?)
    fetchProgress(moduleId)
```

### 5.3 文件变更清单

**Go 后端**

| 文件 | 操作 | 说明 |
|------|------|------|
| `model/module.go` | 修改 | 添加 ModuleIssue 关联表，添加关系字段 |
| `service/module_service.go` | 修改 | 新增 6 个方法，改为 AppError + 软删除 |
| `handler/module_handler.go` | 修改 | 新增 6 个 handler，改为 AppError |
| `router/router.go` | 修改 | 新增 6 条路由 |

**Vue 前端**

| 文件 | 操作 | 说明 |
|------|------|------|
| `stores/module.ts` | 新建 | Pinia store |
| `components/ModuleDetailPanel.vue` | 新建 | 侧滑面板 |
| `components/ModuleFormModal.vue` | 新建 | 创建/编辑弹窗 |
| `components/ModuleList.vue` | 修改 | store 驱动 |
| `api/module.ts` | 微调 | URL 对齐 |
| `views/Project.vue` | 微调 | 集成面板 |

---

**文档路径**: `docs/pages/2026-06-21-module-feature-design.md`
