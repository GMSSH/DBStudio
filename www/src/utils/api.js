/**
 * API Client for Database Manager
 * Uses window.$gm.request() to communicate with GMSSH backend
 */

class DatabaseAPI {
    constructor() {
        this.baseURL = '/api/call/xiaojun/dbmanager'
        this.connectionRecoveryHandler = null
        this.connectionAliases = new Map()
        this.pendingRecoveries = new Map()
        // Track failed recoveries to prevent infinite retry loops
        this.failedRecoveries = new Map() // connId -> timestamp of last failure
        this.recoveryAttempts = new Map()  // connId -> attempt count
    }

    setConnectionRecoveryHandler(handler) {
        this.connectionRecoveryHandler = typeof handler === 'function' ? handler : null
    }

    isReconnectableError(error) {
        const message = String(error?.message || '').toLowerCase()
        // Only match specific RPC-level "connection not found" errors,
        // NOT gateway-level failures (e.g. backend completely down)
        return message.includes('connection not found') && !message.includes('gateway')
    }

    shouldTryRecover(method, params, error, allowRecover = true) {
        if (!allowRecover || !this.connectionRecoveryHandler) {
            return false
        }

        if (!params?.connId) {
            return false
        }

        if (method === 'db.connect' || method === 'db.disconnect') {
            return false
        }

        if (!this.isReconnectableError(error)) {
            return false
        }

        // Check cooldown: don't retry within 30s of a failed recovery
        const lastFailure = this.failedRecoveries.get(params.connId)
        if (lastFailure && Date.now() - lastFailure < 30000) {
            return false
        }

        // Max 2 recovery attempts per connId
        const attempts = this.recoveryAttempts.get(params.connId) || 0
        if (attempts >= 2) {
            return false
        }

        return true
    }

    resolveConnId(connId) {
        let currentConnId = connId
        const visited = new Set()

        while (this.connectionAliases.has(currentConnId) && !visited.has(currentConnId)) {
            visited.add(currentConnId)
            currentConnId = this.connectionAliases.get(currentConnId)
        }

        return currentConnId
    }

    async recoverConnection(connId) {
        if (!this.pendingRecoveries.has(connId)) {
            const attempts = (this.recoveryAttempts.get(connId) || 0) + 1
            this.recoveryAttempts.set(connId, attempts)

            const recoveryTask = Promise.resolve()
                .then(() => this.connectionRecoveryHandler(connId))
                .then((result) => {
                    // Recovery succeeded — clear failure state
                    this.failedRecoveries.delete(connId)
                    this.recoveryAttempts.delete(connId)
                    return result?.connId || result || null
                })
                .catch((err) => {
                    // Recovery failed — record cooldown
                    this.failedRecoveries.set(connId, Date.now())
                    return null
                })
                .finally(() => {
                    this.pendingRecoveries.delete(connId)
                })

            this.pendingRecoveries.set(connId, recoveryTask)
        }

        return this.pendingRecoveries.get(connId)
    }

    /**
     * Make a request to the backend
     * Response structure:
     * {
     *   code: 200000,           // GMSSH gateway code
     *   data: {
     *     code: 200/400,        // RPC business code
     *     data: {...},          // Actual data
     *     msg: "...",           // RPC error message
     *     meta: {...}
     *   },
     *   msg: "...",             // Gateway message
     *   meta: {...}
     * }
     */
    async request(method, params = {}, options = {}) {
        const { allowRecover = true } = options
        const normalizedParams = params?.connId
            ? {
                ...params,
                connId: this.resolveConnId(params.connId)
            }
            : params

        try {
            // Note: window.$gm is provided by gmAppSdk.js
            if (!window.$gm) {
                throw new Error('GMSSH SDK not loaded')
            }

            const response = await window.$gm.request({
                url: `${this.baseURL}/${method}`,
                method: 'POST',
                data: { params: normalizedParams }
            })

            // Check gateway-level response
            if (!response || response.code !== 200000) {
                throw new Error(response?.msg || 'Gateway request failed')
            }

            // Extract RPC response
            const rpcResponse = response.data

            // Check RPC-level response
            if (!rpcResponse || rpcResponse.code !== 200) {
                throw new Error(rpcResponse?.msg || 'RPC request failed')
            }

            // Return actual data
            return rpcResponse.data
        } catch (error) {
            if (this.shouldTryRecover(method, normalizedParams, error, allowRecover)) {
                const recoveredConnId = await this.recoverConnection(normalizedParams.connId)

                if (recoveredConnId) {
                    this.connectionAliases.set(normalizedParams.connId, recoveredConnId)

                    return this.request(
                        method,
                        {
                            ...normalizedParams,
                            connId: recoveredConnId
                        },
                        { allowRecover: false }
                    )
                }
            }

            // Silent error in production
            throw error
        }
    }

    // Connection Management
    async testConnection(config) {
        return this.request('db.testConnection', config)
    }

    async connect(config) {
        return this.request('db.connect', config)
    }

    async disconnect(connId) {
        return this.request('db.disconnect', { connId })
    }

    async listConnections() {
        return this.request('db.listConnections')
    }

    // Database Operations
    async listDatabases(connId) {
        return this.request('db.listDatabases', { connId })
    }

    async getDatabaseInfo(connId, database) {
        return this.request('db.getDatabaseInfo', { connId, database })
    }

    async createDatabase(connId, database) {
        return this.request('db.createDatabase', { connId, database })
    }

    async dropDatabase(connId, database) {
        return this.request('db.dropDatabase', { connId, database })
    }

    async switchDatabase(connId, database) {
        return this.request('db.switchDatabase', { connId, database })
    }

    // Schema Operations (PostgreSQL only)
    async listSchemas(connId, database) {
        return this.request('db.listSchemas', { connId, database })
    }

    // Table Operations
    async listTables(connId, database) {
        return this.request('db.listTables', { connId, database })
    }

    async getTableSchema(connId, database, table) {
        return this.request('db.getTableSchema', { connId, database, table })
    }

    async getTableData(connId, database, table, page = 1, pageSize = 100, sortCol = '', sortDir = '') {
        return this.request('db.getTableData', {
            connId,
            database,
            table,
            page,
            pageSize,
            sortCol,
            sortDir
        })
    }

    // Query Operations
    async executeSQL(connId, database, sql) {
        return this.request('db.executeSQL', { connId, database, sql })
    }

    // Data Modification
    async updateRow(connId, database, table, pkValues, updates) {
        return this.request('db.updateRow', { connId, database, table, pkValues, updates })
    }

    async insertRow(connId, database, table, values) {
        return this.request('db.insertRow', { connId, database, table, values })
    }

    async deleteRow(connId, database, table, pkValues) {
        return this.request('db.deleteRow', { connId, database, table, pkValues })
    }

    async batchModify(connId, database, table, ops) {
        return this.request('db.batchModify', { connId, database, table, ops })
    }

    // Connection Configuration
    async saveConnectionConfig(config) {
        return this.request('db.saveConnectionConfig', config)
    }

    async listConnectionConfigs() {
        return this.request('db.listConnectionConfigs')
    }

    async getConnectionConfig(id) {
        return this.request('db.getConnectionConfig', { id })
    }

    async deleteConnectionConfig(id) {
        return this.request('db.deleteConnectionConfig', { id })
    }

    // Table Designer (DDL)
    async createTable(connId, database, tableDef) {
        return this.request('db.createTable', { connId, database, tableDef })
    }

    async alterTable(connId, database, table, ops) {
        return this.request('db.alterTable', { connId, database, table, ops })
    }

    async dropTable(connId, database, table) {
        return this.request('db.dropTable', { connId, database, table })
    }

    async renameTable(connId, database, oldName, newName) {
        return this.request('db.renameTable', { connId, database, oldName, newName })
    }

    async getTableDDL(connId, database, table) {
        const result = await this.request('db.getTableDDL', { connId, database, table })
        return typeof result === 'string' ? result : result?.ddl
    }

    // Import / Export
    async exportTable(connId, database, table, format = 'csv', limit = 0) {
        return this.request('db.exportTable', { connId, database, table, format, limit })
    }

    async exportTableDDL(connId, database, table) {
        return this.request('db.exportTableDDL', { connId, database, table })
    }

    async importCSV(connId, database, table, csvData, mapping = {}, headerRow = true) {
        return this.request('db.importCSV', { connId, database, table, csvData, mapping, headerRow })
    }

    async importTableSQL(connId, database, table, sqlData) {
        return this.request('db.importTableSQL', { connId, database, table, sqlData })
    }

    async exportDatabase(connId, database, target = 'app', targetPath = '', options = {}) {
        return this.request('db.exportDatabase', { connId, database, target, targetPath, options })
    }

    async importDatabase(connId, database, filePath, options = {}) {
        return this.request('db.importDatabase', {
            connId,
            database,
            filePath,
            options: {
                createDatabase: !!options.createDatabase,
                strategy: options.strategy || 'source'
            }
        })
    }

    async getDatabaseDumpTask(taskId) {
        return this.request('db.getDatabaseDumpTask', { taskId })
    }

    async listDatabaseDumpTasks() {
        return this.request('db.listDatabaseDumpTasks')
    }

    async getDatabaseDumpTaskLogs(taskId) {
        return this.request('db.getDatabaseDumpTaskLogs', { taskId })
    }

    async cancelDatabaseDumpTask(taskId) {
        return this.request('db.cancelDatabaseDumpTask', { taskId })
    }
}

export default new DatabaseAPI()
