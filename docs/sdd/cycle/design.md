# SDD Design - Cycle（周期）模块

**文档版本**: v1.0  
**创建日期**: 2026-06-14  
**功能模块**: Cycle（周期/迭代）  
**状态**: ✅ 已完成

---

## 1. 功能概述

### 1.1 模块描述

Cycle（周期）是用于组织和管理迭代工作的功能，类似于Scrum中的Sprint。它帮助团队在固定的时间盒内完成一组相关的工作项，支持周期规划、进度跟踪、报告生成等功能。

### 1.2 核心功能

| 功能 | 描述 | 优先级 |
|------|--------|--------|
| 创建周期 | 创建新的迭代周期 | P0 |
| 分配工作项 | 将工作项分配到周期 | P0 |
| 开始/结束周期 | 控制周期的活跃状态 | P0 |
| 周期进度 | 跟踪周期完成率 | P1 |
| 周期报告 | 燃尽图、统计视图 | P1 |
| 周期模板 | 复用周期配置 | P2 |

### 1.3 数据模型

#### 1.3.1 Cycle Model

```
Cycle Table
├── id (UUID) - 主键
├── name (String[255]) - 周期名称
├── description (Text) - 周期描述
├── start_date (Date) - 开始日期
├── end_date (Date) - 结束日期
├── status (String[30]) - 状态: upcoming/active/completed
├── progress (Float) - 进度百分比
├── project_id (UUID) - 所属项目
├── workspace_id (UUID) - 所属工作空间
├── created_at (DateTime)
├── updated_at (DateTime)
└── is_deleted (Boolean) - 软删除标记
```

#### 1.3.2 CycleIssue Model（关联表）

```
CycleIssue Table
├── id (UUID) - 主键
├── cycle_id (UUID) - 周期ID
├── issue_id (UUID) - 工作项ID
├── created_at (DateTime)
└── is_deleted (Boolean)
```

### 1.4 状态流转

```
Upcoming → Active → Completed
  ↑         ↓
  └─────────┴───── Cancelled
```

---

## 2. API接口设计

### 2.1 Endpoints

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/cycles` | 创建周期 |
| GET | `/cycles` | 列出周期 |
| GET | `/cycles/{cycle_id}` | 获取周期详情 |
| PUT | `/cycles/{cycle_id}` | 更新周期 |
| DELETE | `/cycles/{cycle_id}` | 删除周期 |
| POST | `/cycles/{cycle_id}/start` | 开始周期 |
| POST | `/cycles/{cycle_id}/end` | 结束周期 |
| POST | `/cycles/{cycle_id}/issues` | 添加工作项到周期 |
| DELETE | `/cycles/{cycle_id}/issues/{issue_id}` | 从周期移除工作项 |
| GET | `/cycles/{cycle_id}/progress` | 获取周期进度 |
| GET | `/cycles/{cycle_id}/report` | 获取周期报告 |

### 2.2 Schema定义

```python
class CycleCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description: Optional[str] = None
    start_date: date
    end_date: date

class CycleUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None

class CycleResponse(BaseModel):
    id: UUID
    name: str
    description: Optional[str]
    start_date: date
    end_date: date
    status: str
    progress: float
    total_issues: int
    completed_issues: int
    # ...
```

---

## 3. 前端类型定义

```typescript
export interface Cycle {
  id: string
  name: string
  description?: string
  start_date: string
  end_date: string
  status: 'upcoming' | 'active' | 'completed'
  progress: number
  total_issues: number
  completed_issues: number
  project_id: string
  workspace_id: string
}

export enum CycleStatus {
  UPCOMING = 'upcoming',
  ACTIVE = 'active',
  COMPLETED = 'completed'
}
```

---

## 4. 业务逻辑

### 4.1 创建周期

1. 验证日期范围有效（开始日期 < 结束日期）
2. 验证项目存在
3. 创建周期记录

### 4.2 开始周期

1. 验证周期状态为upcoming
2. 更新状态为active
3. 更新started_at时间戳

### 4.3 结束周期

1. 验证周期状态为active
2. 更新状态为completed
3. 更新completed_at时间戳
4. 计算最终进度

### 4.4 进度计算

```
progress = (completed_issues / total_issues) * 100
```

---

## 5. 依赖关系

- `app.models.cycle.Cycle` - 周期模型
- `app.models.cycle.CycleIssue` - 关联模型
- `app.models.issue.Issue` - 工作项模型
- Cycle关联Issue需要先完成Issue模块

---

## 6. TODO

- [ ] 创建Cycle Schema
- [ ] 创建Cycle Service
- [ ] 创建Cycle API Endpoints
- [ ] 创建前端TypeScript类型
- [ ] 创建前端API模块
- [ ] 创建前端组件

---

**文档作者**: AI Assistant  
**审核状态**: 待审核
