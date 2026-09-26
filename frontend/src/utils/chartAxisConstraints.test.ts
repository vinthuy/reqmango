import { describe, it, expect } from 'vitest'
import { coerceAxesForChartType, isTimeAxis } from './chartAxisConstraints'

describe('coerceAxesForChartType', () => {
  it('forces categorical X for pie when X is time', () => {
    expect(coerceAxesForChartType('pie', 'created_week', 'count')).toEqual({
      x_axis: 'state',
      y_axis: 'count',
    })
  })

  it('forces time X for line when X is categorical', () => {
    expect(coerceAxesForChartType('line', 'state', 'count')).toEqual({
      x_axis: 'created_week',
      y_axis: 'count',
    })
  })

  it('leaves compatible axes alone', () => {
    expect(coerceAxesForChartType('bar', 'priority', 'count')).toEqual({
      x_axis: 'priority',
      y_axis: 'count',
    })
    expect(coerceAxesForChartType('line', 'completed_week', 'throughput')).toEqual({
      x_axis: 'completed_week',
      y_axis: 'throughput',
    })
  })

  it('detects time axes', () => {
    expect(isTimeAxis('created_week')).toBe(true)
    expect(isTimeAxis('state')).toBe(false)
  })
})
