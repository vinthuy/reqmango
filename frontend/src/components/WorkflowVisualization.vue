<template>
  <div class="workflow-visualization" ref="containerRef">
    <svg :width="svgWidth" :height="svgHeight" class="w-full h-full">
      <defs>
        <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto">
          <polygon points="0 0, 10 3.5, 0 7" fill="#6366f1" />
        </marker>
        <marker id="arrowhead-approval" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto">
          <polygon points="0 0, 10 3.5, 0 7" fill="#f59e0b" />
        </marker>
      </defs>
      
      <!-- Transition lines -->
      <g v-for="transition in transitions" :key="`t-${transition.id || transition.source_state_id + '-' + transition.target_state_id}`">
        <path
          :d="getTransitionPath(transition)"
          :stroke="transition.rule_type === 'approval' ? '#f59e0b' : '#6366f1'"
          :stroke-width="2"
          :stroke-dasharray="transition.rule_type === 'approval' ? '5,5' : 'none'"
          fill="none"
          :marker-end="transition.rule_type === 'approval' ? 'url(#arrowhead-approval)' : 'url(#arrowhead)'"
          class="transition-line"
        />
        <text
          :x="getTransitionMidpoint(transition).x"
          :y="getTransitionMidpoint(transition).y - 8"
          text-anchor="middle"
          class="text-[10px] fill-gray-500"
        >
          {{ transition.rule_type === 'approval' ? t('workflow.approval') : t('workflow.allow') }}
        </text>
      </g>
      
      <!-- State nodes -->
      <g v-for="node in nodes" :key="node.id" class="state-node" @click="$emit('select-state', node.state)">
        <rect
          :x="node.x - nodeWidth / 2"
          :y="node.y - nodeHeight / 2"
          :width="nodeWidth"
          :height="nodeHeight"
          rx="8"
          :fill="node.state.color + '20'"
          :stroke="node.state.color"
          stroke-width="2"
          class="cursor-pointer hover:shadow-lg transition-shadow"
        />
        <circle
          :cx="node.x"
          :cy="node.y - 10"
          r="6"
          :fill="node.state.color"
        />
        <text
          :x="node.x"
          :y="node.y + 12"
          text-anchor="middle"
          class="text-xs font-medium fill-gray-700"
        >
          {{ node.state.name }}
        </text>
        <text
          :x="node.x"
          :y="node.y + 26"
          text-anchor="middle"
          class="text-[10px] fill-gray-400"
        >
          {{ getStateGroupLabel(node.state.group) }}
        </text>
      </g>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

interface State {
  id: number
  name: string
  color: string
  group: string
}

interface Transition {
  id?: number
  source_state_id: number
  target_state_id: number
  rule_type: string
  source_name?: string
  target_name?: string
}

const props = defineProps<{
  states: State[]
  transitions: Transition[]
}>()

defineEmits<{
  (e: 'select-state', state: State): void
}>()

const containerRef = ref<HTMLElement | null>(null)
const nodeWidth = 120
const nodeHeight = 70
const padding = 40

// Calculate node positions using a simple grid layout
const nodes = computed(() => {
  
  const cols = Math.ceil(Math.sqrt(props.states.length))
  
  return props.states.map((state, index) => {
    const col = index % cols
    const row = Math.floor(index / cols)
    return {
      id: state.id,
      state,
      x: padding + nodeWidth / 2 + col * (nodeWidth + 60),
      y: padding + nodeHeight / 2 + row * (nodeHeight + 40)
    }
  })
})

const svgWidth = computed(() => {
  const cols = Math.ceil(Math.sqrt(props.states.length))
  return Math.max(400, cols * (nodeWidth + 60) + padding * 2)
})

const svgHeight = computed(() => {
  const cols = Math.ceil(Math.sqrt(props.states.length))
  const rows = Math.ceil(props.states.length / cols)
  return Math.max(200, rows * (nodeHeight + 40) + padding * 2)
})

function getNodeCenter(stateId: number) {
  const node = nodes.value.find(n => n.id === stateId)
  return node ? { x: node.x, y: node.y } : { x: 0, y: 0 }
}

function getTransitionPath(transition: Transition) {
  const from = getNodeCenter(transition.source_state_id)
  const to = getNodeCenter(transition.target_state_id)
  
  // Calculate control points for a curved line
  const dx = to.x - from.x
  const dy = to.y - from.y
  const dist = Math.sqrt(dx * dx + dy * dy)
  
  // Offset start and end to be at the edge of the node
  const angle = Math.atan2(dy, dx)
  const startX = from.x + Math.cos(angle) * (nodeWidth / 2)
  const startY = from.y + Math.sin(angle) * (nodeHeight / 2)
  const endX = to.x - Math.cos(angle) * (nodeWidth / 2)
  const endY = to.y - Math.sin(angle) * (nodeHeight / 2)
  
  // Control point offset for curve
  const controlOffset = Math.min(dist * 0.2, 50)
  const midX = (startX + endX) / 2
  const midY = (startY + endY) / 2 - controlOffset
  
  return `M ${startX} ${startY} Q ${midX} ${midY} ${endX} ${endY}`
}

function getTransitionMidpoint(transition: Transition) {
  const from = getNodeCenter(transition.source_state_id)
  const to = getNodeCenter(transition.target_state_id)
  return {
    x: (from.x + to.x) / 2,
    y: (from.y + to.y) / 2
  }
}

function getStateGroupLabel(group: string) {
  const labels: Record<string, string> = {
    backlog: t('settings.stateGroupBacklogName'),
    unstarted: t('settings.stateGroupUnstartedName'),
    started: t('settings.stateGroupStartedName'),
    completed: t('settings.stateGroupCompletedName'),
    cancelled: t('settings.stateGroupCancelledName')
  }
  return labels[group] || group
}
</script>

<style scoped>
.workflow-visualization {
  min-height: 300px;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 0.5rem;
  overflow: auto;
}

.transition-line {
  transition: stroke-width 0.2s;
}

.transition-line:hover {
  stroke-width: 3;
}

.state-node:hover rect {
  filter: drop-shadow(0 4px 6px rgba(0, 0, 0, 0.1));
}
</style>
