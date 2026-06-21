# SDD Design - Module（模块）模块

**文档版本**: v1.0  
**创建日期**: 2026-06-14  
**功能模块**: Module（模块/功能分组）  
**状态**: ✅ 已完成

---

## 1. 功能概述

### 1.1 模块描述

Module（模块）用于按功能或业务领域对工作项进行分组。与Cycle的时间维度不同，Module是功能维度的组织方式，帮助团队将大型项目拆分为更小的功能模块。

### 1.2 核心功能

| 功能 | 描述 | 优先级 |
|------|--------|--------|
| 创建模块 | 创建功能模块 | P0 |
| 添加工作项 | 将工作项纳入模块 | P0 |
| 模块进度 | 查看模块整体完成度 | P1 |
| 模块成员 | 指定负责人 | P1 |
| 模块时间线 | 设置目标和截止日期 | P2 |

### 1.3 数据模型

#### 1.3.1 Module Model

```
Module Table
├── id (UUID) - 主键
├── name (String[255]) - 模块名称
├── description (Text) - 模块描述
├── color (String[7]) - 颜色代码
├── start_date (Date) - 开始日期
├── target_date (Date) - 目标日期
├── status (String[30]) - 状态
├── project_id (UUID) - 所属项目
├── workspace_id (UUID) - 所属工作空间
├── created_at (DateTime)
├── updated_at (DateTime)
└── is_deleted (Boolean) - 软删除标记
```

#### 1.3.2 ModuleIssue Model（关联表）

```
ModuleIssue Table
├── id (UUID) - 主键
├── module_id (UUID) - 模块ID
├── issue_id (UUID) - 工作项ID
├── created_at (DateTime)
└── is_deleted (Boolean)
```

### 1.4 与Cycle的区别

| 维度 | Cycle | Module |
|------|-------|--------|
| 组织方式 | 时间维度 | 功能维度 |
| 生命周期 | 固定时间盒 | 持续存在 |
| 典型用途 | 迭代冲刺 | 功能分组 |
| 状态流转 | 有状态流转 | 相对静态 |

---

## 2. API接口设计

### 2.1 Endpoints

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/modules` | 创建模块 |
| GET | `/modules` | 列出模块 |
| GET | `/modules/{module_id}` | 获取模块详情 |
| PUT | `/modules/{module_id}` | 更新模块 |
| DELETE | `/modules/{module_id}` | 删除模块 |
| POST | `/modules/{module_id}/issues` | 添加工作项到模块 |
| DELETE | `/modules/{module_id}/issues/{issue_id}` | 从模块移除工作项 |
| GET | `/modules/{module_id}/issues` | 获取模块工作项列表 |
| GET | `/modules/{module_id}/progress` | 获取模块进度 |

### 2.2 Schema定义

```python
class ModuleCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description: Optional[str] = None
    color: Optional[str] = None
    start_date: Optional[date] = None
    target_date: Optional[date] = None

class ModuleUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    color: Optional[str] = None
    start_date: Optional[date] = None
    target_date: Optional[date] = None
    status: Optional[str] = None

class ModuleResponse(BaseModel):
    id: UUID
    name: str
    description: Optional[str]
    color: Optional[str]
    start_date: Optional[date]
    target_date: Optional[date]
    status: str
    progress: float
    total_issues: int
    completed_issues: int
    project_id: UUID
    workspace_id: UUID
    # ...
```

---

## 3. 前端类型定义

```typescript
export interface Module {
  id: string
  name: string
  description?: string
  color?: string
  start_date?: string
  target_date?: string
  status: string
  progress: number
  total_issues: number
  completed_issues: number
  project_id: string
  workspace_id: string
}

export interface ModuleCreate {
  name: string
  description?: string
  color?: string
  start_date?: string
  target_date?: string
}

export interface ModuleUpdate {
  name?: string
  description?: string
  color?: string
  start_date?: string
  target_date?: string
  status?: string
}
```

---

## 4. 业务逻辑

### 4.1 创建模块

1. 验证项目存在
2. 验证模块名称唯一
3. 创建模块记录

### 4.2 添加工作项到模块

1. 验证模块存在
2. 验证工作项存在
3. 验证工作项属于同一项目
4. 创建关联记录

### 4.3 进度计算

```
progress = (completed_issues / total_issues) * 100
```

---

## 5. 依赖关系

- `app.models.module.Module` - 模块模型
- `app.models.module.ModuleIssue` - 关联模型
- `app.models.issue.Issue` - 工作项模型
- Module关联Issue需要先完成Issue模块

---

## 6. TODO

- [ ] 创建Module Schema
- [ ] 创建Module Service
- [ ] 创建Module API Endpoints
- [ ] 创建前端TypeScript类型
- [ ] 创建前端API模块
- [ ] 创建前端组件

---

**文档作者**: AI Assistant  
**审核状态**: 待审核
