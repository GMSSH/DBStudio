<template>
  <n-form ref="formRef" :model="formModel" :rules="rules" label-placement="top">
    <!-- Security Notice -->
    <n-alert 
      v-if="showSecurityNotice"
      type="info" 
      :bordered="false" 
      closable
      @close="handleCloseNotice"
      style="margin-bottom: 16px;"
    >
      <template #icon>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
          <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
        </svg>
      </template>
      {{ t('connection.securityNotice') }}
    </n-alert>

    <n-form-item :label="t('connection.name')" path="name">
      <n-input v-model:value="formModel.name" :placeholder="t('connection.namePlaceholder')" />
    </n-form-item>

    <n-form-item :label="t('connection.dbType')" path="dbType">
      <n-select v-model:value="formModel.dbType" :options="dbTypeOptions" @update:value="handleTypeChange" />
    </n-form-item>

    <template v-if="formModel.dbType !== 'sqlite'">
      <n-grid :cols="2" :x-gap="12">
        <n-grid-item>
          <n-form-item :label="t('connection.host')" path="host">
            <n-input v-model:value="formModel.host" placeholder="localhost" />
          </n-form-item>
        </n-grid-item>
        <n-grid-item>
          <n-form-item :label="t('connection.port')" path="port">
            <n-input-number v-model:value="formModel.port" placeholder="3306" style="width: 100%;" />
          </n-form-item>
        </n-grid-item>
      </n-grid>

      <n-form-item :label="t('connection.username')" path="username">
        <n-input v-model:value="formModel.username" :placeholder="t('connection.usernamePlaceholder')" />
      </n-form-item>

      <n-form-item :label="t('connection.password')" path="password">
        <n-input v-model:value="formModel.password" type="password" show-password-on="click" />
      </n-form-item>

      <n-form-item :label="t('connection.database')">
        <n-input v-model:value="formModel.database" :placeholder="t('connection.databasePlaceholder')" />
      </n-form-item>
    </template>

    <template v-else>
      <n-form-item :label="t('connection.filePath')" path="filePath">
        <n-input 
          v-model:value="formModel.filePath" 
          placeholder="/path/to/database.db"
        >
          <template #suffix>
            <n-button text @click="handleChooseFile">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/>
                <polyline points="13 2 13 9 20 9"/>
              </svg>
            </n-button>
          </template>
        </n-input>
      </n-form-item>
    </template>

    <n-space v-if="!hideActions">
      <n-button type="primary" @click="handleConnect" :loading="connecting">{{ t('connection.btnConnect') }}</n-button>
      <n-button @click="handleTest" :loading="testing">{{ t('connection.btnTest') }}</n-button>
      <n-button @click="handleSave" :loading="saving">{{ t('connection.btnSave') }}</n-button>
    </n-space>
  </n-form>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import connectionApi from '@/utils/connectionApi'
import api from '@/utils/api'
import { useDatabaseStore } from '@/stores/database'
import gmssh from '@/utils/gmssh'

const { t } = useI18n()

const props = defineProps({
  initialData: {
    type: Object,
    default: () => ({})
  },
  hideActions: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['success', 'cancel'])

const store = useDatabaseStore()

const formRef = ref(null)
const testing = ref(false)
const saving = ref(false)
const connecting = ref(false)

// Security notice state - check localStorage
const SECURITY_NOTICE_KEY = 'sqlmanager_hide_security_notice'
const showSecurityNotice = ref(localStorage.getItem(SECURITY_NOTICE_KEY) !== 'true')

const handleCloseNotice = () => {
  showSecurityNotice.value = false
  localStorage.setItem(SECURITY_NOTICE_KEY, 'true')
}

const dbTypeOptions = [
  { label: 'MySQL', value: 'mysql' },
  { label: 'PostgreSQL', value: 'postgres' },
  { label: 'SQLite', value: 'sqlite' }
]

const formModel = ref({
  id: '',
  name: '',
  dbType: 'mysql',
  host: 'localhost',
  port: 3306,
  username: 'root',
  password: '',
  database: '',
  filePath: '',
  ...props.initialData
})

const rules = ref({
  name: { required: true, message: () => t('connection.nameRequired'), trigger: 'blur' },
  dbType: { required: true, message: () => t('connection.dbTypeRequired') },
  host: { 
    required: () => formModel.value.dbType !== 'sqlite', 
    message: () => t('connection.hostRequired'), 
    trigger: 'blur' 
  },
  port: { 
    required: () => formModel.value.dbType !== 'sqlite', 
    type: 'number', 
    message: () => t('connection.portRequired')
  },
  username: { 
    required: () => formModel.value.dbType !== 'sqlite', 
    message: () => t('connection.usernameRequired'), 
    trigger: 'blur' 
  },
  filePath: { 
    required: () => formModel.value.dbType === 'sqlite', 
    message: () => t('connection.filePathRequired'), 
    trigger: ['blur', 'change'] 
  }
})

const handleTypeChange = () => {
  if (formModel.value.dbType === 'postgres') {
    formModel.value.port = 5432
    formModel.value.database = 'postgres'  // Default database
  } else if (formModel.value.dbType === 'mysql') {
    formModel.value.port = 3306
    formModel.value.database = ''  // Clear for MySQL
  } else if (formModel.value.dbType === 'sqlite') {
    formModel.value.database = ''  // SQLite doesn't use database field
  }
}

const handleTest = async () => {
  testing.value = true
  try {
    await api.testConnection(formModel.value)
    gmssh.success(t('connection.testSuccess'))
  } catch (error) {
    gmssh.error(t('connection.testFailed', { msg: error.message }))
  } finally {
    testing.value = false
  }
}

const handleSave = async () => {
  try {
    await formRef.value?.validate()
    saving.value = true
    await connectionApi.saveConnection(formModel.value)
    gmssh.success(t('connection.saveSuccess'))
    emit('success', { action: 'save', data: formModel.value })
  } catch (error) {
    if (error.errors) return
    gmssh.error(t('connection.saveFailed', { msg: error.message }))
  } finally {
    saving.value = false
  }
}

// Handle file picker for SQLite
const handleChooseFile = () => {
  const gm = gmssh.getGmApi()
  if (gm && gm.chooseFile) {
    gm.chooseFile((selectedPath) => {
      if (selectedPath) {
        formModel.value.filePath = selectedPath
      }
    }, '/home/')  // Default to user's home directory
  } else {
    gmssh.warning(t('connection.filePickerNotAvailable'))
  }
}

const handleConnect = async () => {
  try {
    await formRef.value?.validate()
    connecting.value = true
    
    // Auto-save if new connection
    if (!formModel.value.id) {
      try {
        const result = await connectionApi.saveConnection(formModel.value)
        formModel.value.id = result.id
      } catch (saveError) {
        // Silently handle save error
      }
    }
    
    await store.connect(formModel.value)
    gmssh.success(t('connection.connectSuccess'))
    emit('success', { action: 'connect', data: formModel.value })
  } catch (error) {
    if (error.errors) return
    gmssh.error(t('connection.connectFailed', { msg: error.message }))
  } finally {
    connecting.value = false
  }
}

// Watch for external data updates
watch(() => props.initialData, (newData) => {
  formModel.value = { ...formModel.value, ...newData }
}, { deep: true })

defineExpose({
  formModel,
  validate: () => formRef.value?.validate(),
  handleConnect,
  handleTest,
  handleSave
})
</script>
