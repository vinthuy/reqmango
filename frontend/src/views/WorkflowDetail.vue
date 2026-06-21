<script setup lang="ts">import { ref } from 'vue';
import type { State, WorkflowTransition } from '@/types';
// 模拟数据
const workflowName = ref('Default Workflow');
const workflowDescription = ref('Default workflow for all work items');
const isActive = ref(true);
const states = ref<State[]>([
 { id: 1, name: 'Backlog', category: 'unstarted', color: '#9CA3AF', allowNew: true, isDefault: true },
 { id: 2, name: 'In Progress', category: 'started', color: '#3B82F6', allowNew: false, isDefault: false },
 { id: 3, name: 'Review', category: 'started', color: '#F59E0B', allowNew: false, isDefault: false },
 { id: 4, name: 'Done', category: 'completed', color: '#22C55E', allowNew: false, isDefault: false },
]);
const transitions = ref<WorkflowTransition[]>([
 { id: 1, fromStateId: 1, toStateId: 2, type: 'transition', allowedRoles: ['all'] },
 { id: 2, fromStateId: 2, toStateId: 1, type: 'transition', allowedRoles: ['all'] },
 { id: 3, fromStateId: 2, toStateId: 3, type: 'transition', allowedRoles: ['all'] },
 { id: 4, fromStateId: 3, toStateId: 2, type: 'transition', allowedRoles: ['all'] },
 { id: 5, fromStateId: 3, toStateId: 4, type: 'approval', allowedRoles: ['reviewers'], approvers: ['admin'] },
 { id: 6, fromStateId: 4, toStateId: 2, type: 'transition', allowedRoles: ['all'] },
]);
const showAddStateModal = ref(false);
const showAddTransitionModal = ref(false);
const selectedFromState = ref<number | null>(null);
const newStateName = ref('');
const newStateColor = ref('#3B82F6');
const newStateCategory = ref<'unstarted' | 'started' | 'completed'>('unstarted');
const newTransitionTo = ref<number | null>(null);
const newTransitionType = ref<'transition' | 'approval'>('transition');
const getStateById = (id: number) => {
 return states.value.find(s => s.id === id);
};
const getTransitionsFromState = (stateId: number) => {
 return transitions.value.filter(t => t.fromStateId === stateId);
};
const handleAddState = () => {
 showAddStateModal.value = true;
};
const handleSaveState = () => {
 if (newStateName.value.trim()) {
 states.value.push({
 id: Date.now(),
 name: newStateName.value,
 category: newStateCategory.value,
 color: newStateColor.value,
 allowNew: false,
 isDefault: false
 });
 newStateName.value = '';
 newStateColor.value = '#3B82F6';
 newStateCategory.value = 'unstarted';
 }
 showAddStateModal.value = false;
};
const handleAddTransition = (stateId: number) => {
 selectedFromState.value = stateId;
 newTransitionTo.value = null;
 newTransitionType.value = 'transition';
 showAddTransitionModal.value = true;
};
const handleSaveTransition = () => {
 if (selectedFromState.value && newTransitionTo.value) {
 transitions.value.push({
 id: Date.now(),
 fromStateId: selectedFromState.value,
 toStateId: newTransitionTo.value,
 type: newTransitionType.value,
 allowedRoles: ['all'],
 ...(newTransitionType.value === 'approval' && { approvers: ['admin'] })
 });
 }
 showAddTransitionModal.value = false;
};
const handleDeleteTransition = (transitionId: number) => {
 transitions.value = transitions.value.filter(t => t.id !== transitionId);
};
const toggleAllowNew = (stateId: number) => {
 const state = states.value.find(s => s.id === stateId);
 if (state) {
 state.allowNew = !state.allowNew;
 }
};
const toggleDefault = (stateId: number) => {
 states.value.forEach(s => {
 s.isDefault = s.id === stateId;
 });
};
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <button class="text-gray-500 hover:text-gray-700">
            ← Back
          </button>
          <div>
            <h1 class="text-xl font-semibold text-gray-800">{{ workflowName }}</h1>
            <p class="text-sm text-gray-500">{{ workflowDescription }}</p>
          </div>
        </div>
        <div class="flex items-center space-x-3">
          <span
            :class="[
              'px-3 py-1 rounded-full text-sm font-medium',
              isActive 
                ? 'bg-green-100 text-green-700' 
                : 'bg-gray-100 text-gray-600'
            ]"
          >
            {{ isActive ? 'Active' : 'Inactive' }}
          </span>
          <button class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 text-sm">
            Edit Workflow
          </button>
        </div>
      </div>
    </header>

    <!-- Content -->
    <main class="p-6">
      <div class="max-w-5xl mx-auto">
        <!-- States Section -->
        <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
            <h2 class="font-semibold text-gray-800">Define Workflow</h2>
            <button
              @click="handleAddState"
              class="text-blue-600 hover:text-blue-700 text-sm font-medium"
            >
              + Add States
            </button>
          </div>

          <div class="p-6">
            <!-- States Grid -->
            <div class="space-y-4">
              <div
                v-for="state in states"
                :key="state.id"
                class="border border-gray-200 rounded-lg overflow-hidden"
              >
                <div class="flex items-center justify-between p-4 bg-gray-50">
                  <div class="flex items-center space-x-4">
                    <div
                      class="w-4 h-4 rounded-full"
                      :style="{ backgroundColor: state.color }"
                    ></div>
                    <div>
                      <div class="flex items-center space-x-2">
                        <span class="font-medium text-gray-800">{{ state.name }}</span>
                        <span
                          v-if="state.isDefault"
                          class="px-2 py-0.5 bg-blue-100 text-blue-600 rounded text-xs font-medium"
                        >
                          Default
                        </span>
                      </div>
                      <span class="text-xs text-gray-500 uppercase">{{ state.category }}</span>
                    </div>
                  </div>
                  <div class="flex items-center space-x-4">
                    <label class="flex items-center space-x-2 cursor-pointer">
                      <input
                        type="checkbox"
                        :checked="state.allowNew"
                        @change="toggleAllowNew(state.id)"
                        class="rounded border-gray-300 text-blue-600"
                      />
                      <span class="text-sm text-gray-600">Allow new work items</span>
                    </label>
                    <button
                      @click="toggleDefault(state.id)"
                      :class="[
                        'px-2 py-1 rounded text-xs font-medium transition-colors',
                        state.isDefault 
                          ? 'bg-blue-100 text-blue-700' 
                          : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                      ]"
                    >
                      {{ state.isDefault ? 'Default' : 'Set as Default' }}
                    </button>
                  </div>
                </div>

                <!-- Transitions -->
                <div class="p-4 border-t border-gray-200">
                  <div class="flex items-center justify-between mb-3">
                    <span class="text-sm text-gray-500">Flows</span>
                    <button
                      @click="handleAddTransition(state.id)"
                      class="text-blue-600 hover:text-blue-700 text-sm"
                    >
                      + Add flow
                    </button>
                  </div>
                  
                  <div class="space-y-2">
                    <div
                      v-for="transition in getTransitionsFromState(state.id)"
                      :key="transition.id"
                      class="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                    >
                      <div class="flex items-center space-x-3">
                        <span
                          :class="[
                            'px-2 py-1 rounded text-xs font-medium',
                            transition.type === 'transition' 
                              ? 'bg-blue-100 text-blue-700' 
                              : 'bg-amber-100 text-amber-700'
                          ]"
                        >
                          {{ transition.type === 'transition' ? 'Transition' : 'Approval' }}
                        </span>
                        <span class="text-gray-700">→</span>
                        <span class="font-medium text-gray-800">
                          {{ getStateById(transition.toStateId)?.name }}
                        </span>
                        <span class="text-sm text-gray-500">
                          by {{ transition.allowedRoles.includes('all') ? 'All' : 'Specific roles' }}
                        </span>
                      </div>
                      <button
                        @click="handleDeleteTransition(transition.id)"
                        class="text-gray-400 hover:text-red-500"
                      >
                        ✕
                      </button>
                    </div>

                    <div
                      v-if="getTransitionsFromState(state.id).length === 0"
                      class="text-center py-4 text-gray-400 text-sm"
                    >
                      No flows defined. Click "Add flow" to define transitions.
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Workflow Visualization -->
        <div class="mt-6 bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div class="px-6 py-4 border-b border-gray-200">
            <h2 class="font-semibold text-gray-800">Visualize Workflow</h2>
          </div>
          <div class="p-6">
            <div class="flex items-center justify-center space-x-4 py-8">
              <div
                v-for="state in states"
                :key="state.id"
                class="flex flex-col items-center"
              >
                <div
                  class="w-20 h-20 rounded-xl flex flex-col items-center justify-center shadow-md"
                  :style="{ backgroundColor: state.color + '20', borderColor: state.color, borderWidth: '2px' }"
                >
                  <div
                    class="w-4 h-4 rounded-full mb-1"
                    :style="{ backgroundColor: state.color }"
                  ></div>
                  <span class="text-xs font-medium text-gray-700 text-center">{{ state.name }}</span>
                </div>
                <div
                  v-if="state.id < states.length"
                  class="flex items-center justify-center w-8 h-8"
                >
                  <span class="text-gray-400">→</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Add State Modal -->
    <div
      v-if="showAddStateModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="showAddStateModal = false"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold text-gray-800">Add State</h3>
          <button @click="showAddStateModal = false" class="text-gray-400 hover:text-gray-600">
            ✕
          </button>
        </div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              v-model="newStateName"
              type="text"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Enter state name"
            />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Color</label>
              <input
                v-model="newStateColor"
                type="color"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg cursor-pointer"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Category</label>
              <select
                v-model="newStateCategory"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="unstarted">Unstarted</option>
                <option value="started">Started</option>
                <option value="completed">Completed</option>
              </select>
            </div>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showAddStateModal = false"
            class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            @click="handleSaveState"
            class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
          >
            Add State
          </button>
        </div>
      </div>
    </div>

    <!-- Add Transition Modal -->
    <div
      v-if="showAddTransitionModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="showAddTransitionModal = false"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold text-gray-800">Add Flow</h3>
          <button @click="showAddTransitionModal = false" class="text-gray-400 hover:text-gray-600">
            ✕
          </button>
        </div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">From</label>
            <div class="flex items-center space-x-2 p-3 bg-gray-50 rounded-lg">
              <div
                class="w-3 h-3 rounded-full"
                :style="{ backgroundColor: getStateById(selectedFromState!)?.color }"
              ></div>
              <span class="font-medium text-gray-800">
                {{ getStateById(selectedFromState!)?.name }}
              </span>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Via</label>
            <div class="flex space-x-2">
              <button
                @click="newTransitionType = 'transition'"
                :class="[
                  'flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                  newTransitionType === 'transition'
                    ? 'bg-blue-600 text-white'
                    : 'border border-gray-300 hover:bg-gray-50'
                ]"
              >
                Transition
              </button>
              <button
                @click="newTransitionType = 'approval'"
                :class="[
                  'flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                  newTransitionType === 'approval'
                    ? 'bg-amber-600 text-white'
                    : 'border border-gray-300 hover:bg-gray-50'
                ]"
              >
                Approval
              </button>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Move to</label>
            <select
              v-model="newTransitionTo"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">Select destination state</option>
              <option
                v-for="state in states.filter(s => s.id !== selectedFromState)"
                :key="state.id"
                :value="state.id"
              >
                {{ state.name }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">By</label>
            <select class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent">
              <option value="all">All members</option>
              <option value="specific">Specific roles/members</option>
            </select>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showAddTransitionModal = false"
            class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            @click="handleSaveTransition"
            class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
          >
            Add Flow
          </button>
        </div>
      </div>
    </div>
  </div>
</template>