# SDD Design - Issue（工作项）模块

**文档版本**: v1.0  
**创建日期**: 2026-06-14  
**功能模块**: Issue（工作项）  
**状态**: ✅ 已完成

---

## 1. 功能概述

### 1.1 模块描述

Issue（工作项）是项目管理平台的核心实体，用于表示任务、缺陷、用户故事等各类工作内容。每个工作项都属于特定项目，支持状态流转、优先级管理、负责人分配、标签分类等功能。

### 1.2 核心功能

| 功能 | 描述 | 优先级 |
|------|------|--------|
| 创建工作项 | 在项目中创建Issue/Task/Bug/Story/Epic | P0 |
| 更新工作项 | 修改工作项的各种属性 | P0 |
| 删除工作项 | 软删除工作项 | P0 |
| 归档/恢复 | 归档和恢复工作项 | P1 |
| 活动历史 | 记录工作项的变更历史 | P0 |
| 统计信息 | 项目工作项统计 | P1 |
| 搜索功能 | 跨项目搜索工作项 | P1 |
| 批量操作 | 批量更新/删除工作项 | P1 |

### 1.3 数据模型

#### 1.3.1 Issue Model（数据库模型）

```
Issue Table
├── id (UUID) - 主键
├── name (String[255]) - 工作项名称
├── description_html (Text) - HTML格式描述
├── description_json (JSON) - JSON格式描述
├── description_stripped (Text) - 纯文本描述
├── priority (String[30]) - 优先级
├── sequence_id (Integer) - 项目内序号
├── sort_order (Float) - 排序权重
├── start_date (Date) - 开始日期
├── target_date (Date) - 目标日期
├── completed_at (DateTime) - 完成时间
├── is_draft (Boolean) - 是否草稿
├── archived_at (Date) - 归档时间
├── project_id (UUID) - 所属项目
├── workspace_id (UUID) - 所属工作空间
├── parent_id (UUID) - 父工作项
├── state_id (UUID) - 当前状态
├── external_id (String[255]) - 外部系统ID
├── external_source (String[255]) - 外部系统来源
├── created_at (DateTime) - 创建时间
├── updated_at (DateTime) - 更新时间
└── is_deleted (Boolean) - 软删除标记
```

#### 1.3.2 IssueActivity Model（活动历史）

```
IssueActivity Table
├── id (UUID) - 主键
├── issue_id (UUID) - 关联工作项
├── verb (String[255]) - 动作类型
├── field (String[255]) - 变更字段
├── old_value (Text) - 旧值
├── new_value (Text) - 新值
├── comment (Text) - 评论内容
├── actor_id (UUID) - 操作人
├── created_at (DateTime) - 创建时间
└── is_deleted (Boolean) - 软删除标记
```

### 1.4 状态流转

```
Backlog → Todo → In Progress → In Review → Done
  ↓        ↓         ↓
  └────────┴─────────┴────→ Cancelled
```

---

## 2. API接口设计

### 2.1 RESTful API Endpoints

#### 2.1.1 工作项 CRUD

| 方法 | 路径 | 描述 | 请求体 | 响应 |
|------|------|------|--------|------|
| POST | `/issues` | 创建工作项 | IssueCreate | IssueResponse |
| GET | `/issues` | 列出工作项 | - | List[IssueResponse] |
| GET | `/issues/{issue_id}` | 获取工作项详情 | - | IssueResponse |
| PUT | `/issues/{issue_id}` | 更新工作项 | IssueUpdate | IssueResponse |
| DELETE | `/issues/{issue_id}` | 删除工作项 | - | 204 No Content |

#### 2.1.2 工作项状态

| 方法 | 路径 | 描述 | 请求体 | 响应 |
|------|------|------|--------|------|
| POST | `/issues/{issue_id}/archive` | 归档工作项 | - | 204 No Content |
| POST | `/issues/{issue_id}/restore` | 恢复工作项 | - | IssueResponse |

#### 2.1.3 活动与统计

| 方法 | 路径 | 描述 | 请求体 | 响应 |
|------|------|------|--------|------|
| GET | `/issues/{issue_id}/activities` | 获取活动历史 | - | List[Activity] |
| GET | `/issues/statistics` | 获取项目统计 | - | IssueStatistics |
| GET | `/issues/search` | 搜索工作项 | - | List[IssueSearchResult] |

#### 2.1.4 批量操作

| 方法 | 路径 | 描述 | 请求体 | 响应 |
|------|------|------|--------|------|
| POST | `/issues/bulk/update` | 批量更新 | BulkUpdate | List[IssueResponse] |
| POST | `/issues/bulk/delete` | 批量删除 | BulkDelete | 204 No Content |

### 2.2 Schema定义

#### 2.2.1 IssueCreate

```python
class IssueCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description_html: str = "<p></p>"
    description_json: Dict[str, Any] = {}
    priority: IssuePriority = IssuePriority.NONE
    start_date: Optional[date] = None
    target_date: Optional[date] = None
    parent_id: Optional[UUID] = None
    state_id: Optional[UUID] = None
    assignee_ids: Optional[List[UUID]] = []
    label_ids: Optional[List[UUID]] = []
    estimate_point_id: Optional[UUID] = None
    type_id: Optional[UUID] = None
    external_id: Optional[str] = None
    external_source: Optional[str] = None
```

#### 2.2.2 IssueUpdate

```python
class IssueUpdate(BaseModel):
    name: Optional[str] = None
    description_html: Optional[str] = None
    priority: Optional[IssuePriority] = None
    state_id: Optional[UUID] = None
    assignee_ids: Optional[List[UUID]] = None
    label_ids: Optional[List[UUID]] = None
    start_date: Optional[date] = None
    target_date: Optional[date] = None
    estimate_point_id: Optional[UUID] = None
    cycle_id: Optional[UUID] = None
    module_ids: Optional[List[UUID]] = None
```

#### 2.2.3 IssueResponse

```python
class IssueResponse(AuditSchema, SoftDeleteSchema, IssueBase):
    id: UUID
    project: ProjectLite
    sequence_id: int
    state_id: UUID
    state_name: str
    state_group: str
    assignees: List[UserLite]
    labels: List[UUID]
    sub_issues_count: int
    link_count: int
    attachment_count: int
    completed_at: Optional[datetime] = None
    is_draft: bool = False
    parent_id: Optional[UUID] = None
    estimate_point_id: Optional[UUID] = None
    cycle_id: Optional[UUID] = None
    module_ids: List[UUID] = []
```

---

## 3. 前端类型定义

### 3.1 TypeScript Interfaces

```typescript
// 优先级枚举
export enum IssuePriority {
  URGENT = 'urgent',
  HIGH = 'high',
  MEDIUM = 'medium',
  LOW = 'low',
  NONE = 'none'
}

// 创建请求
export interface IssueCreate {
  name: string
  description_html?: string
  description_json?: Record<string, any>
  priority: IssuePriority
  start_date?: string
  target_date?: string
  parent_id?: string
  state_id?: string
  assignee_ids?: string[]
  label_ids?: string[]
  estimate_point_id?: string
  type_id?: string
  external_id?: string
  external_source?: string
}

// 更新请求
export interface IssueUpdate {
  name?: string
  description_html?: string
  priority?: IssuePriority
  state_id?: string
  assignee_ids?: string[]
  label_ids?: string[]
  start_date?: string
  target_date?: string
  estimate_point_id?: string
  cycle_id?: string
  module_ids?: string[]
}

// 响应
export interface IssueResponse extends IssueBase {
  id: string
  sequence_id: number
  project_id: string
  state_id: string
  state_name: string
  state_group: string
  assignees: UserLite[]
  labels: string[]
  sub_issues_count: number
  completed_at?: string
  is_draft: boolean
  // ...
}
```

---

## 4. 业务逻辑

### 4.1 创建工作项

1. 验证项目存在且未被删除
2. 验证状态存在（如提供）
3. 如未提供状态，使用项目默认状态（Backlog）
4. 生成下一个sequence_id
5. 创建工作项记录
6. 记录创建活动
7. 处理assignees、labels关联（预留）

### 4.2 更新工作项

1. 验证工作项存在
2. 更新指定字段
3. 如果状态变更：
   - 验证新状态存在
   - 如果变为Done状态，记录completed_at
4. 记录每个字段的变更活动

### 4.3 删除工作项

1. 验证工作项存在
2. 执行软删除（设置is_deleted=True）
3. 保留数据可恢复

### 4.4 归档工作项

1. 验证工作项存在
2. 设置archived_at时间戳
3. 从常规列表中隐藏

---

## 5. 辅助函数

### 5.1 优先级映射

| 优先级 | 显示名称 | 颜色代码 |
|--------|----------|----------|
| urgent | 紧急 | #EF4444 |
| high | 高 | #F59E0B |
| medium | 中 | #3B82F6 |
| low | 低 | #10B981 |
| none | 无 | #6B7280 |

### 5.2 状态分组

| 分组 | 显示名称 | 颜色代码 |
|------|----------|----------|
| backlog | 待办 | #6B7280 |
| todo | 计划中 | #3B82F6 |
| in_progress | 进行中 | #F59E0B |
| done | 已完成 | #10B981 |
| cancelled | 已取消 | #EF4444 |

---

## 6. 依赖关系

### 6.1 后端依赖

- `app.models.issue.Issue` - 工作项模型
- `app.models.issue.IssueActivity` - 活动历史模型
- `app.models.project.Project` - 项目模型
- `app.models.state.State` - 状态模型
- `app.schemas.issue` - Schema定义

### 6.2 前端依赖

- `app/types/issue.ts` - TypeScript类型
- `app/api/issue.ts` - API调用模块
- Pinia Store（待创建）

---

## 7. 已知限制与TODO

### 7.1 已实现

- ✅ 基本CRUD操作
- ✅ 软删除和恢复
- ✅ 归档和解档
- ✅ 活动历史记录
- ✅ 统计信息
- ✅ 批量操作

### 7.2 待实现

- ⏳ Assignee关联（需要IssueAssignee模型）
- ⏳ Label关联（需要IssueLabel模型）
- ⏳ Cycle关联（需要IssueCycle模型）
- ⏳ Module关联（需要ModuleIssue模型）
- ⏳ 估算点关联（需要EstimatePoint模型）
- ⏳ 附件管理
- ⏳ 评论功能
- ⏳ 父子工作项层级展示
- ⏳ 排序拖拽

---

**文档作者**: AI Assistant  
**审核状态**: 待审核
