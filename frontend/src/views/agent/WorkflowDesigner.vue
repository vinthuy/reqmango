<template>
  <div class="workflow-designer" @keydown.esc="clearSelection" tabindex="0">
    <!-- Top Toolbar -->
    <div class="toolbar">
      <div class="toolbar-left">
        <a-button type="text" @click="goBack" class="toolbar-btn">
          <template #icon><ArrowLeftOutlined /></template>
          返回
        </a-button>
        <a-divider type="vertical" />
        <a-input
          v-model:value="workflowName"
          class="workflow-name-input"
          :bordered="false"
          placeholder="工作流名称"
          @blur="handleNameChange"
        />
        <a-tag :color="workflowDetail?.is_active ? 'green' : 'default'" style="margin-left: 8px;">
          {{ workflowDetail?.is_active ? '激活' : '草稿' }}
        </a-tag>
      </div>
      <div class="toolbar-center">
        <a-button-group>
          <a-button @click="zoomOut" :disabled="zoom <= 0.3">
            <template #icon><ZoomOutOutlined /></template>
          </a-button>
          <a-button disabled class="zoom-display">{{ Math.round(zoom * 100) }}%</a-button>
          <a-button @click="zoomIn" :disabled="zoom >= 2">
            <template #icon><ZoomInOutlined /></template>
          </a-button>
          <a-button @click="zoomReset">重置</a-button>
        </a-button-group>
      </div>
      <div class="toolbar-right">
        <a-button @click="handleSave" :loading="saving">
          <template #icon><SaveOutlined /></template>
          保存
        </a-button>
        <a-button type="primary" @click="handleRun" :loading="running">
          <template #icon><CaretRightOutlined /></template>
          运行
        </a-button>
      </div>
    </div>

    <div class="designer-body">
      <!-- Left Panel: Node Palette -->
      <div class="left-panel">
        <div class="panel-header">节点面板</div>
        <div class="node-palette">
          <div
            v-for="paletteNode in paletteNodes"
            :key="paletteNode.type"
            class="palette-item"
            :style="{ borderColor: paletteNode.color }"
            draggable="true"
            @dragstart="onPaletteDragStart($event, paletteNode.type)"
          >
            <div class="palette-icon" :style="{ backgroundColor: paletteNode.color }">
              <component :is="paletteNode.icon" />
            </div>
            <div class="palette-info">
              <div class="palette-name">{{ paletteNode.label }}</div>
              <div class="palette-desc">{{ paletteNode.desc }}</div>
            </div>
          </div>
        </div>
        <div class="panel-section">
          <div class="panel-section-title">节点列表</div>
          <div class="node-list">
            <div
              v-for="node in canvasNodes"
              :key="node.id"
              class="node-list-item"
              :class="{ active: selectedNodeId === node.id }"
              @click="selectNode(node.id)"
            >
              <span class="node-list-dot" :style="{ backgroundColor: getNodeColor(node.type) }"></span>
              <span class="node-list-name">{{ node.name }}</span>
            </div>
            <div v-if="canvasNodes.length === 0" class="empty-hint">
              拖拽节点到画布
            </div>
          </div>
        </div>
      </div>

      <!-- Center: Canvas -->
      <div
        class="canvas-container"
        ref="canvasContainerRef"
        @drop.prevent="onCanvasDrop"
        @dragover.prevent
        @mousedown="onCanvasMouseDown"
        @mousemove="onCanvasMouseMove"
        @mouseup="onCanvasMouseUp"
        @wheel="onCanvasWheel"
        @contextmenu.prevent="onCanvasContextMenu"
      >
        <div
          class="canvas"
          ref="canvasRef"
          :style="{
            transform: `translate(${panX}px, ${panY}px) scale(${zoom})`,
            transformOrigin: '0 0'
          }"
        >
          <!-- SVG layer for edges -->
          <svg class="edges-layer" :style="{ width: canvasWidth + 'px', height: canvasHeight + 'px' }">
            <defs>
              <marker
                id="arrowhead"
                markerWidth="10"
                markerHeight="7"
                refX="10"
                refY="3.5"
                orient="auto"
              >
                <polygon points="0 0, 10 3.5, 0 7" fill="#999" />
              </marker>
              <marker
                id="arrowhead-selected"
                markerWidth="10"
                markerHeight="7"
                refX="10"
                refY="3.5"
                orient="auto"
              >
                <polygon points="0 0, 10 3.5, 0 7" fill="#1890ff" />
              </marker>
            </defs>
            <g v-for="edge in canvasEdges" :key="edge.id">
              <path
                :d="getEdgePath(edge)"
                :stroke="selectedEdgeId === edge.id ? '#1890ff' : '#999'"
                :stroke-width="selectedEdgeId === edge.id ? 2.5 : 1.5"
                fill="none"
                :marker-end="selectedEdgeId === edge.id ? 'url(#arrowhead-selected)' : 'url(#arrowhead)'"
                class="edge-path"
                @click.stop="selectEdge(edge.id)"
                @contextmenu.stop.prevent="onEdgeContextMenu($event, edge.id)"
              />
              <text
                v-if="edge.label"
                :x="getEdgeMidpoint(edge).x"
                :y="getEdgeMidpoint(edge).y - 8"
                text-anchor="middle"
                class="edge-label"
                @click.stop="selectEdge(edge.id)"
              >
                {{ edge.label }}
              </text>
            </g>
            <!-- Temp edge while connecting -->
            <path
              v-if="connectingEdge"
              :d="getTempEdgePath()"
              stroke="#1890ff"
              stroke-width="2"
              stroke-dasharray="6,3"
              fill="none"
            />
          </svg>

          <!-- Node layer -->
          <div
            v-for="node in canvasNodes"
            :key="node.id"
            class="canvas-node"
            :class="{
              selected: selectedNodeId === node.id,
              [`type-${node.type}`]: true
            }"
            :style="{
              left: node.x + 'px',
              top: node.y + 'px',
              width: nodeWidth + 'px'
            }"
            @mousedown.stop="onNodeMouseDown($event, node.id)"
            @click.stop="selectNode(node.id)"
          >
            <div class="node-header" :style="{ backgroundColor: getNodeColor(node.type) }">
              <component :is="getNodeIcon(node.type)" class="node-header-icon" />
              <span class="node-header-text">{{ node.name }}</span>
            </div>
            <div class="node-body">
              <div class="node-type-label">{{ getNodeTypeName(node.type) }}</div>
              <div class="node-config-summary">{{ getNodeSummary(node) }}</div>
            </div>
            <!-- Input connection point -->
            <div
              class="connection-point input-point"
              :class="{ active: connectingEdge && connectingEdge.sourceNodeId !== node.id }"
              @mousedown.stop
              @click.stop="onInputPointClick(node.id)"
            ></div>
            <!-- Output connection point -->
            <div
              class="connection-point output-point"
              :class="{ active: connectingEdge?.sourceNodeId === node.id }"
              @mousedown.stop
              @click.stop="onOutputPointClick(node.id)"
            ></div>
          </div>
        </div>

        <!-- Empty state -->
        <div v-if="canvasNodes.length === 0" class="canvas-empty">
          <div class="canvas-empty-icon">
            <BranchesOutlined style="font-size: 48px; color: #bbb;" />
          </div>
          <div class="canvas-empty-text">从左侧面板拖拽节点到画布开始设计工作流</div>
        </div>

        <!-- Edge context menu -->
        <div
          v-if="edgeContextMenu.visible"
          class="context-menu"
          :style="{ left: edgeContextMenu.x + 'px', top: edgeContextMenu.y + 'px' }"
        >
          <div class="context-menu-item danger" @click="deleteEdgeFromMenu">
            <DeleteOutlined /> 删除连线
          </div>
        </div>
      </div>

      <!-- Right Panel: Properties -->
      <div class="right-panel" v-if="selectedNodeId || selectedEdgeId">
        <div class="panel-header">
          {{ selectedEdgeId ? '连线属性' : '节点属性' }}
          <a-button type="text" size="small" @click="clearSelection">
            <template #icon><CloseOutlined /></template>
          </a-button>
        </div>

        <!-- Node Properties -->
        <div v-if="selectedNodeId && selectedNode" class="properties-content">
          <div class="prop-section">
            <div class="prop-label">类型</div>
            <a-tag :color="getNodeColor(selectedNode.type)">{{ getNodeTypeName(selectedNode.type) }}</a-tag>
          </div>
          <div class="prop-section">
            <div class="prop-label">名称</div>
            <a-input v-model:value="selectedNode.name" size="small" />
          </div>

          <!-- Agent Node Properties -->
          <template v-if="selectedNode.type === 'agent'">
            <div class="prop-section">
              <div class="prop-label">Agent 成员</div>
              <a-select
                v-model:value="selectedNode.config.agent_member_id"
                placeholder="选择 Agent"
                size="small"
                style="width: 100%;"
                show-search
                :filter-option="filterAgentOption"
              >
                <a-select-option v-for="agent in agentMembers" :key="agent.id" :value="agent.id">
                  {{ agent.agent_name }}
                </a-select-option>
              </a-select>
            </div>
            <div class="prop-section">
              <div class="prop-label">任务描述</div>
              <a-textarea
                v-model:value="selectedNode.config.task_description"
                :rows="4"
                size="small"
                placeholder="描述此 Agent 需要执行的任务..."
              />
            </div>
            <div class="prop-section">
              <div class="prop-label">超时时间（秒）</div>
              <a-input-number
                v-model:value="selectedNode.config.timeout_seconds"
                :min="1"
                :max="3600"
                size="small"
                style="width: 100%;"
              />
            </div>
            <div class="prop-section">
              <div class="prop-label">重试次数</div>
              <a-input-number
                v-model:value="selectedNode.config.retry_count"
                :min="0"
                :max="10"
                size="small"
                style="width: 100%;"
              />
            </div>
          </template>

          <!-- Condition Node Properties -->
          <template v-if="selectedNode.type === 'condition'">
            <div class="prop-section">
              <div class="prop-label">条件表达式</div>
              <a-textarea
                v-model:value="selectedNode.config.condition_expression"
                :rows="4"
                size="small"
                class="monospace-input"
                placeholder='例如: output.status === "success"'
              />
            </div>
          </template>

          <!-- Merge Node Properties -->
          <template v-if="selectedNode.type === 'merge'">
            <div class="prop-section">
              <div class="prop-label">合并策略</div>
              <a-radio-group v-model:value="selectedNode.config.strategy" size="small">
                <a-radio value="wait_all">等待全部</a-radio>
                <a-radio value="first_success">首次成功</a-radio>
                <a-radio value="custom">自定义</a-radio>
              </a-radio-group>
            </div>
          </template>

          <!-- Notification Node Properties -->
          <template v-if="selectedNode.type === 'notification'">
            <div class="prop-section">
              <div class="prop-label">通知渠道</div>
              <a-select
                v-model:value="selectedNode.config.channel"
                size="small"
                style="width: 100%;"
                placeholder="选择通知渠道"
              >
                <a-select-option value="email">邮件</a-select-option>
                <a-select-option value="webhook">Webhook</a-select-option>
                <a-select-option value="slack">Slack</a-select-option>
                <a-select-option value="feishu">飞书</a-select-option>
              </a-select>
            </div>
            <div class="prop-section">
              <div class="prop-label">接收者</div>
              <a-textarea
                v-model:value="selectedNode.config.recipients"
                :rows="3"
                size="small"
                placeholder="每行一个接收者"
              />
            </div>
            <div class="prop-section">
              <div class="prop-label">通知模板</div>
              <a-textarea
                v-model:value="selectedNode.config.template"
                :rows="4"
                size="small"
                placeholder="通知消息模板"
              />
            </div>
          </template>

          <a-divider />
          <a-button danger block @click="deleteSelectedNode">
            <template #icon><DeleteOutlined /></template>
            删除节点
          </a-button>
        </div>

        <!-- Edge Properties -->
        <div v-if="selectedEdgeId && selectedEdge" class="properties-content">
          <div class="prop-section">
            <div class="prop-label">标签</div>
            <a-input v-model:value="selectedEdge.label" size="small" placeholder="连线标签" />
          </div>
          <div class="prop-section">
            <div class="prop-label">条件表达式</div>
            <a-textarea
              v-model:value="selectedEdge.condition"
              :rows="3"
              size="small"
              class="monospace-input"
              placeholder="可选的条件表达式"
            />
          </div>
          <div class="prop-section">
            <div class="prop-label">源节点</div>
            <a-input :value="getNodeNameById(selectedEdge.sourceNodeId)" size="small" disabled />
          </div>
          <div class="prop-section">
            <div class="prop-label">目标节点</div>
            <a-input :value="getNodeNameById(selectedEdge.targetNodeId)" size="small" disabled />
          </div>
          <a-divider />
          <a-button danger block @click="deleteSelectedEdge">
            <template #icon><DeleteOutlined /></template>
            删除连线
          </a-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  ArrowLeftOutlined,
  ZoomInOutlined,
  ZoomOutOutlined,
  SaveOutlined,
  CaretRightOutlined,
  CloseOutlined,
  DeleteOutlined,
  BranchesOutlined,
  RobotOutlined,
  ForkOutlined,
  MergeCellsOutlined,
  BellOutlined,
} from '@ant-design/icons-vue'
import { workflowApi } from '@/api/workflow'
import { agentMemberApi } from '@/api/agent-member'
import type { AgentMember } from '@/api/agent-member'
import type { WorkflowDetail } from '@/api/workflow'

// ==================== Types ====================

interface CanvasNode {
  id: string
  type: 'agent' | 'condition' | 'merge' | 'notification'
  name: string
  x: number
  y: number
  config: Record<string, any>
  serverNodeId?: number
}

interface CanvasEdge {
  id: string
  sourceNodeId: string
  targetNodeId: string
  label: string
  condition: string
  serverEdgeId?: number
}

interface PaletteNodeDef {
  type: CanvasNode['type']
  label: string
  desc: string
  color: string
  icon: any
}

// ==================== Route & Params ====================

const route = useRoute()
const router = useRouter()

const projectId = computed(() => Number(route.params.id || route.params.projectId))
const workflowId = computed(() => Number(route.params.workflowId))

// ==================== Palette ====================

const paletteNodes: PaletteNodeDef[] = [
  { type: 'agent', label: 'Agent 节点', desc: '执行 AI Agent 任务', color: '#1890ff', icon: RobotOutlined },
  { type: 'condition', label: '条件节点', desc: '条件分支判断', color: '#faad14', icon: ForkOutlined },
  { type: 'merge', label: '合并节点', desc: '合并多个流程', color: '#52c41a', icon: MergeCellsOutlined },
  { type: 'notification', label: '通知节点', desc: '发送通知消息', color: '#722ed1', icon: BellOutlined },
]

// ==================== State ====================

const workflowDetail = ref<WorkflowDetail | null>(null)
const workflowName = ref('')
const saving = ref(false)
const running = ref(false)
const agentMembers = ref<AgentMember[]>([])

// Canvas state
const canvasNodes = ref<CanvasNode[]>([])
const canvasEdges = ref<CanvasEdge[]>([])
const selectedNodeId = ref<string | null>(null)
const selectedEdgeId = ref<string | null>(null)
const zoom = ref(1)
const panX = ref(0)
const panY = ref(0)
const nodeWidth = 200
const nodeHeight = 90
const canvasWidth = 5000
const canvasHeight = 5000

// Interaction state
const isPanning = ref(false)
const panStartX = ref(0)
const panStartY = ref(0)
const draggingNodeId = ref<string | null>(null)
const dragOffsetX = ref(0)
const dragOffsetY = ref(0)
const connectingEdge = ref<{ sourceNodeId: string; sourcePoint: { x: number; y: number }; mouseX: number; mouseY: number } | null>(null)

// Context menu
const edgeContextMenu = reactive({ visible: false, x: 0, y: 0, edgeId: '' })

// Counters for naming
const nodeCounters: Record<string, number> = { agent: 0, condition: 0, merge: 0, notification: 0 }

// Refs
const canvasContainerRef = ref<HTMLDivElement | null>(null)
const canvasRef = ref<HTMLDivElement | null>(null)

// ==================== Computed ====================

const selectedNode = computed(() => {
  if (!selectedNodeId.value) return null
  return canvasNodes.value.find(n => n.id === selectedNodeId.value) || null
})

const selectedEdge = computed(() => {
  if (!selectedEdgeId.value) return null
  return canvasEdges.value.find(e => e.id === selectedEdgeId.value) || null
})

// ==================== Helpers ====================

function getNodeColor(type: CanvasNode['type']): string {
  return paletteNodes.find(p => p.type === type)?.color || '#999'
}

function getNodeIcon(type: CanvasNode['type']) {
  return paletteNodes.find(p => p.type === type)?.icon || RobotOutlined
}

function getNodeTypeName(type: CanvasNode['type']): string {
  const map: Record<string, string> = { agent: 'Agent', condition: '条件', merge: '合并', notification: '通知' }
  return map[type] || type
}

function getNodeSummary(node: CanvasNode): string {
  if (node.type === 'agent') {
    return node.config.task_description ? node.config.task_description.substring(0, 40) : '未设置任务'
  }
  if (node.type === 'condition') {
    return node.config.condition_expression ? node.config.condition_expression.substring(0, 40) : '未设置条件'
  }
  if (node.type === 'merge') {
    const map: Record<string, string> = { wait_all: '等待全部', first_success: '首次成功', custom: '自定义' }
    return map[node.config.strategy] || '等待全部'
  }
  if (node.type === 'notification') {
    return node.config.channel ? `渠道: ${node.config.channel}` : '未设置渠道'
  }
  return ''
}

function getNodeNameById(id: string): string {
  return canvasNodes.value.find(n => n.id === id)?.name || id
}

function generateNodeId(): string {
  return 'local_' + Date.now() + '_' + Math.random().toString(36).substring(2, 8)
}

function generateEdgeId(): string {
  return 'edge_' + Date.now() + '_' + Math.random().toString(36).substring(2, 8)
}

function getInputPointPos(node: CanvasNode): { x: number; y: number } {
  return { x: node.x, y: node.y + nodeHeight / 2 }
}

function getOutputPointPos(node: CanvasNode): { x: number; y: number } {
  return { x: node.x + nodeWidth, y: node.y + nodeHeight / 2 }
}

function getEdgePath(edge: CanvasEdge): string {
  const sourceNode = canvasNodes.value.find(n => n.id === edge.sourceNodeId)
  const targetNode = canvasNodes.value.find(n => n.id === edge.targetNodeId)
  if (!sourceNode || !targetNode) return ''

  const start = getOutputPointPos(sourceNode)
  const end = getInputPointPos(targetNode)

  const dx = Math.abs(end.x - start.x)
  const cp = Math.max(dx * 0.4, 50)

  return `M ${start.x} ${start.y} C ${start.x + cp} ${start.y}, ${end.x - cp} ${end.y}, ${end.x} ${end.y}`
}

function getEdgeMidpoint(edge: CanvasEdge): { x: number; y: number } {
  const sourceNode = canvasNodes.value.find(n => n.id === edge.sourceNodeId)
  const targetNode = canvasNodes.value.find(n => n.id === edge.targetNodeId)
  if (!sourceNode || !targetNode) return { x: 0, y: 0 }

  const start = getOutputPointPos(sourceNode)
  const end = getInputPointPos(targetNode)
  return { x: (start.x + end.x) / 2, y: (start.y + end.y) / 2 }
}

function getTempEdgePath(): string {
  if (!connectingEdge.value) return ''
  const start = connectingEdge.value.sourcePoint
  const end = { x: connectingEdge.value.mouseX, y: connectingEdge.value.mouseY }

  const dx = Math.abs(end.x - start.x)
  const cp = Math.max(dx * 0.4, 50)

  return `M ${start.x} ${start.y} C ${start.x + cp} ${start.y}, ${end.x - cp} ${end.y}, ${end.x} ${end.y}`
}

function getDefaultConfig(type: CanvasNode['type']): Record<string, any> {
  if (type === 'agent') {
    return { agent_member_id: undefined, task_description: '', timeout_seconds: 300, retry_count: 0 }
  }
  if (type === 'condition') {
    return { condition_expression: '' }
  }
  if (type === 'merge') {
    return { strategy: 'wait_all' }
  }
  if (type === 'notification') {
    return { channel: undefined, recipients: '', template: '' }
  }
  return {}
}

function filterAgentOption(input: string, option: any): boolean {
  const agent = agentMembers.value.find(a => a.id === option.value)
  return agent ? agent.agent_name.toLowerCase().includes(input.toLowerCase()) : false
}

// ==================== Selection ====================

function selectNode(id: string) {
  selectedNodeId.value = null
  selectedEdgeId.value = null
  nextTick(() => { selectedNodeId.value = id })
}

function selectEdge(id: string) {
  selectedNodeId.value = null
  selectedEdgeId.value = null
  nextTick(() => { selectedEdgeId.value = id })
}

function clearSelection() {
  selectedNodeId.value = null
  selectedEdgeId.value = null
}

// ==================== Palette Drag & Drop ====================

function onPaletteDragStart(event: DragEvent, type: CanvasNode['type']) {
  if (event.dataTransfer) {
    event.dataTransfer.setData('nodeType', type)
    event.dataTransfer.effectAllowed = 'copy'
  }
}

function onCanvasDrop(event: DragEvent) {
  const nodeType = event.dataTransfer?.getData('nodeType') as CanvasNode['type']
  if (!nodeType) return

  const container = canvasContainerRef.value
  if (!container) return

  const rect = container.getBoundingClientRect()
  const x = (event.clientX - rect.left - panX.value) / zoom.value - nodeWidth / 2
  const y = (event.clientY - rect.top - panY.value) / zoom.value - nodeHeight / 2

  nodeCounters[nodeType]++
  const newNode: CanvasNode = {
    id: generateNodeId(),
    type: nodeType,
    name: `${getNodeTypeName(nodeType)} ${nodeCounters[nodeType]}`,
    x: Math.round(x),
    y: Math.round(y),
    config: getDefaultConfig(nodeType),
  }

  canvasNodes.value.push(newNode)
  selectNode(newNode.id)
}

// ==================== Canvas Panning ====================

function onCanvasMouseDown(event: MouseEvent) {
  if (event.button === 0 && !draggingNodeId.value) {
    isPanning.value = true
    panStartX.value = event.clientX - panX.value
    panStartY.value = event.clientY - panY.value
    clearSelection()
  }
}

function onCanvasMouseMove(event: MouseEvent) {
  if (isPanning.value) {
    panX.value = event.clientX - panStartX.value
    panY.value = event.clientY - panStartY.value
    return
  }

  if (draggingNodeId.value) {
    const node = canvasNodes.value.find(n => n.id === draggingNodeId.value)
    if (node) {
      const container = canvasContainerRef.value
      if (!container) return
      const rect = container.getBoundingClientRect()
      node.x = Math.round((event.clientX - rect.left - panX.value) / zoom.value - dragOffsetX.value)
      node.y = Math.round((event.clientY - rect.top - panY.value) / zoom.value - dragOffsetY.value)
    }
    return
  }

  if (connectingEdge.value) {
    const container = canvasContainerRef.value
    if (!container) return
    const rect = container.getBoundingClientRect()
    connectingEdge.value.mouseX = (event.clientX - rect.left - panX.value) / zoom.value
    connectingEdge.value.mouseY = (event.clientY - rect.top - panY.value) / zoom.value
  }
}

function onCanvasMouseUp() {
  isPanning.value = false
  draggingNodeId.value = null
  if (connectingEdge.value) {
    connectingEdge.value = null
  }
  edgeContextMenu.visible = false
}

function onCanvasWheel(event: WheelEvent) {
  event.preventDefault()
  const delta = event.deltaY > 0 ? -0.08 : 0.08
  const newZoom = Math.min(2, Math.max(0.3, zoom.value + delta))

  const container = canvasContainerRef.value
  if (container) {
    const rect = container.getBoundingClientRect()
    const mouseX = event.clientX - rect.left
    const mouseY = event.clientY - rect.top

    panX.value = mouseX - (mouseX - panX.value) * (newZoom / zoom.value)
    panY.value = mouseY - (mouseY - panY.value) * (newZoom / zoom.value)
  }

  zoom.value = newZoom
}

function onCanvasContextMenu(event: MouseEvent) {
  edgeContextMenu.visible = false
}

// ==================== Node Interaction ====================

function onNodeMouseDown(event: MouseEvent, nodeId: string) {
  if (event.button !== 0) return
  const node = canvasNodes.value.find(n => n.id === nodeId)
  if (!node) return

  draggingNodeId.value = nodeId
  const container = canvasContainerRef.value
  if (!container) return
  const rect = container.getBoundingClientRect()
  dragOffsetX.value = (event.clientX - rect.left - panX.value) / zoom.value - node.x
  dragOffsetY.value = (event.clientY - rect.top - panY.value) / zoom.value - node.y
}

// ==================== Edge Connection ====================

function onOutputPointClick(nodeId: string) {
  const node = canvasNodes.value.find(n => n.id === nodeId)
  if (!node) return

  const pos = getOutputPointPos(node)
  connectingEdge.value = {
    sourceNodeId: nodeId,
    sourcePoint: pos,
    mouseX: pos.x,
    mouseY: pos.y,
  }
}

function onInputPointClick(nodeId: string) {
  if (!connectingEdge.value) return
  if (connectingEdge.value.sourceNodeId === nodeId) {
    connectingEdge.value = null
    return
  }

  // Check for duplicate edge
  const exists = canvasEdges.value.some(
    e => e.sourceNodeId === connectingEdge.value!.sourceNodeId && e.targetNodeId === nodeId
  )
  if (exists) {
    message.warning('已存在相同的连线')
    connectingEdge.value = null
    return
  }

  const newEdge: CanvasEdge = {
    id: generateEdgeId(),
    sourceNodeId: connectingEdge.value.sourceNodeId,
    targetNodeId: nodeId,
    label: '',
    condition: '',
  }

  canvasEdges.value.push(newEdge)
  connectingEdge.value = null
  selectEdge(newEdge.id)
}

// ==================== Edge Context Menu ====================

function onEdgeContextMenu(event: MouseEvent, edgeId: string) {
  edgeContextMenu.visible = true
  edgeContextMenu.x = event.clientX
  edgeContextMenu.y = event.clientY
  edgeContextMenu.edgeId = edgeId
}

function deleteEdgeFromMenu() {
  const idx = canvasEdges.value.findIndex(e => e.id === edgeContextMenu.edgeId)
  if (idx !== -1) {
    canvasEdges.value.splice(idx, 1)
    if (selectedEdgeId.value === edgeContextMenu.edgeId) {
      selectedEdgeId.value = null
    }
  }
  edgeContextMenu.visible = false
}

// ==================== Zoom Controls ====================

function zoomIn() {
  zoom.value = Math.min(2, zoom.value + 0.1)
}

function zoomOut() {
  zoom.value = Math.max(0.3, zoom.value - 0.1)
}

function zoomReset() {
  zoom.value = 1
  panX.value = 0
  panY.value = 0
}

// ==================== Node / Edge Deletion ====================

function deleteSelectedNode() {
  if (!selectedNodeId.value) return
  const id = selectedNodeId.value

  // Remove connected edges
  canvasEdges.value = canvasEdges.value.filter(
    e => e.sourceNodeId !== id && e.targetNodeId !== id
  )

  canvasNodes.value = canvasNodes.value.filter(n => n.id !== id)
  selectedNodeId.value = null
}

function deleteSelectedEdge() {
  if (!selectedEdgeId.value) return
  const idx = canvasEdges.value.findIndex(e => e.id === selectedEdgeId.value)
  if (idx !== -1) {
    canvasEdges.value.splice(idx, 1)
  }
  selectedEdgeId.value = null
}

// ==================== Data Loading ====================

async function loadWorkflow() {
  try {
    const res = await workflowApi.get(projectId.value, workflowId.value)
    const detail = res.data
    workflowDetail.value = detail
    workflowName.value = detail.name

    // Convert server nodes to canvas nodes
    if (detail.nodes && detail.nodes.length > 0) {
      canvasNodes.value = detail.nodes.map((node, index) => {
        const nodeType = (node.node_type || 'agent') as CanvasNode['type']
        nodeCounters[nodeType]++
        return {
          id: `server_${node.id}`,
          type: nodeType,
          name: node.name,
          x: 100 + (index % 4) * 250,
          y: 100 + Math.floor(index / 4) * 180,
          config: node.config || getDefaultConfig(nodeType),
          serverNodeId: node.id,
        }
      })
    }

    // Convert server edges to canvas edges
    if (detail.edges && detail.edges.length > 0) {
      canvasEdges.value = detail.edges.map(edge => ({
        id: `server_${edge.id}`,
        sourceNodeId: `server_${edge.source_node_id}`,
        targetNodeId: `server_${edge.target_node_id}`,
        label: '',
        condition: edge.condition || '',
        serverEdgeId: edge.id,
      }))
    }
  } catch (e: any) {
    message.error('加载工作流失败：' + (e.message || '未知错误'))
  }
}

async function loadAgentMembers() {
  try {
    const res = await agentMemberApi.list(projectId.value)
    agentMembers.value = res.data.data || []
  } catch (e: any) {
    message.error('加载 Agent 列表失败')
  }
}

// ==================== Save ====================

async function handleSave() {
  saving.value = true
  try {
    // Save workflow name
    await workflowApi.update(projectId.value, workflowId.value, {
      name: workflowName.value,
    })

    // Delete nodes that were removed
    if (workflowDetail.value?.nodes) {
      const currentServerIds = canvasNodes.value
        .filter(n => n.serverNodeId)
        .map(n => n.serverNodeId!)
      const toDelete = workflowDetail.value.nodes.filter(n => !currentServerIds.includes(n.id))
      for (const node of toDelete) {
        try {
          await workflowApi.deleteNode(projectId.value, workflowId.value, node.id)
        } catch {
          // node may already be deleted
        }
      }
    }

    // Delete edges that were removed
    if (workflowDetail.value?.edges) {
      const currentEdgeIds = canvasEdges.value
        .filter(e => e.serverEdgeId)
        .map(e => e.serverEdgeId!)
      const toDeleteEdges = workflowDetail.value.edges.filter(e => !currentEdgeIds.includes(e.id))
      for (const edge of toDeleteEdges) {
        try {
          await workflowApi.deleteEdge(projectId.value, workflowId.value, edge.id)
        } catch {
          // edge may already be deleted
        }
      }
    }

    // Create or update nodes
    for (const node of canvasNodes.value) {
      const agentId = node.config.agent_member_id || 0
      const payload = {
        agent_id: agentId,
        node_type: node.type,
        name: node.name,
        config: node.config,
        timeout: node.config.timeout_seconds || 300,
        max_retries: node.config.retry_count || 0,
      }

      if (node.serverNodeId) {
        try {
          await workflowApi.updateNode(projectId.value, workflowId.value, node.serverNodeId, payload)
        } catch {
          // if update fails, try creating new
          const res = await workflowApi.addNode(projectId.value, workflowId.value, payload)
          node.serverNodeId = res.data.id
        }
      } else {
        const res = await workflowApi.addNode(projectId.value, workflowId.value, payload)
        node.serverNodeId = res.data.id
      }
    }

    // Create edges
    for (const edge of canvasEdges.value) {
      const sourceNode = canvasNodes.value.find(n => n.id === edge.sourceNodeId)
      const targetNode = canvasNodes.value.find(n => n.id === edge.targetNodeId)

      if (!sourceNode?.serverNodeId || !targetNode?.serverNodeId) continue

      if (!edge.serverEdgeId) {
        try {
          const res = await workflowApi.addEdge(projectId.value, workflowId.value, {
            source_node_id: sourceNode.serverNodeId,
            target_node_id: targetNode.serverNodeId,
            condition: edge.condition || undefined,
          })
          edge.serverEdgeId = res.data.id
        } catch {
          // skip failed edge creation
        }
      }
    }

    message.success('工作流保存成功')
  } catch (e: any) {
    message.error('保存失败：' + (e.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

// ==================== Run ====================

async function handleRun() {
  running.value = true
  try {
    await workflowApi.execute(projectId.value, workflowId.value)
    message.success('工作流已启动运行')
  } catch (e: any) {
    message.error('运行失败：' + (e.message || '未知错误'))
  } finally {
    running.value = false
  }
}

// ==================== Navigation ====================

function goBack() {
  router.back()
}

function handleNameChange() {
  // Name is saved on explicit save
}

// ==================== Keyboard Shortcuts ====================

function onKeyDown(event: KeyboardEvent) {
  if (event.key === 'Delete' || event.key === 'Backspace') {
    if (selectedNodeId.value && !(event.target as HTMLElement).closest('input, textarea, [contenteditable]')) {
      deleteSelectedNode()
    }
    if (selectedEdgeId.value && !(event.target as HTMLElement).closest('input, textarea, [contenteditable]')) {
      deleteSelectedEdge()
    }
  }
  if (event.key === 'Escape') {
    connectingEdge.value = null
    edgeContextMenu.visible = false
    clearSelection()
  }
}

// ==================== Lifecycle ====================

onMounted(async () => {
  document.addEventListener('keydown', onKeyDown)
  await Promise.all([loadWorkflow(), loadAgentMembers()])
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeyDown)
})
</script>

<style scoped>
.workflow-designer {
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f0f2f5;
  overflow: hidden;
  outline: none;
}

/* ==================== Toolbar ==================== */
.toolbar {
  height: 52px;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  flex-shrink: 0;
  z-index: 100;
}

.toolbar-left,
.toolbar-center,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-btn {
  color: #666;
}

.workflow-name-input {
  font-size: 16px;
  font-weight: 600;
  width: 260px;
}

.workflow-name-input :deep(.ant-input) {
  font-size: 16px;
  font-weight: 600;
}

.zoom-display {
  min-width: 60px;
  cursor: default;
  font-size: 12px;
  color: #666;
}

/* ==================== Designer Body ==================== */
.designer-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* ==================== Left Panel ==================== */
.left-panel {
  width: 240px;
  background: #1a1a2e;
  color: #e0e0e0;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow-y: auto;
}

.panel-header {
  padding: 14px 16px;
  font-size: 13px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #aaa;
  border-bottom: 1px solid #2a2a40;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.node-palette {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.palette-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: #252540;
  border: 1px solid #333;
  border-left: 3px solid;
  border-radius: 6px;
  cursor: grab;
  transition: all 0.15s;
  user-select: none;
}

.palette-item:hover {
  background: #2e2e50;
  transform: translateX(2px);
}

.palette-item:active {
  cursor: grabbing;
}

.palette-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 16px;
  flex-shrink: 0;
}

.palette-info {
  flex: 1;
  min-width: 0;
}

.palette-name {
  font-size: 13px;
  font-weight: 500;
  color: #e0e0e0;
}

.palette-desc {
  font-size: 11px;
  color: #888;
  margin-top: 2px;
}

.panel-section {
  border-top: 1px solid #2a2a40;
  padding: 12px;
}

.panel-section-title {
  font-size: 12px;
  font-weight: 600;
  color: #aaa;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
}

.node-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.node-list-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  color: #ccc;
  transition: background 0.15s;
}

.node-list-item:hover {
  background: #2a2a45;
}

.node-list-item.active {
  background: #2a2a60;
  color: #fff;
}

.node-list-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.node-list-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-hint {
  font-size: 12px;
  color: #666;
  text-align: center;
  padding: 16px 0;
}

/* ==================== Canvas ==================== */
.canvas-container {
  flex: 1;
  position: relative;
  overflow: hidden;
  cursor: grab;
  background-color: #fafafa;
  background-image:
    linear-gradient(rgba(0, 0, 0, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 0, 0, 0.04) 1px, transparent 1px);
  background-size: 20px 20px;
}

.canvas-container:active {
  cursor: grabbing;
}

.canvas {
  position: absolute;
  width: 5000px;
  height: 5000px;
  will-change: transform;
}

.edges-layer {
  position: absolute;
  top: 0;
  left: 0;
  pointer-events: none;
}

.edges-layer .edge-path {
  pointer-events: stroke;
  cursor: pointer;
}

.edges-layer .edge-label {
  font-size: 11px;
  fill: #666;
  pointer-events: all;
  cursor: pointer;
}

.canvas-empty {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  pointer-events: none;
}

.canvas-empty-text {
  font-size: 14px;
  color: #aaa;
  margin-top: 12px;
}

/* ==================== Canvas Node ==================== */
.canvas-node {
  position: absolute;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: move;
  user-select: none;
  transition: box-shadow 0.15s;
}

.canvas-node:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.canvas-node.selected {
  box-shadow: 0 0 0 2px #1890ff, 0 4px 16px rgba(24, 144, 255, 0.2);
}

.node-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 8px 8px 0 0;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
}

.node-header-icon {
  font-size: 14px;
}

.node-header-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-body {
  padding: 8px 12px;
}

.node-type-label {
  font-size: 10px;
  color: #999;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.node-config-summary {
  font-size: 12px;
  color: #666;
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ==================== Connection Points ==================== */
.connection-point {
  position: absolute;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #d9d9d9;
  border: 2px solid #fff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  cursor: crosshair;
  z-index: 10;
  transition: all 0.15s;
}

.connection-point:hover,
.connection-point.active {
  background: #1890ff;
  transform: scale(1.3);
}

.input-point {
  left: -6px;
  top: 50%;
  margin-top: -6px;
}

.output-point {
  right: -6px;
  top: 50%;
  margin-top: -6px;
}

/* ==================== Right Panel ==================== */
.right-panel {
  width: 300px;
  background: #fff;
  border-left: 1px solid #e8e8e8;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow-y: auto;
}

.properties-content {
  padding: 16px;
}

.prop-section {
  margin-bottom: 14px;
}

.prop-label {
  font-size: 12px;
  color: #888;
  margin-bottom: 4px;
  font-weight: 500;
}

.monospace-input :deep(.ant-input),
.monospace-input :deep(textarea) {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
}

/* ==================== Context Menu ==================== */
.context-menu {
  position: fixed;
  background: #fff;
  border-radius: 6px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  padding: 4px 0;
  z-index: 1000;
  min-width: 140px;
}

.context-menu-item {
  padding: 8px 16px;
  font-size: 13px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: background 0.1s;
}

.context-menu-item:hover {
  background: #f5f5f5;
}

.context-menu-item.danger {
  color: #ff4d4f;
}

.context-menu-item.danger:hover {
  background: #fff1f0;
}
</style>
