<template>
  <div class="timeline-editor-page">
    <div class="editor-header">
      <el-button link @click="goBack" class="back-button">
        <el-icon><ArrowLeft /></el-icon>
        {{ $t('timeline.backToEditor') }}
      </el-button>
      <h2>{{ $t('timeline.title') }}</h2>
    </div>
    
    <LoadingSection
      class="editor-content"
      :loading="pageLoading"
      :text="$t('common.loading')"
    >
      <VideoTimelineEditor 
        v-if="scenes.length > 0"
        :scenes="scenes" 
        :episode-id="episodeId" 
      />
      <el-empty v-else :description="$t('timeline.noScenes')" />
    </LoadingSection>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import VideoTimelineEditor from '@/components/editor/VideoTimelineEditor.vue'
import { LoadingSection } from '@/components/common'
import { getCache, setCache } from '@/utils/cache'

const route = useRoute()
const router = useRouter()

const episodeId = route.params.id as string
const scenes = ref<any[]>([])
const pageLoading = ref(false)
const cacheTTL = 30 * 1000
let loadingTimer: number | null = null

const getScenesCacheKey = () => `storyboards:episode:${episodeId}`

const startPageLoading = () => {
  if (loadingTimer) return
  loadingTimer = window.setTimeout(() => {
    pageLoading.value = true
  }, 200)
}

const stopPageLoading = () => {
  if (loadingTimer) {
    window.clearTimeout(loadingTimer)
    loadingTimer = null
  }
  pageLoading.value = false
}

const loadScenes = async () => {
  const cacheKey = getScenesCacheKey()
  const cached = getCache<{ storyboards: any[] }>(cacheKey, cacheTTL)
  if (cached) {
    scenes.value = cached.storyboards || []
  } else {
    startPageLoading()
  }
  try {
    const res = await dramaAPI.getStoryboards(episodeId)
    scenes.value = res.storyboards || []
    setCache(cacheKey, { storyboards: scenes.value })
  } catch (error: any) {
    ElMessage.error($t('timeline.loadFailed'))
  } finally {
    if (!cached) {
      stopPageLoading()
    }
  }
}

const goBack = () => {
  router.back()
}

onMounted(() => {
  loadScenes()
})

onBeforeUnmount(() => {
  if (loadingTimer) {
    window.clearTimeout(loadingTimer)
    loadingTimer = null
  }
})
</script>

<style scoped lang="scss">
.timeline-editor-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;

  .editor-header {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px 24px;
    background: white;
    border-bottom: 1px solid #e4e7ed;

    .back-button {
      display: flex;
      align-items: center;
      gap: 4px;
      color: #606266;
      font-size: 14px;
      
      &:hover {
        color: #409eff;
      }
    }

    h2 {
      margin: 0;
      font-size: 18px;
      font-weight: 500;
    }
  }

  .editor-content {
    flex: 1;
    overflow: hidden;
  }
}
</style>
