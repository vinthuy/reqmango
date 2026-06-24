<script setup lang="ts">import { ref, computed, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import TypeTemplateManager from '@/components/TypeTemplateManager.vue';
import ProjectTemplateManager from '@/components/ProjectTemplateManager.vue';
import RelationTypeManager from '@/components/RelationTypeManager.vue';
import WorkspaceIssueTypeManager from '@/components/WorkspaceIssueTypeManager.vue';
import CustomFieldManager from '@/components/CustomFieldManager.vue';
import api from '@/api';
import { workspaceApi } from '@/api/workspace';
import { listProjects } from '@/api/project';
import * as issueTypeApi from '@/api/issue-type';
import * as customFieldApi from '@/api/custom-field';
import * as workflowApi from '@/api/workflow';
import * as templateApi from '@/api/template';
import * as relationApi from '@/api/relation';
import { useConfirm } from '@/composables/useConfirm';

const { confirm } = useConfirm();

const router = useRouter();
const route = useRoute();
const slug = computed(() => (route.params as any).slug as string || '');

const loading = ref(false);
const workspaceId = ref(0);
const defaultProjectId = ref(0);
const activeSection = ref('states');

// ===== Data (loaded from API) =====
const states = ref<any[]>([]);
const issueTypes = ref<any[]>([]);
const customFields = ref<any[]>([]);
const labels = ref<any[]>([]);
const workflows = ref<any[]>([]);
const automations = ref<any[]>([]);
const typeTemplates = ref<any[]>([]);
const relationTypes = ref<any[]>([]);

const stateGroups = computed(() => {
  const groups: Record<string, any[]> = {};
  for (const s of states.value) {
    const g = s.group || 'backlog';
    if (!groups[g]) groups[g] = [];
    groups[g].push(s);
  }
  const names: Record<string, string> = { backlog: 'Backlog', unstarted: 'Unstarted', started: 'Started', completed: 'Completed', cancelled: 'Cancelled' };
  return Object.entries(groups).map(([id, sts]) => ({ id, name: names[id] || id, states: sts }));
});

const totalStates = computed(() => states.value.length);

// ===== Load workspace and data =====
async function loadWorkspace() {
  if (!slug.value) return;
  try {
    const ws = await workspaceApi.getBySlug(slug.value);
    workspaceId.value = ws.id;
    const projects = await listProjects(ws.id);
    if (projects.length > 0) {
      defaultProjectId.value = projects[0].id;
      await loadAllData();
    }
  } catch (e) { console.error('Failed to load workspace:', e); }
}

async function loadAllData() {
  loading.value = true;
  try {
    const pid = defaultProjectId.value;
    const wid = workspaceId.value;
    const results = await Promise.allSettled([
      api.get(`/projects/${pid}/settings/states`).then(r => r.data),
      issueTypeApi.getIssueTypes(wid, pid),
      customFieldApi.listCustomFields(wid),
      api.get(`/projects/${pid}/settings/labels`).then(r => r.data),
      workflowApi.listWorkflows(pid),
      workflowApi.listAutomations(pid),
      templateApi.listTypeTemplates(wid),
      relationApi.listRelationTypes(wid),
    ]);
    states.value = results[0].status === 'fulfilled' ? (Array.isArray(results[0].value) ? results[0].value : []) : [];
    issueTypes.value = results[1].status === 'fulfilled' ? (Array.isArray(results[1].value) ? results[1].value : []) : [];
    customFields.value = results[2].status === 'fulfilled' ? (Array.isArray(results[2].value) ? results[2].value : []) : [];
    labels.value = results[3].status === 'fulfilled' ? (Array.isArray(results[3].value) ? results[3].value : []) : [];
    workflows.value = results[4].status === 'fulfilled' ? (Array.isArray(results[4].value) ? results[4].value : []) : [];
    automations.value = results[5].status === 'fulfilled' ? (Array.isArray(results[5].value) ? results[5].value : []) : [];
    typeTemplates.value = results[6].status === 'fulfilled' ? (Array.isArray(results[6].value) ? results[6].value : []) : [];
    relationTypes.value = results[7].status === 'fulfilled' ? (Array.isArray(results[7].value) ? results[7].value : []) : [];
  } catch (e) { console.error('Failed to load data:', e); }
  finally { loading.value = false; }
}

// ===== Modal state =====
const showStateModal = ref(false);
const showLabelModal = ref(false);
const showAutomationModal = ref(false);
const editingState = ref<{ groupId: string; state: any } | null>(null);
const editingLabel = ref<any>(null);

const newStateForm = ref({ name: '', color: '#3B82F6', groupId: '' });
const newLabelForm = ref({ name: '', color: '#3B82F6' });
const newAutomationForm = ref({ name: '', description: '', trigger: 'issue_created', conditions: '[]', actions: '[]' });

// ===== State handlers =====
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
const handleSaveState = async () => {
  if (!newStateForm.value.name || !defaultProjectId.value) return;
  try {
    if (editingState.value) {
      await api.put(`/projects/${defaultProjectId.value}/settings/states/${editingState.value.state.id}`, {
        name: newStateForm.value.name, color: newStateForm.value.color
      });
    } else {
      await api.post(`/projects/${defaultProjectId.value}/settings/states?workspace_id=${workspaceId.value}`, {
        name: newStateForm.value.name, color: newStateForm.value.color, group: newStateForm.value.groupId
      });
    }
    showStateModal.value = false;
    await loadAllData();
  } catch (e) { console.error('Failed to save state:', e); }
};
const handleDeleteState = async (_groupId: string, state: any) => {
  if (!defaultProjectId.value) return;
  if (!(await confirm({ title: '删除状态', message: `确定要删除状态 "${state.name}" 吗？此操作不可恢复。`, danger: true, confirmText: '删除' }))) return;
  try { await api.delete(`/projects/${defaultProjectId.value}/settings/states/${state.id}`); await loadAllData(); }
  catch (e) { console.error('Failed to delete state:', e); }
};

// ===== Label handlers =====
const handleLabelClick = (label: any) => {
  editingLabel.value = { ...label };
  newLabelForm.value = { name: label.name, color: label.color };
  showLabelModal.value = true;
};
const handleAddLabel = () => {
  editingLabel.value = null;
  newLabelForm.value = { name: '', color: '#3B82F6' };
  showLabelModal.value = true;
};
const handleSaveLabel = async () => {
  if (!newLabelForm.value.name || !defaultProjectId.value) return;
  try {
    if (editingLabel.value) {
      await api.put(`/projects/${defaultProjectId.value}/settings/labels/${editingLabel.value.id}`, { name: newLabelForm.value.name, color: newLabelForm.value.color });
    } else {
      await api.post(`/projects/${defaultProjectId.value}/settings/labels?workspace_id=${workspaceId.value}`, { name: newLabelForm.value.name, color: newLabelForm.value.color });
    }
    showLabelModal.value = false;
    await loadAllData();
  } catch (e) { console.error('Failed to save label:', e); }
};
const handleDeleteLabel = async (label: any) => {
  if (!defaultProjectId.value) return;
  if (!(await confirm({ title: '删除标签', message: `确定要删除标签 "${label.name}" 吗？此操作不可恢复。`, danger: true, confirmText: '删除' }))) return;
  try { await api.delete(`/projects/${defaultProjectId.value}/settings/labels/${label.id}`); await loadAllData(); }
  catch (e) { console.error('Failed to delete label:', e); }
};

// ===== Workflow / Automation navigation =====
const handleViewWorkflow = (workflowId: number) => {
  router.push(`/workspace/${slug.value}/settings/workflows/${workflowId}`);
};
const handleAddAutomation = () => {
  newAutomationForm.value = { name: '', description: '', trigger: 'issue_created', conditions: '[]', actions: '[]' };
  showAutomationModal.value = true;
};
const handleSaveAutomation = async () => {
  if (!newAutomationForm.value.name || !defaultProjectId.value) return;
  try {
    let conds = '[]', acts = '[]';
    try { conds = JSON.stringify(JSON.parse(newAutomationForm.value.conditions || '[]')); } catch { conds = '[]'; }
    try { acts = JSON.stringify(JSON.parse(newAutomationForm.value.actions || '[]')); } catch { acts = '[]'; }
    await workflowApi.createAutomation(defaultProjectId.value, { name: newAutomationForm.value.name, description: newAutomationForm.value.description, trigger_type: newAutomationForm.value.trigger, conditions: conds, actions: acts });
    showAutomationModal.value = false;
    await loadAllData();
  } catch (e) { console.error('Failed to create automation:', e); }
};

onMounted(() => { loadWorkspace(); });
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex">
    <!-- Left Sidebar -->
    <aside class="w-64 bg-white border-r border-gray-200 flex flex-col">
      <div class="p-4 border-b border-gray-200">
        <h2 class="text-lg font-semibold text-gray-800">Workspace Settings</h2>
        <p class="text-sm text-gray-500">Configure workspace</p>
      </div>
      <nav class="flex-1 p-4 space-y-1">
        <button
          v-for="item in [
            { id: 'states', label: 'States', icon: '🔄', count: totalStates },
            { id: 'types', label: 'Work Item Types', icon: '📋', count: issueTypes.length },
            { id: 'fields', label: 'Custom Fields', icon: '📝', count: customFields.length },
            { id: 'labels', label: 'Labels', icon: '🏷️', count: labels.length },
            { id: 'workflows', label: 'Workflows', icon: '⚙️', count: workflows.length },
            { id: 'automations', label: 'Automations', icon: '🤖', count: automations.length },
            { id: 'type-templates', label: 'Type Templates', icon: '📋', count: typeTemplates.length },
            { id: 'relations', label: 'Relations', icon: '🔗', count: relationTypes.length },
          ]"
          :key="item.id"
          @click="activeSection = item.id"
          :class="['w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm font-medium transition-colors', activeSection === item.id ? 'bg-blue-50 text-blue-700' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-800']"
        >
          <span class="flex items-center space-x-3">
            <span>{{ item.icon }}</span><span>{{ item.label }}</span>
          </span>
          <span class="text-xs bg-gray-100 px-2 py-0.5 rounded-full">{{ item.count }}</span>
        </button>
      </nav>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 overflow-auto">
      <div v-if="loading" class="flex items-center justify-center h-64"><div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div></div>

      <!-- States Section -->
      <div v-if="!loading && activeSection === 'states'" class="p-6">
        <div class="flex items-center justify-between mb-6"><div><h1 class="text-xl font-semibold text-gray-900">States</h1><p class="text-sm text-gray-500 mt-1">Manage work item states grouped by workflow stage</p></div></div>
        <div class="space-y-6">
          <div v-for="group in stateGroups" :key="group.id" class="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div class="px-4 py-3 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
              <div class="flex items-center space-x-3"><span class="text-sm font-semibold text-gray-700 uppercase tracking-wide">{{ group.name }}</span><span class="text-xs text-gray-400">{{ group.states.length }} states</span></div>
              <button @click="handleAddState(group.id)" class="text-blue-600 hover:text-blue-700 text-sm font-medium">+ Add state</button>
            </div>
            <div class="divide-y divide-gray-100">
              <div v-for="state in group.states" :key="state.id" class="px-4 py-3 flex items-center justify-between hover:bg-gray-50 transition-colors cursor-pointer" @click="handleStateClick(group.id, state)">
                <div class="flex items-center space-x-3"><div class="w-3 h-3 rounded-full" :style="{ backgroundColor: state.color }"></div><span class="text-sm text-gray-800">{{ state.name }}</span><span v-if="state.is_default" class="px-2 py-0.5 bg-blue-100 text-blue-600 rounded text-xs font-medium">Default</span></div>
                <div class="flex items-center space-x-2">
                  <button @click.stop="handleStateClick(group.id, state)" class="p-1 text-gray-400 hover:text-gray-600">✏️</button>
                  <button @click.stop="handleDeleteState(group.id, state)" class="p-1 text-gray-400 hover:text-red-500">🗑️</button>
                </div>
              </div>
              <div v-if="group.states.length === 0" class="px-4 py-6 text-center text-gray-400 text-sm">No states in this group.</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Types Section (workspace-level, inline management) -->
      <div v-if="!loading && activeSection === 'types'" class="p-6">
        <WorkspaceIssueTypeManager :workspace-id="workspaceId" ref="typeManagerRef" />
      </div>

      <!-- Custom Fields Section (workspace-level) -->
      <div v-if="!loading && activeSection === 'fields'" class="p-6">
        <CustomFieldManager :workspace-id="workspaceId" />
      </div>

      <!-- Labels Section -->
      <div v-if="!loading && activeSection === 'labels'" class="p-6">
        <div class="flex items-center justify-between mb-6"><div><h1 class="text-xl font-semibold text-gray-900">Labels</h1><p class="text-sm text-gray-500 mt-1">Categorize work items with color-coded labels</p></div><button @click="handleAddLabel" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">+ Add label</button></div>
        <div class="flex flex-wrap gap-3">
          <div v-for="label in labels" :key="label.id" @click="handleLabelClick(label)" class="inline-flex items-center px-3 py-1.5 rounded-full cursor-pointer hover:opacity-80 transition-opacity" :style="{ backgroundColor: label.color + '20', borderColor: label.color }">
            <div class="w-2 h-2 rounded-full mr-2" :style="{ backgroundColor: label.color }"></div>
            <span class="text-sm font-medium" :style="{ color: label.color }">{{ label.name }}</span>
            <button @click.stop="handleDeleteLabel(label)" class="ml-2 text-gray-400 hover:text-red-500">✕</button>
          </div>
          <div v-if="labels.length === 0" class="w-full text-center text-gray-400 py-8">No labels. Click "Add label" to create.</div>
        </div>
      </div>

      <!-- Workflows Section -->
      <div v-if="!loading && activeSection === 'workflows'" class="p-6">
        <div class="flex items-center justify-between mb-6"><div><h1 class="text-xl font-semibold text-gray-900">Workflows</h1><p class="text-sm text-gray-500 mt-1">Manage workflow configurations</p></div></div>
        <div class="grid gap-4">
          <div v-for="workflow in workflows" :key="workflow.id" @click="handleViewWorkflow(workflow.id)" class="bg-white rounded-xl border border-gray-200 p-4 hover:border-gray-300 hover:shadow-md transition-all cursor-pointer">
            <div class="flex items-center justify-between">
              <div><h3 class="font-medium text-gray-900">{{ workflow.name }}</h3><p class="mt-1 text-sm text-gray-500">{{ (workflow.transitions || []).length }} transitions</p></div>
              <span class="text-gray-400">→</span>
            </div>
          </div>
          <div v-if="workflows.length === 0" class="text-center text-gray-400 py-12">No workflows configured. Create one from a project.</div>
        </div>
      </div>

      <!-- Automations Section -->
      <div v-if="!loading && activeSection === 'automations'" class="p-6">
        <div class="flex items-center justify-between mb-6"><div><h1 class="text-xl font-semibold text-gray-900">Automations</h1><p class="text-sm text-gray-500 mt-1">Automate repetitive tasks with trigger-based rules</p></div><button @click="handleAddAutomation" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">+ Create automation</button></div>
        <div class="space-y-4">
          <div v-for="automation in automations" :key="automation.id" class="bg-white rounded-xl border border-gray-200 p-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-3">
                <div :class="['w-10 h-10 rounded-lg flex items-center justify-center', automation.is_enabled ? 'bg-green-100' : 'bg-gray-100']"><span class="text-lg">🤖</span></div>
                <div><h3 class="font-medium text-gray-900">{{ automation.name }}</h3><p class="text-sm text-gray-500">{{ automation.description }}</p></div>
              </div>
              <div class="flex items-center space-x-3">
                <span :class="['px-3 py-1 rounded-full text-xs font-medium', automation.is_enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500']">{{ automation.is_enabled ? 'Enabled' : 'Disabled' }}</span>
              </div>
            </div>
            <div class="mt-4 pt-4 border-t border-gray-100">
              <div class="flex items-center space-x-2 text-sm">
                <span class="px-2 py-1 bg-blue-100 text-blue-700 rounded font-medium">Trigger: {{ automation.trigger_type }}</span>
              </div>
            </div>
          </div>
          <div v-if="automations.length === 0" class="text-center text-gray-400 py-12">No automations. Click "Create automation" to add.</div>
        </div>
      </div>

      <!-- Embedded components for workspace-scoped sections -->
      <div v-if="activeSection === 'relations'" class="p-0"><RelationTypeManager :workspace-id="workspaceId" /></div>
      <div v-if="activeSection === 'type-templates'" class="p-0"><TypeTemplateManager :workspace-id="workspaceId" /></div>
      <div v-if="activeSection === 'templates'" class="p-0"><ProjectTemplateManager :workspace-id="workspaceId" /></div>
    </main>

    <!-- State Modal -->
    <div v-if="showStateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showStateModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ editingState ? 'Edit State' : 'Add State' }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Name</label><input v-model="newStateForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Color</label><input v-model="newStateForm.color" type="color" class="w-full h-12 border border-gray-300 rounded-lg cursor-pointer" /></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showStateModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="handleSaveState" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">{{ editingState ? 'Update' : 'Create' }}</button>
        </div>
      </div>
    </div>

    <!-- Label Modal -->
    <div v-if="showLabelModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showLabelModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ editingLabel ? 'Edit Label' : 'Add Label' }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Name</label><input v-model="newLabelForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Color</label><input v-model="newLabelForm.color" type="color" class="w-full h-12 border border-gray-300 rounded-lg cursor-pointer" /></div>
          <div class="mt-2"><div class="inline-flex items-center px-4 py-2 rounded-full" :style="{ backgroundColor: newLabelForm.color + '20' }"><div class="w-2 h-2 rounded-full mr-2" :style="{ backgroundColor: newLabelForm.color }"></div><span class="text-sm font-medium" :style="{ color: newLabelForm.color }">{{ newLabelForm.name || 'Preview' }}</span></div></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showLabelModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="handleSaveLabel" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">{{ editingLabel ? 'Update' : 'Create' }}</button>
        </div>
      </div>
    </div>

    <!-- Automation Modal -->
    <div v-if="showAutomationModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showAutomationModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-lg">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">Create Automation</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Name</label><input v-model="newAutomationForm.name" type="text" placeholder="e.g., Auto-assign bugs to QA" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Description (optional)</label><input v-model="newAutomationForm.description" type="text" placeholder="What does this automation do?" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent" /></div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">Trigger</div>
            <select v-model="newAutomationForm.trigger" class="w-full px-4 py-2 border border-gray-300 rounded-lg">
              <option value="issue_created">Issue Created</option><option value="issue_updated">Issue Updated</option><option value="state_changed">State Changed</option><option value="assignee_changed">Assignee Changed</option><option value="comment_added">Comment Added</option>
            </select>
          </div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">Conditions (JSON)</div>
            <textarea v-model="newAutomationForm.conditions" rows="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-xs font-mono" placeholder='[{"field":"priority","operator":"equals","value":"urgent"}]'></textarea>
          </div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">Actions (JSON)</div>
            <textarea v-model="newAutomationForm.actions" rows="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-xs font-mono" placeholder='[{"type":"assign","field":"assignee","value":"1"}]'></textarea>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showAutomationModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="handleSaveAutomation" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">Create</button>
        </div>
      </div>
    </div>
  </div>
</template>