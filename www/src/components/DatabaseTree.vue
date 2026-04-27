<template>
  <div class="database-tree-wrapper">
    <!-- Loading state -->
    <div v-if="loading" class="loading-state">
      <n-spin size="small" />
      <span class="loading-text">{{ t('tree.loading') }}</span>
    </div>
    
    <!-- Empty state - no connections -->
    <div v-else-if="allConnections.length === 0" class="empty-state-wrapper">
      <!-- minimal empty state: icon + label + text button -->
      <div class="empty-icon">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round">
          <ellipse cx="12" cy="5" rx="9" ry="3"/>
          <path d="M3 5v6c0 1.66 4.03 3 9 3s9-1.34 9-3V5"/>
          <path d="M3 11v6c0 1.66 4.03 3 9 3s9-1.34 9-3v-6"/>
        </svg>
      </div>
      <p class="empty-label">{{ t('tree.noConnections') }}</p>
      <button class="empty-add-btn" @click="$emit('add-connection')">
        {{ t('tree.addConnection') }}
      </button>
    </div>
    
    <!-- Tree with connections -->
    <template v-else>
      <div class="tree-toolbar">
        <n-input
          v-model:value="treeSearch"
          clearable
          size="small"
          class="tree-search"
          :placeholder="t('tree.searchPlaceholder')"
        >
          <template #prefix>
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="7"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
          </template>
        </n-input>
      </div>

      <n-tree
        block-line
        :data="treeData"
        :expanded-keys="expandedKeys"
        :selected-keys="selectedKeys"
        :on-update:expanded-keys="handleExpand"
        :on-update:selected-keys="handleSelect"
        :render-prefix="renderPrefix"
        :render-label="renderLabel"
        :node-props="nodeProps"
        :override-default-node-click-behavior="resolveNodeClickBehavior"
        expand-on-click
        selectable
      />
    </template>
    
    <!-- Context Menu -->
    <n-dropdown
      placement="bottom-start"
      trigger="manual"
      :x="contextMenuX"
      :y="contextMenuY"
      :options="contextMenuOptions"
      :show="showContextMenu"
      :on-clickoutside="closeContextMenu"
      @select="handleContextMenuSelect"
    />

    <FormModal
      v-model="renameDialogVisible"
      :title="t('tree.renameTable')"
      :width="420"
    >
      <n-form>
        <n-form-item :label="t('tree.renameTargetLabel')">
          <n-input :value="renameTarget.oldName" disabled />
        </n-form-item>
        <n-form-item :label="t('tree.renameNewLabel')">
          <n-input v-model:value="renameTarget.newName" :placeholder="t('tree.renamePlaceholder')" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="renameDialogVisible = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="renameSubmitting" @click="submitRename">
            {{ t('common.confirm') }}
          </n-button>
        </n-space>
      </template>
    </FormModal>

    <FormModal
      v-model="dropDatabaseDialogVisible"
      :title="t('tree.dropDbTitle')"
      :width="460"
    >
      <div class="drop-db-dialog">
        <p class="drop-db-dialog__lead">
          {{ t('tree.dropDbContent', { name: dropDatabaseTarget.dbName }) }}
        </p>
        <p class="drop-db-dialog__hint">
          {{ t('tree.dropDbConfirmHint') }}
        </p>

        <n-form>
          <n-form-item :label="t('tree.dropDbNameLabel')">
            <n-input :value="dropDatabaseTarget.dbName" disabled />
          </n-form-item>
          <n-form-item :label="t('tree.dropDbConfirmLabel')">
            <n-input
              v-model:value="dropDatabaseTarget.confirmText"
              :placeholder="t('tree.dropDbConfirmPlaceholder')"
              @keydown.enter.prevent="submitDropDatabase"
            />
          </n-form-item>
        </n-form>
      </div>
      <template #footer>
        <n-space justify="end">
          <n-button @click="closeDropDatabaseDialog">{{ t('tree.dropDbCancel') }}</n-button>
          <n-button
            type="error"
            :disabled="!isDropDatabaseConfirmMatched"
            :loading="dropDatabaseSubmitting"
            @click="submitDropDatabase"
          >
            {{ t('tree.dropDbOk') }}
          </n-button>
        </n-space>
      </template>
    </FormModal>

  </div>
</template>

<script setup>
import { computed, ref, h, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTooltip } from 'naive-ui'
import { useDatabaseStore } from '@/stores/database'
import FormModal from '@/components/FormModal.vue'
import connectionApi from '@/utils/connectionApi'
import api from '@/utils/api'
import gmssh from '@/utils/gmssh'

const { t } = useI18n()

const store = useDatabaseStore()

const emit = defineEmits(['add-connection', 'edit-connection'])

const treeData = ref([])
const expandedKeys = ref([])
const selectedKeys = ref([])
const databaseObjectCache = ref({})
// PostgreSQL schema cache: key = 'connId::dbName' → { schemas: string[], tables: { [schema]: object[] } }
const pgSchemaCache = ref({})
const treeSearch = ref('')
const workspaceSnapshot = ref({ byConn: {}, active: null })
const hoveredKey = ref('')

// Render label with tooltip for long names
const renderLabel = ({ option }) => {
  const labelClass = ['tree-node-label', `tree-node-label--${option.type}`]
  const trigger = h('div', { class: ['tree-node-label-wrap', `tree-node-label-wrap--${option.type}`] }, [
    h('span', { class: labelClass }, option.label),
    option.connectionBadge ? h('span', { class: ['tree-node-meta', `tree-node-meta--${option.connectionBadge}`] }, option.connectionBadgeLabel) : null,
    typeof option.count === 'number' ? h('span', { class: 'tree-node-pill' }, String(option.count)) : null
  ].filter(Boolean))

  return h(NTooltip, { trigger: 'hover', placement: 'top', delay: 300 }, {
    trigger: () => trigger,
    default: () => option.label
  })
}
const allConnections = ref([])
const loading = ref(true)
const connecting = ref(false)
const connectingId = ref(null)
const expandingDbKey = ref(null)  // Track which database is loading tables

// Context menu state
const showContextMenu = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuNode = ref(null)
const renameDialogVisible = ref(false)
const renameSubmitting = ref(false)
const renameTarget = ref({
  connId: '',
  database: '',
  oldName: '',
  newName: ''
})
const dropDatabaseDialogVisible = ref(false)
const dropDatabaseSubmitting = ref(false)
const dropDatabaseTarget = ref({
  connId: '',
  dbName: '',
  confirmText: ''
})
const isDropDatabaseConfirmMatched = computed(() => {
  const value = dropDatabaseTarget.value.confirmText.trim().toLowerCase()
  return value === 'delete' || value === 'confirm delete' || value === '确认删除'
})

const getDatabaseCacheKey = (connId, dbName) => `${connId}::${dbName}`
const getSavedConnectionConfig = (savedConnId) => allConnections.value.find((item) => item.id === savedConnId) || null
const getConnectionDbType = (savedConnId) => getSavedConnectionConfig(savedConnId)?.dbType || 'mysql'
const MYSQL_SYSTEM_DATABASES = new Set(['information_schema', 'mysql', 'performance_schema', 'sys'])

const quoteIdentifier = (identifier, dbType = 'mysql') => {
  const value = String(identifier || '')
  if (dbType === 'postgres') {
    return `"${value.replace(/"/g, '""')}"`
  }

  return `\`${value.replace(/`/g, '``')}\``
}

function isProtectedMySQLDatabase(node) {
  const dbType = String(getConnectionDbType(node?.connId) || '').toLowerCase()
  const dbName = String(node?.dbName || '').toLowerCase()
  return dbType === 'mysql' && MYSQL_SYSTEM_DATABASES.has(dbName)
}

const normalizeObjectType = (type) => type === 'view' ? 'view' : 'table'

const getCachedDatabaseObjects = (connId, dbName) => {
  const key = getDatabaseCacheKey(connId, dbName)
  if (!Object.prototype.hasOwnProperty.call(databaseObjectCache.value, key)) {
    return null
  }

  return databaseObjectCache.value[key]
}

const cacheDatabaseObjects = (connId, dbName, objects) => {
  databaseObjectCache.value = {
    ...databaseObjectCache.value,
    [getDatabaseCacheKey(connId, dbName)]: Array.isArray(objects) ? objects : []
  }
}

const pruneDatabaseObjectCache = () => {
  const validConnIds = new Set(allConnections.value.map((item) => String(item.id)))
  const nextCache = {}

  Object.entries(databaseObjectCache.value).forEach(([key, value]) => {
    const [connId] = key.split('::')
    if (validConnIds.has(connId)) {
      nextCache[key] = value
    }
  })

  databaseObjectCache.value = nextCache
}

const matchesTreeSearch = (text) => {
  const search = treeSearch.value.trim().toLowerCase()
  if (!search) return true
  return String(text || '').toLowerCase().includes(search)
}

const getWorkspaceConnSummary = (connId) => workspaceSnapshot.value.byConn?.[connId] || {
  overviews: [],
  objects: [],
  queries: []
}

const isNodeOpened = (node) => {
  const connId = node.runtimeConnId || node.connId
  if (!connId) return false

  const summary = getWorkspaceConnSummary(connId)

  if (node.type === 'database') {
    return summary.overviews.includes(node.dbName)
  }

  if (node.type === 'table' || node.type === 'view') {
    return summary.objects.includes(`${node.dbName}::${node.tableName}`)
  }

  return false
}

const isNodeActive = (node) => {
  const active = workspaceSnapshot.value.active
  if (!active) return false

  const connId = node.runtimeConnId || node.connId
  if (String(active.connId || '') !== String(connId || '')) return false

  if (node.type === 'database') {
    return active.type === 'overview' && active.database === node.dbName
  }

  if (node.type === 'table' || node.type === 'view') {
    return active.database === node.dbName && active.table === node.tableName
  }

  return false
}

const buildObjectLeafNode = (connId, dbName, object, schemaName = null) => ({
  key: `${normalizeObjectType(object.type)}-${connId}-${dbName}${schemaName ? `-${schemaName}` : ''}-${object.name}`,
  label: object.name,
  isLeaf: true,
  type: normalizeObjectType(object.type),
  connId,
  dbName,
  schemaName,
  tableName: object.name,
  opened: false,
  active: false
})

const buildObjectGroupChildren = (connId, dbName, objects) => {
  if (!Array.isArray(objects)) return []

  const tables = objects.filter((object) => normalizeObjectType(object.type) === 'table' && matchesTreeSearch(object.name))
  const views = objects.filter((object) => normalizeObjectType(object.type) === 'view' && matchesTreeSearch(object.name))
  const children = []

  if (tables.length > 0) {
    children.push({
      key: `group-${connId}-${dbName}-tables`,
      label: t('tree.groupTables'),
      type: 'object-group',
      groupKind: 'table',
      isLeaf: false,
      children: tables.map((object) => {
        const node = buildObjectLeafNode(connId, dbName, object)
        node.opened = isNodeOpened(node)
        node.active = isNodeActive(node)
        return node
      })
    })
  }

  if (views.length > 0) {
    children.push({
      key: `group-${connId}-${dbName}-views`,
      label: t('tree.groupViews'),
      type: 'object-group',
      groupKind: 'view',
      isLeaf: false,
      children: views.map((object) => {
        const node = buildObjectLeafNode(connId, dbName, object)
        node.opened = isNodeOpened(node)
        node.active = isNodeActive(node)
        return node
      })
    })
  }

  return children
}

const buildDatabaseNodeChildren = (conn, dbName, objects) => {
  if (!objects && conn.dbType !== 'postgres') return []

  if (conn.dbType === 'postgres') {
    // Build from pgSchemaCache
    const cacheKey = getDatabaseCacheKey(conn.id, dbName)
    const cached = pgSchemaCache.value[cacheKey]
    if (!cached?.schemas?.length) return []
    return cached.schemas.map(schemaName => {
      const schemaTables = cached.tables?.[schemaName] || []
      return {
        key: `schema-${conn.id}-${dbName}-${schemaName}`,
        label: schemaName,
        isLeaf: false,
        children: schemaTables.map(table => {
          const node = buildObjectLeafNode(conn.id, dbName, table, schemaName)
          node.opened = isNodeOpened(node)
          node.active = isNodeActive(node)
          return node
        }),
        type: 'schema',
        connId: conn.id,
        dbName: dbName,
        schemaName: schemaName
      }
    })
  }

  return buildObjectGroupChildren(conn.id, dbName, objects)
}

// Load all saved connections
const loadConnections = async () => {
  loading.value = true
  try {
    const result = await connectionApi.listConnections()
    allConnections.value = result || []
    pruneDatabaseObjectCache()
    buildTreeData()
  } catch (error) {

    allConnections.value = []
  } finally {
    loading.value = false
  }
}

const ensureConnectionContext = async (savedConnId) => {
  let runtimeConnId = store.getRuntimeConnIdByConfigId(savedConnId)

  if (runtimeConnId) {
    await store.switchConnection(runtimeConnId)
    return runtimeConnId
  }

  connecting.value = true
  connectingId.value = savedConnId
  try {
    const fullConnData = await connectionApi.getConnection(savedConnId)
    const result = await store.connect(fullConnData)
    runtimeConnId = result?.connId || store.activeConnection
    gmssh.success(t('tree.connectSuccess'))
    buildTreeData()
    return runtimeConnId
  } finally {
    connecting.value = false
    connectingId.value = null
  }
}

// Build tree data structure with all connections
const buildTreeData = () => {
  const connectionNodes = []
  
  // Show all saved connections
  for (const conn of allConnections.value) {
    const runtimeConnId = store.getRuntimeConnIdByConfigId(conn.id)
    const connState = runtimeConnId ? store.getConnectionState(runtimeConnId) : null
    const isActive = !!runtimeConnId
    const isCurrent = runtimeConnId && store.currentConnId === runtimeConnId
    
    const connNode = {
      key: `conn-${conn.id}`,
      label: conn.name || `${conn.dbType}@${conn.host}`,
      isLeaf: false,
      children: [],
      type: 'connection',
      connId: conn.id,
      runtimeConnId,
      connData: conn,
      isConnected: isActive,
      isCurrent,
      opened: false,
      active: false
    }
    
    // If this connection is active, show its databases from isolated state
    if (connState?.databases?.length > 0) {
      connNode.children = connState.databases
        .filter((db) => matchesTreeSearch(db.name) || (getCachedDatabaseObjects(conn.id, db.name) || []).some((object) => matchesTreeSearch(object.name)))
        .map((db) => {
        const cachedObjects = getCachedDatabaseObjects(conn.id, db.name)
        const selectedObjects = connState.selectedDatabase === db.name ? (connState.tables || []) : null
        const objects = cachedObjects ?? selectedObjects
        const opened = isNodeOpened({ type: 'database', connId: conn.id, runtimeConnId, dbName: db.name })
        const active = isNodeActive({ type: 'database', connId: conn.id, runtimeConnId, dbName: db.name })

        return {
          key: `db-${conn.id}-${db.name}`,
          label: db.name,
          isLeaf: false,
          children: buildDatabaseNodeChildren(conn, db.name, objects),
          type: 'database',
          connId: conn.id,
          dbName: db.name,
          isSystem: db.isSystem,
          opened,
          active
        }
      })
    }
    
    connectionNodes.push(connNode)
  }
  
  // Wrap in root node
  treeData.value = [{
    key: 'root-connections',
    label: t('tree.myConnections'),
    isLeaf: false,
    children: connectionNodes,
    type: 'root'
  }]
  
  // Auto-expand root node
  if (!expandedKeys.value.includes('root-connections')) {
    expandedKeys.value = ['root-connections', ...expandedKeys.value]
  }
}

// Load schemas for PostgreSQL database
const loadSchemasForDatabase = async (connId, dbName, dbType) => {
  try {
    const activeConnId = await ensureConnectionContext(connId)
    
    const schemas = await api.listSchemas(activeConnId, dbName)
    
    if (schemas && schemas.length > 0) {
      const cacheKey = getDatabaseCacheKey(connId, dbName)
      pgSchemaCache.value = {
        ...pgSchemaCache.value,
        [cacheKey]: {
          schemas,
          tables: pgSchemaCache.value[cacheKey]?.tables || {}
        }
      }
      buildTreeData()
    }
  } catch (error) {
    // silently ignore
  }
}

// Load tables for a specific schema (PostgreSQL)
const loadTablesForSchema = async (connId, dbName, schemaName) => {
  try {
    const runtimeConnId = await ensureConnectionContext(connId)
    await store.selectDatabase(dbName, runtimeConnId)
    
    const state = store.getConnectionState(runtimeConnId)
    if (state?.tables) {
      const cacheKey = getDatabaseCacheKey(connId, dbName)
      const existing = pgSchemaCache.value[cacheKey] || { schemas: [], tables: {} }
      pgSchemaCache.value = {
        ...pgSchemaCache.value,
        [cacheKey]: {
          ...existing,
          tables: {
            ...existing.tables,
            [schemaName]: state.tables
          }
        }
      }
      buildTreeData()
    }
  } catch (error) {
    // silently ignore
  }
}

// Load tables when database is expanded (MySQL/SQLite)
const loadTablesForDatabase = async (connId, dbName) => {
  try {
    const runtimeConnId = await ensureConnectionContext(connId)
    const objects = await store.selectDatabase(dbName, runtimeConnId)
    cacheDatabaseObjects(connId, dbName, objects || [])
    buildTreeData()
  } catch (error) {
    console.error('Failed to load tables:', error)
  }
}

const findNodeByKey = (key) => {
  const find = (nodes) => {
    for (const node of nodes) {
      if (node.key === key) return node
      if (node.children) {
        const found = find(node.children)
        if (found) return found
      }
    }
    return null
  }
  return find(treeData.value)
}

const handleExpand = async (keys) => {
  const previousKeys = expandedKeys.value
  expandedKeys.value = keys
  
  // Find newly expanded keys
  const newlyExpanded = keys.filter(k => !previousKeys.includes(k))
  
  for (const key of newlyExpanded) {
    // Handle connection expansion - connect if not connected
    if (key.startsWith('conn-')) {
      const node = findNodeByKey(key)
      if (node?.connId) {
        try {
          await ensureConnectionContext(node.connId)
          buildTreeData()
        } catch (error) {
          gmssh.error(t('tree.connectFailed', { msg: error.message }))
        }
      }
    }
    // Handle database expansion - load schemas (PostgreSQL) or tables (MySQL/SQLite)
    else if (key.startsWith('db-')) {
      const node = findNodeByKey(key)
      if (node) {
        expandingDbKey.value = key
        try {
          // Check database type from connection config
          const conn = allConnections.value.find(c => c.id === node.connId)
          const isPostgreSQL = conn && conn.dbType === 'postgres'
          const cachedObjects = getCachedDatabaseObjects(node.connId, node.dbName)
          
          if (isPostgreSQL) {
            // PostgreSQL: check cache first, load schemas if needed
            const pgCacheKey = getDatabaseCacheKey(node.connId, node.dbName)
            if (pgSchemaCache.value[pgCacheKey]?.schemas?.length) {
              buildTreeData()
            } else {
              await loadSchemasForDatabase(node.connId, node.dbName, conn.dbType)
            }
          } else if (cachedObjects !== null) {
            buildTreeData()
          } else {
            // MySQL/SQLite: load tables directly
            await loadTablesForDatabase(node.connId, node.dbName)
          }
        } catch (error) {

          const errorMsg = error.message || String(error)
          
          // Provide user-friendly error messages
          if (errorMsg.includes('Connection not found')) {
            gmssh.error(t('tree.connInvalidMsg'))
          } else if (errorMsg.includes('timeout') || errorMsg.includes('deadline exceeded')) {
            gmssh.error(t('tree.connTimeoutMsg'))
          } else {
            gmssh.error(t('tree.loadFailed', { msg: errorMsg }))
          }
        } finally {
          expandingDbKey.value = null
        }
      }
    }
    // Handle schema expansion (PostgreSQL) - load tables
    else if (key.startsWith('schema-')) {
      const node = findNodeByKey(key)
      if (node) {
        expandingDbKey.value = key
        try {
          await loadTablesForSchema(node.connId, node.dbName, node.schemaName)
        } catch (error) {

          gmssh.error(t('tree.loadTablesFailed', { msg: error.message || String(error) }))
        } finally {
          expandingDbKey.value = null
        }
      }
    }
  }
}

const handleSelect = async (keys) => {
  selectedKeys.value = keys
  if (keys.length === 0) return

  const key = keys[0]
  const node = findNodeByKey(key)
  
  if (node && (node.type === 'table' || node.type === 'view')) {
    try {
      const runtimeConnId = await ensureConnectionContext(node.connId)
      await store.selectTable(node.tableName, runtimeConnId)
    } catch (error) {
      gmssh.error(t('tree.loadFailed', { msg: error.message }))
    }
  }
}

const handleWorkspaceTabsUpdated = (event) => {
  workspaceSnapshot.value = event.detail || { byConn: {}, active: null }
  buildTreeData()
}

const resolveNodeClickBehavior = ({ option }) => {
  if (option.type === 'table' || option.type === 'view') {
    return 'toggleSelect'
  }

  if (option.type === 'connection' || option.type === 'database' || option.type === 'schema' || option.type === 'root' || option.type === 'object-group') {
    return 'toggleExpand'
  }

  return 'default'
}

const openDatabaseOverview = async (node) => {
  const runtimeConnId = await ensureConnectionContext(node.connId)
  const objects = await store.selectDatabase(node.dbName, runtimeConnId)
  cacheDatabaseObjects(node.connId, node.dbName, objects || [])
  buildTreeData()
  window.dispatchEvent(new CustomEvent('open-database-overview', {
    detail: {
      database: node.dbName,
      connId: runtimeConnId
    }
  }))
}

// Node props for right-click and double-click
const nodeProps = ({ option }) => {
  return {
    class: [
      `tree-node--${option.type}`,
      option.groupKind ? `tree-node--group-${option.groupKind}` : '',
      option.isCurrent ? 'is-current' : '',
      option.isSystem ? 'is-system' : '',
      option.opened ? 'is-opened' : '',
      option.active ? 'is-active' : ''
    ].filter(Boolean).join(' '),
    onMouseenter() {
      hoveredKey.value = option.key
    },
    onMouseleave() {
      if (hoveredKey.value === option.key) {
        hoveredKey.value = ''
      }
    },
    onContextmenu(e) {
      e.preventDefault()
      showContextMenu.value = false
      contextMenuNode.value = option
      selectedKeys.value = [option.key]
      
      nextTick(() => {
        contextMenuX.value = e.clientX
        contextMenuY.value = e.clientY
        showContextMenu.value = true
      })
    },
    async onClick() {
      if (option.type === 'connection' && option.runtimeConnId && option.runtimeConnId !== store.currentConnId) {
        try {
          await store.switchConnection(option.runtimeConnId)
        } catch (error) {
          gmssh.error(t('tree.loadFailed', { msg: error.message }))
        }
      }

      if (option.type === 'database') {
        selectedKeys.value = [option.key]
        try {
          await openDatabaseOverview(option)
        } catch (error) {
          gmssh.error(t('tree.loadFailed', { msg: error.message }))
        }
      }
    },
    onDblclick() {
      // Double-click on table opens data view
      if (option.type === 'table' || option.type === 'view') {
        ensureConnectionContext(option.connId).then((runtimeConnId) => {
          window.dispatchEvent(new CustomEvent('open-table-data', {
            detail: { database: option.dbName, table: option.tableName, connId: runtimeConnId }
          }))
        }).catch((error) => {
          gmssh.error(t('tree.loadFailed', { msg: error.message }))
        })
      }
    }
  }
}

const renderPrefix = ({ option }) => {
  const elements = []
  
  // SVG icon helper function
  const svgIcon = (paths, size = 14, color = 'currentColor', opacity = 0.85) => {
    return h('svg', {
      width: size,
      height: size,
      viewBox: '0 0 24 24',
      fill: 'none',
      stroke: color,
      'stroke-width': '1.5',
      'stroke-linecap': 'round',
      'stroke-linejoin': 'round',
      style: `margin-right: 8px; opacity: ${opacity}; flex-shrink: 0; transition: opacity 0.2s;`
    }, paths.map(d => h('path', { d })))
  }
  
  // Status dot for connection nodes
  if (option.type === 'connection') {
    const isConnecting = connecting.value && connectingId.value === option.connId
    
    if (isConnecting) {
      elements.push(h('span', { class: 'conn-spinner' }))
    } else {
      const dotColor = option.isConnected ? 'var(--ref-color-green-6)' : 'var(--ref-color-white-15)'
      const shadow = option.isConnected ? '0 0 8px rgba(50, 178, 93, 0.8), 0 0 16px rgba(50, 178, 93, 0.36)' : 'none'
      const transition = 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)'
      elements.push(h('span', {
        class: option.isConnected ? 'status-led-active status-led-active--green' : 'status-led-inactive',
        style: `width: 8px; height: 8px; border-radius: 50%; background: ${dotColor}; box-shadow: ${shadow}; transition: ${transition}; margin-right: 12px; flex-shrink: 0;`
      }))
    }
  }
  
  // SVG Icon based on type
  let iconElement = null
  
  if (option.type === 'root') {
    // Folder icon for root
    const isExpanded = expandedKeys.value.includes(option.key)
    if (isExpanded) {
      iconElement = svgIcon(['M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z', 'M2 10h20'])
    } else {
      iconElement = svgIcon(['M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z'])
    }
  } else if (option.type === 'connection') {
    // DB-type-specific brand icons from public/ folder
    const dbType = option.connData?.dbType || ''
    const active = option.isConnected
    
    // Add inner glow for 3D resin feel
    const innerGlow = active ? 'filter: drop-shadow(0 0 2px rgba(255,255,255,0.4));' : 'opacity: 0.4;'
    const iconColor = 'currentColor'

    const mysqlPath = 'M13 21c-1.427-1.026-3.59-3.854-4-6c-.486.77-1.501 2-2 2c-1.499-.888-.574-3.973 0-6c-1.596-1.433-2.468-2.458-2.5-4C1.15 3.56 4.056 1.73 7 4h1c8.482.5 6.421 8.07 9 11.5c2.295.522 3.665 2.254 5 3.5c-2.086-.2-2.784-.344-3.5 0c.478 1.64 2.123 2.2 3.5 3M9 7h.01'
    
    // Custom SVG renderer for complex path icons
    const renderComplexIcon = (pathData, size = 16, color = iconColor, isFill = false) => {
      const pathAttrs = isFill ? { d: pathData, fill: 'currentColor' } : { d: pathData }
      return h('svg', {
        width: size,
        height: size,
        viewBox: '0 0 24 24',
        fill: isFill ? 'currentColor' : 'none',
        stroke: isFill ? 'none' : color,
        'stroke-width': '1.2',
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        style: `margin-right: 12px; flex-shrink: 0; transition: color 0.3s; ${innerGlow}`
      }, [h('path', pathAttrs)])
    }

    if (dbType === 'mysql') {
      iconElement = renderComplexIcon(mysqlPath, 16, iconColor, false)
    } else if (dbType === 'postgres') {
      iconElement = renderComplexIcon('M23.56 14.723a.5.5 0 0 0-.057-.12q-.21-.395-1.007-.231c-1.654.34-2.294.13-2.526-.02c1.342-2.048 2.445-4.522 3.041-6.83c.272-1.05.798-3.523.122-4.73a1.6 1.6 0 0 0-.15-.236C21.693.91 19.8.025 17.51.001c-1.495-.016-2.77.346-3.116.479a10 10 0 0 0-.516-.082a8 8 0 0 0-1.312-.127c-1.182-.019-2.203.264-3.05.84C8.66.79 4.729-.534 2.296 1.19C.935 2.153.309 3.873.43 6.304c.041.818.507 3.334 1.243 5.744q.69 2.26 1.433 3.582q.83 1.493 1.714 1.79c.448.148 1.133.143 1.858-.729a56 56 0 0 1 1.945-2.206c.435.235.906.362 1.39.377v.004a11 11 0 0 0-.247.305c-.339.43-.41.52-1.5.745c-.31.064-1.134.233-1.146.811a.6.6 0 0 0 .091.327c.227.423.922.61 1.015.633c1.335.333 2.505.092 3.372-.679c-.017 2.231.077 4.418.345 5.088c.221.553.762 1.904 2.47 1.904q.375.001.829-.094c1.782-.382 2.556-1.17 2.855-2.906c.15-.87.402-2.875.539-4.101c.017-.07.036-.12.057-.136c0 0 .07-.048.427.03l.044.007l.254.022l.015.001c.847.039 1.911-.142 2.531-.43c.644-.3 1.806-1.033 1.595-1.67M2.37 11.876c-.744-2.435-1.178-4.885-1.212-5.571c-.109-2.172.417-3.683 1.562-4.493c1.837-1.299 4.84-.54 6.108-.13l-.01.01C6.795 3.734 6.843 7.226 6.85 7.44c0 .082.006.199.016.36c.034.586.1 1.68-.074 2.918c-.16 1.15.194 2.276.973 3.089q.12.126.252.237c-.347.371-1.1 1.193-1.903 2.158c-.568.682-.96.551-1.088.508c-.392-.13-.813-.587-1.239-1.322c-.48-.839-.963-2.032-1.415-3.512m6.007 5.088a1.6 1.6 0 0 1-.432-.178c.089-.039.237-.09.483-.14c1.284-.265 1.482-.451 1.915-1a8 8 0 0 1 .367-.443a.4.4 0 0 0 .074-.13c.17-.151.272-.11.436-.042c.156.065.308.26.37.475c.03.102.062.295-.045.445c-.904 1.266-2.222 1.25-3.168 1.013m2.094-3.988l-.052.14c-.133.357-.257.689-.334 1.004c-.667-.002-1.317-.288-1.81-.803c-.628-.655-.913-1.566-.783-2.5c.183-1.308.116-2.447.08-3.059l-.013-.22c.296-.262 1.666-.996 2.643-.772c.446.102.718.406.83.928c.585 2.704.078 3.83-.33 4.736a9 9 0 0 0-.23.546m7.364 4.572q-.024.266-.062.596l-.146.438a.4.4 0 0 0-.018.108c-.006.475-.054.649-.115.87a4.8 4.8 0 0 0-.18 1.057c-.11 1.414-.878 2.227-2.417 2.556c-1.515.325-1.784-.496-2.02-1.221a7 7 0 0 0-.078-.227c-.215-.586-.19-1.412-.157-2.555c.016-.561-.025-1.901-.33-2.646q.006-.44.019-.892a.4.4 0 0 0-.016-.113a2 2 0 0 0-.044-.208c-.122-.428-.42-.786-.78-.935c-.142-.059-.403-.167-.717-.087c.067-.276.183-.587.309-.925l.053-.142c.06-.16.134-.325.213-.5c.426-.948 1.01-2.246.376-5.178c-.237-1.098-1.03-1.634-2.232-1.51c-.72.075-1.38.366-1.709.532a6 6 0 0 0-.196.104c.092-1.106.439-3.174 1.736-4.482a4 4 0 0 1 .303-.276a.35.35 0 0 0 .145-.064c.752-.57 1.695-.85 2.802-.833q.616.01 1.174.081c1.94.355 3.244 1.447 4.036 2.383c.814.962 1.255 1.931 1.431 2.454c-1.323-.134-2.223.127-2.68.78c-.992 1.418.544 4.172 1.282 5.496c.135.242.252.452.289.54c.24.583.551.972.778 1.256c.07.087.138.171.189.245c-.4.116-1.12.383-1.055 1.717a35 35 0 0 1-.084.815c-.046.208-.07.46-.1.766m.89-1.621c-.04-.832.27-.919.597-1.01l.135-.041a1 1 0 0 0 .134.103c.57.376 1.583.421 3.007.134c-.202.177-.519.4-.953.601c-.41.19-1.096.333-1.747.364c-.72.034-1.086-.08-1.173-.151m.57-9.271a7 7 0 0 1-.105 1.001c-.055.358-.112.728-.127 1.177c-.014.436.04.89.093 1.33c.107.887.216 1.8-.207 2.701a4 4 0 0 1-.188-.385a8 8 0 0 0-.325-.617c-.616-1.104-2.057-3.69-1.32-4.744c.38-.543 1.342-.566 2.179-.463m.228 7.013l-.085-.107l-.035-.044c.726-1.2.584-2.387.457-3.439c-.052-.432-.1-.84-.088-1.222c.013-.407.066-.755.118-1.092c.064-.415.13-.844.111-1.35a.6.6 0 0 0 .012-.19c-.046-.486-.6-1.938-1.73-3.253a7.8 7.8 0 0 0-2.688-2.04A9.3 9.3 0 0 1 17.62.746c2.052.046 3.675.814 4.824 2.283a1 1 0 0 1 .067.1c.723 1.356-.276 6.275-2.987 10.54m-8.816-6.116c-.025.18-.31.423-.621.423l-.081-.006a.8.8 0 0 1-.506-.315c-.046-.06-.12-.178-.106-.285a.22.22 0 0 1 .093-.149c.118-.089.352-.122.61-.086c.316.044.642.193.61.418m7.93-.411c.011.08-.049.2-.153.31a.72.72 0 0 1-.408.223l-.075.005c-.293 0-.541-.234-.56-.371c-.024-.177.264-.31.56-.352c.298-.042.612.009.636.185', 16, iconColor, true)
    } else if (dbType === 'sqlite') {
      iconElement = renderComplexIcon('M21.678.521c-1.032-.92-2.28-.55-3.513.544a9 9 0 0 0-.547.535c-2.109 2.237-4.066 6.38-4.674 9.544c.237.48.422 1.093.544 1.561a13 13 0 0 1 .164.703s-.019-.071-.096-.296l-.05-.146l-.033-.08c-.138-.32-.518-.995-.686-1.289c-.143.423-.27.818-.376 1.176c.484.884.778 2.4.778 2.4s-.025-.099-.147-.442c-.107-.303-.644-1.244-.772-1.464c-.217.804-.304 1.346-.226 1.478c.152.256.296.698.422 1.186c.286 1.1.485 2.44.485 2.44l.017.224a22 22 0 0 0 .056 2.748c.095 1.146.273 2.13.5 2.657l.155-.084c-.334-1.038-.47-2.399-.41-3.967c.09-2.398.642-5.29 1.661-8.304c1.723-4.55 4.113-8.201 6.3-9.945c-1.993 1.8-4.692 7.63-5.5 9.788c-.904 2.416-1.545 4.684-1.931 6.857c.666-2.037 2.821-2.912 2.821-2.912s1.057-1.304 2.292-3.166c-.74.169-1.955.458-2.362.629c-.6.251-.762.337-.762.337s1.945-1.184 3.613-1.72C21.695 7.9 24.195 2.767 21.678.521m-18.573.543A1.84 1.84 0 0 0 1.27 2.9v16.608a1.84 1.84 0 0 0 1.835 1.834h9.418a23 23 0 0 1-.052-2.707c-.006-.062-.011-.141-.016-.2a27 27 0 0 0-.473-2.378c-.121-.47-.275-.898-.369-1.057c-.116-.197-.098-.31-.097-.432c0-.12.015-.245.037-.386a10 10 0 0 1 .234-1.045l.217-.028c-.017-.035-.014-.065-.031-.097l-.041-.381a33 33 0 0 1 .382-1.194l.2-.019c-.008-.016-.01-.038-.018-.053l-.043-.316c.63-3.28 2.587-7.443 4.8-9.791c.066-.069.133-.128.198-.194Z', 16, iconColor, true)
    } else {
      // Fallback generic database icon
      iconElement = svgIcon([
        'M12 2C6.48 2 2 4.24 2 7v10c0 2.76 4.48 5 10 5s10-2.24 10-5V7c0-2.76-4.48-5-10-5z',
        'M2 7c0 2.76 4.48 5 10 5s10-2.24 10-5',
        'M2 12c0 2.76 4.48 5 10 5s10-2.24 10-5'
      ], 14, active ? 'currentColor' : 'currentColor', active ? 1 : 0.4)
    }
  } else if (option.type === 'database') {
    const isLoading = expandingDbKey.value === option.key
    
    if (isLoading) {
      elements.push(h('span', { class: 'conn-spinner', style: 'margin-right: 8px;' }))
    } else {
      if (option.isSystem) {
        // System databases: dimmed gear icon (40% opacity — visually "pushed back")
        iconElement = svgIcon([
          'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z',
          'M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z'
        ], 13, 'currentColor', 0.35)
      } else {
        // User databases: solid filled cylinder — confident, content-bearing
        iconElement = svgIcon([
          'M12 2C6.48 2 2 4.24 2 7v10c0 2.76 4.48 5 10 5s10-2.24 10-5V7c0-2.76-4.48-5-10-5z',
          'M2 7c0 2.76 4.48 5 10 5s10-2.24 10-5',
          'M2 12c0 2.76 4.48 5 10 5s10-2.24 10-5'
        ], 14, 'currentColor', 0.82)
      }
    }
  } else if (option.type === 'schema') {
    // Schema icon — always dark mode: invert to white with soft glow
    const iconStyle = 'width: 14px; height: 14px; margin-right: 12px; flex-shrink: 0; filter: brightness(0) invert(1) drop-shadow(0 0 2px rgba(255,255,255,0.3));'
    iconElement = h('img', { src: './scheme.svg', style: iconStyle })
  } else if (option.type === 'object-group') {
    if (option.groupKind === 'view') {
      iconElement = svgIcon([
        'M2 12s4-5 10-5 10 5 10 5-4 5-10 5-10-5-10-5z',
        'M12 14a2 2 0 1 0 0-4 2 2 0 0 0 0 4z'
      ], 13, 'currentColor', 0.46)
    } else {
      iconElement = svgIcon([
        'M4 5h16v14H4z',
        'M4 10h16',
        'M10 5v14'
      ], 13, 'currentColor', 0.42)
    }
  } else if (option.type === 'table') {
    // Table/grid icon — slightly dimmed to show table is a leaf
    iconElement = svgIcon([
      'M3 3h18v18H3z',
      'M3 9h18',
      'M3 15h18',
      'M9 3v18',
      'M15 3v18'
    ], 13, 'currentColor', 0.45)
  } else if (option.type === 'view') {
    iconElement = svgIcon([
      'M2 12s4-6 10-6 10 6 10 6-4 6-10 6-10-6-10-6z',
      'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z'
    ], 13, 'currentColor', 0.55)
  }
  
  if (iconElement) {
    elements.push(iconElement)
  }

  if (option.opened && (option.type === 'database' || option.type === 'table' || option.type === 'view')) {
    elements.push(h('span', { class: ['opened-indicator', option.active ? 'opened-indicator--active' : ''] }))
  }
  
  return h('span', { style: 'display: flex; align-items: center;' }, elements)
}

// Context menu options
const contextMenuOptions = ref([])

watch(contextMenuNode, (node) => {
  if (!node) return
  
  if (node.type === 'connection') {
    contextMenuOptions.value = [
      { label: t('tree.refresh'), key: 'refresh' },
      { label: t('tree.newDatabase'), key: 'new-database' },
      { label: t('tree.importDatabase'), key: 'import-database' },
      { label: t('tree.newQuery'), key: 'new-query' },
      { label: t('tree.copyName'), key: 'copy-name' },
      { label: t('tree.editConn'), key: 'edit' },
      { type: 'divider' },
      { label: t('tree.disconnect'), key: 'disconnect' },
      { label: t('tree.deleteConn'), key: 'delete-connection', props: { style: { color: '#ff4d4f' } } }
    ]
  } else if (node.type === 'database') {
    contextMenuOptions.value = [
      { label: t('tree.overview'), key: 'overview' },
      { label: t('tree.refreshDb'), key: 'refresh-db' },
      { label: t('tree.newDatabase'), key: 'new-database' },
      { label: t('tree.exportDatabase'), key: 'export-database' },
      { label: t('tree.importDatabase'), key: 'import-database' },
      { label: t('tree.newQuery'), key: 'new-query' },
      { label: t('tree.copyName'), key: 'copy-name' },
      { type: 'divider' },
      {
        label: t('tree.dropDb'),
        key: 'drop-db',
        disabled: isProtectedMySQLDatabase(node),
        props: { style: { color: '#ff4d4f' } }
      }
    ]
  } else if (node.type === 'table') {
    contextMenuOptions.value = [
      { label: t('tree.viewData'), key: 'view-data' },
      { label: t('tree.viewStructure'), key: 'view-structure' },
      { label: t('tree.queryTable'), key: 'query-table' },
      { label: t('tree.renameTable'), key: 'rename-table' },
      { label: t('tree.copyName'), key: 'copy-name' },
      { label: t('tree.copyDDL'), key: 'copy-ddl' },
      { type: 'divider' },
      { label: t('tree.exportData'), key: 'export-data' },
      { label: t('tree.importData'), key: 'import-data' },
      { type: 'divider' },
      { label: t('tree.truncateTable'), key: 'truncate-table' },
      { label: t('tree.dropTable'), key: 'drop-table', props: { style: { color: '#ff4d4f' } } }
    ]
  } else if (node.type === 'view') {
    contextMenuOptions.value = [
      { label: t('tree.viewData'), key: 'view-data' },
      { label: t('tree.viewStructure'), key: 'view-structure' },
      { label: t('tree.properties'), key: 'properties' },
      { label: t('tree.queryTable'), key: 'query-table' },
      {
        label: t('tree.generateSQL'),
        key: 'generate-sql',
        children: [
          { label: t('tree.sqlSelect'), key: 'sql-select' },
          { label: t('tree.sqlCount'), key: 'sql-count' },
          { label: t('tree.sqlCreate'), key: 'sql-create' }
        ]
      },
      { label: t('tree.copyName'), key: 'copy-name' },
      { label: t('tree.copyDDL'), key: 'copy-ddl' },
      { type: 'divider' },
      { label: t('tree.exportData'), key: 'export-data' }
    ]
  }
})

const closeContextMenu = () => {
  showContextMenu.value = false
}

const openRenameDialog = (node) => {
  renameTarget.value = {
    connId: node.connId,
    database: node.dbName,
    oldName: node.tableName,
    newName: node.tableName
  }
  renameDialogVisible.value = true
}

const closeDropDatabaseDialog = () => {
  dropDatabaseDialogVisible.value = false
  dropDatabaseSubmitting.value = false
  dropDatabaseTarget.value = {
    connId: '',
    dbName: '',
    confirmText: ''
  }
}

const openDropDatabaseDialog = (node) => {
  if (isProtectedMySQLDatabase(node)) {
    gmssh.warning(t('tree.systemDbProtected', { name: node.dbName }))
    return
  }

  dropDatabaseTarget.value = {
    connId: node.connId,
    dbName: node.dbName,
    confirmText: ''
  }
  dropDatabaseDialogVisible.value = true
}

const submitDropDatabase = async () => {
  if (!isDropDatabaseConfirmMatched.value) {
    gmssh.warning(t('tree.dropDbConfirmMismatch'))
    return
  }

  dropDatabaseSubmitting.value = true
  try {
    const runtimeConnId = await ensureConnectionContext(dropDatabaseTarget.value.connId)
    await store.dropDatabase(dropDatabaseTarget.value.dbName, runtimeConnId)
    gmssh.success(t('tree.dropDbSuccess'))
    closeDropDatabaseDialog()
  } catch (error) {
    gmssh.error(t('tree.dropDbFailed', { msg: error.message }))
  } finally {
    dropDatabaseSubmitting.value = false
  }
}

const copyNodeName = async (node) => {
  const value = node?.tableName || node?.dbName || node?.label || ''
  const copied = await gmssh.copyToClipboard(value)
  if (copied) {
    gmssh.success(t('tree.copyNameSuccess'))
  } else {
    gmssh.warning(t('designer.copyUnavailable'))
  }
}

const copyNodeDDL = async (node) => {
  try {
    const runtimeConnId = await ensureConnectionContext(node.connId)
    const ddl = await api.getTableDDL(runtimeConnId, node.dbName, node.tableName)
    const copied = await gmssh.copyToClipboard(ddl)
    if (copied) {
      gmssh.success(t('overview.ddlCopied'))
    } else {
      gmssh.warning(t('designer.copyUnavailable'))
    }
  } catch (error) {
    gmssh.error(t('tree.copyDDLFailed', { msg: error.message }))
  }
}

const buildPredicateTemplate = (columns = [], primaryKeys = [], dbType = 'mysql') => {
  const keyColumns = primaryKeys.length > 0
    ? primaryKeys
    : (columns[0]?.name ? [columns[0].name] : [])

  if (keyColumns.length === 0) {
    return '/* condition */ 1 = 0'
  }

  return keyColumns
    .map((column) => `${quoteIdentifier(column, dbType)} = /* value */`)
    .join('\n  AND ')
}

const buildSqlTemplate = async (node, templateKey) => {
  const runtimeConnId = await ensureConnectionContext(node.connId)
  const dbType = getConnectionDbType(node.connId)
  const objectName = quoteIdentifier(node.tableName, dbType)
  const schema = await api.getTableSchema(runtimeConnId, node.dbName, node.tableName).catch(() => null)
  const columns = Array.isArray(schema?.columns) ? schema.columns : []
  const primaryKeys = Array.isArray(schema?.primaryKey) ? schema.primaryKey : []
  const columnNames = columns.map((column) => column.name).filter(Boolean)
  const quotedColumns = columnNames.map((name) => quoteIdentifier(name, dbType))

  switch (templateKey) {
    case 'sql-select':
      return `SELECT *\nFROM ${objectName}\nLIMIT 200;`
    case 'sql-count':
      return `SELECT COUNT(*) AS total\nFROM ${objectName};`
    case 'sql-insert':
      if (quotedColumns.length === 0) {
        return `INSERT INTO ${objectName} (\n  /* columns */\n) VALUES (\n  /* values */\n);`
      }
      return `INSERT INTO ${objectName} (\n  ${quotedColumns.join(',\n  ')}\n) VALUES (\n  ${quotedColumns.map(() => '/* value */ NULL').join(',\n  ')}\n);`
    case 'sql-update':
      if (quotedColumns.length === 0) {
        return `UPDATE ${objectName}\nSET /* column */ = /* value */\nWHERE /* condition */ 1 = 0;`
      }
      return `UPDATE ${objectName}\nSET\n  ${quotedColumns.map((column) => `${column} = /* value */`).join(',\n  ')}\nWHERE ${buildPredicateTemplate(columns, primaryKeys, dbType)};`
    case 'sql-delete':
      return `DELETE FROM ${objectName}\nWHERE ${buildPredicateTemplate(columns, primaryKeys, dbType)};`
    case 'sql-create':
      return await api.getTableDDL(runtimeConnId, node.dbName, node.tableName)
    default:
      return ''
  }
}

const openGeneratedSql = async (node, templateKey) => {
  try {
    const runtimeConnId = await ensureConnectionContext(node.connId)
    const sql = await buildSqlTemplate(node, templateKey)
    window.dispatchEvent(new CustomEvent('open-query', {
      detail: {
        database: node.dbName,
        table: node.tableName,
        connId: runtimeConnId,
        sql
      }
    }))
  } catch (error) {
    gmssh.error(t('tree.generateSQLFailed', { msg: error.message }))
  }
}

const openObjectProperties = async (node) => {
  const runtimeConnId = await ensureConnectionContext(node.connId)
  window.dispatchEvent(new CustomEvent('open-object-properties', {
    detail: { database: node.dbName, table: node.tableName, connId: runtimeConnId }
  }))
}

const dispatchCreateDatabase = async (connId) => {
  const runtimeConnId = await ensureConnectionContext(connId)
  window.dispatchEvent(new CustomEvent('open-create-database', {
    detail: { connId: runtimeConnId }
  }))
}

const dispatchNewQuery = async (node) => {
  const runtimeConnId = await ensureConnectionContext(node.connId)
  window.dispatchEvent(new CustomEvent('open-query', {
    detail: { database: node?.dbName || '', connId: runtimeConnId }
  }))
}

const openTransferCenter = async (node, mode) => {
  const runtimeConnId = await ensureConnectionContext(node.connId)
  window.dispatchEvent(new CustomEvent('open-transfer-center', {
    detail: {
      connId: runtimeConnId,
      database: node.type === 'database' ? node.dbName : '',
      action: mode,
      autoLaunch: true
    }
  }))
}

const truncateTable = async (node) => {
  const confirmed = await gmssh.confirm({
    title: t('tree.truncateTableTitle'),
    content: t('tree.truncateTableContent', { name: node.tableName }),
    positiveText: t('tree.truncateTableOk'),
    negativeText: t('tree.truncateTableCancel')
  })

  if (!confirmed) return

  try {
    const runtimeConnId = await ensureConnectionContext(node.connId)
    const dbType = getConnectionDbType(node.connId)
    const objectName = quoteIdentifier(node.tableName, dbType)
    const sql = dbType === 'sqlite'
      ? `DELETE FROM ${objectName};`
      : `TRUNCATE TABLE ${objectName};`

    await api.executeSQL(runtimeConnId, node.dbName, sql)
    window.dispatchEvent(new CustomEvent('refresh-table-data', {
      detail: { database: node.dbName, table: node.tableName, connId: runtimeConnId }
    }))
    gmssh.success(t('tree.truncateTableSuccess'))
  } catch (error) {
    gmssh.error(t('tree.truncateTableFailed', { msg: error.message }))
  }
}

const submitRename = async () => {
  const payload = renameTarget.value
  const newName = payload.newName.trim()

  if (!newName) {
    gmssh.warning(t('tree.renameNameRequired'))
    return
  }

  if (newName === payload.oldName) {
    renameDialogVisible.value = false
    return
  }

  renameSubmitting.value = true
  try {
    const runtimeConnId = await ensureConnectionContext(payload.connId)
    await api.renameTable(runtimeConnId, payload.database, payload.oldName, newName)
    await loadTablesForDatabase(payload.connId, payload.database)
    window.dispatchEvent(new CustomEvent('table-renamed', {
      detail: {
        connId: runtimeConnId,
        database: payload.database,
        oldName: payload.oldName,
        newName
      }
    }))
    gmssh.success(t('tree.renameSuccess', { name: newName }))
    renameDialogVisible.value = false
  } catch (error) {
    gmssh.error(t('tree.renameFailed', { msg: error.message }))
  } finally {
    renameSubmitting.value = false
  }
}

const handleContextMenuSelect = async (key) => {
  closeContextMenu()
  const node = contextMenuNode.value
  
  switch (key) {
    case 'refresh':
      await ensureConnectionContext(node.connId)
      await store.loadDatabases()
      gmssh.success(t('tree.refreshSuccess'))
      break
      
    case 'edit':
      // Emit event to parent to show edit dialog
      emit('edit-connection', node.connData)
      break

    case 'new-database':
      await dispatchCreateDatabase(node.connId)
      break

    case 'new-query':
      await dispatchNewQuery(node)
      break

    case 'export-database':
      await openTransferCenter(node, 'export')
      break

    case 'import-database':
      await openTransferCenter(node, 'import')
      break
      
    case 'disconnect':
      {
        const runtimeConnId = store.getRuntimeConnIdByConfigId(node.connId)
        if (!runtimeConnId) break
        const confirmed = await gmssh.confirm({
          title: t('tree.disconnectConfirmTitle'),
          content: t('tree.disconnectConfirmContent'),
          positiveText: t('tree.disconnectConfirmOk'),
          negativeText: t('tree.disconnectConfirmCancel')
        })
        
        if (confirmed) {
          await store.disconnect(runtimeConnId)
          gmssh.success(t('tree.disconnected'))
        }
      }
      break
      
    case 'delete-connection':
      {
        const confirmed = await gmssh.confirm({
          title: t('tree.deleteConnTitle'),
          content: t('tree.deleteConnContent', { name: node.label }),
          positiveText: t('tree.deleteConnOk'),
          negativeText: t('tree.deleteConnCancel')
        })
        
        if (confirmed) {
          try {
            // If this connection is currently active, disconnect first
            const runtimeConnId = store.getRuntimeConnIdByConfigId(node.connId)
            if (runtimeConnId) {
              await store.disconnect(runtimeConnId)
            }
            
            // Delete the connection from backend
            await connectionApi.deleteConnection(node.connId)
            
            // Refresh connection list
            await loadConnections()
            
            gmssh.success(t('tree.deleteConnSuccess'))
          } catch (error) {
            gmssh.error(t('tree.deleteConnFailed', { msg: error.message }))
          }
        }
      }
      break

    case 'overview':
      await openDatabaseOverview(node)
      break
      
    case 'refresh-db':
      await loadTablesForDatabase(node.connId, node.dbName)
      gmssh.success(t('tree.refreshSuccess'))
      break
      
    case 'copy-name':
      await copyNodeName(node)
      break

    case 'copy-ddl':
      await copyNodeDDL(node)
      break

    case 'properties':
      await openObjectProperties(node)
      break

    case 'rename-table':
      openRenameDialog(node)
      break
      
    case 'drop-db':
      openDropDatabaseDialog(node)
      break
      
    case 'view-data':
      {
        const runtimeConnId = await ensureConnectionContext(node.connId)
      window.dispatchEvent(new CustomEvent('open-table-data', {
          detail: { database: node.dbName, table: node.tableName, connId: runtimeConnId }
      }))
      }
      break
      
    case 'view-structure':
      {
        const runtimeConnId = await ensureConnectionContext(node.connId)
      window.dispatchEvent(new CustomEvent('open-structure', {
          detail: { database: node.dbName, table: node.tableName, connId: runtimeConnId }
      }))
      }
      break

    case 'query-table':
      {
        const runtimeConnId = await ensureConnectionContext(node.connId)
      window.dispatchEvent(new CustomEvent('open-query', {
          detail: { database: node.dbName, table: node.tableName, connId: runtimeConnId }
      }))
      }
      break

    case 'sql-select':
    case 'sql-count':
    case 'sql-insert':
    case 'sql-update':
    case 'sql-delete':
    case 'sql-create':
      await openGeneratedSql(node, key)
      break

    case 'export-data':
      {
        const runtimeConnId = await ensureConnectionContext(node.connId)
      window.dispatchEvent(new CustomEvent('open-export', {
          detail: { database: node.dbName, table: node.tableName, connId: runtimeConnId }
      }))
      }
      break

    case 'import-data':
      {
        const runtimeConnId = await ensureConnectionContext(node.connId)
      window.dispatchEvent(new CustomEvent('open-import', {
          detail: { database: node.dbName, table: node.tableName, connId: runtimeConnId }
      }))
      }
      break

    case 'truncate-table':
      await truncateTable(node)
      break
      
    case 'drop-table':
      {
        const confirmed = await gmssh.confirm({
          title: t('tree.dropTableTitle'),
          content: t('tree.dropTableContent', { name: node.tableName }),
          positiveText: t('tree.dropTableOk'),
          negativeText: t('tree.dropTableCancel')
        })
        
        if (confirmed) {
          try {
            const runtimeConnId = await ensureConnectionContext(node.connId)
            await api.dropTable(runtimeConnId, node.dbName, node.tableName)
            await loadTablesForDatabase(node.connId, node.dbName)
            gmssh.success(t('tree.dropTableSuccess'))
          } catch (error) {
            gmssh.error(t('tree.dropTableFailed', { msg: error.message }))
          }
        }
      }
      break
  }
}

// Watch for database changes to rebuild tree
watch(() => store.databases, () => {
  buildTreeData()
}, { deep: true })

// Watch for connection status changes
watch(() => store.currentConnId, () => {
  buildTreeData()
  // Auto-expand current connection
  if (store.isConnected && store.connectionConfig?.id) {
    const connKey = `conn-${store.connectionConfig.id}`
    if (!expandedKeys.value.includes(connKey)) {
      expandedKeys.value = [...expandedKeys.value, connKey]
    }
  }
})

watch(() => store.connections.map(item => `${item.connId}:${item.config?.id || ''}`).join('|'), () => {
  buildTreeData()
})

watch(treeSearch, () => {
  buildTreeData()
})

const handleDatabaseDumpFinished = async (event) => {
  const { connId } = event.detail || {}
  if (!connId) return

  databaseObjectCache.value = {}
  await store.loadDatabases(connId).catch(() => null)
  buildTreeData()
}

onMounted(() => {
  // Load connections asynchronously without blocking initial render
  loadConnections().then(() => {
    // If already connected, expand and load databases
    if (store.isConnected && store.connectionConfig?.id) {
      store.loadDatabases().then(() => {
        buildTreeData()
        expandedKeys.value = [`conn-${store.connectionConfig.id}`]
      })
    }
  })
  
  // Listen for connection refresh events
  window.addEventListener('refresh-connections', loadConnections)
  window.addEventListener('workspace-tabs-updated', handleWorkspaceTabsUpdated)
  window.addEventListener('database-dump-finished', handleDatabaseDumpFinished)
})

onUnmounted(() => {
  window.removeEventListener('refresh-connections', loadConnections)
  window.removeEventListener('workspace-tabs-updated', handleWorkspaceTabsUpdated)
  window.removeEventListener('database-dump-finished', handleDatabaseDumpFinished)
})
</script>

<style scoped>
/* ─── wrapper ─────────────────────────────────────────── */
/* NO height/overflow here — let .sidebar-scroll control scrolling */
.database-tree-wrapper {
  width: 100%;
  box-sizing: border-box;
  padding: 8px 10px 20px;
  /* height is auto — content determines it */
}

.tree-toolbar {
  padding: 2px 2px 12px;
}

.tree-search {
  width: 100%;
}

.tree-search :deep(.n-input-wrapper) {
  background: rgba(255, 255, 255, 0.04) !important;
  border: 1px solid rgba(255, 255, 255, 0.06) !important;
  box-shadow: none !important;
}

.tree-search :deep(.n-input__input-el) {
  color: var(--sys-color-text-primary) !important;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  gap: 12px;
}

.loading-text {
  color: rgba(255, 255, 255, 0.5);
  font-size: 13px;
}

/* ── Empty state: minimal ───────────────────────────────── */
.empty-state-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 40px 20px 24px;
  text-align: center;
}

.empty-icon {
  color: var(--ref-color-white-20);
  margin-bottom: 2px;
  line-height: 0;
}

.empty-label {
  font-size: var(--ref-font-size-sm);
  color: var(--sys-color-text-tertiary);
  font-weight: var(--ref-font-weight-medium);
  letter-spacing: 0.01em;
}

/* Text-style + button: no filled bg, just subtle hover */
.empty-add-btn {
  all: unset;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 14px;
  border-radius: var(--ref-radius-pill);
  border: 1px solid var(--ref-color-white-15);
  font-size: var(--ref-font-size-xs);
  font-weight: var(--ref-font-weight-medium);
  color: var(--sys-color-text-secondary);
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease,
    color 0.18s ease,
    box-shadow 0.18s ease,
    opacity 0.18s ease;
  font-family: var(--ref-font-family-base);
}

.empty-add-btn:hover {
  border-color: var(--ref-color-brand-6);
  color: var(--ref-color-brand-5);
  background: rgba(87,114,255,0.08);
}

.empty-add-btn:active {
  background: rgba(87,114,255,0.14);
  border-color: var(--ref-color-brand-7);
}

/* ── tree node shell ────────────────────────────────── */
:deep(.n-tree-node-wrapper) {
  /* overflow managed by .sidebar-scroll native div */
  min-width: 0;
  width: 100%;
  padding: 0 !important;
}

:deep(.n-tree-node) {
  align-items: center;
  overflow: visible;
  width: 100%;
  min-width: 0;
  min-height: 40px;
  padding: 0 0 0 2px;
  margin-bottom: 3px;
  border-radius: var(--ref-radius-sm);
  position: relative;
  transition:
    background-color 0.18s ease,
    color 0.18s ease;
}

:deep(.tree-node--root) {
  min-height: 28px;
  margin: 8px 0 6px;
}

:deep(.tree-node--object-group) {
  min-height: 32px;
  margin-top: 4px;
}

:deep(.n-tree-node)::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  width: 3px;
  height: 16px;
  border-radius: 2px;
  background: var(--comp-sidebar-nav-item-selected-bar);
  opacity: 0;
  transform: translateY(-50%);
  transition: opacity 0.18s ease;
}

/* ─── row content ──────────────────────────────────────── */
:deep(.n-tree-node-content) {
  width: auto;
  box-sizing: border-box;
  gap: 8px;
  padding: 0 12px 0 8px;
  height: 40px;
  min-height: 40px;
  overflow: visible;
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  background: transparent;   /* no !important — parent .n-tree-node controls bg */
  color: inherit;
  border-radius: 0;
  position: static;
  cursor: pointer;
  transition: color 0.18s ease;
}

:deep(.tree-node--root .n-tree-node-content) {
  min-height: 28px;
  height: 28px;
  padding: 0 8px 0 6px;
  background: transparent !important;
  border-radius: 0;
  color: var(--sys-color-text-tertiary);
}

:deep(.tree-node--object-group .n-tree-node-content) {
  min-height: 32px;
  height: 32px;
  padding: 0 12px 0 6px;
  cursor: default;
}

/* ─ hover: bg-white-8, label goes full white ──────── */
:deep(.n-tree-node:hover) {
  background: rgba(255, 255, 255, 0.07) !important;
  color: var(--sys-color-text-title) !important;
}
:deep(.n-tree-node:hover) .tree-node-label {
  color: var(--sys-color-text-title);
}

:deep(.n-tree-node.is-opened .tree-node-label) {
  color: var(--sys-color-text-primary);
}

:deep(.n-tree-node.is-active) {
  background: rgba(255, 255, 255, 0.07) !important;
}

:deep(.n-tree-node.is-active)::before {
  opacity: 1;
}

:deep(.tree-node--root:hover),
:deep(.tree-node--root:active),
:deep(.tree-node--object-group:hover),
:deep(.tree-node--object-group:active) {
  background: transparent !important;
  color: var(--sys-color-text-tertiary) !important;
}

/* ─ selected: bg-white-8 + brand-6 left bar (3×16px) ── */
/* --comp-sidebar-nav-item-bg-selected = var(--ref-color-white-8) */
:deep(.n-tree-node.n-tree-node--selected) {
  background: rgba(255, 255, 255, 0.07) !important;
  color: var(--sys-color-text-title) !important;
}
:deep(.n-tree-node.n-tree-node--selected) .tree-node-label {
  color: var(--sys-color-text-title);
  font-weight: var(--ref-font-weight-medium);
}

/* Reveal left bar on selected */
:deep(.n-tree-node.n-tree-node--selected)::before {
  opacity: 1;
}

/* Root group header: never gets selected-bar or heavy bg */
:deep(.tree-node--root)::before {
  display: none;
}

:deep(.tree-node--object-group)::before {
  display: none;
}

:deep(.n-tree-node:active) {
  background: rgba(255,255,255,0.10) !important;
}

:deep(.tree-node--connection.is-current) {
  background: transparent !important;
}

:deep(.tree-node--connection.is-current)::before {
  opacity: 0;
}

:deep(.tree-node--connection.is-current .tree-node-label) {
  color: var(--sys-color-text-title);
  font-weight: var(--ref-font-weight-medium);
}

:deep(.tree-node--connection.n-tree-node--selected) {
  background: transparent !important;
}

:deep(.tree-node--connection.n-tree-node--selected)::before {
  opacity: 0;
}

:deep(.tree-node--connection:hover) {
  background: var(--comp-sidebar-nav-item-bg-hover) !important;
}

/* tree indentation should stay clean in sidebar navigation */
:deep(.n-tree-node-indent) {
  width: 14px !important;
  position: relative;
  flex-shrink: 0;
}
:deep(.n-tree-node-indent::before) {
  display: none;
}

:deep(.n-tree-node-switcher) {
  width: 20px;
  height: 40px;
  margin-right: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: currentColor;
  border-radius: 0;
  transition:
    background-color 0.18s ease,
    color 0.18s ease;
}

:deep(.n-tree-node-switcher:hover) {
  background: transparent;
  color: var(--sys-color-text-secondary);
}

:deep(.tree-node--root .n-tree-node-switcher) {
  width: 16px;
  height: 28px;
  color: var(--sys-color-text-tertiary);
}

:deep(.n-tree-node-switcher__icon) {
  transition: transform 0.18s ease;
}

/* ─── prefix / text layout ─────────────────────────────── */
:deep(.n-tree-node-content__prefix) {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  margin-right: 8px;
}

:deep(.n-tree-node-content svg),
:deep(.n-tree-node-content img) {
  width: 16px;
  height: 16px;
  margin-right: 0 !important;
  opacity: 1 !important;
  transform: none !important;
}

:deep(.n-tree-node-content__text) {
  overflow: hidden;
  flex: 1;
  min-width: 0;
}

:deep(.n-tree-node-content__suffix) {
  display: flex;
  align-items: center;
  margin-left: auto;
  width: 0;
  overflow: hidden;
}

:deep(.n-tree-node-content__label) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  display: block;
}

/* ─── label typography ─────────────────────────────────── */
.tree-node-label {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
  font-family: var(--ref-font-family-base);
  font-size: var(--ref-font-size-sm);
  font-weight: var(--ref-font-weight-regular);
  color: var(--comp-sidebar-nav-item-text-default);
  transition: color 0.15s ease;
}

.tree-node-label-wrap {
  display: inline-flex;
  align-items: center;
  gap: var(--ref-space-6);
  min-width: 0;
  max-width: 100%;
}

.tree-node-label-wrap--object-group {
  width: 100%;
}

.tree-node-meta,
.tree-node-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 18px;
  padding: 0 var(--ref-space-6);
  border-radius: var(--ref-radius-pill);
  font-size: var(--ref-font-size-xs);
  line-height: 1;
  flex-shrink: 0;
}

.tree-node-meta {
  background: var(--ref-color-white-8);
  color: var(--sys-color-text-tertiary);
}

.tree-node-meta--current {
  background: rgba(87, 114, 255, 0.16);
  color: var(--ref-color-brand-4);
}

.tree-node-meta--connected {
  background: rgba(50, 178, 93, 0.14);
  color: var(--ref-color-green-6);
}

.tree-node-meta--saved {
  background: var(--ref-color-white-6);
  color: var(--sys-color-text-tertiary);
}

.tree-node-pill {
  min-width: 18px;
  background: var(--ref-color-white-6);
  color: var(--sys-color-text-secondary);
}

.opened-indicator {
  width: 6px;
  height: 6px;
  margin-left: 8px;
  border-radius: 50%;
  background: var(--ref-color-white-25);
  flex-shrink: 0;
}

.opened-indicator--active {
  background: var(--ref-color-brand-6);
  box-shadow: 0 0 8px rgba(87, 114, 255, 0.55);
}

.tree-node-label--root {
  font-size: var(--ref-font-size-xs);
  font-weight: var(--ref-font-weight-medium);
  letter-spacing: 0.04em;
  color: var(--sys-color-text-tertiary);
  text-transform: uppercase;
}

.tree-node-label--object-group {
  font-size: var(--ref-font-size-xs);
  letter-spacing: 0.04em;
  color: var(--sys-color-text-tertiary);
  text-transform: uppercase;
}

:deep(.tree-node--object-group .tree-node-label-wrap) {
  display: flex;
}

:deep(.tree-node--object-group .tree-node-label) {
  flex: 1;
}

:deep(.tree-node--database.is-system .tree-node-label) {
  color: var(--sys-color-text-tertiary);
}

:deep(.tree-node--database.is-system .n-tree-node-content svg) {
  color: var(--sys-color-text-tertiary);
}

:deep(.n-tree-node.n-tree-node--selected) .tree-node-label {
  font-weight: var(--ref-font-weight-medium);
}

:deep(.tree-node--object-group .tree-node-pill) {
  min-width: 22px;
  margin-left: auto;
  background: transparent;
  border: 1px solid var(--ref-color-white-10);
  color: var(--sys-color-text-tertiary);
}

:deep(.tree-node--object-group .n-tree-node-content svg) {
  color: var(--sys-color-text-tertiary);
}

:deep(.tree-node--group-view .n-tree-node-content svg) {
  color: var(--ref-color-cyan-6);
}

:deep(.tree-node--view .n-tree-node-content svg) {
  color: var(--ref-color-cyan-6);
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.drop-db-dialog {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.drop-db-dialog__lead {
  margin: 0;
  color: var(--sys-color-text-primary);
  line-height: 1.6;
}

.drop-db-dialog__hint {
  margin: 0;
  color: var(--sys-color-text-tertiary);
  font-size: var(--ref-font-size-sm);
  line-height: 1.6;
}

/* ─── status LED ───────────────────────────────────────── */
.status-led-active {
  animation: led-pulse-green 2s ease-in-out infinite;
}

/* ─── spinner ──────────────────────────────────────────── */
:deep(.conn-spinner) {
  display: inline-block;
  width: 9px;
  height: 9px;
  margin-right: 8px;
  border: 1.5px solid rgba(255, 255, 255, 0.15);
  border-top-color: var(--ref-color-brand-6);
  border-radius: 50%;
  animation: spin 0.75s linear infinite;
}

/* ─── keyframes ────────────────────────────────────────── */
@keyframes spin {
  from { transform: rotate(0deg); }
  to   { transform: rotate(360deg); }
}

@keyframes led-pulse-green {
  0%, 100% {
    opacity: 0.45;
    box-shadow: 0 0 3px rgba(50, 178, 93, 0.32), 0 0 7px rgba(50, 178, 93, 0.16);
  }
  50% {
    opacity: 1;
    box-shadow: 0 0 7px rgba(50, 178, 93, 0.8), 0 0 14px rgba(50, 178, 93, 0.4);
  }
}
</style>
