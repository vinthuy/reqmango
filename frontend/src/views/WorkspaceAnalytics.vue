<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { workspaceApi } from '@/api/workspace'
import { projectApi } from '@/api/project'

const route = useRoute()
const router = useRouter()

onMounted(async () => {
  const slug = route.params.slug as string
  try {
    const ws = await workspaceApi.getBySlug(slug)
    const projects = await projectApi.listProjects(ws.id)
    if (projects.length > 0) {
      router.replace(`/workspace/${slug}/project/${projects[0].id}/analytics`)
    } else {
      router.replace(`/workspace/${slug}`)
    }
  } catch {
    router.replace(`/workspace/${slug}`)
  }
})
</script>

<template>
  <div class="flex items-center justify-center h-64 text-gray-400">跳转中...</div>
</template>
