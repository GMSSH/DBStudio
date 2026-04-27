/**
 * Connection Configuration API Client
 * Communicates with backend to manage saved connections
 */

class ConnectionConfigAPI {
    constructor() {
        this.baseURL = '/api/call/xiaojun/dbmanager'
        this.initialized = false
        this.initPromise = null
    }

    /**
     * Get $gm API safely
     */
    getGmApi() {
        try {
            if (window.$gm) return window.$gm
            if (window.parent && window.parent.$gm) return window.parent.$gm
            if (window.top && window.top.$gm) return window.top.$gm
        } catch (e) {
            // Silently fail
        }
        return null
    }

    /**
     * Ensure SDK is initialized before any request
     */
    async ensureInitialized() {
        if (this.initialized) {
            return
        }

        // If initialization is in progress, wait for it
        if (this.initPromise) {
            return this.initPromise
        }

        // Start initialization
        this.initPromise = (async () => {
            try {
                const gm = this.getGmApi()
                if (gm) {
                    // First: check backend service status
                    if (typeof gm.checkStatus === 'function') {
                        await gm.checkStatus()
                    }
                    // Then: initialize SDK (will auto-start backend if needed)
                    if (typeof gm.init === 'function') {
                        await gm.init()
                    }
                }
                this.initialized = true
            } catch (error) {
                // Silently handle initialization errors
                this.initialized = true // Mark as initialized to avoid blocking
            }
        })()

        return this.initPromise
    }

    /**
     * Make a request to backend
     */
    async request(method, params = {}) {
        // Ensure SDK is initialized before making request
        await this.ensureInitialized()

        const gm = this.getGmApi()
        if (!gm) {
            throw new Error('GMSSH SDK not loaded')
        }

        const response = await gm.request({
            url: `${this.baseURL}/${method}`,
            method: 'POST',
            data: { params }
        })

        // Check gateway response
        if (!response || response.code !== 200000) {
            throw new Error(response?.msg || 'Gateway request failed')
        }

        // Check RPC response
        const rpcResponse = response.data
        if (!rpcResponse || rpcResponse.code !== 200) {
            throw new Error(rpcResponse?.msg || 'RPC request failed')
        }

        return rpcResponse.data
    }

    /**
     * Save connection configuration
     */
    async saveConnection(config) {
        return this.request('db.saveConnectionConfig', config)
    }

    /**
     * List all saved connections (without passwords)
     */
    async listConnections() {
        return this.request('db.listConnectionConfigs')
    }

    /**
     * Get single connection (with password)
     */
    async getConnection(id) {
        return this.request('db.getConnectionConfig', { id })
    }

    /**
     * Delete connection
     */
    async deleteConnection(id) {
        return this.request('db.deleteConnectionConfig', { id })
    }
}

export default new ConnectionConfigAPI()
