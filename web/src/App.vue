<template>
  <SideNavLayout v-if="showSideNav">
    <router-view />
  </SideNavLayout>
  <router-view v-else />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { SideNavLayout } from '@/components/common'
import { disableWorkspace, getWorkspaceDramaId, isWorkspaceActive } from '@/utils/workspace'

const route = useRoute()
const workspaceActive = ref(isWorkspaceActive())

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
