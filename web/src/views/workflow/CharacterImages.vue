<template>
  <div class="character-images-container">
    <el-page-header @back="goBack" title="返回项目">
      <template #content>
        <h2>角色形象生成</h2>
      </template>
      <template #extra>
        <el-button type="primary" @click="batchGenerate" :loading="batchGenerating" :disabled="selectedCharacters.length === 0">
          <el-icon><Picture /></el-icon>
          批量生成 ({{ selectedCharacters.length }})
        </el-button>
        <el-button @click="goToCharacterManagement">
          <el-icon><Edit /></el-icon>
          管理角色
        </el-button>
      </template>
    </el-page-header>

    <el-card
      shadow="never"
      class="main-card"
      v-loading="pageLoading"
      :element-loading-text="$t('common.loading')"
      element-loading-background="transparent"
      element-loading-custom-class="glass-loading"
    >
      <div class="toolbar">
        <el-checkbox v-model="selectAll" @change="handleSelectAll" :indeterminate="isIndeterminate">
          全选
        </el-checkbox>
        <span class="selection-info">已选择 {{ selectedCharacters.length }} / {{ characters.length }} 个角色</span>
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
                <span v-if="generatingIds.includes(character.id)">生成中...</span>
                <span v-else>{{ character.image_url ? '重新生成' : '生成形象' }}</span>
              </el-button>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <div class="actions">
        <el-button type="success" size="large" @click="goToNextStep" :disabled="!allImagesGenerated">
          完成并返回项目
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Edit, Picture } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { characterLibraryAPI } from '@/api/character-library'
import type { Character } from '@/types/drama'
import { buildSSEUrl, subscribeSSE } from '@/utils/sse'

const route = useRoute()
const router = useRouter()
const dramaId = route.params.id as string

const characters = ref<Character[]>([])
const generatingIds = ref<(number | string)[]>([])
const batchGenerating = ref(false)
const selectedCharacters = ref<(number | string)[]>([])
const selectAll = ref(false)
const pageLoading = ref(false)

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
    
    ElMessage.success(`${character.name}的形象生成成功`)
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || `${character.name}生成失败`)
  } finally {
    const index = generatingIds.value.indexOf(character.id)
    if (index > -1) {
      generatingIds.value.splice(index, 1)
    }
  }
}

const batchGenerate = async () => {
  if (selectedCharacters.value.length === 0) {
    ElMessage.warning('请选择要生成的角色')
    return
  }

  if (selectedCharacters.value.length > 10) {
    ElMessage.warning('单次最多生成10个角色')
    return
  }

  batchGenerating.value = true
  generatingIds.value = [...selectedCharacters.value]
  
  try {
    await characterLibraryAPI.batchGenerateCharacterImages(
      selectedCharacters.value.map(id => String(id))
    )
    
    ElMessage.success(`批量生成任务已提交，正在后台生成 ${selectedCharacters.value.length} 个角色形象`)
    
    // 轮询检查生成状态
    startPolling()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '批量生成失败')
    batchGenerating.value = false
    generatingIds.value = []
  }
}

let pollingTimer: number | null = null
let sseStop: (() => void) | null = null

const stopSSE = () => {
  if (sseStop) {
    sseStop()
    sseStop = null
  }
}

const startPolling = () => {
  if (pollingTimer || sseStop) return

  const handleImageEvent = (imageGen: any) => {
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

    const allGenerated = selectedCharacters.value.every(id => {
      const char = characters.value.find(c => c.id === id)
      return char?.image_url || char?.image_generation_status === 'failed'
    })

    if (allGenerated) {
      stopPolling()
      ElMessage.success('批量生成完成')
    }
  }

  const startFallback = () => {
    if (pollingTimer) return
    pollingTimer = window.setInterval(async () => {
      try {
        const drama = await dramaAPI.get(dramaId)
        if (drama.characters) {
          characters.value = drama.characters

          const allGenerated = selectedCharacters.value.every(id => {
            const char = characters.value.find(c => c.id === id)
            return char?.image_url || char?.image_generation_status === 'failed'
          })

          if (allGenerated) {
            stopPolling()
            ElMessage.success('批量生成完成')
          }
        }
      } catch (error) {
        console.error('轮询错误:', error)
      }
    }, 5000)
    return () => {
      if (pollingTimer) {
        clearInterval(pollingTimer)
        pollingTimer = null
      }
    }
  }

  const url = buildSSEUrl('/api/v1/events/image-generations', { drama_id: dramaId })
  sseStop = subscribeSSE({
    url,
    event: 'image_generation',
    onMessage: handleImageEvent,
    fallback: startFallback
  }).close
}

const stopPolling = () => {
  stopSSE()
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
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
  pageLoading.value = true
  try {
    const drama = await dramaAPI.get(dramaId)
    if (drama.characters && drama.characters.length > 0) {
      characters.value = drama.characters
    } else {
      ElMessage.warning('未找到角色信息，请先完成剧本生成')
      router.push(`/dramas/${dramaId}`)
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载角色失败')
    router.push(`/dramas/${dramaId}`)
  } finally {
    pageLoading.value = false
  }
})

// 组件销毁时清理轮询
import { onBeforeUnmount } from 'vue'
onBeforeUnmount(() => {
  stopPolling()
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
