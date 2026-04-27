<template>
  <div class="transfer-center">
    <div class="transfer-toolbar">
      <div class="transfer-toolbar__left">
        <n-button
          type="primary"
          class="transfer-toolbar__button"
          :disabled="!currentConnId"
          @click="openTaskModal('export')"
        >
          {{ t('transfer.newExport') }}
        </n-button>
        <n-button
          class="transfer-toolbar__button"
          :disabled="!currentConnId"
          @click="openTaskModal('import')"
        >
          {{ t('transfer.newImport') }}
        </n-button>
      </div>
      <div class="transfer-toolbar__right">
        <n-tooltip trigger="hover">
          <template #trigger>
            <button class="bar-icon-btn" :disabled="loadingTasks" @click="refreshTasks()">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>
            </button>
          </template>
          {{ t('transfer.refresh') }}
        </n-tooltip>
      </div>
    </div>

    <div class="transfer-shell">
      <section class="transfer-list-panel panel-surface">
        <div class="panel-head">
          <div>
            <div class="panel-title">{{ t('transfer.taskList') }}</div>
          </div>
        </div>

        <div v-if="filteredTasks.length" class="transfer-task-list">
          <button
            v-for="task in filteredTasks"
            :key="task.id"
            type="button"
            class="transfer-task-item"
            :class="{ 'is-active': activeTask?.id === task.id }"
            @click="selectedTaskId = task.id"
          >
            <div class="transfer-task-item__head">
              <div class="transfer-task-item__title">
                <span class="transfer-task-item__name">{{ task.database || t('transfer.unknownDatabase') }}</span>
                <span class="transfer-task-item__type">{{ task.type === 'export' ? t('transfer.exportLabel') : t('transfer.importLabel') }}</span>
              </div>
              <span class="task-badge" :class="`is-${task.status}`">
                {{ getTaskStatusLabel(task.status) }}
              </span>
            </div>

            <div class="transfer-task-item__meta">
              <span>{{ getTaskConnectionName(task) }}</span>
              <span>{{ formatDateTime(task.createdAt) }}</span>
            </div>

            <div class="transfer-task-item__phase">
              {{ formatTaskSummary(task) }}
            </div>
          </button>
        </div>

        <n-empty
          v-else
          class="transfer-empty"
          :description="t('transfer.empty')"
        />
      </section>

      <section class="transfer-detail-panel panel-surface">
        <template v-if="activeTask">
          <div class="panel-head panel-head--detail">
            <div>
              <div class="panel-title">{{ summarizeTask(activeTask) }}</div>
              <div class="panel-subtitle">{{ formatTaskSummary(activeTask) }}</div>
            </div>

            <div class="transfer-detail-actions">
              <n-popover trigger="hover" placement="bottom-end" :show-arrow="false">
                <template #trigger>
                  <n-button quaternary size="small" class="detail-info-button">
                    {{ t('transfer.detailInfo') }}
                  </n-button>
                </template>

                <div class="detail-popover">
                  <div class="detail-pair">
                    <span>{{ t('transfer.detailConnection') }}</span>
                    <strong>{{ getTaskConnectionName(activeTask) }}</strong>
                  </div>
                  <div class="detail-pair">
                    <span>{{ t('transfer.detailDatabase') }}</span>
                    <strong>{{ activeTask.database || t('transfer.unknownDatabase') }}</strong>
                  </div>
                  <div class="detail-pair">
                    <span>{{ t('transfer.detailType') }}</span>
                    <strong>{{ activeTask.type === 'export' ? t('transfer.exportLabel') : t('transfer.importLabel') }}</strong>
                  </div>
                  <div v-if="activeTask.strategy" class="detail-pair">
                    <span>{{ t('transfer.detailStrategy') }}</span>
                    <strong>{{ formatStrategy(activeTask.strategy) }}</strong>
                  </div>
                  <div class="detail-pair">
                    <span>{{ t('transfer.detailPhase') }}</span>
                    <strong>{{ formatTaskPhase(activeTask.phase || 'queued') }}</strong>
                  </div>
                  <div class="detail-pair">
                    <span>{{ t('transfer.detailStartedAt') }}</span>
                    <strong>{{ formatDateTime(activeTask.startedAt || activeTask.createdAt) }}</strong>
                  </div>
                  <div class="detail-pair">
                    <span>{{ t('transfer.detailFinishedAt') }}</span>
                    <strong>{{ formatDateTime(activeTask.finishedAt) }}</strong>
                  </div>
                  <div class="detail-pair">
                    <span>{{ t('transfer.detailFileSize') }}</span>
                    <strong>{{ formatBytes(activeTask.size || activeTask.totalBytes) }}</strong>
                  </div>
                  <div v-if="activeTask.fileName" class="detail-pair detail-pair--wide">
                    <span>{{ t('transfer.detailFileName') }}</span>
                    <strong>{{ activeTask.fileName }}</strong>
                  </div>
                  <div v-if="activeTask.error" class="detail-pair detail-pair--wide">
                    <span>{{ t('transfer.detailError') }}</span>
                    <strong class="detail-error">{{ activeTask.error }}</strong>
                  </div>
                </div>
              </n-popover>
              <n-button
                v-if="canCancel(activeTask)"
                quaternary
                type="warning"
                @click="cancelTask(activeTask)"
              >
                {{ t('transfer.cancelTask') }}
              </n-button>
            </div>
          </div>

          <div v-if="activeTask.filePath" class="transfer-path-bar">
            <span class="transfer-path-label">{{ t('transfer.detailPath') }}</span>
            <strong class="transfer-path-value">{{ activeTask.filePath }}</strong>
            <n-button
              v-if="showFolderLink(activeTask)"
              quaternary
              size="small"
              @click="openTaskFolder(activeTask)"
            >
              {{ t('transfer.openFolder') }}
            </n-button>
          </div>

          <div class="panel-head panel-head--logs">
            <div>
              <div class="panel-title">{{ t('transfer.logsTitle') }}</div>
            </div>
          </div>

          <div v-if="taskLogs.length" class="transfer-log-list">
            <div v-for="log in taskLogs" :key="log.id" class="transfer-log-item">
              <span class="log-time">{{ formatDateTime(log.createdAt) }}</span>
              <span class="log-level" :class="`is-${log.level}`">{{ formatLogLevel(log.level) }}</span>
              <span class="log-phase">{{ formatTaskPhase(log.phase || 'task') }}</span>
              <span class="log-message">{{ formatTaskMessage(log.message, log.phase) }}</span>
            </div>
          </div>
          <n-empty v-else class="transfer-empty transfer-empty--logs" :description="t('transfer.logsEmpty')" />
        </template>

        <n-empty
          v-else
          class="transfer-empty transfer-empty--detail"
          :description="t('transfer.detailEmpty')"
        />
      </section>
    </div>

    <DatabaseDumpModal
      v-model="dumpVisible"
      :mode="dumpMode"
      :conn-id="currentConnId"
      :database="currentDatabase"
      @done="handleTaskDone"
    />
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import DatabaseDumpModal from '@/components/DatabaseDumpModal.vue'
import { useDatabaseStore } from '@/stores/database'
import api from '@/utils/api'
import gmssh from '@/utils/gmssh'

const consumedAutoLaunchKeys = new Set()

const props = defineProps({
  connId: { type: String, default: '' },
  initialDatabase: { type: String, default: '' },
  initialAction: { type: String, default: '' },
  autoLaunch: { type: Boolean, default: false },
  launchKey: { type: [String, Number], default: 0 }
})
const emit = defineEmits(['auto-launch-consumed'])

const { t } = useI18n()
const store = useDatabaseStore()

const tasks = ref([])
const taskLogs = ref([])
const selectedTaskId = ref('')
const loadingTasks = ref(false)
const dumpVisible = ref(false)
const dumpMode = ref('export')
const pollingTimer = ref(null)

const currentConnId = computed(() => props.connId || store.currentConnId || '')
const currentDatabase = computed(() => {
  if (!currentConnId.value) return ''

  if (currentConnId.value === store.currentConnId) {
    return store.selectedDatabase || props.initialDatabase || ''
  }

  return props.initialDatabase || store.getConnectionState(currentConnId.value)?.selectedDatabase || ''
})

const selectedConfigId = computed(() => {
  if (!currentConnId.value) return ''
  const activeConnection = (store.connections || []).find((item) => item.connId === currentConnId.value)
  return activeConnection?.config?.id || ''
})

const filteredTasks = computed(() => {
  return (tasks.value || []).filter((task) => {
    if (selectedConfigId.value) {
      if (task.connectionConfigId) {
        if (task.connectionConfigId !== selectedConfigId.value) {
          return false
        }
      } else if (currentConnId.value && task.connId !== currentConnId.value) {
        return false
      }
    } else if (currentConnId.value && task.connId !== currentConnId.value) {
      return false
    }

    if (currentDatabase.value && task.database !== currentDatabase.value) {
      return false
    }

    return true
  })
})

const activeTask = computed(() => {
  if (!filteredTasks.value.length) return null
  return filteredTasks.value.find((task) => task.id === selectedTaskId.value) || filteredTasks.value[0]
})

const runningTaskCount = computed(() => filteredTasks.value.filter((task) => ['pending', 'running'].includes(task.status)).length)
watch(
  () => activeTask.value?.id,
  (taskId) => {
    selectedTaskId.value = taskId || ''
    loadTaskLogs(taskId)
  },
  { immediate: true }
)

watch(
  () => currentConnId.value,
  async (connId) => {
    if (!connId) return
    await ensureDatabasesLoaded(connId)
  },
  { immediate: true }
)

watch(
  () => [props.launchKey, props.connId, props.initialDatabase, props.initialAction, props.autoLaunch],
  async () => {
    const nextConnId = props.connId || store.currentConnId || store.connections[0]?.connId || ''

    if (nextConnId) {
      await ensureDatabasesLoaded(nextConnId)
    }

    const launchToken = `${props.connId || nextConnId}::${props.initialAction}::${props.launchKey}`
    if (props.autoLaunch && props.initialAction && nextConnId && !consumedAutoLaunchKeys.has(launchToken)) {
      consumedAutoLaunchKeys.add(launchToken)
      openTaskModal(props.initialAction)
      emit('auto-launch-consumed')
    }
  },
  { immediate: true }
)

onMounted(() => {
  refreshTasks()
})

onBeforeUnmount(() => {
  stopPolling()
})

function stopPolling() {
  if (pollingTimer.value) {
    window.clearTimeout(pollingTimer.value)
    pollingTimer.value = null
  }
}

function schedulePolling() {
  stopPolling()

  if (!tasks.value.some((task) => ['pending', 'running'].includes(task.status))) {
    return
  }

  pollingTimer.value = window.setTimeout(() => {
    refreshTasks()
  }, 1800)
}

async function ensureDatabasesLoaded(connId) {
  const state = store.getConnectionState(connId)
  if (state?.databases?.length) {
    return state.databases
  }

  try {
    return await store.loadDatabases(connId)
  } catch (error) {
    gmssh.error(error.message)
    return []
  }
}

async function refreshTasks() {
  loadingTasks.value = true

  try {
    const result = await api.listDatabaseDumpTasks()
    tasks.value = Array.isArray(result) ? result : []

    if (!tasks.value.some((task) => task.id === selectedTaskId.value)) {
      selectedTaskId.value = filteredTasks.value[0]?.id || tasks.value[0]?.id || ''
    }
  } catch (error) {
    gmssh.error(error.message)
  } finally {
    loadingTasks.value = false
    schedulePolling()
  }
}

async function loadTaskLogs(taskId) {
  if (!taskId) {
    taskLogs.value = []
    return
  }

  try {
    const result = await api.getDatabaseDumpTaskLogs(taskId)
    taskLogs.value = Array.isArray(result) ? result : []
  } catch (error) {
    taskLogs.value = []
    gmssh.error(error.message)
  }
}

function openTaskModal(mode) {
  dumpMode.value = mode
  dumpVisible.value = true
}

async function handleTaskDone(task) {
  await refreshTasks()

  if (task?.id) {
    selectedTaskId.value = task.id
    await loadTaskLogs(task.id)
  }

  if (task?.type === 'import' && currentConnId.value) {
    await ensureDatabasesLoaded(currentConnId.value)
    window.dispatchEvent(new CustomEvent('database-dump-finished', {
      detail: {
        mode: 'import',
        connId: currentConnId.value,
        database: task.database || currentDatabase.value || ''
      }
    }))
  }
}

async function cancelTask(task) {
  if (!task?.id) return

  try {
    const nextTask = await api.cancelDatabaseDumpTask(task.id)
    selectedTaskId.value = nextTask?.id || task.id
    await refreshTasks()
    await loadTaskLogs(selectedTaskId.value)
  } catch (error) {
    gmssh.error(error.message)
  }
}

function canCancel(task) {
  return ['pending', 'running'].includes(task?.status)
}

function getConnectionName(connId) {
  const item = (store.connections || []).find((entry) => entry.connId === connId)
  return item?.config?.name || connId || '—'
}

function getTaskConnectionName(task) {
  if (!task) return '—'
  if (task.connectionName) return task.connectionName
  return getConnectionName(task.connId)
}

function getTaskStatusLabel(status) {
  if (status === 'pending') return t('dbDump.statusPending')
  if (status === 'running') return t('dbDump.statusRunning')
  if (status === 'success') return t('dbDump.statusSuccess')
  if (status === 'failed') return t('dbDump.statusFailed')
  if (status === 'canceled') return t('dbDump.statusCanceled')
  if (status === 'interrupted') return t('dbDump.statusInterrupted')
  return formatTaskPhase(status || 'task')
}

function summarizeTask(task) {
  if (!task) return ''
  const label = task.type === 'export' ? t('transfer.exportLabel') : t('transfer.importLabel')
  return `${label} · ${task.database || t('transfer.unknownDatabase')}`
}

function formatTaskSummary(task) {
  if (!task) return '—'
  return formatTaskMessage(task.message, task.phase)
}

function formatTaskPhase(phase) {
  const normalized = String(phase || '').trim()
  if (!normalized) return '—'
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

  return formatTaskPhase(phase || normalized || 'task')
}

function formatLogLevel(level) {
  const normalized = String(level || '').trim().toLowerCase()
  if (!normalized) return '—'
  const translated = t(`transfer.logLevel.${normalized}`)
  return translated === `transfer.logLevel.${normalized}` ? normalized : translated
}

function formatStrategy(strategy) {
  if (strategy === 'target') return t('dbDump.importStrategyTarget')
  if (strategy === 'replace') return t('dbDump.importStrategyReplace')
  return t('dbDump.importStrategySource')
}

function formatDateTime(value) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function formatBytes(value) {
  const size = Number(value || 0)
  if (!size) return '—'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`
}

function showFolderLink(task) {
  return task?.type === 'export' && task?.target === 'app' && !!task?.filePath
}

function openTaskFolder(task) {
  const targetPath = task?.filePath?.replace(/\/[^/]+$/, '')
  if (!targetPath) return

  if (!gmssh.openFolder(targetPath)) {
    gmssh.warning(t('transfer.openFolderUnavailable'))
  }
}
</script>

<style scoped>
.transfer-center {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
}

.transfer-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.transfer-toolbar__left,
.transfer-toolbar__right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.transfer-toolbar__button {
  min-width: 92px;
}

.panel-surface {
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.04);
  border-radius: 10px;
  min-height: 0;
  overflow: hidden;
}

.transfer-shell {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(300px, 380px) minmax(0, 1fr);
  gap: 12px;
}

.transfer-list-panel,
.transfer-detail-panel {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 18px;
  border-bottom: 1px solid var(--ref-color-white-8);
}

.panel-head--detail,
.panel-head--logs {
  border-bottom: none;
  padding-bottom: 10px;
}

.panel-title {
  color: var(--sys-color-text-primary);
  font-size: 14px;
  font-weight: 600;
}

.panel-subtitle {
  margin-top: 4px;
  color: var(--sys-color-text-secondary);
  font-size: 12px;
}

.transfer-task-list,
.transfer-log-list {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 12px;
}

.transfer-task-item {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 14px;
  margin-bottom: 8px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.03);
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.18s ease, transform 0.18s ease, background-color 0.18s ease, box-shadow 0.18s ease;
}

.transfer-task-item:hover,
.transfer-task-item.is-active {
  border-color: rgba(82, 104, 239, 0.28);
  background: rgba(82, 104, 239, 0.08);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.22);
}

.transfer-task-item__head,
.transfer-task-item__title,
.transfer-task-item__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.transfer-task-item__title {
  justify-content: flex-start;
}

.transfer-task-item__name {
  color: var(--sys-color-text-primary);
  font-size: 14px;
  font-weight: 600;
}

.transfer-task-item__type,
.transfer-task-item__meta,
.transfer-task-item__phase {
  color: var(--sys-color-text-secondary);
  font-size: 12px;
}

.transfer-task-item__meta {
  justify-content: flex-start;
}

.transfer-task-item__phase {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.task-badge {
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  color: var(--sys-color-text-secondary);
  font-size: 11px;
  white-space: nowrap;
}

.task-badge.is-pending,
.task-badge.is-running {
  color: #8fc0ff;
}

.task-badge.is-success {
  color: #74d3a0;
}

.task-badge.is-failed,
.task-badge.is-canceled {
  color: #ff9d9d;
}

.transfer-detail-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.detail-info-button {
  min-width: 0;
}

.transfer-path-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 18px 10px;
  padding: 10px 12px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.03);
}

.transfer-path-label {
  flex-shrink: 0;
  color: var(--sys-color-text-secondary);
  font-size: 12px;
}

.transfer-path-value {
  flex: 1;
  min-width: 0;
  color: var(--sys-color-text-primary);
  font-size: 13px;
  font-weight: 500;
  word-break: break-all;
}

.task-progress :deep(.n-progress-graph-line-rail) {
  background: rgba(255, 255, 255, 0.06) !important;
  border: 1px solid rgba(255, 255, 255, 0.04);
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.18);
}

.task-progress :deep(.n-progress-graph-line-fill) {
  background: #5268EF !important;
  box-shadow: 0 0 14px rgba(82, 104, 239, 0.18);
}

.task-progress :deep(.n-progress-content__text) {
  color: var(--sys-color-text-secondary) !important;
  font-size: 12px;
  font-weight: 600;
}

.detail-pair {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  padding-bottom: 8px;
}

.detail-pair--wide {
  grid-column: 1 / -1;
}

.detail-pair span {
  color: var(--sys-color-text-secondary);
  font-size: 12px;
}

.detail-pair strong {
  color: var(--sys-color-text-primary);
  font-size: 13px;
  font-weight: 500;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.detail-path {
  word-break: break-all;
}

.detail-popover {
  width: min(460px, 60vw);
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 16px;
}

.detail-error {
  color: #ff9d9d !important;
  white-space: pre-wrap;
}

.transfer-log-item {
  display: grid;
  grid-template-columns: 168px 68px 96px minmax(0, 1fr);
  gap: 10px;
  align-items: start;
  padding: 10px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.transfer-log-item:last-child {
  border-bottom: none;
}

.log-time,
.log-phase,
.log-message {
  color: var(--sys-color-text-secondary);
  font-size: 12px;
}

.log-level {
  display: inline-flex;
  justify-content: center;
  min-width: 56px;
  padding: 3px 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  color: var(--sys-color-text-secondary);
  font-size: 11px;
  text-transform: uppercase;
}

.log-level.is-info {
  color: #8fc0ff;
}

.log-level.is-warning {
  color: #ffd27d;
}

.log-level.is-error {
  color: #ff9d9d;
}

.transfer-empty {
  margin: auto;
}

.transfer-empty--logs,
.transfer-empty--detail {
  padding: 24px 0;
}

@media (max-width: 1280px) {
  .transfer-shell {
    grid-template-columns: minmax(280px, 320px) minmax(0, 1fr);
  }
}
</style>
