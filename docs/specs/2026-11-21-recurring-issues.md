# 重复工作项 — 需求规格说明书（SRS）

| 项 | 值 |
|----|-----|
| 状态 | 已冻结 |
| 日期 | 2026-11-21 |
| 关联 | ROADMAP Later · [sdd.md](../dev/sdd.md) |
| 依赖 | 既有 `recurrence_rules` API + `RecurrenceConfig.vue`（未挂载） |

---

## 1. 背景与目标

规则 CRUD API 与配置组件已存在，但未挂到工作项详情，且无后台按 `next_run` 生成实例。P0 目标：**在详情页配置 daily/weekly/monthly 规则**；**可选**同 PR 加最简调度（或明确标 P1）。

## 2. 范围 / 非目标

**范围（P0）**

- 在 Issue 详情 / 面板挂载 `RecurrenceConfig`（折叠或属性区）。
- UI 仅暴露 daily / weekly / monthly（隐藏 cron）。
- 创建 / 读取 / 暂停 / 删除；toast i18n。
- 写路由加 `issue:edit`。

**P0+（同里程碑若时间够）**

- 定时任务：扫描 `is_active && next_run <= now`，克隆模板 Issue，推进 `next_run`，尊重 `end_date`。

**非目标**

- cron 表达式；复杂字段继承策略；跨项目；审批流挂钩。

## 3. 验收场景

### AC-01 挂载与配置
Given 有 `issue:edit` 的用户打开工作项  
When 配置每周重复并保存  
Then GET recurrence 可读回；刷新后 UI 仍显示

### AC-02 暂停与删除
Given 已有活跃规则  
When 暂停或删除  
Then `is_active=false` 或 404

### AC-03 无权限
Given 无 `issue:edit`  
When POST/PUT/DELETE recurrence  
Then 403

### AC-04（若做执行器）生成实例
Given 活跃规则且 `next_run` 已过  
When 调度跑一轮  
Then 新建 Issue 且 `next_run` 前进

## 4. 验收清单

| AC-ID | 优先级 | 结果 |
|-------|--------|------|
| AC-01 | P0 | |
| AC-02 | P0 | |
| AC-03 | P0 | |
| AC-04 | P1 | |
