<template>
  <n-drawer v-model:show="visible" :width="420" placement="right">
    <n-drawer-content :title="$t('query.historyTitle')" closable>
      <div class="history-toolbar">
        <span class="history-count">{{ items.length }} {{ $t('query.historyCount') }}</span>
        <n-button text type="error" :disabled="items.length === 0" @click="clearHistory">
          {{ $t('query.clearHistory') }}
        </n-button>
      </div>

      <n-empty v-if="items.length === 0" :description="$t('query.historyEmpty')" class="history-empty" />

      <div v-else class="history-list">
        <button
          v-for="item in items"
          :key="item.id"
          class="history-item"
          @click="selectHistory(item.sql)"
        >
          <div class="history-meta">
            <span>{{ formatTime(item.timestamp) }}</span>
            <span v-if="item.database">{{ item.database }}</span>
          </div>
          <div class="history-summary">{{ item.summary || item.sql }}</div>
          <div class="history-sql">{{ item.sql }}</div>
        </button>
      </div>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup>
import { ref, watch } from 'vue'

const SQL_HISTORY_KEY = 'sql-history'

const emit = defineEmits(['select'])
const visible = defineModel({ default: false })
const items = ref([])

function loadHistory() {
  try {
    items.value = JSON.parse(localStorage.getItem(SQL_HISTORY_KEY) || '[]')
  } catch {
    items.value = []
  }
}

function selectHistory(sql) {
  emit('select', sql)
  visible.value = false
}

function clearHistory() {
  localStorage.removeItem(SQL_HISTORY_KEY)
  items.value = []
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  return new Date(timestamp).toLocaleString()
}

watch(visible, (show) => {
  if (show) {
    loadHistory()
  }
}, { immediate: true })
</script>

<style scoped>
.history-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.history-count {
  font-size: 12px;
  color: rgba(255,255,255,0.45);
}

.history-empty {
  margin-top: 48px;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.history-item {
  all: unset;
  display: block;
  cursor: pointer;
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 10px;
  padding: 12px;
  background: rgba(255,255,255,0.02);
  transition: border-color 0.18s ease, background 0.18s ease;
}

.history-item:hover {
  border-color: rgba(87,114,255,0.45);
  background: rgba(87,114,255,0.06);
}

.history-meta {
  display: flex;
  gap: 8px;
  font-size: 11px;
  color: rgba(255,255,255,0.4);
  margin-bottom: 6px;
}

.history-summary {
  font-size: 13px;
  font-weight: 600;
  color: rgba(255,255,255,0.82);
  margin-bottom: 4px;
  line-height: 1.5;
}

.history-sql {
  font-size: 12px;
  color: rgba(255,255,255,0.52);
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
