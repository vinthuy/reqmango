<template>
  <div class="h-full flex flex-col bg-gray-50/50">
    <!-- Header Bar -->
    <div class="flex items-center justify-between px-6 py-3 bg-white border-b border-gray-200 shrink-0">
      <div class="flex items-center gap-3">
        <h2 class="text-base font-semibold text-gray-800">{{ t('metrics.title') }}</h2>
        <span class="text-xs text-gray-400">{{ t('metrics.chartCount', { count: charts.length }) }}</span>
      </div>
      <button @click="openPanel('new')"
        class="inline-flex items-center gap-1.5 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
        </svg>
        {{ t('metrics.createChart') }}
      </button>
    </div>

    <!-- Main Content -->
    <div class="flex-1 overflow-y-auto p-6">
      <!-- Loading -->
      <div v-if="loading && charts.length === 0" class="flex items-center justify-center py-20">
        <svg class="animate-spin h-6 w-6 text-gray-300" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      </div>

      <!-- Empty State -->
      <div v-else-if="charts.length === 0" class="flex flex-col items-center justify-center py-20">
        <div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
          <svg class="w-8 h-8 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
          </svg>
        </div>
        <p class="text-sm text-gray-500 mb-1">{{ t('metrics.empty') }}</p>
        <p class="text-xs text-gray-400 mb-4">{{ t('metrics.emptyHint') }}</p>
        <button @click="openPanel('new')"
          class="inline-flex items-center gap-1.5 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
          </svg>
          {{ t('metrics.createChart') }}
        </button>
      </div>

      <!-- Chart Grid -->
      <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <MetricsChartCard
          v-for="chart in charts"
          :key="chart.id"
          :chart="chart"
          :project-id="projectId"
          @edit="openPanel('edit', chart)"
          @delete="handleDeleteChart(chart.id)"
        />
      </div>
    </div>

    <!-- Side Panel -->
    <Teleport to="body">
      <div v-if="panel.visible" class="fixed inset-0 z-50 flex justify-end">
        <div class="absolute inset-0 bg-black/30" @click="closePanel"></div>
        <div class="relative w-[680px] bg-white shadow-2xl flex flex-col animate-slide-in">
          <!-- Header -->
          <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 shrink-0">
            <h3 class="text-sm font-semibold text-gray-800">
              {{ panel.mode === 'edit' ? t('metrics.editChart') : t('metrics.createChart') }}
            </h3>
            <button @click="closePanel" class="p-1 text-gray-400 hover:text-gray-600 rounded">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>

          <!-- Body: Two-column layout -->
          <div class="flex-1 flex overflow-hidden">
            <!-- Left: Config -->
            <div class="w-[340px] border-r border-gray-100 overflow-y-auto p-4 space-y-4">
              <!-- Step 1: Template Selection (new mode only) -->
              <div v-if="panel.mode === 'new' && !panel.selectedTemplate && !panel.useCustom">
                <p class="text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-2">{{ t('metrics.selectTemplate') }}</p>
                <div class="flex gap-1 mb-3 bg-gray-100 rounded-lg p-0.5">
                  <button v-for="cat in categories" :key="cat.id"
                    @click="panel.activeCategory = cat.id"
                    :class="['flex-1 px-2 py-1 text-[11px] font-medium rounded-md transition-colors',
                      panel.activeCategory === cat.id ? 'bg-white shadow-sm text-gray-800' : 'text-gray-500 hover:text-gray-700']"
                  >{{ cat.name }}</button>
                </div>
                <div class="space-y-1.5">
                  <button v-for="tpl in currentCategoryTemplates" :key="tpl.id"
                    @click="selectTemplate(tpl)"
                    class="w-full flex items-center gap-2.5 p-2.5 bg-gray-50 hover:bg-indigo-50 hover:border-indigo-200 border border-gray-100 rounded-lg text-left transition-colors group"
                  >
                    <span class="text-base">{{ getTemplateIcon(tpl.icon) }}</span>
                    <div class="min-w-0 flex-1">
                      <p class="text-xs font-medium text-gray-700 group-hover:text-indigo-700 truncate">{{ tpl.name }}</p>
                      <p class="text-[10px] text-gray-400 truncate">{{ tpl.description }}</p>
                    </div>
                    <svg class="w-3.5 h-3.5 text-gray-300 group-hover:text-indigo-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                    </svg>
                  </button>
                </div>
                <div class="mt-3 pt-3 border-t border-gray-100">
                  <button @click="panel.useCustom = true; panel.selectedTemplate = null; resetForm()"
                    class="w-full flex items-center justify-center gap-2 py-2 border border-dashed border-gray-300 rounded-lg text-xs text-gray-500 hover:text-indigo-600 hover:border-indigo-300 transition-colors">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                    </svg>
                    {{ t('metrics.customCreate') }}
                  </button>
                </div>
              </div>

              <!-- Step 2: Configuration Form -->
              <div v-if="panel.mode === 'edit' || panel.selectedTemplate || panel.useCustom">
                <button v-if="panel.mode === 'new'" @click="backToTemplates"
                  class="flex items-center gap-1 text-[11px] text-gray-400 hover:text-indigo-600 transition-colors mb-2">
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
                  </svg>
                  {{ t('metrics.backToTemplates') }}
                </button>

                <!-- Chart Name -->
                <div>
                  <label class="block text-[11px] font-medium text-gray-500 mb-1">{{ t('metrics.chartName') }}</label>
                  <input v-model="form.name" type="text" :placeholder="t('metrics.chartNamePlaceholder')"
                    class="w-full px-3 py-1.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
                </div>

                <!-- Chart Type -->
                <div>
                  <label class="block text-[11px] font-medium text-gray-500 mb-1">
                    {{ t('metrics.chartType') }}
                    <span v-if="isTemplateMode" class="text-indigo-400 ml-1">{{ t('metrics.templateLocked') }}</span>
                    <button v-if="isTemplateMode" @click="panel.selectedTemplate = null"
                      class="ml-2 text-[10px] text-gray-400 hover:text-indigo-500 underline">{{ t('metrics.switchToCustom') }}</button>
                  </label>
                  <div v-if="isTemplateMode" class="flex items-center gap-2 px-3 py-2 bg-indigo-50 rounded-lg border border-indigo-100">
                    <span class="text-sm">{{ chartTypeOptions.find(c => c.value === form.chart_type)?.icon }}</span>
                    <span class="text-[11px] text-indigo-700 font-medium">{{ chartTypeOptions.find(c => c.value === form.chart_type)?.label }}</span>
                    <svg class="w-3 h-3 text-indigo-400 ml-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
                  </div>
                  <div v-else class="grid grid-cols-5 gap-1">
                    <button v-for="ct in chartTypeOptions" :key="ct.value" @click="updateChartType(ct.value)"
                      :class="['flex flex-col items-center gap-0.5 px-1 py-1.5 rounded-md border text-[10px] transition-all',
                        form.chart_type === ct.value
                          ? 'border-indigo-300 bg-indigo-50 text-indigo-700 font-medium'
                          : 'border-gray-100 bg-white text-gray-500 hover:border-gray-200']"
                    >
                      <span class="text-sm">{{ ct.icon }}</span>
                      <span>{{ ct.label }}</span>
                    </button>
                  </div>
                </div>

                <!-- X Axis -->
                <div>
                  <label class="block text-[11px] font-medium text-gray-500 mb-1">
                    {{ t('metrics.xAxisDimension') }}
                    <span v-if="isTemplateMode" class="text-indigo-400 ml-1">{{ t('metrics.templateLocked') }}</span>
                  </label>
                  <div v-if="isTemplateMode" class="flex items-center gap-2 px-3 py-2 bg-indigo-50 rounded-lg border border-indigo-100">
                    <span class="text-[11px] text-indigo-700 font-medium">{{ form.x_axis }}</span>
                    <svg class="w-3 h-3 text-indigo-400 ml-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
                  </div>
                  <select v-else v-model="form.x_axis"
                    class="w-full px-3 py-1.5 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                    <optgroup :label="t('metrics.categoryDimensions')">
                      <option v-for="d in xAxisCategoryOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                    <optgroup :label="t('metrics.timeCreated')">
                      <option v-for="d in xAxisTimeCreatedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                    <optgroup :label="t('metrics.timeCompleted')">
                      <option v-for="d in xAxisTimeCompletedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                    <optgroup :label="t('metrics.timeUpdated')">
                      <option v-for="d in xAxisTimeUpdatedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                    <optgroup v-if="customFields.length > 0" :label="t('metrics.customFields')">
                      <option v-for="cf in customFields" :key="'cf-'+cf.id" :value="'custom_field:'+cf.id">{{ cf.name }}</option>
                    </optgroup>
                  </select>
                </div>

                <!-- Y Axis -->
                <div>
                  <label class="block text-[11px] font-medium text-gray-500 mb-1">{{ t('metrics.yAxisMetric') }}</label>
                  <div v-if="isTemplateMode" class="flex items-center gap-2 px-3 py-2 bg-indigo-50 rounded-lg border border-indigo-100">
                    <span class="text-[11px] text-indigo-700 font-medium">{{ yAxisOptions.find(m => m.value === form.y_axis)?.label || form.y_axis }}</span>
                    <svg class="w-3 h-3 text-indigo-400 ml-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
                  </div>
                  <select v-else v-model="form.y_axis"
                    class="w-full px-3 py-1.5 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                    <optgroup :label="t('metrics.builtinMetrics')">
                      <option v-for="m in yAxisOptions" :key="m.value" :value="m.value">{{ m.label }}</option>
                    </optgroup>
                    <optgroup v-if="customFields.length > 0" :label="t('metrics.customFields')">
                      <option v-for="cf in customFields" :key="'cf-avg-'+cf.id" :value="'custom_field_avg:'+cf.id">{{ cf.name }} ({{ t('metrics.average') }})</option>
                      <option v-for="cf in customFields" :key="'cf-count-'+cf.id" :value="'custom_field_count:'+cf.id">{{ cf.name }} ({{ t('metrics.count') }})</option>
                    </optgroup>
                  </select>
                </div>

                <!-- Jira-style Filters -->
                <div>
                  <div class="flex items-center justify-between mb-2">
                    <label class="text-[11px] font-medium text-gray-500">{{ t('metrics.filterConditions') }}</label>
                    <button @click="addFilter" class="inline-flex items-center gap-0.5 text-[11px] text-indigo-500 hover:text-indigo-700">
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
                      {{ t('metrics.addCondition') }}
                    </button>
                  </div>
                  <div v-if="filters.length === 0" class="flex items-center gap-2 px-3 py-2 bg-gray-50 rounded-lg border border-dashed border-gray-200">
                    <svg class="w-4 h-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"/></svg>
                    <span class="text-[11px] text-gray-400">{{ t('metrics.noFiltersHint') }}</span>
                  </div>
                  <div v-for="(f, i) in filters" :key="i" class="relative mb-2">
                    <div class="flex items-center gap-1.5">
                      <!-- AND badge -->
                      <span v-if="i > 0" class="text-[9px] font-bold text-indigo-400 bg-indigo-50 px-1.5 py-0.5 rounded shrink-0">{{ t('metrics.and') }}</span>
                      <!-- Field selector -->
                      <select v-model="f.field" @change="f.values = []; f.value = ''"
                        class="flex-1 px-2 py-1.5 border border-gray-200 rounded text-[11px] bg-white focus:outline-none focus:ring-1 focus:ring-indigo-400">
                        <option value="">{{ t('metrics.selectField') }}</option>
                        <option value="state">{{ t('metrics.fieldState') }}</option>
                        <option value="priority">{{ t('metrics.fieldPriority') }}</option>
                        <option value="assignee">{{ t('metrics.fieldAssignee') }}</option>
                        <option value="type">{{ t('metrics.fieldType') }}</option>
                        <option value="label">{{ t('metrics.fieldLabel') }}</option>
                        <option value="module">{{ t('metrics.fieldModule') }}</option>
                        <option value="created_by">{{ t('metrics.fieldCreatedBy') }}</option>
                        <option value="title">{{ t('metrics.fieldTitle') }}</option>
                        <optgroup v-if="customFields.length > 0" :label="t('metrics.customFields')">
                          <option v-for="cf in customFields" :key="'filter-cf-'+cf.id" :value="'custom_field:'+cf.id">{{ cf.name }}</option>
                        </optgroup>
                      </select>
                      <!-- Operator selector -->
                      <select v-model="f.operator" @change="f.values = []; f.value = ''"
                        class="w-24 px-2 py-1.5 border border-gray-200 rounded text-[11px] bg-white focus:outline-none focus:ring-1 focus:ring-indigo-400">
                        <option value="=">{{ t('metrics.opEqual') }}</option>
                        <option value="!=">{{ t('metrics.opNotEqual') }}</option>
                        <option value="in">{{ t('metrics.opIn') }}</option>
                        <option value="not_in">{{ t('metrics.opNotIn') }}</option>
                        <option value="contains">{{ t('metrics.opContains') }}</option>
                        <option value="empty">{{ t('metrics.opEmpty') }}</option>
                        <option value="not_empty">{{ t('metrics.opNotEmpty') }}</option>
                      </select>
                      <!-- Value: Multi-select for in/not_in -->
                      <div v-if="f.operator === 'in' || f.operator === 'not_in'" class="flex-1 relative" @click="$event.stopPropagation()">
                        <div @click="f._showDropdown = !f._showDropdown"
                          class="flex items-center gap-1 px-2 py-1.5 border border-gray-200 rounded text-[11px] bg-white cursor-pointer hover:border-indigo-300 min-h-[30px] flex-wrap">
                          <span v-for="(v, vi) in f.values" :key="vi"
                            class="inline-flex items-center gap-0.5 bg-indigo-50 text-indigo-700 px-1.5 py-0.5 rounded text-[10px]">
                            {{ v }}
                            <button @click.stop="f.values.splice(vi, 1)" class="text-indigo-400 hover:text-indigo-600">&times;</button>
                          </span>
                          <span v-if="f.values.length === 0" class="text-gray-400">{{ t('metrics.selectMultipleValues') }}</span>
                        </div>
                        <div v-if="f._showDropdown" class="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg max-h-40 overflow-y-auto">
                          <!-- Search input -->
                          <div class="sticky top-0 bg-white border-b border-gray-100 px-2 py-1">
                            <input v-model="f._search" :placeholder="t('metrics.search')" @input="$event.stopPropagation()"
                              class="w-full px-2 py-1 border border-gray-200 rounded text-[11px] focus:outline-none focus:ring-1 focus:ring-indigo-400" />
                          </div>
                          <label v-for="v in filterDropdownValues(f)" :key="v"
                            class="flex items-center gap-2 px-3 py-1.5 text-[11px] hover:bg-gray-50 cursor-pointer">
                            <input type="checkbox" :value="v" v-model="f.values"
                              class="w-3 h-3 rounded border-gray-300 text-indigo-500 focus:ring-indigo-400">
                            <span class="text-gray-700">{{ v }}</span>
                          </label>
                          <div v-if="filterDropdownValues(f).length === 0" class="px-3 py-2 text-[11px] text-gray-400">{{ t('metrics.noMatches') }}</div>
                        </div>
                      </div>
                      <!-- Value: Single select/input for other operators -->
                      <template v-else-if="f.operator !== 'empty' && f.operator !== 'not_empty'">
                        <div class="flex-1 relative" @click="$event.stopPropagation()">
                          <select v-if="filterDropdownValues(f).length > 0 && !f._showDropdown"
                            v-model="f.value" @click="f._showDropdown = true"
                            class="flex-1 px-2 py-1.5 border border-gray-200 rounded text-[11px] bg-white focus:outline-none focus:ring-1 focus:ring-indigo-400">
                            <option value="">{{ t('metrics.selectValue') }}</option>
                            <option v-for="v in filterDropdownValues(f)" :key="v" :value="v">{{ v }}</option>
                          </select>
                          <div v-else-if="f._showDropdown" class="w-full">
                            <input v-model="f._search" :placeholder="t('metrics.searchField', { field: fieldLabelMap[f.field] })"
                              class="w-full px-2 py-1.5 border border-gray-200 rounded text-[11px] focus:outline-none focus:ring-1 focus:ring-indigo-400"
                              @input="$event.stopPropagation()" @blur="f._showDropdown = false" />
                            <div class="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg max-h-40 overflow-y-auto">
                              <div v-if="f.value" @click="f.value = ''; f._search = ''; f._showDropdown = false"
                                class="px-3 py-1.5 text-[11px] text-red-500 hover:bg-red-50 cursor-pointer border-b border-gray-100">{{ t('metrics.clearSelection') }}</div>
                              <div v-for="v in filterDropdownValues(f)" :key="v"
                                @click="f.value = v; f._search = ''; f._showDropdown = false"
                                class="px-3 py-1.5 text-[11px] hover:bg-indigo-50 cursor-pointer"
                                :class="{'bg-indigo-50 text-indigo-700': f.value === v}">
                                {{ v }}
                              </div>
                              <div v-if="filterDropdownValues(f).length === 0" class="px-3 py-2 text-[11px] text-gray-400">{{ t('metrics.noMatches') }}</div>
                            </div>
                          </div>
                          <input v-else v-model="f.value" :placeholder="t('metrics.inputValue')"
                            class="flex-1 px-2 py-1.5 border border-gray-200 rounded text-[11px] focus:outline-none focus:ring-1 focus:ring-indigo-400" />
                        </div>
                      </template>
                      <!-- Remove button -->
                      <button @click="removeFilter(i)" class="p-1 text-gray-300 hover:text-red-500 shrink-0 rounded hover:bg-red-50 transition-colors">
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Right: Preview -->
            <div class="flex-1 flex flex-col bg-gray-50">
              <div class="flex items-center justify-between px-4 py-2 border-b border-gray-100 bg-white shrink-0">
                <span class="text-[11px] font-medium text-gray-400">{{ t('metrics.chartPreview') }}</span>
                <button @click="fetchPreview" :disabled="previewLoading"
                  class="px-3 py-1 text-[11px] font-medium text-indigo-600 bg-indigo-50 rounded-md hover:bg-indigo-100 disabled:opacity-50 transition-colors">
                  {{ previewLoading ? t('metrics.loading') : t('metrics.preview') }}
                </button>
              </div>
              <div class="flex-1 flex items-center justify-center p-4">
                <div v-if="!hasData && !previewLoading" class="text-center">
                  <svg class="w-10 h-10 text-gray-200 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                  </svg>
                  <p class="text-xs text-gray-400">{{ previewError || t('metrics.previewHint') }}</p>
                </div>
                <div v-else-if="previewLoading" class="text-center">
                  <svg class="animate-spin h-6 w-6 text-indigo-300 mx-auto mb-2" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  <p class="text-xs text-gray-400">{{ t('metrics.loadingPreviewData') }}</p>
                </div>
                <div v-else class="w-full h-full">
                  <MetricsChartCard
                    :chart="previewChart"
                    :project-id="projectId"
                    class="shadow-none border-0 h-full"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="flex items-center justify-end gap-2 px-5 py-3 border-t border-gray-100 shrink-0">
            <button @click="closePanel"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">
              {{ t('common.cancel') }}
            </button>
            <button @click="handleSave" :disabled="saving || !canSave"
              class="px-5 py-2 text-sm font-medium bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
              {{ saving ? t('common.saving') : (panel.mode === 'edit' ? t('common.update') : t('common.create')) }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import type { TemplateCategory, MetricTemplate, MetricChart, CreateChartPayload, MetricChartConfig } from '@/types/metrics'
import { metricsApi } from '@/api/metrics'
import { useConfirm } from '@/composables/useConfirm'
import { useToast } from '@/composables/useToast'
import { useI18n } from '@/composables/useI18n'
import MetricsChartCard from '@/components/metrics/MetricsChartCard.vue'

const props = defineProps<{ projectId: number }>()
const { confirm } = useConfirm()
const toast = useToast()
const { t } = useI18n()

// ── Data ──
const categories = ref<TemplateCategory[]>([])
const charts = ref<MetricChart[]>([])
const loading = ref(false)
const saving = ref(false)

// ── Side Panel ──
const panel = reactive({
  visible: false,
  mode: 'new' as 'new' | 'edit',
  activeCategory: 'agile',
  selectedTemplate: null as MetricTemplate | null,
  useCustom: false,
  editingChart: null as MetricChart | null,
})

// ── Form ──
const form = reactive({ name: '', chart_type: 'bar', x_axis: 'state', y_axis: 'count' })
const filters = reactive<Array<{ field: string; operator: string; value: string; values: string[]; _showDropdown: boolean; _search: string }>>([])
const advancedConfig = reactive<MetricChartConfig>({ stack_mode: 'none', show_labels: false, reference_lines: [] })
const canSave = computed(() => form.name.trim().length > 0)
const isTemplateMode = computed(() => !!panel.selectedTemplate && panel.useCustom)

// ── Custom Fields ──
const customFields = ref<Array<{ id: number; name: string; field_type: string }>>([])

async function loadCustomFields() {
  try {
    customFields.value = await metricsApi.getCustomFields(props.projectId)
  } catch { /* ignore */ }
}

// ── Filter Values (for dropdown) ──
const filterValues = ref<Record<string, string[]> & { custom_fields?: Record<string, string[]> }>({})
const filterValuesLoaded = ref(false)
async function loadFilterValues() {
  try {
    const data = await metricsApi.getFilterValues(props.projectId)
    filterValues.value = data || {}
    filterValuesLoaded.value = true
  } catch {
    filterValues.value = {}
    filterValuesLoaded.value = true
  }
}

// Local fallback values for common enum fields (used when API returns empty)
const localFilterDefaults: Record<string, string[]> = {
  state: ['待处理', '进行中', '已完成', '已关闭', '已拒绝'],
  priority: ['高', '中', '低', '紧急', '非常紧急'],
  type: ['Bug', 'Feature', 'Improvement', 'Task', 'Story', 'Epic'],
  label: [],
  module: [],
  assignee: [],
  reporter: [],
  created_by: [],
}

function getFilterFieldValues(field: string): string[] {
  if (!field) return []
  if (field.startsWith('custom_field:')) {
    const fieldId = field.split(':')[1]
    return filterValues.value.custom_fields?.[fieldId] || filterValues.value.custom_fields?.[String(fieldId)] || []
  }
  // Use API values first, fallback to local defaults
  const apiValues = filterValues.value[field] || []
  if (apiValues.length > 0) return apiValues
  return localFilterDefaults[field] || []
}

const fieldLabelMap: Record<string, string> = {
  state: t('metrics.dimensions.state'), priority: t('metrics.dimensions.priority'), assignee: t('metrics.dimensions.assignee'), type: t('metrics.dimensions.type'),
  label: t('metrics.dimensions.label'), module: t('metrics.dimensions.module'), created_by: t('metrics.fieldCreatedBy'), title: t('metrics.fieldTitle'),
}

function filterDropdownValues(f: { field: string; _search?: string }): string[] {
  const all = getFilterFieldValues(f.field)
  const q = (f._search || '').toLowerCase()
  if (!q) return all
  return all.filter(v => v.toLowerCase().includes(q))
}

// ── Build RQL from filters ──
function buildFiltersPayload(validFilters: Array<{ field: string; operator: string; value: string; values: string[] }>) {
  if (validFilters.length === 0) return undefined
  const rqlParts = validFilters.map(f => {
    if (f.operator === 'empty') return `${f.field} IS NULL`
    if (f.operator === 'not_empty') return `${f.field} IS NOT NULL`
    if (f.operator === 'in') return `(${f.values.map(v => `${f.field} = "${v}"`).join(' OR ')})`
    if (f.operator === 'not_in') return `${f.field} NOT IN (${f.values.map(v => `"${v}"`).join(', ')})`
    if (f.operator === 'contains') return `${f.field} ~ "${f.value}"`
    return `${f.field} ${f.operator} "${f.value}"`
  })
  return { rql: rqlParts.join(' AND '), conditions: validFilters }
}

// ── Preview ──
const previewData = ref<{ labels: string[]; values: number[]; colors?: string[] } | null>(null)
const previewLoading = ref(false)
const previewError = ref('')
const hasData = computed(() => previewData.value && previewData.value.labels.length > 0)

const previewChart = computed<MetricChart>(() => ({
  id: 0, project_id: props.projectId, creator_id: 0, name: form.name || t('metrics.preview'),
  chart_type: form.chart_type, x_axis: form.x_axis, y_axis: form.y_axis,
  filters: '', config: '', template_id: '',
  data_labels: previewData.value?.labels || [],
  data_values: previewData.value?.values || [],
  data_colors: previewData.value?.colors || [],
  sort_order: 0, is_visible: true, created_at: '', updated_at: '',
}))

// ── Options ──
const chartTypeOptions = [
  { value: 'bar', label: t('metrics.chartTypes.bar'), icon: '📊' },
  { value: 'line', label: t('metrics.chartTypes.line'), icon: '📈' },
  { value: 'pie', label: t('metrics.chartTypes.pie'), icon: '🥧' },
  { value: 'doughnut', label: t('metrics.chartTypes.doughnut'), icon: '🍩' },
  { value: 'area', label: t('metrics.chartTypes.area'), icon: '🏔' },
  { value: 'radar', label: t('metrics.chartTypes.radar'), icon: '🕷' },
  { value: 'scatter', label: t('metrics.chartTypes.scatter'), icon: '⚬' },
  { value: 'bubble', label: t('metrics.chartTypes.bubble'), icon: '🫧' },
  { value: 'mixed', label: t('metrics.chartTypes.mixed'), icon: '📋' },
  { value: 'table', label: t('metrics.chartTypes.table'), icon: '📑' },
]
const xAxisCategoryOptions = [
  { value: 'state', label: t('metrics.dimensions.state') }, { value: 'priority', label: t('metrics.dimensions.priority') },
  { value: 'assignee', label: t('metrics.dimensions.assignee') }, { value: 'type', label: t('metrics.dimensions.type') },
  { value: 'label', label: t('metrics.dimensions.label') }, { value: 'cycle', label: t('metrics.dimensions.cycle') },
  { value: 'module', label: t('metrics.dimensions.module') }, { value: 'state_group', label: t('metrics.dimensions.state_group') },
  { value: 'created_by', label: t('metrics.dimensions.created_by') },
]
const xAxisTimeCreatedOptions = [
  { value: 'created_day', label: t('metrics.dimensions.byDay') }, { value: 'created_week', label: t('metrics.dimensions.byWeek') }, { value: 'created_month', label: t('metrics.dimensions.byMonth') },
]
const xAxisTimeCompletedOptions = [
  { value: 'completed_day', label: t('metrics.dimensions.byDay') }, { value: 'completed_week', label: t('metrics.dimensions.byWeek') }, { value: 'completed_month', label: t('metrics.dimensions.byMonth') },
]
const xAxisTimeUpdatedOptions = [
  { value: 'updated_day', label: t('metrics.dimensions.byDay') }, { value: 'updated_week', label: t('metrics.dimensions.byWeek') }, { value: 'updated_month', label: t('metrics.dimensions.byMonth') },
]
const yAxisOptions = [
  { value: 'count', label: t('metrics.yAxisOptions.count') },
  { value: 'avg_processing_time', label: t('metrics.yAxisOptions.avg_processing_time') },
  { value: 'current_retention', label: t('metrics.yAxisOptions.current_retention') },
  { value: 'created_vs_resolved', label: t('metrics.yAxisOptions.created_vs_resolved') },
  { value: 'completion_rate', label: t('metrics.yAxisOptions.completion_rate') },
  { value: 'avg_cycle_time', label: t('metrics.yAxisOptions.avg_cycle_time') },
  { value: 'throughput', label: t('metrics.yAxisOptions.throughput') },
  { value: 'wip_count', label: t('metrics.yAxisOptions.wip_count') },
  { value: 'backlog_count', label: t('metrics.yAxisOptions.backlog_count') },
  { value: 'overdue_count', label: t('metrics.yAxisOptions.overdue_count') },
]
const currentCategoryTemplates = computed(() =>
  categories.value.find(c => c.id === panel.activeCategory)?.templates || []
)
const templateIconMap: Record<string, string> = {
  flame: '🔥', 'trending-up': '📈', layers: '📚', clock: '⏱', timer: '⏲', columns: '📊',
  package: '📦', target: '🎯', 'alert-triangle': '⚠️', 'bar-chart-2': '📊', users: '👥',
  search: '🔍', shield: '🛡', 'check-circle': '✅', 'pause-circle': '⏸',
}
function getTemplateIcon(icon: string) { return templateIconMap[icon] || '📊' }

// ── Data Loading ──
async function loadData() {
  loading.value = true
  try {
    const [tplData, chartData] = await Promise.all([
      metricsApi.listTemplates(props.projectId),
      metricsApi.listCharts(props.projectId),
    ])
    categories.value = tplData.categories || tplData
    charts.value = chartData.charts || chartData
  } catch (e) {
    console.error('Failed to load metrics data:', e)
  } finally {
    loading.value = false
  }
}

// ── Panel Actions ──
function openPanel(mode: 'new' | 'edit', chart?: MetricChart) {
  panel.mode = mode
  panel.visible = true
  panel.selectedTemplate = null
  panel.useCustom = false
  panel.editingChart = null
  previewData.value = null
  previewError.value = ''
  filters.splice(0)
  loadFilterValues()
  loadCustomFields()

  if (mode === 'edit' && chart) {
    panel.editingChart = chart
    panel.useCustom = true
    form.name = chart.name
    form.chart_type = chart.chart_type
    form.x_axis = chart.x_axis
    form.y_axis = chart.y_axis
    // Parse filters from JSON
    try {
      const f = JSON.parse(chart.filters || '{}')
      if (f.conditions) {
        f.conditions.forEach((c: any) => {
          const values = c.values || (c.value ? [c.value] : [])
          filters.push({ field: c.field || '', operator: c.operator || '=', value: c.value || '', values, _showDropdown: false, _search: '' })
        })
      } else if (f.rql) {
        // Convert legacy RQL to visual filters (best effort)
        const match = f.rql.match(/^(\w+)\s*(!?=)\s*"?([^"]+)"?$/)
        if (match) filters.push({ field: match[1], operator: match[2], value: match[3], values: [match[3]], _showDropdown: false, _search: '' })
      }
    } catch { /* ignore */ }
    try {
      const cfg = JSON.parse(chart.config || '{}')
      advancedConfig.stack_mode = cfg.stack_mode || 'none'
      advancedConfig.show_labels = cfg.show_labels || false
      advancedConfig.reference_lines = cfg.reference_lines || []
    } catch { /* ignore */ }
  } else {
    resetForm()
  }
}

function closePanel() {
  panel.visible = false
  panel.selectedTemplate = null
  panel.useCustom = false
  panel.editingChart = null
  previewData.value = null
}

function selectTemplate(tpl: MetricTemplate) {
  panel.selectedTemplate = tpl
  panel.useCustom = true
  form.name = tpl.name
  form.chart_type = tpl.chart_type
  form.x_axis = tpl.default_x_axis
  form.y_axis = tpl.default_y_axis
  filters.splice(0)
  // 应用模板默认筛选条件
  if (tpl.default_filters) {
    if (tpl.default_filters.type_filter) {
      const typeMap: Record<string, string> = {
        bug: 'Bug', feature: 'Feature', improvement: 'Improvement',
        task: 'Task', story: 'Story', epic: 'Epic',
      }
      const typeVal = typeMap[tpl.default_filters.type_filter] || tpl.default_filters.type_filter
      filters.push({ field: 'type', operator: '=', value: typeVal, values: [], _showDropdown: false, _search: '' })
    }
  }
  if (tpl.default_config) {
    advancedConfig.stack_mode = tpl.default_config.stack_mode || 'none'
    advancedConfig.show_labels = tpl.default_config.show_labels || false
    advancedConfig.reference_lines = tpl.default_config.reference_lines || []
  }
}

function backToTemplates() {
  panel.selectedTemplate = null
  panel.useCustom = false
  previewData.value = null
  resetForm()
}

function resetForm() {
  form.name = ''
  form.chart_type = 'bar'
  form.x_axis = 'state'
  form.y_axis = 'count'
  filters.splice(0)
  advancedConfig.stack_mode = 'none'
  advancedConfig.show_labels = false
  advancedConfig.reference_lines = []
}

function updateChartType(type: string) {
  form.chart_type = type
}

function addFilter() {
  filters.push({ field: '', operator: '=', value: '', values: [], _showDropdown: false, _search: '' })
}

function removeFilter(index: number) {
  filters.splice(index, 1)
}

// ── Preview ──

async function fetchPreview() {
  if (!form.x_axis || !form.y_axis) return
  previewLoading.value = true
  previewError.value = ''
  try {
    const payload: any = {
      name: form.name || 'preview',
      chart_type: form.chart_type,
      x_axis: form.x_axis,
      y_axis: form.y_axis,
    }
    // Build filters
    const validFilters = filters.filter(f => {
      if (!f.field) return false
      if (f.operator === 'empty' || f.operator === 'not_empty') return true
      if (f.operator === 'in' || f.operator === 'not_in') return f.values.length > 0
      return f.value !== ''
    })
    payload.filters = buildFiltersPayload(validFilters)
    const res = await metricsApi.previewChart(props.projectId, payload)
    previewData.value = { labels: res.labels || [], values: res.values || [], colors: res.colors || [] }
  } catch (e: any) {
    previewError.value = e?.response?.data?.error || t('metrics.previewFailed')
    previewData.value = null
  } finally {
    previewLoading.value = false
  }
}

// ── Save ──
async function handleSave() {
  if (!canSave.value || saving.value) return
  saving.value = true
  try {
    const validFilters = filters.filter(f => {
      if (!f.field) return false
      if (f.operator === 'empty' || f.operator === 'not_empty') return true
      if (f.operator === 'in' || f.operator === 'not_in') return f.values.length > 0
      return f.value !== ''
    })
    const filtersPayload = buildFiltersPayload(validFilters)
    const payload: CreateChartPayload = {
      name: form.name.trim(),
      chart_type: form.chart_type,
      x_axis: form.x_axis,
      y_axis: form.y_axis,
      filters: filtersPayload,
      config: { stack_mode: advancedConfig.stack_mode, show_labels: advancedConfig.show_labels, reference_lines: advancedConfig.reference_lines },
    }
    if (panel.selectedTemplate) payload.template_id = panel.selectedTemplate.id

    if (panel.mode === 'edit' && panel.editingChart) {
      await metricsApi.updateChart(props.projectId, panel.editingChart.id, payload)
      toast.success(t('metrics.chartUpdated'))
    } else {
      await metricsApi.createChart(props.projectId, payload)
      toast.success(t('metrics.chartCreated'))
    }
    closePanel()
    await loadData()
  } catch (e: any) {
    console.error('Failed to save chart:', e)
    toast.error(e?.response?.data?.message || t('metrics.saveFailed'))
  } finally {
    saving.value = false
  }
}

// ── Delete ──
async function handleDeleteChart(chartId: number) {
  const ok = await confirm({ title: t('metrics.deleteChart'), message: t('metrics.deleteConfirmSimple'), confirmText: t('common.delete'), danger: true })
  if (!ok) return
  try {
    await metricsApi.deleteChart(props.projectId, chartId)
    await loadData()
  } catch (e) {
    console.error('Failed to delete chart:', e)
  }
}

onMounted(loadData)
</script>

<style scoped>
@keyframes slide-in { from { transform: translateX(100%); } to { transform: translateX(0); } }
.animate-slide-in { animation: slide-in 0.25s ease-out; }
</style>
