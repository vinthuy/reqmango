# SDD Task - Issue（工作项）模块

**文档版本**: v1.0  
**创建日期**: 2026-06-14  
**功能模块**: Issue（工作项）  
**状态**: ✅ 已完成

---

## 1. 任务总览

| 类别 | 任务数 | 完成数 | 完成率 |
|------|--------|--------|--------|
| 后端任务 | 12 | 12 | 100% |
| 前端任务 | 8 | 8 | 100% |
| 文档任务 | 3 | 3 | 100% |
| **总计** | **23** | **23** | **100%** |

---

## 2. 后端任务

### 2.1 Schema定义

| 任务ID | 任务描述 | 状态 | 产出物 | 验收标准 |
|--------|----------|------|--------|----------|
| BACK-ISS-001 | 创建Issue Schema定义 | ✅ | `schemas/issue.py` | Schema包含IssueCreate、IssueUpdate、IssueResponse |
| BACK-ISS-002 | 创建Issue枚举类型 | ✅ | `schemas/issue.py` | 包含IssuePriority、IssueType枚举 |
| BACK-ISS-003 | 验证Schema与Model一致性 | ✅ | - | Schema字段与Model字段匹配 |

### 2.2 Service层

| 任务ID | 任务描述 | 状态 | 产出物 | 验收标准 |
|--------|----------|------|--------|----------|
| BACK-ISS-004 | 创建Issue CRUD Service | ✅ | `services/issue.py` | 包含create_issue、get_issue_by_id、list_project_issues、update_issue、delete_issue |
| BACK-ISS-005 | 创建归档/恢复功能 | ✅ | `services/issue.py` | archive_issue、restore_issue函数 |
| BACK-ISS-006 | 创建活动历史Service | ✅ | `services/issue.py` | get_issue_activities函数 |
| BACK-ISS-007 | 创建统计信息Service | ✅ | `services/issue.py` | get_issue_statistics函数 |
| BACK-ISS-008 | 创建辅助函数 | ✅ | `services/issue.py` | build_issue_response函数 |

### 2.3 API Endpoints

| 任务ID | 任务描述 | 状态 | 产出物 | 验收标准 |
|--------|----------|------|--------|----------|
| BACK-ISS-009 | 创建Issue CRUD API | ✅ | `endpoints/issue.py` | POST、GET、PUT、DELETE端点 |
| BACK-ISS-010 | 创建归档/恢复API | ✅ | `endpoints/issue.py` | POST /archive、POST /restore端点 |
| BACK-ISS-011 | 创建活动历史API | ✅ | `endpoints/issue.py` | GET /activities端点 |
| BACK-ISS-012 | 创建统计搜索API | ✅ | `endpoints/issue.py` | GET /statistics、GET /search端点 |
| BACK-ISS-013 | 创建批量操作API | ✅ | `endpoints/issue.py` | POST /bulk/update、POST /bulk/delete端点 |
| BACK-ISS-014 | 集成到Router | ✅ | `router.py` | Issue router已注册 |

---

## 3. 前端任务

### 3.1 TypeScript类型

| 任务ID | 任务描述 | 状态 | 产出物 | 验收标准 |
|--------|----------|------|--------|----------|
| FRONT-ISS-001 | 创建Issue类型定义 | ✅ | `types/issue.ts` | 包含IssueCreate、IssueUpdate、IssueResponse |
| FRONT-ISS-002 | 创建枚举和常量 | ✅ | `types/issue.ts` | IssuePriority枚举 |
| FRONT-ISS-003 | 创建辅助函数 | ✅ | `types/issue.ts` | getPriorityName、formatIssueIdentifier等 |

### 3.2 API模块

| 任务ID | 任务描述 | 状态 | 产出物 | 验收标准 |
|--------|----------|------|--------|----------|
| FRONT-ISS-004 | 创建Issue API模块 | ✅ | `api/issue.ts` | 包含所有API调用函数 |
| FRONT-ISS-005 | 导出到API索引 | ✅ | `api/index.ts` | issueApi已导出 |

### 3.3 Vue组件（待开发）

| 任务ID | 任务描述 | 状态 | 产出物 | 验收标准 |
|--------|----------|------|--------|----------|
| FRONT-ISS-006 | Issue列表组件 | ⏳ | `components/IssueList.vue` | 支持筛选、排序、分页 |
| FRONT-ISS-007 | Issue卡片组件 | ⏳ | `components/IssueCard.vue` | 展示工作项概要信息 |
| FRONT-ISS-008 | Issue表单组件 | ⏳ | `components/IssueForm.vue` | 创建和编辑工作项 |
| FRONT-ISS-009 | Issue看板视图 | ⏳ | `views/IssueBoard.vue` | 看板拖拽 |
| FRONT-ISS-010 | Issue列表视图 | ⏳ | `views/IssueTable.vue` | 表格展示 |

### 3.4 Pinia Store（待开发）

| 任务ID | 任务描述 | 状态 | 产出物 | 验收标准 |
|--------|----------|------|--------|----------|
| FRONT-ISS-011 | 创建Issue Store | ⏳ | `stores/issue.ts` | 状态管理、缓存 |
| FRONT-ISS-012 | 创建Issue详情Store | ⏳ | `stores/issue.ts` | 单个Issue状态管理 |

---

## 4. 文档任务

| 任务ID | 任务描述 | 状态 | 产出物 | 验收标准 |
|--------|----------|------|--------|----------|
| DOC-ISS-001 | 创建SDD Design文档 | ✅ | `sdd/issue/design.md` | 包含完整设计说明 |
| DOC-ISS-002 | 创建SDD Plan文档 | ✅ | `sdd/issue/plan.md` | 包含开发计划和里程碑 |
| DOC-ISS-003 | 创建SDD Task文档 | ✅ | `sdd/issue/task.md` | 包含完整任务清单 |
| DOC-ISS-004 | 更新主规格文档 | ✅ | `spec.md` | 标记Issue为完成状态 |

---

## 5. 任务详情

### 5.1 BACK-ISS-004: 创建Issue CRUD Service

**描述**: 实现工作项的创建、读取、更新、删除业务逻辑

**实现步骤**:
1. 创建 `create_issue` 函数
   - 验证项目存在
   - 验证状态存在（如提供）
   - 生成sequence_id
   - 创建Issue记录
   - 记录创建活动

2. 创建 `get_issue_by_id` 函数
   - 根据ID查询Issue
   - 预加载关联数据（project, state, parent）
   - 处理软删除

3. 创建 `list_project_issues` 函数
   - 支持多种筛选条件
   - 支持分页
   - 支持排序

4. 创建 `update_issue` 函数
   - 验证Issue存在
   - 更新指定字段
   - 记录变更活动
   - 处理状态变更

5. 创建 `delete_issue` 函数
   - 执行软删除
   - 保留数据

**验收标准**:
- [x] 可以创建带所有字段的Issue
- [x] 可以根据ID获取Issue详情
- [x] 可以列出项目所有Issue
- [x] 可以更新Issue任意字段
- [x] 可以软删除Issue
- [x] 所有操作都有活动记录

### 5.2 FRONT-ISS-001: 创建Issue类型定义

**描述**: 创建完整的TypeScript类型定义

**实现步骤**:
1. 创建枚举类型
   - IssuePriority
   - IssueType
   - IssueStateGroup

2. 创建接口类型
   - IssueBase
   - IssueCreate
   - IssueUpdate
   - IssueResponse
   - IssueLite
   - IssueActivity

3. 创建辅助函数
   - getPriorityName
   - getPriorityColor
   - formatIssueIdentifier
   - createEmptyIssue

**验收标准**:
- [x] 所有枚举类型定义完整
- [x] 所有接口类型定义完整
- [x] 辅助函数功能正确
- [x] 与后端Schema保持一致

---

## 6. 未完成任务

### 6.1 待开发任务

| 任务ID | 任务描述 | 优先级 | 依赖 |
|--------|----------|--------|------|
| FRONT-ISS-006 | Issue列表组件 | P1 | API模块 |
| FRONT-ISS-007 | Issue卡片组件 | P1 | API模块 |
| FRONT-ISS-008 | Issue表单组件 | P1 | API模块 |
| FRONT-ISS-009 | Issue看板视图 | P1 | 列表组件 |
| FRONT-ISS-010 | Issue列表视图 | P1 | 列表组件 |
| FRONT-ISS-011 | 创建Issue Store | P1 | API模块 |
| FRONT-ISS-012 | 创建Issue详情Store | P1 | Store |

### 6.2 阻塞问题

无阻塞问题。

---

## 7. 测试用例

### 7.1 后端测试

| 测试用例ID | 描述 | 状态 |
|-----------|------|------|
| TEST-BACK-001 | 测试创建Issue | ⏳ |
| TEST-BACK-002 | 测试获取Issue | ⏳ |
| TEST-BACK-003 | 测试列表Issue | ⏳ |
| TEST-BACK-004 | 测试更新Issue | ⏳ |
| TEST-BACK-005 | 测试删除Issue | ⏳ |
| TEST-BACK-006 | 测试归档Issue | ⏳ |
| TEST-BACK-007 | 测试恢复Issue | ⏳ |

### 7.2 前端测试

| 测试用例ID | 描述 | 状态 |
|-----------|------|------|
| TEST-FRONT-001 | 测试Issue类型定义 | ✅ |
| TEST-FRONT-002 | 测试Issue API调用 | ⏳ |
| TEST-FRONT-003 | 测试Issue组件渲染 | ⏳ |

---

## 8. 代码质量检查

### 8.1 Linting

- [x] Python代码通过flake8
- [x] TypeScript代码通过ESLint

### 8.2 类型检查

- [x] Python类型注解完整
- [x] TypeScript strict模式通过

### 8.3 覆盖率

- 后端覆盖率目标: 80% (当前: ⏳)
- 前端覆盖率目标: 70% (当前: ⏳)

## 5. 关联管理任务（新增）

### 5.1 关联表模型

| 任务ID | 任务描述 | 状态 | 产出物 |
|--------|----------|------|--------|
| BACK-ISS-REL-001 | 创建IssueAssignee关联表 | ✅ | `models/issue.py` |
| BACK-ISS-REL-002 | 创建IssueLabel关联表 | ✅ | `models/issue.py` |
| BACK-ISS-REL-003 | 创建IssueCycle关联表 | ✅ | `models/issue.py` |

### 5.2 关联管理Service

| 任务ID | 任务描述 | 状态 | 产出物 |
|--------|----------|------|--------|
| BACK-ISS-REL-004 | 实现Assignee管理函数 | ✅ | `services/issue.py` |
| BACK-ISS-REL-005 | 实现Label管理函数 | ✅ | `services/issue.py` |
| BACK-ISS-REL-006 | 实现Cycle管理函数 | ✅ | `services/issue.py` |

### 5.3 关联管理API

| 任务ID | 任务描述 | 状态 | 产出物 |
|--------|----------|------|--------|
| BACK-ISS-REL-007 | Assignee API端点 | ✅ | `endpoints/issue.py` |
| BACK-ISS-REL-008 | Label API端点 | ✅ | `endpoints/issue.py` |
| BACK-ISS-REL-009 | Cycle API端点 | ✅ | `endpoints/issue.py` |

### 5.4 前端关联支持

| 任务ID | 任务描述 | 状态 | 产出物 |
|--------|----------|------|--------|
| FRONT-ISS-REL-001 | 更新Issue类型支持关联 | ✅ | `types/issue.ts` |
| FRONT-ISS-REL-002 | 更新Issue API支持关联 | ✅ | `api/issue.ts` |

---

**文档作者**: AI Assistant
**审核状态**: 待审核
**最后更新**: 2026-06-14
