/**
 * AIChartRenderer 组件测试
 * 覆盖：组件挂载/卸载、chart 配置传递
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'

const { mockDestroy, mockChartCtor } = vi.hoisted(() => {
  const destroy = vi.fn()
  const ctor = vi.fn()
  return { mockDestroy: destroy, mockChartCtor: ctor }
})

vi.mock('chart.js', () => {
  class MockChart {
    destroy = mockDestroy
    data = {}
    options = {}
    update = vi.fn()
    resize = vi.fn()
    constructor(..._args: any[]) {
      mockChartCtor(..._args)
      // return this implicitly
    }
    static register = vi.fn()
    static defaults = {}
    static overrides = {}
  }
  class Base {}
  return {
    Chart: MockChart,
    CategoryScale: class extends Base {},
    LinearScale: class extends Base {},
    BarElement: class extends Base {},
    PointElement: class extends Base {},
    LineElement: class extends Base {},
    ArcElement: class extends Base {},
    RadialLinearScale: class extends Base {},
    Filler: class extends Base {},
    Title: class extends Base {},
    Tooltip: class extends Base {},
    Legend: class extends Base {},
    BarController: class extends Base {},
    LineController: class extends Base {},
    PieController: class extends Base {},
    DoughnutController: class extends Base {},
    PolarAreaController: class extends Base {},
    RadarController: class extends Base {},
    BubbleController: class extends Base {},
    ScatterController: class extends Base {},
    register: vi.fn(),
  }
})

import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import AIChartRenderer from './AIChartRenderer.vue'

describe('AIChartRenderer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // jsdom does not support canvas.getContext
    HTMLCanvasElement.prototype.getContext = vi.fn().mockReturnValue({
      canvas: {} as any,
    } as any)
  })

  describe('rendering', () => {
    it('should render a canvas element', () => {
      const wrapper = mount(AIChartRenderer, {
        props: {
          config: {
            chart_type: 'bar',
            labels: ['A', 'B'],
            datasets: [{ label: 'Data', data: [1, 2] }],
          },
        },
      })
      expect(wrapper.find('canvas').exists()).toBe(true)
    })
  })

  describe('props validation', () => {
    it('should accept bar chart config', () => {
      const wrapper = mount(AIChartRenderer, {
        props: {
          config: {
            chart_type: 'bar',
            labels: ['Red', 'Blue', 'Green'],
            datasets: [
              { label: 'Votes', data: [12, 19, 3] },
              { label: 'Comments', data: [5, 2, 8] },
            ],
            title: 'Sample Chart',
          },
        },
      })
      expect(wrapper.props('config').chart_type).toBe('bar')
      expect(wrapper.props('config').labels).toHaveLength(3)
      expect(wrapper.props('config').datasets).toHaveLength(2)
    })

    it('should accept pie chart config', () => {
      const wrapper = mount(AIChartRenderer, {
        props: {
          config: {
            chart_type: 'pie',
            labels: ['Bug', 'Feature', 'Task'],
            datasets: [{ label: 'Types', data: [30, 50, 20] }],
          },
        },
      })
      expect(wrapper.props('config').chart_type).toBe('pie')
    })

    it('should accept doughnut chart config', () => {
      const wrapper = mount(AIChartRenderer, {
        props: {
          config: {
            chart_type: 'doughnut',
            labels: ['Open', 'Closed'],
            datasets: [{ label: 'Status', data: [25, 75] }],
          },
        },
      })
      expect(wrapper.props('config').chart_type).toBe('doughnut')
    })

    it('should accept line chart config', () => {
      const wrapper = mount(AIChartRenderer, {
        props: {
          config: {
            chart_type: 'line',
            labels: ['W1', 'W2', 'W3', 'W4'],
            datasets: [{ label: 'Velocity', data: [8, 12, 10, 15] }],
          },
        },
      })
      expect(wrapper.props('config').chart_type).toBe('line')
    })

    it('should accept radar chart config', () => {
      const wrapper = mount(AIChartRenderer, {
        props: {
          config: {
            chart_type: 'radar',
            labels: ['Speed', 'Quality', 'Cost'],
            datasets: [{ label: 'Sprint A', data: [80, 70, 60] }],
          },
        },
      })
      expect(wrapper.props('config').chart_type).toBe('radar')
    })
  })

  describe('lifecycle', () => {
    it('should destroy chart on unmount', async () => {
      const wrapper = mount(AIChartRenderer, {
        props: {
          config: {
            chart_type: 'bar',
            labels: ['A'],
            datasets: [{ label: 'X', data: [1] }],
          },
        },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 100))

      wrapper.unmount()
      expect(mockDestroy).toHaveBeenCalled()
    })
  })

  describe('watch reactivity', () => {
    it('should recreate chart when config changes', async () => {
      const wrapper = mount(AIChartRenderer, {
        props: {
          config: {
            chart_type: 'bar',
            labels: ['A'],
            datasets: [{ label: 'X', data: [1] }],
          },
        },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 100))

      const beforeCount = mockChartCtor.mock.calls.length
      await wrapper.setProps({
        config: {
          chart_type: 'line',
          labels: ['B', 'C'],
          datasets: [{ label: 'Y', data: [3, 4] }],
        },
      })
      await nextTick()
      await new Promise(r => setTimeout(r, 150))

      expect(mockChartCtor.mock.calls.length).toBeGreaterThan(beforeCount)
    })
  })
})
