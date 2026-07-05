# Custom Report Builder Redesign - X/Y Axis Design

## Overview

Replace the current `report_type` + `group_by` coupling with explicit X-axis (dimension) and Y-axis (metric) selectors. Implement a step-by-step flow: filter first, then configure chart.

## Current Problems

1. `report_type` and `group_by` implicitly control X/Y axes - users can't see what maps to what
2. No clear separation between data filtering and chart configuration
3. `created_vs_resolved` and `created_trend` silently ignore `group_by`, confusing users

## Design

### UX Flow

```
Step 1: Filter Data
  - Visual filter builder (field + operator + value)
  - RQL advanced editor
  - Date range
  - [Apply Filter] -> shows "N issues matched"

Step 2: Configure Chart
  - X-axis: dimension selector (category/time/custom field)
  - Y-axis: metric selector (count/avg time/created vs resolved)
  - Chart type: bar/pie/line/table/...
  - Live chart preview
  - Export / Save report
```

### X-Axis Options (Dimensions)

| Group | Options | SQL |
|-------|---------|-----|
| Category | State, Priority, Assignee, Type, Label, Cycle, Module | Existing `buildAggregation()` |
| Time - Created | Day/Week/Month | `TO_CHAR(created_at, format)` |
| Time - Completed | Day/Week/Month | `TO_CHAR(completed_at, format)` |
| Time - Updated | Day/Week/Month | `TO_CHAR(updated_at, format)` |
| Custom Fields | Dynamic enum fields from project | `COALESCE(cf.value, 'None')` with JOIN |

### Y-Axis Options (Metrics)

| Metric | Description | SQL |
|--------|-------------|-----|
| Issue Count | Count per group | `COUNT(*)` |
| Avg Processing Time | Avg days for resolved issues | `AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400)` |
| Current Retention | Avg days for unresolved issues | `AVG(EXTRACT(EPOCH FROM (NOW() - created_at)) / 86400)` |
| Created vs Resolved | Dual dataset (count created + count resolved) | Two separate `COUNT(*)` queries |

### Backend Changes

#### New Report Service Method

Replace the current 5-type routing with a unified query builder:

```go
type ReportRequest struct {
    XAxis    string `json:"x_axis"`     // dimension key
    YAxis    string `json:"y_axis"`     // metric key
    Interval string `json:"interval"`   // day/week/month (for time axes)
    RQL      string `json:"rql"`        // filter conditions
    DateFrom string `json:"date_from"`
    DateTo   string `json:"date_to"`
}

type ReportResponse struct {
    Type    string            `json:"type"`
    Labels  []string          `json:"labels"`       // X-axis labels
    Values  []int             `json:"values"`       // Y-axis values (primary)
    Values2 []int             `json:"values2,omitempty"` // Y-axis values (secondary, for created_vs_resolved)
    Total   int               `json:"total"`
    Colors  map[string]string `json:"colors,omitempty"`
}
```

#### X-Axis SQL Builder

```go
func buildXAxisSQL(xAxis string, interval string) (selectExpr string, joinClause string, groupBy string) {
    switch xAxis {
    case "state":
        return "COALESCE(s.name, 'Unknown')", "LEFT JOIN states s ON issues.state_id = s.id", "1"
    case "priority":
        return "COALESCE(issues.priority, 'none')", "", "1"
    case "assignee":
        return "COALESCE(u.display_name, 'Unassigned')", "LEFT JOIN issue_assignees ia ON ... LEFT JOIN users u ON ...", "1"
    case "created_day":
        return "TO_CHAR(issues.created_at, 'YYYY-MM-DD')", "", "1"
    case "created_week":
        return "TO_CHAR(issues.created_at, 'IYYY-IW')", "", "1"
    case "created_month":
        return "TO_CHAR(issues.created_at, 'YYYY-MM')", "", "1"
    case "completed_day":
        return "TO_CHAR(issues.completed_at, 'YYYY-MM-DD')", "", "1"
    // ... etc
    case "custom_field_{id}":
        return "COALESCE(cf.value, 'None')", "LEFT JOIN custom_field_values cf ON ...", "1"
    }
}
```

#### Y-Axis SQL Builder

```go
func buildYAxisSQL(yAxis string) (selectExpr string, havingClause string) {
    switch yAxis {
    case "count":
        return "COUNT(*)::int", ""
    case "avg_processing_time":
        return "AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400)::int", "HAVING AVG(...) IS NOT NULL"
    case "current_retention":
        return "AVG(EXTRACT(EPOCH FROM (NOW() - created_at)) / 86400)::int", "HAVING AVG(...) IS NOT NULL"
    case "created_vs_resolved":
        // Special: needs two queries
    }
}
```

### Frontend Changes

#### ReportBuilder.vue Refactor

- Remove `reportType` ref
- Add `xAxis` ref (default: 'state')
- Add `yAxis` ref (default: 'count')
- Add `interval` ref (default: 'week', only visible when time axis selected)
- Split into two visual sections:
  - Section 1: Filter builder (existing, but with explicit "Apply" button)
  - Section 2: X/Y axis selectors + chart type + preview
- Show match count after applying filter
- Chart only renders after both filter is applied AND X/Y are configured

#### useReportChart.ts Changes

- Remove `renderMixed`, `renderBubble`, `renderScatter` if not needed for new design (keep them for backwards compat)
- The render functions stay the same - they take `labels` + `values` and render charts

#### API Layer

- Add new `reportApi.generateV2()` method or update existing to accept `x_axis`/`y_axis`
- Keep old `reportApi.generate()` for backward compatibility with saved reports

### Custom Fields Support

- Frontend: on mount, fetch project custom fields, filter to enum type, show as X-axis options
- Backend: new endpoint or extend existing to return available custom fields for reports
- SQL: join `custom_field_values` table with appropriate field ID

## Files to Modify

| File | Changes |
|------|---------|
| `frontend/src/components/ReportBuilder.vue` | Major refactor: step-by-step UI, X/Y selectors |
| `frontend/src/composables/useReportChart.ts` | Minor: ensure compatibility with new data format |
| `frontend/src/api/report.ts` | Add V2 generate method |
| `backend/internal/service/report_service.go` | New unified query builder |
| `backend/internal/handler/report_handler.go` | Accept new request format |
| `backend/internal/dto/request/report.go` | New request DTO |
| `frontend/src/locales/en-US.json` | New i18n keys |
| `frontend/src/locales/zh-CN.json` | New i18n keys |

## Migration Strategy

- Keep old `reportApi.generate()` working for Quick Charts tab and saved reports
- New X/Y design only applies to Custom Reports tab
- Existing saved reports continue to work with old format
