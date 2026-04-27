// Design system tokens (define first, before everything else)
import './styles/tokens.css'
// gm-app-components library CSS (provides --jm-* compat tokens)
import 'gm-app-components/index.css'
import './styles/global.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import naive from 'naive-ui'
import i18n from './i18n/index.js'
import App from './App.vue'

// Create app
const app = createApp(App)

// Use plugins
app.use(createPinia())
app.use(naive)
app.use(i18n)

// Mount app
app.mount('#app')
