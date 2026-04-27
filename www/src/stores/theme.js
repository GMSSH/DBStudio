// Theme store stub — dark mode only, no toggle
// Kept for backward compatibility; isDark is always true.
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export const useThemeStore = defineStore('theme', () => {
    const theme = ref('dark')
    // Always dark; toggle removed by design
    const isDark = computed(() => true)

    return { theme, isDark }
})
