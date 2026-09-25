# Release preflight — human-like QA (2026-09-25)

**Verdict:** PASS  
**Tester:** Cursor Browser (simulated human) + API smoke  
**Environment:** http://localhost:5173 · API :8000 · Project 4 (OPENAPI / 开放平台 & API)  
**Git:** `master` @ `a5128a4` pushed to `github` + `origin` (HTTP/1.1 + Clash 7897)

## Paths exercised

| Area | Result | Evidence |
|------|--------|----------|
| P1 sidebar | PASS | Workspace nav: 项目 / 战略目标 / 设置 only（无 Agents 一级） |
| Copilot Ctrl+J | PASS | Ask/Build；文案「Ask 为只读查询…」 |
| C1 项目 AI 总结 | PASS | 按钮触发加载中…（此前 API 已验存 Page） |
| B1 自动化模板 | PASS | 设置→自动化：新建即分诊 / 高优即风险评估 / 新建即 Spec 草稿 + 规则列表并存 |
| B2 NL 配规则 | PASS | 「用自然语言创建」+ 生成规则草稿；`test_ai_b2_smoke.ps1` PASS |
| C2 Dashboard AI | PASS | 编辑→添加「AI 项目摘要」→组件数 5；页面含洞察/瓶颈；`test_ai_c2_smoke.ps1` PASS |
| Issue 指派 Agent | PASS | Issue 详情侧栏可见「AI 智能体」列表（含请求分诊/交付风险/Spec 草稿）+「指派 Agent」 |
## API smoke

- `backend/test_ai_b2_smoke.ps1` → `B2_SMOKE_PASS`
- `backend/test_ai_c2_smoke.ps1` → `C2_SMOKE_PASS` (summary + insights)

## Product debt noted (not blockers)

1. TopBar 曾暴露 Agent 成员/工作流/Agent 任务/预算与SLA → **本轮已从主导航移除**（深链仍可达）。
2. Harness/Loop/高级 `/agents/*` 控制台代码仍在；**不删包**，仅降权（见 KB / inventory）。
3. 仪表盘空状态文案偶发与侧栏并存（「创建第一个」与已有仪表盘）— UX 毛边，非功能阻塞。

## Sign-off

Phase 1–2 缺陷 DEF-01…07 均已 fixed 且合入；本轮复测主路径通过 → **提升为通过**。
