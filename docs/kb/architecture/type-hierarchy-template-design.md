# 类型层级 & 模板架构设计

> 参考 PlaneAI Enterprise Grid 设计，适配当前项目

---

## 1. PlaneAI 参考模型

### 1.1 类型层级 (Hierarchy)

PlaneAI 的类型层级由 Workspace Admin 在工作空间级别定义：

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
POST   /api/v1/projects/:id/issue-types/import    从模板导入类型
```

### 3.4 层级查询

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
| 字段绑定 | ✅ | ✅ |
| 层级定义 (Level/ParentTypeID) | ✅ | ✅ |
| 项目模板 CRUD | ✅ | ⏳ |
| 模板添加类型 | ✅ | ⏳ |
| 应用到项目(实例化) | ⏳ | ⏳ |
| 层级校验 (创建/编辑) | ✅ | ⏳ |
| 层级树查询 | ⏳ | ⏳ |
| 项目模板 UI | ⏳ | ⏳ |

---

## 5. 待实现优先级

1. **模板实例化** (apply to project) — 核心闭环
2. **项目模板管理 UI** — 可用性
3. **层级树查询** — 展示
4. **关系系统** (依赖/关联) — 后续增强
