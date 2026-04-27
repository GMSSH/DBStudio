/**
 * GMSSH SDK Wrapper for UI Utilities
 */

class GMSSHUtils {
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
     * Show success message
     */
    success(message, options) {
        const gm = this.getGmApi()
        if (gm && gm.message) {
            gm.message.success(message, options)
        } else {
            // Production: silent success
        }
    }

    /**
     * Show error message
     */
    error(message, options) {
        const gm = this.getGmApi()
        if (gm && gm.message) {
            gm.message.error(message, options)
        } else {
            // Production: silent error
        }
    }

    /**
     * Show info message
     */
    info(message, options) {
        const gm = this.getGmApi()
        if (gm && gm.message) {
            gm.message.info(message, options)
        } else {
            // Production: silent info
        }
    }

    /**
     * Show warning message
     */
    warning(message, options) {
        const gm = this.getGmApi()
        if (gm && gm.message) {
            gm.message.warning(message, options)
        } else {
            // Production: silent warning
        }
    }

    /**
     * Show confirm dialog
     * SDK uses dialog.warning() with onPositiveClick/onNegativeClick callbacks
     */
    async confirm(options) {
        const gm = this.getGmApi()

        // Use GMSSH dialog.warning for confirm dialogs
        if (gm && gm.dialog && gm.dialog.warning) {
            return new Promise((resolve) => {
                gm.dialog.warning({
                    title: options.title || '确认',
                    content: options.content || '',
                    positiveText: options.positiveText || '确定',
                    negativeText: options.negativeText || '取消',
                    maskClosable: false,
                    onPositiveClick: () => {
                        resolve(true)
                    },
                    onNegativeClick: () => {
                        resolve(false)
                    },
                    onClose: () => {
                        resolve(false)
                    }
                })
            })
        }

        // Fallback to native confirm
        return window.confirm(options.content || options.title)
    }

    /**
     * Show alert dialog
     * SDK uses dialog.info() for alerts
     */
    async alert(options) {
        const gm = this.getGmApi()

        // Use GMSSH dialog.info for alerts
        if (gm && gm.dialog && gm.dialog.info) {
            return new Promise((resolve) => {
                gm.dialog.info({
                    title: options.title || '提示',
                    content: options.content || '',
                    positiveText: options.positiveText || '确定',
                    onPositiveClick: () => {
                        resolve(true)
                    },
                    onClose: () => {
                        resolve(true)
                    }
                })
            })
        }

        // Fallback to native alert
        window.alert(options.content || options.title)
        return true
    }

    /**
     * Copy text to clipboard through GMSSH shell APIs when available.
     */
    async copyToClipboard(text) {
        const value = String(text || '')
        if (!value) return false

        const gm = this.getGmApi()

        if (gm && typeof gm.copyToClipboard === 'function') {
            return !!gm.copyToClipboard(value)
        }

        if (navigator.clipboard?.writeText) {
            await navigator.clipboard.writeText(value)
            return true
        }

        return false
    }

    async chooseFile(initialPath = '/home/') {
        const gm = this.getGmApi()
        if (!gm || typeof gm.chooseFile !== 'function') {
            return ''
        }

        return new Promise((resolve) => {
            gm.chooseFile((selectedPath) => {
                resolve(selectedPath || '')
            }, initialPath)
        })
    }

    async chooseFolder(initialPath = '/home/') {
        const gm = this.getGmApi()
        if (!gm || typeof gm.chooseFolder !== 'function') {
            return ''
        }

        return new Promise((resolve) => {
            gm.chooseFolder((selectedPath) => {
                resolve(selectedPath || '')
            }, initialPath)
        })
    }

    async readFileText(path) {
        const gm = this.getGmApi()
        if (!gm || typeof gm.request !== 'function' || !path) {
            return ''
        }

        const response = await gm.request({
            url: '/api/files/read_file_body',
            method: 'post',
            data: {
                path,
                auto_create: '0'
            }
        })

        if (!response || response.code !== 200000) {
            throw new Error(response?.msg || 'Read file failed')
        }

        return response?.data?.data || ''
    }

    openFolder(path) {
        const gm = this.getGmApi()
        if (!gm || typeof gm.openFolder !== 'function' || !path) {
            return false
        }

        gm.openFolder(path)
        return true
    }
}

export default new GMSSHUtils()
