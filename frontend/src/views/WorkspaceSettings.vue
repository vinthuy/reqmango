<script setup lang="ts">import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import TypeTemplateManager from '@/components/TypeTemplateManager.vue';
import ProjectTemplateManager from '@/components/ProjectTemplateManager.vue';
import type { WorkItemType, CustomField, State, Workflow, Automation, Label } from '@/types';

const router = useRouter();

// 导航状态
const activeSection = ref('states');

// 状态分组数据 - 符合Plane设计
const stateGroups = ref([
  { 
    id: 'backlog', 
    name: 'Backlog', 
    states: [
      { id: 1, name: 'Backlog', color: '#9CA3AF', is_default: true }
    ]
  },
  { 
    id: 'unstarted', 
    name: 'Unstarted', 
    states: [
      { id: 2, name: 'Todo', color: '#6B7280', is_default: false }
    ]
  },
  { 
    id: 'started', 
    name: 'Started', 
    states: [
      { id: 3, name: 'In Progress', color: '#3B82F6', is_default: false },
      { id: 4, name: 'In Review', color: '#F59E0B', is_default: false }
    ]
  },
  { 
    id: 'completed', 
    name: 'Completed', 
    states: [
      { id: 5, name: 'Done', color: '#22C55E', is_default: false }
    ]
  },
  { 
    id: 'cancelled', 
    name: 'Cancelled', 
    states: [
      { id: 6, name: 'Cancelled', color: '#EF4444', is_default: false }
    ]
  }
]);

// 工作项类型数据
const workItemTypes = ref<WorkItemType[]>([
  { id: 1, name: 'Task', description: 'A task to be completed', icon: '📋', color: '#3B82F6', is_active: true },
  { id: 2, name: 'Bug', description: 'A bug or issue', icon: '🐛', color: '#EF4444', is_active: true },
  { id: 3, name: 'Feature', description: 'A new feature', icon: '✨', color: '#22C55E', is_active: true },
  { id: 4, name: 'Epic', description: 'A large feature', icon: '🎯', color: '#8B5CF6', is_active: true },
  { id: 5, name: 'Story', description: 'A user story', icon: '📖', color: '#F59E0B', is_active: true },
]);

// 自定义字段数据
const customFields = ref<CustomField[]>([
  { id: 1, name: 'Story Points', type: 'number', description: 'Effort estimation', is_required: false },
  { id: 2, name: 'Sprint', type: 'dropdown', description: 'Sprint number', options: ['Sprint 1', 'Sprint 2', 'Sprint 3'], is_required: false },
  { id: 3, name: 'Severity', type: 'dropdown', description: 'Issue severity', options: ['Critical', 'High', 'Medium', 'Low'], is_required: true },
  { id: 4, name: 'Test Suite', type: 'text', description: 'Related test suite', is_required: false },
  { id: 5, name: 'Release Date', type: 'date', description: 'Target release date', is_required: false },
]);

// 标签数据
const labels = ref<Label[]>([
  { id: 1, name: 'Frontend', color: '#3B82F6' },
  { id: 2, name: 'Backend', color: '#22C55E' },
  { id: 3, name: 'Documentation', color: '#F59E0B' },
  { id: 4, name: 'DevOps', color: '#8B5CF6' },
  { id: 5, name: 'Security', color: '#EF4444' },
]);

// 自动化规则数据
const automations = ref<Automation[]>([
  { 
    id: 1, 
    name: 'Auto-assign bugs', 
    description: 'Assign new bugs to QA lead',
    trigger: 'work_item_created',
    conditions: [{ field: 'type', operator: 'is', value: 'Bug' }],
    actions: [{ type: 'change_property', field: 'assignee', value: 'qa-lead' }],
    is_enabled: true 
  },
  { 
    id: 2, 
    name: 'Mark stale items', 
    description: 'Mark items inactive for 30 days',
    trigger: 'scheduled',
    conditions: [],
    actions: [{ type: 'add_label', field: 'label', value: 'stale' }],
    is_enabled: false 
  },
]);

// 工作流数据
const workflows = ref([
  { id: 1, name: 'Default Workflow', states: ['Backlog', 'Todo', 'In Progress', 'In Review', 'Done'] },
  { id: 2, name: 'Bug Workflow', states: ['Backlog', 'Reproducing', 'Fixing', 'Testing', 'Done'] },
]);

// 模板数据
const templates = ref([
  { id: 1, name: 'Software Development', description: 'Standard SDLC template', work_item_count: 12 },
  { id: 2, name: 'Bug Tracking', description: 'For bug tracking workflow', work_item_count: 5 },
]);

// Modal状态
const showStateModal = ref(false);
const showTypeModal = ref(false);
const showFieldModal = ref(false);
const showLabelModal = ref(false);
const showAutomationModal = ref(false);
const showTemplateModal = ref(false);
const editingState = ref<{ groupId: string; state: any } | null>(null);
const editingType = ref<WorkItemType | null>(null);
const editingField = ref<CustomField | null>(null);
const editingLabel = ref<Label | null>(null);

// 新建表单
const newStateForm = ref({ name: '', color: '#3B82F6', groupId: '' });
const newTypeForm = ref({ name: '', description: '', icon: '📋', color: '#3B82F6' });
const newFieldForm = ref({ name: '', type: 'text', description: '', options: [] as string[], is_required: false });
const newLabelForm = ref({ name: '', color: '#3B82F6' });

// 计算属性
const totalStates = computed(() => stateGroups.value.reduce((sum, group) => sum + group.states.length, 0));

// 方法
const handleStateClick = (groupId: string, state: any) => {
  editingState.value = { groupId, state };
  newStateForm.value = { name: state.name, color: state.color, groupId };
  showStateModal.value = true;
};

const handleAddState = (groupId: string) => {
  newStateForm.value = { name: '', color: '#3B82F6', groupId };
  editingState.value = null;
  showStateModal.value = true;
};

const handleSaveState = () => {
  if (!newStateForm.value.name) return;
  
  if (editingState.value) {
    const group = stateGroups.value.find(g => g.id === editingState.value?.groupId);
    if (group) {
      const state = group.states.find(s => s.id === editingState.value?.state.id);
      if (state) {
        state.name = newStateForm.value.name;
        state.color = newStateForm.value.color;
      }
    }
  } else {
    const group = stateGroups.value.find(g => g.id === newStateForm.value.groupId);
    if (group) {
      group.states.push({
        id: Date.now(),
        name: newStateForm.value.name,
        color: newStateForm.value.color,
        is_default: false
      });
    }
  }
  showStateModal.value = false;
};

const handleDeleteState = (groupId: string, stateId: number) => {
  const group = stateGroups.value.find(g => g.id === groupId);
  if (group) {
    group.states = group.states.filter(s => s.id !== stateId);
  }
};

const handleTypeClick = (type: WorkItemType) => {
  editingType.value = { ...type };
  newTypeForm.value = { name: type.name, description: type.description || '', icon: type.icon, color: type.color };
  showTypeModal.value = true;
};

const handleAddType = () => {
  editingType.value = null;
  newTypeForm.value = { name: '', description: '', icon: '📋', color: '#3B82F6' };
  showTypeModal.value = true;
};

const handleSaveType = () => {
  if (!newTypeForm.value.name) return;
  
  if (editingType.value) {
    const type = workItemTypes.value.find(t => t.id === editingType.value?.id);
    if (type) {
      type.name = newTypeForm.value.name;
      type.description = newTypeForm.value.description;
      type.icon = newTypeForm.value.icon;
      type.color = newTypeForm.value.color;
    }
  } else {
    workItemTypes.value.push({
      id: Date.now(),
      name: newTypeForm.value.name,
      description: newTypeForm.value.description,
      icon: newTypeForm.value.icon,
      color: newTypeForm.value.color,
      is_active: true
    });
  }
  showTypeModal.value = false;
};

const handleFieldClick = (field: CustomField) => {
  editingField.value = { ...field, options: field.options || [] };
  newFieldForm.value = { 
    name: field.name, 
    type: field.type, 
    description: field.description || '', 
    options: field.options || [],
    is_required: field.is_required 
  };
  showFieldModal.value = true;
};

const handleAddField = () => {
  editingField.value = null;
  newFieldForm.value = { name: '', type: 'text', description: '', options: [], is_required: false };
  showFieldModal.value = true;
};

const handleSaveField = () => {
  if (!newFieldForm.value.name) return;
  
  if (editingField.value) {
    const field = customFields.value.find(f => f.id === editingField.value?.id);
    if (field) {
      field.name = newFieldForm.value.name;
      field.type = newFieldForm.value.type;
      field.description = newFieldForm.value.description;
      field.options = newFieldForm.value.options;
      field.is_required = newFieldForm.value.is_required;
    }
  } else {
    customFields.value.push({
      id: Date.now(),
      name: newFieldForm.value.name,
      type: newFieldForm.value.type,
      description: newFieldForm.value.description,
      options: newFieldForm.value.options,
      is_required: newFieldForm.value.is_required
    });
  }
  showFieldModal.value = false;
};

const handleLabelClick = (label: Label) => {
  editingLabel.value = { ...label };
  newLabelForm.value = { name: label.name, color: label.color };
  showLabelModal.value = true;
};

const handleAddLabel = () => {
  editingLabel.value = null;
  newLabelForm.value = { name: '', color: '#3B82F6' };
  showLabelModal.value = true;
};

const handleSaveLabel = () => {
  if (!newLabelForm.value.name) return;
  
  if (editingLabel.value) {
    const label = labels.value.find(l => l.id === editingLabel.value?.id);
    if (label) {
      label.name = newLabelForm.value.name;
      label.color = newLabelForm.value.color;
    }
  } else {
    labels.value.push({
      id: Date.now(),
      name: newLabelForm.value.name,
      color: newLabelForm.value.color
    });
  }
  showLabelModal.value = false;
};

const handleDeleteLabel = (labelId: number) => {
  labels.value = labels.value.filter(l => l.id !== labelId);
};

const handleAddAutomation = () => {
  showAutomationModal.value = true;
};

const handleViewWorkflow = (workflowId: number) => {
  router.push(`/workspace/demo1/settings/workflows/${workflowId}`);
};
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex">
    <!-- Left Sidebar - Navigation -->
    <aside class="w-64 bg-white border-r border-gray-200 flex flex-col">
      <div class="p-4 border-b border-gray-200">
        <h2 class="text-lg font-semibold text-gray-800">Workspace Settings</h2>
        <p class="text-sm text-gray-500">Configure workspace</p>
      </div>
      
      <nav class="flex-1 p-4 space-y-1">
        <button
          v-for="item in [
            { id: 'states', label: 'States', icon: '🔄', count: totalStates },
            { id: 'types', label: 'Work Item Types', icon: '📋', count: workItemTypes.length },
            { id: 'fields', label: 'Custom Fields', icon: '📝', count: customFields.length },
            { id: 'labels', label: 'Labels', icon: '🏷️', count: labels.length },
            { id: 'workflows', label: 'Workflows', icon: '⚙️', count: workflows.length },
            { id: 'automations', label: 'Automations', icon: '🤖', count: automations.length },
            { id: 'type-templates', label: 'Type Templates', icon: '📋', count: 0 },
            { id: 'templates', label: 'Templates', icon: '📄', count: templates.length },
          ]"
          :key="item.id"
          @click="activeSection = item.id"
          :class="[
            'w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm font-medium transition-colors',
            activeSection === item.id
              ? 'bg-blue-50 text-blue-700'
              : 'text-gray-600 hover:bg-gray-50 hover:text-gray-800'
          ]"
        >
          <span class="flex items-center space-x-3">
            <span>{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </span>
          <span class="text-xs bg-gray-100 px-2 py-0.5 rounded-full">{{ item.count }}</span>
        </button>
      </nav>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 overflow-auto">
      <!-- States Section -->
      <div v-if="activeSection === 'states'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-xl font-semibold text-gray-900">States</h1>
            <p class="text-sm text-gray-500 mt-1">Manage work item states grouped by workflow stage</p>
          </div>
        </div>

        <!-- State Groups -->
        <div class="space-y-6">
          <div 
            v-for="group in stateGroups" 
            :key="group.id"
            class="bg-white rounded-xl border border-gray-200 overflow-hidden"
          >
            <!-- Group Header -->
            <div class="px-4 py-3 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
              <div class="flex items-center space-x-3">
                <span class="text-sm font-semibold text-gray-700 uppercase tracking-wide">
                  {{ group.name }}
                </span>
                <span class="text-xs text-gray-400">{{ group.states.length }} states</span>
              </div>
              <button
                @click="handleAddState(group.id)"
                class="text-blue-600 hover:text-blue-700 text-sm font-medium"
              >
                + Add state
              </button>
            </div>

            <!-- States List -->
            <div class="divide-y divide-gray-100">
              <div
                v-for="state in group.states"
                :key="state.id"
                class="px-4 py-3 flex items-center justify-between hover:bg-gray-50 transition-colors cursor-pointer"
                @click="handleStateClick(group.id, state)"
              >
                <div class="flex items-center space-x-3">
                  <div
                    class="w-3 h-3 rounded-full"
                    :style="{ backgroundColor: state.color }"
                  ></div>
                  <span class="text-sm text-gray-800">{{ state.name }}</span>
                  <span
                    v-if="state.is_default"
                    class="px-2 py-0.5 bg-blue-100 text-blue-600 rounded text-xs font-medium"
                  >
                    Default
                  </span>
                </div>
                <div class="flex items-center space-x-2">
                  <button
                    @click.stop="handleStateClick(group.id, state)"
                    class="p-1 text-gray-400 hover:text-gray-600"
                  >
                    ✏️
                  </button>
                  <button
                    @click.stop="handleDeleteState(group.id, state.id)"
                    class="p-1 text-gray-400 hover:text-red-500"
                  >
                    🗑️
                  </button>
                </div>
              </div>

              <!-- Empty State -->
              <div
                v-if="group.states.length === 0"
                class="px-4 py-6 text-center text-gray-400 text-sm"
              >
                No states in this group. Click "Add state" to create one.
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Work Item Types Section -->
      <div v-if="activeSection === 'types'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-xl font-semibold text-gray-900">Work Item Types</h1>
            <p class="text-sm text-gray-500 mt-1">Define the types of work items available in your projects</p>
          </div>
          <button
            @click="handleAddType"
            class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
          >
            + Add type
          </button>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div
            v-for="type in workItemTypes"
            :key="type.id"
            @click="handleTypeClick(type)"
            class="bg-white rounded-xl border border-gray-200 p-4 hover:border-gray-300 hover:shadow-md transition-all cursor-pointer"
          >
            <div class="flex items-start justify-between">
              <div
                class="w-12 h-12 rounded-xl flex items-center justify-center text-2xl"
                :style="{ backgroundColor: type.color + '15' }"
              >
                {{ type.icon }}
              </div>
              <span
                :class="[
                  'px-2 py-1 rounded-full text-xs font-medium',
                  type.is_active 
                    ? 'bg-green-100 text-green-700' 
                    : 'bg-gray-100 text-gray-500'
                ]"
              >
                {{ type.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>
            <h3 class="mt-4 font-medium text-gray-900">{{ type.name }}</h3>
            <p class="mt-1 text-sm text-gray-500 line-clamp-2">{{ type.description }}</p>
          </div>
        </div>
      </div>

      <!-- Custom Fields Section -->
      <div v-if="activeSection === 'fields'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-xl font-semibold text-gray-900">Custom Fields</h1>
            <p class="text-sm text-gray-500 mt-1">Add custom properties to work items</p>
          </div>
          <button
            @click="handleAddField"
            class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
          >
            + Add field
          </button>
        </div>

        <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <table class="w-full">
            <thead class="bg-gray-50 border-b border-gray-200">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wide">Field</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wide">Type</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wide">Description</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wide">Required</th>
                <th class="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100">
              <tr
                v-for="field in customFields"
                :key="field.id"
                @click="handleFieldClick(field)"
                class="hover:bg-gray-50 cursor-pointer transition-colors"
              >
                <td class="px-4 py-3">
                  <span class="font-medium text-gray-900">{{ field.name }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="px-2 py-1 bg-gray-100 text-gray-600 rounded text-xs font-medium capitalize">
                    {{ field.type }}
                  </span>
                </td>
                <td class="px-4 py-3 text-sm text-gray-500">{{ field.description }}</td>
                <td class="px-4 py-3">
                  <span
                    :class="[
                      'px-2 py-1 rounded-full text-xs font-medium',
                      field.is_required 
                        ? 'bg-red-100 text-red-700' 
                        : 'bg-gray-100 text-gray-500'
                    ]"
                  >
                    {{ field.is_required ? 'Required' : 'Optional' }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right">
                  <button class="text-gray-400 hover:text-gray-600">✏️</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Labels Section -->
      <div v-if="activeSection === 'labels'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-xl font-semibold text-gray-900">Labels</h1>
            <p class="text-sm text-gray-500 mt-1">Categorize work items with color-coded labels</p>
          </div>
          <button
            @click="handleAddLabel"
            class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
          >
            + Add label
          </button>
        </div>

        <div class="flex flex-wrap gap-3">
          <div
            v-for="label in labels"
            :key="label.id"
            @click="handleLabelClick(label)"
            class="inline-flex items-center px-3 py-1.5 rounded-full cursor-pointer hover:opacity-80 transition-opacity"
            :style="{ backgroundColor: label.color + '20', borderColor: label.color }"
          >
            <div
              class="w-2 h-2 rounded-full mr-2"
              :style="{ backgroundColor: label.color }"
            ></div>
            <span class="text-sm font-medium" :style="{ color: label.color }">
              {{ label.name }}
            </span>
            <button
              @click.stop="handleDeleteLabel(label.id)"
              class="ml-2 text-gray-400 hover:text-red-500"
            >
              ✕
            </button>
          </div>
        </div>
      </div>

      <!-- Workflows Section -->
      <div v-if="activeSection === 'workflows'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-xl font-semibold text-gray-900">Workflows</h1>
            <p class="text-sm text-gray-500 mt-1">Manage workflow configurations</p>
          </div>
        </div>

        <div class="grid gap-4">
          <div
            v-for="workflow in workflows"
            :key="workflow.id"
            @click="handleViewWorkflow(workflow.id)"
            class="bg-white rounded-xl border border-gray-200 p-4 hover:border-gray-300 hover:shadow-md transition-all cursor-pointer"
          >
            <div class="flex items-center justify-between">
              <div>
                <h3 class="font-medium text-gray-900">{{ workflow.name }}</h3>
                <p class="mt-1 text-sm text-gray-500">
                  {{ workflow.states.length }} states: {{ workflow.states.join(' → ') }}
                </p>
              </div>
              <span class="text-gray-400">→</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Automations Section -->
      <div v-if="activeSection === 'automations'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-xl font-semibold text-gray-900">Automations</h1>
            <p class="text-sm text-gray-500 mt-1">Automate repetitive tasks with trigger-based workflows</p>
          </div>
          <button
            @click="handleAddAutomation"
            class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
          >
            + Create automation
          </button>
        </div>

        <div class="space-y-4">
          <div
            v-for="automation in automations"
            :key="automation.id"
            class="bg-white rounded-xl border border-gray-200 p-4"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-3">
                <div
                  :class="[
                    'w-10 h-10 rounded-lg flex items-center justify-center',
                    automation.is_enabled ? 'bg-green-100' : 'bg-gray-100'
                  ]"
                >
                  <span class="text-lg">🤖</span>
                </div>
                <div>
                  <h3 class="font-medium text-gray-900">{{ automation.name }}</h3>
                  <p class="text-sm text-gray-500">{{ automation.description }}</p>
                </div>
              </div>
              <div class="flex items-center space-x-3">
                <span
                  :class="[
                    'px-3 py-1 rounded-full text-xs font-medium',
                    automation.is_enabled 
                      ? 'bg-green-100 text-green-700' 
                      : 'bg-gray-100 text-gray-500'
                  ]"
                >
                  {{ automation.is_enabled ? 'Enabled' : 'Disabled' }}
                </span>
                <button class="px-3 py-1 border border-gray-300 rounded-lg text-sm hover:bg-gray-50">
                  {{ automation.is_enabled ? 'Disable' : 'Enable' }}
                </button>
                <button class="px-3 py-1 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
                  Edit
                </button>
              </div>
            </div>
            <div class="mt-4 pt-4 border-t border-gray-100">
              <div class="flex items-center space-x-2 text-sm">
                <span class="px-2 py-1 bg-blue-100 text-blue-700 rounded font-medium">
                  Trigger: {{ automation.trigger }}
                </span>
                <span class="text-gray-400">→</span>
                <span class="px-2 py-1 bg-amber-100 text-amber-700 rounded">
                  {{ automation.conditions.length }} conditions
                </span>
                <span class="text-gray-400">→</span>
                <span class="px-2 py-1 bg-green-100 text-green-700 rounded">
                  {{ automation.actions.length }} actions
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Templates Section -->
      <div v-if="activeSection === 'type-templates'" class="p-0">
        <TypeTemplateManager :workspace-id="1" />
      </div>

      <div v-if="activeSection === 'templates'" class="p-0">
        <ProjectTemplateManager :workspace-id="1" />
      </div>
    </main>

    <!-- State Modal -->
    <div
      v-if="showStateModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="showStateModal = false"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">
          {{ editingState ? 'Edit State' : 'Add State' }}
        </h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              v-model="newStateForm.name"
              type="text"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Color</label>
            <input
              v-model="newStateForm.color"
              type="color"
              class="w-full h-12 border border-gray-300 rounded-lg cursor-pointer"
            />
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showStateModal = false"
            class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            @click="handleSaveState"
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            {{ editingState ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Type Modal -->
    <div
      v-if="showTypeModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="showTypeModal = false"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">
          {{ editingType ? 'Edit Work Item Type' : 'Add Work Item Type' }}
        </h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              v-model="newTypeForm.name"
              type="text"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea
              v-model="newTypeForm.description"
              rows="3"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            ></textarea>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Icon</label>
              <input
                v-model="newTypeForm.icon"
                type="text"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Color</label>
              <input
                v-model="newTypeForm.color"
                type="color"
                class="w-full h-10 border border-gray-300 rounded-lg cursor-pointer"
              />
            </div>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showTypeModal = false"
            class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            @click="handleSaveType"
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            {{ editingType ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Field Modal -->
    <div
      v-if="showFieldModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="showFieldModal = false"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">
          {{ editingField ? 'Edit Custom Field' : 'Add Custom Field' }}
        </h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              v-model="newFieldForm.name"
              type="text"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <select
              v-model="newFieldForm.type"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="text">Text</option>
              <option value="number">Number</option>
              <option value="dropdown">Dropdown</option>
              <option value="date">Date</option>
              <option value="boolean">Checkbox</option>
              <option value="url">URL</option>
              <option value="email">Email</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea
              v-model="newFieldForm.description"
              rows="2"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            ></textarea>
          </div>
          <div v-if="newFieldForm.type === 'dropdown'" class="space-y-2">
            <label class="block text-sm font-medium text-gray-700">Options</label>
            <div
              v-for="(option, index) in newFieldForm.options"
              :key="index"
              class="flex items-center space-x-2"
            >
              <input
                v-model="newFieldForm.options[index]"
                type="text"
                class="flex-1 px-3 py-1 border border-gray-300 rounded"
              />
              <button
                @click="newFieldForm.options.splice(index, 1)"
                class="text-red-500 hover:text-red-700"
              >
                ✕
              </button>
            </div>
            <button
              @click="newFieldForm.options.push('')"
              class="text-blue-600 hover:text-blue-700 text-sm"
            >
              + Add option
            </button>
          </div>
          <div class="flex items-center">
            <input
              v-model="newFieldForm.is_required"
              type="checkbox"
              id="field_required"
              class="mr-2"
            />
            <label for="field_required" class="text-sm text-gray-700">Required field</label>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showFieldModal = false"
            class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            @click="handleSaveField"
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            {{ editingField ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Label Modal -->
    <div
      v-if="showLabelModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="showLabelModal = false"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">
          {{ editingLabel ? 'Edit Label' : 'Add Label' }}
        </h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              v-model="newLabelForm.name"
              type="text"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Color</label>
            <input
              v-model="newLabelForm.color"
              type="color"
              class="w-full h-12 border border-gray-300 rounded-lg cursor-pointer"
            />
          </div>
          <div class="mt-2">
            <div
              class="inline-flex items-center px-4 py-2 rounded-full"
              :style="{ backgroundColor: newLabelForm.color + '20', borderColor: newLabelForm.color }"
            >
              <div
                class="w-2 h-2 rounded-full mr-2"
                :style="{ backgroundColor: newLabelForm.color }"
              ></div>
              <span class="text-sm font-medium" :style="{ color: newLabelForm.color }">
                {{ newLabelForm.name || 'Label preview' }}
              </span>
            </div>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showLabelModal = false"
            class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            @click="handleSaveLabel"
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            {{ editingLabel ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Automation Modal -->
    <div
      v-if="showAutomationModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="showAutomationModal = false"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-lg">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">
          Create Automation
        </h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              type="text"
              placeholder="e.g., Auto-assign bugs to QA"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description (optional)</label>
            <input
              type="text"
              placeholder="What does this automation do?"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          
          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">Trigger</div>
            <select class="w-full px-4 py-2 border border-gray-300 rounded-lg">
              <option value="work_item_created">Work item created</option>
              <option value="work_item_updated">Work item updated</option>
              <option value="state_changed">State changed</option>
              <option value="assignee_changed">Assignee changed</option>
              <option value="scheduled">Scheduled</option>
            </select>
          </div>

          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">Conditions (optional)</div>
            <div class="text-sm text-gray-500 mb-2">No conditions set. Automation will run on all triggers.</div>
            <button class="text-blue-600 hover:text-blue-700 text-sm">+ Add condition</button>
          </div>

          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">Actions</div>
            <div class="text-sm text-gray-500 mb-2">No actions set.</div>
            <button class="text-blue-600 hover:text-blue-700 text-sm">+ Add action</button>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showAutomationModal = false"
            class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            @click="showAutomationModal = false"
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Create
          </button>
        </div>
      </div>
    </div>
  </div>
</template>