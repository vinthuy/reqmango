<template>
  <div class="custom-fields-page min-h-screen bg-gray-50">
    <!-- 头部 -->
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <button @click="goBack" class="text-gray-500 hover:text-gray-700">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1 class="text-lg font-semibold text-gray-900">{{ t('customFieldsPage.title') }}</h1>
        </div>
      </div>
    </div>

    <!-- 主内容 -->
    <div class="max-w-4xl mx-auto px-6 py-6">
      <div class="bg-white rounded-lg shadow-sm p-6">
        <CustomFieldManager
          :workspace-id="workspaceId"
          :project-id="projectId"
          mode="manage"
          @field-created="onFieldCreated"
          @field-updated="onFieldUpdated"
          @field-deleted="onFieldDeleted"
        />
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import CustomFieldManager from '@/components/CustomFieldManager.vue'
import type { CustomField } from '@/types/custom-field'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const workspaceId = parseInt(route.params.workspaceId as string, 10)
const projectId = parseInt(route.params.projectId as string, 10)

function goBack() {
  router.back()
}

function onFieldCreated(field: CustomField) {
  console.log('Field created:', field)
}

function onFieldUpdated(field: CustomField) {
  console.log('Field updated:', field)
}

function onFieldDeleted(fieldId: number) {
  console.log('Field deleted:', fieldId)
}
</script>

<style scoped>
.custom-fields-page {
  min-height: 100vh;
}
</style>
