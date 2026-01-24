<template>
  <!-- Create Drama Dialog / 创建短剧弹窗 -->
  <el-dialog
    v-model="visible"
    :title="$t('drama.createNew')"
    width="1200px"
    :close-on-click-modal="false"
    class="create-dialog"
    @closed="handleClosed"
  >
    <div class="dialog-desc">{{ $t('drama.createDesc') }}</div>
    
    <el-form 
      ref="formRef" 
      :model="form" 
      :rules="rules" 
      label-position="top"
      class="create-form"
      @submit.prevent="handleSubmit"
    >
      <div class="create-layout">
        <div class="form-left">
          <el-form-item :label="$t('drama.projectName')" prop="title" required>
            <el-input 
              v-model="form.title" 
              :placeholder="$t('drama.projectNamePlaceholder')"
              size="large"
              maxlength="100"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="广告主" prop="brand_id">
            <el-skeleton v-if="brandsLoading" :rows="1" animated />
            <template v-else>
              <el-select
                v-model="form.brand_id"
                placeholder="请选择广告主（可选）"
                size="large"
                clearable
                @change="handleBrandChange"
              >
                <el-option
                  v-for="brand in brands"
                  :key="brand.id"
                  :label="brand.display_name || brand.name"
                  :value="brand.id"
                />
              </el-select>
            </template>
          </el-form-item>

          <el-form-item label="素材规范" prop="spec_id">
            <el-select
              v-model="form.spec_id"
              placeholder="请选择规范模板（可选）"
              size="large"
              clearable
              :disabled="specs.length === 0"
            >
              <el-option
                v-for="spec in specs"
                :key="spec.id"
                :label="spec.name"
                :value="spec.id"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="$t('drama.projectDesc')" prop="description">
            <el-input 
              v-model="form.description" 
              type="textarea" 
              :rows="6"
              :placeholder="$t('drama.projectDescPlaceholder')"
              maxlength="500"
              show-word-limit
              resize="none"
            />
          </el-form-item>
        </div>

        <div class="form-right">
          <el-form-item label="项目风格" prop="style" class="style-form-item">
            <el-skeleton v-if="stylesLoading" :rows="3" animated />
            <template v-else>
              <StylePicker v-if="styles.length > 0" v-model="form.style" :styles="styles" :columns="6" />
              <div v-else class="style-empty">暂无可用风格</div>
            </template>
          </el-form-item>
        </div>
      </div>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button size="large" @click="handleClose">
          {{ $t('common.cancel') }}
        </el-button>
        <el-button 
          type="primary" 
          size="large"
          :loading="loading"
          :disabled="stylesLoading || styles.length === 0"
          @click="handleSubmit"
        >
          <el-icon v-if="!loading"><Plus /></el-icon>
          {{ $t('drama.createNew') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { brandAPI } from '@/api/brand'
import type { CreateDramaRequest } from '@/types/drama'
import type { Brand, BrandSpec } from '@/types/brand'
import type { StyleOption } from '@/types/style'
import { styleAPI } from '@/api/style'
import StylePicker from './StylePicker.vue'

/**
 * CreateDramaDialog - Reusable dialog for creating new drama projects
 * 创建短剧弹窗 - 可复用的创建短剧项目弹窗
 */
const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'created': [id: string]
}>()

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const stylesLoading = ref(false)
const styles = ref<StyleOption[]>([])
const brandsLoading = ref(false)
const brands = ref<Brand[]>([])
const specs = ref<BrandSpec[]>([])

// v-model binding / 双向绑定
const visible = ref(props.modelValue)
watch(() => props.modelValue, (val) => {
  visible.value = val
})
watch(visible, (val) => {
  emit('update:modelValue', val)
})

// Form data / 表单数据
const form = reactive<CreateDramaRequest>({
  title: '',
  description: '',
  style: '',
  brand_id: undefined,
  spec_id: undefined
})

// Validation rules / 验证规则
const rules: FormRules = {
  title: [
    { required: true, message: '请输入项目标题', trigger: 'blur' },
    { min: 1, max: 100, message: '标题长度在 1 到 100 个字符', trigger: 'blur' }
  ]
}

// Reset form when dialog closes / 关闭时重置表单
const handleClosed = () => {
  form.title = ''
  form.description = ''
  form.style = ''
  form.brand_id = undefined
  form.spec_id = undefined
  specs.value = []
  formRef.value?.resetFields()
}

const applyDefaultStyle = () => {
  if (styles.value.length === 0) {
    form.style = ''
    return
  }
  const defaultStyle = styles.value.find((style) => style.is_default) || styles.value[0]
  form.style = defaultStyle?.key || ''
}

const loadStyles = async () => {
  if (stylesLoading.value) return
  stylesLoading.value = true
  try {
    styles.value = await styleAPI.list()
    applyDefaultStyle()
  } catch (error: any) {
    styles.value = []
  } finally {
    stylesLoading.value = false
  }
}

const loadBrands = async () => {
  if (brandsLoading.value) return
  brandsLoading.value = true
  try {
    brands.value = await brandAPI.list({ include_inactive: false, with_specs: true })
  } catch (error: any) {
    brands.value = []
  } finally {
    brandsLoading.value = false
  }
}

const handleBrandChange = (value: number | undefined) => {
  const brand = brands.value.find((item) => item.id === value)
  specs.value = brand?.specs || []
  if (specs.value.length === 0) {
    form.spec_id = undefined
    return
  }
  const defaultSpec = specs.value.find((spec) => spec.is_default) || specs.value[0]
  form.spec_id = defaultSpec?.id
}

watch(visible, (val) => {
  if (val) {
    loadStyles()
    loadBrands()
  }
})

// Close dialog / 关闭弹窗
const handleClose = () => {
  visible.value = false
}

// Submit form / 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const drama = await dramaAPI.create(form)
        ElMessage.success('创建成功')
        visible.value = false
        emit('created', drama.id)
        // Navigate to drama detail page / 跳转到短剧详情页
        router.push(`/dramas/${drama.id}`)
      } catch (error: any) {
        ElMessage.error(error.message || '创建失败')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
/* ========================================
   Dialog Styles / 弹窗样式
   ======================================== */
.create-dialog :deep(.el-dialog) {
  border-radius: var(--radius-xl);
  max-width: calc(100vw - 48px);
}

.create-dialog :deep(.el-dialog__header) {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-primary);
  margin-right: 0;
}

.create-dialog :deep(.el-dialog__title) {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.create-dialog :deep(.el-dialog__body) {
  padding: 1.5rem;
}

.dialog-desc {
  margin-bottom: 1.5rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

/* ========================================
   Form Styles / 表单样式
   ======================================== */
.create-form :deep(.el-form-item) {
  margin-bottom: 1.25rem;
}

.create-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.create-form :deep(.el-input__wrapper),
.create-form :deep(.el-textarea__inner) {
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
  transition: all var(--transition-fast);
}

.create-form :deep(.el-input__wrapper:hover),
.create-form :deep(.el-textarea__inner:hover) {
  box-shadow: 0 0 0 1px var(--border-secondary) inset;
}

.create-form :deep(.el-input__wrapper.is-focus),
.create-form :deep(.el-textarea__inner:focus) {
  box-shadow: 0 0 0 2px var(--accent) inset;
}

.create-form :deep(.el-input__inner),
.create-form :deep(.el-textarea__inner) {
  color: var(--text-primary);
}

.create-form :deep(.el-input__inner::placeholder),
.create-form :deep(.el-textarea__inner::placeholder) {
  color: var(--text-muted);
}

.create-form :deep(.el-input__count) {
  color: var(--text-muted);
  background: transparent;
}

.create-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(560px, 680px);
  gap: 24px;
  align-items: start;
}

.form-right {
  padding-left: 16px;
  border-left: 1px solid var(--border-primary);
}

.style-form-item :deep(.el-form-item__label) {
  margin-bottom: 0.75rem;
}

.style-empty {
  padding: 12px;
  border: 1px dashed var(--border-primary);
  border-radius: var(--radius-md);
  color: var(--text-muted);
  font-size: 13px;
  text-align: center;
}

/* ========================================
   Footer Styles / 底部样式
   ======================================== */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.dialog-footer .el-button {
  min-width: 100px;
}

@media (max-width: 900px) {
  .create-layout {
    grid-template-columns: 1fr;
  }

  .form-right {
    padding-left: 0;
    border-left: none;
  }
}
</style>
