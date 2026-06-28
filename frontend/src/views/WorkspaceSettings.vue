<script setup lang="ts">import { ref, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import RelationTypeManager from '@/components/RelationTypeManager.vue';
import WorkspaceIssueTypeManager from '@/components/WorkspaceIssueTypeManager.vue';
import WorkspaceMemberList from '@/components/WorkspaceMemberList.vue';
import ProjectTemplateManager from '@/components/ProjectTemplateManager.vue';
import AISettingsPanel from '@/components/AISettingsPanel.vue';
import CustomFieldManager from '@/components/CustomFieldManager.vue';
import AutomationForm from '@/components/AutomationForm.vue';
import AutomationList from '@/components/AutomationList.vue';
import WorkspaceIntegrations from '@/components/WorkspaceIntegrations.vue'
import RoleManagement from '@/components/RoleManagement.vue';
import * as workflowApi from '@/api/workflow';
import * as issueTypeApi from '@/api/issue-type';
import * as customFieldApi from '@/api/custom-field';
import * as relationApi from '@/api/relation';
import { workspaceApi } from '@/api/workspace';
import { listProjects } from '@/api/project';
import { useConfirm } from '@/composables/useConfirm';

const { confirm } = useConfirm();

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
const relationTypes = ref<any[]>([]);
const memberCount = ref(0);

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
    const pid = firstProjectId.value;

    const results = await Promise.allSettled([
      issueTypeApi.getIssueTypes(wid),
      customFieldApi.listCustomFields(wid),
      pid ? workflowApi.listAutomations(pid) : Promise.resolve([]),
      relationApi.listRelationTypes(wid),
      slug.value ? workspaceApi.listMembers(slug.value) : Promise.resolve([]),
    ]);
    issueTypes.value = results[0].status === 'fulfilled' ? (Array.isArray(results[0].value) ? results[0].value : []) : [];
    customFields.value = results[1].status === 'fulfilled' ? (Array.isArray(results[1].value) ? results[1].value : []) : [];
    automations.value = results[2].status === 'fulfilled' ? (Array.isArray(results[2].value) ? results[2].value : []) : [];
    relationTypes.value = results[3].status === 'fulfilled' ? (Array.isArray(results[3].value) ? results[3].value : []) : [];
    memberCount.value = results[4].status === 'fulfilled' ? (Array.isArray(results[4].value) ? results[4].value.length : 0) : 0;
  } catch (e) { console.error('Failed to load data:', e); }
  finally { loading.value = false; }
}

// ===== Automation handlers (workspace-level) =====
const editingAutomation = ref<any>(null);
const showAutomationForm = ref(false);

function handleAddAutomation() {
  editingAutomation.value = null;
  showAutomationForm.value = true;
}
function handleEditAutomation(automation: any) {
  editingAutomation.value = automation;
  showAutomationForm.value = true;
}
async function handleSaveAutomation(data: any) {
  try {
    const pid = firstProjectId.value;
    if (!pid) return;
    if (editingAutomation.value) {
      await workflowApi.updateAutomation(pid, editingAutomation.value.id, data);
    } else {
      await workflowApi.createAutomation(pid, data);
    }
    showAutomationForm.value = false;
    await loadAllData();
  } catch (e) { console.error('Failed to save automation:', e); }
}
async function handleDeleteAutomation(automation: any) {
  if (!(await confirm({ title: '删除自动化', message: `确定要删除自动化规则 "${automation.name}" 吗？`, danger: true, confirmText: '删除' }))) return;
  try {
    const pid = firstProjectId.value;
    if (pid) await workflowApi.deleteAutomation(pid, automation.id);
    await loadAllData();
  } catch (e) { console.error('Failed to delete automation:', e); }
}
async function handleToggleAutomation(automation: any) {
  const newStatus = !automation.is_enabled;
  const actionText = newStatus ? '启用' : '停用';
  if (!(await confirm({ title: `${actionText}自动化规则`, message: `确定要${actionText}规则 "${automation.name}" 吗？`, danger: !newStatus, confirmText: actionText }))) return;
  try {
    const pid = firstProjectId.value;
    if (pid) {
      await workflowApi.updateAutomation(pid, automation.id, { is_enabled: newStatus });
      await loadAllData();
    }
  } catch (e) { console.error('Failed to toggle automation:', e); }
}

onMounted(() => { loadWorkspace(); });
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex">
    <!-- Left Sidebar -->
    <aside class="w-64 bg-white border-r border-gray-200 flex flex-col">
      <div class="p-4 border-b border-gray-200">
        <h2 class="text-lg font-semibold text-gray-800">Workspace Settings</h2>
        <p class="text-sm text-gray-500">Configure workspace-wide settings</p>
      </div>
      <nav class="flex-1 p-4 space-y-1">
        <button
          v-for="item in [
            { id: 'members', label: 'Members', icon: '👥', count: memberCount },
            { id: 'types', label: 'Work Item Types', icon: '📋', count: issueTypes.length },
            { id: 'templates', label: 'Templates', icon: '📦', count: 0 },
            { id: 'ai', label: 'AI', icon: '🤖', count: 0 },
            { id: 'fields', label: 'Custom Fields', icon: '📝', count: customFields.length },
            { id: 'automations', label: 'Automations', icon: '🤖', count: automations.length },
            { id: 'relations', label: 'Relations', icon: '🔗', count: relationTypes.length },
            { id: 'integrations', label: 'Integrations', icon: '🔌', count: 0 },
            { id: 'roles', label: 'Roles & Permissions', icon: '🔑', count: 0 },
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

      <!-- Members Section -->
      <div v-if="!loading && activeSection === 'members'" class="p-6">
        <WorkspaceMemberList :workspace-id="workspaceId" :workspace-slug="slug" />
      </div>

      <!-- Work Item Types Section -->
      <div v-if="!loading && activeSection === 'types'" class="p-6">
        <WorkspaceIssueTypeManager :workspace-id="workspaceId" />
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

      <!-- Automations Section (workspace-level) -->
      <div v-if="!loading && activeSection === 'automations'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-xl font-semibold text-gray-900">Automations</h1>
            <p class="text-sm text-gray-500 mt-1">Workspace-level automation rules</p>
          </div>
        </div>
        <AutomationList
          :automations="automations"
          @create="handleAddAutomation"
          @edit="handleEditAutomation"
          @delete="handleDeleteAutomation"
          @toggle="handleToggleAutomation"
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
    </main>

    <!-- Automation Form Modal -->
    <div v-if="showAutomationForm" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showAutomationForm = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-900">{{ editingAutomation ? 'Edit Automation' : 'Create Automation' }}</h3>
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
  </div>
</template>
