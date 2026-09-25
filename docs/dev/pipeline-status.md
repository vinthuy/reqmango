# Pipeline Status（功能管线状态）

最后更新：2026-09-25

当前没有走 `docs/dev/features/` 流程的活跃功能。2026-06-21 之后的能力大多直接进了代码，没有对应的 Spec / Design / Plan / Review 目录。不要再按下面「已过期」那份 Backlog 开工。

---

## 已走完管线并归档

| 功能 | 后端 | Spec | Design | Plan | Implement | Review | KB | 备注 |
|------|------|------|--------|------|-----------|--------|----|------|
| Cycle（周期） | Go | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | CRUD + 状态流转 + 进度 + 燃尽图 |
| Module（模块） | Go | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | CRUD + 树形 + Issue 关联 + 统计 |

设计文档仍在 `docs/dev/features/2026-06-21-cycle/` 与 `docs/dev/features/2026-06-21-module/`。

---

## 已在代码中落地（2026-06-21 清单已过期）

这些条目当时标成「待启动」，现在都有对应 handler，产品侧视为已实现：

| 当时的 Backlog | 代码入口 |
|----------------|----------|
| CustomField | `custom_field_handler.go`、`conditional_field_handler.go` |
| Workflow + Automation | `workflow_handler.go`、`automation_handler.go`、`approval_handler.go` |
| IssueType | `issue_type_handler.go`、`project_issue_type_handler.go` |
| AI Assistant | `chat_handler.go` 及 `agent_*` / `ai` 相关 handler |
| EstimatePoint | `estimate_handler.go` |
| Comments | `comment_handler.go` |
| Notifications | `notification_handler.go`、`sse_handler.go` |
| Attachments | `attachment_handler.go` |

已知缺陷清单 `docs/bug-list.md` 在 2026-09-25 标为 52/52 已修复。

---

## 接下来

发布核对还没收口，优先于新功能：

1. **GitHub Actions 仍是红的。** `github/master` 停在 `687fade`，落后 gitcode 上的 `origin/master`（`34ebee3`）。最近一次运行（2026-09-24）里 Lint、Security、Test 通过，E2E 失败：chromium **12 failed / 381 passed**。失败集中在 `ai-phase1-e2e`、`automation-full-validation`、`chat-e2e`、`issue-detail-acceptance`、`pages-e2e`、`report-e2e`、`workflow-automation-ui`。Build 因 E2E 失败被跳过。gitcode 上更晚的修复还没在 GitHub 上跑过。
2. **`tests/` 套件未进 CI。** CI 只跑 `frontend/e2e` 的 chromium。`tests/` 仍有大量 `waitForTimeout`，并依赖固定账号与项目 id。
3. **一键脚本已跑过一轮**（2026-09-25，`powershell -NoProfile -File scripts/run-full-e2e.ps1`）。本机没有 `pwsh`；脚本须保持 UTF-8 BOM。冷启动时 `go run` 下载依赖超过原来的 180 秒，监听等待已改为 600 秒。本轮结果：Go 测试通过；vitest **833/833** 通过；`tests/` **63 failed / 298 passed**（1.1h，失败集中在项目页、Issue 详情/看板、评论，典型是找不到「查看」按钮）；`frontend/e2e` **231 failed / 141 passed / 26 did not run**（2.1m）。后者大量是 `browserType.launch: Executable doesn't exist`（Playwright 自带 Chromium 不在当前缓存路径），不是产品断言失败。通过的多半是不启动浏览器的 API 用例。

---

## 图例

| 标记 | 含义 |
|------|------|
| ✅ | 完成 |
| 🔄 | 进行中 |
| ⏳ | 待开始 |
| ❌ | 取消 |
| - | 不适用 |
