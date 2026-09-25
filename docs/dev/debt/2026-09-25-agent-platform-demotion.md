# Demoted Agent-platform surface — cleanup backlog

**Status:** Stage A + Stage B done (2026-09-26). Stage C = code deletion (high risk, later).  
**Keep forever (PM path):** Copilot, ensure-pm, `dispatch_agent`, analyze, automation-preview, `ai_summary`, core Agent assign/@Agent.

## Done

### Stage A

- [x] TopBar：移除项目主导航「Agent 成员 / 工作流 / Agent 任务 / 预算与SLA」
- [x] Docs / KB 标明 Harness/Loop 非产品方向

### Stage B

- [x] `/workspace/:slug/agents*` 等默认重定向（`localStorage.rm_advanced_agents !== '1'`）
- [x] AI Settings 隐藏控制台深链 + 提示文案
- [x] 设计文档归档至 [superseded/agent-platform](../../superseded/agent-platform/README.md)

## Later (Stage C — high risk)

- [ ] 审计后删除 `backend/internal/ai/{harness,loop,registry}` 等与 skill 执行耦合部分
- [ ] 删除 `frontend/src/views/agents/**` 等控制台视图
- [ ] **禁止**误删 `pm_agents.go` / `automation_preview.go` / `dashboard_ai_summary.go`

## 产品真相

排期与 PRD 以 [kb/PRD.md](../../kb/PRD.md) 与 [AI PM redesign](../../superpowers/specs/2026-09-25-ai-project-management-redesign.md) 为准；Agent 平台旧 PRD **已废止**。
