<template>
  <n-modal
    v-model:show="showProxy"
    preset="card"
    class="form-modal"
    header-class="form-modal-head"
    content-class="form-modal-body"
    footer-class="form-modal-footer"
    :title="title"
    :style="modalStyle"
    v-bind="attrs"
  >
    <slot />
    <template #footer>
      <slot name="footer" />
    </template>
  </n-modal>
</template>

<script setup>
import { computed, useAttrs } from 'vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '' },
  width: { type: [Number, String], default: 420 }
})

const emit = defineEmits(['update:modelValue'])
const attrs = useAttrs()

const showProxy = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const modalStyle = computed(() => {
  if (typeof props.width === 'number') {
    return { width: `${props.width}px` }
  }

  return props.width ? { width: props.width } : undefined
})
</script>
