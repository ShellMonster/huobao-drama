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

          <div class="prompt-list" v-if="textPrompts.length > 0">
            <div class="prompt-list-header">
              <span>提示词列表</span>
              <el-button size="small" @click="selectAll(textPrompts, selectedTextPrompts)">全选</el-button>
            </div>
            <el-checkbox-group v-model="selectedTextPrompts">
              <div v-for="prompt in textPrompts" :key="prompt" class="prompt-item">
                <el-checkbox :label="prompt">
                  <span class="prompt-text">{{ prompt }}</span>
                </el-checkbox>
              </div>
            </el-checkbox-group>
            <div class="prompt-actions">
              <el-button
                type="primary"
                :loading="imageGenerating"
                :disabled="selectedTextPrompts.length === 0"
                @click="generateImages(selectedTextPrompts)"
              >
                生成图片
              </el-button>
            </div>
          </div>
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

          <div class="prompt-list" v-if="imagePrompts.length > 0">
            <div class="prompt-list-header">
              <span>提示词列表</span>
              <el-button size="small" @click="selectAll(imagePrompts, selectedImagePrompts)">全选</el-button>
            </div>
            <el-checkbox-group v-model="selectedImagePrompts">
              <div v-for="prompt in imagePrompts" :key="prompt" class="prompt-item">
                <el-checkbox :label="prompt">
                  <span class="prompt-text">{{ prompt }}</span>
                </el-checkbox>
              </div>
            </el-checkbox-group>
            <div class="prompt-actions">
              <el-button
                type="primary"
                :loading="imageGenerating"
                :disabled="selectedImagePrompts.length === 0"
                @click="generateImages(selectedImagePrompts, referenceImage)"
              >
                生成图片
              </el-button>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <div class="generated-section">
      <div class="section-title">已生成图片</div>
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
                  :preview-src-list="[image.image_url]"
                />
                <div v-else class="image-placeholder">
                  <el-icon class="loading-icon" v-if="image.status === 'processing'"><Loading /></el-icon>
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
        <el-empty v-if="!imageListLoading && adImages.length === 0" description="暂无图片" />
      </LoadingSection>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading, Picture } from '@element-plus/icons-vue'
import { adPromptAPI } from '@/api/ad-image-prompt'
import { imageAPI } from '@/api/image'
import { dramaAPI } from '@/api/drama'
import type { Drama } from '@/types/drama'
import type { ImageGeneration } from '@/types/image'
import { LoadingSection } from '@/components/common'

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

const brandId = computed(() => drama.value?.brand_id)
const specId = computed(() => drama.value?.spec_id)

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

const selectAll = (source: string[], target: string[]) => {
  target.splice(0, target.length, ...source)
}

const truncateText = (text: string, max = 80) => {
  if (!text) return ''
  if (text.length <= max) return text
  return text.slice(0, max) + '...'
}

loadDrama()
loadImages()
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

.prompt-list {
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 12px;
}

.prompt-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
}

.prompt-item {
  padding: 6px 0;
  border-bottom: 1px dashed var(--border-secondary);
}

.prompt-item:last-child {
  border-bottom: none;
}

.prompt-text {
  display: inline-block;
  max-width: 100%;
  word-break: break-word;
}

.prompt-actions {
  margin-top: 12px;
  text-align: right;
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
}

.image-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  color: var(--text-secondary, #666);
}
</style>
