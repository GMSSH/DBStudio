<template>
  <FormModal
    v-if="mode === 'export'"
    v-model="visible"
    :title="$t('ie.exportTitle', { table: tableName })"
    class="ie-modal"
    :width="500"
  >
    <div class="ie-body">
      <!-- Format Cards -->
      <div class="ie-section-label">{{ $t('ie.format') }}</div>
      <div class="format-cards">
        <button
          v-for="fmt in formatOptions"
          :key="fmt.value"
          class="format-card"
          :class="{ 'format-card--active': exportOptions.format === fmt.value }"
          @click="exportOptions.format = fmt.value"
          type="button"
        >
          <span class="format-card__icon" v-html="fmt.icon" />
          <span class="format-card__name">{{ fmt.label }}</span>
          <span class="format-card__desc">{{ $t(fmt.descKey) }}</span>
        </button>
      </div>

      <!-- Format tip -->
      <div class="format-tip">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        {{ $t(currentFormatTip) }}
      </div>

      <!-- Range -->
      <div class="ie-section-label" style="margin-top: 4px;">{{ $t('ie.range') }}</div>
      <div class="range-options">
        <label v-if="hasCurrentPageExport" class="range-option" :class="{ 'range-option--active': exportOptions.range === 'current' }">
          <input type="radio" v-model="exportOptions.range" value="current" />
          <span class="range-option__text">
            {{ $t('ie.currentPage') }}
            <em>{{ pageRows.length }} {{ $t('ie.rows') }}</em>
          </span>
        </label>
        <label class="range-option" :class="{ 'range-option--active': exportOptions.range === 'all' }">
          <input type="radio" v-model="exportOptions.range" value="all" />
          <span class="range-option__text">
            {{ $t('ie.allRows') }}
            <em v-if="totalRows > 0">≈ {{ totalRows.toLocaleString() }} {{ $t('ie.rows') }}</em>
          </span>
        </label>
        <label class="range-option" :class="{ 'range-option--active': exportOptions.range === 'custom' }">
          <input type="radio" v-model="exportOptions.range" value="custom" />
          <span class="range-option__text">{{ $t('ie.customRows') }}</span>
        </label>
      </div>

      <n-input-number
        v-if="exportOptions.range === 'custom'"
        v-model:value="exportOptions.limit"
        :min="1"
        :max="100000"
        size="small"
        style="width: 160px; margin-top: 4px;"
        :placeholder="$t('ie.rowCount')"
      />
    </div>

    <template #footer>
      <div class="ie-footer">
        <n-button @click="visible = false">{{ $t('common.cancel') }}</n-button>
        <n-button type="primary" :loading="loading" @click="doExport">
          <template #icon>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          </template>
          {{ $t('ie.download') }}
        </n-button>
      </div>
    </template>
  </FormModal>

  <FormModal
    v-if="mode === 'import'"
    v-model="visible"
    :title="$t('ie.importTitle', { table: tableName })"
    class="ie-modal"
    :width="600"
  >
    <div class="ie-body">

      <!-- Format Toggle (inline, compact) -->
      <div class="import-format-row">
        <span class="ie-section-label" style="margin: 0;">{{ $t('ie.importFileType') }}</span>
        <div class="format-toggle">
          <button
            class="format-toggle__btn"
            :class="{ 'format-toggle__btn--active': importOptions.format === 'csv' }"
            @click="importOptions.format = 'csv'"
            type="button"
          >CSV</button>
          <button
            class="format-toggle__btn"
            :class="{ 'format-toggle__btn--active': importOptions.format === 'sql' }"
            @click="importOptions.format = 'sql'"
            type="button"
          >SQL</button>
        </div>
      </div>

      <!-- Drop Zone -->
      <div
        class="drop-zone"
        :class="{ 'drop-zone-active': isDragging, 'drop-zone-ready': importFile }"
        @dragover.prevent="isDragging = true"
        @dragleave="isDragging = false"
        @drop.prevent="onFileDrop"
        @click="fileInput?.click()"
      >
        <input ref="fileInput" type="file" :accept="fileAccept" style="display:none" @change="onFileSelect" />

        <template v-if="!importFile">
          <div class="drop-zone__icon">
            <svg v-if="importOptions.format === 'sql'" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="rgba(232,177,60,0.7)" stroke-width="1.2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="8" y1="13" x2="16" y2="13"/><line x1="8" y1="17" x2="16" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
            <svg v-else width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="rgba(87,114,255,0.7)" stroke-width="1.2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="8" y1="13" x2="16" y2="13"/><line x1="8" y1="17" x2="12" y2="17"/></svg>
          </div>
          <div class="drop-hint">{{ importOptions.format === 'sql' ? $t('ie.dropHintSql') : $t('ie.dropHint') }}</div>
          <div class="drop-sub">{{ importOptions.format === 'sql' ? $t('ie.dropSubSql') : $t('ie.dropSub') }}</div>
        </template>
        <template v-else>
          <div class="drop-zone__file-ready">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="rgba(77,204,115,0.9)" stroke-width="1.8"><polyline points="20 6 9 17 4 12"/></svg>
            <div>
              <div class="drop-hint drop-hint--ready">{{ importFile.name }}</div>
              <div class="drop-sub">{{ formatBytes(importFile.size) }} · {{ $t('ie.clickToChange') }}</div>
            </div>
          </div>
        </template>
      </div>

      <!-- SQL restriction notice -->
      <n-alert v-if="importOptions.format === 'sql'" type="warning" :bordered="false" size="small">
        {{ $t('ie.sqlImportRestriction', { table: tableName }) }}
      </n-alert>

      <!-- CSV: header row toggle + mapping + preview in one section -->
      <template v-if="importOptions.format === 'csv' && importFile">
        <div class="csv-options-row">
          <label class="csv-toggle-label">
            <n-switch v-model:value="importOptions.headerRow" size="small" />
            <span>{{ $t('ie.headerRow') }}</span>
            <span class="form-hint">{{ $t('ie.headerRowHint') }}</span>
          </label>
        </div>

        <!-- Compact Mapping + Preview combined table -->
        <template v-if="csvHeaders.length > 0">
          <div class="section-header">
            <span class="section-title">{{ $t('ie.mapping') }}</span>
            <span class="section-meta">{{ csvHeaders.length }} {{ $t('ie.columns') }}</span>
          </div>
          <div class="mapping-preview-wrap">
            <table class="mapping-table">
              <thead>
                <tr>
                  <th class="mth mth-csv">{{ $t('ie.csvColumn') }}</th>
                  <th class="mth mth-arrow"></th>
                  <th class="mth mth-target">{{ $t('ie.targetColumn') }}</th>
                  <th v-for="(_, di) in csvDataRows.slice(0, 3)" :key="di" class="mth mth-preview">
                    {{ $t('ie.previewRow') }} {{ di + 1 }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(header, hi) in csvHeaders" :key="header">
                  <td class="mtd mtd-csv">
                    <span class="csv-col-name">{{ header }}</span>
                  </td>
                  <td class="mtd mtd-arrow">→</td>
                  <td class="mtd mtd-target">
                    <n-select
                      v-model:value="fieldMappings[header]"
                      :options="[{ label: $t('ie.skip'), value: '' }, ...mappingOptions]"
                      size="tiny"
                      :placeholder="$t('ie.skip')"
                    />
                  </td>
                  <td
                    v-for="(dataRow, di) in csvDataRows.slice(0, 3)"
                    :key="di"
                    class="mtd mtd-preview"
                  >
                    <span class="preview-cell">{{ dataRow[hi] }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </template>

      <!-- SQL: preview -->
      <template v-if="importOptions.format === 'sql' && sqlPreviewText">
        <div class="section-header">
          <span class="section-title">{{ $t('ie.preview') }}</span>
        </div>
        <div class="sql-preview-wrap">
          <pre class="sql-preview">{{ sqlPreviewText }}</pre>
        </div>
      </template>

      <!-- Import Result -->
      <div v-if="importResult" class="import-result" :class="resultClass">
        <div class="import-result__summary">
          <svg v-if="importResult.errors?.length === 0" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="rgba(77,204,115,1)" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="rgba(232,177,60,1)" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          <span>{{ $t('ie.importResult', { n: importResult.imported }) }}</span>
          <span v-if="importResult.errors?.length > 0" class="result-error-count">
            · {{ $t('ie.importErrors', { n: importResult.errors.length }) }}
          </span>
          <button
            v-if="importResult.errors?.length > 0"
            class="result-expand-btn"
            @click="showErrors = !showErrors"
            type="button"
          >{{ showErrors ? $t('ie.hideErrors') : $t('ie.showErrors') }} {{ showErrors ? '▲' : '▼' }}</button>
        </div>
        <div v-if="showErrors && importResult.errors?.length > 0" class="import-errors-list">
          <div v-for="(err, i) in importResult.errors.slice(0, 20)" :key="i" class="import-error-item">
            {{ err }}
          </div>
          <div v-if="importResult.errors.length > 20" class="import-error-more">
            + {{ importResult.errors.length - 20 }} {{ $t('ie.moreErrors') }}
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="ie-footer">
        <n-button @click="visible = false">{{ $t('common.cancel') }}</n-button>
        <n-button type="primary" :loading="loading" :disabled="!importFile" @click="doImport">
          <template #icon>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          </template>
          {{ $t('ie.startImport') }}
        </n-button>
      </div>
    </template>
  </FormModal>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import gmssh from '@/utils/gmssh'
import FormModal from '@/components/FormModal.vue'
import api from '@/utils/api'
import { useDatabaseStore } from '@/stores/database'

const props = defineProps({
  mode: { type: String, default: 'export' },
  tableName: { type: String, required: true },
  database: { type: String, default: '' },
  connId: { type: String, default: '' },
  pageRows: { type: Array, default: () => [] },
  pageColumns: { type: Array, default: () => [] },
  totalRows: { type: Number, default: 0 }
})

const emit = defineEmits(['imported'])

const { t } = useI18n()
const store = useDatabaseStore()
const visible = defineModel({ default: false })

const loading = ref(false)

const exportOptions = ref({
  format: 'csv',
  range: props.pageRows?.length ? 'current' : 'all',
  limit: 1000
})

const importFile = ref(null)
const importContent = ref('')
const isDragging = ref(false)
const importOptions = ref({ format: 'csv', headerRow: true })
const importResult = ref(null)
const fileInput = ref(null)
const tableColumns = ref([])
const fieldMappings = ref({})
const showErrors = ref(false)

const hasCurrentPageExport = computed(() => props.pageRows.length > 0 && props.pageColumns.length > 0)
const fileAccept = computed(() => (importOptions.value.format === 'sql' ? '.sql,.txt' : '.csv,.txt'))
const totalRows = computed(() => props.totalRows ?? 0)
const mappingOptions = computed(() => tableColumns.value.map((column) => ({
  label: column.name,
  value: column.name
})))

// Format card definitions for export
const formatOptions = [
  {
    value: 'csv',
    label: 'CSV',
    descKey: 'ie.fmtDescCsv',
    icon: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="8" y1="13" x2="16" y2="13"/><line x1="8" y1="17" x2="16" y2="17"/><line x1="8" y1="9" x2="10" y2="9"/></svg>'
  },
  {
    value: 'json',
    label: 'JSON',
    descKey: 'ie.fmtDescJson',
    icon: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>'
  },
  {
    value: 'sql',
    label: 'SQL',
    descKey: 'ie.fmtDescSql',
    icon: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/></svg>'
  }
]

const formatTipKeys = {
  csv: 'ie.tipCsv',
  json: 'ie.tipJson',
  sql: 'ie.tipSql'
}
const currentFormatTip = computed(() => formatTipKeys[exportOptions.value.format] || 'ie.tipCsv')

const resultClass = computed(() => {
  if (!importResult.value) return ''
  return importResult.value.errors?.length === 0 ? 'import-result--success' : 'import-result--warning'
})

const csvPreviewRows = computed(() => {
  if (!importContent.value || importOptions.value.format !== 'csv') return []
  return importContent.value
    .split(/\r?\n/)
    .filter((line) => line.trim())
    .slice(0, 6)
    .map((line) => parseCSVLine(line))
})

const csvHeaders = computed(() => {
  if (!csvPreviewRows.value.length) return []
  if (importOptions.value.headerRow) {
    return csvPreviewRows.value[0]
  }
  return csvPreviewRows.value[0].map((_, index) => `col${index + 1}`)
})

const csvDataRows = computed(() => {
  if (!csvPreviewRows.value.length) return []
  return importOptions.value.headerRow ? csvPreviewRows.value.slice(1) : csvPreviewRows.value
})

const sqlPreviewText = computed(() => {
  if (importOptions.value.format !== 'sql' || !importContent.value) return ''
  const normalized = importContent.value.trim()
  if (normalized.length <= 1600) return normalized
  return `${normalized.slice(0, 1600)}\n...`
})

watch(() => props.pageRows.length, (length) => {
  if (length > 0 && exportOptions.value.range === 'all') {
    exportOptions.value.range = 'current'
  }
})

watch([csvHeaders, tableColumns], () => {
  if (importOptions.value.format !== 'csv') {
    fieldMappings.value = {}
    return
  }
  const nextMapping = {}
  csvHeaders.value.forEach((header, index) => {
    const directMatch = tableColumns.value.find((column) => column.name === header)?.name
    nextMapping[header] = directMatch || tableColumns.value[index]?.name || ''
  })
  fieldMappings.value = nextMapping
})

watch(() => importOptions.value.format, () => {
  importFile.value = null
  importContent.value = ''
  importResult.value = null
  fieldMappings.value = {}
  if (fileInput.value) {
    fileInput.value.value = ''
  }
})

watch(
  () => [visible.value, props.mode, props.tableName, props.database, props.connId],
  async ([show, mode]) => {
    if (!show || mode !== 'import') return

    try {
      const connId = props.connId || store.activeConnection
      const database = props.database || store.selectedDatabase
      const schema = await api.getTableSchema(connId, database, props.tableName)
      tableColumns.value = schema?.columns || []
    } catch {
      tableColumns.value = []
    }
  },
  { immediate: true }
)

function parseCSVLine(line) {
  const result = []
  let current = ''
  let inQuotes = false

  for (let i = 0; i < line.length; i += 1) {
    const char = line[i]
    const next = line[i + 1]

    if (char === '"' && next === '"') {
      current += '"'
      i += 1
      continue
    }

    if (char === '"') {
      inQuotes = !inQuotes
      continue
    }

    if (char === ',' && !inQuotes) {
      result.push(current)
      current = ''
      continue
    }

    current += char
  }

  result.push(current)
  return result
}

function toCSV(columns, rows) {
  const escapeCell = (value) => {
    if (value == null) return ''
    const text = String(value)
    if (/[",\n]/.test(text)) {
      return `"${text.replace(/"/g, '""')}"`
    }
    return text
  }

  const lines = [
    columns.map(escapeCell).join(',')
  ]

  for (const row of rows) {
    lines.push(columns.map((column) => escapeCell(row[column])).join(','))
  }

  return lines.join('\n')
}

function toJSON(rows) {
  return JSON.stringify(rows, null, 2)
}

function toSQL(tableName, columns, rows, database) {
  const lines = []

  // Database context header
  if (database) {
    lines.push(`CREATE DATABASE IF NOT EXISTS \`${database}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;`)
    lines.push(`USE \`${database}\`;`)
    lines.push('')
  }

  if (!rows.length) {
    lines.push(`-- ${tableName}: no data`)
    return lines.join('\n')
  }

  // INSERT statements
  for (const row of rows) {
    const values = columns.map((column) => {
      const value = row[column]
      if (value == null || value === '') return 'NULL'
      return `'${String(value).replace(/'/g, "''")}'`
    })
    const columnList = columns.map((column) => `\`${column}\``).join(', ')
    lines.push(`INSERT INTO \`${tableName}\` (${columnList}) VALUES (${values.join(', ')});`)
  }

  return lines.join('\n')
}

function downloadText(data, fileName, format) {
  const mimeTypes = {
    csv: 'text/csv;charset=utf-8',
    json: 'application/json;charset=utf-8',
    sql: 'text/plain;charset=utf-8'
  }

  const blob = new Blob([data], { type: mimeTypes[format] || 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

async function doExport() {
  loading.value = true

  try {
    if (exportOptions.value.range === 'current' && hasCurrentPageExport.value) {
      const rows = props.pageRows
      const columns = props.pageColumns
      const format = exportOptions.value.format
      const ext = format === 'sql' ? 'sql' : format
      const fileName = `${props.tableName}_page.${ext}`

      let data = ''
      if (format === 'json') data = toJSON(rows)
      else if (format === 'sql') data = toSQL(props.tableName, columns, rows, props.database || store.selectedDatabase)
      else data = toCSV(columns, rows)

      downloadText(data, fileName, format)
      gmssh.success(t('ie.exportSuccess', { n: rows.length }))
      visible.value = false
      return
    }

    const connId = props.connId || store.activeConnection
    const database = props.database || store.selectedDatabase
    const limit = exportOptions.value.range === 'custom' ? exportOptions.value.limit : 0
    const result = await api.exportTable(connId, database, props.tableName, exportOptions.value.format, limit)

    downloadText(result.data, result.fileName, exportOptions.value.format)
    gmssh.success(t('ie.exportSuccess', { n: result.rowCount }))
    visible.value = false
  } catch (error) {
    gmssh.error(error.message)
  } finally {
    loading.value = false
  }
}

function onFileDrop(event) {
  isDragging.value = false
  const file = event.dataTransfer.files[0]
  if (file) loadFile(file)
}

function onFileSelect(event) {
  const file = event.target.files[0]
  if (file) loadFile(file)
}

function loadFile(file) {
  importFile.value = file
  importResult.value = null

  const reader = new FileReader()
  reader.onload = (event) => {
    importContent.value = event.target.result
  }
  reader.readAsText(file, 'UTF-8')
}

async function doImport() {
  if (!importContent.value) return

  loading.value = true
  importResult.value = null

  try {
    const connId = props.connId || store.activeConnection
    const database = props.database || store.selectedDatabase
    const result = importOptions.value.format === 'sql'
      ? await api.importTableSQL(connId, database, props.tableName, importContent.value)
      : await api.importCSV(
        connId,
        database,
        props.tableName,
        importContent.value,
        fieldMappings.value,
        importOptions.value.headerRow
      )

    importResult.value = result
    showErrors.value = false
    if (result.imported > 0) {
      gmssh.success(t('ie.importSuccess', { n: result.imported }))
      emit('imported')
    }
    if (result.errors?.length > 0) {
      gmssh.warning(t('ie.importPartial', { errors: result.errors.length }))
    }
  } catch (error) {
    gmssh.error(error.message)
  } finally {
    loading.value = false
  }
}

function formatBytes(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}
</script>

<style scoped>
/* ── Layout ─────────────────────────────────────────── */
.ie-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ie-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.ie-section-label {
  font-size: 11px;
  font-weight: 600;
  color: rgba(255,255,255,0.4);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 2px;
}

/* ── Export: Format Cards ────────────────────────────── */
.format-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.format-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  padding: 12px 8px;
  border-radius: 9px;
  border: 1px solid rgba(255,255,255,0.08);
  background: rgba(255,255,255,0.03);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s, box-shadow 0.15s;
  text-align: center;
  color: rgba(255,255,255,0.65);
}

.format-card:hover {
  border-color: rgba(87,114,255,0.35);
  background: rgba(87,114,255,0.06);
}

.format-card--active {
  border-color: rgba(87,114,255,0.7);
  background: rgba(87,114,255,0.1);
  box-shadow: 0 0 0 1px rgba(87,114,255,0.3) inset;
  color: rgba(255,255,255,0.92);
}

.format-card__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: inherit;
  opacity: 0.8;
}
.format-card--active .format-card__icon { opacity: 1; }

.format-card__name {
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.format-card__desc {
  font-size: 11px;
  color: rgba(255,255,255,0.35);
  line-height: 1.3;
}
.format-card--active .format-card__desc { color: rgba(255,255,255,0.55); }

/* ── Export: Format Tip ─────────────────────────────── */
.format-tip {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 12px;
  color: rgba(255,255,255,0.4);
  line-height: 1.5;
  padding: 6px 10px;
  background: rgba(255,255,255,0.02);
  border-radius: 6px;
  border: 1px solid rgba(255,255,255,0.05);
}
.format-tip svg { flex-shrink: 0; margin-top: 1px; }

/* ── Export: Range Options ──────────────────────────── */
.range-options {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.range-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border-radius: 7px;
  border: 1px solid rgba(255,255,255,0.06);
  background: rgba(255,255,255,0.02);
  cursor: pointer;
  transition: border-color 0.12s, background 0.12s;
}

.range-option:hover { border-color: rgba(255,255,255,0.12); }

.range-option--active {
  border-color: rgba(87,114,255,0.5);
  background: rgba(87,114,255,0.07);
}

.range-option input[type="radio"] { accent-color: #5772FF; }

.range-option__text {
  font-size: 13px;
  color: rgba(255,255,255,0.75);
  display: flex;
  align-items: center;
  gap: 6px;
}

.range-option__text em {
  font-style: normal;
  font-size: 11px;
  color: rgba(255,255,255,0.35);
  font-variant-numeric: tabular-nums;
}

/* ── Import: Format Row ─────────────────────────────── */
.import-format-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.format-toggle {
  display: flex;
  border-radius: 7px;
  border: 1px solid rgba(255,255,255,0.1);
  overflow: hidden;
  background: rgba(255,255,255,0.03);
}

.format-toggle__btn {
  padding: 4px 16px;
  font-size: 12px;
  font-weight: 600;
  color: rgba(255,255,255,0.5);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: background 0.14s, color 0.14s;
  letter-spacing: 0.04em;
}

.format-toggle__btn:hover { background: rgba(255,255,255,0.06); color: rgba(255,255,255,0.8); }
.format-toggle__btn--active {
  background: rgba(87,114,255,0.2);
  color: rgba(87,114,255,0.95);
}

/* ── Drop Zone ──────────────────────────────────────── */
.drop-zone {
  border: 1.5px dashed rgba(255,255,255,0.12);
  border-radius: 10px;
  padding: 22px 28px;
  text-align: center;
  cursor: pointer;
  transition: border-color 0.18s, background 0.18s;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 7px;
  background: rgba(255,255,255,0.02);
}

.drop-zone:hover {
  border-color: rgba(87,114,255,0.45);
  background: rgba(87,114,255,0.04);
}

.drop-zone-active {
  border-color: rgba(87,114,255,0.75);
  background: rgba(87,114,255,0.08);
}

.drop-zone-ready {
  border-color: rgba(77,204,115,0.45);
  background: rgba(77,204,115,0.04);
  padding: 16px 28px;
}

.drop-zone__icon { opacity: 0.85; }

.drop-zone__file-ready {
  display: flex;
  align-items: center;
  gap: 10px;
  text-align: left;
}

.drop-hint {
  font-size: 13px;
  color: rgba(255,255,255,0.65);
  font-weight: 500;
}

.drop-hint--ready { color: rgba(77,204,115,0.92); }
.drop-sub { font-size: 12px; color: rgba(255,255,255,0.35); }

/* ── CSV Options Row ────────────────────────────────── */
.csv-options-row {
  display: flex;
  align-items: center;
}

.csv-toggle-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: rgba(255,255,255,0.72);
}

.form-hint {
  font-size: 11px;
  color: rgba(255,255,255,0.35);
  margin-left: 2px;
}

/* ── Section Header ─────────────────────────────────── */
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2px;
}

.section-title {
  font-size: 11px;
  font-weight: 600;
  color: rgba(255,255,255,0.4);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.section-meta {
  font-size: 11px;
  color: rgba(255,255,255,0.28);
}

/* ── Compact Mapping + Preview Table ────────────────── */
.mapping-preview-wrap {
  border-radius: 7px;
  border: 1px solid rgba(255,255,255,0.07);
  overflow: auto;
  max-height: 260px;
}

.mapping-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.mth {
  position: sticky;
  top: 0;
  padding: 5px 8px;
  background: rgba(255,255,255,0.04);
  backdrop-filter: blur(8px);
  color: rgba(255,255,255,0.4);
  font-weight: 600;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  border-bottom: 1px solid rgba(255,255,255,0.07);
  white-space: nowrap;
  text-align: left;
}

.mth-csv { width: 28%; min-width: 90px; }
.mth-arrow { width: 24px; text-align: center; }
.mth-target { width: 28%; min-width: 110px; }
.mth-preview { color: rgba(87,114,255,0.6); min-width: 80px; }

.mtd {
  padding: 5px 8px;
  border-bottom: 1px solid rgba(255,255,255,0.04);
  vertical-align: middle;
}

.mapping-table tr:last-child .mtd { border-bottom: none; }
.mapping-table tr:hover .mtd { background: rgba(255,255,255,0.02); }

.mtd-arrow {
  text-align: center;
  color: rgba(255,255,255,0.2);
  font-size: 12px;
  padding: 0 2px;
}

.csv-col-name {
  font-family: var(--ref-font-family-mono);
  font-size: 11.5px;
  color: rgba(255,255,255,0.75);
  background: rgba(255,255,255,0.05);
  border-radius: 3px;
  padding: 1px 5px;
}

.preview-cell {
  font-size: 11px;
  color: rgba(255,255,255,0.45);
  max-width: 100px;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ── SQL Preview ─────────────────────────────────────── */
.sql-preview-wrap {
  border-radius: 6px;
  border: 1px solid rgba(255,255,255,0.07);
  background: rgba(255,255,255,0.02);
  max-height: 200px;
  overflow: auto;
}

.sql-preview {
  margin: 0;
  padding: 12px 14px;
  color: rgba(255,255,255,0.78);
  font-size: 12px;
  line-height: 1.6;
  font-family: var(--ref-font-family-mono);
  white-space: pre-wrap;
  word-break: break-word;
}

/* ── Import Result ───────────────────────────────────── */
.import-result {
  border-radius: 8px;
  border: 1px solid;
  padding: 10px 12px;
  font-size: 13px;
}

.import-result--success {
  border-color: rgba(77,204,115,0.3);
  background: rgba(77,204,115,0.06);
}

.import-result--warning {
  border-color: rgba(232,177,60,0.3);
  background: rgba(232,177,60,0.05);
}

.import-result__summary {
  display: flex;
  align-items: center;
  gap: 7px;
  flex-wrap: wrap;
}

.result-error-count {
  color: rgba(232,177,60,0.85);
  font-size: 12px;
}

.result-expand-btn {
  margin-left: auto;
  font-size: 11px;
  color: rgba(255,255,255,0.4);
  background: none;
  border: none;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  transition: background 0.12s, color 0.12s;
}
.result-expand-btn:hover {
  background: rgba(255,255,255,0.06);
  color: rgba(255,255,255,0.7);
}

.import-errors-list {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid rgba(255,255,255,0.06);
  display: flex;
  flex-direction: column;
  gap: 3px;
  max-height: 140px;
  overflow-y: auto;
}

.import-error-item {
  font-size: 11px;
  font-family: var(--ref-font-family-mono);
  color: rgba(232,177,60,0.75);
  line-height: 1.5;
  padding: 1px 4px;
}

.import-error-more {
  font-size: 11px;
  color: rgba(255,255,255,0.3);
  padding: 2px 4px;
  font-style: italic;
}
</style>

