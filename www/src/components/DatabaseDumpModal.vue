<template>
  <FormModal
    v-model="visible"
    :title="modalTitle"
    class="db-dump-modal"
    :width="560"
  >
    <div class="db-dump-body">
      <div v-if="taskState" class="task-status-card">
        <div class="task-status-head">
          <div>
            <div class="task-status-title">{{ taskStatusText }}</div>
            <div class="task-status-meta">{{ formatTaskMessage(taskState.message, taskState.phase) }}</div>
          </div>
          <div class="task-status-badge" :class="`is-${taskState.status}`">
            {{ taskStatusText }}
          </div>
        </div>

        <div class="task-status-grid">
          <div class="task-status-item">
            <span>{{ $t('dbDump.phase') }}</span>
            <strong>{{ formatTaskPhase(taskState.phase || '-') }}</strong>
          </div>
          <div class="task-status-item" v-if="taskState.mode">
            <span>{{ $t('dbDump.mode') }}</span>
            <strong>{{ formatMode(taskState.mode) }}</strong>
          </div>
          <div class="task-status-item">
            <span>{{ $t('dbDump.startedAt') }}</span>
            <strong>{{ formatDateTime(taskState.startedAt || taskState.createdAt) }}</strong>
          </div>
          <div class="task-status-item" v-if="taskState.fileName">
            <span>{{ $t('dbDump.outputFile') }}</span>
            <strong>{{ taskState.fileName }}</strong>
          </div>
          <div
            v-if="showServerPath"
            class="task-status-item task-status-item--path"
          >
            <span>{{ $t('dbDump.outputPath') }}</span>
            <button class="path-link" type="button" @click="openOutputFolder">
              {{ taskState.filePath }}
            </button>
          </div>
          <div class="task-status-item" v-if="taskState.error">
            <span>{{ $t('dbDump.error') }}</span>
            <strong class="is-error">{{ taskState.error }}</strong>
          </div>
        </div>
      </div>

      <template v-if="mode === 'export'">
        <n-form label-placement="left" label-width="108" size="small">
          <n-form-item :label="$t('dbDump.sourceDatabase')">
            <n-select
              v-model:value="selectedExportDatabase"
              :options="databaseOptions"
              :placeholder="$t('dbDump.databaseSelectPlaceholder')"
              filterable
            />
          </n-form-item>

          <n-form-item :label="$t('dbDump.mode')">
            <n-radio-group v-model:value="exportOptions.mode">
              <n-space vertical>
                <n-radio value="auto">{{ $t('dbDump.modeAuto') }}</n-radio>
                <n-radio value="full">{{ $t('dbDump.modeFull') }}</n-radio>
                <n-radio value="compatible">{{ $t('dbDump.modeCompatible') }}</n-radio>
              </n-space>
            </n-radio-group>
          </n-form-item>

          <n-form-item :label="$t('dbDump.targetPath')">
            <div class="path-row">
              <n-input v-model:value="exportOptions.targetPath" :placeholder="$t('dbDump.targetPathPlaceholder')" />
              <n-button @click="chooseExportPath">{{ $t('dbDump.browse') }}</n-button>
            </div>
          </n-form-item>
        </n-form>

        <div class="advanced-card">
          <button type="button" class="advanced-toggle" @click="showAdvanced = !showAdvanced">
            <span>{{ $t('dbDump.advanced') }}</span>
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              :class="{ 'is-open': showAdvanced }"
            >
              <polyline points="6 9 12 15 18 9" />
            </svg>
          </button>

          <div v-if="showAdvanced" class="advanced-options">
            <n-checkbox v-model:checked="exportOptions.includeRoutines" :disabled="exportOptions.mode === 'compatible'">
              {{ $t('dbDump.includeRoutines') }}
            </n-checkbox>
            <n-checkbox v-model:checked="exportOptions.includeTriggers" :disabled="exportOptions.mode === 'compatible'">
              {{ $t('dbDump.includeTriggers') }}
            </n-checkbox>
            <n-checkbox v-model:checked="exportOptions.includeEvents" :disabled="exportOptions.mode === 'compatible'">
              {{ $t('dbDump.includeEvents') }}
            </n-checkbox>
            <n-checkbox v-model:checked="exportOptions.includeTablespace" :disabled="exportOptions.mode === 'compatible'">
              {{ $t('dbDump.includeTablespace') }}
            </n-checkbox>
          </div>
        </div>
      </template>

      <template v-else>
        <n-form label-placement="left" label-width="auto" size="small">
          <n-form-item :label="$t('dbDump.filePath')">
            <div class="path-row">
              <n-input v-model:value="importOptions.filePath" :placeholder="$t('dbDump.filePathPlaceholder')" />
              <n-button @click="chooseImportFile">{{ $t('dbDump.browse') }}</n-button>
            </div>
          </n-form-item>

          <n-form-item :label="$t('dbDump.targetDatabase')">
            <n-input
              v-model:value="importOptions.database"
              :placeholder="$t('dbDump.targetDatabasePlaceholder')"
              clearable
            />
          </n-form-item>

          <n-form-item :label="$t('dbDump.importStrategy')">
            <div class="strategy-field">
              <n-radio-group v-model:value="importOptions.strategy">
                <n-space vertical>
                  <n-radio value="source">{{ $t('dbDump.importStrategySource') }}</n-radio>
                  <n-radio value="target">{{ $t('dbDump.importStrategyTarget') }}</n-radio>
                  <n-radio value="replace">{{ $t('dbDump.importStrategyReplace') }}</n-radio>
                </n-space>
              </n-radio-group>

              <!-- 当填了目标库但选了「按源库」策略时提示冲突 -->
              <div
                v-if="importOptions.strategy === 'source' && importOptions.database"
                class="strategy-warn"
              >
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="flex-shrink:0;margin-top:1px"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                {{ $t('dbDump.strategySourceConflict', { database: importOptions.database }) }}
              </div>
            </div>
          </n-form-item>
        </n-form>

        <n-alert
          v-if="importOptions.strategy === 'replace'"
          type="warning"
          :bordered="false"
          size="small"
        >
          {{ $t('dbDump.replaceWarning') }}
        </n-alert>

      </template>
    </div>

    <template #footer>
      <div class="db-dump-footer">
        <n-button @click="visible = false">{{ $t('common.cancel') }}</n-button>
        <n-button
          v-if="canCancelTask"
          quaternary
          type="warning"
          :loading="canceling"
          @click="cancelTask"
        >
          {{ $t('dbDump.cancelTask') }}
        </n-button>
        <n-button type="primary" :loading="loading" :disabled="isTaskRunning" @click="submit">
          {{ mode === 'export' ? $t('dbDump.startExport') : $t('dbDump.startImport') }}
        </n-button>
      </div>
    </template>
  </FormModal>
</template>

<script setup>
import { ref, watch, computed, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import FormModal from '@/components/FormModal.vue'
import { useDatabaseStore } from '@/stores/database'
import api from '@/utils/api'
import gmssh from '@/utils/gmssh'

const props = defineProps({
  mode: { type: String, default: 'export' },
  connId: { type: String, required: true },
  database: { type: String, required: true }
})

const emit = defineEmits(['done'])

const { t } = useI18n()
const store = useDatabaseStore()
const visible = defineModel({ default: false })

const loading = ref(false)
const canceling = ref(false)
const taskState = ref(null)
const pollingTimer = ref(null)
const lastTaskByKey = ref({})
const showAdvanced = ref(false)
const selectedExportDatabase = ref('')
const exportOptions = ref({
  targetPath: '',
  mode: 'auto',
  includeRoutines: false,
  includeTriggers: false,
  includeEvents: false,
  includeTablespace: false
})
const importOptions = ref({
  filePath: '',
  database: '',
  createDatabase: false,
  strategy: 'source'
})

const databaseOptions = computed(() => {
  const state = store.getConnectionState(props.connId)
  const databases = state?.databases || []
  return databases.map((item) => ({
    label: item.name,
    value: item.name
  }))
})

const modalTitle = computed(() => {
  if (props.mode === 'export') {
    return selectedExportDatabase.value
      ? t('dbDump.exportTitle', { database: selectedExportDatabase.value })
      : t('dbDump.exportTitleEmpty')
  }

  return props.database
    ? t('dbDump.importTitle', { database: props.database })
    : t('dbDump.importTitleEmpty')
})

const isTaskRunning = computed(() => ['pending', 'running'].includes(taskState.value?.status))
const canCancelTask = computed(() => ['pending', 'running'].includes(taskState.value?.status))
const taskStatusText = computed(() => {
  const status = taskState.value?.status
  if (status === 'pending') return t('dbDump.statusPending')
  if (status === 'running') return t('dbDump.statusRunning')
  if (status === 'success') return t('dbDump.statusSuccess')
  if (status === 'failed') return t('dbDump.statusFailed')
  if (status === 'canceled') return t('dbDump.statusCanceled')
  if (status === 'interrupted') return t('dbDump.statusInterrupted')
  return ''
})

watch(
  () => [visible.value, props.database, props.mode, props.connId],
  ([show]) => {
    if (!show) {
      stopPolling()
      return
    }
    const userName = window.$gm?.userName || 'root'
    const homeDir = userName === 'root' ? '/root' : `/home/${userName}`
    const defaultExportPath = `${homeDir}/db-exports`
    // Ensure the default export directory exists on the server
    window.$gm?.execShell?.({ cmd: `mkdir -p ${defaultExportPath}`, timeout: '5000' }).catch(() => {})
    exportOptions.value = {
      targetPath: defaultExportPath,
      mode: 'auto',
      includeRoutines: false,
      includeTriggers: false,
      includeEvents: false,
      includeTablespace: false
    }
    selectedExportDatabase.value = props.database || ''
    showAdvanced.value = false
    importOptions.value = {
      filePath: '',
      database: '',
      createDatabase: false,
      strategy: 'source'
    }
    const cacheKey = getTaskCacheKey()
    const cachedTask = lastTaskByKey.value[cacheKey]
    const isRunning = cachedTask && ['pending', 'running'].includes(cachedTask.status)
    taskState.value = isRunning ? cachedTask : null
    if (isRunning) {
      pollTask(cachedTask.id)
    }
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  stopPolling()
})

function formatDateTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function formatMode(value) {
  if (value === 'full') return t('dbDump.modeFull')
  if (value === 'compatible') return t('dbDump.modeCompatible')
  return t('dbDump.modeAuto')
}

function formatTaskPhase(phase) {
  const normalized = String(phase || '').trim()
  if (!normalized || normalized === '-') return normalized || '-'
  const key = normalized.toLowerCase()
  const translated = t(`transfer.phaseMap.${key}`)
  return translated === `transfer.phaseMap.${key}` ? normalized : translated
}

function formatTaskMessage(messageText, phase) {
  const normalized = String(messageText || '').trim()
  if (normalized) {
    const key = normalized
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '_')
      .replace(/^_+|_+$/g, '')
    const translated = t(`transfer.messageMap.${key}`)
    if (translated !== `transfer.messageMap.${key}`) {
      return translated
    }
  }

  return formatTaskPhase(phase || normalized || '-')
}

function stopPolling() {
  if (pollingTimer.value) {
    window.clearTimeout(pollingTimer.value)
    pollingTimer.value = null
  }
}

function getTaskCacheKey() {
  const scopedDatabase = props.mode === 'export'
    ? (selectedExportDatabase.value || props.database)
    : props.database
  return `${props.connId}::${props.mode}::${scopedDatabase || 'global'}`
}

function rememberTask(task) {
  if (!task?.id) return
  lastTaskByKey.value = {
    ...lastTaskByKey.value,
    [getTaskCacheKey()]: task
  }
}

async function pollTask(taskId) {
  stopPolling()

  try {
    const task = await api.getDatabaseDumpTask(taskId)
    taskState.value = task
    rememberTask(task)

    if (task.status === 'success') {
      gmssh.success(
        props.mode === 'export'
          ? t('dbDump.exportSuccess', { file: task.fileName || task.filePath || '' })
          : t('dbDump.importSuccess', { database: task.database || importOptions.value.database || props.database })
      )
      emit('done', task)
      return
    }

    if (task.status === 'failed') {
      gmssh.error(task.error || t('dbDump.taskFailed'))
      emit('done', task)
      return
    }

    if (task.status === 'canceled') {
      gmssh.warning(t('dbDump.taskCanceled'))
      emit('done', task)
      return
    }

    pollingTimer.value = window.setTimeout(() => {
      pollTask(taskId)
    }, 1500)
  } catch (error) {
    gmssh.error(error.message)
  }
}

async function chooseImportFile() {
  const selectedPath = await gmssh.chooseFile('/home/')
  if (selectedPath) {
    importOptions.value.filePath = selectedPath
  }
}

async function chooseExportPath() {
  const selectedPath = await gmssh.chooseFolder(exportOptions.value.targetPath || '/home/')
  if (selectedPath) {
    exportOptions.value.targetPath = selectedPath
  }
}

async function submit() {
  loading.value = true
  try {
    if (props.mode === 'export') {
      if (!selectedExportDatabase.value) {
        gmssh.warning(t('dbDump.databaseRequired'))
        return
      }

      if (!exportOptions.value.targetPath) {
        gmssh.warning(t('dbDump.targetPathRequired'))
        return
      }

      const task = await api.exportDatabase(
        props.connId,
        selectedExportDatabase.value,
        'app',
        exportOptions.value.targetPath,
        {
          mode: exportOptions.value.mode,
          includeRoutines: exportOptions.value.includeRoutines,
          includeTriggers: exportOptions.value.includeTriggers,
          includeEvents: exportOptions.value.includeEvents,
          includeTablespace: exportOptions.value.includeTablespace
        }
      )
      rememberTask(task)
      emit('done', task)
      visible.value = false
      //message.info(t('dbDump.taskCreated'))
      $gm.message.info(t('dbDump.taskCreated'));
      return
    }

    if (!importOptions.value.filePath) {
      //message.warning(t('dbDump.filePathRequired'))
      $gm.message.warning(t('dbDump.filePathRequired'));
      return
    }

    // 替换策略需二次确认（会 DROP DATABASE）
    if (importOptions.value.strategy === 'replace') {
      const confirmed = await gmssh.confirm({
        title: t('dbDump.replaceConfirmTitle'),
        content: t('dbDump.replaceConfirmContent', { database: importOptions.value.database || props.database }),
        positiveText: t('dbDump.replaceConfirmOk'),
        negativeText: t('common.cancel')
      })
      if (!confirmed) return
    }

    const task = await api.importDatabase(
      props.connId,
      importOptions.value.database,
      importOptions.value.filePath,
      {
        createDatabase: !!importOptions.value.database,
        strategy: importOptions.value.strategy
      }
    )
    rememberTask(task)
    emit('done', task)
    visible.value = false
    gmssh.info(t('dbDump.taskCreated'))
  } catch (error) {
    gmssh.error(error.message)
  } finally {
    loading.value = false
  }
}

async function cancelTask() {
  if (!taskState.value?.id) return

  canceling.value = true
  try {
    const task = await api.cancelDatabaseDumpTask(taskState.value.id)
    taskState.value = task
    pollTask(task.id)
  } catch (error) {
    gmssh.error(error.message)
  } finally {
    canceling.value = false
  }
}


function openOutputFolder() {
  if (!taskState.value?.filePath) return
  gmssh.openFolder(taskState.value.filePath.replace(/\/[^/]+$/, ''))
}
</script>

<style scoped>
.db-dump-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-status-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  border: 1px solid var(--ref-color-white-8);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.04);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    0 12px 30px rgba(0, 0, 0, 0.16);
  backdrop-filter: blur(18px) saturate(126%);
  -webkit-backdrop-filter: blur(18px) saturate(126%);
}

.task-status-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.task-status-title {
  color: var(--sys-color-text-primary);
  font-size: 14px;
  font-weight: 600;
}

.task-status-meta {
  margin-top: 4px;
  color: var(--sys-color-text-secondary);
  font-size: 12px;
}

.task-status-badge {
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  color: var(--sys-color-text-secondary);
  font-size: 11px;
}

.task-status-badge.is-running,
.task-status-badge.is-pending {
  color: #9ed0ff;
}

.task-status-badge.is-success {
  color: #6dd39e;
}

.task-status-badge.is-failed,
.task-status-badge.is-canceled {
  color: #ff9b9b;
}

.task-status-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 12px;
}

.task-status-item {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.task-status-item--path {
  grid-column: 1 / -1;
}

.path-link {
  all: unset;
  cursor: pointer;
  color: #8fc0ff;
  font-size: 13px;
  line-height: 1.5;
  word-break: break-all;
}

.path-link:hover {
  text-decoration: underline;
}

.advanced-options {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 12px;
  margin-top: 12px;
}

.advanced-card {
  padding: 12px 14px;
  border: 1px solid var(--ref-color-white-8);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.03);
}

.advanced-toggle {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--sys-color-text-primary);
  font-size: 13px;
  cursor: pointer;
}

.advanced-toggle svg {
  transition: transform 0.18s ease;
}

.advanced-toggle svg.is-open {
  transform: rotate(180deg);
}

.task-status-item span {
  color: var(--sys-color-text-secondary);
  font-size: 12px;
}

.task-status-item strong {
  color: var(--sys-color-text-primary);
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-status-item .is-error {
  color: #ff9b9b;
}

.path-row {
  display: flex;
  width: 100%;
  gap: 8px;
}

/* ── strategy field + conflict warning ─── */
.strategy-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.strategy-warn {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 12px;
  color: var(--ref-color-yellow-6);
  line-height: 1.5;
  padding: 6px 10px;
  background: rgba(250,173,20,0.06);
  border-radius: 6px;
  border: 1px solid rgba(250,173,20,0.2);
}

/* ── db-name-field: input + inline create hint ── */
.db-name-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.create-db-hint {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  user-select: none;
  width: fit-content;
}

.create-db-hint input[type="checkbox"] {
  width: 14px;
  height: 14px;
  accent-color: var(--ref-color-brand-6);
  cursor: pointer;
  flex-shrink: 0;
  margin: 0;
}

.create-db-hint span {
  font-size: 12px;
  color: var(--sys-color-text-secondary);
  line-height: 1;
}

.create-db-hint:hover span {
  color: var(--sys-color-text-primary);
}

.create-db-hint.is-disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.create-db-hint.is-disabled input { cursor: not-allowed; }

.db-dump-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
