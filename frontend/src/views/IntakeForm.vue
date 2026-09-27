<template>
  <div class="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4">
    <div class="bg-white rounded-xl shadow-lg p-8 w-full max-w-lg" data-testid="intake-form">
      <div class="text-center mb-6">
        <h1 class="text-2xl font-bold text-gray-900" data-testid="intake-form-project">{{ projectName || t('intakeForm.title') }}</h1>
        <p class="text-sm text-gray-500 mt-1">{{ t('intakeForm.subtitle') }}</p>
      </div>

      <div v-if="notFound" class="text-center py-8 text-sm text-gray-500">{{ t('intakeForm.notFound') }}</div>
      <div v-else-if="disabled" class="text-center py-8 text-sm text-gray-500" data-testid="intake-form-disabled">{{ t('intakeForm.disabled') }}</div>

      <div v-else-if="submitted" class="text-center py-8" data-testid="intake-form-done">
        <div class="text-4xl mb-3">✅</div>
        <h2 class="text-lg font-semibold text-gray-800">{{ t('intakeForm.doneTitle') }}</h2>
        <p class="text-sm text-gray-500 mt-1">{{ t('intakeForm.doneDesc', { n: String(submittedSeq) }) }}</p>
        <button class="mt-4 text-sm text-indigo-600 hover:underline" @click="reset">{{ t('intakeForm.another') }}</button>
      </div>

      <form v-else class="space-y-4" @submit.prevent="submit">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('intakeForm.name') }} *</label>
          <input v-model="form.name" name="name" maxlength="255" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" :placeholder="t('intakeForm.namePh')" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('intakeForm.description') }}</label>
          <textarea v-model="form.description" name="description" rows="4" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm resize-none" :placeholder="t('intakeForm.descriptionPh')"></textarea>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('intake.yourName') }}</label>
            <input v-model="form.submitter" name="submitter" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" :placeholder="t('intakeForm.optional')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('intakeForm.email') }}</label>
            <input v-model="form.email" name="email" type="email" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" :placeholder="t('intakeForm.optional')" />
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('intakeForm.priority') }}</label>
          <select v-model="form.priority" name="priority" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
            <option v-for="p in priorities" :key="p" :value="p">{{ t('intakeForm.p.' + p) }}</option>
          </select>
        </div>
        <button type="submit" :disabled="submitting" class="w-full py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 font-medium" data-testid="intake-form-submit">
          {{ submitting ? t('intakeForm.submitting') : t('intakeForm.submit') }}
        </button>
        <p v-if="error" class="text-red-500 text-xs text-center" data-testid="intake-form-error">{{ error }}</p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { intakeApi } from '@/api/intake'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()
const route = useRoute()
const priorities = ['none', 'low', 'medium', 'high', 'urgent']
const projectId = ref(0)
const projectName = ref('')
const notFound = ref(false)
const disabled = ref(false)
const submitted = ref(false)
const submittedSeq = ref(0)
const submitting = ref(false)
const error = ref('')
const form = reactive({ name: '', description: '', priority: 'none', submitter: '', email: '' })

onMounted(async () => {
  projectId.value = parseInt(String(route.params.projectId), 10)
  try {
    const info = await intakeApi.info(projectId.value)
    projectName.value = info.name
    disabled.value = !info.form_enabled
  } catch {
    notFound.value = true
  }
})

function reset() {
  Object.assign(form, { name: '', description: '', priority: 'none' })
  submitted.value = false
}

async function submit() {
  if (!form.name.trim()) {
    error.value = t('intakeForm.nameRequired')
    return
  }
  if (form.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
    error.value = t('intakeForm.emailInvalid')
    return
  }
  submitting.value = true
  error.value = ''
  try {
    const res = await intakeApi.submit(projectId.value, { ...form })
    submittedSeq.value = res.sequence_id
    submitted.value = true
  } catch (e: any) {
    error.value = e?.response?.data?.message || t('intakeForm.failed')
  } finally {
    submitting.value = false
  }
}
</script>
