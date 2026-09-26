/**
 * Coerce X/Y axes when chart type and axis family disagree.
 * Line/area → time axis; pie/doughnut → categorical axis.
 */

export const TIME_X_AXES = [
  'created_day', 'created_week', 'created_month',
  'completed_day', 'completed_week', 'completed_month',
  'updated_day', 'updated_week', 'updated_month',
] as const

export const CATEGORICAL_X_AXES = [
  'state', 'state_group', 'priority', 'assignee', 'type',
  'label', 'module', 'cycle', 'created_by', 'reporter', 'title',
] as const

export function isTimeAxis(x: string): boolean {
  return (TIME_X_AXES as readonly string[]).includes(x) || x.endsWith('_day') || x.endsWith('_week') || x.endsWith('_month')
}

export function isCategoricalAxis(x: string): boolean {
  return (CATEGORICAL_X_AXES as readonly string[]).includes(x) || x.startsWith('custom_field:')
}

export function coerceAxesForChartType(
  chartType: string,
  xAxis: string,
  yAxis: string,
): { x_axis: string; y_axis: string } {
  const type = (chartType || 'bar').toLowerCase()
  let x = xAxis || 'state'
  let y = yAxis || 'count'

  if (['pie', 'doughnut'].includes(type) && isTimeAxis(x)) {
    x = 'state'
    if (!y) y = 'count'
  }
  if (['line', 'area'].includes(type) && !isTimeAxis(x)) {
    x = 'created_week'
    if (!y) y = 'count'
  }
  return { x_axis: x, y_axis: y }
}
