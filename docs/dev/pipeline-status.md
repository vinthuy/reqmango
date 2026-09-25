# Pipeline Status（功能管线状态）

最后更新：2026-09-25

---

## 活跃功能

| 功能 | Spec | Design | Plan | Implement | 备注 |
|------|------|--------|------|-----------|------|
| AI 项目管理重设计（对标 Plane AI） | ✅ | ✅ | ✅ | 🔄 Phase1 Tasks 1–8 代码已接线 | 剩手测 + Task9 文档勾选；可选本地 commit |

---

## 已走完管线并归档

| 功能 | 后端 | Spec | Design | Plan | Implement | Review | KB | 备注 |
|------|------|------|--------|------|-----------|--------|----|------|
| Cycle（周期） | Go | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | CRUD + 状态流转 + 进度 + 燃尽图 |
| Module（模块） | Go | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | CRUD + 树形 + Issue 关联 + 统计 |

---

## 已在代码中落地（2026-06 Backlog，文档曾过期）

CustomField / Workflow / Automation / IssueType / Comments / Notifications / Attachments / Estimate 等均已有 handler；详见历史提交与 `docs/bug-list.md`（52/52 已修复）。

---

## 废止 / 降级

| 文档 | 状态 |
|------|------|
| `docs/AI_AGENT_PRD.md`（Multica / 全 SDLC） | Superseded |
| `docs/superpowers/specs/2026-07-18-reqmango-agent-platform-design.md`（Harness/Loop 主叙事） | Demoted → Phase 3+ |
| MCP/CLI 分发作为增长主路径 | 非当前 AI 产品排期 |

---

## 图例

| 标记 | 含义 |
|------|------|
| ✅ | 完成 |
| 🔄 | 进行中 |
| ⏳ | 待开始 |
| ❌ | 取消 |
| - | 不适用 |
