<template>
  <div class="storyboard-generation">
    <el-empty :description="$t('storyboard.inDevelopment')" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  dramaId: string
  episodeId: string
}>()

defineEmits<{
  storyboardGenerated: []
}>()

const { t: $t } = useI18n()
</script>

<style scoped lang="scss">
.storyboard-generation {
  padding: 24px;
}
</style>
