import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN.js'
import enUS from './en-US.js'

// Read language from GMSSH SDK: gm.lang is 'zh-CN' or 'en'
function getGmLang() {
    try {
        const gm = window.$gm || (window.parent && window.parent.$gm) || (window.top && window.top.$gm)
        if (gm && gm.lang) {
            return gm.lang
        }
    } catch (e) {
        // ignore cross-origin errors
    }
    return 'zh-CN'
}

// Normalize gm.lang to our internal locale key
// gm.lang: 'zh-CN' -> 'zh-CN', 'en' -> 'en-US'
function normalizeLocale(gmLang) {
    if (!gmLang) return 'zh-CN'
    if (gmLang.startsWith('zh')) return 'zh-CN'
    return 'en-US'
}

const i18n = createI18n({
    legacy: false,
    locale: normalizeLocale(getGmLang()),
    fallbackLocale: 'zh-CN',
    messages: {
        'zh-CN': zhCN,
        'en-US': enUS,
    },
})

export default i18n

// Export helper so components can get gm.lang-normalized locale on demand
export { normalizeLocale, getGmLang }
