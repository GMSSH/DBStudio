import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import api from '@/utils/api'

function createConnectionState(connId, config = null) {
  return {
    connId,
    config: config ? { ...config } : null,
    databases: [],
    selectedDatabase: null,
    tables: [],
    selectedTable: null,
    tableData: null,
    tableSchema: null
  }
}

const MYSQL_SYSTEM_DATABASES = new Set(['information_schema', 'mysql', 'performance_schema', 'sys'])

export const useDatabaseStore = defineStore('database', () => {
  const activeMux = ref({})
  const connections = ref([])
  const currentConnId = ref(null)
  const isLoading = ref(false)
  const error = ref(null)

  function cloneConnectionState(state, connId, config = null) {
    return {
      connId,
      config: config ? { ...config } : state?.config ? { ...state.config } : null,
      databases: Array.isArray(state?.databases) ? [...state.databases] : [],
      selectedDatabase: state?.selectedDatabase || null,
      tables: Array.isArray(state?.tables) ? [...state.tables] : [],
      selectedTable: state?.selectedTable || null,
      tableData: state?.tableData || null,
      tableSchema: state?.tableSchema || null
    }
  }

  function ensureConnState(connId, config = null) {
    if (!connId) return null

    if (!activeMux.value[connId]) {
      activeMux.value = {
        ...activeMux.value,
        [connId]: createConnectionState(connId, config)
      }
    } else if (config) {
      activeMux.value[connId].config = {
        ...(activeMux.value[connId].config || {}),
        ...config
      }
    }

    return activeMux.value[connId]
  }

  function getConnectionState(connId = currentConnId.value) {
    if (!connId) return null
    return activeMux.value[connId] || null
  }

  function getRuntimeConnIdByConfigId(configId) {
    return connections.value.find((item) => item.config?.id === configId)?.connId || null
  }

  function getConnectionStateByConfigId(configId) {
    const connId = getRuntimeConnIdByConfigId(configId)
    return connId ? getConnectionState(connId) : null
  }

  function isConnectionActive(configId) {
    return !!getRuntimeConnIdByConfigId(configId)
  }

  function updateConnectionConfig(connId, partialConfig) {
    const state = ensureConnState(connId)
    if (!state) return

    state.config = {
      ...(state.config || {}),
      ...(partialConfig || {})
    }

    const index = connections.value.findIndex((item) => item.connId === connId)
    if (index >= 0) {
      connections.value[index] = {
        ...connections.value[index],
        config: {
          ...(connections.value[index].config || {}),
          ...(partialConfig || {})
        }
      }
    }
  }

  const currentConnection = computed(() => getConnectionState())
  const activeConnection = computed(() => currentConnId.value)
  const connectionConfig = computed(() => currentConnection.value?.config || null)
  const databases = computed({
    get: () => currentConnection.value?.databases || [],
    set: (value) => {
      const state = ensureConnState(currentConnId.value)
      if (state) state.databases = value || []
    }
  })
  const selectedDatabase = computed({
    get: () => currentConnection.value?.selectedDatabase || null,
    set: (value) => {
      const state = ensureConnState(currentConnId.value)
      if (state) state.selectedDatabase = value || null
    }
  })
  const tables = computed({
    get: () => currentConnection.value?.tables || [],
    set: (value) => {
      const state = ensureConnState(currentConnId.value)
      if (state) state.tables = value || []
    }
  })
  const selectedTable = computed({
    get: () => currentConnection.value?.selectedTable || null,
    set: (value) => {
      const state = ensureConnState(currentConnId.value)
      if (state) state.selectedTable = value || null
    }
  })
  const tableData = computed({
    get: () => currentConnection.value?.tableData || null,
    set: (value) => {
      const state = ensureConnState(currentConnId.value)
      if (state) state.tableData = value || null
    }
  })
  const tableSchema = computed({
    get: () => currentConnection.value?.tableSchema || null,
    set: (value) => {
      const state = ensureConnState(currentConnId.value)
      if (state) state.tableSchema = value || null
    }
  })
  const isConnected = computed(() => !!currentConnId.value)

  async function switchConnection(connId) {
    if (!activeMux.value[connId]) {
      throw new Error('Connection not found')
    }
    currentConnId.value = connId
    return getConnectionState(connId)
  }

  async function addConnection(config, existingConnId = null) {
    let connId = existingConnId
    if (!connId) {
      const result = await api.connect(config)
      connId = result?.connId || result
    }

    if (!connId) {
      throw new Error('No connection ID returned from server')
    }

    ensureConnState(connId, config)

    const existingIndex = connections.value.findIndex((item) => item.connId === connId)
    if (existingIndex >= 0) {
      connections.value[existingIndex] = { connId, config: { ...config } }
    } else {
      connections.value = [...connections.value, { connId, config: { ...config } }]
    }

    currentConnId.value = connId
    return connId
  }

  async function connect(config) {
    isLoading.value = true
    error.value = null

    try {
      if (config?.id) {
        const existingConnId = getRuntimeConnIdByConfigId(config.id)
        if (existingConnId) {
          updateConnectionConfig(existingConnId, config)
          await switchConnection(existingConnId)
          if ((getConnectionState(existingConnId)?.databases || []).length === 0) {
            await loadDatabases(existingConnId)
          }
          return { connId: existingConnId }
        }
      }

      const connId = await addConnection(config)
      await loadDatabases(connId)
      return { connId }
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function recoverConnection(connId) {
    if (!connId) {
      throw new Error('No active connection')
    }

    const state = getConnectionState(connId)
    const connection = connections.value.find((item) => item.connId === connId)
    const config = {
      ...(connection?.config || {}),
      ...(state?.config || {})
    }

    if (!config || Object.keys(config).length === 0) {
      throw new Error('Connection config not found')
    }

    const restoredState = cloneConnectionState(state, connId, config)
    const result = await api.connect(config)
    const newConnId = result?.connId || result

    if (!newConnId) {
      throw new Error('No connection ID returned from server')
    }

    const nextMux = { ...activeMux.value }
    delete nextMux[connId]
    nextMux[newConnId] = cloneConnectionState(restoredState, newConnId, config)
    activeMux.value = nextMux

    const nextConnections = connections.value.filter((item) => item.connId !== connId && item.connId !== newConnId)
    nextConnections.push({ connId: newConnId, config: { ...config } })
    connections.value = nextConnections

    if (currentConnId.value === connId) {
      currentConnId.value = newConnId
    }

    return { connId: newConnId }
  }

  async function removeConnection(connId = currentConnId.value, shouldDisconnect = true) {
    if (!connId) return

    if (shouldDisconnect) {
      try {
        await api.disconnect(connId)
      } catch (err) {
        error.value = err.message
        throw err
      }
    }

    const nextMux = { ...activeMux.value }
    delete nextMux[connId]
    activeMux.value = nextMux
    connections.value = connections.value.filter((item) => item.connId !== connId)

    if (currentConnId.value === connId) {
      currentConnId.value = connections.value[0]?.connId || null
    }
  }

  async function disconnect(connId = currentConnId.value) {
    await removeConnection(connId, true)
  }

  async function loadDatabases(connId = currentConnId.value) {
    if (!connId) {
      throw new Error('No active connection')
    }

    isLoading.value = true
    try {
      const result = await api.listDatabases(connId)
      const state = ensureConnState(connId)
      state.databases = result || []
      return state.databases
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function selectDatabase(dbName, connId = currentConnId.value) {
    if (!connId) {
      throw new Error('No active connection')
    }

    const state = ensureConnState(connId)
    if (!state) {
      throw new Error('Connection not found')
    }

    if (currentConnId.value !== connId) {
      currentConnId.value = connId
    }

    if (state.selectedDatabase === dbName && state.tables.length > 0) {
      return state.tables
    }

    isLoading.value = true
    error.value = null

    try {
      await api.switchDatabase(connId, dbName)
      state.selectedDatabase = dbName
      state.tables = []
      state.selectedTable = null
      state.tableData = null
      state.tableSchema = null
      updateConnectionConfig(connId, { database: dbName })
      await loadTables(connId, dbName)
      return state.tables
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function loadTables(connId = currentConnId.value, database = null) {
    const state = getConnectionState(connId)
    const dbName = database || state?.selectedDatabase
    if (!connId || !dbName) return []

    isLoading.value = true
    try {
      const result = await api.listTables(connId, dbName)
      state.tables = result || []
      return state.tables
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function selectTable(tableName, connId = currentConnId.value) {
    const state = ensureConnState(connId)
    if (!state) {
      throw new Error('No active connection')
    }

    currentConnId.value = connId
    state.selectedTable = tableName

    await Promise.all([
      loadTableSchema(connId),
      loadTableData(1, 500, connId)
    ])
  }

  async function loadTableSchema(connId = currentConnId.value, database = null, table = null) {
    const state = getConnectionState(connId)
    const dbName = database || state?.selectedDatabase
    const tableName = table || state?.selectedTable
    if (!connId || !dbName || !tableName) return null

    try {
      const result = await api.getTableSchema(connId, dbName, tableName)
      if (!table || table === state.selectedTable) {
        state.tableSchema = result
      }
      return result
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function getTableSchema() {
    return tableSchema.value || await loadTableSchema()
  }

  async function loadTableData(page = 1, pageSize = 100, connId = currentConnId.value, database = null, table = null, options = {}) {
    const state = getConnectionState(connId)
    const dbName = database || state?.selectedDatabase
    const tableName = table || state?.selectedTable
    if (!connId || !dbName || !tableName) return null

    isLoading.value = true
    try {
      const result = await api.getTableData(
        connId,
        dbName,
        tableName,
        page,
        pageSize,
        options.sortCol || '',
        options.sortDir || ''
      )

      if (!table || table === state.selectedTable) {
        state.tableData = result
      }
      return result
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function getTableData(tableName, page = 1, pageSize = 100, database = null, options = {}, connId = currentConnId.value) {
    const state = getConnectionState(connId)
    const dbName = database || state?.selectedDatabase
    if (!connId || !dbName) {
      throw new Error('No active connection or database')
    }

    return api.getTableData(
      connId,
      dbName,
      tableName,
      page,
      pageSize,
      options.sortCol || '',
      options.sortDir || ''
    )
  }

  async function getTableSchemaFor(tableName, database = null, connId = currentConnId.value) {
    const state = getConnectionState(connId)
    const dbName = database || state?.selectedDatabase
    if (!connId || !dbName) {
      throw new Error('No active connection or database')
    }

    return api.getTableSchema(connId, dbName, tableName)
  }

  async function executeSQL(sql, database = null, connId = currentConnId.value) {
    const state = getConnectionState(connId)
    const dbName = database || state?.selectedDatabase

    if (!connId) {
      throw new Error('No active connection')
    }
    if (!dbName) {
      throw new Error('No database selected')
    }

    try {
      return await api.executeSQL(connId, dbName, sql)
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function createDatabase(dbName, connId = currentConnId.value) {
    if (!connId) {
      throw new Error('No active connection')
    }

    const result = await api.createDatabase(connId, dbName)
    await loadDatabases(connId)
    return result
  }

  async function dropDatabase(dbName, connId = currentConnId.value) {
    if (!connId) {
      throw new Error('No active connection')
    }

    const state = getConnectionState(connId)
    const connection = connections.value.find((item) => item.connId === connId)
    const dbType = String(state?.config?.dbType || connection?.config?.dbType || '').toLowerCase()
    const normalizedDbName = String(dbName || '').toLowerCase()

    if (dbType === 'mysql' && MYSQL_SYSTEM_DATABASES.has(normalizedDbName)) {
      throw new Error('MySQL system database cannot be deleted')
    }

    const result = await api.dropDatabase(connId, dbName)
    if (state?.selectedDatabase === dbName) {
      state.selectedDatabase = null
      state.tables = []
      state.selectedTable = null
      state.tableData = null
      state.tableSchema = null
    }
    await loadDatabases(connId)
    return result
  }

  function reset() {
    activeMux.value = {}
    connections.value = []
    currentConnId.value = null
    error.value = null
  }

  api.setConnectionRecoveryHandler(recoverConnection)

  return {
    activeMux,
    connections,
    currentConnId,
    activeConnection,
    connectionConfig,
    databases,
    selectedDatabase,
    tables,
    selectedTable,
    tableData,
    tableSchema,
    currentConnection,
    isLoading,
    error,
    isConnected,
    ensureConnState,
    getConnectionState,
    getRuntimeConnIdByConfigId,
    getConnectionStateByConfigId,
    isConnectionActive,
    addConnection,
    connect,
    recoverConnection,
    switchConnection,
    removeConnection,
    disconnect,
    loadDatabases,
    selectDatabase,
    loadTables,
    selectTable,
    loadTableSchema,
    getTableSchema,
    loadTableData,
    getTableData,
    getTableSchemaFor,
    executeSQL,
    createDatabase,
    dropDatabase,
    reset
  }
})
