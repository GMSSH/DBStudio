<template>
  <div class="db-overview">
    <div class="overview-toolbar">
      <div class="overview-toolbar-left">
        <n-input
          v-model:value="keyword"
          clearable
          size="small"
          class="overview-search"
          :placeholder="t('overview.searchPlaceholder')"
        />
        <div class="overview-meta">
          <span>{{ t('overview.objectCount', { n: filteredObjects.length }) }}</span>
        </div>
      </div>
      <div class="overview-toolbar-actions">
        <n-tooltip trigger="hover">
          <template #trigger>
            <button class="bar-icon-btn" @click="emit('refresh')">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>
            </button>
          </template>
          {{ t('overview.refresh') }}
        </n-tooltip>
      </div>
    </div>

    <div class="overview-table-panel">
      <n-spin :show="loading">
        <template v-if="filteredObjects.length > 0">
          <div class="overview-table-shell">
            <div class="overview-table-wrap overview-table-wrap--main">
              <table class="overview-table">
                <colgroup>
                  <col class="col-name" />
                  <col class="col-type" />
                  <col class="col-rows" />
                  <col class="col-engine" />
                  <col class="col-size" />
                  <col class="col-updated" />
                  <col class="col-comment" />
                  <col class="col-actions" />
                </colgroup>
                <thead>
                  <tr>
                    <th class="gm-pth overview-name-col">{{ t('overview.tableName') }}</th>
                    <th class="gm-pth">{{ t('overview.type') }}</th>
                    <th class="gm-pth">{{ t('overview.rows') }}</th>
                    <th class="gm-pth">{{ t('overview.engine') }}</th>
                    <th class="gm-pth">{{ t('overview.size') }}</th>
                    <th class="gm-pth">{{ t('overview.updatedAt') }}</th>
                    <th class="gm-pth overview-comment-col">{{ t('overview.comment') }}</th>
                    <th class="gm-pth">{{ t('overview.actions') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="object in filteredObjects"
                    :key="`main-${object.name}`"
                    :class="[
                      'gm-ptr',
                      `overview-row--${normalizeObjectType(object.type)}`
                    ]"
                    @dblclick="emit('open-data', object)"
                  >
                    <td class="gm-ptd overview-name-cell">
                      <div class="object-main">
                        <span class="object-name">{{ object.name }}</span>
                      </div>
                    </td>
                    <td class="gm-ptd">
                      <span class="cell-text">{{ objectTypeLabel(object.type) }}</span>
                    </td>
                    <td class="gm-ptd"><span class="cell-text">{{ formatRowCount(object) }}</span></td>
                    <td class="gm-ptd"><span class="cell-text">{{ object.engine || t('overview.unknown') }}</span></td>
                    <td class="gm-ptd"><span class="cell-text">{{ formatSize(object.size) }}</span></td>
                    <td class="gm-ptd"><span class="cell-text">{{ object.updatedAt || t('overview.unknown') }}</span></td>
                    <td class="gm-ptd overview-comment-cell">
                      <span class="overview-comment" :title="object.comment || ''">
                        {{ object.comment || t('overview.unknown') }}
                      </span>
                    </td>
                    <td class="gm-ptd overview-actions-cell">
                      <n-space :size="4" justify="start">
                        <n-button text size="tiny" @click.stop="emit('open-data', object)">
                          {{ t('overview.openData') }}
                        </n-button>
                        <n-button text size="tiny" @click.stop="emit('open-structure', object)">
                          {{ t('overview.openStructure') }}
                        </n-button>
                        <n-button text size="tiny" @click.stop="previewDDL(object)">
                          {{ t('overview.previewDDL') }}
                        </n-button>
                        <n-dropdown
                          trigger="click"
                          :options="getObjectMenuOptions(object)"
                          @select="(key) => handleObjectMenu(key, object)"
                        >
                          <n-button text size="tiny" @click.stop>
                            {{ t('overview.more') }}
                          </n-button>
                        </n-dropdown>
                      </n-space>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </template>
        <n-empty
          v-else
          class="overview-empty"
          :description="keyword ? t('overview.emptyFiltered') : t('overview.empty')"
        />
      </n-spin>
    </div>

    <n-drawer v-model:show="ddlDrawerVisible" width="560" placement="right" class="overview-ddl-drawer">
      <n-drawer-content
        :title="t('overview.ddlTitle', { name: ddlTargetName })"
        closable
        class="overview-ddl-content"
      >
        <n-spin :show="ddlLoading">
          <pre class="ddl-preview">{{ ddlText || t('overview.ddlEmpty') }}</pre>
        </n-spin>
        <template #footer>
          <n-space justify="end">
            <n-button @click="ddlDrawerVisible = false">{{ t('common.close') }}</n-button>
            <n-button type="primary" :disabled="!ddlText" @click="copyText(ddlText)">
              {{ t('overview.copyDDL') }}
            </n-button>
          </n-space>
        </template>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/utils/api'
import gmssh from '@/utils/gmssh'

const props = defineProps({
  connId: { type: String, required: true },
  database: { type: String, required: true },
  cache: {
    type: Object,
    default: () => ({
      loaded: false,
      loading: false,
      objects: []
    })
  }
})

const emit = defineEmits(['refresh', 'open-data', 'open-structure', 'open-export', 'open-import'])

const { t } = useI18n()

const keyword = ref('')
const ddlDrawerVisible = ref(false)
const ddlLoading = ref(false)
const ddlText = ref('')
const ddlTargetName = ref('')

const loading = computed(() => !!props.cache?.loading)
const objects = computed(() => Array.isArray(props.cache?.objects) ? props.cache.objects : [])
const hasLoaded = computed(() => !!props.cache?.loaded)

const filteredObjects = computed(() => {
  const search = keyword.value.trim().toLowerCase()
  if (!search) return objects.value

  return objects.value.filter((object) => {
    const haystack = [
      object.name,
      object.comment,
      object.engine,
      normalizeObjectType(object.type)
    ]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()

    return haystack.includes(search)
  })
})

function normalizeObjectType(type) {
  return type === 'view' ? 'view' : 'table'
}

function objectTypeLabel(type) {
  return normalizeObjectType(type) === 'view' ? t('overview.objectView') : t('overview.objectTable')
}

function formatSize(size) {
  if (!size) return t('overview.unknown')
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`
}

function formatRowCount(object) {
  if (normalizeObjectType(object.type) === 'view') {
    return t('overview.unknown')
  }

  if (typeof object.rowCount !== 'number') {
    return t('overview.unknown')
  }

  return new Intl.NumberFormat().format(object.rowCount)
}

function getObjectMenuOptions(object) {
  return [
    { label: t('overview.exportData'), key: 'export' },
    normalizeObjectType(object.type) === 'table' ? { label: t('overview.importData'), key: 'import' } : null,
    { label: t('overview.copyDDL'), key: 'copy-ddl' }
  ].filter(Boolean)
}

function handleObjectMenu(key, object) {
  if (key === 'export') {
    emit('open-export', object)
    return
  }

  if (key === 'import') {
    emit('open-import', object)
    return
  }

  if (key === 'copy-ddl') {
    copyDDL(object)
  }
}

async function fetchDDL(object) {
  if (!object?.name) return ''
  const result = await api.getTableDDL(props.connId, props.database, object.name)
  return typeof result === 'string' ? result : (result?.ddl || '')
}

async function previewDDL(object) {
  ddlTargetName.value = object?.name || ''
  ddlText.value = ''
  ddlDrawerVisible.value = true
  ddlLoading.value = true
  try {
    ddlText.value = await fetchDDL(object)
  } catch (error) {
    gmssh.error(error.message)
  } finally {
    ddlLoading.value = false
  }
}

async function copyText(text) {
  if (!text) return

  const copied = await gmssh.copyToClipboard(text)
  if (copied) {
    gmssh.success(t('overview.ddlCopied'))
    return
  }

  gmssh.warning(t('designer.copyUnavailable'))
}

async function copyDDL(object) {
  if (!object?.name) return
  try {
    const ddl = await fetchDDL(object)
    await copyText(ddl)
  } catch (error) {
    gmssh.error(error.message)
  }
}
watch(
  () => [props.connId, props.database],
  () => {
    keyword.value = ''
  },
  { immediate: true }
)

watch(
  () => [props.connId, props.database, hasLoaded.value, loading.value],
  () => {
    if (!hasLoaded.value && !loading.value) {
      emit('refresh')
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.db-overview {
  display: flex;
  flex-direction: column;
  gap: var(--ref-space-16);
  min-height: 100%;
  padding-bottom: 12px;
}

.overview-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--ref-space-16);
  padding: 0 2px;
}

.overview-toolbar-left,
.overview-toolbar-actions {
  display: flex;
  align-items: center;
  gap: var(--ref-space-10);
}

.overview-toolbar-left {
  min-width: 0;
  flex-wrap: nowrap;
}

.overview-search {
  flex: 0 0 320px;
  max-width: 320px;
}

.overview-meta {
  display: inline-flex;
  align-items: center;
  flex-wrap: nowrap;
  white-space: nowrap;
  gap: var(--ref-space-8);
  color: var(--sys-color-text-secondary);
  font-size: var(--ref-font-size-xs);
  flex-shrink: 0;
}

.overview-table-panel {
  min-height: 0;
  flex: none;
  overflow: visible;
}

.overview-table-panel :deep(.n-spin-container),
.overview-table-panel :deep(.n-spin-content) {
  height: auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.overview-table-shell {
  height: auto;
  border: none;
  border-radius: 0;
  background: transparent;
  overflow: hidden;
}

.overview-table-wrap {
  height: auto;
  background: transparent;
}

.overview-table-wrap--main {
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: thin;
  scrollbar-color: var(--ref-color-white-12) transparent;
}

.overview-table-wrap--main::-webkit-scrollbar {
  width: var(--scroll-size-default);
  height: var(--scroll-size-default);
}

.overview-table-wrap--main::-webkit-scrollbar-track {
  background: var(--scroll-track-bg);
}

.overview-table-wrap--main::-webkit-scrollbar-thumb {
  background: var(--scroll-thumb-bg);
  border-radius: var(--scroll-thumb-radius);
}

.overview-table-wrap--main::-webkit-scrollbar-thumb:hover {
  background: var(--scroll-thumb-bg-hover);
}

.overview-table {
  width: max(100%, 1220px);
  min-width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  font-family: var(--ref-font-family-base);
  font-size: var(--ref-font-size-sm);
  line-height: 1.5;
}

.overview-table .col-name {
  width: 180px;
}

.overview-table .col-type {
  width: 80px;
}

.overview-table .col-rows {
  width: 92px;
}

.overview-table .col-engine {
  width: 112px;
}

.overview-table .col-size {
  width: 104px;
}

.overview-table .col-updated {
  width: 156px;
}

.overview-table .col-comment {
  width: 220px;
}

.overview-table .col-actions {
  width: 184px;
}



.overview-name-cell,
.overview-comment-cell {
  max-width: 0;
}

.object-main {
  display: flex;
  align-items: center;
  gap: var(--ref-space-8);
  min-width: 0;
}

.object-name,
.overview-comment {
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.5;
}

.cell-text {
  display: inline-block;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.5;
}

.overview-comment {
  width: 100%;
  color: var(--sys-color-text-secondary);
  vertical-align: middle;
}

.type-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 20px;
  padding: 0 var(--ref-space-8);
  border-radius: var(--ref-radius-pill);
  font-size: var(--ref-font-size-xs);
  font-weight: var(--ref-font-weight-medium);
  line-height: 1;
}

.type-badge--table {
  background: rgba(87, 114, 255, 0.14);
  color: var(--ref-color-brand-4);
}

.type-badge--view {
  background: rgba(15, 198, 194, 0.14);
  color: var(--ref-color-cyan-6);
}

.overview-empty {
  margin-top: 64px;
}

.ddl-preview {
  min-height: 320px;
  max-height: calc(100vh - 180px);
  margin: 0;
  padding: var(--ref-space-14);
  overflow: auto;
  border: 1px solid var(--depth-1-border);
  border-radius: var(--ref-radius-lg);
  background: rgba(0, 0, 0, 0.3);
  color: var(--sys-color-text-primary);
  font-family: var(--ref-font-family-mono);
  font-size: var(--ref-font-size-xs);
  line-height: 1.7;
  white-space: pre-wrap;
}

.overview-ddl-drawer :deep(.n-drawer-content) {
  background: var(--sys-color-bg-container);
}

.overview-ddl-content :deep(.n-drawer-header) {
  border-bottom: 1px solid var(--depth-1-border);
}

.overview-ddl-content :deep(.n-drawer-body),
.overview-ddl-content :deep(.n-drawer-footer) {
  background: var(--sys-color-bg-container);
}

@media (max-width: 860px) {
  .overview-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .overview-toolbar-left,
  .overview-toolbar-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .overview-search {
    flex-basis: auto;
    max-width: none;
  }

  .overview-table .col-name {
    width: 140px;
  }

  .overview-table .col-comment {
    width: 200px;
  }

  .overview-table-shell {
    overflow-x: auto;
  }
}
</style>
