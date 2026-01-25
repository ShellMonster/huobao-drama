<template>
  <div class="character-images-container">
    <el-page-header @back="goBack" :title="$t('workflow.backToProject')">
      <template #content>
        <h2>{{ $t('workflow.characterImagesTitle') }}</h2>
      </template>
      <template #extra>
        <el-button type="primary" @click="batchGenerate" :loading="batchGenerating" :disabled="selectedCharacters.length === 0">
          <el-icon><Picture /></el-icon>
          {{ $t('workflow.batchGenerate') }} ({{ selectedCharacters.length }})
        </el-button>
        <el-button @click="goToCharacterManagement">
          <el-icon><Edit /></el-icon>
          {{ $t('workflow.manageCharacters') }}
        </el-button>
      </template>
    </el-page-header>

    <LoadingSection class="main-card" :loading="pageLoading" :text="$t('common.loading')">
      <el-card shadow="never">
        <div class="toolbar">
        <el-checkbox v-model="selectAll" @change="handleSelectAll" :indeterminate="isIndeterminate">
          {{ $t('common.selectAll') }}
        </el-checkbox>
        <span class="selection-info">{{ $t('workflow.selectedCharacters', { selected: selectedCharacters.length, total: characters.length }) }}</span>
      </div>

      <div class="character-list">
        <el-row :gutter="20">
          <el-col :span="6" v-for="character in characters" :key="character.id">
            <el-card shadow="hover" class="character-card" :class="{ 'has-image': character.image_url, 'selected': isSelected(character.id) }">
              <el-checkbox 
                class="card-checkbox" 
                :model-value="isSelected(character.id)" 
                @change="toggleSelection(character.id)"
              />
              <div class="character-preview">
                <img v-if="character.image_url" :src="character.image_url" :alt="character.name" />
                <el-avatar v-else :size="120">{{ character.name[0] }}</el-avatar>
              </div>
              
              <div class="character-info">
                <h4>{{ character.name }}</h4>
                <p class="role">{{ character.role }}</p>
                <p class="desc">{{ character.appearance }}</p>
              </div>

              <el-button 
                type="primary" 
                @click="generateImage(character)" 
                :loading="generatingIds.includes(character.id)"
                :disabled="batchGenerating || (generatingIds.length > 0 && !generatingIds.includes(character.id))"
                style="width: 100%"
              >
                <span v-if="generatingIds.includes(character.id)">{{ $t('common.generating') }}</span>
                <span v-else>{{ character.image_url ? $t('common.regenerate') : $t('character.generateImage') }}</span>
              </el-button>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <div class="actions">
        <el-button type="success" size="large" @click="goToNextStep" :disabled="!allImagesGenerated">
          {{ $t('workflow.completeAndReturn') }}
        </el-button>
      </div>
      </el-card>
    </LoadingSection>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Edit, Picture } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { characterLibraryAPI } from '@/api/character-library'
import type { Character } from '@/types/drama'
import type { ImageGeneration } from '@/types/image'
import { createListStream } from '@/utils/generationManager'
import { LoadingSection } from '@/components/common'
import { getCache, setCache } from '@/utils/cache'

const route = useRoute()
const router = useRouter()
const { t: $t } = useI18n()
const dramaId = route.params.id as string

const characters = ref<Character[]>([])
const generatingIds = ref<(number | string)[]>([])
const batchGenerating = ref(false)
const selectedCharacters = ref<(number | string)[]>([])
const selectAll = ref(false)
const pageLoading = ref(false)
const cacheTTL = 60 * 1000
let loadingTimer: number | null = null

const getDramaCacheKey = () => `drama:detail:${dramaId}`

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

const applyDramaCharacters = (drama: any, redirectOnEmpty: boolean) => {
  if (drama.characters && drama.characters.length > 0) {
    characters.value = drama.characters
    return
  }
  if (redirectOnEmpty) {
    ElMessage.warning($t('workflow.charactersNotFound'))
    router.push(`/dramas/${dramaId}`)
  }
}

const hydrateDramaFromCache = () => {
  const cached = getCache<any>(getDramaCacheKey(), cacheTTL)
  if (!cached) return false
  applyDramaCharacters(cached, false)
  return true
}

const loadDrama = async (showLoading = true) => {
  if (showLoading) {
    startPageLoading()
  }
  try {
    const drama = await dramaAPI.get(dramaId)
    applyDramaCharacters(drama, true)
    setCache(getDramaCacheKey(), drama)
  } catch (error: any) {
    ElMessage.error(error.message || $t('workflow.loadCharactersFailed'))
    router.push(`/dramas/${dramaId}`)
  } finally {
    if (showLoading) {
      stopPageLoading()
    }
  }
}

const allImagesGenerated = computed(() => {
  return characters.value.length > 0 && characters.value.every(c => c.image_url)
})

const isIndeterminate = computed(() => {
  const selectedCount = selectedCharacters.value.length
  return selectedCount > 0 && selectedCount < characters.value.length
})

const goBack = () => {
  router.push(`/dramas/${dramaId}`)
}

const goToCharacterManagement = () => {
  router.push(`/dramas/${dramaId}/characters`)
}

const isSelected = (id: number | string) => {
  return selectedCharacters.value.includes(id)
}

const toggleSelection = (id: number | string) => {
  const index = selectedCharacters.value.indexOf(id)
  if (index > -1) {
    selectedCharacters.value.splice(index, 1)
  } else {
    selectedCharacters.value.push(id)
  }
  updateSelectAllState()
}

const handleSelectAll = (val: boolean) => {
  if (val) {
    selectedCharacters.value = characters.value.map(c => c.id)
  } else {
    selectedCharacters.value = []
  }
}

const updateSelectAllState = () => {
  selectAll.value = selectedCharacters.value.length === characters.value.length
}

const generateImage = async (character: Character) => {
  if (generatingIds.value.includes(character.id)) return
  
  generatingIds.value.push(character.id)
  try {
    const result = await characterLibraryAPI.generateCharacterImage(character.id as string)
    
    // 更新角色图片
    const index = characters.value.findIndex(c => c.id === character.id)
    if (index !== -1) {
      characters.value[index].image_url = result.image_url
    }
    
    ElMessage.success($t('workflow.characterImageGenerated', { name: character.name }))
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || $t('workflow.characterImageFailed', { name: character.name }))
  } finally {
    const index = generatingIds.value.indexOf(character.id)
    if (index > -1) {
      generatingIds.value.splice(index, 1)
    }
  }
}

const batchGenerate = async () => {
  if (selectedCharacters.value.length === 0) {
    ElMessage.warning($t('workflow.selectCharactersWarning'))
    return
  }

  if (selectedCharacters.value.length > 10) {
    ElMessage.warning($t('workflow.batchGenerateLimit'))
    return
  }

  batchGenerating.value = true
  generatingIds.value = [...selectedCharacters.value]
  
  try {
    const response = await characterLibraryAPI.batchGenerateCharacterImages(
      selectedCharacters.value.map(id => String(id))
    )
    
    ElMessage.success($t('workflow.batchCharacterSubmitted', { count: selectedCharacters.value.length }))
    
    const items = response.items || []
    if (items.length > 0) {
      const successIds: Array<number | string> = []
      const failedItems = items.filter(item => !item.image_generation_id)
      items.forEach((item) => {
        const charId = Number(item.character_id)
        const idx = characters.value.findIndex(c => c.id === charId)
        if (item.image_generation_id) {
          successIds.push(charId)
          if (idx !== -1) {
            characters.value[idx] = {
              ...characters.value[idx],
              image_generation_status: item.status || 'pending',
              image_generation_id: item.image_generation_id
            }
          }
        } else if (idx !== -1) {
          characters.value[idx] = {
            ...characters.value[idx],
            image_generation_status: 'failed',
            image_generation_error: item.error || $t('common.generateFailed')
          }
        }
      })
      generatingIds.value = successIds
      if (failedItems.length > 0) {
        ElMessage.warning($t('workflow.batchSubmitSummary', { success: items.length - failedItems.length, fail: failedItems.length }))
      }
      if (successIds.length > 0) {
        startPolling()
      } else {
        batchGenerating.value = false
      }
    } else {
      // 兼容旧接口响应
      startPolling()
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || $t('workflow.batchGenerateFailed'))
    batchGenerating.value = false
    generatingIds.value = []
  }
}

const isAllSelectedGenerated = () => {
  return selectedCharacters.value.every(id => {
    const char = characters.value.find(c => c.id === id)
    return char?.image_url || char?.image_generation_status === 'failed'
  })
}

const refreshCharacters = async () => {
  try {
    const drama = await dramaAPI.get(dramaId)
    if (drama.characters) {
      characters.value = drama.characters
    }
  } catch (error) {
    console.error('轮询错误:', error)
  }
}

const handleImageEvent = (imageGen: ImageGeneration) => {
  if (!imageGen?.character_id) return
  if (!selectedCharacters.value.includes(imageGen.character_id)) return

  const idx = characters.value.findIndex(c => c.id === imageGen.character_id)
  if (idx !== -1) {
    characters.value[idx] = {
      ...characters.value[idx],
      image_url: imageGen.image_url || characters.value[idx].image_url,
      image_generation_status: imageGen.status
    }
  }

  if (isAllSelectedGenerated()) {
    stopPolling()
    ElMessage.success($t('workflow.batchGenerateComplete'))
  }
}

const hasPendingSelected = () => {
  if (selectedCharacters.value.length === 0) return false
  return !isAllSelectedGenerated()
}

const characterStream = createListStream<ImageGeneration>({
  types: ['image_generation'],
  getParams: () => ({ drama_id: dramaId }),
  onMessage: handleImageEvent,
  poll: async () => {
    await refreshCharacters()
    if (isAllSelectedGenerated()) {
      stopPolling()
      ElMessage.success($t('workflow.batchGenerateComplete'))
    }
  },
  shouldPoll: hasPendingSelected,
  pollIntervalMs: 5000,
  scheduleDelayMs: 500
})

let streamActive = false

const startPolling = () => {
  if (streamActive) return
  if (!hasPendingSelected()) return
  streamActive = true
  characterStream.start()
}

const stopPolling = () => {
  if (streamActive) {
    characterStream.stop()
    streamActive = false
  }
  batchGenerating.value = false
  generatingIds.value = []
  selectedCharacters.value = []
  selectAll.value = false
}

const goToNextStep = () => {
  router.push(`/dramas/${dramaId}`)
}

onMounted(async () => {
  const hasCache = hydrateDramaFromCache()
  await loadDrama(!hasCache)
})

// 组件销毁时清理轮询
onBeforeUnmount(() => {
  stopPolling()
  if (loadingTimer) {
    window.clearTimeout(loadingTimer)
    loadingTimer = null
  }
})
</script>

<style scoped>
.character-images-container {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.main-card {
  margin-top: 20px;
}

.character-card {
  margin-bottom: 20px;
  text-align: center;
}

.character-card.has-image {
  border-color: #67c23a;
}

.character-preview {
  height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.character-preview img {
  max-width: 100%;
  max-height: 180px;
  border-radius: 8px;
}

.character-info h4 {
  margin: 8px 0;
}

.character-info .role {
  color: #909399;
  font-size: 13px;
  margin: 4px 0;
}

.character-info .desc {
  color: #606266;
  font-size: 12px;
  margin: 8px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.toolbar {
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.selection-info {
  color: #606266;
  font-size: 14px;
}

.character-card {
  position: relative;
  transition: all 0.3s;
}

.character-card.selected {
  border-color: #409eff;
  box-shadow: 0 2px 12px 0 rgba(64, 158, 255, 0.3);
}

.card-checkbox {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 1;
}

.actions {
  margin-top: 30px;
  text-align: center;
}
</style>
