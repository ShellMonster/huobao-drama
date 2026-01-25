<template>
  <div class="script-edit-container">
    <el-page-header @back="goBack" :title="$t('script.back')">
      <template #content>
        <h2>{{ $t('script.title') }}</h2>
      </template>
    </el-page-header>
    <p>{{ $t('script.inDevelopment') }}</p>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { t } = useI18n()

const goBack = () => {
  router.back()
}
</script>

<style scoped>
.script-edit-container {
  padding: 20px;
}
</style>
