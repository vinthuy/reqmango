/**
 * Apply missing i18n keys and rewrite locale files (also collapses duplicate top-level keys).
 * Run: node scripts/patch-missing-i18n.mjs
 */
import fs from 'node:fs'
import path from 'node:path'

const root = path.resolve('src/locales')

function deepMerge(target, source) {
  for (const [k, v] of Object.entries(source || {})) {
    if (v && typeof v === 'object' && !Array.isArray(v)) {
      if (!target[k] || typeof target[k] !== 'object' || Array.isArray(target[k])) target[k] = {}
      deepMerge(target[k], v)
    } else {
      target[k] = v
    }
  }
  return target
}

const patches = {
  common: {
    zh: {
      refresh: '刷新',
      add: '添加',
      creating: '创建中...',
      done: '完成',
      error: '错误',
      filter: '筛选',
      justNow: '刚刚',
      label: '标签',
      view: '查看',
      assignee: '负责人',
    },
    en: {
      refresh: 'Refresh',
      add: 'Add',
      creating: 'Creating...',
      done: 'Done',
      error: 'Error',
      filter: 'Filter',
      justNow: 'Just now',
      label: 'Label',
      view: 'View',
      assignee: 'Assignee',
    },
  },
  agent: {
    zh: { feedbackUp: '有用', feedbackDown: '无用' },
    en: { feedbackUp: 'Helpful', feedbackDown: 'Not helpful' },
  },
  ai: {
    zh: {
      saveSuccess: 'AI 设置已保存',
      agentTask: {
        namePlaceholder: '输入任务名称',
        descriptionPlaceholder: '输入任务描述（可选）',
      },
    },
    en: {
      saveSuccess: 'AI settings saved',
      agentTask: {
        namePlaceholder: 'Enter task name',
        descriptionPlaceholder: 'Enter task description (optional)',
      },
    },
  },
  approvals: {
    zh: {
      approveHint: '批准后工作项将按审批流继续流转',
      rejectHint: '拒绝后工作项将回到审批前状态',
    },
    en: {
      approveHint: 'After approval the work item continues along the workflow',
      rejectHint: 'After rejection the work item returns to the pre-approval state',
    },
  },
  auth: {
    zh: { identifierHint: '用于 URL 的唯一标识，例如 DEMO' },
    en: { identifierHint: 'Unique URL slug, e.g. DEMO' },
  },
  automation: {
    zh: { viewHistory: '查看历史' },
    en: { viewHistory: 'View history' },
  },
  automationBuilder: {
    zh: { greaterThan: '大于', lessThan: '小于' },
    en: { greaterThan: 'Greater than', lessThan: 'Less than' },
  },
  automationForm: {
    zh: { agentNamePlaceholder: 'Agent 名称或 ID' },
    en: { agentNamePlaceholder: 'Agent name or ID' },
  },
  import: {
    zh: {
      colTitle: '标题',
      colTitleHint: '必填',
      colDescription: '描述',
      colDescriptionHint: '可选',
      colPriority: '优先级',
      colState: '状态',
      colStateHint: '状态名称',
      colType: '类型',
      colTypeHint: '类型名称',
      colAssignees: '负责人',
      colAssigneesHint: '邮箱，逗号分隔',
      colLabels: '标签',
      colLabelsHint: '标签名，逗号分隔',
      colStartDate: '开始日期',
      colTargetDate: '截止日期',
      colParentTitle: '父工作项标题',
      colParentTitleHint: '用于建立父子关系',
      colModule: '模块',
      colModuleHint: '模块名称',
      colCycle: '周期',
      colCycleHint: '周期名称',
      colEstimate: '估算',
      colEstimateHint: '估算点数',
      exampleTitle: '修复登录问题',
      exampleDescription: '用户无法登录',
      exampleState: '待办',
      exampleType: 'Bug',
      exampleParentTitle: '用户认证',
      examplePlaceholderTitle: '示例工作项 1',
      examplePlaceholderTitle2: '示例工作项 2',
    },
    en: {
      colTitle: 'Title',
      colTitleHint: 'Required',
      colDescription: 'Description',
      colDescriptionHint: 'Optional',
      colPriority: 'Priority',
      colState: 'State',
      colStateHint: 'State name',
      colType: 'Type',
      colTypeHint: 'Type name',
      colAssignees: 'Assignees',
      colAssigneesHint: 'Emails, comma-separated',
      colLabels: 'Labels',
      colLabelsHint: 'Label names, comma-separated',
      colStartDate: 'Start date',
      colTargetDate: 'Due date',
      colParentTitle: 'Parent title',
      colParentTitleHint: 'Used to link parent/child',
      colModule: 'Module',
      colModuleHint: 'Module name',
      colCycle: 'Cycle',
      colCycleHint: 'Cycle name',
      colEstimate: 'Estimate',
      colEstimateHint: 'Estimate points',
      exampleTitle: 'Fix login issue',
      exampleDescription: 'Users cannot sign in',
      exampleState: 'Todo',
      exampleType: 'Bug',
      exampleParentTitle: 'User auth',
      examplePlaceholderTitle: 'Sample issue 1',
      examplePlaceholderTitle2: 'Sample issue 2',
    },
  },
  issue: {
    zh: { issues: '工作项', remove: '移除', unknownError: '未知错误' },
    en: { issues: 'Issues', remove: 'Remove', unknownError: 'Unknown error' },
  },
  issueKanban: {
    zh: { dragRevert: '移动失败，已还原' },
    en: { dragRevert: 'Move failed, reverted' },
  },
  issueTypePage: {
    zh: { field: '字段' },
    en: { field: 'Field' },
  },
  moduleCard: {
    zh: {
      submodule: '子模块',
      topLevel: '顶级模块',
      viewDetails: '查看详情',
      editOverride: '编辑覆盖',
      override: '覆盖',
      exclude: '排除',
      resetOverride: '重置覆盖',
      delete: '删除',
    },
    en: {
      submodule: 'Submodule',
      topLevel: 'Top-level',
      viewDetails: 'View details',
      editOverride: 'Edit override',
      override: 'Override',
      exclude: 'Exclude',
      resetOverride: 'Reset override',
      delete: 'Delete',
    },
  },
  settings: {
    zh: { workspaceStatesDesc: '管理工作区级状态，可被项目继承' },
    en: { workspaceStatesDesc: 'Manage workspace-level states inherited by projects' },
  },
  webhook: {
    zh: { events: '事件' },
    en: { events: 'Events' },
  },
  workflow: {
    zh: {
      selectAtLeastOneApprover: '请至少选择一名审批人',
      selectFromState: '请选择起始状态',
      workflowNotFound: '未找到工作流',
    },
    en: {
      selectAtLeastOneApprover: 'Select at least one approver',
      selectFromState: 'Select a from-state',
      workflowNotFound: 'Workflow not found',
    },
  },
  workflowRule: {
    zh: { unassign: '取消分配' },
    en: { unassign: 'Unassign' },
  },
  workspace: {
    zh: { settingsDesc: '管理工作区基本信息、成员与类型' },
    en: { settingsDesc: 'Manage workspace profile, members, and types' },
  },
  workspaceIssueType: {
    zh: {
      addPropertiesDesc: '选择要绑定到此类型的自定义字段',
      addPropertiesTitle: '添加属性到',
      addProperty: '添加属性',
      allFieldsBound: '所有可用字段均已绑定',
      customProperties: '自定义属性',
      customPropertiesDesc: '绑定到此工作项类型的自定义字段',
      noProperties: '暂无绑定属性',
      required: '必填',
    },
    en: {
      addPropertiesDesc: 'Choose custom fields to bind to this type',
      addPropertiesTitle: 'Add properties to',
      addProperty: 'Add property',
      allFieldsBound: 'All available fields are already bound',
      customProperties: 'Custom properties',
      customPropertiesDesc: 'Custom fields bound to this issue type',
      noProperties: 'No properties bound yet',
      required: 'Required',
    },
  },
}

function patchFile(file, lang) {
  const data = JSON.parse(fs.readFileSync(file, 'utf8'))
  for (const [ns, langs] of Object.entries(patches)) {
    if (!data[ns] || typeof data[ns] !== 'object') data[ns] = {}
    deepMerge(data[ns], langs[lang])
  }
  fs.writeFileSync(file, `${JSON.stringify(data, null, 2)}\n`, 'utf8')
  console.log('patched', path.basename(file), 'common.refresh=', data.common.refresh)
}

patchFile(path.join(root, 'zh-CN.json'), 'zh')
patchFile(path.join(root, 'en-US.json'), 'en')
