# SDD - 工作项空间增删改查及自定义工作项功能

## 1. 概述

### 1.1 背景
基于 PlanAI (Makeplane) 的功能特性，为 reqmango 项目完善工作项（Issue）管理功能，实现完整的增删改查操作，以及自定义工作项类型和自定义字段与类型的关联功能。

### 1.2 目标
1. 完善工作项的创建、编辑、删除、查看功能
2. 实现工作项类型（IssueType）的管理功能
3. 实现自定义字段与工作项类型的关联绑定
4. 提供统一的工作项操作体验

---

## 2. 功能需求

### 2.1 工作项类型管理

#### 2.1.1 类型列表页
- 展示项目中所有工作项类型
- 显示类型名称、图标、颜色、是否默认、状态
- 支持创建、编辑、删除、禁用类型

#### 2.1.2 类型创建/编辑
- 类型名称（必填）
- 类型图标（预设图标库）
- 类型颜色（预设色板）
- 是否默认类型
- 顺序/优先级
- 启用/禁用状态

#### 2.1.3 类型与自定义字段关联
- 在类型详情页查看已关联的自定义字段
- 添加/移除与类型关联的自定义字段
- 设置字段在特定类型中的必填性

### 2.2 工作项创建

#### 2.2.1 创建页面/模态框
- 类型选择器（显示类型图标和颜色）
- 标题输入（必填）
- 描述编辑器
- 快速设置：
  - 状态选择（基于类型的工作流）
  - 优先级
  - 负责人
  - 周期
  - 模块
  - 开始/截止日期
- 自定义字段表单（根据类型显示关联字段）

#### 2.2.2 快速创建
- 在工作项列表顶部提供快捷创建入口
- 简化的表单（仅标题和类型）

### 2.3 工作项编辑

#### 2.3.1 详情页面
- 完整的属性编辑面板
- 左侧：主要信息（标题、描述、子工作项）
- 右侧：属性面板（可折叠）
- 自定义字段显示与编辑
- 活动历史记录

#### 2.3.2 类型切换
- 支持在工作项详情页切换类型
- 切换后更新显示的自定义字段

### 2.4 工作项删除与归档

#### 2.4.1 删除
- 确认对话框
- 软删除（标记为已删除）

#### 2.4.2 归档
- 归档而非删除
- 支持从归档中恢复

---

## 3. 技术设计

### 3.1 后端 API 设计

#### 3.1.1 工作项类型 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/projects/:project_id/issue-types` | 获取项目所有工作项类型 |
| POST | `/api/v1/projects/:project_id/issue-types` | 创建工作项类型 |
| GET | `/api/v1/projects/:project_id/issue-types/:id` | 获取单个类型详情 |
| PUT | `/api/v1/projects/:project_id/issue-types/:id` | 更新工作项类型 |
| DELETE | `/api/v1/projects/:project_id/issue-types/:id` | 删除工作项类型 |
| PATCH | `/api/v1/projects/:project_id/issue-types/:id/disable` | 禁用/启用类型 |

#### 3.1.2 类型-字段关联 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/projects/:project_id/issue-types/:id/fields` | 获取类型关联的字段 |
| POST | `/api/v1/projects/:project_id/issue-types/:id/fields` | 关联字段到类型 |
| DELETE | `/api/v1/projects/:project_id/issue-types/:id/fields/:field_id` | 移除字段关联 |

#### 3.1.3 工作项 API 扩展

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/v1/issues` | 创建工作项（支持 type_id） |
| PUT | `/api/v1/issues/:id` | 更新工作项 |
| DELETE | `/api/v1/issues/:id` | 删除工作项 |
| PATCH | `/api/v1/issues/:id/archive` | 归档/取消归档 |
| POST | `/api/v1/issues/:id/convert-type` | 转换工作项类型 |

### 3.2 数据模型

#### 3.2.1 IssueType 模型

```go
type IssueType struct {
    BaseModel
    Name        string `gorm:"not null;size:255"`
    Color       string `gorm:"size:50;default:'#6B7280'"`
    Icon        string `gorm:"size:50;default:'circle'"`
    IsDefault   bool   `gorm:"default:false"`
    Sequence    int    `gorm:"default:1"`
    IsActive    bool   `gorm:"default:true"`
    ProjectID   uint64 `gorm:"not null;index"`
    WorkspaceID uint64 `gorm:"not null;index"`
    CreatedByID uint64
    UpdatedByID *uint64
    Fields      []IssueTypeField `gorm:"foreignKey:TypeID"`
}

type IssueTypeField struct {
    BaseModel
    TypeID       uint64 `gorm:"not null;index"`
    FieldID      uint64 `gorm:"not null;index"`
    IsRequired   bool   `gorm:"default:false"`
    Sequence     int    `gorm:"default:1"`
}
```

### 3.3 前端组件设计

#### 3.3.1 组件结构

```
src/
├── views/
│   ├── IssueTypeList.vue       # 类型列表页
│   ├── IssueCreate.vue         # 创建工作项页
│   └── IssueDetail.vue         # 编辑工作项页（已有）
├── components/
│   ├── IssueTypeSelector.vue   # 类型选择器
│   ├── IssueTypeCard.vue       # 类型卡片
│   ├── IssueTypeForm.vue       # 类型表单（创建/编辑）
│   └── IssueQuickCreate.vue    # 快速创建组件
└── api/
    └── issue-type.ts           # 类型 API
```

#### 3.3.2 路由设计

```typescript
{
  path: '/workspaces/:workspaceId/projects/:projectId/issue-types',
  name: 'IssueTypeList',
  component: IssueTypeList
},
{
  path: '/workspaces/:workspaceId/projects/:projectId/issues/new',
  name: 'IssueCreate',
  component: IssueCreate
}
```

---

## 4. 实施计划

### Phase 1: 基础数据模型
1. 创建 IssueType 模型和迁移
2. 创建 IssueTypeField 关联模型
3. 实现类型 CRUD API

### Phase 2: 前端类型管理
1. 创建类型 API 模块
2. 创建类型列表页面
3. 创建类型表单组件
4. 在项目设置中添加类型管理入口

### Phase 3: 工作项创建/编辑完善
1. 创建工作项创建页面
2. 集成类型选择器
3. 根据类型加载自定义字段
4. 实现类型切换功能

### Phase 4: 功能完善
1. 实现快速创建
2. 实现归档功能
3. 批量操作（可选）

---

## 5. 数据库变更

### 5.1 新增表

#### issue_types
| 列名 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| name | varchar(255) | 类型名称 |
| color | varchar(50) | 颜色 |
| icon | varchar(50) | 图标 |
| is_default | boolean | 是否默认 |
| sequence | integer | 排序 |
| is_active | boolean | 是否启用 |
| project_id | bigint | 所属项目 |
| workspace_id | bigint | 所属工作空间 |

#### issue_type_fields
| 列名 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| type_id | bigint | 类型ID |
| field_id | bigint | 字段ID |
| is_required | boolean | 是否必填 |
| sequence | integer | 排序 |

---

## 6. UI 设计

### 6.1 类型选择器
- 下拉选择或网格展示
- 显示图标、颜色、名称
- 支持搜索过滤

### 6.2 工作项创建页
- 顶部：类型选择 + 标题输入
- 左侧：描述编辑器
- 右侧：属性面板（折叠）
- 底部：自定义字段表单

### 6.3 类型管理页
- 卡片网格布局
- 每个卡片显示：图标、名称、字段数量、状态
- 操作按钮：编辑、禁用、删除

---

## 7. 验收标准

1. ✅ 可以创建、编辑、删除工作项类型
2. ✅ 可以为类型关联自定义字段
3. ✅ 创建工作项时可选择类型
4. ✅ 根据选择的类型显示对应的自定义字段
5. ✅ 可以切换工作项的类型
6. ✅ 工作项详情页正确显示所有字段
7. ✅ API 和前端类型匹配，无编译错误
8. ✅ 集成测试通过
