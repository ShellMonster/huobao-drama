<template>
  <div
    v-loading="loading"
    :element-loading-text="text"
    :element-loading-custom-class="loadingClass"
    :element-loading-background="background"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  loading: boolean
  text?: string
  loadingClass?: string
  background?: string
}>(), {
  text: '',
  loadingClass: 'storyboard-glass-loading',
  background: 'transparent'
})
</script>
