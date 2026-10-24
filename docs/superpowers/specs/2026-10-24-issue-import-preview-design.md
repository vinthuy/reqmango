# 工作项导入：预校验、字段映射与加固（设计）

日期：2026-10-24
状态：待实现
上游需求：[Plane 差距 Task 3 Bulk Import CSV/JSON](../../dev/plane-gap-implementation-plan.md)

---

## 1. 背景

`POST /issues/import/csv`、`POST /issues/import/json`、`GET /issues/export` 与
`ImportIssuesModal.vue` 已经存在，"批量导入"并非从零开发。本轮只补齐缺口：

| 缺口 | 现状 |
|------|------|
| 无预校验 | 只能真跑一次才知道哪几行会失败，失败行会与成功行混在一起落库 |
| 无字段映射 | 表头必须命中内置别名（`name`/`标题`/`summary`/`title`），自定义列名无法导入，用户只能拿到"标题为空"的报错 |
| 静默降级 | `state` / `type` / `labels` / `assignees` 匹配不到时被悄悄忽略：状态回落到项目默认状态，类型变为空，标签与负责人直接丢弃 |
| 无体量上限 | `csv.ReadAll` 一次性读全文件，行数、文件大小都没有限制 |
| 项目与工作空间不校验 | `?project_id=` 与 `?workspace_id=` 可指向不同工作空间，写入的 issue 会带上错误的 `workspace_id` |

跨租户读取/写入本身已被 `ScopedAuthMiddleware` 的 `project_id` 作用域检查挡住
（外部账号 `GET /issues/export?project_id=<他人项目>` 实测 403），本轮不重复处理。

## 2. 目标

1. 导入前可预览：解析出的表头、自动识别的字段映射、每行的校验结果、无法识别的状态/类型/标签/负责人。
2. 字段映射可手工调整：任意列名都能映射到标准字段，预览与最终导入使用同一份映射。
3. 未知取值不再静默：预览以 warning 呈现，最终导入把 warning 计入响应。
4. 有界：单次导入行数上限 5000，CSV 文件上限 5 MB。
5. `project_id` 必须属于 `workspace_id`，否则 400。

## 3. 设计

### 3.1 复用既有路由 + `dry_run`

不新增路由，`/issues/import/csv` 与 `/issues/import/json` 接受两个新参数：

| 参数 | 位置 | 说明 |
|------|------|------|
| `dry_run=true` | query | 只解析与校验，不落库 |
| `mapping` | CSV 的 multipart 表单字段（JSON 字符串） | `{"name":"标题列","state":"状态列",...}`，键为规范字段名，值为 CSV 表头原文 |

`mapping` 缺省时按内置别名自动识别。自动识别与手工映射共用一套代码路径，
避免"预览通过、导入失败"。

### 3.2 响应

`ImportResult` 增加三个字段（向后兼容，旧客户端忽略即可）：

```go
type ImportResult struct {
    SuccessCount int           `json:"success_count"`
    FailCount    int           `json:"fail_count"`
    Errors       []ImportError `json:"errors"`
    ImportedIDs  []uint64      `json:"imported_ids"`

    DryRun       bool              `json:"dry_run,omitempty"`
    TotalRows    int               `json:"total_rows"`
    Headers      []string          `json:"headers,omitempty"`  // 仅 CSV
    Mapping      map[string]string `json:"mapping,omitempty"`  // 规范字段 -> 源表头
    Warnings     []ImportWarning   `json:"warnings,omitempty"`
}

type ImportWarning struct {
    Row     int    `json:"row"`
    Field   string `json:"field"`   // state / type / label / assignee / parent
    Message string `json:"message"`
}
```

`dry_run` 响应中 `SuccessCount` 表示"通过校验的行数"，`ImportedIDs` 为空。

### 3.3 校验顺序

1. `len(items) > 5000` → 400 `导入行数超过上限（5000 行）`。
2. `project.WorkspaceID != workspaceID` → 400 `项目不属于所选工作空间`。
3. 逐行：
   - `name` 为空 → error。
   - `state` 非空但未命中 → warning，回落默认状态（保留现有行为，但现在可见）。
   - `type` 非空但未命中或不可用于该项目 → warning，类型留空。
   - `labels` / `assignees` 未命中 → warning。
   - `parent_title` 未在本次导入或库内命中 → warning，按顶层创建。
4. 未标 error 的行才创建；创建失败按 error 计入。

### 3.4 前端

`ImportIssuesModal.vue` 的 CSV 分支改为两步：

1. 选择文件后自动调用 `dry_run`：渲染表头、映射下拉（每个标准字段一个）、前 5 行数据预览、错误与警告列表、总行数。
2. 点击"开始导入"用同一份映射正式导入；预览有 error 时按钮不阻止（部分成功是既有语义），但提示会失败的行数。

标准字段固定为：`name`（必填）、`description`、`priority`、`state`、`type`、`assignees`、`labels`、`start_date`、`target_date`、`parent_title`、`module`、`cycle`、`estimate`。

JSON 分支保持现状（结构即字段，无需映射），但同样先走 `dry_run` 展示校验结果。

### 3.5 非目标

- 不做并发导入队列、断点续传。
- 不做导入模板持久化 / 保存常用映射。
- 不改变 `module`/`cycle` 目前"解析但不写入"的既有行为（另开任务）。

## 4. 测试

| 层 | 用例 |
|----|------|
| Go service | 手工映射命中自定义表头；`dry_run` 不落库；未知状态产生 warning 且回落默认；行数超限 400；项目与工作空间不匹配 400；CSV 超过 5 MB 400 |
| Go handler | `dry_run` 返回 `headers` 与 `mapping`；非法 `mapping` JSON 400 |
| Vitest | 选择文件后调用预览接口；渲染映射下拉；改映射后正式导入携带映射；预览错误行计数展示 |
| Playwright | 导入 CSV 走预览 → 映射 → 导入 → 列表出现新工作项（复用既有 `import` spec） |
