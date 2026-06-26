<template>
  <div class="rich-text-editor">
    <div class="toolbar">
      <button @click="formatText('bold')" class="toolbar-btn" title="加粗">
        <strong>B</strong>
      </button>
      <button @click="formatText('italic')" class="toolbar-btn" title="斜体">
        <em>I</em>
      </button>
      <button @click="formatText('underline')" class="toolbar-btn" title="下划线">
        <u>U</u>
      </button>
      <button @click="formatText('strikeThrough')" class="toolbar-btn" title="删除线">
        <s>S</s>
      </button>
      <span class="toolbar-divider"></span>
      <button @click="formatText('insertUnorderedList')" class="toolbar-btn" title="无序列表">
        ☰
      </button>
      <button @click="formatText('insertOrderedList')" class="toolbar-btn" title="有序列表">
        ≡
      </button>
      <span class="toolbar-divider"></span>
      <button @click="formatText('justifyLeft')" class="toolbar-btn" title="左对齐">
        ⌈
      </button>
      <button @click="formatText('justifyCenter')" class="toolbar-btn" title="居中">
        ⌉
      </button>
      <button @click="formatText('justifyRight')" class="toolbar-btn" title="右对齐">
        ⌊
      </button>
    </div>
    <div
      ref="editorRef"
      contenteditable="true"
      class="editor-content"
      :placeholder="placeholder"
      @input="handleInput"
      @paste="handlePaste"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

const props = defineProps<{
  modelValue?: string
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const editorRef = ref<HTMLElement | null>(null)

onMounted(() => {
  if (editorRef.value && props.modelValue) {
    editorRef.value.innerHTML = props.modelValue
  }
})

watch(() => props.modelValue, (val) => {
  if (editorRef.value && val !== editorRef.value.innerHTML) {
    editorRef.value.innerHTML = val || ''
  }
})

function handleInput() {
  if (editorRef.value) {
    emit('update:modelValue', editorRef.value.innerHTML)
  }
}

function handlePaste(e: ClipboardEvent) {
  e.preventDefault()
  const text = e.clipboardData?.getData('text/html') || e.clipboardData?.getData('text/plain') || ''
  document.execCommand('insertHTML', false, text)
  handleInput()
}

function formatText(command: string) {
  document.execCommand(command, false)
  editorRef.value?.focus()
  handleInput()
}
</script>

<style scoped>
.rich-text-editor {
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  overflow: hidden;
}

.toolbar {
  display: flex;
  align-items: center;
  padding: 0.5rem;
  background-color: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  margin: 0 0.25rem;
  border: none;
  background-color: transparent;
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 0.875rem;
  color: #4b5563;
}

.toolbar-btn:hover {
  background-color: #e5e7eb;
}

.toolbar-divider {
  width: 1px;
  height: 1.5rem;
  background-color: #e5e7eb;
  margin: 0 0.25rem;
}

.editor-content {
  min-height: 150px;
  padding: 0.75rem;
  outline: none;
  font-family: inherit;
  font-size: inherit;
  color: inherit;
}

.editor-content:empty:before {
  content: attr(placeholder);
  color: #9ca3af;
  pointer-events: none;
}

.editor-content:focus {
  outline: none;
}

.editor-content p {
  margin: 0 0 0.5rem 0;
}

.editor-content ul,
.editor-content ol {
  margin: 0.5rem 0;
  padding-left: 1.5rem;
}

.editor-content li {
  margin: 0.25rem 0;
}
</style>
