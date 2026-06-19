import { defineStore } from 'pinia'
import type { Issue } from '@/types/issue'

export interface IssueFilters {
  search: string
  states: number[]
  priorities: string[]
  assignees: number[]
  labels: number[]
  cycleId: number | null
}

export const useIssueStore = defineStore('issue', {
  state: () => ({
    // 视图模式
    currentView: 'list' as 'list' | 'kanban',

    // 筛选条件
    filters: {
      search: '',
      states: [] as number[],
      priorities: [] as string[],
      assignees: [] as number[],
      labels: [] as number[],
      cycleId: null as number | null,
    },

    // 列表数据
    issues: [] as Issue[],
    totalIssues: 0,
    currentPage: 1,
    pageSize: 20,

    // 批量选择
    selectedIds: [] as number[],

    // 详情抽屉
    detailDrawerOpen: false,
    currentIssueId: null as number | null,

    // 加载状态
    loading: false,
    error: null as string | null,
  }),

  getters: {
    // 是否全选
    isAllSelected: (state) => {
      return state.issues.length > 0 &&
             state.selectedIds.length === state.issues.length
    },

    // 是否部分选择
    isPartialSelected: (state) => {
      return state.selectedIds.length > 0 &&
             state.selectedIds.length < state.issues.length
    },

    // 选中数量
    selectedCount: (state) => state.selectedIds.length,
  },

  actions: {
    // 切换视图
    setView(view: 'list' | 'kanban') {
      this.currentView = view
      localStorage.setItem('issue-view', view)
    },

    // 初始化视图（从 localStorage）
    initView() {
      const saved = localStorage.getItem('issue-view')
      if (saved === 'list' || saved === 'kanban') {
        this.currentView = saved
      }
    },

    // 设置筛选条件
    setFilters(filters: Partial<IssueFilters>) {
      this.filters = { ...this.filters, ...filters }
    },

    // 清除筛选条件
    clearFilters() {
      this.filters = {
        search: '',
        states: [],
        priorities: [],
        assignees: [],
        labels: [],
        cycleId: null,
      }
    },

    // 选择工作项
    selectIssue(issueId: number) {
      if (!this.selectedIds.includes(issueId)) {
        this.selectedIds.push(issueId)
      }
    },

    // 取消选择工作项
    unselectIssue(issueId: number) {
      this.selectedIds = this.selectedIds.filter(id => id !== issueId)
    },

    // 全选
    selectAll() {
      this.selectedIds = this.issues.map(issue => issue.id)
    },

    // 取消全选
    unselectAll() {
      this.selectedIds = []
    },

    // 打开详情抽屉
    openDetailDrawer(issueId: number) {
      this.currentIssueId = issueId
      this.detailDrawerOpen = true
    },

    // 关闭详情抽屉
    closeDetailDrawer() {
      this.detailDrawerOpen = false
      this.currentIssueId = null
    },

    // 设置分页
    setPage(page: number) {
      this.currentPage = page
    },

    // 设置每页数量
    setPageSize(size: number) {
      this.pageSize = size
      this.currentPage = 1
    },
  },
})