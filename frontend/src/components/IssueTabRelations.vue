<template>
  <div class="space-y-4">
    <RelationParentCard
      :parent="parent"
      :issue-types="issueTypes"
      @change="$emit('changeParent')"
      @remove="removeParent"
      @navigate="(id: number) => $emit('navigate', id)"
    />

    <RelationSubIssuesCard
      :sub-issues="subIssues"
      @add="createSubIssue"
      @toggle="toggleSubIssue"
      @navigate="(id: number) => $emit('navigate', id)"
    />

    <RelationTypeCard
      v-for="group in relationGroups"
      :key="group.typeId"
      :type-name="group.typeName"
      :items="group.items"
      :color="group.color"
      :issue-types="issueTypes"
      @add="$emit('addRelation', group.typeId)"
      @remove="removeRelation"
      @navigate="(id: number) => $emit('navigate', id)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listIssueRelations, listRelationTypes, deleteIssueRelation } from '@/api/relation'
import { issueApi } from '@/api/issue'
import RelationParentCard from './RelationParentCard.vue'
import RelationSubIssuesCard from './RelationSubIssuesCard.vue'
import RelationTypeCard from './RelationTypeCard.vue'

const COLORS = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316']

const router = useRouter()

interface IssueType {
  id: number
  name: string
  color: string
}

interface ParentIssue {
  id: number
  sequence_id: number
  name: string
  state_name: string
  state_group: string
  priority: string
  assignees?: Array<{ id: number; display_name?: string; username?: string }>
  target_date?: string | null
  issue_type?: IssueType
}

interface SubIssue {
  id: number
  sequence_id: number
  name: string
  state_name: string
  state_group: string
  priority: string
  assignees?: Array<{ id: number; display_name?: string; username?: string }>
  target_date?: string | null
  issue_type?: IssueType
}

interface RelationItem {
  id: number
  relation_type_id: number
  relation_type: {
    id: number
    name: string
    outward_name: string
  }
  related_issue_id: number
  related_issue: {
    id: number
    sequence_id: number
    name: string
    state_name: string
    state_group: string
    priority: string
    assignees?: Array<{ id: number; display_name?: string; username?: string }>
    start_date?: string | null
    target_date?: string | null
    issue_type?: IssueType
  }
}

interface RelationGroup {
  typeId: number
  typeName: string
  items: RelationItem[]
  color: string
}

const props = defineProps<{
  issueId: number
  projectId: number
  workspaceId: number
  parent: ParentIssue | null
  subIssues: SubIssue[]
  issueTypes: IssueType[]
}>()

defineEmits<{
  navigate: [issueId: number]
  changeParent: []
  addRelation: [relationTypeId: number]
}>()

const relations = ref<RelationItem[]>([])
const relationTypes = ref<any[]>([])
const loading = ref(false)

const relationGroups = computed<RelationGroup[]>(() => {
  const groups: Map<number, RelationGroup> = new Map()
  let colorIndex = 0

  for (const rel of relations.value) {
    if (!groups.has(rel.relation_type_id)) {
      groups.set(rel.relation_type_id, {
        typeId: rel.relation_type_id,
        typeName: rel.relation_type.outward_name || rel.relation_type.name,
        items: [],
        color: COLORS[colorIndex % COLORS.length],
      })
      colorIndex++
    }
    groups.get(rel.relation_type_id)!.items.push(rel)
  }

  return Array.from(groups.values())
})

onMounted(async () => {
  loading.value = true
  try {
    const [rels, types] = await Promise.all([
      listIssueRelations(props.issueId),
      listRelationTypes(props.workspaceId),
    ])
    relations.value = rels || []
    relationTypes.value = types || []
  } catch (err) {
    console.error('Failed to load relations:', err)
  } finally {
    loading.value = false
  }
})

async function removeRelation(relationId: number) {
  try {
    await deleteIssueRelation(relationId)
    relations.value = relations.value.filter((r) => r.id !== relationId)
  } catch (err) {
    console.error('Failed to remove relation:', err)
  }
}

async function removeParent() {
  try {
    await issueApi.updateIssue(props.issueId, { parent_id: undefined } as any)
  } catch (err) {
    console.error('Failed to remove parent:', err)
  }
}

async function toggleSubIssue(subIssueId: number) {
  try {
    const issue = props.subIssues.find((s) => s.id === subIssueId)
    if (issue) {
      const newStateGroup = issue.state_group === 'done' ? 'todo' : 'done'
      await issueApi.updateIssue(subIssueId, { state_group: newStateGroup } as any)
    }
  } catch (err) {
    console.error('Failed to toggle sub-issue:', err)
  }
}

function createSubIssue() {
  router.push({ name: 'create-issue', query: { parent_id: props.issueId } })
}
</script>
