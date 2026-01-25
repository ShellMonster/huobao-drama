<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false" :show-logo="false">
        <template #left>
          <div class="page-title">
            <h1>{{ $t('styleManagement.title') }}</h1>
            <span class="subtitle">{{ $t('styleManagement.subtitle') }}</span>
          </div>
        </template>
        <template #right>
          <el-button type="primary" class="header-btn primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            <span class="btn-text">{{ $t('styleManagement.create') }}</span>
          </el-button>
        </template>
      </AppHeader>

      <LoadingSection :loading="loading">
        <EmptyState
          v-if="!loading && styles.length === 0"
          :title="$t('styleManagement.empty.title')"
          :description="$t('styleManagement.empty.description')"
          :icon="Brush"
        >
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            {{ $t('styleManagement.create') }}
          </el-button>
        </EmptyState>

        <div v-else class="style-grid">
          <div v-for="style in styles" :key="style.id" class="style-card">
            <div class="style-thumb">
              <img
                v-if="style.preview_url && !failedImages[style.key]"
                :src="style.preview_url"
                :alt="style.name"
                @error="markFailed(style.key)"
              />
              <div v-else class="style-thumb-placeholder">
                <span>{{ $t('styleManagement.noPreview') }}</span>
              </div>
              <div class="style-label">
                <span>{{ style.name }}</span>
              </div>
            </div>

            <div class="style-meta">
              <el-tag size="small" type="info">
                {{ style.is_system ? $t('styleManagement.tags.system') : $t('styleManagement.tags.custom') }}
              </el-tag>
              <el-tag v-if="style.is_default" size="small" type="success">{{ $t('styleManagement.tags.default') }}</el-tag>
              <el-tag v-if="style.is_active === false" size="small" type="warning">{{ $t('styleManagement.tags.inactive') }}</el-tag>
            </div>

            <div class="style-actions">
              <el-button
                size="small"
                type="primary"
                link
                :disabled="style.is_default"
                @click="setDefault(style)"
              >
                {{ $t('styleManagement.actions.setDefault') }}
              </el-button>
              <el-button size="small" link @click="openEditDialog(style)">{{ $t('common.edit') }}</el-button>
              <el-popconfirm
                v-if="!style.is_system"
                :title="$t('styleManagement.messages.deleteConfirm')"
                :confirm-button-text="$t('common.delete')"
                :cancel-button-text="$t('common.cancel')"
                @confirm="removeStyle(style)"
              >
                <template #reference>
                  <el-button size="small" type="danger" link>{{ $t('common.delete') }}</el-button>
                </template>
              </el-popconfirm>
              <el-popconfirm
                v-else
                :title="style.is_active === false ? $t('styleManagement.messages.enableConfirm') : $t('styleManagement.messages.disableConfirm')"
                :confirm-button-text="$t('common.confirm')"
                :cancel-button-text="$t('common.cancel')"
                @confirm="toggleActive(style)"
              >
                <template #reference>
                  <el-button size="small" type="warning" link>
                    {{ style.is_active === false ? $t('styleManagement.actions.enable') : $t('styleManagement.actions.disable') }}
                  </el-button>
                </template>
              </el-popconfirm>
            </div>
          </div>
        </div>
      </LoadingSection>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('styleManagement.dialog.editTitle') : $t('styleManagement.dialog.createTitle')"
      width="720px"
      :close-on-click-modal="false"
      destroy-on-close
      class="style-dialog"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item :label="$t('styleManagement.form.name')" prop="name">
          <el-input v-model="form.name" :placeholder="$t('styleManagement.form.namePlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('styleManagement.form.preview')" prop="preview_url">
          <div class="preview-row">
            <el-input v-model="form.preview_url" :placeholder="$t('styleManagement.form.previewPlaceholder')" />
            <el-upload
              class="upload-btn"
              :action="uploadAction"
              :show-file-list="false"
              :on-success="handleUploadSuccess"
              :on-error="handleUploadError"
              accept="image/jpeg,image/png,image/jpg,image/webp"
            >
              <el-button>{{ $t('common.upload') }}</el-button>
            </el-upload>
          </div>
          <div v-if="form.preview_url" class="preview-thumb">
            <img :src="form.preview_url" :alt="$t('styleManagement.form.preview')" />
          </div>
        </el-form-item>

        <el-form-item :label="$t('styleManagement.form.promptZh')" prop="prompt_zh">
          <el-input v-model="form.prompt_zh" type="textarea" :rows="2" :placeholder="$t('styleManagement.form.promptPlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('styleManagement.form.promptEn')" prop="prompt_en">
          <el-input v-model="form.prompt_en" type="textarea" :rows="2" :placeholder="$t('styleManagement.form.promptPlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('styleManagement.form.sortOrder')">
          <el-input-number v-model="form.sort_order" :min="0" :max="9999" />
        </el-form-item>

        <el-form-item :label="$t('styleManagement.form.active')">
          <el-switch v-model="form.is_active" />
        </el-form-item>

        <el-form-item :label="$t('styleManagement.form.isDefault')">
          <el-switch v-model="form.is_default" />
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
          <el-button type="primary" :loading="saving" @click="submitForm">
            {{ isEdit ? $t('common.save') : $t('common.create') }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Brush, Plus } from '@element-plus/icons-vue'
import { AppHeader, EmptyState, LoadingSection } from '@/components/common'
import { styleAPI } from '@/api/style'
import type { StyleCreateRequest, StyleOption, StyleUpdateRequest } from '@/types/style'
import { useI18n } from 'vue-i18n'

const loading = ref(false)
const saving = ref(false)
const styles = ref<StyleOption[]>([])
const dialogVisible = ref(false)
const editingStyle = ref<StyleOption | null>(null)
const failedImages = reactive<Record<string, boolean>>({})
const { t } = useI18n()

const formRef = ref()
const form = reactive<StyleCreateRequest>({
  name: '',
  preview_url: '',
  prompt_zh: '',
  prompt_en: '',
  sort_order: 0,
  is_active: true,
  is_default: false
})

const rules = {
  name: [{ required: true, message: t('styleManagement.validation.nameRequired'), trigger: 'blur' }]
}

const uploadAction = '/api/v1/upload/style'

const isEdit = computed(() => !!editingStyle.value)

const markFailed = (key: string) => {
  failedImages[key] = true
}

const loadStyles = async () => {
  if (loading.value) return
  loading.value = true
  try {
    styles.value = await styleAPI.list({ include_inactive: true })
  } catch (error: any) {
    ElMessage.error(error?.message || t('styleManagement.messages.loadFailed'))
    styles.value = []
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.name = ''
  form.preview_url = ''
  form.prompt_zh = ''
  form.prompt_en = ''
  form.sort_order = styles.value.length + 1
  form.is_active = true
  form.is_default = false
}

const openCreateDialog = () => {
  editingStyle.value = null
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (style: StyleOption) => {
  editingStyle.value = style
  form.name = style.name
  form.preview_url = style.preview_url || ''
  form.prompt_zh = style.prompt_zh || ''
  form.prompt_en = style.prompt_en || ''
  form.sort_order = style.sort_order || 0
  form.is_active = style.is_active !== false
  form.is_default = style.is_default === true
  dialogVisible.value = true
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate()
  saving.value = true
  try {
    if (editingStyle.value) {
      const payload: StyleUpdateRequest = {
        name: form.name,
        preview_url: form.preview_url,
        prompt_zh: form.prompt_zh,
        prompt_en: form.prompt_en,
        sort_order: form.sort_order,
        is_active: form.is_active,
        is_default: form.is_default
      }
      await styleAPI.update(editingStyle.value.id, payload)
      ElMessage.success(t('styleManagement.messages.updateSuccess'))
    } else {
      await styleAPI.create(form)
      ElMessage.success(t('styleManagement.messages.createSuccess'))
    }
    dialogVisible.value = false
    await loadStyles()
  } catch (error: any) {
    ElMessage.error(error?.message || t('common.saveFailed'))
  } finally {
    saving.value = false
  }
}

const setDefault = async (style: StyleOption) => {
  try {
    await styleAPI.update(style.id, { is_default: true, is_active: true })
    ElMessage.success(t('styleManagement.messages.setDefaultSuccess'))
    await loadStyles()
  } catch (error: any) {
    ElMessage.error(error?.message || t('styleManagement.messages.setDefaultFailed'))
  }
}

const toggleActive = async (style: StyleOption) => {
  const currentActive = style.is_active !== false
  try {
    await styleAPI.update(style.id, { is_active: !currentActive })
    ElMessage.success(currentActive ? t('styleManagement.messages.disabled') : t('styleManagement.messages.enabled'))
    await loadStyles()
  } catch (error: any) {
    ElMessage.error(error?.message || t('styleManagement.messages.toggleFailed'))
  }
}

const removeStyle = async (style: StyleOption) => {
  try {
    await styleAPI.remove(style.id)
    ElMessage.success(t('styleManagement.messages.deleteSuccess'))
    await loadStyles()
  } catch (error: any) {
    ElMessage.error(error?.message || t('common.deleteFailed'))
  }
}

const handleUploadSuccess = (response: any) => {
  if (response?.success && response?.data?.url) {
    form.preview_url = response.data.url
  } else if (response?.url) {
    form.preview_url = response.url
  }
}

const handleUploadError = (error: any) => {
  ElMessage.error(error?.message || t('styleManagement.messages.uploadFailed'))
}

onMounted(() => {
  loadStyles()
})
</script>

<style scoped>
.style-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
  margin-top: var(--space-4);
}

.style-card {
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  box-shadow: var(--shadow-card);
}

.style-thumb {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  border-radius: var(--radius-lg);
  overflow: hidden;
  background: linear-gradient(135deg, rgba(231, 236, 242, 0.9), rgba(205, 214, 226, 0.9));
}

.style-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.style-thumb-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #5f6b7a;
  font-size: 12px;
  letter-spacing: 0.4px;
}

.style-label {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 8px 10px;
  background: rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(8px);
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  text-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
}

.style-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.style-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.preview-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
}

.preview-row :deep(.el-input) {
  width: 100%;
  min-width: 0;
}

.preview-row :deep(.el-upload) {
  justify-self: end;
}

.preview-row :deep(.el-button) {
  min-width: 96px;
}

.preview-thumb {
  margin-top: 10px;
  width: 160px;
  height: 160px;
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 1px solid var(--border-primary);
}

.preview-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

@media (max-width: 640px) {
  .preview-row {
    grid-template-columns: 1fr;
    align-items: stretch;
  }

  .preview-row :deep(.el-upload) {
    justify-self: start;
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
