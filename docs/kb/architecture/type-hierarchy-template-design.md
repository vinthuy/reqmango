# 类型层级 & 模板架构设计

> 参考企业级平台设计，适配当前项目

**最后更新**: 2026-10-25

---

## 1. 企业级平台参考模型

这一节是当初对照企业级工具写的参照，不是当前校验规则。当前规则在 2.3。

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

### 2.3 创建子工作项时的校验

代码在 `IssueService` 创建和更新路径里做这些检查，不是「子类型 level 必须正好等于父类型 level + 1」：

1. 父工作项必须存在，不能把自己或子孙设成父级。
2. 父级 `depth >= 5` 时拒绝。子项 `depth = 父级 + 1`，因此 depth 取值 0–5。错误文案是 “Maximum hierarchy depth (6 levels) exceeded”。
3. 子类型 `level == 0` 且 `parent_type_id` 为空时，不能作为子工作项。
4. 父类型 `level > 0` 时，子类型的 level 不能比父类型更浅。更深或相同都允许。
5. 子类型写了 `parent_type_id` 时，父工作项的类型必须就是这个 id。
6. 父类型的 `allowed_child_type_ids` 非空时，子类型必须在这个名单里。

`IssueType.level` 的模型注释是 L0–L5，0 表示顶层。这和第一节里「Epic 在 level 2」的企业参考不是同一套编号。

工作项树的接口是 `GET /issues/tree?project_id=`，子项懒加载是 `GET /issues/:issueId/children`。没有 `/issues/hierarchy`。

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

### 3.3 项目里的类型

项目不复制工作空间类型，而是用 `issue_type_imports` 引用。没有 `copy-from-workspace` 这条路由。

```
GET    /projects/:projectId/issue-types
POST   /projects/:projectId/issue-types
PATCH  /projects/:projectId/issue-types/reorder
GET    /projects/:projectId/issue-types/importable
POST   /projects/:projectId/issue-types/:typeId/import
DELETE /projects/:projectId/issue-types/:typeId/import
```

创建、排序、导入、取消导入要 `settings:manage`。

创建工作项时套用的标题和描述模板是另一张表 `work_item_templates`，路由是 `/projects/:projectId/work-item-templates`。它不是项目模板 `project_templates`。

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
GET    /issues/tree?project_id=N
GET    /issues/:issueId/children
```

---

## 4. 和代码的对应

| 能力 | 代码 |
|------|------|
| 工作空间自定义字段 | `/custom-fields`，另有 `/conditional-fields` |
| 类型蓝图 | `/type-templates`，字段绑定是 `POST/DELETE /:id/fields` |
| 项目模板 | `/templates`，含 `/:templateId/types` 和 `/:templateId/apply` |
| 工作空间类型 | `/issue-types`，含 `PATCH /:typeId/disable` 和 `PATCH /reorder-workspace` |
| 类型上的字段 | `/issue-types/:typeId/fields` |
| 项目引用工作空间类型 | `/projects/:projectId/issue-types/:typeId/import` |
| 工作项模板 | `/projects/:projectId/work-item-templates` |
| 层级校验 | `IssueService` 创建和更新，规则见 2.3 |
| 关联类型 | 见 [relation-system-design.md](relation-system-design.md)，已经实现 |
