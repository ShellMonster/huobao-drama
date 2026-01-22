<template>
  <!-- Drama Create Page / 创建短剧页面 -->
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <!-- Header / 头部 -->
      <AppHeader :fixed="false" :show-logo="false">
        <template #left>
          <el-button text @click="goBack" class="back-btn">
            <el-icon><ArrowLeft /></el-icon>
            <span>返回</span>
          </el-button>
          <div class="page-title">
            <h1>创建新项目</h1>
            <span class="subtitle">填写基本信息来创建你的短剧项目</span>
          </div>
        </template>
      </AppHeader>

      <!-- Form Card / 表单卡片 -->
      <div class="form-card">

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
              <el-form-item label="项目标题" prop="title" required>
                <el-input 
                  v-model="form.title" 
                  placeholder="给你的短剧起个名字"
                  size="large"
                  maxlength="100"
                  show-word-limit
                />
              </el-form-item>

              <el-form-item label="项目描述" prop="description">
                <el-input 
                  v-model="form.description" 
                  type="textarea" 
                  :rows="7"
                  placeholder="简要描述你的短剧内容、风格或创意（可选）"
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
                  <StylePicker v-if="styles.length > 0" v-model="form.style" :styles="styles" :columns="4" />
                  <div v-else class="style-empty">暂无可用风格</div>
                </template>
              </el-form-item>
            </div>
          </div>

          <div class="form-actions">
            <el-button size="large" @click="goBack">取消</el-button>
            <el-button 
              type="primary" 
              size="large"
              :loading="loading"
              :disabled="stylesLoading || styles.length === 0"
              @click="handleSubmit"
            >
              <el-icon v-if="!loading"><Plus /></el-icon>
              创建项目
            </el-button>
          </div>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { ArrowLeft, Plus } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { styleAPI } from '@/api/style'
import type { CreateDramaRequest } from '@/types/drama'
import type { StyleOption } from '@/types/style'
import { AppHeader, StylePicker } from '@/components/common'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const stylesLoading = ref(false)
const styles = ref<StyleOption[]>([])

const form = reactive<CreateDramaRequest>({
  title: '',
  description: '',
  style: ''
})

const rules: FormRules = {
  title: [
    { required: true, message: '请输入项目标题', trigger: 'blur' },
    { min: 1, max: 100, message: '标题长度在 1 到 100 个字符', trigger: 'blur' }
  ]
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

// Submit form / 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const drama = await dramaAPI.create(form)
        ElMessage.success('创建成功')
        router.push(`/dramas/${drama.id}`)
      } catch (error: any) {
        ElMessage.error(error.message || '创建失败')
      } finally {
        loading.value = false
      }
    }
  })
}

// Go back / 返回上一页
const goBack = () => {
  router.back()
}

onMounted(() => {
  loadStyles()
})
</script>

<style scoped>
/* ========================================
   Page Layout / 页面布局 - 紧凑边距
   ======================================== */
.page-container {
  min-height: 100vh;
  background-color: var(--bg-primary);
  padding: var(--space-2) var(--space-3);
  transition: background-color var(--transition-normal);
}

@media (min-width: 768px) {
  .page-container {
    padding: var(--space-3) var(--space-4);
  }
}

.content-wrapper {
  max-width: 980px;
  margin: 0 auto;
}

/* ========================================
   Form Card / 表单卡片
   ======================================== */
.form-card {
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-card);
}

/* ========================================
   Form Styles / 表单样式 - 紧凑内边距
   ======================================== */
.create-form {
  padding: var(--space-4);
}

.create-form :deep(.el-form-item) {
  margin-bottom: var(--space-4);
}

.create-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(320px, 460px);
  gap: var(--space-5);
  align-items: start;
}

.form-right {
  padding-left: var(--space-4);
  border-left: 1px solid var(--border-primary);
}

.style-form-item :deep(.el-form-item__label) {
  margin-bottom: var(--space-2);
}

/* ========================================
   Form Actions / 表单操作区
   ======================================== */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  padding-top: var(--space-4);
  border-top: 1px solid var(--border-primary);
  margin-top: var(--space-2);
}

.form-actions .el-button {
  min-width: 100px;
}

.style-empty {
  padding: var(--space-3);
  border: 1px dashed var(--border-primary);
  border-radius: var(--radius-md);
  color: var(--text-muted);
  font-size: 13px;
  text-align: center;
}

@media (max-width: 960px) {
  .content-wrapper {
    max-width: 720px;
  }

  .create-layout {
    grid-template-columns: 1fr;
  }

  .form-right {
    padding-left: 0;
    border-left: none;
  }
}
</style>
