<template>
  <div class="issue-list bg-white rounded-lg border border-gray-200">
    <!-- Row 1: 主工具栏 -->
    <div class="px-4 py-2.5 border-b border-gray-100">
      <div class="flex items-center gap-3">
        <!-- 搜索框 -->
        <div class="relative flex-1 max-w-sm">
          <svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="filters.search"
            type="text"
            :placeholder="t('issueList.searchPlaceholder')"

            class="w-full pl-9 pr-3 py-1.5 border border-gray-200 rounded-md text-sm bg-gray-50 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-colors"
            @keydown.enter="search"
          />
        </div>

        <!-- 快速筛选芯片 -->
        <QuickFilterChips @filter="handleQuickFilter" />

        <div class="flex-1" />

        <!-- 右侧操作按钮 -->
        <div class="flex items-center gap-1.5">
          <!-- RQL 按钮 -->
          <button
            @click="showRQL = !showRQL"
            class="px-2.5 py-1.5 text-xs text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-md transition-colors"
            :class="{ 'bg-gray-100 text-gray-700': showRQL }"
            :title="t('issueList.rqlAdvanced')"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
          </button>

          <!-- 筛选按钮 -->
          <div class="relative">
            <button
              @click="showFilterDropdown = !showFilterDropdown"
              class="flex items-center gap-1.5 px-2.5 py-1.5 text-xs text-gray-600 border border-gray-200 rounded-md hover:bg-gray-50 transition-colors"
              :class="{ 'bg-gray-100 border-gray-300': activeFilterCount > 0 }"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
              </svg>
              <span>{{ t('issueList.filter') }}</span>
              <span v-if="activeFilterCount > 0" class="w-4 h-4 rounded-full bg-indigo-600 text-white text-[10px] flex items-center justify-center">{{ activeFilterCount }}</span>
            </button>

            <!-- 筛选下拉菜单 -->
            <div v-if="showFilterDropdown" class="absolute right-0 top-full mt-1 w-56 bg-white border border-gray-200 rounded-lg shadow-lg z-30 py-1" @click.stop>
              <div class="px-3 py-1.5 text-[11px] text-gray-400 font-medium uppercase tracking-wider">{{ t('issueList.addFilterCondition') }}</div>
              <button @click="addFilter('state_id')" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 flex items-center justify-between">
                {{ t('issue.state') }} <span class="text-gray-400 text-xs">{{ getFilterValueLabel('state_id') }}</span>
              </button>
              <button @click="addFilter('priority')" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 flex items-center justify-between">
                {{ t('issue.priority') }} <span class="text-gray-400 text-xs">{{ getFilterValueLabel('priority') }}</span>
              </button>
              <button @click="addFilter('assignee_id')" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 flex items-center justify-between">
                {{ t('issue.assignee') }} <span class="text-gray-400 text-xs">{{ getFilterValueLabel('assignee_id') }}</span>
              </button>
              <button @click="addFilter('cycle_id')" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 flex items-center justify-between">
                {{ t('issue.cycle') }} <span class="text-gray-400 text-xs">{{ getFilterValueLabel('cycle_id') }}</span>
              </button>
              <button @click="addFilter('start_date')" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 flex items-center justify-between">
                {{ t('issue.startDate') }} <span class="text-gray-400 text-xs">{{ getFilterValueLabel('start_date') }}</span>
              </button>
              <button @click="addFilter('target_date')" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 flex items-center justify-between">
                {{ t('issue.targetDate') }} <span class="text-gray-400 text-xs">{{ getFilterValueLabel('target_date') }}</span>
              </button>
              <div v-if="customFields.length > 0" class="border-t border-gray-100 mt-1 pt-1">
                <div class="px-3 py-1.5 text-[11px] text-gray-400 font-medium uppercase tracking-wider">{{ t('issue.customFields') }}</div>
                <button v-for="cf in customFields" :key="cf.id" @click="addCFFilter(cf.id)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50">
                  {{ cf.name }}
                </button>
              </div>
            </div>
          </div>

          <!-- 列配置 -->
          <div class="relative">
            <button
              @click="showColumns = !showColumns"
              class="flex items-center gap-1.5 px-2.5 py-1.5 text-xs text-gray-600 border border-gray-200 rounded-md hover:bg-gray-50 transition-colors"
              :class="{ 'bg-gray-100 border-gray-300': showColumns }"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
              </svg>
              <span>{{ t('issueList.columnConfig') }}</span>
            </button>
            <div v-if="showColumns" class="absolute right-0 top-full mt-1 w-44 bg-white border border-gray-200 rounded-lg shadow-lg z-20 py-1">
              <div class="px-3 py-1.5 text-[11px] text-gray-400 font-medium uppercase tracking-wider border-b border-gray-100">{{ t('issueList.displayColumns') }}</div>
              <label v-for="col in effectiveColumns" :key="col.key" class="flex items-center px-3 py-1.5 hover:bg-gray-50 cursor-pointer text-sm">
                <input type="checkbox" :checked="visibleColumnKeys.has(col.key)" @change="toggleColumn(col.key)" class="rounded border-gray-300 mr-2" />
                {{ t(col.labelKey || "") || col.label }}
              </label>
            </div>
          </div>

          <!-- 导入 -->
          <button @click="showImportModal = true" class="flex items-center gap-1.5 px-2.5 py-1.5 text-xs text-gray-600 border border-gray-200 rounded-md hover:bg-gray-50 transition-colors">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
            </svg>
            <span>{{ t('common.import') }}</span>
          </button>

          <!-- 新建 -->
          <button @click="goToCreate" class="flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 text-white text-xs rounded-md hover:bg-indigo-700 transition-colors font-medium">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>{{ t('project.create') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- RQL 查询栏（可收起） -->
    <div v-if="showRQL" class="px-4 py-2 border-b border-gray-100 bg-gray-50 flex items-center gap-2">
      <div class="flex-1">
        <RQLInput
          v-model="rqlQuery"
          :placeholder="t('issueList.rqlPlaceholder')"
          :error="rqlError"
          show-history
          @search="onRQLSearch"
        />
      </div>
      <svg v-if="rqlLoading" class="animate-spin h-4 w-4 text-indigo-600 shrink-0" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
      </svg>
    </div>

    <!-- Row 2: 激活的筛选条件芯片 -->
    <div v-if="activeFilterChips.length > 0" class="px-4 py-2 border-b border-gray-100 bg-gray-50/50 flex items-center gap-2 flex-wrap">
      <span class="text-[11px] text-gray-400 font-medium shrink-0">{{ t('issueList.filterConditions') }}:</span>
      <span
        v-for="chip in activeFilterChips"
        :key="chip.key"
        class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-white border border-gray-200 text-gray-700 shadow-sm"
      >
        <span class="text-gray-400">{{ chip.label }}:</span>
        <span class="font-medium">{{ chip.value }}</span>
        <button @click="removeFilter(chip.key)" class="ml-0.5 text-gray-400 hover:text-red-500 transition-colors">
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </span>
      <button @click="clearAllFilters" class="text-[11px] text-gray-400 hover:text-indigo-600 transition-colors ml-1">{{ t('issueList.clearAll') }}</button>
    </div>

    <!-- Row 3: 快速创建 -->
    <QuickCreateInput
      :project-id="projectId"
      :workspace-id="workspaceId"
      :issue-types="issueTypes"
      @created="onQuickCreated"
    />

    <!-- 筛选值选择弹窗（当点击添加筛选时出现） -->
    <div v-if="activeFilterPicker" class="px-4 py-3 border-b border-gray-100 bg-gray-50 flex items-center gap-3">
      <span class="text-xs text-gray-500 shrink-0">{{ activeFilterPicker.label }}:</span>
      <!-- 状态选择 -->
      <select v-if="activeFilterPicker.key === 'state_id'" v-model="filters.state_id" @change="applyFilterPicker" class="px-2 py-1 border border-gray-300 rounded text-sm">
        <option :value="0">{{ t('issueList.allStates') }}</option>
        <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
      </select>
      <!-- 优先级选择 -->
      <select v-else-if="activeFilterPicker.key === 'priority'" v-model="filters.priority" @change="applyFilterPicker" class="px-2 py-1 border border-gray-300 rounded text-sm">
        <option value="">{{ t('issueList.allPriorities') }}</option>
        <option value="urgent">{{ t('issue.priorityUrgent') }}</option><option value="high">{{ t('issue.priorityHigh') }}</option><option value="medium">{{ t('issue.priorityMedium') }}</option><option value="low">{{ t('issue.priorityLow') }}</option><option value="none">{{ t('issue.priorityNone') }}</option>
      </select>
      <!-- 周期选择 -->
      <select v-else-if="activeFilterPicker.key === 'cycle_id'" v-model="filters.cycle_id" @change="applyFilterPicker" class="px-2 py-1 border border-gray-300 rounded text-sm">
        <option :value="0">{{ t('common.all') }}</option>
        <option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
      <!-- 负责人选择 -->
      <UserSelect
        v-else-if="activeFilterPicker.key === 'assignee_id'"
        v-model="filtersAssignee"
        :users="memberOptions"
        :placeholder="t('issueList.selectAssignee')"
        :clearable="true"
        @update:model-value="applyFilterPicker"
      />
      <!-- 日期选择 -->
      <input v-else-if="activeFilterPicker.key === 'start_date'" type="date" v-model="filters.start_date" @change="applyFilterPicker" class="px-2 py-1 border border-gray-300 rounded text-sm" />
      <input v-else-if="activeFilterPicker.key === 'target_date'" type="date" v-model="filters.end_date" @change="applyFilterPicker" class="px-2 py-1 border border-gray-300 rounded text-sm" />
      <!-- 自定义字段 -->
      <input v-else-if="activeFilterPicker.key === 'cf'" type="text" :placeholder="t('issueList.inputFilterValue').replace('{0}', activeFilterPicker.label)" v-model="activeCFValue" @keydown.enter="applyFilterPicker" class="px-2 py-1 border border-gray-300 rounded text-sm flex-1 max-w-xs" />
      <button @click="cancelFilterPicker" class="text-xs text-gray-400 hover:text-gray-600">{{ t('common.cancel') }}</button>
    </div>

    <!-- 批量操作工具栏 -->
    <div v-if="selectedIds.size > 0" class="sticky top-0 z-20 bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-200 dark:border-indigo-800 rounded-lg px-4 py-2 flex items-center gap-3 mb-3 flex-wrap">
      <span class="text-sm text-indigo-700 dark:text-indigo-300 font-medium">{{ t('issueList.selected') }} {{ selectedIds.size }} {{ t('common.items') }}</span>

      <!-- 更改状态 -->
      <div class="relative">
        <button @click="showBatchState = !showBatchState" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          {{ t('issueList.changeState') }}
        </button>
        <div v-if="showBatchState" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 py-1 w-32">
          <button v-for="s in states" :key="s.id" @click="batchChangeState(s.id)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-200">{{ s.name }}</button>
        </div>
      </div>

      <!-- 更改优先级 -->
      <div class="relative">
        <button @click="showBatchPriority = !showBatchPriority" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          {{ t('issueList.changePriority') }}
        </button>
        <div v-if="showBatchPriority" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 py-1 w-32">
          <button v-for="p in priorityOptions" :key="p.value" @click="batchChangePriority(p.value)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-200">{{ p.label }}</button>
        </div>
      </div>

      <!-- 批量分配 -->
      <div class="relative">
        <button @click="showBatchAssign = !showBatchAssign" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          {{ t('issueList.batchAssign') }}
        </button>
        <div v-if="showBatchAssign" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 p-2 w-48">
          <UserSelect v-model="batchAssigneeId" :users="memberOptions" :placeholder="t('issueList.selectAssignee')" @update:model-value="batchAssign" />
        </div>
      </div>

      <!-- 批量删除 -->
      <button @click="execBatchDelete" class="px-2.5 py-1 text-xs border border-red-300 dark:border-red-700 rounded-md bg-white dark:bg-gray-700 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors">
        {{ t('issueList.batchDelete') }}
      </button>

      <button @click="clearSelection" class="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">{{ t('issueList.clearSelection') }}</button>
    </div>

    <!-- Success toast -->
    <div v-if="toastMessage" class="fixed top-4 right-4 z-50 bg-green-50 dark:bg-green-900/50 border border-green-200 dark:border-green-800 text-green-700 dark:text-green-300 text-sm px-4 py-2 rounded-md shadow-lg transition-opacity">
      {{ toastMessage }}
    </div>

    <!-- 列表内容 -->
    <div class="overflow-x-auto">
      <div v-if="loading" class="text-center py-16">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
        <p class="mt-2 text-gray-500 text-sm">{{ t('common.loading') }}</p>
      </div>
      <div v-else-if="issues.length === 0" class="text-center py-16">
        <svg class="h-12 w-12 text-gray-300 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
        <p class="mt-2 text-gray-500">{{ t('cycle.noIssues') }}</p>
        <p class="mt-1 text-sm text-gray-400">{{ t('issueList.noIssuesHint') }}</p>
      </div>
      <table v-else class="w-full">
        <thead class="bg-gray-50 border-b border-gray-200 sticky top-0">
          <tr>
            <th class="w-10 px-3 py-2.5 text-left">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" class="rounded border-gray-300" />
            </th>
            <th v-for="col in visibleColumns" :key="col.key"
              class="px-3 py-2.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
              :class="col.width"
            >{{ t(col.labelKey || "") || col.label }}</th>
            <th class="px-3 py-2.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-20">{{ t('issueList.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="issue in issues" :key="issue.id"
            class="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer transition-colors"
            :class="{ 'bg-indigo-50/50': selectedIds.has(issue.id) }">
            <td class="px-3 py-2.5" @click.stop>
              <input type="checkbox" :checked="selectedIds.has(issue.id)" @change="toggleSelect(issue.id)" class="rounded border-gray-300" />
            </td>
            <td v-for="col in visibleColumns" :key="col.key" class="px-3 py-2.5" @click="col.key !== 'actions' && $emit('select', issue)">
              <!-- 编号 -->
              <span v-if="col.key === 'sequence_id'" class="text-xs text-gray-400 font-mono">{{ projectIdentifier }}-{{ issue.sequence_id }}</span>
              <!-- 标题 -->
              <span v-else-if="col.key === 'name'" class="text-sm text-gray-800 font-medium line-clamp-2 hover:text-indigo-600 transition-colors">{{ issue.name }}</span>
              <!-- 优先级 -->
              <span v-else-if="col.key === 'priority'" :class="priorityClass(issue.priority)" class="text-xs px-1.5 py-0.5 rounded whitespace-nowrap">{{ priorityLabel(issue.priority) }}</span>
              <!-- 类型 -->
              <span v-else-if="col.key === 'issue_type'" class="text-xs whitespace-nowrap">
                <span v-if="issue.issue_type" class="inline-flex items-center gap-1">
                  <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: issue.issue_type.color }"></span>
                  <span class="text-gray-600">{{ issue.issue_type.name }}</span>
                </span>
                <span v-else class="text-gray-400">-</span>
              </span>
              <!-- 状态 -->
              <span v-else-if="col.key === 'state'" class="text-xs text-gray-600 whitespace-nowrap">{{ getStateName(issue.state_id) }}</span>
              <!-- 负责人 -->
              <div v-else-if="col.key === 'assignees'" class="flex -space-x-1">
                <div v-for="(a, idx) in (issue.assignees || []).slice(0, 3)" :key="a.id"
                  class="w-5 h-5 rounded-full border border-white flex items-center justify-center text-[10px] font-medium text-white ring-2 ring-white"
                  :style="{ backgroundColor: assigneeColor(idx) }" :title="a.display_name || a.username">{{ getInitials(a.display_name || a.username) }}</div>
                <span v-if="!issue.assignees?.length" class="text-xs text-gray-400">-</span>
              </div>
              <!-- 周期 -->
              <span v-else-if="col.key === 'cycle'" class="text-xs text-gray-500 whitespace-nowrap">{{ getCycleName(issue) }}</span>
              <!-- 日期 -->
              <span v-else-if="col.key === 'start_date'" class="text-xs text-gray-500">{{ formatDate(issue.start_date) }}</span>
              <span v-else-if="col.key === 'target_date'" class="text-xs text-gray-500">{{ formatDate(issue.target_date) }}</span>
              <span v-else-if="col.key === 'created_at'" class="text-xs text-gray-400">{{ formatDate(issue.created_at) }}</span>
              <!-- 自定义字段 -->
              <span v-else-if="col.key.startsWith('cf_')" class="text-xs text-gray-600 whitespace-nowrap">{{ getCFValue(issue.id, col.key) }}</span>
            </td>
            <td class="px-3 py-2.5" @click.stop>
              <div class="flex items-center gap-1">
                <button @click="$emit('select', issue)" class="text-xs text-indigo-600 hover:text-indigo-800 font-medium">{{ t('issueList.view') }}</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div v-if="totalPages > 1" class="px-4 py-3 border-t border-gray-200 flex items-center justify-between bg-gray-50/50">
      <span class="text-sm text-gray-500">{{ t('issueList.total') }} {{ totalCount }} {{ t('common.items') }}</span>
      <div class="flex items-center gap-1">
        <button @click="page--" :disabled="page <= 1" class="px-3 py-1 border rounded text-sm disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-100 transition-colors">{{ t('issueList.prevPage') }}</button>
        <template v-for="p in visiblePages" :key="p">
          <button v-if="p === '...'" disabled class="px-2 py-1 text-sm text-gray-400">...</button>
          <button v-else @click="page = Number(p)" class="px-3 py-1 border rounded text-sm transition-colors"
            :class="page === Number(p) ? 'bg-indigo-600 text-white border-indigo-600' : 'hover:bg-gray-100'">{{ p }}</button>
        </template>
        <button @click="page++" :disabled="page >= totalPages" class="px-3 py-1 border rounded text-sm disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-100 transition-colors">{{ t('issueList.nextPage') }}</button>
      </div>
    </div>

    <ImportIssuesModal
      :visible="showImportModal"
      :project-id="projectId"
      :workspace-id="workspaceId"
      @close="showImportModal = false"
      @success="onImportSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import issueApi from '@/api/issue'
import customFieldApi from '@/api/custom-field'
import projectApi from '@/api/project'
import api from '@/api'
import UserSelect from '@/components/UserSelect.vue'
import { RQLInput } from '@/components/RQL'
import { useRQL } from '@/composables/useRQL'
import { useConfirm } from '@/composables/useConfirm'
import QuickCreateInput from '@/components/QuickCreateInput.vue'
import ImportIssuesModal from '@/components/ImportIssuesModal.vue'
import QuickFilterChips from '@/components/QuickFilterChips.vue'
import * as issueTypeApi from '@/api/issue-type'

const props = defineProps<{ projectId: number; workspaceId: number }>()
const router = useRouter()
const { t, locale } = useI18n()

const emit = defineEmits<{
  (e: 'select', issue: any): void
  (e: 'delete', issue: any): void
}>()

// ---- Project Identifier ----
const projectIdentifier = ref('')

// ---- Column config ----
interface ColumnDef { key: string; label?: string; labelKey?: string; width: string; defaultVisible: boolean }
const staticColumns: ColumnDef[] = [
  { key: 'sequence_id', labelKey: 'issueList.columnSequenceId', width: 'w-20', defaultVisible: true },
  { key: 'name',         labelKey: 'issueList.columnName', width: '',       defaultVisible: true },
  { key: 'priority',     labelKey: 'issue.priority', width: 'w-20', defaultVisible: true },
  { key: 'issue_type',   labelKey: 'issue.type', width: 'w-20', defaultVisible: true },
  { key: 'state',        labelKey: 'issue.state',   width: 'w-28', defaultVisible: true },
  { key: 'assignees',    labelKey: 'issue.assignee', width: 'w-28', defaultVisible: true },
  { key: 'cycle',        labelKey: 'issue.cycle',   width: 'w-24', defaultVisible: true },
  { key: 'start_date',   labelKey: 'issue.startDate', width: 'w-28', defaultVisible: false },
  { key: 'target_date',  labelKey: 'issue.targetDate', width: 'w-28', defaultVisible: false },
  { key: 'created_at',   labelKey: 'issueList.columnCreatedAt', width: 'w-36', defaultVisible: false },
]

const STORAGE_KEY = 'issuelist_columns'

const customFields = ref<any[]>([])
const cfFilters = ref<Map<number, string>>(new Map())

// Dynamic columns: static + custom fields
const effectiveColumns = computed(() => {
  const cols = [...staticColumns]
  for (const cf of customFields.value) {
    cols.push({ key: 'cf_' + cf.id, label: cf.name, width: 'w-28', defaultVisible: false } as ColumnDef)
  }
  return cols
})

function loadColumnPrefs(): Set<string> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return new Set(JSON.parse(raw))
  } catch { /* */ }
  return new Set(effectiveColumns.value.filter(c => c.defaultVisible).map(c => c.key))
}

const visibleColumnKeys = ref(loadColumnPrefs())
const showColumns = ref(false)
const showRQL = ref(false)

const visibleColumns = computed(() => effectiveColumns.value.filter(c => visibleColumnKeys.value.has(c.key)))

function toggleColumn(key: string) {
  const s = new Set(visibleColumnKeys.value)
  s.has(key) ? s.delete(key) : s.add(key)
  visibleColumnKeys.value = s
  saveColumnPrefs()
}

function saveColumnPrefs() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify([...visibleColumnKeys.value]))
}

// ---- RQL Search ----
const { rql: rqlQuery, loading: rqlLoading, error: rqlError, search: doRQLSearch, results: rqlResults } = useRQL()

const onRQLSearch = async () => {
  await doRQLSearch(props.projectId, 'issue')
  if (!rqlError.value && rqlResults.value.length > 0) {
    issues.value = rqlResults.value
  }
}

// ---- Filter System ----
const issues = ref<any[]>([])
const states = ref<any[]>([])
const cycles = ref<any[]>([])
const members = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const limit = ref(20)
const totalCount = ref(0)
const totalPages = ref(1)
const selectedIds = ref(new Set<number>())
const showFilterDropdown = ref(false)
const showImportModal = ref(false)

// Active filter picker state
const activeFilterPicker = ref<{ key: string; label: string } | null>(null)
const activeCFValue = ref('')

const filters = ref({
  search: '', state_id: 0, priority: '', cycle_id: 0, assignee_id: 0,
  start_date: '', end_date: ''
})

// Track which filters are "active" (user has explicitly set them)
const activeFilterKeys = ref<Set<string>>(new Set())

const activeFilterCount = computed(() => activeFilterKeys.value.size + cfFilters.value.size)

interface FilterChip { key: string; label: string; value: string }

const activeFilterChips = computed<FilterChip[]>(() => {
  const chips: FilterChip[] = []
  if (activeFilterKeys.value.has('state_id') && filters.value.state_id > 0) {
    const s = states.value.find((s: any) => s.id === filters.value.state_id)
    chips.push({ key: 'state_id', label: t('issue.state'), value: s?.name || String(filters.value.state_id) })
  }
  if (activeFilterKeys.value.has('priority') && filters.value.priority) {
    const labels: Record<string, string> = { urgent: t('issue.priorityUrgent'), high: t('issue.priorityHigh'), medium: t('issue.priorityMedium'), low: t('issue.priorityLow'), none: t('issue.priorityNone') }
    chips.push({ key: 'priority', label: t('issue.priority'), value: labels[filters.value.priority] || filters.value.priority })
  }
  if (activeFilterKeys.value.has('assignee_id') && filters.value.assignee_id > 0) {
    const m = members.value.find((m: any) => m.user_id === filters.value.assignee_id)
    chips.push({ key: 'assignee_id', label: t('issue.assignee'), value: m?.user?.display_name || m?.user?.username || String(filters.value.assignee_id) })
  }
  if (activeFilterKeys.value.has('cycle_id') && filters.value.cycle_id > 0) {
    const c = cycles.value.find((c: any) => c.id === filters.value.cycle_id)
    chips.push({ key: 'cycle_id', label: t('issue.cycle'), value: c?.name || String(filters.value.cycle_id) })
  }
  if (activeFilterKeys.value.has('start_date') && filters.value.start_date) {
    chips.push({ key: 'start_date', label: t('issue.startDate'), value: filters.value.start_date })
  }
  if (activeFilterKeys.value.has('target_date') && filters.value.end_date) {
    chips.push({ key: 'target_date', label: t('issue.targetDate'), value: filters.value.end_date })
  }
  cfFilters.value.forEach((value, fieldId) => {
    const cf = customFields.value.find((f: any) => f.id === fieldId)
    chips.push({ key: 'cf_' + fieldId, label: cf?.name || t('issue.customFields'), value })
  })
  return chips
})

function getFilterValueLabel(key: string): string {
  if (key === 'state_id' && filters.value.state_id > 0) {
    const s = states.value.find((s: any) => s.id === filters.value.state_id)
    return s?.name || ''
  }
  if (key === 'priority' && filters.value.priority) {
    const labels: Record<string, string> = { urgent: t('issue.priorityUrgent'), high: t('issue.priorityHigh'), medium: t('issue.priorityMedium'), low: t('issue.priorityLow'), none: t('issue.priorityNone') }
    return labels[filters.value.priority] || ''
  }
  if (key === 'assignee_id' && filters.value.assignee_id > 0) {
    const m = members.value.find((m: any) => m.user_id === filters.value.assignee_id)
    return m?.user?.display_name || m?.user?.username || ''
  }
  if (key === 'cycle_id' && filters.value.cycle_id > 0) {
    const c = cycles.value.find((c: any) => c.id === filters.value.cycle_id)
    return c?.name || ''
  }
  if (key === 'start_date' && filters.value.start_date) return filters.value.start_date
  if (key === 'target_date' && filters.value.end_date) return filters.value.end_date
  return ''
}

function addFilter(key: string) {
  showFilterDropdown.value = false
  const labels: Record<string, string> = {
    state_id: t('issue.state'), priority: t('issue.priority'), assignee_id: t('issue.assignee'),
    cycle_id: t('issue.cycle'), start_date: t('issue.startDate'), target_date: t('issue.targetDate')
  }
  activeFilterPicker.value = { key, label: labels[key] || key }
}

function addCFFilter(fieldId: number) {
  showFilterDropdown.value = false
  const cf = customFields.value.find((f: any) => f.id === fieldId)
  activeFilterPicker.value = { key: 'cf', label: cf?.name || t('issue.customFields') }
}

function applyFilterPicker() {
  if (!activeFilterPicker.value) return
  if (activeFilterPicker.value.key === 'cf') {
    if (activeCFValue.value.trim()) {
      const fieldId = customFields.value.find((f: any) => f.name === activeFilterPicker.value!.label)?.id
      if (fieldId) cfFilters.value.set(fieldId, activeCFValue.value.trim())
    }
    activeCFValue.value = ''
  } else {
    activeFilterKeys.value.add(activeFilterPicker.value.key)
  }
  activeFilterPicker.value = null
  showFilterDropdown.value = false
  search()
}

function cancelFilterPicker() {
  activeFilterPicker.value = null
  activeCFValue.value = ''
}

function removeFilter(key: string) {
  if (key.startsWith('cf_')) {
    const fieldId = parseInt(key.replace('cf_', ''))
    cfFilters.value.delete(fieldId)
  } else {
    activeFilterKeys.value.delete(key)
    // Reset the filter value
    if (key === 'state_id') filters.value.state_id = 0
    else if (key === 'priority') filters.value.priority = ''
    else if (key === 'cycle_id') filters.value.cycle_id = 0
    else if (key === 'assignee_id') filters.value.assignee_id = 0
    else if (key === 'start_date') filters.value.start_date = ''
    else if (key === 'target_date') filters.value.end_date = ''
  }
  search()
}

function clearAllFilters() {
  activeFilterKeys.value.clear()
  cfFilters.value.clear()
  filters.value = { search: '', state_id: 0, priority: '', cycle_id: 0, assignee_id: 0, start_date: '', end_date: '' }
  search()
}

// Close filter dropdown when clicking outside
function handleClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.relative')) {
    showFilterDropdown.value = false
    showColumns.value = false
  }
}

// ---- Batch Operations ----
const showBatchState = ref(false)
const showBatchPriority = ref(false)
const showBatchAssign = ref(false)
const batchAssigneeId = ref<number | undefined>(undefined)

const priorityOptions = computed(() => [
  { value: 'urgent', label: t('issue.priorityUrgent') },
  { value: 'high', label: t('issue.priorityHigh') },
  { value: 'medium', label: t('issue.priorityMedium') },
  { value: 'low', label: t('issue.priorityLow') },
  { value: 'none', label: t('issue.priorityNone') },
])

const toastMessage = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

function showToast(msg: string) {
  toastMessage.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMessage.value = '' }, 2500)
}

const { confirm } = useConfirm()

// ---- Computed ----
const isAllSelected = computed(() => issues.value.length > 0 && issues.value.every(i => selectedIds.value.has(i.id)))

const memberOptions = computed(() => members.value.map((m: any) => ({
  id: m.user_id,
  display_name: m.user?.display_name || m.user?.username,
  email: m.user?.email
})))

const filtersAssignee = computed({
  get: () => filters.value.assignee_id > 0 ? filters.value.assignee_id : undefined,
  set: (v: number | undefined) => { filters.value.assignee_id = v || 0 }
})

const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const tp = totalPages.value; const p = page.value
  if (tp <= 7) { for (let i = 1; i <= tp; i++) pages.push(i); return pages }
  pages.push(1); if (p > 3) pages.push('...')
  for (let i = Math.max(2, p - 1); i <= Math.min(tp - 1, p + 1); i++) pages.push(i)
  if (p < tp - 2) pages.push('...'); pages.push(tp)
  return pages
})

// ---- Helpers ----
function priorityClass(p: string) {
  const m: Record<string, string> = { urgent: 'bg-red-100 text-red-700', high: 'bg-orange-100 text-orange-700', medium: 'bg-yellow-100 text-yellow-700', low: 'bg-green-100 text-green-700', none: 'bg-gray-100 text-gray-500' }
  return m[p] || m.none
}
function priorityLabel(p: string) { const m: Record<string, string> = { urgent: t('issue.priorityUrgent'), high: t('issue.priorityHigh'), medium: t('issue.priorityMedium'), low: t('issue.priorityLow'), none: t('issue.priorityNone') }; return m[p] || p }
function getStateName(id: number) { return states.value.find((s: any) => s.id === id)?.name || '-' }
function getCycleName(i: any) { return i.cycle?.name || i.cycle_link?.name || '-' }
function getInitials(n: string) { return (n || '?')[0]?.toUpperCase() || '?' }
function assigneeColor(i: number) { return ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'][i % 6] }
function formatDate(d: string | null | undefined) { if (!d) return '-'; return new Date(d).toLocaleDateString(locale.value) }

// ---- Actions ----
function goToCreate() { router.push(`/workspaces/${props.workspaceId}/projects/${props.projectId}/issues/new`) }

function onQuickCreated() { page.value = 1; loadIssues() }
function onImportSuccess() { page.value = 1; loadIssues() }
function search() { page.value = 1; loadIssues() }

function handleQuickFilter(quickFilters: Record<string, any>) {
  if (quickFilters.assignee_id === 'me') {
    filters.value.assignee_id = 1
    activeFilterKeys.value.add('assignee_id')
  } else if (quickFilters.assignee_id === null) {
    filters.value.assignee_id = 0
    activeFilterKeys.value.delete('assignee_id')
  }
  if (quickFilters.priority) {
    filters.value.priority = quickFilters.priority
    activeFilterKeys.value.add('priority')
  }
  search()
}

function toggleSelectAll() {
  if (isAllSelected.value) selectedIds.value.clear()
  else issues.value.forEach(i => selectedIds.value.add(i.id))
  selectedIds.value = new Set(selectedIds.value)
}
function toggleSelect(id: number) {
  const s = new Set(selectedIds.value); s.has(id) ? s.delete(id) : s.add(id); selectedIds.value = s
}

async function batchChangeState(stateId: number) {
  showBatchState.value = false
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { state_id: stateId })
    clearSelection()
    showToast(t('issueList.toastStateUpdated'))
    loadIssues()
  } catch (e) { console.error('Batch state failed:', e) }
}

async function batchChangePriority(priority: string) {
  showBatchPriority.value = false
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { priority: priority as any })
    clearSelection()
    showToast(t('issueList.toastPriorityUpdated'))
    loadIssues()
  } catch (e) { console.error('Batch priority failed:', e) }
}

async function batchAssign(userId: string | number | undefined) {
  showBatchAssign.value = false
  if (!userId) return
  const uid = typeof userId === 'string' ? Number(userId) : userId
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { assignee_ids: [uid] })
    clearSelection()
    showToast(t('issueList.toastAssigned'))
    loadIssues()
  } catch (e) { console.error('Batch assign failed:', e) }
}

function clearSelection() {
  selectedIds.value = new Set()
  showBatchState.value = false
  showBatchPriority.value = false
  showBatchAssign.value = false
}

async function execBatchDelete() {
  if (!(await confirm(t('issueList.confirmDelete').replace('{0}', String(selectedIds.value.size))))) return
  try {
    await issueApi.bulkDeleteIssues([...selectedIds.value])
    clearSelection()
    showToast(t('issueList.toastDeleted'))
    loadIssues()
  } catch (e) { console.error('Batch delete failed:', e) }
}

// ---- Data loading ----
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
	  if (props.workspaceId > 0) Promise.all([loadIssues(), loadStates(), loadCycles(), loadMembers(), loadCustomFields(), loadIssueTypes(), loadProjectInfo()])
	})

	watch(page, () => loadIssues())
	watch(() => props.workspaceId, (id) => {
	  if (id > 0) Promise.all([loadIssues(), loadStates(), loadCycles(), loadMembers(), loadCustomFields(), loadIssueTypes(), loadProjectInfo()])
	})
async function loadProjectInfo() {
  try {
    const project = await projectApi.getProject(props.projectId)
    projectIdentifier.value = project.identifier || 'PROJ'
  } catch { /* */ }
}

async function loadIssues() {
  loading.value = true
  try {
    const params: any = { limit: limit.value, offset: (page.value - 1) * limit.value }
    if (filters.value.state_id && filters.value.state_id > 0) params.state_id = filters.value.state_id
    if (filters.value.priority) params.priority = filters.value.priority
    if (filters.value.cycle_id && filters.value.cycle_id > 0) params.cycle_id = filters.value.cycle_id
    if (filters.value.assignee_id && filters.value.assignee_id > 0) params.assignee_id = filters.value.assignee_id
    if (filters.value.search) params.search = filters.value.search
    // Pass CF filters
    if (cfFilters.value.size > 0) {
      const conditions = Array.from(cfFilters.value.entries()).map(([field_id, value]) => ({ field_id, value }))
      params.cf_and = JSON.stringify(conditions)
    }
    const result = await issueApi.listIssues(props.projectId, props.workspaceId, params)
    issues.value = result.items; totalCount.value = result.total; totalPages.value = Math.max(1, Math.ceil(result.total / limit.value))
  } catch (e) { console.error('Failed to load issues:', e) }
  finally { loading.value = false }
}

async function loadStates() { try { const r = await api.get(`/projects/${props.projectId}/settings/states`); states.value = r.data } catch (e) { /* */ } }
async function loadCycles() { try { const r = await api.get(`/projects/${props.projectId}/cycles`); cycles.value = r.data } catch (e) { /* */ } }
async function loadMembers() { try { const r = await api.get(`/workspaces/${props.workspaceId}/members`); members.value = r.data } catch (e) { /* */ } }
async function loadCustomFields() {
  try { customFields.value = await customFieldApi.listCustomFields(props.workspaceId, props.projectId) } catch (e) { /* */ }
}
const issueTypes = ref<any[]>([])
async function loadIssueTypes() {
  try { issueTypes.value = await issueTypeApi.getIssueTypes(props.workspaceId, props.projectId) } catch (e) { /* */ }
}

// Custom field value cache
const cfValueCache = ref<Record<number, Record<number, string>>>({})

async function loadCFValues(issueIds: number[]) {
  if (!issueIds.length) return
  try {
    const results = await Promise.all(issueIds.map(id =>
      customFieldApi.listIssueCustomFieldValues(id).catch(() => [] as any[])
    ))
    const cache: Record<number, Record<number, string>> = {}
    issueIds.forEach((id, idx) => {
      cache[id] = {}
      for (const v of (results[idx] || [])) {
        cache[id][v.field_id] = v.value || ''
      }
    })
    cfValueCache.value = cache
  } catch { /* */ }
}

function getCFValue(issueId: number, colKey: string): string {
  const fieldId = parseInt(colKey.replace('cf_', ''))
  return cfValueCache.value[issueId]?.[fieldId] || '-'
}

watch(issues, (newIssues) => {
  if (newIssues.length) loadCFValues(newIssues.map(i => i.id))
})
</script>
