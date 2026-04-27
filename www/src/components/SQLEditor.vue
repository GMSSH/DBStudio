<template>
  <div ref="editorContainer" class="sql-editor" :class="{ 'full-height': fullHeight }"></div>
</template>

<script setup>
import { ref, onMounted, watch, onBeforeUnmount } from 'vue'
import { EditorView, Decoration, keymap } from '@codemirror/view'
import { basicSetup } from 'codemirror'
import { Compartment, StateField } from '@codemirror/state'
import { sql, MySQL, PostgreSQL, SQLite } from '@codemirror/lang-sql'
import { oneDark } from '@codemirror/theme-one-dark'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  fullHeight: {
    type: Boolean,
    default: false
  },
  dbType: {
    type: String,
    default: 'mysql'
  },
  completionSchema: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['update:modelValue', 'execute'])

const editorContainer = ref(null)
const languageCompartment = new Compartment()
let editorView = null

function normalizeRange(text, from, to) {
  let start = from
  let end = to

  while (start < end && /\s/.test(text[start])) start += 1
  while (end > start && /\s/.test(text[end - 1])) end -= 1

  return {
    from: start,
    to: end,
    text: text.slice(start, end)
  }
}

function splitStatements(text) {
  const segments = []
  let start = 0
  let inSingle = false
  let inDouble = false
  let inBacktick = false
  let inLineComment = false
  let inBlockComment = false

  for (let i = 0; i < text.length; i += 1) {
    const char = text[i]
    const next = text[i + 1]

    if (inLineComment) {
      if (char === '\n') inLineComment = false
      continue
    }

    if (inBlockComment) {
      if (char === '*' && next === '/') {
        inBlockComment = false
        i += 1
      }
      continue
    }

    if (!inSingle && !inDouble && !inBacktick) {
      if (char === '-' && next === '-') {
        inLineComment = true
        i += 1
        continue
      }
      if (char === '/' && next === '*') {
        inBlockComment = true
        i += 1
        continue
      }
    }

    if (char === '\'' && !inDouble && !inBacktick) {
      if (inSingle && next === '\'') {
        i += 1
        continue
      }
      inSingle = !inSingle
      continue
    }

    if (char === '"' && !inSingle && !inBacktick) {
      if (inDouble && next === '"') {
        i += 1
        continue
      }
      inDouble = !inDouble
      continue
    }

    if (char === '`' && !inSingle && !inDouble) {
      inBacktick = !inBacktick
      continue
    }

    if (char === ';' && !inSingle && !inDouble && !inBacktick) {
      segments.push(normalizeRange(text, start, i))
      start = i + 1
    }
  }

  segments.push(normalizeRange(text, start, text.length))
  return segments.filter((segment) => segment.to > segment.from)
}

function getCurrentStatement(text, position) {
  const statements = splitStatements(text)

  if (!statements.length) {
    return normalizeRange(text, 0, text.length)
  }

  return (
    statements.find((statement) => position >= statement.from && position <= statement.to) ||
    statements.find((statement) => position < statement.from) ||
    statements[statements.length - 1]
  )
}

const activeStatementField = StateField.define({
  create(state) {
    const doc = state.doc.toString()
    const statement = getCurrentStatement(doc, state.selection.main.head)
    if (!statement.text.trim()) return Decoration.none
    return Decoration.set([
      Decoration.mark({ class: 'cm-active-statement' }).range(statement.from, statement.to)
    ])
  },
  update(_, transaction) {
    if (!transaction.docChanged && !transaction.selection) {
      return _
    }

    const doc = transaction.state.doc.toString()
    const statement = getCurrentStatement(doc, transaction.state.selection.main.head)
    if (!statement.text.trim()) return Decoration.none
    return Decoration.set([
      Decoration.mark({ class: 'cm-active-statement' }).range(statement.from, statement.to)
    ])
  },
  provide: (field) => EditorView.decorations.from(field)
})

const completionTheme = EditorView.theme({
  '.cm-tooltip.cm-tooltip-autocomplete': {
    overflow: 'hidden',
    minWidth: '260px',
    maxWidth: '460px',
    border: '1px solid rgba(255, 255, 255, 0.12)',
    borderRadius: '12px',
    background: 'rgba(24, 28, 40, 0.98)',
    boxShadow: '0 18px 50px rgba(0, 0, 0, 0.42), 0 0 0 1px rgba(255, 255, 255, 0.04) inset',
    backdropFilter: 'blur(18px)',
    color: 'var(--sys-color-text-primary)',
    fontFamily: 'var(--ref-font-family-base)',
    padding: '6px',
    zIndex: '20'
  },
  '.cm-tooltip.cm-tooltip-autocomplete > ul': {
    maxHeight: '300px',
    padding: '2px',
    fontFamily: 'var(--ref-font-family-base)',
    scrollbarWidth: 'thin',
    scrollbarColor: 'rgba(255, 255, 255, 0.18) transparent'
  },
  '.cm-tooltip.cm-tooltip-autocomplete > ul::-webkit-scrollbar': {
    width: '6px'
  },
  '.cm-tooltip.cm-tooltip-autocomplete > ul::-webkit-scrollbar-track': {
    background: 'transparent'
  },
  '.cm-tooltip.cm-tooltip-autocomplete > ul::-webkit-scrollbar-thumb': {
    borderRadius: '999px',
    background: 'rgba(255, 255, 255, 0.18)'
  },
  '.cm-tooltip.cm-tooltip-autocomplete > ul > li': {
    position: 'relative',
    minHeight: '34px',
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    padding: '7px 10px 7px 34px',
    borderRadius: '9px',
    color: 'var(--sys-color-text-secondary)',
    fontSize: '13px',
    lineHeight: '20px',
    cursor: 'pointer',
    transition: 'background-color 0.14s ease, color 0.14s ease, transform 0.14s ease',
    background: 'transparent'
  },
  '.cm-tooltip.cm-tooltip-autocomplete > ul > li:hover': {
    background: 'rgba(255, 255, 255, 0.07)',
    color: 'var(--sys-color-text-title)'
  },
  '.cm-tooltip.cm-tooltip-autocomplete > ul > li[aria-selected]': {
    background: 'rgba(87, 114, 255, 0.16)',
    color: 'var(--sys-color-text-title)',
    boxShadow: 'inset 3px 0 0 var(--ref-color-brand-6)'
  },
  '.cm-completionIcon': {
    position: 'absolute',
    left: '10px',
    top: '50%',
    width: '16px',
    height: '16px',
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    transform: 'translateY(-50%)',
    opacity: '1',
    color: 'var(--ref-color-cyan-6)',
    fontSize: '0'
  },
  '.cm-completionIcon::after': {
    content: '""',
    width: '7px',
    height: '7px',
    borderRadius: '999px',
    background: 'currentColor',
    boxShadow: '0 0 10px currentColor'
  },
  '.cm-completionLabel': {
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    fontFamily: 'var(--ref-font-family-mono)',
    fontWeight: '500'
  },
  '.cm-completionMatchedText': {
    color: 'var(--ref-color-brand-4)',
    textDecoration: 'none',
    fontWeight: '700'
  },
  '.cm-completionDetail': {
    marginLeft: 'auto',
    maxWidth: '150px',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    color: 'var(--sys-color-text-tertiary)',
    fontSize: '12px',
    fontFamily: 'var(--ref-font-family-base)'
  },
  '.cm-tooltip.cm-completionInfo': {
    maxWidth: '360px',
    border: '1px solid rgba(255, 255, 255, 0.12)',
    borderRadius: '10px',
    background: 'rgba(22, 24, 34, 0.98)',
    color: 'var(--sys-color-text-secondary)',
    boxShadow: '0 18px 50px rgba(0, 0, 0, 0.38)',
    padding: '10px 12px',
    fontSize: '12px',
    lineHeight: '1.6',
    fontFamily: 'var(--ref-font-family-base)'
  }
}, { dark: true })

function getDialect() {
  if (props.dbType === 'postgres') return PostgreSQL
  if (props.dbType === 'sqlite') return SQLite
  return MySQL
}

function normalizeCompletionSchemaValue(value) {
  if (Array.isArray(value)) {
    return value.filter((item) => typeof item === 'string' || (item && typeof item.label === 'string'))
  }

  if (!value || typeof value !== 'object') {
    return []
  }

  if (value.self && typeof value.self.label === 'string' && value.children && typeof value.children === 'object') {
    return {
      self: value.self,
      children: normalizeCompletionSchemaValue(value.children)
    }
  }

  const normalized = {}
  for (const [key, child] of Object.entries(value)) {
    if (!key) continue

    if (Array.isArray(child)) {
      normalized[key] = normalizeCompletionSchemaValue(child)
      continue
    }

    if (child && typeof child === 'object') {
      normalized[key] = normalizeCompletionSchemaValue(child)
      continue
    }

    if (typeof child === 'string') {
      normalized[key] = [child]
    }
  }
  return normalized
}

function createSqlExtension() {
  return sql({
    dialect: getDialect(),
    schema: normalizeCompletionSchemaValue(props.completionSchema || {})
  })
}

function focus() {
  editorView?.focus()
}

function getSelectedText() {
  if (!editorView) return ''
  const { from, to } = editorView.state.selection.main
  if (from === to) return ''
  return editorView.state.sliceDoc(from, to)
}

function getExecutionSQL() {
  if (!editorView) return ''

  const selected = getSelectedText()
  if (selected.trim()) {
    return selected.trim()
  }

  const doc = editorView.state.doc.toString()
  const statement = getCurrentStatement(doc, editorView.state.selection.main.head)
  return (statement.text || doc).trim()
}

function insertText(text) {
  if (!editorView) return

  const { from, to } = editorView.state.selection.main
  editorView.dispatch({
    changes: { from, to, insert: text },
    selection: { anchor: from + text.length }
  })
  focus()
}

onMounted(() => {
  editorView = new EditorView({
    doc: props.modelValue,
    extensions: [
      basicSetup,
      languageCompartment.of(createSqlExtension()),
      oneDark,
      activeStatementField,
      completionTheme,
      keymap.of([
        {
          key: 'F8',
          run: () => {
            emit('execute')
            return true
          }
        },
        {
          key: 'Mod-Enter',
          run: () => {
            emit('execute')
            return true
          }
        }
      ]),
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          emit('update:modelValue', update.state.doc.toString())
        }
      })
    ],
    parent: editorContainer.value
  })

  focus()
})

watch(() => props.modelValue, (newValue) => {
  if (!editorView) return

  const currentValue = editorView.state.doc.toString()
  if (newValue !== currentValue) {
    editorView.dispatch({
      changes: { from: 0, to: currentValue.length, insert: newValue }
    })
  }
})

watch(
  () => JSON.stringify({ dbType: props.dbType, schema: props.completionSchema || {} }),
  () => {
    if (!editorView) return
    editorView.dispatch({
      effects: languageCompartment.reconfigure(createSqlExtension())
    })
  }
)

onBeforeUnmount(() => {
  editorView?.destroy()
})

defineExpose({
  focus,
  getSelectedText,
  getExecutionSQL,
  insertText
})
</script>

<style scoped>
.sql-editor {
  overflow: hidden;
  font-size: var(--ref-font-size-md);
}

/* 覆盖 oneDark 主题的 #282c34 背景，改为透明 */
.sql-editor :deep(.cm-editor),
.sql-editor :deep(.cm-gutters),
.sql-editor :deep(.cm-gutters-before),
.sql-editor :deep(.cm-gutter),
.sql-editor :deep(.cm-gutterElement),
.sql-editor :deep(.cm-activeLineGutter) {
  background: transparent !important;
}

.sql-editor.full-height {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.sql-editor :deep(.cm-editor) {
  height: 300px;
}

.sql-editor.full-height :deep(.cm-editor) {
  height: 100%;
  flex: 1;
}

.sql-editor :deep(.cm-scroller) {
  overflow: auto;
  font-family: var(--ref-font-family-mono);
}

.sql-editor :deep(.cm-active-statement) {
  background: rgba(87, 114, 255, 0.08);
}

.sql-editor :deep(.cm-scroller::-webkit-scrollbar) {
  width: var(--scroll-size-default);
  height: var(--scroll-size-default);
}

.sql-editor :deep(.cm-scroller::-webkit-scrollbar-track) {
  background: var(--scroll-track-bg);
  border-radius: var(--scroll-track-radius);
}

.sql-editor :deep(.cm-scroller::-webkit-scrollbar-thumb) {
  background: var(--scroll-thumb-bg);
  border-radius: var(--scroll-thumb-radius);
}

.sql-editor :deep(.cm-scroller::-webkit-scrollbar-thumb:hover) {
  background: var(--scroll-thumb-bg-hover);
}
</style>
