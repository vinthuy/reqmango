# 实施计划：工作项导入预校验与字段映射

日期：2026-10-24
设计：[2026-10-24-issue-import-preview-design.md](../specs/2026-10-24-issue-import-preview-design.md)

## Task 1：后端解析与校验重构

- [x] 1.1 `dto/response/issue.go`：`ImportResult` 增加 `dry_run` / `total_rows` / `headers` / `mapping` / `warnings`，新增 `ImportWarning`
- [x] 1.2 `dto/request/issue.go`：新增 `ImportOptions{ DryRun bool; Mapping map[string]string }`
- [x] 1.3 `service/issue_service.go`：
  - [x] 抽出 `parseCSVRecords(records [][]string, mapping map[string]string) (headers []string, items []request.ImportIssueItem, err error)`
  - [x] `ImportFromJSON` / `ImportFromCSV` 增加 `ImportOptions`，校验与创建共用 `importItems` 单遍实现
  - [x] 校验顺序按设计 §3.3；`dry_run` 只校验不创建
  - [x] 行数上限常量 `maxImportRows = 5000`
- [x] 1.4 `handler/issue_handler.go`：
  - [x] `dry_run` query、CSV `mapping` 表单字段解析
  - [x] `http.MaxBytesReader` 5 MB
  - [x] 非法 `mapping` JSON → 400

## Task 2：后端测试

- [x] 2.1 `service/issue_import_test.go`：自定义表头映射、显式映射关闭别名回退、dry_run 不落库、未知状态 warning、负责人非项目成员 warning、行数超限、工作空间不匹配
- [x] 2.2 `handler/issue_import_test.go`：dry_run 返回 headers/mapping、非法 mapping、跨工作空间拒绝

## Task 3：前端两步导入

- [x] 3.1 `api/issue.ts`：`previewImportCSV/JSON`（带 `dry_run=true` 与 `mapping`）
- [x] 3.2 `ImportIssuesModal.vue`：文件选择后自动预览；映射下拉（自动识别失败时用文件表头兜底）；错误/警告列表；正式导入复用映射
- [x] 3.3 i18n `import.*` 新增键（中英）

## Task 4：前端测试与文档

- [x] 4.1 Vitest：预览调用、映射下拉渲染、导入携带 mapping、自动识别失败后手工映射可恢复
- [x] 4.2 Playwright：`e2e/issue-import-preview.spec.ts` CSV 预览 → 映射 → 导入闭环（chromium/firefox/webkit 通过）
- [x] 4.3 `plane-gap-implementation-plan.md` Task 3 标记完成；`pipeline-status.md` 增行
