<script setup lang="ts">import { ref, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from '@/composables/useI18n';
import RelationTypeManager from '@/components/RelationTypeManager.vue';
import WorkspaceIssueTypeManager from '@/components/WorkspaceIssueTypeManager.vue';
import WorkspaceMemberList from '@/components/WorkspaceMemberList.vue';
import ProjectTemplateManager from '@/components/ProjectTemplateManager.vue';
import AISettingsPanel from '@/components/AISettingsPanel.vue';
import CustomFieldManager from '@/components/CustomFieldManager.vue';
import AutomationForm from '@/components/AutomationForm.vue';
import AutomationList from '@/components/AutomationList.vue';
import AutomationExecutionLog from '@/components/AutomationExecutionLog.vue';
import WorkspaceIntegrations from '@/components/WorkspaceIntegrations.vue'
import RoleManagement from '@/components/RoleManagement.vue';
import PluginManager from '@/views/PluginManager.vue';
import WorkflowManager from '@/components/WorkflowManager.vue';
import * as workflowApi from '@/api/workflow';
import * as issueTypeApi from '@/api/issue-type';
import api from '@/api';
import * as customFieldApi from '@/api/custom-field';
import * as relationApi from '@/api/relation';
import { workspaceApi } from '@/api/workspace';
import { listProjects } from '@/api/project';
import templateApi from '@/api/template';
import { pluginApi } from '@/api/plugin';
import { roleApi } from '@/api/role';
import { mcpApi } from '@/api/mcp';
import { githubApi } from '@/api/github';
import { slackApi } from '@/api/slack';
import { useConfirm } from '@/composables/useConfirm';
import { useToast } from '@/composables/useToast';

const { confirm } = useConfirm();
const { t } = useI18n();
const toast = useToast();

const route = useRoute();
const slug = computed(() => (route.params as any).slug as string || '');

const loading = ref(false);
const workspaceId = ref(0);
const firstProjectId = ref(0);
const activeSection = ref('types');

// ===== Data =====
const issueTypes = ref<any[]>([]);
const customFields = ref<any[]>([]);
const automations = ref<any[]>([]);
const workflows = ref<any[]>([]);
const relationTypes = ref<any[]>([]);
const memberCount = ref(0);
const templateCount = ref(0);
const pluginCount = ref(0);
const roleCount = ref(0);
const integrationCount = ref(0);
const workspaceStates = ref<any[]>([]);
const workspaceLabels = ref<any[]>([]);
const workspaceModules = ref<any[]>([]);

const navItems = computed(() => [
  { id: 'members', label: t('settings.members'), icon: '👥', count: memberCount.value },
  { id: 'types', label: t('settings.workItemTypes'), icon: '📋', count: issueTypes.value.length },
  { id: 'states', label: t('settings.states'), icon: '🔄', count: workspaceStates.value.length },
  { id: 'labels', label: t('settings.labels'), icon: '🏷️', count: workspaceLabels.value.length },
  { id: 'modules', label: t('settings.modules'), icon: '📦', count: workspaceModules.value.length },
  { id: 'templates', label: t('settings.templates'), icon: '📦', count: templateCount.value },
  { id: 'ai', label: t('settings.ai'), icon: '🤖', count: 0 },
  { id: 'fields', label: t('settings.fields'), icon: '📝', count: customFields.value.length },
  { id: 'workflows', label: t('settings.workflows'), icon: '🔄', count: workflows.value.length },
  { id: 'automations', label: t('settings.automations'), icon: '🤖', count: automations.value.length },
  { id: 'relations', label: t('settings.relations'), icon: '🔗', count: relationTypes.value.length },
  { id: 'integrations', label: t('settings.integrations'), icon: '🔌', count: integrationCount.value },
  { id: 'roles', label: t('settings.roles'), icon: '🔑', count: roleCount.value },
  { id: 'plugins', label: t('settings.plugins'), icon: '🧩', count: pluginCount.value },
])

// ===== Load workspace and data =====
async function loadWorkspace() {
  if (!slug.value) return;
  try {
    const ws = await workspaceApi.getBySlug(slug.value);
    workspaceId.value = ws.id;
    const projects = await listProjects(ws.id);
    firstProjectId.value = projects.length > 0 ? projects[0].id : 0;
    await loadAllData();
  } catch (e) { console.error('Failed to load workspace:', e); }
}

async function loadAllData() {
  loading.value = true;
  try {
    const wid = workspaceId.value;

    const results = await Promise.allSettled([
      issueTypeApi.getIssueTypes(wid),
      customFieldApi.listCustomFields(wid),
      workflowApi.listWorkspaceWorkflows(wid),
      workflowApi.listWorkspaceAutomations(wid),
      relationApi.listRelationTypes(wid),
      slug.value ? workspaceApi.listMembers(slug.value) : Promise.resolve([]),
      templateApi.listTemplates(wid),
      pluginApi.list(wid),
      roleApi.listRoles(wid),
      mcpApi.list(wid),
      githubApi.list(wid),
      slackApi.list(wid),
      api.get(`/workspaces/${wid}/settings/states`).then(r => r.data),
      api.get(`/workspaces/${wid}/settings/labels`).then(r => r.data),
      api.get(`/workspaces/${wid}/modules`).then(r => r.data),
    ]);
    issueTypes.value = results[0].status === 'fulfilled' ? (Array.isArray(results[0].value) ? results[0].value : []) : [];
    customFields.value = results[1].status === 'fulfilled' ? (Array.isArray(results[1].value) ? results[1].value : []) : [];
    workflows.value = results[2].status === 'fulfilled' ? (Array.isArray(results[2].value) ? results[2].value : []) : [];
    automations.value = results[3].status === 'fulfilled' ? (Array.isArray(results[3].value) ? results[3].value : []) : [];
    relationTypes.value = results[4].status === 'fulfilled' ? (Array.isArray(results[4].value) ? results[4].value : []) : [];
    memberCount.value = results[5].status === 'fulfilled' ? (Array.isArray(results[5].value) ? results[5].value.length : 0) : 0;
    templateCount.value = results[6].status === 'fulfilled' ? (Array.isArray(results[6].value) ? results[6].value.length : 0) : 0;
    pluginCount.value = results[7].status === 'fulfilled' ? (Array.isArray(results[7].value) ? results[7].value.length : 0) : 0;
    const roles = results[8].status === 'fulfilled' ? results[8].value : null;
    const roleData = roles?.data;
    roleCount.value = Array.isArray(roleData) ? roleData.length : (Array.isArray(roles) ? roles.length : 0);
    const mcp = results[9].status === 'fulfilled' ? (Array.isArray(results[9].value) ? results[9].value : []) : [];
    const github = results[10].status === 'fulfilled' ? (Array.isArray(results[10].value) ? results[10].value : []) : [];
    const slack = results[11].status === 'fulfilled' ? (Array.isArray(results[11].value) ? results[11].value : []) : [];
    integrationCount.value = mcp.length + github.length + slack.length;
    workspaceStates.value = results[12].status === 'fulfilled' ? (Array.isArray(results[12].value) ? results[12].value : []) : [];
    workspaceLabels.value = results[13].status === 'fulfilled' ? (Array.isArray(results[13].value) ? results[13].value : []) : [];
    workspaceModules.value = results[14].status === 'fulfilled' ? (Array.isArray(results[14].value) ? results[14].value : []) : [];
  } catch (e) { console.error('Failed to load data:', e); }
  finally { loading.value = false; }
}

// ===== Automation handlers (workspace-level) =====
const editingAutomation = ref<any>(null);
const showAutomationForm = ref(false);
const showAutomationLogModal = ref(false);
const viewingAutomationId = ref<number | null>(null);

function handleAddAutomation() {
  editingAutomation.value = null;
  showAutomationForm.value = true;
}
function handleEditAutomation(automation: any) {
  editingAutomation.value = automation;
  showAutomationForm.value = true;
}
function handleViewAutomationLog(automation: any) {
  viewingAutomationId.value = automation.id;
  showAutomationLogModal.value = true;
}
async function handleSaveAutomation(data: any) {
  try {
    const wid = workspaceId.value;
    if (!wid) return;
    if (editingAutomation.value) {
      await workflowApi.updateWorkspaceAutomation(wid, editingAutomation.value.id, data);
    } else {
      await workflowApi.createWorkspaceAutomation(wid, data);
    }
    showAutomationForm.value = false;
    await loadAllData();
  } catch (e) { console.error('Failed to save automation:', e); }
}
async function handleDeleteAutomation(automation: any) {
  if (!(await confirm({ title: t('settings.deleteAutomationGeneric'), message: t('settings.confirmDeleteAutomationGeneric', { name: automation.name }), danger: true, confirmText: t('common.delete') }))) return;
  try {
    const wid = workspaceId.value;
    if (wid) await workflowApi.deleteWorkspaceAutomation(wid, automation.id);
    await loadAllData();
  } catch (e) { console.error('Failed to delete automation:', e); }
}
async function handleToggleAutomation(automation: any) {
  const newStatus = !automation.is_enabled;
  const action = newStatus ? t('workflow.enable') : t('workflow.disable');
  if (!(await confirm({ title: newStatus ? t('settings.enableAutomation') : t('settings.disableAutomation'), message: t('settings.confirmToggleAutomation', { action, name: automation.name }), danger: !newStatus, confirmText: action }))) return;
  try {
    const wid = workspaceId.value;
    if (wid) {
      await workflowApi.updateWorkspaceAutomation(wid, automation.id, { is_enabled: newStatus });
      await loadAllData();
    }
  } catch (e) { console.error('Failed to toggle automation:', e); }
}

// ===== Workspace States handlers =====
const wsShowStateModal = ref(false);
const wsEditingState = ref<{ groupId: string; state: any } | null>(null);
const wsNewStateForm = ref({ name: '', color: '#3B82F6', groupId: '' });
const WS_STATE_GROUP_KEYS = ['backlog', 'unstarted', 'started', 'completed', 'cancelled'];

const wsStateGroups = computed(() => {
  const groupNames: Record<string, string> = {
    backlog: t('settings.stateGroupBacklogName'),
    unstarted: t('settings.stateGroupUnstartedName'),
    started: t('settings.stateGroupStartedName'),
    completed: t('settings.stateGroupCompletedName'),
    cancelled: t('settings.stateGroupCancelledName'),
  };
  return WS_STATE_GROUP_KEYS.map(id => ({
    id, name: groupNames[id],
    states: workspaceStates.value.filter((s: any) => s.group === id)
  }));
});

function wsHandleAddState(groupId: string) {
  wsNewStateForm.value = { name: '', color: '#3B82F6', groupId };
  wsEditingState.value = null;
  wsShowStateModal.value = true;
}
function wsHandleEditState(groupId: string, state: any) {
  wsEditingState.value = { groupId, state };
  wsNewStateForm.value = { name: state.name, color: state.color, groupId };
  wsShowStateModal.value = true;
}
async function wsHandleSaveState() {
  if (!wsNewStateForm.value.name || !workspaceId.value) return;
  try {
    const wid = workspaceId.value;
    if (wsEditingState.value?.state) {
      await api.put(`/workspaces/${wid}/settings/states/${wsEditingState.value.state.id}`, {
        name: wsNewStateForm.value.name, color: wsNewStateForm.value.color, group: wsNewStateForm.value.groupId
      });
    } else {
      await api.post(`/workspaces/${wid}/settings/states`, {
        name: wsNewStateForm.value.name, color: wsNewStateForm.value.color, group: wsNewStateForm.value.groupId
      });
    }
    wsShowStateModal.value = false;
    await loadAllData();
  } catch (e: any) { console.error('Failed to save state:', e); toast.error(e?.response?.data?.message || 'Failed to save state'); }
}
async function wsHandleDeleteState(_groupId: string, state: any) {
  if (!workspaceId.value) return;
  if (!(await confirm({ title: t('settings.deleteState'), message: t('settings.confirmDeleteState', { 0: state.name }), danger: true, confirmText: t('common.delete') }))) return;
  try { await api.delete(`/workspaces/${workspaceId.value}/settings/states/${state.id}`); await loadAllData(); }
  catch (e: any) { console.error('Failed to delete state:', e); toast.error(e?.response?.data?.message || 'Failed to delete state'); }
}

// ===== Workspace Labels handlers =====
const wsShowLabelModal = ref(false);
const wsEditingLabel = ref<any>(null);
const wsNewLabelForm = ref({ name: '', color: '#3B82F6' });

function wsHandleAddLabel() {
  wsEditingLabel.value = null;
  wsNewLabelForm.value = { name: '', color: '#3B82F6' };
  wsShowLabelModal.value = true;
}
function wsHandleEditLabel(label: any) {
  wsEditingLabel.value = { ...label };
  wsNewLabelForm.value = { name: label.name, color: label.color };
  wsShowLabelModal.value = true;
}
async function wsHandleSaveLabel() {
  if (!wsNewLabelForm.value.name || !workspaceId.value) return;
  try {
    const wid = workspaceId.value;
    if (wsEditingLabel.value) {
      await api.put(`/workspaces/${wid}/settings/labels/${wsEditingLabel.value.id}`, {
        name: wsNewLabelForm.value.name, color: wsNewLabelForm.value.color
      });
    } else {
      await api.post(`/workspaces/${wid}/settings/labels`, {
        name: wsNewLabelForm.value.name, color: wsNewLabelForm.value.color
      });
    }
    wsShowLabelModal.value = false;
    await loadAllData();
  } catch (e: any) { console.error('Failed to save label:', e); toast.error(e?.response?.data?.message || 'Failed to save label'); }
}
async function wsHandleDeleteLabel(label: any) {
  if (!workspaceId.value) return;
  if (!(await confirm({ title: t('settings.deleteLabel'), message: t('settings.confirmDeleteLabel', { 0: label.name }), danger: true, confirmText: t('common.delete') }))) return;
  try { await api.delete(`/workspaces/${workspaceId.value}/settings/labels/${label.id}`); await loadAllData(); }
  catch (e: any) { console.error('Failed to delete label:', e); toast.error(e?.response?.data?.message || 'Failed to delete label'); }
}

// ===== Workspace Modules handlers =====
const wsShowModuleModal = ref(false);
const wsEditingModule = ref<any>(null);
const wsNewModuleForm = ref({ name: '', description: '', parentId: null as number | null });

function wsHandleAddModule() {
  wsEditingModule.value = null;
  wsNewModuleForm.value = { name: '', description: '', parentId: null };
  wsShowModuleModal.value = true;
}
function wsHandleEditModule(module: any) {
  wsEditingModule.value = { ...module };
  wsNewModuleForm.value = { name: module.name, description: module.description || '', parentId: module.parent_id || null };
  wsShowModuleModal.value = true;
}
async function wsHandleSaveModule() {
  if (!wsNewModuleForm.value.name || !workspaceId.value) return;
  try {
    const wid = workspaceId.value;
    if (wsEditingModule.value) {
      await api.put(`/workspaces/${wid}/modules/${wsEditingModule.value.id}`, {
        name: wsNewModuleForm.value.name, description: wsNewModuleForm.value.description, parent_id: wsNewModuleForm.value.parentId
      });
    } else {
      await api.post(`/workspaces/${wid}/modules`, {
        name: wsNewModuleForm.value.name, description: wsNewModuleForm.value.description, parent_id: wsNewModuleForm.value.parentId
      });
    }
    wsShowModuleModal.value = false;
    await loadAllData();
  } catch (e: any) { console.error('Failed to save module:', e); toast.error(e?.response?.data?.message || 'Failed to save module'); }
}
async function wsHandleDeleteModule(module: any) {
  if (!workspaceId.value) return;
  if (!(await confirm({ title: t('settings.delete'), message: t('settings.confirmDelete', { 0: module.name }), danger: true, confirmText: t('common.delete') }))) return;
  try { await api.delete(`/workspaces/${workspaceId.value}/modules/${module.id}`); await loadAllData(); }
  catch (e: any) { console.error('Failed to delete module:', e); toast.error(e?.response?.data?.message || 'Failed to delete module'); }
}

onMounted(() => { loadWorkspace(); });
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex">
    <!-- Left Sidebar -->
    <aside class="w-64 bg-white border-r border-gray-200 flex flex-col">
      <div class="p-4 border-b border-gray-200">
        <h2 class="text-lg font-semibold text-gray-800">{{ t('workspace.settings') }}</h2>
        <p class="text-sm text-gray-500">{{ t('workspace.settingsDesc', 'Configure workspace-wide settings') }}</p>
      </div>
      <nav class="flex-1 p-4 space-y-1">
        <button
          v-for="item in navItems"
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

      <!-- Members Section -->
      <div v-if="!loading && activeSection === 'members'" class="p-6">
        <WorkspaceMemberList :workspace-id="workspaceId" :workspace-slug="slug" />
      </div>

      <!-- Work Item Types Section -->
      <div v-if="!loading && activeSection === 'types'" class="p-6">
        <WorkspaceIssueTypeManager :workspace-id="workspaceId" />
      </div>

      <!-- Workspace States Section -->
      <div v-if="!loading && activeSection === 'states'" class="p-6 space-y-6">
        <div class="flex items-center justify-between">
          <div><h2 class="text-lg font-semibold text-gray-900">{{ t('settings.workItemStates') }}</h2><p class="text-sm text-gray-500 mt-1">{{ t('settings.workspaceStatesDesc') }}</p></div>
        </div>
        <div v-for="group in wsStateGroups" :key="group.id" class="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div class="px-4 py-3 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <span class="text-sm font-semibold text-gray-700 uppercase tracking-wide">{{ group.name }}</span>
              <span class="text-xs bg-gray-200 px-2 py-0.5 rounded-full">{{ group.states.length }}</span>
            </div>
            <button @click="wsHandleAddState(group.id)" class="text-indigo-600 hover:text-indigo-700 text-sm font-medium">+ {{ t('settings.addState') }}</button>
          </div>
          <div class="divide-y divide-gray-100">
            <div v-for="state in group.states" :key="state.id" class="px-4 py-3 flex items-center justify-between hover:bg-gray-50 transition-colors cursor-pointer" @click="wsHandleEditState(group.id, state)">
              <div class="flex items-center space-x-3">
                <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: state.color }"></div>
                <span class="text-sm text-gray-800">{{ state.name }}</span>
                <span v-if="state.is_default" class="px-2 py-0.5 bg-indigo-100 text-indigo-600 rounded text-xs font-medium">{{ t('settings.default') }}</span>
              </div>
              <div class="flex items-center space-x-2">
                <button @click.stop="wsHandleEditState(group.id, state)" class="p-1 text-gray-400 hover:text-gray-600">✏️</button>
                <button @click.stop="wsHandleDeleteState(group.id, state)" class="p-1 text-gray-400 hover:text-red-500">🗑️</button>
              </div>
            </div>
            <div v-if="group.states.length === 0" class="px-4 py-6 text-center text-gray-400 text-sm">{{ t('settings.noStates') }}</div>
          </div>
        </div>
      </div>

      <!-- Workspace Labels Section -->
      <div v-if="!loading && activeSection === 'labels'" class="p-6 space-y-6">
        <div class="flex items-center justify-between">
          <div><h2 class="text-lg font-semibold text-gray-900">{{ t('settings.labels') }}</h2><p class="text-sm text-gray-500 mt-1">{{ t('settings.workspaceLabelsDesc') }}</p></div>
          <button @click="wsHandleAddLabel" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">+ {{ t('settings.addLabel') }}</button>
        </div>
        <div class="bg-white rounded-xl border border-gray-200 p-6">
          <div class="flex flex-wrap gap-3">
            <div v-for="label in workspaceLabels" :key="label.id" @click="wsHandleEditLabel(label)" class="inline-flex items-center px-3 py-1.5 rounded-full cursor-pointer hover:opacity-80 transition-opacity" :style="{ backgroundColor: label.color + '20', borderColor: label.color }">
              <div class="w-2 h-2 rounded-full mr-2" :style="{ backgroundColor: label.color }"></div>
              <span class="text-sm font-medium" :style="{ color: label.color }">{{ label.name }}</span>
              <button @click.stop="wsHandleDeleteLabel(label)" class="ml-2 text-gray-400 hover:text-red-500">✕</button>
            </div>
            <div v-if="workspaceLabels.length === 0" class="w-full text-center text-gray-400 py-8">{{ t('settings.noLabels') }}</div>
          </div>
        </div>
      </div>

      <!-- Workspace Modules Section -->
      <div v-if="!loading && activeSection === 'modules'" class="p-6 space-y-6">
        <div class="flex items-center justify-between">
          <div><h2 class="text-lg font-semibold text-gray-900">{{ t('settings.modules') }}</h2><p class="text-sm text-gray-500 mt-1">{{ t('settings.workspaceModulesDesc') }}</p></div>
          <button @click="wsHandleAddModule" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">+ {{ t('settings.addModule') }}</button>
        </div>
        <div class="bg-white rounded-xl border border-gray-200">
          <div class="divide-y divide-gray-100">
            <div v-for="module in workspaceModules" :key="module.id" class="px-4 py-3 flex items-center justify-between hover:bg-gray-50 transition-colors">
              <div>
                <span class="text-sm text-gray-800 font-medium">{{ module.name }}</span>
                <span v-if="module.description" class="text-xs text-gray-500 ml-2">{{ module.description }}</span>
              </div>
              <div class="flex items-center space-x-2">
                <button @click.stop="wsHandleEditModule(module)" class="p-1 text-gray-400 hover:text-gray-600">✏️</button>
                <button @click.stop="wsHandleDeleteModule(module)" class="p-1 text-gray-400 hover:text-red-500">🗑️</button>
              </div>
            </div>
            <div v-if="workspaceModules.length === 0" class="px-4 py-8 text-center text-gray-400 text-sm">{{ t('settings.noModules') }}</div>
          </div>
        </div>
      </div>

      <!-- Templates Section -->
      <div v-if="!loading && activeSection === 'templates'" class="p-6">
        <ProjectTemplateManager :workspace-id="workspaceId" />
      </div>

      <!-- AI Settings Section -->
      <div v-if="!loading && activeSection === 'ai'" class="p-6">
        <AISettingsPanel :workspace-id="workspaceId" />
      </div>

      <!-- Custom Fields Section -->
      <div v-if="!loading && activeSection === 'fields'" class="p-6">
        <CustomFieldManager :workspace-id="workspaceId" />
      </div>

      <!-- Workflows Section (workspace-level) -->
      <div v-if="!loading && activeSection === 'workflows'" class="p-6">
        <WorkflowManager :workspace-id="workspaceId" />
      </div>

      <!-- Automations Section (workspace-level) -->
      <div v-if="!loading && activeSection === 'automations'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-xl font-semibold text-gray-900">{{ t('settings.automations') }}</h1>
            <p class="text-sm text-gray-500 mt-1">{{ t('settings.workspaceAutomationsDesc') }}</p>
          </div>
        </div>
        <AutomationList
          :automations="automations"
          @create="handleAddAutomation"
          @edit="handleEditAutomation"
          @delete="handleDeleteAutomation"
          @toggle="handleToggleAutomation"
          @viewHistory="handleViewAutomationLog"
        />
      </div>

      <!-- Relations Section -->
      <div v-if="activeSection === 'relations'" class="p-0">
        <RelationTypeManager :workspace-id="workspaceId" />
      </div>

      <!-- Integrations Section -->
      <div v-if="activeSection === 'integrations'" class="p-0">
        <WorkspaceIntegrations :workspace-id="workspaceId" :slug="slug" />
      </div>

      <!-- Roles & Permissions Section -->
      <div v-if="activeSection === 'roles'" class="p-0">
        <RoleManagement />
      </div>

      <!-- Plugins Section -->
      <div v-if="activeSection === 'plugins'" class="p-0">
        <PluginManager :workspace-id="workspaceId" />
      </div>
    </main>

    <!-- Automation Form Modal -->
    <div v-if="showAutomationForm" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showAutomationForm = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-900">{{ editingAutomation ? t('settings.editAutomation') : t('settings.createAutomation') }}</h3>
          <button @click="showAutomationForm = false" class="text-gray-400 hover:text-gray-600">✕</button>
        </div>
        <div class="flex-1 overflow-y-auto">
          <AutomationForm
            :project-id="firstProjectId"
            :workspace-id="workspaceId"
            :automation="editingAutomation"
            @submit="handleSaveAutomation"
            @cancel="showAutomationForm = false"
          />
        </div>
      </div>
    </div>

    <!-- Automation Execution Log Drawer -->
    <AutomationExecutionLog
      :visible="showAutomationLogModal"
      :rule-id="viewingAutomationId || undefined"
      @close="showAutomationLogModal = false"
    />

    <!-- Workspace State Modal -->
    <div v-if="wsShowStateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="wsShowStateModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ wsEditingState ? t('settings.editState') : t('settings.addState') }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.name') }}</label><input v-model="wsNewStateForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.color') }}</label><input v-model="wsNewStateForm.color" type="color" class="w-full h-12 border border-gray-300 rounded-lg cursor-pointer" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.stateGroup') }}</label>
            <select v-model="wsNewStateForm.groupId" class="w-full px-4 py-2 border border-gray-300 rounded-lg">
              <option v-for="g in wsStateGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="wsShowStateModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('settings.cancel') }}</button>
          <button @click="wsHandleSaveState" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ wsEditingState ? t('settings.update') : t('settings.create') }}</button>
        </div>
      </div>
    </div>

    <!-- Workspace Label Modal -->
    <div v-if="wsShowLabelModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="wsShowLabelModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ wsEditingLabel ? t('settings.editLabel') : t('settings.addLabel') }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.name') }}</label><input v-model="wsNewLabelForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.color') }}</label><input v-model="wsNewLabelForm.color" type="color" class="w-full h-12 border border-gray-300 rounded-lg cursor-pointer" /></div>
          <div class="mt-2"><div class="inline-flex items-center px-4 py-2 rounded-full" :style="{ backgroundColor: wsNewLabelForm.color + '20' }"><div class="w-2 h-2 rounded-full mr-2" :style="{ backgroundColor: wsNewLabelForm.color }"></div><span class="text-sm font-medium" :style="{ color: wsNewLabelForm.color }">{{ wsNewLabelForm.name || t('settings.preview') }}</span></div></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="wsShowLabelModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('settings.cancel') }}</button>
          <button @click="wsHandleSaveLabel" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ wsEditingLabel ? t('settings.update') : t('settings.create') }}</button>
        </div>
      </div>
    </div>

    <!-- Workspace Module Modal -->
    <div v-if="wsShowModuleModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="wsShowModuleModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ wsEditingModule ? t('settings.editModule') : t('settings.addModule') }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.name') }}</label><input v-model="wsNewModuleForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.descriptionOptional') }}</label><textarea v-model="wsNewModuleForm.description" rows="2" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"></textarea></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.parentModule') }}</label>
            <select v-model="wsNewModuleForm.parentId" class="w-full px-4 py-2 border border-gray-300 rounded-lg">
              <option :value="null">{{ t('settings.none') }}</option>
              <option v-for="m in workspaceModules.filter((mod: any) => mod.id !== wsEditingModule?.id)" :key="m.id" :value="m.id">{{ m.name }}</option>
            </select>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="wsShowModuleModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('settings.cancel') }}</button>
          <button @click="wsHandleSaveModule" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ wsEditingModule ? t('settings.update') : t('settings.create') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
