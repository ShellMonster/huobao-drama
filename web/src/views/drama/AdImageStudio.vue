<template>
  <div class="ad-image-studio">
    <el-tabs v-model="activeTab" class="studio-tabs">
      <el-tab-pane label="文生图" name="text">
        <div class="studio-panel">
          <el-form label-position="top" class="prompt-form">
            <el-form-item label="广告需求描述">
              <el-input
                v-model="textDemand"
                type="textarea"
                :rows="4"
                placeholder="描述产品、目标受众、场景、氛围、卖点等"
                maxlength="4000"
                show-word-limit
              />
            </el-form-item>
            <el-form-item label="生成数量">
              <el-input-number v-model="textPromptCount" :min="1" :max="30" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="textPromptLoading" @click="generateTextPrompts">
                生成提示词
              </el-button>
            </el-form-item>
          </el-form>

        </div>
      </el-tab-pane>

      <el-tab-pane label="图生图" name="image">
        <div class="studio-panel">
          <el-form label-position="top" class="prompt-form">
            <el-form-item label="上传参考图">
              <el-upload
                class="upload-box"
                :action="uploadAction"
                :show-file-list="false"
                :on-success="handleImageUpload"
                :on-error="handleUploadError"
                accept="image/jpeg,image/png,image/jpg,image/webp"
              >
                <el-button>上传图片</el-button>
              </el-upload>
              <div v-if="referenceImage" class="reference-preview">
                <img :src="referenceImage" alt="参考图" />
              </div>
            </el-form-item>
            <el-form-item label="生成数量">
              <el-input-number v-model="imagePromptCount" :min="1" :max="30" />
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                :loading="imagePromptLoading"
                :disabled="!referenceImage"
                @click="generateImagePrompts"
              >
                解析提示词
              </el-button>
            </el-form-item>
          </el-form>

        </div>
      </el-tab-pane>
    </el-tabs>

    <div class="generated-section">
      <div class="section-title">已生成图片</div>
      <div class="prompt-grid" v-if="activeTab === 'text' && textPrompts.length > 0" v-loading="imageGenerating">
        <div class="prompt-toolbar">
          <el-checkbox
            v-model="textSelectAll"
            :indeterminate="textIndeterminate"
            @change="handleSelectAll('text')"
          >
            全选
          </el-checkbox>
          <span class="selection-info">
            已选 {{ selectedTextPrompts.length }} / {{ textPrompts.length }}
          </span>
          <el-button
            type="primary"
            size="small"
            :loading="imageGenerating"
            :disabled="selectedTextPrompts.length === 0 || imageGenerating"
            @click="generateImages(selectedTextPrompts)"
          >
            批量生成
          </el-button>
        </div>

        <el-row :gutter="16">
          <el-col
            v-for="(prompt, index) in textPrompts"
            :key="`${prompt}-${index}`"
            :xs="24"
            :sm="12"
            :md="8"
            :lg="6"
          >
            <PromptCard
              :prompt="prompt"
              :selected="isPromptSelected('text', prompt)"
              :loading="imageGenerating"
              :disabled="imageGenerating"
              @select="togglePromptSelection('text', prompt)"
              @preview="openPromptPreview('text', prompt)"
              @edit="openEditPrompt('text', prompt, index)"
              @copy="copyPrompt(prompt)"
              @generate="generateImages([prompt])"
            />
          </el-col>
        </el-row>
      </div>
      <div class="prompt-grid" v-if="activeTab === 'image' && imagePrompts.length > 0" v-loading="imageGenerating">
        <div class="prompt-toolbar">
          <el-checkbox
            v-model="imageSelectAll"
            :indeterminate="imageIndeterminate"
            @change="handleSelectAll('image')"
          >
            全选
          </el-checkbox>
          <span class="selection-info">
            已选 {{ selectedImagePrompts.length }} / {{ imagePrompts.length }}
          </span>
          <el-button
            type="primary"
            size="small"
            :loading="imageGenerating"
            :disabled="selectedImagePrompts.length === 0 || imageGenerating"
            @click="generateImages(selectedImagePrompts, referenceImage)"
          >
            批量生成
          </el-button>
        </div>

        <el-row :gutter="16">
          <el-col
            v-for="(prompt, index) in imagePrompts"
            :key="`${prompt}-${index}`"
            :xs="24"
            :sm="12"
            :md="8"
            :lg="6"
          >
            <PromptCard
              :prompt="prompt"
              :selected="isPromptSelected('image', prompt)"
              :loading="imageGenerating"
              :disabled="imageGenerating"
              @select="togglePromptSelection('image', prompt)"
              @preview="openPromptPreview('image', prompt)"
              @edit="openEditPrompt('image', prompt, index)"
              @copy="copyPrompt(prompt)"
              @generate="generateImages([prompt], referenceImage)"
            />
          </el-col>
        </el-row>
      </div>
      <LoadingSection :loading="imageListLoading">
        <el-row :gutter="16">
          <el-col
            v-for="image in adImages"
            :key="image.id"
            :xs="24"
            :sm="12"
            :md="8"
            :lg="6"
          >
            <el-card shadow="hover" class="image-card">
              <div class="image-wrapper">
                <el-image
                  v-if="image.status === 'completed' && image.image_url"
                  :src="image.image_url"
                  fit="cover"
                  class="image"
                  @click="viewDetails(image)"
                />
                <div v-else class="image-placeholder">
                  <LoadingIcon v-if="image.status === 'processing'" />
                  <el-icon v-else><Picture /></el-icon>
                  <span>
                    {{ image.status === 'processing' ? '生成中' : image.status === 'failed' ? '生成失败' : '等待生成' }}
                  </span>
                </div>
              </div>
              <div class="image-info">
                <div class="prompt-text">{{ truncateText(image.prompt, 60) }}</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-empty v-if="!imageListLoading && adImages.length === 0 && !hasActivePrompts" description="暂无图片" />
      </LoadingSection>
    </div>
  </div>

  <el-dialog
    v-model="showPromptDialog"
    title="提示词预览"
    width="860px"
    :close-on-click-modal="true"
  >
    <div class="prompt-preview">
      <el-row :gutter="20">
        <el-col :span="10">
          <div class="prompt-preview-left">
            <el-icon :size="32"><Document /></el-icon>
            <p>{{ previewTypeLabel }}</p>
          </div>
        </el-col>
        <el-col :span="14">
          <div class="prompt-preview-right">
            <div class="prompt-preview-header">
              <span>提示词内容</span>
              <el-button text size="small" @click="copyPrompt(previewPrompt)">
                复制
              </el-button>
            </div>
            <div class="prompt-preview-text">{{ previewPrompt }}</div>
          </div>
        </el-col>
      </el-row>
    </div>
  </el-dialog>

  <el-dialog v-model="showEditDialog" title="编辑提示词" width="720px">
    <el-input
      v-model="editPromptValue"
      type="textarea"
      :rows="6"
      maxlength="4000"
      show-word-limit
      placeholder="请输入提示词"
    />
    <template #footer>
      <el-button @click="showEditDialog = false">取消</el-button>
      <el-button type="primary" @click="applyPromptEdit">保存</el-button>
    </template>
  </el-dialog>

  <ImageDetailDialog
    v-model="showDetailDialog"
    :image="selectedImage"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Document, Picture } from '@element-plus/icons-vue'
import { adPromptAPI } from '@/api/ad-image-prompt'
import { imageAPI } from '@/api/image'
import { dramaAPI } from '@/api/drama'
import type { Drama } from '@/types/drama'
import type { ImageGeneration } from '@/types/image'
import { LoadingIcon, LoadingSection } from '@/components/common'
import ImageDetailDialog from '@/views/generation/components/ImageDetailDialog.vue'
import PromptCard from './components/PromptCard.vue'

const props = defineProps<{
  dramaId: string
}>()

const uploadAction = '/api/v1/upload/image'
const activeTab = ref<'text' | 'image'>('text')
const textDemand = ref('')
const textPromptCount = ref(10)
const imagePromptCount = ref(10)
const textPrompts = ref<string[]>([])
const imagePrompts = ref<string[]>([])
const selectedTextPrompts = ref<string[]>([])
const selectedImagePrompts = ref<string[]>([])
const textPromptLoading = ref(false)
const imagePromptLoading = ref(false)
const imageGenerating = ref(false)
const referenceImage = ref('')
const adImages = ref<ImageGeneration[]>([])
const imageListLoading = ref(false)
const drama = ref<Drama | null>(null)
const showDetailDialog = ref(false)
const selectedImage = ref<ImageGeneration>()
const showPromptDialog = ref(false)
const previewPrompt = ref('')
const previewType = ref<'text' | 'image'>('text')
const showEditDialog = ref(false)
const editPromptValue = ref('')
const editPromptIndex = ref<number | null>(null)
const editPromptType = ref<'text' | 'image'>('text')
const textSelectAll = ref(false)
const imageSelectAll = ref(false)

const brandId = computed(() => drama.value?.brand_id)
const specId = computed(() => drama.value?.spec_id)
const hasActivePrompts = computed(() =>
  activeTab.value === 'text' ? textPrompts.value.length > 0 : imagePrompts.value.length > 0
)
const previewTypeLabel = computed(() => (previewType.value === 'text' ? '文生图提示词' : '图生图提示词'))
const textIndeterminate = computed(() =>
  selectedTextPrompts.value.length > 0 && selectedTextPrompts.value.length < textPrompts.value.length
)
const imageIndeterminate = computed(() =>
  selectedImagePrompts.value.length > 0 && selectedImagePrompts.value.length < imagePrompts.value.length
)

const loadDrama = async () => {
  try {
    drama.value = await dramaAPI.get(props.dramaId)
  } catch (error: any) {
    ElMessage.error(error?.message || '加载项目信息失败')
  }
}

const loadImages = async () => {
  if (imageListLoading.value) return
  imageListLoading.value = true
  try {
    const res = await imageAPI.listImages({
      drama_id: props.dramaId,
      image_type: 'ad',
      page: 1,
      page_size: 50
    })
    adImages.value = res.items || []
  } catch (error: any) {
    ElMessage.error(error?.message || '加载图片失败')
    adImages.value = []
  } finally {
    imageListLoading.value = false
  }
}

const loadLatestPrompts = async () => {
  try {
    const res = await adPromptAPI.getLatest({
      drama_id: props.dramaId,
      brand_id: brandId.value,
      spec_id: specId.value
    })
    if (res.text_prompts?.length) {
      textPrompts.value = res.text_prompts
      selectedTextPrompts.value = [...textPrompts.value]
    }
    if (res.image_prompts?.length) {
      imagePrompts.value = res.image_prompts
      selectedImagePrompts.value = [...imagePrompts.value]
    }
  } catch (error: any) {
    ElMessage.error(error?.message || '加载提示词失败')
  }
}

const isPromptSelected = (type: 'text' | 'image', prompt: string) => {
  return type === 'text'
    ? selectedTextPrompts.value.includes(prompt)
    : selectedImagePrompts.value.includes(prompt)
}

const togglePromptSelection = (type: 'text' | 'image', prompt: string) => {
  const list = type === 'text' ? selectedTextPrompts.value : selectedImagePrompts.value
  const index = list.indexOf(prompt)
  if (index >= 0) {
    list.splice(index, 1)
  } else {
    list.push(prompt)
  }
  updateSelectAllState(type)
}

const handleSelectAll = (type: 'text' | 'image') => {
  if (type === 'text') {
    selectedTextPrompts.value = textSelectAll.value ? [...textPrompts.value] : []
  } else {
    selectedImagePrompts.value = imageSelectAll.value ? [...imagePrompts.value] : []
  }
}

const updateSelectAllState = (type: 'text' | 'image') => {
  if (type === 'text') {
    textSelectAll.value = selectedTextPrompts.value.length === textPrompts.value.length && textPrompts.value.length > 0
  } else {
    imageSelectAll.value = selectedImagePrompts.value.length === imagePrompts.value.length && imagePrompts.value.length > 0
  }
}

const openPromptPreview = (type: 'text' | 'image', prompt: string) => {
  previewType.value = type
  previewPrompt.value = prompt
  showPromptDialog.value = true
}

const copyPrompt = async (prompt?: string) => {
  if (!prompt) return
  try {
    await navigator.clipboard.writeText(prompt)
    ElMessage.success('已复制')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const openEditPrompt = (type: 'text' | 'image', prompt: string, index: number) => {
  editPromptType.value = type
  editPromptIndex.value = index
  editPromptValue.value = prompt
  showEditDialog.value = true
}

const applyPromptEdit = () => {
  if (!editPromptValue.value.trim() || editPromptIndex.value === null) {
    ElMessage.warning('请输入提示词')
    return
  }
  const list = editPromptType.value === 'text' ? textPrompts.value : imagePrompts.value
  const selected = editPromptType.value === 'text' ? selectedTextPrompts.value : selectedImagePrompts.value
  const index = editPromptIndex.value
  const oldPrompt = list[index]
  list.splice(index, 1, editPromptValue.value.trim())
  const selectedIndex = selected.indexOf(oldPrompt)
  if (selectedIndex >= 0) {
    selected.splice(selectedIndex, 1, editPromptValue.value.trim())
  }
  showEditDialog.value = false
}

const viewDetails = (image: ImageGeneration) => {
  selectedImage.value = image
  showDetailDialog.value = true
}

const generateTextPrompts = async () => {
  if (!textDemand.value.trim()) {
    ElMessage.warning('请输入广告需求描述')
    return
  }
  textPromptLoading.value = true
  try {
    const res = await adPromptAPI.generateText({
      drama_id: props.dramaId,
      brand_id: brandId.value,
      spec_id: specId.value,
      prompt: textDemand.value,
      count: textPromptCount.value
    })
    textPrompts.value = res.prompts || []
    selectedTextPrompts.value = [...textPrompts.value]
    updateSelectAllState('text')
  } catch (error: any) {
    ElMessage.error(error?.message || '生成提示词失败')
  } finally {
    textPromptLoading.value = false
  }
}

const generateImagePrompts = async () => {
  if (!referenceImage.value) {
    ElMessage.warning('请先上传参考图')
    return
  }
  imagePromptLoading.value = true
  try {
    const res = await adPromptAPI.generateFromImage({
      drama_id: props.dramaId,
      brand_id: brandId.value,
      spec_id: specId.value,
      image_url: referenceImage.value,
      count: imagePromptCount.value
    })
    imagePrompts.value = res.prompts || []
    selectedImagePrompts.value = [...imagePrompts.value]
    updateSelectAllState('image')
  } catch (error: any) {
    ElMessage.error(error?.message || '解析提示词失败')
  } finally {
    imagePromptLoading.value = false
  }
}

const generateImages = async (prompts: string[], reference?: string) => {
  if (imageGenerating.value) return
  imageGenerating.value = true
  try {
    for (const prompt of prompts) {
      await imageAPI.generateImage({
        drama_id: props.dramaId,
        prompt,
        image_type: 'ad',
        reference_images: reference ? [reference] : undefined
      })
    }
    ElMessage.success('已提交生成任务')
    await loadImages()
  } catch (error: any) {
    ElMessage.error(error?.message || '生成图片失败')
  } finally {
    imageGenerating.value = false
  }
}

const handleImageUpload = (res: any) => {
  if (res?.url) {
    referenceImage.value = res.url
    ElMessage.success('上传成功')
  } else {
    ElMessage.error('上传失败')
  }
}

const handleUploadError = () => {
  ElMessage.error('上传失败')
}

const truncateText = (text: string, max = 80) => {
  if (!text) return ''
  if (text.length <= max) return text
  return text.slice(0, max) + '...'
}

onMounted(() => {
  void loadDrama()
  void loadImages()
  void loadLatestPrompts()
})
</script>

<style scoped>
.ad-image-studio {
  padding: 8px 4px;
}

.studio-panel {
  padding: 8px 0;
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}

.prompt-grid {
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
}

.prompt-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.selection-info {
  color: #606266;
  font-size: 13px;
}

.prompt-text {
  display: inline-block;
  max-width: 100%;
  word-break: break-word;
}

.prompt-preview-left {
  height: 100%;
  min-height: 180px;
  border: 1px dashed var(--border-primary);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
  gap: 8px;
}

.prompt-preview-right {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.prompt-preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
}

.prompt-preview-text {
  padding: 12px;
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  min-height: 180px;
  white-space: pre-wrap;
  word-break: break-word;
  color: #303133;
}

.reference-preview {
  margin-top: 12px;
  width: 160px;
  height: 160px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--border-primary);
}

.reference-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.generated-section {
  margin-top: 16px;
}

.section-title {
  font-weight: 600;
  margin-bottom: 8px;
}

.image-card .image-wrapper {
  height: 160px;
  background: var(--bg-muted, #f5f5f5);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-card .image {
  width: 100%;
  height: 100%;
  cursor: pointer;
}

.image-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  color: var(--text-secondary, #666);
}
</style>
