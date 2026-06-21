# RQL (Reqman Query Language) 设计文档

**最后更新**: 2026-06-22
**状态**: 已批准

---

## 1. 概述

RQL 是一种类 SQL 的查询语言，用于在 Reqman 项目管理工具中筛选和查找工作项、周期、模块等实体。RQL 支持三种查询方式：文本查询、可视化构建器、AI 自然语言解释器。

### 1.1 设计目标

- 提供强大而直观的查询能力
- 支持复杂条件组合 (AND/OR/括号分组)
- 零外部依赖，纯前端实现
- 查询可保存和复用

### 1.2 技术选型

| 组件 | 技术方案 |
|------|----------|
| 词法分析 | 手写词法分析器 (Tokenize) |
| 语法分析 | 递归下降_parser 或 Pratt Parser |
| AI 解释器 | 前端规则引擎 (无外部 API) |
| 存储 | localStorage |

---

## 2. RQL 语法

### 2.1 词法规范

| Token 类型 | 描述 | 示例 |
|-----------|------|------|
| `IDENTIFIER` | 字段名 | `state`, `priority`, `assignee` |
| `STRING` | 字符串值 | `"待处理"`, `'张三'` |
| `NUMBER` | 数字 | `123`, `45.67` |
| `DATE` | 日期 | `2026-01-01` |
| `DATETIME` | 日期时间 | `2026-01-01 12:00:00` |
| `OPERATOR` | 操作符 | `=`, `!=`, `>`, `<`, `>=`, `<=` |
| `LIKE` | 模糊匹配 | `LIKE` |
| `IN` | 集合查询 | `IN (...)` |
| `AND` | 逻辑与 | `AND` |
| `OR` | 逻辑或 | `OR` |
| `NOT` | 逻辑非 | `NOT` |
| `LPAREN` | 左括号 | `(` |
| `RPAREN` | 右括号 | `)` |
| `COMMA` | 逗号 | `,` |

### 2.2 支持字段

#### 2.2.1 工作项 (Issue) 字段

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `id` | number | 工作项 ID | `id = 123` |
| `sequence_id` | string | 序列号 | `sequence_id LIKE "DEMO-"` |
| `name` | string | 标题 | `name LIKE "登录"` |
| `description` | string | 描述 | `description LIKE "%bug%"` |
| `state` | string | 状态 | `state = "待处理"` |
| `priority` | enum | 优先级 | `priority = "high"` |
| `assignee` | string | 负责人 | `assignee = "张三"` |
| `reporter` | string | 创建者 | `reporter = "张三"` |
| `label` | string | 标签 | `label = "bug"` |
| `cycle` | string | 所属周期 | `cycle = "Sprint 1"` |
| `module` | string | 所属模块 | `module = "前端"` |
| `created_at` | datetime | 创建时间 | `created_at > "2026-01-01"` |
| `updated_at` | datetime | 更新时间 | `updated_at < "2026-06-01"` |
| `due_date` | date | 截止日期 | `due_date = "2026-06-30"` |

#### 2.2.2 周期 (Cycle) 字段

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `name` | string | 周期名称 | `name LIKE "Sprint"` |
| `status` | enum | 状态 | `status = "active"` |
| `start_date` | date | 开始日期 | `start_date > "2026-01-01"` |
| `end_date` | date | 结束日期 | `end_date < "2026-06-30"` |

#### 2.2.3 模块 (Module) 字段

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `name` | string | 模块名称 | `name LIKE "用户"` |
| `parent` | string | 父模块 | `parent = "后端"` |

### 2.3 语法规则

```ebnf
query       ::= or_expr
or_expr     ::= and_expr ( "OR" and_expr )*
and_expr    ::= not_expr ( "AND" not_expr )*
not_expr    ::= "NOT" not_expr | primary
primary     ::= comparison | "(" query ")"
comparison  ::= field operator value
field       ::= IDENTIFIER
operator    ::= "=" | "!=" | ">" | "<" | ">=" | "<=" | "LIKE" | "IN"
value       ::= STRING | NUMBER | DATE | DATETIME | in_list
in_list     ::= "(" value ( "," value )* ")"
```

### 2.4 示例

```sql
-- 基础比较
state = "待处理"
priority = "high"
created_at > "2026-01-01"

-- 模糊匹配
name LIKE "登录"
description LIKE "%bug%"

-- IN 查询
state IN ("待处理", "进行中", "已完成")

-- 逻辑组合
state = "待处理" AND priority = "high"
assignee = "张三" OR assignee = "李四"

-- 复杂组合 (使用括号)
(state = "待处理" OR state = "进行中") AND priority = "high" AND assignee = "张三"

-- 多条件
label = "bug" AND (priority = "high" OR priority = "urgent")

-- 自然语言等效 (Phase 3)
"张三负责的高优先级待处理工作项"
= assignee = "张三" AND priority = "high" AND state = "待处理"
```

---

## 3. 架构设计

### 3.1 组件结构

```
src/
├── components/
│   ├── RQL/
│   │   ├── RQLInput.vue          # RQL 文本输入框
│   │   ├── RQLVisualBuilder.vue  # 可视化构建器
│   │   ├── RQLAIRecognizer.vue   # AI 自然语言解释
│   │   ├── RQLSuggest.vue         # 语法提示/补全
│   │   ├── RQLHistory.vue        # 查询历史
│   │   └── RQLSavedQueries.vue   # 保存的查询
│   └── GlobalSearch.vue           # 全局搜索入口
├── composables/
│   └── useRQL.ts                 # RQL 核心逻辑
├── utils/
│   └── rql/
│       ├── lexer.ts              # 词法分析器
│       ├── parser.ts             # 语法分析器
│       ├── executor.ts           # 查询执行器
│       ├── interpreter.ts        # AI 解释器
│       ├── validator.ts          # 语法验证
│       └── types.ts              # 类型定义
└── stores/
    └── rql.ts                    # RQL 状态管理
```

### 3.2 核心类设计

#### RQLLexer

```typescript
interface Token {
  type: TokenType
  value: string
  position: number
}

class RQLLexer {
  tokenize(input: string): Token[]
  hasErrors(): boolean
  getErrors(): string[]
}
```

#### RQLParser

```typescript
interface ASTNode {
  type: 'BinaryExpr' | 'Comparison' | 'InExpr' | 'LikeExpr'
}

interface BinaryExpr extends ASTNode {
  type: 'BinaryExpr'
  operator: 'AND' | 'OR'
  left: ASTNode
  right: ASTNode
}

interface Comparison extends ASTNode {
  type: 'Comparison'
  field: string
  operator: string
  value: string | number | Date
}

class RQLParser {
  parse(tokens: Token[]): ASTNode
  hasErrors(): boolean
  getErrors(): string[]
}
```

#### RQLExecutor

```typescript
interface QueryParams {
  search?: string
  state_id?: number
  priority?: string
  assignee_id?: number
  cycle_id?: number
  module_id?: number
  label_ids?: number[]
  // ... 其他参数
}

class RQLExecutor {
  execute(ast: ASTNode, context: QueryContext): QueryParams
}
```

#### AIRQLInterpreter

```typescript
class AIRQLInterpreter {
  // 规则映射表
  private rules: AIRule[]

  // 关键词到字段的映射
  private keywordMap: Map<string, string>

  interpret(naturalText: string): string | null // 返回 RQL 或 null
}
```

### 3.3 数据流

```
用户输入 (文本/可视化/自然语言)
         │
         ▼
┌─────────────────┐
│   RQLInput      │ ←→ RQLSuggest (语法提示)
└────────┬────────┘
         │ parse()
         ▼
┌─────────────────┐
│   RQLLexer      │ → Token[]
└────────┬────────┘
         │ parse()
         ▼
┌─────────────────┐
│   RQLParser     │ → AST
└────────┬────────┘
         │ execute()
         ▼
┌─────────────────┐
│   RQLExecutor   │ → QueryParams
└────────┬────────┘
         │ buildQuery()
         ▼
┌─────────────────┐
│   API Client   │ → Server
└─────────────────┘
```

---

## 4. AI 自然语言解释

### 4.1 规则引擎设计

不依赖外部 LLM API，使用前端规则引擎实现。

#### 4.1.1 意图识别规则

```typescript
interface AIRule {
  // 匹配模式 (正则或关键词)
  pattern: RegExp | string[]

  // 字段映射
  field: string

  // 值映射
  valueMap?: Record<string, string>

  // 优先级 (数字越小优先级越高)
  priority: number
}

// 示例规则
const rules: AIRule[] = [
  {
    pattern: /负责人[是为]*(.+)/,
    field: 'assignee',
    valueMap: { '我': 'current_user' },
    priority: 1
  },
  {
    pattern: /(张三|李四|王五)/,
    field: 'assignee',
    priority: 2
  },
  {
    pattern: /(高|紧急)优先级/,
    field: 'priority',
    valueMap: { '高': 'high', '紧急': 'urgent' },
    priority: 3
  },
  {
    pattern: /(待处理|进行中|已完成)/,
    field: 'state',
    priority: 4
  },
  {
    pattern: /(bug|缺陷)/,
    field: 'label',
    valueMap: { 'bug': 'bug', '缺陷': 'bug' },
    priority: 5
  }
]
```

#### 4.1.2 解释流程

```
输入: "张三负责的高优先级待处理工作项"
         │
         ▼
┌─────────────────┐
│ 分词/断句       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 规则匹配        │
├─────────────────┤
│ "张三" → assignee │
│ "高优先级" → priority = high  │
│ "待处理" → state          │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 组合 RQL        │
└────────┬────────┘
         │
         ▼
输出: assignee = "张三" AND priority = "high" AND state = "待处理"
```

### 4.2 预定义短语映射

| 中文短语 | RQL |
|---------|-----|
| 我负责的 | assignee = "current_user" |
| 我创建的 | reporter = "current_user" |
| 我参与的 | assignee = "current_user" OR reporter = "current_user" |
| 未分配的 | assignee IS NULL |
| 已完成的 | state = "已完成" |
| 进行中的 | state = "进行中" |
| 高优先级 | priority = "high" |
| 紧急的 | priority = "urgent" |
| 过期的 | due_date < today |
| 今天到期 | due_date = today |
| 本周到期 | due_date >= start_of_week AND due_date <= end_of_week |
| 最近创建的 | created_at >= "-7d" |
| 最近更新的 | updated_at >= "-3d" |

---

## 5. 存储设计

### 5.1 localStorage 结构

```typescript
interface RQLStorage {
  // 查询历史
  history: {
    id: string
    rql: string
    timestamp: number
    entityType: 'issue' | 'cycle' | 'module'
    context: { projectId: number }
  }[]

  // 保存的查询
  saved: {
    id: string
    name: string
    rql: string
    entityType: 'issue' | 'cycle' | 'module'
    context: { projectId: number }
    createdAt: number
  }[]

  // 用户偏好
  preferences: {
    defaultEntity: 'issue' | 'cycle' | 'module'
    defaultView: 'list' | 'kanban'
    autoSave: boolean
  }
}
```

### 5.2 存储限制

- 历史记录最多保存 50 条
- 保存的查询最多 20 条
- 超出时自动清理最旧记录

---

## 6. UI 设计

### 6.1 RQL 输入框组件

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 🔍 state = "待处理" AND priority = "high"              [AI] [保存] [历史] │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─ 语法提示 ────────────────────────────────────────────────────────┐ │
│  │ 可用字段: state, priority, assignee, label, cycle, module, name...   │ │
│  │ 操作符: =, !=, >, <, >=, <=, LIKE, IN, AND, OR, NOT                │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌─ 语法错误 ────────────────────────────────────────────────────────┐ │
│  │ ✗ 期望操作符 (=, LIKE, IN)，遇到: identifier                       │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6.2 可视化构建器

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 条件 (AND)                                              [+ 添加条件]  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─ 条件 1 ─────────────────────────────────────────────────────────┐ │
│  │ [字段 ▼]     [操作符 ▼]     [值输入框]           [× 删除]          │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌─ 条件组 A ────────────────────────────────────────────────────────┐ │
│  │ 条件 (OR)                                          [+ 添加条件]    │ │
│  │ ┌─ 条件 2 ─────────────────────────────────────────────────────┐  │ │
│  │ │ [state ▼]   [= ▼]     [待处理 ▼]             [× 删除]        │  │ │
│  │ └────────────────────────────────────────────────────────────────┘  │ │
│  │ ┌─ 条件 3 ─────────────────────────────────────────────────────┐  │ │
│  │ │ [state ▼]   [= ▼]     [进行中 ▼]             [× 删除]        │  │ │
│  │ └────────────────────────────────────────────────────────────────┘  │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  转换后的 RQL: state IN ("待处理", "进行中") AND priority = "high"      │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6.3 AI 解释器面板

```
┌─────────────────────────────────────────────────────────────────────────┐
│  🤖 AI 智能查询                                                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  请输入您的需求（支持中文）                                              │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │ 显示张三负责的所有高优先级 bug                                   │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  [转换为 RQL]  [直接搜索]                                               │
│                                                                         │
│  ── 转换结果 ────────────────────────────────────────────────────────── │
│                                                                         │
│  ✅ 理解为您想要查找：                                                   │
│                                                                         │
│  • 负责人: 张三                                                        │
│  • 优先级: 高 (high)                                                   │
│  • 标签: bug                                                           │
│                                                                         │
│  生成的 RQL:                                                            │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │ assignee = "张三" AND priority = "high" AND label = "bug"        │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  [编辑 RQL]  [使用此查询]                                               │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6.4 全局搜索入口

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 🔍 [全局搜索... (按 / 聚焦)]                                    [ESC]  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─ 搜索范围 ────────────────────────────────────────────────────────┐ │
│  │ ○ 所有   ● 工作项   ○ 周期   ○ 模块   ○ 成员                      │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌─ RQL 模式 ────────────────────────────────────────────────────────┐ │
│  │ > _                                                           │   │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  热门查询:                                                              │
│  • 我负责的工作项: assignee = "current_user"                            │
│  • 高优先级 bug: priority = "high" AND label = "bug"                    │
│  • 未完成的周期: status != "completed"                                  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 7. 实现计划

### Phase 1: 核心引擎 (MVP)

- [ ] `utils/rql/types.ts` - 类型定义
- [ ] `utils/rql/lexer.ts` - 词法分析器
- [ ] `utils/rql/parser.ts` - 语法分析器
- [ ] `utils/rql/executor.ts` - 查询执行器
- [ ] `utils/rql/validator.ts` - 语法验证
- [ ] `composables/useRQL.ts` - 核心逻辑封装
- [ ] `components/RQL/RQLInput.vue` - RQL 输入框
- [ ] 集成到 IssueList 和 IssueKanban
- [ ] 集成到全局搜索 GlobalSearch

### Phase 2: 可视化构建器

- [ ] `components/RQL/RQLVisualBuilder.vue` - 可视化构建器
- [ ] 双向同步：表单 ↔ RQL
- [ ] 集成到工作项和全局搜索

### Phase 3: 查询管理

- [ ] `stores/rql.ts` - RQL 状态管理
- [ ] `components/RQL/RQLHistory.vue` - 查询历史
- [ ] `components/RQL/RQLSavedQueries.vue` - 保存的查询
- [ ] localStorage 持久化

### Phase 4: AI 解释器

- [ ] `utils/rql/interpreter.ts` - AI 规则引擎
- [ ] `components/RQL/RQLAIRecognizer.vue` - AI 面板
- [ ] 短语映射表完善

---

## 8. 风险与注意事项

### 8.1 潜在风险

| 风险 | 缓解措施 |
|------|----------|
| SQL 注入 | RQL 是受限查询语言，不支持子查询和函数 |
| 性能问题 | 后端添加查询超时限制 |
| 解析复杂性 | 从简单场景开始，逐步增加语法支持 |

### 8.2 语法限制

- 不支持 `SELECT` 子句（自动返回所有字段）
- 不支持 `JOIN` 操作
- 不支持聚合函数 (`COUNT`, `SUM` 等)
- 不支持子查询
- 字段值必须用引号包裹（防止注入）
- 最大嵌套深度 3 层

---

## 9. 测试计划

### 9.1 单元测试

- [ ] 词法分析器测试 (Tokenize)
- [ ] 语法分析器测试 (Parse)
- [ ] 执行器测试 (Execute)
- [ ] AI 解释器测试 (Interpret)

### 9.2 集成测试

- [ ] IssueList RQL 查询测试
- [ ] IssueKanban RQL 查询测试
- [ ] 全局搜索测试

---

## 10. 附录

### 10.1 参考资料

- [Pratt Parser 技术](https://tdop.dev/posts/pratt-parsing/)
- [GitHub Issues 搜索语法](https://docs.github.com/en/search-github/searching-on-github)
- [Linear 搜索语法](https://linear.app/docs/search)

### 10.2 优先级枚举

| 值 | 描述 |
|----|------|
| `urgent` | 紧急 |
| `high` | 高 |
| `medium` | 中 |
| `low` | 低 |
| `none` | 无 |

### 10.3 状态组

| 组 | 状态示例 |
|----|----------|
| `backlog` | 待处理 |
| `unstarted` | 未开始 |
| `started` | 进行中 |
| `completed` | 已完成 |
| `cancelled` | 已取消 |
