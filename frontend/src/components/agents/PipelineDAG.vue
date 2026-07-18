<template>
  <div class="pipeline-dag">
    <svg :width="svgWidth" :height="svgHeight" class="w-full">
      <defs>
        <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto">
          <polygon points="0 0, 10 3.5, 0 7" fill="#9ca3af" />
        </marker>
      </defs>
      <!-- Draw edges -->
      <line v-for="(edge, i) in edges" :key="'e'+i"
        :x1="edge.x1" :y1="edge.y1" :x2="edge.x2" :y2="edge.y2"
        stroke="#9ca3af" stroke-width="2" marker-end="url(#arrowhead)" />
      <!-- Draw nodes -->
      <g v-for="(node, i) in nodes" :key="'n'+i">
        <rect :x="node.x-60" :y="node.y-20" width="120" height="40" rx="8"
          :fill="nodeColor(node.type)" stroke="#e5e7eb" stroke-width="1.5" />
        <text :x="node.x" :y="node.y+5" text-anchor="middle" fill="#1f2937" font-size="12" font-weight="600">
          {{ node.label }}
        </text>
      </g>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ stages: { name: string; stage_type: string }[]; mode: string }>()

type DAGNode = { label: string; type: string; x: number; y: number }
type DAGEdge = { x1: number; y1: number; x2: number; y2: number }

const nodeWidth = 140; const nodeHeight = 50; const hGap = 80; const vGap = 60

const nodes = computed<DAGNode[]>(() => {
  if (props.mode === 'fan_out' || props.mode === 'tournament') {
    const ns: DAGNode[] = []
    // Planner at top
    const planners = props.stages.filter(s=>s.stage_type==='planner')
    const executors = props.stages.filter(s=>s.stage_type==='executor')
    const reviewers = props.stages.filter(s=>s.stage_type==='reviewer'||s.stage_type==='judge')
    planners.forEach((s,i)=>ns.push({label:s.name,type:s.stage_type,x:250,y:40+i*80}))
    const execWidth = executors.length * (nodeWidth + hGap) - hGap
    const startX = 250 - execWidth/2 + nodeWidth/2
    executors.forEach((s,i)=>ns.push({label:s.name,type:s.stage_type,x:startX+i*(nodeWidth+hGap),y:180}))
    reviewers.forEach((s,i)=>ns.push({label:s.name,type:s.stage_type,x:250,y:320+i*80}))
    return ns
  }
  // Sequential: vertical line
  return props.stages.map((s, i) => ({
    label: s.name, type: s.stage_type, x: 250, y: 40 + i * (nodeHeight + vGap)
  }))
})

const edges = computed<DAGEdge[]>(() => {
  const es: DAGEdge[] = []
  const ns = nodes.value
  if (props.mode === 'fan_out' || props.mode === 'tournament') {
    const planner = ns.find(n=>n.type==='planner')
    const executors = ns.filter(n=>n.type==='executor')
    const reviewer = ns.find(n=>n.type==='reviewer'||n.type==='judge')
    if (planner) executors.forEach(e=>es.push({x1:planner.x,y1:planner.y+20,x2:e.x,y2:e.y-20}))
    if (reviewer) executors.forEach(e=>es.push({x1:e.x,y1:e.y+20,x2:reviewer.x,y2:reviewer.y-20}))
    return es
  }
  for (let i = 0; i < ns.length-1; i++) {
    es.push({x1:ns[i].x,y1:ns[i].y+20,x2:ns[i+1].x,y2:ns[i+1].y-20})
  }
  return es
})

const svgWidth = computed(() => {
  if (props.mode==='fan_out'||props.mode==='tournament') return Math.max(500, nodes.value.length*(nodeWidth+hGap))
  return 500
})
const svgHeight = computed(() => {
  if (props.mode==='fan_out'||props.mode==='tournament') return 420
  return nodes.value.length * (nodeHeight+vGap) + 40
})

function nodeColor(type: string): string {
  switch(type) { case 'planner': return '#e3f2fd'; case 'executor': return '#e8f5e9'; case 'reviewer': return '#fff3e0'; case 'judge': return '#fce4ec'; default: return '#f5f5f5' }
}
</script>
