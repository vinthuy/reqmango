# 类型层级 & 模板架构设计

> 参考企业级平台设计，适配当前项目

**最后更新**: 2026-07-04

---

## 1. 企业级平台参考模型

### 1.1 类型层级 (Hierarchy)

企业级平台的类型层级由 Workspace Admin 在工作空间级别定义：

| Level | 类型示例 | 角色 |
|-------|----------|------|
| 2 | Epic | 大功能/交付物 |
| 1 | Story, Bug | 独立工作单元 |
| 0 | Task, Spike | 叶子执行项 |

**核心规则**：
- 多个类型可共享同一层级
- 子工作项只能是"直接下一层"的类型
- 一旦启用不可禁用
- 类型变更时自动校验层级合法性

### 1.2 类型管理

- **Enterprise Grid**: 类型由 Workspace Admin 集中定义，项目从工作空间导入
- **Pro/Business**: 类型在项目级别管理
- 每种类型可绑定自定义属性（字段）
- 类型导入到项目后，Workspace 变更自动传播

### 1.3 模板系统

| 模板类型 | 级别 | 包含内容 |
|----------|------|----------|
| Work Item 模板 | 项目/工作空间 | 标题格式、描述、字段默认值、标签、子工作项 |
| Project 模板 | 工作空间 | 状态、标签、类型、初始工作项 |
| Page 模板 | 项目/工作空间 | 页面布局、内容结构 |

---

## 2. 当前项目架构设计

### 2.1 数据模型

```
┌──────────────────────────────────────────────────────────────┐
│                    WORKSPACE LEVEL                           │
│                                                              │
│  ┌─────────────────┐    ┌──────────────────────────┐        │
│  │ CustomField      │    │ IssueTypeTemplate        │        │
│  │ (字段库)         │    │ (类型蓝图)               │        │
│  │ - name          │    │ - name, color, icon      │        │
│  │ - field_type    │    │ - level (0-5)            │        │
│  │ - workspace_id  │    │ - parent_type_id (层级)  │        │
│  │                 │    │ - workspace_id            │        │
│  └────────┬────────┘    └──────────┬───────────────┘        │
│           │                        │                         │
│           │  ┌─────────────────────┘                         │
│           │  │                                               │
│           ▼  ▼                                               │
│  ┌─────────────────────────┐                                │
│  │ IssueTypeTemplateField  │  类型↔字段绑定                  │
│  │ - template_type_id      │                                │
│  │ - field_id              │                                │
│  │ - is_required           │                                │
│  └─────────────────────────┘                                │
│                                                              │
│  ┌──────────────────────────┐                               │
│  │ ProjectTemplate          │  项目模板                      │
│  │ - name, description      │                                │
│  │ - workspace_id           │                               │
│  └──────────┬───────────────┘                               │
│             │                                                 │
│             ▼                                                 │
│  ┌──────────────────────────┐                               │
│  │ ProjectTemplateType      │  模板↔类型关联                 │
│  │ - template_id            │                                │
│  │ - type_template_id       │                                │
│  │ - is_required            │                                │
│  └──────────────────────────┘                               │
└──────────────────────────────────────────────────────────────┘
                              │
                              │ Apply to Project
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                    PROJECT LEVEL                              │
│                                                              │
│  ┌─────────────────┐    ┌──────────────────────────┐        │
│  │ Project          │    │ IssueType                │        │
│  │ - template_id    │    │ (从模板实例化)            │        │
│  │                 │    │ - template_type_id (溯源) │        │
│  └─────────────────┘    │ - level, parent_type_id   │        │
│                         └──────────┬───────────────┘        │
│                                    │                         │
│  ┌─────────────────┐               │                         │
│  │ Issue           │               │                         │
│  │ - parent_id     │               │                         │
│  │ - depth (0-5)   │               │                         │
│  │ - issue_type_id │───────────────┘                         │
│  └─────────────────┘                                        │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 数据流

```
1. Workspace Admin 定义自定义字段
   POST /custom-fields → workspace 字段库

2. Workspace Admin 创建类型模板(蓝图)
   POST /type-templates → 定义 name/level/icon
   POST /type-templates/:id/fields → 绑定字段

3. Workspace Admin 创建项目模板
   POST /templates → 创建模板
   POST /templates/:id/types → 添加类型模板

4. Project Admin 应用模板到项目
   POST /templates/:id/apply → 实例化:
     a. 从类型模板创建 IssueType (复制 name/color/level/fields)
     b. 建立项目内层级关系
     c. 复制字段绑定
     d. 设置项目 template_id
```

### 2.3 层级校验规则

```
创建子工作项时:
  1. 父工作项存在 (parent_id 有效)
  2. 子类型.level == 父类型.level + 1 (只能下一层)
  3. 如果子类型有 parent_type_id 约束，父类型必须匹配
  4. depth < 5 (最多6层)

类型变更时:
  1. 如果工作项有子项，新类型的 level 必须当前所有子项的 level - 1
  2. 如果工作项有父项，新类型必须满足父项 level + 1
```

---

## 3. API 设计

### 3.1 类型模板 (Workspace 级)

```
POST   /api/v1/type-templates?workspace_id=N    创建
GET    /api/v1/type-templates?workspace_id=N    列表(按level排序)
GET    /api/v1/type-templates/:id                详情(含字段)
PUT    /api/v1/type-templates/:id                更新
DELETE /api/v1/type-templates/:id                删除
POST   /api/v1/type-templates/:id/fields         绑定字段
DELETE /api/v1/type-templates/:id/fields/:fid    解绑字段
```

### 3.2 项目模板

```
POST   /api/v1/templates?workspace_id=N          创建
GET    /api/v1/templates?workspace_id=N          列表
GET    /api/v1/templates/:id                      详情(含关联类型)
PUT    /api/v1/templates/:id                      更新
DELETE /api/v1/templates/:id                      删除
POST   /api/v1/templates/:id/types                添加类型模板
DELETE /api/v1/templates/:id/types/:typeId        移除
POST   /api/v1/templates/:id/apply                应用到项目(实例化)
```

### 3.3 项目内类型管理

```
GET    /api/v1/projects/:id/issue-types            项目内已有类型
POST   /api/v1/projects/:id/issue-types            项目内创建类型
POST   /api/v1/projects/:id/issue-types/copy-from-workspace  从工作空间复制
PATCH  /api/v1/projects/:id/issue-types/reorder    项目内排序
```

### 3.4 工作空间级别类型管理

```
POST   /api/v1/issue-types?workspace_id=N          创建类型
GET    /api/v1/issue-types?workspace_id=N          列表
GET    /api/v1/issue-types/:typeId                 详情
PUT    /api/v1/issue-types/:typeId                 更新
DELETE /api/v1/issue-types/:typeId                 删除
PATCH  /api/v1/issue-types/:typeId/disable         启用/禁用
PATCH  /api/v1/issue-types/reorder-workspace?workspace_id=N  工作空间级别排序
```

### 3.5 类型字段绑定

```
GET    /api/v1/issue-types/:typeId/fields          获取已绑定字段
POST   /api/v1/issue-types/:typeId/fields          添加字段绑定
PUT    /api/v1/issue-types/:typeId/fields/:fieldId 更新字段绑定（切换必填等）
DELETE /api/v1/issue-types/:typeId/fields/:fieldId 移除字段绑定
```

### 3.6 层级查询

```
GET    /api/v1/issues/hierarchy?project_id=N       项目工作项层级树
GET    /api/v1/issues/:id/children                  子工作项列表(按层级)
```

---

## 4. 当前完成度

| 功能 | 后端 | 前端 |
|------|------|------|
| 自定义字段 CRUD | ✅ | ✅ |
| 类型模板 CRUD | ✅ | ✅ TypeTemplateManager |
| 字段绑定 | ✅ | ✅ IssueTypeList页面 |
| 层级定义 (Level/ParentTypeID) | ✅ | ✅ |
| 项目模板 CRUD | ✅ | ✅ |
| 模板添加类型 | ✅ | ✅ |
| 应用到项目(实例化) | ✅ | ✅ |
| 层级校验 (创建/编辑) | ✅ | ✅ |
| 层级树查询 | ✅ | ✅ |
| 项目模板 UI | ✅ | ✅ |
| 工作空间级别排序 | ✅ | ✅ |
| 启用/禁用类型 | ✅ | ✅ |
| 从工作空间复制到项目 | ✅ | ✅ |

---

## 5. 最新改进 (2026-07-04)

### IssueTypeList页面增强
- 新增自定义字段绑定UI
- 支持显示已绑定字段列表
- 支持添加新字段绑定
- 支持移除字段绑定
- 支持切换字段必填状态

### 模板应用修复
- `CopyFromWorkspace` 方法现在会正确复制字段关联
- 从工作空间复制类型时，会同时复制 `IssueTypeField` 关联

### 工作空间级别排序
- 新增 `ReorderWorkspace` API：`PATCH /issue-types/reorder-workspace?workspace_id=`
- 支持工作空间级别类型的排序操作

---

## 6. 待实现优先级

1. ~~**模板实例化** (apply to project) — 核心闭环~~ ✅ 已完成
2. ~~**项目模板管理 UI** — 可用性~~ ✅ 已完成
3. ~~**层级树查询** — 展示~~ ✅ 已完成
4. **关系系统** (依赖/关联) — 后续增强
5. **批量创建/导入类型** — 提升效率
6. **类型使用统计** — 显示每个类型关联的Issue数量
