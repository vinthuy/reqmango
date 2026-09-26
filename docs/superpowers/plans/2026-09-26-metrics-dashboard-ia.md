# Metrics ↔ Dashboard IA Refactor

> **For agentic workers:** Implement task-by-task. Checkboxes track progress.

**Goal:** Split responsibilities — Metrics authors charts; Dashboards compose panels — and remove duplicate chart configuration UX.

**Architecture:** New dashboard widget type `metric_chart` references `metric_charts.id`. Dashboard add-widget flow prefers picking existing metrics charts. Inline bar/pie/line config remains for backward compat but is demoted. Project tab shows「度量」; dashboards stay a separate board surface.

**Tech Stack:** Vue 3 + Go/Gin + existing metrics/dashboard APIs

**Spec:** Chat design 2026-09-26 (P0–P2)

## Global Constraints

- Do not break existing dashboards with bar_chart/pie_chart/etc. widgets
- Keep `?tab=reports` as alias for metrics tab
- True burndown only via cycle burndown widget / cycle detail

---

### Task 1: Rename + IA copy
- [ ] Locales: project tab/nav use 度量/Metrics; dashboard empty states mention「来自度量的图」
- [ ] `?tab=reports` still opens metrics; prefer `tab=metrics`

### Task 2: Backend `metric_chart` widget
- [ ] `renderWidget` case `metric_chart` → MetricService.RenderChart
- [ ] Types/docs for config `{metric_chart_id}`

### Task 3: Frontend metric_chart widget
- [ ] WidgetType + WidgetCard Chart.js render (reuse canvas path)
- [ ] WidgetConfigPanel: pick metric chart / link to metrics

### Task 4: Metrics → Add to dashboard
- [ ] MetricsChartCard action; create widget on default/selected dashboard

### Task 5: Dashboard add-widget redesign
- [ ] Primary: list project metric charts
- [ ] Secondary: number_card, recent_list, burndown
- [ ] Legacy chart types under「高级」collapse

### Task 6: Axis constraints (P2)
- [ ] MetricsView: line/area require time x_axis; pie/doughnut require categorical

### Task 7: Verify
- [ ] Unit tests + API smoke
