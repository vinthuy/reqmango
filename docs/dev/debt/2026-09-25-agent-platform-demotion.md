# Demoted Agent-platform surface — cleanup backlog

**Status:** Stage A started (2026-09-25). No package deletion yet.  
**Keep forever (PM path):** Copilot, ensure-pm, `dispatch_agent`, analyze, automation-preview, `ai_summary`, core Agent assign/@Agent.

## Done (Stage A)

- [x] TopBar：移除项目主导航「Agent 成员 / 工作流 / Agent 任务 / 预算与SLA」（深链仍可用）
- [x] Docs / KB 标明 Harness/Loop 非产品方向

## Next (Stage B — hide, don't delete)

- [ ] Feature-flag or drop `/workspace/:slug/agents/*` 默认可达性（保留高级设置深链）
- [ ] 清理 locales 中 `ai.dashboard.*` 平台文案曝光

## Later (Stage C — high risk)

- [ ] 审计后删除 `backend/internal/ai/{harness,loop,registry}` 等与 skill 执行耦合部分
- [ ] 删除 `frontend/src/views/agents/**` 等控制台视图
- [ ] **禁止**误删 `pm_agents.go` / `automation_preview.go` / `dashboard_ai_summary.go`

详见探索结论（本会话 inventory）。
