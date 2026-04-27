<template>
  <div class="table-designer">
    <!-- Toolbar -->
    <div class="designer-toolbar">
      <div class="toolbar-left">
        <span class="designer-label">{{ isNew ? $t('designer.newTable') : $t('designer.editTable') }}</span>
        <n-input
          v-model:value="tableDef.name"
          :placeholder="$t('designer.tableNamePlaceholder')"
          size="small"
          style="width: 200px; margin-left: 12px;"
          :status="!tableDef.name ? 'error' : undefined"
        />
        <n-input
          v-model:value="tableDef.comment"
          :placeholder="$t('designer.tableCommentPlaceholder')"
          size="small"
          style="width: 180px; margin-left: 8px;"
        />
        <template v-if="dbType === 'mysql'">
          <n-select
            v-model:value="tableDef.engine"
            :options="engineOptions"
            size="small"
            style="width: 110px; margin-left: 8px;"
          />
          <n-select
            v-model:value="tableDef.charset"
            :options="charsetOptions"
            size="small"
            style="width: 120px; margin-left: 8px;"
          />
        </template>
      </div>
      <div class="toolbar-right">
        <n-button size="small" @click="showDDL = true">
          <template #icon><i class="icon-code" /></template>
          {{ $t('designer.previewDDL') }}
        </n-button>
        <n-button size="small" :loading="saving" type="primary" @click="save" :disabled="!tableDef.name">
          {{ isNew ? $t('designer.createTable') : $t('designer.saveChanges') }}
        </n-button>
      </div>
    </div>

    <!-- Sub Tabs: Columns / Indexes -->
    <n-tabs v-model:value="activeTab" type="line" size="small" class="designer-tabs">
      <!-- Columns Tab -->
      <n-tab-pane name="columns" :tab="$t('designer.columns')">
        <div class="columns-editor">
          <!-- Column list header -->
          <div class="col-header-row">
            <div class="col-h col-pk">PK</div>
            <div class="col-h col-name">{{ $t('designer.colName') }}</div>
            <div class="col-h col-type">{{ $t('designer.colType') }}</div>
            <div class="col-h col-len">{{ $t('designer.colLength') }}</div>
            <div class="col-h col-notnull">NOT NULL</div>
            <div class="col-h col-ai" v-if="dbType !== 'postgres'">AUTO INC</div>
            <div class="col-h col-default">{{ $t('designer.colDefault') }}</div>
            <div class="col-h col-comment">{{ $t('designer.colComment') }}</div>
            <div class="col-h col-actions"></div>
          </div>

          <!-- Column rows -->
          <div
            v-for="(col, idx) in tableDef.columns"
            :key="idx"
            class="col-row"
            :class="{ 'col-row-active': activeColIdx === idx }"
            @click="activeColIdx = idx"
          >
            <div class="col-cell col-pk">
              <n-checkbox v-model:checked="col.isPrimaryKey" @update:checked="v => onPKChange(idx, v)" />
            </div>
            <div class="col-cell col-name">
              <n-input v-model:value="col.name" size="tiny" :placeholder="$t('designer.colNamePlaceholder')" />
            </div>
            <div class="col-cell col-type">
              <n-select
                v-model:value="col.type"
                :options="getTypeOptions()"
                size="tiny"
                filterable
                style="min-width: 120px;"
              />
            </div>
            <div class="col-cell col-len">
              <n-input
                v-model:value="col.length"
                size="tiny"
                :placeholder="getDefaultLength(col.type)"
                :disabled="!supportsLength(col.type)"
              />
            </div>
            <div class="col-cell col-notnull">
              <n-checkbox v-model:checked="col.notNull" :disabled="col.isPrimaryKey" />
            </div>
            <div class="col-cell col-ai" v-if="dbType !== 'postgres'">
              <n-checkbox v-model:checked="col.autoIncrement" :disabled="!col.isPrimaryKey" />
            </div>
            <div class="col-cell col-default">
              <n-input v-model:value="col.defaultValue" size="tiny" :placeholder="$t('designer.nullPlaceholder')" />
            </div>
            <div class="col-cell col-comment">
              <n-input v-model:value="col.comment" size="tiny" />
            </div>
            <div class="col-cell col-actions">
              <n-button text size="tiny" type="error" @click.stop="removeColumn(idx)" :disabled="tableDef.columns.length <= 1">
                <template #icon>
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </template>
              </n-button>
              <n-button text size="tiny" @click.stop="moveColumn(idx, -1)" :disabled="idx === 0" style="margin-left: 2px;">
                <template #icon><svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="18 15 12 9 6 15"/></svg></template>
              </n-button>
              <n-button text size="tiny" @click.stop="moveColumn(idx, 1)" :disabled="idx === tableDef.columns.length - 1">
                <template #icon><svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="6 9 12 15 18 9"/></svg></template>
              </n-button>
            </div>
          </div>

          <!-- Add column button -->
          <div class="add-col-row" @click="addColumn">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            {{ $t('designer.addColumn') }}
          </div>
        </div>
      </n-tab-pane>

      <!-- Indexes Tab -->
      <n-tab-pane name="indexes" :tab="$t('designer.indexes')">
        <div class="indexes-editor">
          <div class="idx-header-row">
            <div class="idx-h idx-name">{{ $t('designer.idxName') }}</div>
            <div class="idx-h idx-columns">{{ $t('designer.idxColumns') }}</div>
            <div class="idx-h idx-unique">{{ $t('designer.idxUnique') }}</div>
            <div class="idx-h idx-actions"></div>
          </div>
          <div v-for="(idx, i) in tableDef.indexes" :key="i" class="idx-row">
            <div class="idx-cell idx-name">
              <n-input v-model:value="idx.name" size="tiny" :placeholder="$t('designer.idxNamePlaceholder')" />
            </div>
            <div class="idx-cell idx-columns">
              <n-select
                v-model:value="idx.columns"
                multiple
                :options="columnNameOptions"
                size="tiny"
              />
            </div>
            <div class="idx-cell idx-unique">
              <n-checkbox v-model:checked="idx.unique" />
            </div>
            <div class="idx-cell idx-actions">
              <n-button text size="tiny" type="error" @click="removeIndex(i)">
                <template #icon><svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></template>
              </n-button>
            </div>
          </div>
          <div class="add-col-row" @click="addIndex">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            {{ $t('designer.addIndex') }}
          </div>
        </div>
      </n-tab-pane>
    </n-tabs>

    <!-- DDL Preview Drawer -->
    <n-drawer v-model:show="showDDL" :width="560" placement="right">
      <n-drawer-content :title="$t('designer.ddlPreview')" closable>
        <n-code :code="generatedDDL" language="sql" />
        <template #footer>
          <n-button type="primary" @click="copyDDL">{{ $t('designer.copyDDL') }}</n-button>
        </template>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDatabaseStore } from '@/stores/database'
import api from '@/utils/api'
import gmssh from '@/utils/gmssh'

const props = defineProps({
  tableName: { type: String, default: '' },      // empty = new table
  database: { type: String, default: '' },
  connId: { type: String, default: '' },
  dbType: { type: String, default: 'mysql' },    // mysql | postgres | sqlite
  initialSchema: { type: Object, default: null } // for edit mode, pass existing schema
})

const emit = defineEmits(['saved', 'close'])

const { t } = useI18n()
const store = useDatabaseStore()

const isNew = computed(() => !props.tableName)
const saving = ref(false)
const showDDL = ref(false)
const activeTab = ref('columns')
const activeColIdx = ref(0)
const originalDef = ref(null)

// Table definition state
const tableDef = ref({
  name: props.tableName || '',
  comment: '',
  engine: 'InnoDB',
  charset: 'utf8mb4',
  columns: [defaultColumn()],
  indexes: []
})

// Load existing structure if in edit mode
watch(() => props.initialSchema, (schema) => {
  if (!schema || !schema.columns) return
  tableDef.value.columns = schema.columns.map(col => ({
    name: col.name,
    type: col.type.split('(')[0].toLowerCase(),
    length: col.type.includes('(') ? col.type.match(/\(([^)]+)\)/)?.[1] || '' : '',
    notNull: !col.nullable,
    isPrimaryKey: !!col.isPrimaryKey,
    autoIncrement: col.type.toLowerCase().includes('auto_increment'),
    defaultValue: col.defaultValue || '',
    comment: col.comment || '',
    unsigned: false,
    _originName: col.name
  }))
  tableDef.value.indexes = (schema.indexes || []).map(idx => ({
    name: idx.name,
    columns: idx.columns,
    unique: idx.unique
  }))
  originalDef.value = snapshotCurrentDef()
}, { immediate: true })

// Column name options for index selector
const columnNameOptions = computed(() =>
  tableDef.value.columns
    .filter(c => c.name)
    .map(c => ({ label: c.name, value: c.name }))
)

// Engine / charset options for MySQL
const engineOptions = [
  { label: 'InnoDB', value: 'InnoDB' },
  { label: 'MyISAM', value: 'MyISAM' },
  { label: 'MEMORY', value: 'MEMORY' }
]
const charsetOptions = [
  { label: 'utf8mb4', value: 'utf8mb4' },
  { label: 'utf8', value: 'utf8' },
  { label: 'latin1', value: 'latin1' },
  { label: 'gbk', value: 'gbk' }
]

// Column type options based on DB type
function getTypeOptions() {
  if (props.dbType === 'postgres') {
    return [
      { label: 'INTEGER', value: 'integer' },
      { label: 'BIGINT', value: 'bigint' },
      { label: 'SMALLINT', value: 'smallint' },
      { label: 'SERIAL', value: 'serial' },
      { label: 'BIGSERIAL', value: 'bigserial' },
      { label: 'NUMERIC', value: 'numeric' },
      { label: 'REAL', value: 'real' },
      { label: 'DOUBLE PRECISION', value: 'double precision' },
      { label: 'BOOLEAN', value: 'boolean' },
      { label: 'VARCHAR', value: 'varchar' },
      { label: 'CHAR', value: 'char' },
      { label: 'TEXT', value: 'text' },
      { label: 'BYTEA', value: 'bytea' },
      { label: 'DATE', value: 'date' },
      { label: 'TIMESTAMP', value: 'timestamp' },
      { label: 'TIMESTAMPTZ', value: 'timestamptz' },
      { label: 'JSON', value: 'json' },
      { label: 'JSONB', value: 'jsonb' },
      { label: 'UUID', value: 'uuid' }
    ]
  } else if (props.dbType === 'sqlite') {
    return [
      { label: 'INTEGER', value: 'INTEGER' },
      { label: 'REAL', value: 'REAL' },
      { label: 'TEXT', value: 'TEXT' },
      { label: 'BLOB', value: 'BLOB' },
      { label: 'NUMERIC', value: 'NUMERIC' }
    ]
  } else {
    // MySQL
    return [
      { label: 'INT', value: 'int' },
      { label: 'BIGINT', value: 'bigint' },
      { label: 'SMALLINT', value: 'smallint' },
      { label: 'TINYINT', value: 'tinyint' },
      { label: 'FLOAT', value: 'float' },
      { label: 'DOUBLE', value: 'double' },
      { label: 'DECIMAL', value: 'decimal' },
      { label: 'BIT', value: 'bit' },
      { label: 'VARCHAR', value: 'varchar' },
      { label: 'CHAR', value: 'char' },
      { label: 'TEXT', value: 'text' },
      { label: 'MEDIUMTEXT', value: 'mediumtext' },
      { label: 'LONGTEXT', value: 'longtext' },
      { label: 'BLOB', value: 'blob' },
      { label: 'JSON', value: 'json' },
      { label: 'DATE', value: 'date' },
      { label: 'DATETIME', value: 'datetime' },
      { label: 'TIMESTAMP', value: 'timestamp' },
      { label: 'TIME', value: 'time' },
      { label: 'YEAR', value: 'year' },
      { label: 'ENUM', value: 'enum' },
      { label: 'BOOLEAN', value: 'boolean' }
    ]
  }
}

function supportsLength(type) {
  if (!type) return false
  const noLen = ['text', 'mediumtext', 'longtext', 'blob', 'json', 'date', 'datetime',
    'timestamp', 'time', 'year', 'boolean', 'tinyint', 'bigserial', 'serial',
    'integer', 'bigint', 'smallint', 'real', 'double precision', 'boolean',
    'bytea', 'uuid', 'jsonb', 'TEXT', 'REAL', 'BLOB', 'INTEGER', 'NUMERIC']
  return !noLen.includes(type)
}

function getDefaultLength(type) {
  if (!type) return ''
  const map = { varchar: '255', char: '10', decimal: '10,2', float: '10,2', numeric: '10,2', int: '11' }
  return map[type.toLowerCase()] || ''
}

// --- Column/Index management ---

function defaultColumn() {
  return { name: '', type: props.dbType === 'postgres' ? 'varchar' : 'varchar', length: '255', notNull: false, isPrimaryKey: false, autoIncrement: false, defaultValue: '', comment: '', unsigned: false, _originName: '' }
}

function addColumn() {
  tableDef.value.columns.push(defaultColumn())
  activeColIdx.value = tableDef.value.columns.length - 1
}

function removeColumn(idx) {
  tableDef.value.columns.splice(idx, 1)
  if (activeColIdx.value >= tableDef.value.columns.length) {
    activeColIdx.value = tableDef.value.columns.length - 1
  }
}

function moveColumn(idx, dir) {
  const cols = tableDef.value.columns
  const newIdx = idx + dir
  if (newIdx < 0 || newIdx >= cols.length) return
  ;[cols[idx], cols[newIdx]] = [cols[newIdx], cols[idx]]
  activeColIdx.value = newIdx
}

function onPKChange(idx, v) {
  if (v) {
    tableDef.value.columns[idx].notNull = true
  }
}

function addIndex() {
  tableDef.value.indexes.push({ name: '', columns: [], unique: false })
}

function removeIndex(i) {
  tableDef.value.indexes.splice(i, 1)
}

function cleanColumn(col) {
  return {
    name: col.name,
    type: col.type,
    length: supportsLength(col.type) ? (col.length || '') : '',
    notNull: !!col.notNull,
    defaultValue: col.defaultValue || '',
    comment: col.comment || '',
    isPrimaryKey: !!col.isPrimaryKey,
    autoIncrement: !!col.autoIncrement,
    unsigned: !!col.unsigned
  }
}

function cleanIndex(index) {
  return {
    name: index.name,
    columns: [...(index.columns || [])],
    unique: !!index.unique
  }
}

function snapshotCurrentDef() {
  return {
    name: tableDef.value.name,
    comment: tableDef.value.comment,
    engine: tableDef.value.engine,
    charset: tableDef.value.charset,
    columns: tableDef.value.columns.map((column) => ({
      ...cleanColumn(column),
      _originName: column._originName || column.name
    })),
    indexes: tableDef.value.indexes.map(cleanIndex)
  }
}

function columnsDiffer(previousColumn, nextColumn, ignoreName = false) {
  const previousValue = cleanColumn(previousColumn)
  const nextValue = cleanColumn(nextColumn)

  if (ignoreName) {
    previousValue.name = ''
    nextValue.name = ''
  }

  return JSON.stringify(previousValue) !== JSON.stringify(nextValue)
}

function indexesDiffer(previousIndex, nextIndex) {
  return JSON.stringify(cleanIndex(previousIndex)) !== JSON.stringify(cleanIndex(nextIndex))
}

function buildAlterOps(previousDef, nextDef) {
  if (!previousDef) return []

  const operations = []
  const previousColumnsByName = new Map(previousDef.columns.map((column) => [column.name, column]))
  const keptColumns = new Set()

  for (const column of nextDef.columns) {
    const originName = column._originName || ''
    const previousColumn = originName ? previousColumnsByName.get(originName) : null

    if (!previousColumn) {
      operations.push({ action: 'addColumn', column: cleanColumn(column) })
      continue
    }

    keptColumns.add(originName)

    if (originName !== column.name) {
      operations.push({
        action: 'renameColumn',
        oldName: originName,
        column: cleanColumn(column)
      })
    }

    if (columnsDiffer(previousColumn, column, originName !== column.name)) {
      operations.push({
        action: 'modifyColumn',
        column: cleanColumn(column)
      })
    }
  }

  for (const column of previousDef.columns) {
    if (!keptColumns.has(column.name)) {
      operations.push({
        action: 'dropColumn',
        column: { name: column.name }
      })
    }
  }

  const previousIndexesByName = new Map(previousDef.indexes.map((index) => [index.name, index]))
  const nextIndexesByName = new Map(nextDef.indexes.map((index) => [index.name, index]))

  for (const [name, previousIndex] of previousIndexesByName.entries()) {
    const nextIndex = nextIndexesByName.get(name)
    if (!nextIndex) {
      operations.push({ action: 'dropIndex', index: { name } })
      continue
    }

    if (indexesDiffer(previousIndex, nextIndex)) {
      operations.push({ action: 'dropIndex', index: { name } })
      operations.push({ action: 'addIndex', index: cleanIndex(nextIndex) })
    }
  }

  for (const [name, nextIndex] of nextIndexesByName.entries()) {
    if (!previousIndexesByName.has(name)) {
      operations.push({ action: 'addIndex', index: cleanIndex(nextIndex) })
    }
  }

  return operations
}

// --- DDL Generation ---
const generatedDDL = computed(() => {
  try {
    return buildDDL()
  } catch (e) {
    return `-- Error: ${e.message}`
  }
})

function buildDDL() {
  const def = tableDef.value
  if (!def.name) return '-- Table name required'
  const dbType = props.dbType

  const lines = []
  const pks = def.columns.filter(c => c.isPrimaryKey).map(c => c.name)

  if (dbType === 'mysql') {
    lines.push(`CREATE TABLE \`${def.name}\` (`)
    const colLines = def.columns.map(col => {
      let s = `  \`${col.name}\` ${col.type.toUpperCase()}`
      if (supportsLength(col.type) && col.length) s += `(${col.length})`
      if (col.unsigned) s += ' UNSIGNED'
      s += col.notNull ? ' NOT NULL' : ' NULL'
      if (col.autoIncrement) s += ' AUTO_INCREMENT'
      if (col.defaultValue) s += ` DEFAULT '${col.defaultValue}'`
      if (col.comment) s += ` COMMENT '${col.comment}'`
      return s
    })
    if (pks.length > 0) colLines.push(`  PRIMARY KEY (\`${pks.join('`, `')}\`)`)
    def.indexes.forEach(idx => {
      if (!idx.name || idx.columns.length === 0) return
      const cols = idx.columns.map(c => `\`${c}\``).join(', ')
      colLines.push(`  ${idx.unique ? 'UNIQUE ' : ''}KEY \`${idx.name}\` (${cols})`)
    })
    lines.push(colLines.join(',\n'))
    lines.push(`) ENGINE=${def.engine || 'InnoDB'} DEFAULT CHARSET=${def.charset || 'utf8mb4'}${def.comment ? ` COMMENT='${def.comment}'` : ''};`)
  } else if (dbType === 'postgres') {
    lines.push(`CREATE TABLE "${def.name}" (`)
    const colLines = def.columns.map(col => {
      let t = col.type
      if (supportsLength(col.type) && col.length) t += `(${col.length})`
      let s = `  "${col.name}" ${t.toUpperCase()}`
      if (col.notNull) s += ' NOT NULL'
      if (col.defaultValue) s += ` DEFAULT '${col.defaultValue}'`
      return s
    })
    if (pks.length > 0) colLines.push(`  PRIMARY KEY ("${pks.join('", "')}")`)
    lines.push(colLines.join(',\n'))
    lines.push(');')
    def.indexes.forEach(idx => {
      if (!idx.name || idx.columns.length === 0) return
      const cols = idx.columns.map(c => `"${c}"`).join(', ')
      lines.push(`CREATE ${idx.unique ? 'UNIQUE ' : ''}INDEX "${idx.name}" ON "${def.name}" (${cols});`)
    })
  } else {
    // SQLite
    lines.push(`CREATE TABLE \`${def.name}\` (`)
    const colLines = def.columns.map(col => {
      let s = `  \`${col.name}\` ${col.type.toUpperCase()}`
      if (col.isPrimaryKey && col.autoIncrement) s += ' PRIMARY KEY AUTOINCREMENT'
      else if (col.isPrimaryKey && pks.length === 1) s += ' PRIMARY KEY'
      if (!col.isPrimaryKey && col.notNull) s += ' NOT NULL'
      if (col.defaultValue) s += ` DEFAULT '${col.defaultValue}'`
      return s
    })
    if (pks.length > 1) colLines.push(`  PRIMARY KEY (\`${pks.join('`, `')}\`)`)
    lines.push(colLines.join(',\n'))
    lines.push(');')
  }

  return lines.join('\n')
}

// --- Save ---
async function save() {
  if (!tableDef.value.name) {
    gmssh.error(t('designer.tableNameRequired'))
    return
  }
  const emptyCol = tableDef.value.columns.find(c => !c.name)
  if (emptyCol) {
    gmssh.error(t('designer.colNameRequired'))
    return
  }

  saving.value = true
  try {
    const def = {
      name: tableDef.value.name,
      comment: tableDef.value.comment,
      engine: tableDef.value.engine,
      charset: tableDef.value.charset,
      columns: tableDef.value.columns.map(col => ({
        ...cleanColumn(col),
        _originName: col._originName || ''
      })),
      indexes: tableDef.value.indexes.filter(i => i.name && i.columns.length > 0)
    }

    const connId = props.connId || store.activeConnection
    const database = props.database || store.selectedDatabase

    if (isNew.value) {
      await api.createTable(connId, database, def)
      gmssh.success(t('designer.tableCreated', { name: def.name }))
    } else {
      const previousName = originalDef.value?.name || props.tableName
      if (previousName && previousName !== def.name) {
        await api.renameTable(connId, database, previousName, def.name)
      }

      const operations = buildAlterOps(originalDef.value, def)
      if (operations.length > 0) {
        await api.alterTable(connId, database, def.name, operations)
      }

      gmssh.success(t('designer.tableUpdated', { name: def.name }))
    }

    emit('saved', def.name)
  } catch (e) {
    gmssh.error(e.message)
  } finally {
    saving.value = false
  }
}

async function copyDDL() {
  const copied = await gmssh.copyToClipboard(generatedDDL.value)
  if (!copied) {
    gmssh.error(t('designer.copyUnavailable'))
    return
  }
  gmssh.success(t('designer.ddlCopied'))
}
</script>

<style scoped>
.table-designer {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: transparent;
  overflow: hidden;
}

.designer-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  border-bottom: 1px solid rgba(255,255,255,0.07);
  background: rgba(255,255,255,0.02);
  gap: 8px;
  flex-shrink: 0;
}

.toolbar-left, .toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.designer-label {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255,255,255,0.5);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  flex-shrink: 0;
}

.designer-tabs {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.designer-tabs :deep(.n-tabs-pane-wrapper) {
  flex: 1;
  overflow-y: auto;
}

/* Columns editor table */
.columns-editor, .indexes-editor {
  width: 100%;
}

.col-header-row, .idx-header-row {
  display: flex;
  align-items: center;
  padding: 5px 12px;
  background: rgba(255,255,255,0.04);
  border-bottom: 1px solid rgba(255,255,255,0.06);
  position: sticky;
  top: 0;
  z-index: 1;
}

.col-h, .idx-h {
  font-size: 11px;
  font-weight: 600;
  color: rgba(255,255,255,0.4);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.col-row, .idx-row {
  display: flex;
  align-items: center;
  padding: 5px 12px;
  border-bottom: 1px solid rgba(255,255,255,0.04);
  cursor: pointer;
  transition: background 0.12s;
}

.col-row:hover { background: rgba(255,255,255,0.04); }
.col-row-active { background: rgba(87,114,255,0.08) !important; }

/* Column widths */
.col-pk, .col-notnull, .col-ai { width: 52px; flex-shrink: 0; text-align: center; }
.col-name { flex: 2; min-width: 100px; padding-right: 8px; }
.col-type { flex: 2; min-width: 120px; padding-right: 8px; }
.col-len { flex: 1; min-width: 80px; padding-right: 8px; }
.col-default { flex: 1.5; min-width: 90px; padding-right: 8px; }
.col-comment { flex: 2; min-width: 100px; padding-right: 8px; }
.col-actions { width: 64px; flex-shrink: 0; display: flex; align-items: center; gap: 2px; justify-content: flex-end; }

.col-cell, .idx-cell { display: flex; align-items: center; }

/* Index widths */
.idx-name { flex: 1.5; padding-right: 8px; }
.idx-columns { flex: 3; padding-right: 8px; }
.idx-unique { width: 70px; flex-shrink: 0; }
.idx-actions { width: 40px; flex-shrink: 0; }

.add-col-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  font-size: 12px;
  color: rgba(255,255,255,0.35);
  cursor: pointer;
  border-top: 1px solid rgba(255,255,255,0.04);
  transition: color 0.15s, background 0.15s;
  user-select: none;
}
.add-col-row:hover {
  color: rgba(87,114,255,0.9);
  background: rgba(87,114,255,0.05);
}
</style>
