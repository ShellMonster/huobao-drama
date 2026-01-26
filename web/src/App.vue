<template>
  <el-config-provider :locale="elementLocale">
    <SideNavLayout v-if="showSideNav">
      <router-view />
    </SideNavLayout>
    <router-view v-else />
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { SideNavLayout } from '@/components/common'
import { getElementLocale } from '@/locales/element'

const route = useRoute()
const { locale } = useI18n()
const elementLocale = computed(() => getElementLocale(locale.value))

const showSideNav = computed(() => route.meta?.sideNav === true)
</script>

<style>
#app {
  width: 100%;
  min-height: 100vh;
}
</style>
