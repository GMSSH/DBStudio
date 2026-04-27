<template>
  <n-config-provider
    :theme="darkTheme"
    :theme-overrides="themeOverrides"
    :locale="naiveLocale"
    :date-locale="naiveDateLocale"
  >
    <n-message-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <div id="app-window">
            <MainLayout />
          </div>
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
import { computed } from 'vue'
import { darkTheme, zhCN, enUS, dateZhCN, dateEnUS } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import MainLayout from '@/views/MainLayout.vue'
import { themeOverrides } from '@/styles/naive-theme.js'

const { locale } = useI18n()
const naiveLocale = computed(() => locale.value === 'zh-CN' ? zhCN : enUS)
const naiveDateLocale = computed(() => locale.value === 'zh-CN' ? dateZhCN : dateEnUS)
</script>

<style>
.form-modal {
  background: transparent;
}

.form-modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--ref-color-white-8);
  font-size: 16px;
  font-weight: 500;
}

.form-modal-body {
  padding: 20px;
}
</style>
