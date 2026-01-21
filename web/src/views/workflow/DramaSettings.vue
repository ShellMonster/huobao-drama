<template>
  <div class="drama-settings-container">
    <el-page-header @back="goBack" title="返回项目">
      <template #content>
        <h2>项目设置</h2>
      </template>
    </el-page-header>

    <LoadingSection class="main-card" :loading="pageLoading" text="加载中...">
      <el-card shadow="never">
        <el-tabs v-model="activeTab">
        <el-tab-pane label="基本信息" name="basic">
          <el-form :model="form" label-width="100px" style="max-width: 600px">
            <el-form-item label="项目标题">
              <el-input v-model="form.title" />
            </el-form-item>
            <el-form-item label="项目描述">
              <el-input v-model="form.description" type="textarea" :rows="4" />
            </el-form-item>
            <el-form-item label="类型">
              <el-select v-model="form.genre">
                <el-option label="都市" value="都市" />
                <el-option label="古装" value="古装" />
                <el-option label="悬疑" value="悬疑" />
                <el-option label="爱情" value="爱情" />
                <el-option label="喜剧" value="喜剧" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="form.status">
                <el-option label="草稿" value="draft" />
                <el-option label="策划中" value="planning" />
                <el-option label="制作中" value="production" />
                <el-option label="已完成" value="completed" />
                <el-option label="已归档" value="archived" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="saving" @click="saveSettings">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="危险操作" name="danger">
          <el-alert
            title="警告"
            type="warning"
            description="以下操作不可恢复，请谨慎操作"
            :closable="false"
            show-icon
          />
          <div class="danger-zone">
            <el-button type="danger" :loading="deleting" @click="deleteProject">删除项目</el-button>
          </div>
        </el-tab-pane>
      </el-tabs>
      </el-card>
    </LoadingSection>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { dramaAPI } from '@/api/drama'
import { LoadingSection } from '@/components/common'
import { getCache, setCache } from '@/utils/cache'

const route = useRoute()
const router = useRouter()
const dramaId = route.params.id as string

const activeTab = ref('basic')
const pageLoading = ref(false)
const cacheTTL = 60 * 1000
let loadingTimer: number | null = null
const saving = ref(false)
const deleting = ref(false)
const form = reactive({
  title: '',
  description: '',
  genre: '',
  status: 'draft' as any
})

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

const applyDramaToForm = (drama: any) => {
  form.title = drama.title || ''
  form.description = drama.description || ''
  form.genre = drama.genre || ''
  form.status = drama.status || 'draft'
}

const hydrateDramaFromCache = () => {
  const cached = getCache<any>(getDramaCacheKey(), cacheTTL)
  if (!cached) return false
  applyDramaToForm(cached)
  return true
}

const loadDrama = async (showLoading = true) => {
  if (showLoading) {
    startPageLoading()
  }
  try {
    const drama = await dramaAPI.get(dramaId)
    applyDramaToForm(drama)
    setCache(getDramaCacheKey(), drama)
  } catch (error: any) {
    ElMessage.error(error.message || '加载失败')
  } finally {
    if (showLoading) {
      stopPageLoading()
    }
  }
}

const goBack = () => {
  router.push(`/dramas/${dramaId}`)
}

const saveSettings = async () => {
  saving.value = true
  try {
    await dramaAPI.update(dramaId, form)
    ElMessage.success('设置保存成功')
    setCache(getDramaCacheKey(), { ...form, id: dramaId })
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const deleteProject = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要删除此项目吗？此操作不可恢复！',
      '警告',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    deleting.value = true
    await dramaAPI.delete(dramaId)
    ElMessage.success('项目已删除')
    router.push('/dramas')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  } finally {
    deleting.value = false
  }
}

onMounted(async () => {
  const hasCache = hydrateDramaFromCache()
  await loadDrama(!hasCache)
})

onBeforeUnmount(() => {
  if (loadingTimer) {
    window.clearTimeout(loadingTimer)
    loadingTimer = null
  }
})
</script>

<style scoped>
.drama-settings-container {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.main-card {
  margin-top: 20px;
}

.danger-zone {
  margin-top: 20px;
  padding: 20px;
  text-align: center;
}
</style>
