<script setup lang="ts">import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import api from '@/api';
import * as workflowApi from '@/api/workflow';

const route = useRoute();
const router = useRouter();

const workflowId = ref(0);
const projectId = ref(0);

const workflowName = ref('');
const workflowDescription = ref('');
const isActive = ref(true);
const states = ref<any[]>([]);
const transitions = ref<any[]>([]);

const showAddStateModal = ref(false);
const showAddTransitionModal = ref(false);
const selectedFromState = ref<number | null>(null);
const newStateName = ref('');
const newStateColor = ref('#3B82F6');
const newStateGroup = ref('backlog');
const newTransitionTo = ref<number | null>(null);
const newTransitionType = ref('allow');
const newTransitionDesc = ref('');

const loading = ref(true);

function getStateName(id: number): string {
  const s = states.value.find((s: any) => s.id === id);
  return s?.name || `#${id}`;
}
function getStateById(id: number) { return states.value.find((s: any) => s.id === id); }
function getTransitionsFromState(stateId: number) { return transitions.value.filter((t: any) => t.from_state_id === stateId); }

async function loadData() {
  loading.value = true;
  try {
    // Get workflowId from route
    const wfId = parseInt((route.params as any).workflowId, 10);
    if (!wfId) return;
    workflowId.value = wfId;

    // Find the workflow's project by listing all workflows across projects
    // For simplicity, try the default project (ID 1)
    const pid = 1;
    projectId.value = pid;

    // Load workflow
    const wf = await api.get(`/projects/${pid}/workflows/${wfId}`).then(r => r.data);
    workflowName.value = wf.name || '';
    workflowDescription.value = wf.description || '';
    isActive.value = wf.is_active ?? true;
    transitions.value = wf.transitions || [];

    // Load states for the project
    const sts = await api.get(`/projects/${pid}/settings/states`).then(r => r.data);
    states.value = Array.isArray(sts) ? sts : (sts?.data || []);
  } catch (e) {
    console.error('Failed to load workflow:', e);
  } finally {
    loading.value = false;
  }
}

// ===== State handlers =====
async function handleSaveState() {
  if (!newStateName.value.trim() || !projectId.value) return;
  try {
    await api.post(`/projects/${projectId.value}/settings/states?workspace_id=1`, {
      name: newStateName.value, color: newStateColor.value, group: newStateGroup.value
    });
    showAddStateModal.value = false;
    newStateName.value = ''; newStateColor.value = '#3B82F6';
    // Reload states
    const sts = await api.get(`/projects/${projectId.value}/settings/states`).then(r => r.data);
    states.value = Array.isArray(sts) ? sts : (sts?.data || []);
  } catch (e) { console.error('Failed to add state:', e); }
}

// ===== Transition handlers =====
function handleAddTransition(stateId: number) {
  selectedFromState.value = stateId;
  newTransitionTo.value = null;
  newTransitionType.value = 'allow';
  newTransitionDesc.value = '';
  showAddTransitionModal.value = true;
}

async function handleSaveTransition() {
  if (!selectedFromState.value || !newTransitionTo.value || !workflowId.value) return;
  try {
    await workflowApi.addTransition(projectId.value, workflowId.value, {
      from_state_id: selectedFromState.value,
      to_state_id: newTransitionTo.value,
      description: newTransitionDesc.value,
      rule_type: newTransitionType.value
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

function goBack() { router.back(); }

onMounted(loadData);
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <div v-if="loading" class="flex items-center justify-center h-64"><div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div></div>

    <template v-if="!loading">
      <!-- Header -->
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
            <span :class="['px-3 py-1 rounded-full text-sm font-medium', isActive ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600']">{{ isActive ? 'Active' : 'Inactive' }}</span>
          </div>
        </div>
      </header>

      <!-- Content -->
      <main class="p-6">
        <div class="max-w-5xl mx-auto">
          <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h2 class="font-semibold text-gray-800">Define Workflow</h2>
              <button @click="showAddStateModal = true" class="text-blue-600 hover:text-blue-700 text-sm font-medium">+ Add States</button>
            </div>

            <div class="p-6">
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

                  <!-- Transitions from this state -->
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

          <!-- Visualize -->
          <div class="mt-6 bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div class="px-6 py-4 border-b border-gray-200"><h2 class="font-semibold text-gray-800">Visualize Workflow</h2></div>
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

    <!-- Add State Modal -->
    <div v-if="showAddStateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showAddStateModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <div class="flex items-center justify-between mb-6"><h3 class="text-lg font-semibold text-gray-800">Add State</h3><button @click="showAddStateModal = false" class="text-gray-400 hover:text-gray-600">✕</button></div>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Name</label><input v-model="newStateName" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent" placeholder="Enter state name" /></div>
          <div class="grid grid-cols-2 gap-4">
            <div><label class="block text-sm font-medium text-gray-700 mb-1">Color</label><input v-model="newStateColor" type="color" class="w-full h-10 border border-gray-300 rounded-lg cursor-pointer" /></div>
            <div><label class="block text-sm font-medium text-gray-700 mb-1">Group</label>
              <select v-model="newStateGroup" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent">
                <option value="backlog">Backlog</option><option value="unstarted">Unstarted</option><option value="started">Started</option><option value="completed">Completed</option><option value="cancelled">Cancelled</option>
              </select>
            </div>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showAddStateModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="handleSaveState" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700">Add State</button>
        </div>
      </div>
    </div>

    <!-- Add Transition Modal -->
    <div v-if="showAddTransitionModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showAddTransitionModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <div class="flex items-center justify-between mb-6"><h3 class="text-lg font-semibold text-gray-800">Add Transition</h3><button @click="showAddTransitionModal = false" class="text-gray-400 hover:text-gray-600">✕</button></div>
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
              <option v-for="s in states.filter(s => s.id !== selectedFromState)" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <input v-model="newTransitionDesc" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg" placeholder="Optional description" />
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