<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import api from '@/api';
import * as workflowApi from '@/api/workflow';
import { useI18n } from '@/composables/useI18n';
import { useConfirm } from '@/composables/useConfirm';

const { t } = useI18n();
const { confirm } = useConfirm();

const route = useRoute();
const router = useRouter();

const workflowId = ref(0);
const projectId = ref(0);

const workflowName = ref('');
const workflowDescription = ref('');
const isActive = ref(true);
const updatingStatus = ref(false);
const states = ref<any[]>([]);
const transitions = ref<any[]>([]);
const members = ref<any[]>([]);

const showAddTransitionModal = ref(false);
const selectedFromState = ref<number | null>(null);
const newTransitionTo = ref<number | null>(null);
const newTransitionType = ref('allow');
const newTransitionName = ref('');
const newTransitionDesc = ref('');
const newTransitionApproverIds = ref<number[]>([]);

const loading = ref(true);

function getStateName(id: number): string {
  const s = states.value.find((s: any) => s.id === id);
  return s?.name || `#${id}`;
}
function getStateById(id: number) { return states.value.find((s: any) => s.id === id); }
function getTransitionsFromState(stateId: number) { return transitions.value.filter((t: any) => t.from_state_id === stateId); }

async function loadMembers() {
  if (!projectId.value) return;
  try {
    const res = await api.get(`/projects/${projectId.value}/members`).then(r => r.data);
    members.value = Array.isArray(res) ? res : (res?.data || []);
  } catch (e) { console.error('Failed to load members:', e); }
}

async function loadData() {
  loading.value = true;
  try {
    const wfId = parseInt((route.params as any).workflowId, 10);
    if (!wfId) return;
    workflowId.value = wfId;

    const pid = parseInt((route.params as any).id, 10);
    projectId.value = pid;

    const wf = await api.get(`/projects/${pid}/workflows/${wfId}`).then(r => r.data);
    workflowName.value = wf.name || '';
    workflowDescription.value = wf.description || '';
    isActive.value = wf.is_active ?? true;
    transitions.value = wf.transitions || [];

    const sts = await api.get(`/projects/${pid}/settings/states`).then(r => r.data);
    states.value = Array.isArray(sts) ? sts : (sts?.data || []);

    await loadMembers();
  } catch (e) {
    console.error('Failed to load workflow:', e);
  } finally {
    loading.value = false;
  }
}

function handleAddTransition(stateId: number) {
  selectedFromState.value = stateId;
  newTransitionTo.value = null;
  newTransitionType.value = 'allow';
  newTransitionName.value = '';
  newTransitionDesc.value = '';
  newTransitionApproverIds.value = [];
  showAddTransitionModal.value = true;
}

function getApproverNames(approverIds: string | null | undefined): string {
  if (!approverIds || !members.value.length) return '';
  try {
    const ids = JSON.parse(approverIds);
    if (!Array.isArray(ids)) return '';
    return ids.map((id: number) => {
      const m = members.value.find(m => m.user_id === id || m.id === id);
      return m?.user?.display_name || m?.display_name || `#${id}`;
    }).join(', ');
  } catch { return ''; }
}

async function handleSaveTransition() {
  if (!selectedFromState.value || !newTransitionTo.value || !workflowId.value) return;
  try {
    const fromState = getStateById(selectedFromState.value);
    const toState = getStateById(newTransitionTo.value);
    const name = newTransitionName.value || `${fromState?.name || ''}→${toState?.name || ''}`;
    
    let approverIds: string | undefined = undefined;
    if (newTransitionType.value === 'approval' && newTransitionApproverIds.value.length > 0) {
      approverIds = JSON.stringify(newTransitionApproverIds.value);
    }

    await workflowApi.addTransition(projectId.value, workflowId.value, {
      from_state_id: selectedFromState.value,
      to_state_id: newTransitionTo.value,
      name: name,
      description: newTransitionDesc.value,
      rule_type: newTransitionType.value,
      approver_ids: approverIds
    });
    showAddTransitionModal.value = false;
    await loadData();
  } catch (e) { console.error('Failed to add transition:', e); }
}

async function handleDeleteTransition(transitionId: number) {
  if (!workflowId.value) return;
  try {
    await workflowApi.deleteTransition(projectId.value, workflowId.value, transitionId);
    await loadData();
  } catch (e) { console.error('Failed to delete transition:', e); }
}

const availableToStates = computed(() => {
  return states.value.filter(s => s.id !== selectedFromState.value);
});

function toggleApprover(userId: number) {
  const idx = newTransitionApproverIds.value.indexOf(userId);
  if (idx > -1) {
    newTransitionApproverIds.value.splice(idx, 1);
  } else {
    newTransitionApproverIds.value.push(userId);
  }
}

function goBack() {
  const slug = (route.params as any).slug as string;
  const pid = (route.params as any).id as string;
  router.push(`/workspace/${slug}/project/${pid}/settings`);
}

async function toggleActive() {
  if (!workflowId.value || updatingStatus.value) return;

  const newStatus = !isActive.value;

  if (!(await confirm({
    title: newStatus ? t('workflow.enableWorkflow') : t('workflow.disableWorkflow'),
    message: newStatus ? t('workflow.confirmEnable', { name: workflowName.value }) : t('workflow.confirmDisable', { name: workflowName.value }),
    danger: !newStatus,
    confirmText: newStatus ? t('workflow.enable') : t('workflow.disable')
  }))) return;
  
  updatingStatus.value = true;
  try {
    await workflowApi.updateWorkflow(projectId.value, workflowId.value, {
      is_active: newStatus
    });
    isActive.value = newStatus;
  } catch (e) { console.error('Failed to toggle workflow status:', e); }
  finally { updatingStatus.value = false; }
}

onMounted(loadData);
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <div v-if="loading" class="flex items-center justify-center h-64"><div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div></div>

    <template v-if="!loading">
      <header class="bg-white border-b border-gray-200 px-6 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4">
            <button @click="goBack" class="text-gray-500 hover:text-gray-700">← Back</button>
            <div>
              <h1 class="text-xl font-semibold text-gray-800">{{ workflowName }}</h1>
              <p class="text-sm text-gray-500">{{ workflowDescription }}</p>
            </div>
          </div>
          <div class="flex items-center space-x-3">
            <button @click="toggleActive" :disabled="updatingStatus" :class="['px-3 py-1 rounded-full text-sm font-medium transition-colors', isActive ? 'bg-green-100 text-green-700 hover:bg-green-200 cursor-pointer' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 cursor-pointer', updatingStatus ? 'opacity-50 cursor-not-allowed' : '']">
              {{ isActive ? 'Active' : 'Inactive' }}
            </button>
          </div>
        </div>
      </header>

      <main class="p-6">
        <div class="max-w-5xl mx-auto">
          <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <div>
                <h2 class="font-semibold text-gray-800">{{ t('workflow.define') }}</h2>
                <p class="text-xs text-gray-500 mt-1">{{ t('settings.states') }} 在项目设置中管理</p>
              </div>
            </div>

            <div class="p-6">
              <div v-if="states.length === 0" class="text-center py-8 text-gray-400">
                <p>No states defined. Add states in project settings first.</p>
              </div>
              <div class="space-y-4">
                <div v-for="state in states" :key="state.id" class="border border-gray-200 rounded-lg overflow-hidden">
                  <div class="flex items-center justify-between p-4 bg-gray-50">
                    <div class="flex items-center space-x-4">
                      <div class="w-4 h-4 rounded-full" :style="{ backgroundColor: state.color }"></div>
                      <div>
                        <div class="flex items-center space-x-2">
                          <span class="font-medium text-gray-800">{{ state.name }}</span>
                          <span v-if="state.is_default" class="px-2 py-0.5 bg-blue-100 text-blue-600 rounded text-xs font-medium">Default</span>
                        </div>
                        <span class="text-xs text-gray-500 uppercase">{{ state.group }}</span>
                      </div>
                    </div>
                  </div>

                  <div class="p-4 border-t border-gray-200">
                    <div class="flex items-center justify-between mb-3">
                      <span class="text-sm text-gray-500">Transitions</span>
                      <button @click="handleAddTransition(state.id)" class="text-blue-600 hover:text-blue-700 text-sm">+ Add transition</button>
                    </div>
                    <div class="space-y-2">
                      <div v-for="t in getTransitionsFromState(state.id)" :key="t.id" class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                        <div class="flex items-center space-x-3">
                          <span :class="['px-2 py-1 rounded text-xs font-medium', t.rule_type === 'approval' ? 'bg-amber-100 text-amber-700' : 'bg-blue-100 text-blue-700']">{{ t.rule_type || 'allow' }}</span>
                          <span class="text-gray-700">→</span>
                          <span class="font-medium text-gray-800">{{ getStateName(t.to_state_id) }}</span>
                          <span v-if="t.description" class="text-xs text-gray-400">{{ t.description }}</span>
                          <span v-if="t.rule_type === 'approval' && t.approver_ids" class="text-xs text-gray-500 ml-2">👤 {{ getApproverNames(t.approver_ids) }}</span>
                        </div>
                        <button @click="handleDeleteTransition(t.id)" class="text-gray-400 hover:text-red-500">✕</button>
                      </div>
                      <div v-if="getTransitionsFromState(state.id).length === 0" class="text-center py-4 text-gray-400 text-sm">No transitions defined.</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="mt-6 bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div class="px-6 py-4 border-b border-gray-200"><h2 class="font-semibold text-gray-800">{{ t('workflow.visualize') }}</h2></div>
            <div class="p-6">
              <div class="flex items-center justify-center space-x-4 py-8 flex-wrap gap-4">
                <div v-for="state in states" :key="state.id" class="flex flex-col items-center">
                  <div class="w-20 h-20 rounded-xl flex flex-col items-center justify-center shadow-md border-2" :style="{ backgroundColor: state.color + '20', borderColor: state.color }">
                    <div class="w-4 h-4 rounded-full mb-1" :style="{ backgroundColor: state.color }"></div>
                    <span class="text-xs font-medium text-gray-700 text-center">{{ state.name }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </template>

    <div v-if="showAddTransitionModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showAddTransitionModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-6"><h3 class="text-lg font-semibold text-gray-800">{{ t('workflow.addTransition') }}</h3><button @click="showAddTransitionModal = false" class="text-gray-400 hover:text-gray-600">✕</button></div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">From</label>
            <div class="flex items-center space-x-2 p-3 bg-gray-50 rounded-lg">
              <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: getStateById(selectedFromState!)?.color }"></div>
              <span class="font-medium text-gray-800">{{ getStateName(selectedFromState!) }}</span>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <div class="flex space-x-2">
              <button @click="newTransitionType = 'allow'" :class="['flex-1 px-4 py-2 rounded-lg text-sm font-medium', newTransitionType === 'allow' ? 'bg-blue-600 text-white' : 'border border-gray-300 hover:bg-gray-50']">Transition</button>
              <button @click="newTransitionType = 'approval'" :class="['flex-1 px-4 py-2 rounded-lg text-sm font-medium', newTransitionType === 'approval' ? 'bg-amber-600 text-white' : 'border border-gray-300 hover:bg-gray-50']">Approval</button>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">To</label>
            <select v-model="newTransitionTo" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent">
              <option :value="null">Select destination state</option>
              <option v-for="s in availableToStates" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name (optional)</label>
            <input v-model="newTransitionName" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg" placeholder="Auto-generated if empty" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <input v-model="newTransitionDesc" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg" placeholder="Optional description" />
          </div>
          <div v-if="newTransitionType === 'approval'" class="p-4 bg-amber-50 rounded-lg border border-amber-200">
            <label class="block text-sm font-medium text-amber-800 mb-2">Approvers</label>
            <p class="text-xs text-amber-600 mb-3">Select users who can approve this transition</p>
            <div class="space-y-2 max-h-48 overflow-y-auto">
              <label v-for="m in members" :key="m.user_id || m.id" class="flex items-center space-x-3 p-2 rounded hover:bg-amber-100 cursor-pointer">
                <input type="checkbox" :checked="newTransitionApproverIds.includes(m.user_id || m.id)" @change="toggleApprover(m.user_id || m.id)" class="rounded border-amber-300 text-amber-600 focus:ring-amber-500" />
                <div class="flex items-center space-x-2">
                  <div class="w-7 h-7 rounded-full bg-gray-200 flex items-center justify-center text-xs font-medium text-gray-600">
                    {{ (m.user?.display_name || m.display_name || 'U').charAt(0).toUpperCase() }}
                  </div>
                  <span class="text-sm text-gray-700">{{ m.user?.display_name || m.display_name || 'Unknown' }}</span>
                </div>
              </label>
              <div v-if="members.length === 0" class="text-center py-4 text-gray-400 text-sm">No members found.</div>
            </div>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showAddTransitionModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="handleSaveTransition" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700">Add Transition</button>
        </div>
      </div>
    </div>
  </div>
</template>
