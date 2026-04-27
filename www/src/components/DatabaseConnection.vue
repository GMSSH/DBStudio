<template>
  <div class="database-connection glass">
    <h2 class="text-gradient">{{ t('connection.title') }}</h2>
    
    <form @submit.prevent="handleConnect" class="connection-form">
      <!-- Database Type -->
      <div class="form-group">
        <label>{{ t('connection.type') }}</label>
        <select v-model="config.dbType" @change="handleTypeChange">
          <option value="mysql">MySQL</option>
          <option value="postgres">PostgreSQL</option>
          <option value="sqlite">SQLite</option>
        </select>
      </div>

      <!-- MySQL/PostgreSQL Fields -->
      <template v-if="config.dbType !== 'sqlite'">
        <div class="form-row">
          <div class="form-group">
            <label>{{ t('connection.host') }}</label>
            <input v-model="config.host" type="text" placeholder="localhost" required />
          </div>
          <div class="form-group">
            <label>{{ t('connection.port') }}</label>
            <input v-model.number="config.port" type="number" placeholder="3306" required />
          </div>
        </div>

        <div class="form-group">
          <label>{{ t('connection.username') }}</label>
          <input v-model="config.username" type="text" placeholder="root" required />
        </div>

        <div class="form-group">
          <label>{{ t('connection.password') }}</label>
          <input v-model="config.password" type="password" autocomplete="off" />
        </div>

        <div class="form-group">
          <label>{{ t('connection.database') }}</label>
          <input v-model="config.database" type="text" placeholder="mydb" />
        </div>
      </template>

      <!-- SQLite Fields -->
      <template v-else>
        <div class="form-group">
          <label>{{ t('connection.filePath') }}</label>
          <div class="file-input-group">
            <input v-model="config.filePath" type="text" placeholder="/path/to/database.db" required />
            <button type="button" @click="selectFile" class="btn-secondary">
              {{ t('connection.selectFile') }}
            </button>
          </div>
        </div>
      </template>

      <!-- Actions -->
      <div class="form-actions">
        <button type="button" @click="handleTest" class="btn-secondary" :disabled="isLoading">
          {{ t('connection.testConnection') }}
        </button>
        <button type="submit" class="btn-primary gradient-primary" :disabled="isLoading">
          {{ isLoading ? t('common.loading') : t('connection.connect') }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDatabaseStore } from '@/stores/database'
import api from '@/utils/api'

const { t } = useI18n()
const store = useDatabaseStore()

const isLoading = ref(false)
const config = reactive({
  dbType: 'mysql',
  host: 'localhost',
  port: 3306,
  username: 'root',
  password: '',
  database: '',
  filePath: ''
})

const handleTypeChange = () => {
  if (config.dbType === 'postgres') {
    config.port = 5432
  } else if (config.dbType === 'mysql') {
    config.port = 3306
  }
}

// Safe access to $gm API (handles cross-origin iframe)
const getGmApi = () => {
  try {
    // Try direct access first
    if (window.$gm) return window.$gm
    // Try parent window (if in iframe)
    if (window.parent && window.parent.$gm) return window.parent.$gm
    // Try top window
    if (window.top && window.top.$gm) return window.top.$gm
  } catch (e) {
    // Silently handle error
  }
  // Fallback to console
  return {
    message: {
      success: () => {},
      error: () => {},
      warning: () => {}
    },
    chooseFile: (callback) => {
      const path = prompt('Enter file path:')
      if (path) callback(path)
    }
  }
}

const selectFile = () => {
  const gm = getGmApi()
  if (gm.chooseFile) {
    gm.chooseFile((filePath) => {
      config.filePath = filePath
    }, '/')
  }
}

const handleTest = async () => {
  isLoading.value = true
  const gm = getGmApi()
  try {
    await api.testConnection(config)
    gm.message.success(t('connection.testConnection') + ' ' + t('common.success'))
  } catch (error) {
    gm.message.error(error.message)
  } finally {
    isLoading.value = false
  }
}

const handleConnect = async () => {
  isLoading.value = true
  const gm = getGmApi()
  try {
    await store.connect(config)
    gm.message.success(t('connection.connect') + ' ' + t('common.success'))
  } catch (error) {
    gm.message.error(error.message)
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped lang="scss">
@use '@/styles/variables.scss' as *;
@use '@/styles/mixins.scss' as *;

.database-connection {
  padding: var(--spacing-xl);
  border-radius: var(--radius-lg);
  max-width: 500px;
  margin: 0 auto;

  h2 {
    margin-bottom: var(--spacing-lg);
    font-size: var(--font-size-2xl);
  }
}

.connection-form {
  .form-group {
    margin-bottom: var(--spacing-md);

    label {
      display: block;
      margin-bottom: var(--spacing-sm);
      color: var(--text-secondary);
      font-size: var(--font-size-sm);
      font-weight: 500;
    }

    input, select {
      width: 100%;
      padding: var(--spacing-sm) var(--spacing-md);
      background: var(--bg-secondary);
      border: 1px solid var(--glass-border);
      border-radius: var(--radius-sm);
      color: var(--text-primary);
      @include smooth-transition(border-color, background);

      &:focus {
        border-color: var(--primary-start);
        background: var(--bg-tertiary);
      }
    }

    select {
      cursor: pointer;
    }
  }

  .form-row {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: var(--spacing-md);
  }

  .file-input-group {
    display: flex;
    gap: var(--spacing-sm);

    input {
      flex: 1;
    }
  }

  .form-actions {
    display: flex;
    gap: var(--spacing-md);
    margin-top: var(--spacing-lg);

    button {
      flex: 1;
      padding: var(--spacing-sm) var(--spacing-lg);
      border-radius: var(--radius-sm);
      font-weight: 500;
      @include hover-lift;

      &:disabled {
        opacity: 0.5;
        cursor: not-allowed;
        transform: none !important;
      }
    }

    .btn-primary {
      color: white;
      box-shadow: var(--shadow-glow);
    }

    .btn-secondary {
      background: var(--bg-secondary);
      color: var(--text-primary);
      border: 1px solid var(--glass-border);
    }
  }
}
</style>
