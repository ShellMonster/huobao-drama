<template>
  <el-config-provider :locale="elementLocale">
    <SideNavLayout v-if="showSideNav">
      <router-view />
    </SideNavLayout>
    <router-view v-else />
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { SideNavLayout } from '@/components/common'
import { disableWorkspace, getWorkspaceDramaId, isWorkspaceActive } from '@/utils/workspace'
import { getElementLocale } from '@/locales/element'

const route = useRoute()
const workspaceActive = ref(isWorkspaceActive())
const { locale } = useI18n()
const elementLocale = computed(() => getElementLocale(locale.value))

watch(
  () => route.fullPath,
  () => {
    const hasWorkspacePage =
      route.name === 'DramaWorkspace' ||
      (typeof route.path === 'string' && route.path.startsWith('/dramas/'))
    const savedDramaId = getWorkspaceDramaId()
    const routeDramaId = typeof route.params?.id === 'string' ? route.params.id : undefined

    if (!hasWorkspacePage) {
      disableWorkspace()
    } else if (savedDramaId && routeDramaId && savedDramaId !== routeDramaId) {
      disableWorkspace()
    }
    workspaceActive.value = isWorkspaceActive()
  }
)

const showSideNav = computed(() => route.meta?.sideNav === true && !workspaceActive.value)
</script>

<style>
#app {
  width: 100%;
  min-height: 100vh;
}
</style>
