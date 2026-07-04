<template>
  <div class="tiptap-editor" :class="{ 'is-locked': locked }">
    <!-- Toolbar -->
    <div v-if="editor && !locked" class="tiptap-toolbar">
      <div class="toolbar-group">
        <button @click="editor.chain().focus().toggleBold().run()" :class="{ 'is-active': editor.isActive('bold') }" :title="t('editor.boldShortcut')">
          <strong>B</strong>
        </button>
        <button @click="editor.chain().focus().toggleItalic().run()" :class="{ 'is-active': editor.isActive('italic') }" :title="t('editor.italicShortcut')">
          <em>I</em>
        </button>
        <button @click="editor.chain().focus().toggleUnderline().run()" :class="{ 'is-active': editor.isActive('underline') }" :title="t('editor.underlineShortcut')">
          <u>U</u>
        </button>
        <button @click="editor.chain().focus().toggleStrike().run()" :class="{ 'is-active': editor.isActive('strike') }" :title="t('editor.strikethrough')">
          <s>S</s>
        </button>
        <button @click="editor.chain().focus().toggleHighlight().run()" :class="{ 'is-active': editor.isActive('highlight') }" :title="t('editor.highlight')">
          <span class="highlight-icon">H</span>
        </button>
        <button @click="editor.chain().focus().toggleCode().run()" :class="{ 'is-active': editor.isActive('code') }" :title="t('editor.inlineCode')">
          <code>&lt;/&gt;</code>
        </button>
      </div>
      <span class="toolbar-divider"></span>
      <div class="toolbar-group">
        <button @click="editor.chain().focus().toggleHeading({ level: 1 }).run()" :class="{ 'is-active': editor.isActive('heading', { level: 1 }) }" :title="t('editor.heading1')">
          H1
        </button>
        <button @click="editor.chain().focus().toggleHeading({ level: 2 }).run()" :class="{ 'is-active': editor.isActive('heading', { level: 2 }) }" :title="t('editor.heading2')">
          H2
        </button>
        <button @click="editor.chain().focus().toggleHeading({ level: 3 }).run()" :class="{ 'is-active': editor.isActive('heading', { level: 3 }) }" :title="t('editor.heading3')">
          H3
        </button>
      </div>
      <span class="toolbar-divider"></span>
      <div class="toolbar-group">
        <button @click="editor.chain().focus().toggleBulletList().run()" :class="{ 'is-active': editor.isActive('bulletList') }" :title="t('editor.bulletList')">
          •≡
        </button>
        <button @click="editor.chain().focus().toggleOrderedList().run()" :class="{ 'is-active': editor.isActive('orderedList') }" :title="t('editor.orderedList')">
          1.≡
        </button>
        <button @click="editor.chain().focus().toggleTaskList().run()" :class="{ 'is-active': editor.isActive('taskList') }" :title="t('editor.taskList')">
          ☑
        </button>
        <button @click="editor.chain().focus().toggleBlockquote().run()" :class="{ 'is-active': editor.isActive('blockquote') }" :title="t('editor.blockquote')">
          ❝
        </button>
        <button @click="editor.chain().focus().toggleCodeBlock().run()" :class="{ 'is-active': editor.isActive('codeBlock') }" :title="t('editor.codeBlock')">
          { }
        </button>
      </div>
      <span class="toolbar-divider"></span>
      <div class="toolbar-group">
        <button @click="editor.chain().focus().setTextAlign('left').run()" :class="{ 'is-active': editor.isActive({ textAlign: 'left' }) }" :title="t('editor.alignLeft')">
          ═╪
        </button>
        <button @click="editor.chain().focus().setTextAlign('center').run()" :class="{ 'is-active': editor.isActive({ textAlign: 'center' }) }" :title="t('editor.alignCenter')">
          ═╪═
        </button>
        <button @click="editor.chain().focus().setTextAlign('right').run()" :class="{ 'is-active': editor.isActive({ textAlign: 'right' }) }" :title="t('editor.alignRight')">
          ╪═
        </button>
      </div>
      <span class="toolbar-divider"></span>
      <div class="toolbar-group">
        <button @click="addLink" :class="{ 'is-active': editor.isActive('link') }" :title="t('editor.addLink')">
          🔗
        </button>
        <button @click="addImage" :title="t('editor.addImage')">
          🖼
        </button>
        <button @click="editor.chain().focus().setHorizontalRule().run()" :title="t('editor.horizontalRule')">
          —
        </button>
      </div>
    </div>

    <!-- Editor Area -->
    <div class="tiptap-content" @click="focusEditor">
      <editor-content :editor="editor" />
    </div>

    <!-- Lock Banner -->
    <div v-if="locked" class="lock-banner">
      <span>🔒 {{ t('pages.lockedBy', { name: lockedByName || t('pages.anotherUser') }) }}</span>
    </div>

    <!-- Character Count -->
    <div v-if="editor && showCharCount" class="char-count">
      {{ (editor.storage.characterCount as any).charCount() }} {{ t('pages.characters') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import Highlight from '@tiptap/extension-highlight'
import TextAlign from '@tiptap/extension-text-align'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import Placeholder from '@tiptap/extension-placeholder'
import CharacterCount from '@tiptap/extension-character-count'
import HorizontalRule from '@tiptap/extension-horizontal-rule'
import Typography from '@tiptap/extension-typography'
import { watch, onBeforeUnmount } from 'vue'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

const props = defineProps<{
  modelValue?: string
  contentJson?: string | object
  placeholder?: string
  locked?: boolean
  lockedByName?: string
  showCharCount?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'update:contentJson', value: string): void
  (e: 'focus'): void
}>()

// Parse initial content
function getInitialContent(): any {
  if (props.contentJson) {
    try {
      const parsed = typeof props.contentJson === 'string'
        ? JSON.parse(props.contentJson)
        : props.contentJson
      if (parsed && typeof parsed === 'object') return parsed
    } catch { /* use HTML */ }
  }
  if (props.modelValue) {
    return props.modelValue
  }
  return ''
}

const editor = useEditor({
  content: getInitialContent(),
  extensions: [
    StarterKit.configure({
      heading: { levels: [1, 2, 3] },
    }),
    Underline,
    Highlight,
    TextAlign.configure({ types: ['heading', 'paragraph'] }),
    Link.configure({ openOnClick: false, HTMLAttributes: { class: 'tiptap-link' } }),
    Image.configure({ inline: true, HTMLAttributes: { class: 'tiptap-image' } }),
    TaskList,
    TaskItem.configure({ nested: true }),
    Placeholder.configure({ placeholder: props.placeholder || t('pages.placeholder') }),
    CharacterCount,
    HorizontalRule,
    Typography,
  ],
  editable: !props.locked,
  onUpdate: () => {
    if (!editor.value) return
    const html = editor.value.getHTML()
    const json = editor.value.getJSON()
    emit('update:modelValue', html)
    emit('update:contentJson', JSON.stringify(json))
  },
})

watch(() => props.modelValue, (val) => {
  if (!editor.value || val === null || val === undefined) return
  const currentHtml = editor.value.getHTML()
  if (val !== currentHtml) {
    // Only reset if external change is different
    const isEmpty = val === '' || val === '<p></p>'
    const currentEmpty = currentHtml === '' || currentHtml === '<p></p>'
    if (isEmpty && currentEmpty) return
    editor.value.commands.setContent(val)
  }
})

watch(() => props.contentJson, (val) => {
  if (!editor.value || !val) return
  try {
    const parsed = typeof val === 'string' ? JSON.parse(val) : val
    if (parsed && typeof parsed === 'object' && parsed.type) {
      const currentJson = editor.value.getJSON()
      if (JSON.stringify(parsed) !== JSON.stringify(currentJson)) {
        editor.value.commands.setContent(parsed)
      }
    }
  } catch { /* ignore */ }
})

watch(() => props.locked, (val) => {
  editor.value?.setEditable(!val)
})

function focusEditor() {
  editor.value?.commands.focus()
  emit('focus')
}

function addLink() {
  if (!editor.value) return
  const url = window.prompt('URL:')
  if (url) {
    editor.value.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
  }
}

function addImage() {
  if (!editor.value) return
  const url = window.prompt('Image URL:')
  if (url) {
    editor.value.chain().focus().setImage({ src: url }).run()
  }
}

onBeforeUnmount(() => {
  editor.value?.destroy()
})
</script>

<style>
/* Editor content styles (unscoped for editor-content rendering) */
.tiptap-editor {
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  background: #fff;
  overflow: hidden;
  min-height: 400px;
  display: flex;
  flex-direction: column;
}

.tiptap-editor.is-locked {
  opacity: 0.9;
}

.tiptap-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  padding: 0.375rem 0.5rem;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
  gap: 0;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 2px;
}

.toolbar-group button {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 1.75rem;
  height: 1.75rem;
  padding: 0 0.375rem;
  border: 1px solid transparent;
  background: transparent;
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 0.8125rem;
  color: #4b5563;
  transition: all 0.15s;
}

.toolbar-group button:hover {
  background: #e5e7eb;
  color: #1f2937;
}

.toolbar-group button.is-active {
  background: #e0e7ff;
  color: #4338ca;
  border-color: #c7d2fe;
}

.toolbar-divider {
  width: 1px;
  height: 1.25rem;
  background: #d1d5db;
  margin: 0 0.375rem;
}

.tiptap-content {
  flex: 1;
  min-height: 400px;
  cursor: text;
}

.tiptap-content .ProseMirror {
  padding: 1rem 1.25rem;
  min-height: 400px;
  outline: none;
  font-size: 0.9375rem;
  line-height: 1.75;
  color: #374151;
}

.tiptap-content .ProseMirror p.is-editor-empty:first-child::before {
  content: attr(data-placeholder);
  float: left;
  color: #9ca3af;
  pointer-events: none;
  height: 0;
}

.tiptap-content .ProseMirror h1 { font-size: 1.75rem; font-weight: 700; margin: 1rem 0 0.5rem; color: #111827; }
.tiptap-content .ProseMirror h2 { font-size: 1.375rem; font-weight: 600; margin: 0.875rem 0 0.5rem; color: #1f2937; }
.tiptap-content .ProseMirror h3 { font-size: 1.125rem; font-weight: 600; margin: 0.75rem 0 0.375rem; color: #374151; }

.tiptap-content .ProseMirror p { margin: 0 0 0.5rem 0; }

.tiptap-content .ProseMirror ul,
.tiptap-content .ProseMirror ol { padding-left: 1.5rem; margin: 0.5rem 0; }

.tiptap-content .ProseMirror li { margin: 0.125rem 0; }

.tiptap-content .ProseMirror ul[data-type="taskList"] {
  list-style: none;
  padding-left: 0;
}

.tiptap-content .ProseMirror ul[data-type="taskList"] li {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
}

.tiptap-content .ProseMirror ul[data-type="taskList"] li label {
  margin-top: 0.25rem;
}

.tiptap-content .ProseMirror blockquote {
  border-left: 3px solid #d1d5db;
  padding-left: 1rem;
  margin: 0.5rem 0;
  color: #6b7280;
  font-style: italic;
}

.tiptap-content .ProseMirror pre {
  background: #1f2937;
  color: #f3f4f6;
  font-family: 'Fira Code', 'Cascadia Code', monospace;
  padding: 0.75rem 1rem;
  border-radius: 0.375rem;
  margin: 0.5rem 0;
  overflow-x: auto;
  font-size: 0.8125rem;
  line-height: 1.5;
}

.tiptap-content .ProseMirror pre code { background: none; color: inherit; padding: 0; font-size: inherit; }

.tiptap-content .ProseMirror code {
  background: #f3f4f6;
  color: #be185d;
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  font-size: 0.875em;
  font-family: 'Fira Code', 'Cascadia Code', monospace;
}

.tiptap-content .ProseMirror mark {
  background: #fef08a;
  border-radius: 0.125rem;
  padding: 0.125rem 0.25rem;
  color: inherit;
}

.tiptap-content .ProseMirror hr {
  border: none;
  border-top: 2px solid #e5e7eb;
  margin: 1rem 0;
}

.tiptap-content .ProseMirror a.tiptap-link {
  color: #4f46e5;
  text-decoration: underline;
  cursor: pointer;
}

.tiptap-content .ProseMirror a.tiptap-link:hover { color: #4338ca; }

.tiptap-content .ProseMirror img.tiptap-image {
  max-width: 100%;
  height: auto;
  border-radius: 0.375rem;
  margin: 0.5rem 0;
}

.tiptap-content .ProseMirror img.tiptap-image.ProseMirror-selectednode {
  outline: 2px solid #6366f1;
  outline-offset: 2px;
}

.lock-banner {
  background: #fef3c7;
  border-top: 1px solid #fcd34d;
  color: #92400e;
  padding: 0.5rem 1rem;
  font-size: 0.8125rem;
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.char-count {
  padding: 0.25rem 1rem;
  text-align: right;
  font-size: 0.75rem;
  color: #9ca3af;
  border-top: 1px solid #f3f4f6;
}
</style>
