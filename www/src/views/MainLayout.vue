<template>
  <n-layout has-sider class="main-layout" style="flex:1;min-height:0;width:100%;display:flex;">
    <!-- Left Sidebar: glass panel per app-window.md -->
    <n-layout-sider
      class="sidebar"
      :width="SIDEBAR_WIDTH"
      content-style="padding:0;height:100%;overflow:hidden;"
    >
      <!-- Native scroll wrapper — replaces n-scrollbar to guarantee overflow works -->
      <div class="sidebar-scroll">
        <DatabaseTree @add-connection="handleNewConnection" @edit-connection="handleEditConnection" />
      </div>
    </n-layout-sider>

    <!-- Main Content Area -->
    <n-layout class="content-layout">
      <!-- Titlebar / Toolbar: glass divider from sidebar -->
      <n-layout-header class="toolbar-header">
        <div class="toolbar-capsule">
          <!-- New Connection -->
          <n-button type="primary" size="small" @click="handleNewConnection" class="toolbar-btn" id="btn-new-connection" :bordered="false">
            <template #icon>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="3"/>
                <path d="M12 3v6M12 15v6M3 12h6M15 12h6"/>
              </svg>
            </template>
            {{ t('toolbar.connect') }}
          </n-button>

          <!-- New Query -->
          <n-button size="small" @click="handleNewQuery" class="toolbar-btn ghost" id="btn-new-query" :bordered="false">
            <template #icon>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="4 17 10 11 4 5"/>
                <line x1="12" y1="19" x2="20" y2="19"/>
              </svg>
            </template>
            {{ t('toolbar.query') }}
          </n-button>

          <n-button size="small" @click="handleOpenTransferCenter" class="toolbar-btn ghost" id="btn-transfer-center" :bordered="false">
            <template #icon>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="17 8 12 3 7 8"/>
                <line x1="12" y1="3" x2="12" y2="15"/>
              </svg>
            </template>
            {{ t('toolbar.transfer') }}
          </n-button>
        </div>

        <div class="toolbar-right-cluster">
          <n-select
            v-if="showConnectionSwitcher"
            :value="store.currentConnId"
            :options="connectionOptions"
            size="small"
            class="connection-switcher"
            :placeholder="t('toolbar.switchConnection')"
            @update:value="handleSwitchConnection"
          />
        </div>
      </n-layout-header>

      <!-- Work Content -->
      <n-layout-content class="main-content">
        <WorkArea />
      </n-layout-content>
    </n-layout>

    <!-- New Database Dialog -->
    <FormModal
      v-model="showNewDbDialog"
      :title="t('newDatabase.title')"
      :width="420"
      id="modal-new-db"
    >
      <n-form>
        <n-form-item :label="t('newDatabase.nameLabel')">
          <n-input v-model:value="newDbName" :placeholder="t('newDatabase.namePlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('newDatabase.charsetLabel')">
          <n-select v-model:value="newDbCharset" :options="charsetOptions" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showNewDbDialog = false">{{ t('newDatabase.cancel') }}</n-button>
          <n-button type="primary" @click="handleCreateDatabase">{{ t('newDatabase.create') }}</n-button>
        </n-space>
      </template>
    </FormModal>

    <!-- New / Edit Connection Dialog -->
    <FormModal
      v-model="showNewConnDialog"
      :title="editingConn ? t('connection.editTitle') : t('connection.newTitle')"
      :width="500"
      :mask-closable="false"
      :close-on-esc="false"
      id="modal-new-conn"
    >
      <ConnectionForm
        ref="newConnectionFormRef"
        :hide-actions="true"
        :initial-data="newConnFormData"
      />
      <template #footer>
        <n-space justify="end">
          <n-button @click="showNewConnDialog = false">{{ t('connection.cancel') }}</n-button>
          <n-button @click="handleTestNewConnection" :loading="testingConn">{{ t('connection.testConn') }}</n-button>
          <n-button type="primary" @click="handleSaveAndConnectNew" :loading="savingConn" strong>{{ t('connection.saveAndConnect') }}</n-button>
        </n-space>
      </template>
    </FormModal>
  </n-layout>
</template>

<script setup>
import { computed, ref, defineAsyncComponent, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDatabaseStore } from '@/stores/database'
import DatabaseTree from '@/components/DatabaseTree.vue'
import FormModal from '@/components/FormModal.vue'
import WorkArea from '@/components/WorkArea.vue'
const ConnectionForm = defineAsyncComponent(() => import('@/components/ConnectionForm.vue'))
import connectionApi from '@/utils/connectionApi'
import api from '@/utils/api'
import gmssh from '@/utils/gmssh'

const { t } = useI18n()
const store = useDatabaseStore()

const SIDEBAR_WIDTH = 240
const showConnectionSwitcher = computed(() => store.connections.length > 1)
const connectionOptions = computed(() => (
  store.connections.map((item) => ({
    label: `${item.config?.name || `${item.config?.dbType || 'db'}@${item.config?.host || 'local'}`} · ${(item.config?.dbType || 'db').toUpperCase()}`,
    value: item.connId
  }))
))

// New database dialog
const showNewDbDialog = ref(false)
const newDbName = ref('')
const newDbCharset = ref('utf8mb4')
const charsetOptions = [
  { label: 'utf8mb4', value: 'utf8mb4' },
  { label: 'utf8',    value: 'utf8'    },
  { label: 'latin1',  value: 'latin1'  },
  { label: 'gbk',     value: 'gbk'     }
]

// Connection dialog
const showNewConnDialog = ref(false)
const editingConn = ref(false)
const newConnectionFormRef = ref(null)
const testingConn = ref(false)
const savingConn = ref(false)
const newConnFormData = ref({
  id: '', name: '', dbType: 'mysql',
  host: 'localhost', port: 3306,
  username: 'root', password: '',
  database: '', filePath: ''
})

const handleNewConnection = () => {
  newConnFormData.value = {
    id: '', name: '', dbType: 'mysql',
    host: 'localhost', port: 3306,
    username: 'root', password: '',
    database: '', filePath: ''
  }
  editingConn.value = false
  showNewConnDialog.value = true
}

const handleEditConnection = async (connData) => {
  try {
    const fullData = await connectionApi.getConnection(connData.id)
    newConnFormData.value = { ...fullData }
    editingConn.value = true
    showNewConnDialog.value = true
  } catch (error) {
    gmssh.error(t('connection.getInfoFailed', { msg: error.message }))
  }
}

const handleTestNewConnection = async () => {
  if (!newConnectionFormRef.value) return
  testingConn.value = true
  try {
    await api.testConnection(newConnectionFormRef.value.formModel)
    gmssh.success(t('connection.testSuccess'))
  } catch (error) {
    gmssh.error(t('connection.testFailed', { msg: error.message }))
  } finally {
    testingConn.value = false
  }
}

const handleSaveAndConnectNew = async () => {
  if (!newConnectionFormRef.value) return
  savingConn.value = true
  try {
    const formData = newConnectionFormRef.value.formModel
    const result = await connectionApi.saveConnection(formData)
    formData.id = result.id
    await store.connect(formData)
    window.dispatchEvent(new CustomEvent('refresh-connections'))
    gmssh.success(t('connection.saveConnectSuccess'))
    showNewConnDialog.value = false
  } catch (error) {
    gmssh.error(t('connection.saveConnectFailed', { msg: error.message }))
  } finally {
    savingConn.value = false
  }
}

const handleNewQuery = () => {
  if (!store.isConnected) {
    gmssh.warning(t('connection.noConnected'))
    return
  }
  window.dispatchEvent(new CustomEvent('switch-to-sql-tab'))
}

const handleOpenTransferCenter = () => {
  if (!store.isConnected) {
    gmssh.warning(t('connection.noConnected'))
    return
  }

  window.dispatchEvent(new CustomEvent('open-transfer-center', {
    detail: {
      connId: store.currentConnId,
      database: store.selectedDatabase || '',
      action: '',
      autoLaunch: false
    }
  }))
}

const handleCreateDatabase = async () => {
  if (!newDbName.value.trim()) {
    gmssh.warning(t('newDatabase.nameRequired'))
    return
  }
  try {
    await store.createDatabase(newDbName.value)
    gmssh.success(t('newDatabase.createSuccess', { name: newDbName.value }))
    showNewDbDialog.value = false
    await store.loadDatabases()
  } catch (error) {
    gmssh.error(t('newDatabase.createFailed', { msg: error.message }))
  }
}

const handleOpenCreateDatabase = async (event) => {
  const { connId } = event.detail || {}

  if (connId && connId !== store.currentConnId) {
    try {
      await store.switchConnection(connId)
    } catch (error) {
      gmssh.error(t('layout.switchConnFailed', { msg: error.message }))
      return
    }
  }

  if (!store.isConnected) {
    gmssh.warning(t('connection.noConnected'))
    return
  }

  newDbName.value = ''
  newDbCharset.value = 'utf8mb4'
  showNewDbDialog.value = true
}

const handleSwitchConnection = async (connId) => {
  if (!connId || connId === store.currentConnId) return

  try {
    await store.switchConnection(connId)
  } catch (error) {
    gmssh.error(t('layout.switchConnFailed', { msg: error.message }))
  }
}

onMounted(() => {
  window.addEventListener('open-create-database', handleOpenCreateDatabase)
})

onUnmounted(() => {
  window.removeEventListener('open-create-database', handleOpenCreateDatabase)
})

</script>

<style scoped>
/* ── Main layout ─────────────────────────────────────────── */
/* n-layout is a flex child of #app-window (via display:contents providers) */
.main-layout {
  background: transparent !important;
  flex: 1 !important;
  min-height: 0 !important;
  align-self: stretch !important;
}
.content-layout {
  background: transparent !important;
  flex: 1 !important;
  min-height: 0 !important;
  overflow: hidden !important;
}

/* ── Sidebar: glass panel ─────────────────────────────────── */
.sidebar {
  /* align-self:stretch fills full flex row height */
  align-self: stretch !important;
  background: rgba(255, 255, 255, 0.1) !important;
  border-right: 1px solid rgba(255, 255, 255, 0.08) !important;
}


/* Hide NaiveUI default sider border line */
.sidebar :deep(.n-layout-sider__border) { display: none !important; }

/*
 * NaiveUI INTERNAL containers — must ALL be transparent and full-height.
 * NaiveUI renders: .n-layout-sider > .n-layout-sider-scroll-container > [slot]
 * The scroll container gets NaiveUI's border-radius + bg by default → we strip both.
 */
.sidebar :deep(.n-layout-sider-scroll-container) {
  background: transparent !important;
  background-color: transparent !important;
  border-radius: 0 !important;
  height: 100% !important;
  width: 100% !important;
  overflow: hidden !important;
  box-shadow: none !important;
  display: flex !important;
  flex-direction: column !important;
  padding: 0 !important;
}

/* ── Native scroll wrapper (replaces NScrollbar) ───────────
   flex:1 + height:100% inside a flex column = fills remaining space.
   overflow-y:auto = real OS scrollbar when tree content overflows.     */
.sidebar-scroll {
  flex: 1;
  min-height: 0;          /* required in flex child to allow scrolling */
  overflow-y: auto;
  overflow-x: hidden;
  /* Firefox scrollbar */
  scrollbar-width: thin;
  scrollbar-color: rgba(255,255,255,0.14) transparent;
}
.sidebar-scroll::-webkit-scrollbar { width: 4px; }
.sidebar-scroll::-webkit-scrollbar-track { background: transparent; }
.sidebar-scroll::-webkit-scrollbar-thumb {
  background: rgba(255,255,255,0.14);
  border-radius: 2px;
}
.sidebar-scroll::-webkit-scrollbar-thumb:hover {
  background: rgba(255,255,255,0.26);
}

/* ── Toolbar Header — subtle glass divider bar ────────────── */
.toolbar-header {
  height: 48px; /* titleBarHeight from app-window.md */
  padding: 12px 18px 6px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: transparent !important;
  border-bottom: none;
  z-index: 5;
  flex-shrink: 0;
}

.toolbar-right-cluster {
  display: flex;
  align-items: center;
  gap: 10px;
}

.connection-switcher {
  width: 220px;
}

/* ── Toolbar Capsule — light, no background ──────────────── */
.toolbar-capsule {
  display: flex;
  align-items: center;
  gap: 4px;
}

.toolbar-btn {
  border-radius: 6px !important;
  font-size: var(--ref-font-size-sm) !important;
  font-weight: var(--ref-font-weight-medium) !important;
  height: 30px !important;
  padding: 0 14px !important;
  transition: background-color 0.15s ease, color 0.15s ease !important;
}

/* Active primary button: #5B63F6 solid */
.toolbar-btn:not(.ghost) {
  background: #5B63F6 !important;
  color: #FFFFFF !important;
  border: none !important;
}
.toolbar-btn:not(.ghost):hover {
  background: #6E76FF !important;
}

/* Ghost (inactive) toolbar buttons: transparent */
.toolbar-btn.ghost {
  background: transparent !important;
  color: var(--sys-color-text-secondary) !important;
  border: none !important;
}
.toolbar-btn.ghost:hover {
  background: transparent !important;
  color: var(--sys-color-text-primary) !important;
}

/* ── Main Content: transparent, lets glass show through ───── */
.main-content {
  height: calc(100dvh - 48px);
  overflow: hidden;
  background: transparent !important;
  padding: 0 14px 14px;
}

/* NaiveUI content override */
.main-content :deep(.n-layout-scroll-container) {
  background: transparent !important;
}
</style>
