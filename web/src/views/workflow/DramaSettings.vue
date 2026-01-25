<template>
  <div class="drama-settings-container">
    <el-page-header @back="goBack" :title="$t('workflow.backToProject')">
      <template #content>
        <h2>{{ $t('projectSettings.title') }}</h2>
      </template>
    </el-page-header>

    <LoadingSection class="main-card" :loading="pageLoading" :text="$t('common.loading')">
      <el-card shadow="never">
        <el-tabs v-model="activeTab">
        <el-tab-pane :label="$t('projectSettings.tabs.basic')" name="basic">
          <el-form :model="form" label-width="100px" style="max-width: 600px">
            <el-form-item :label="$t('projectSettings.form.title')">
              <el-input v-model="form.title" />
            </el-form-item>
            <el-form-item :label="$t('projectSettings.form.description')">
              <el-input v-model="form.description" type="textarea" :rows="4" />
            </el-form-item>
            <el-form-item :label="$t('projectSettings.form.brand')">
              <el-skeleton v-if="brandsLoading" :rows="1" animated />
              <template v-else>
                <el-select
                  v-model="form.brand_id"
                  :placeholder="$t('projectSettings.form.brandPlaceholder')"
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
            <el-form-item :label="$t('projectSettings.form.spec')">
              <el-select
                v-model="form.spec_id"
                :placeholder="$t('projectSettings.form.specPlaceholder')"
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
            <el-form-item :label="$t('projectSettings.form.genre')">
              <el-select v-model="form.genre">
                <el-option :label="$t('genres.urban')" value="都市" />
                <el-option :label="$t('genres.costume')" value="古装" />
                <el-option :label="$t('genres.mystery')" value="悬疑" />
                <el-option :label="$t('genres.romance')" value="爱情" />
                <el-option :label="$t('genres.comedy')" value="喜剧" />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('projectSettings.form.status')">
              <el-select v-model="form.status">
                <el-option :label="$t('projectSettings.status.draft')" value="draft" />
                <el-option :label="$t('projectSettings.status.planning')" value="planning" />
                <el-option :label="$t('projectSettings.status.production')" value="production" />
                <el-option :label="$t('projectSettings.status.completed')" value="completed" />
                <el-option :label="$t('projectSettings.status.archived')" value="archived" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="saving" @click="saveSettings">
                {{ $t('projectSettings.form.save') }}
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane :label="$t('projectSettings.tabs.danger')" name="danger">
          <el-alert
            :title="$t('projectSettings.danger.title')"
            type="warning"
            :description="$t('projectSettings.danger.description')"
            :closable="false"
            show-icon
          />
          <div class="danger-zone">
            <el-button type="danger" :loading="deleting" @click="deleteProject">
              {{ $t('projectSettings.danger.delete') }}
            </el-button>
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
import { brandAPI } from '@/api/brand'
import type { Brand, BrandSpec } from '@/types/brand'
import { LoadingSection } from '@/components/common'
import { getCache, setCache } from '@/utils/cache'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const router = useRouter()
const dramaId = route.params.id as string
const { t } = useI18n()

const activeTab = ref('basic')
const pageLoading = ref(false)
const cacheTTL = 60 * 1000
let loadingTimer: number | null = null
const saving = ref(false)
const deleting = ref(false)
const brandsLoading = ref(false)
const brands = ref<Brand[]>([])
const specs = ref<BrandSpec[]>([])
const form = reactive({
  title: '',
  description: '',
  genre: '',
  status: 'draft' as any,
  brand_id: undefined as number | undefined,
  spec_id: undefined as number | undefined
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
  form.brand_id = drama.brand_id
  form.spec_id = drama.spec_id
  handleBrandChange(form.brand_id, true)
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
    ElMessage.error(error.message || t('common.loadFailed'))
  } finally {
    if (showLoading) {
      stopPageLoading()
    }
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

const handleBrandChange = (value: number | undefined, preserveSpec = false) => {
  if (preserveSpec && brands.value.length === 0) {
    return
  }
  const brand = brands.value.find((item) => item.id === value)
  specs.value = brand?.specs || []
  if (!value || specs.value.length === 0) {
    form.spec_id = undefined
    return
  }
  if (preserveSpec && form.spec_id) {
    const matched = specs.value.find((spec) => spec.id === form.spec_id)
    if (matched) {
      return
    }
  }
  const defaultSpec = specs.value.find((spec) => spec.is_default) || specs.value[0]
  form.spec_id = defaultSpec?.id
}

const goBack = () => {
  router.push(`/dramas/${dramaId}`)
}

const saveSettings = async () => {
  saving.value = true
  try {
    await dramaAPI.update(dramaId, form)
    ElMessage.success(t('projectSettings.messages.saveSuccess'))
    setCache(getDramaCacheKey(), { ...form, id: dramaId })
  } catch (error: any) {
    ElMessage.error(error.message || t('common.saveFailed'))
  } finally {
    saving.value = false
  }
}

const deleteProject = async () => {
  try {
    await ElMessageBox.confirm(
      t('projectSettings.messages.deleteConfirm'),
      t('projectSettings.danger.title'),
      {
        confirmButtonText: t('projectSettings.messages.deleteConfirmButton'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
      }
    )
    
    deleting.value = true
    await dramaAPI.delete(dramaId)
    ElMessage.success(t('projectSettings.messages.deleteSuccess'))
    router.push('/dramas')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || t('common.deleteFailed'))
    }
  } finally {
    deleting.value = false
  }
}

onMounted(async () => {
  const hasCache = hydrateDramaFromCache()
  await loadBrands()
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
