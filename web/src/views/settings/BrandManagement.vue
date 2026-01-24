<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false" :show-logo="false">
        <template #left>
          <div class="page-title">
            <h1>品牌管理</h1>
            <span class="subtitle">维护广告主信息与素材规范</span>
          </div>
        </template>
        <template #right>
          <el-button type="primary" class="header-btn primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            <span class="btn-text">新增品牌</span>
          </el-button>
        </template>
      </AppHeader>

      <LoadingSection :loading="loading">
        <EmptyState
          v-if="!loading && brands.length === 0"
          title="暂无品牌"
          description="请先新增广告主信息"
          :icon="Star"
        >
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            新增品牌
          </el-button>
        </EmptyState>

        <el-table
          v-else
          :data="brands"
          border
          stripe
          class="brand-table"
        >
          <el-table-column prop="name" label="品牌名称" min-width="200" />
          <el-table-column prop="display_name" label="展示名称" min-width="180" />
          <el-table-column label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="row.is_active === false ? 'warning' : 'success'">
                {{ row.is_active === false ? '停用' : '启用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="规范数" width="100">
            <template #default="{ row }">
              {{ row.specs?.length || 0 }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="240" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="openSpecDialog(row)">管理规范</el-button>
              <el-button size="small" type="primary" link @click="openEditDialog(row)">编辑</el-button>
              <el-popconfirm
                title="确定删除该品牌？"
                confirm-button-text="删除"
                cancel-button-text="取消"
                @confirm="removeBrand(row)"
              >
                <template #reference>
                  <el-button size="small" type="danger" link>删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </LoadingSection>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑品牌' : '新增品牌'"
      width="720px"
      :close-on-click-modal="false"
      destroy-on-close
      class="brand-dialog"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="品牌名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入品牌名称" />
        </el-form-item>

        <el-form-item label="展示名称">
          <el-input v-model="form.display_name" placeholder="可选，用于对外展示" />
        </el-form-item>

        <el-form-item label="品牌描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="可选，品牌说明" />
        </el-form-item>

        <el-form-item label="Logo URL">
          <el-input v-model="form.logo_url" placeholder="品牌 Logo URL" />
        </el-form-item>

        <el-form-item label="深色 Logo">
          <el-input v-model="form.logo_dark_url" placeholder="深色背景 Logo URL" />
        </el-form-item>

        <el-form-item label="浅色 Logo">
          <el-input v-model="form.logo_light_url" placeholder="浅色背景 Logo URL" />
        </el-form-item>

        <el-form-item label="启用状态">
          <el-switch v-model="form.is_active" />
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="saving" @click="submitForm">
            {{ isEdit ? '保存' : '创建' }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="specDialogVisible"
      :title="currentBrand ? `品牌规范 - ${currentBrand.display_name || currentBrand.name}` : '品牌规范'"
      width="860px"
      :close-on-click-modal="false"
      destroy-on-close
      class="spec-dialog"
    >
      <LoadingSection :loading="specLoading">
        <el-table :data="specs" border stripe class="spec-table">
          <el-table-column prop="name" label="规范名称" min-width="200" />
          <el-table-column label="默认" width="80">
            <template #default="{ row }">
              <el-tag v-if="row.is_default" type="success" size="small">默认</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_active === false ? 'warning' : 'info'" size="small">
                {{ row.is_active === false ? '停用' : '启用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button size="small" link type="primary" @click="editSpec(row)">编辑</el-button>
              <el-popconfirm
                title="确定删除该规范？"
                confirm-button-text="删除"
                cancel-button-text="取消"
                @confirm="removeSpec(row)"
              >
                <template #reference>
                  <el-button size="small" link type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>

        <div class="spec-form-title">新增/编辑规范</div>
        <el-form ref="specFormRef" :model="specForm" :rules="specRules" label-width="110px">
          <el-form-item label="规范名称" prop="name">
            <el-input v-model="specForm.name" placeholder="请输入规范名称" />
          </el-form-item>
          <el-form-item label="描述">
            <el-input v-model="specForm.description" type="textarea" :rows="2" placeholder="可选描述" />
          </el-form-item>
          <el-form-item label="允许尺寸">
            <el-input v-model="specForm.allowed_sizes_text" placeholder="用逗号分隔，如 1080x1080, 1080x1920" />
          </el-form-item>
          <el-form-item label="允许比例">
            <el-input v-model="specForm.aspect_ratios_text" placeholder="用逗号分隔，如 1:1, 9:16" />
          </el-form-item>
          <el-form-item label="安全区(JSON)">
            <el-input v-model="specForm.safe_area_json" type="textarea" :rows="2" placeholder='例如 {"top":40,"bottom":40}' />
          </el-form-item>
          <el-form-item label="Logo规则(JSON)">
            <el-input v-model="specForm.logo_rules_json" type="textarea" :rows="2" placeholder='例如 {"position":"top-right","min_size":0.08}' />
          </el-form-item>
          <el-form-item label="文字规则(JSON)">
            <el-input v-model="specForm.text_rules_json" type="textarea" :rows="2" placeholder='例如 {"font":"Inter","min_contrast":4.5}' />
          </el-form-item>
          <el-form-item label="排序权重">
            <el-input-number v-model="specForm.sort_order" :min="0" :max="9999" />
          </el-form-item>
          <el-form-item label="启用状态">
            <el-switch v-model="specForm.is_active" />
          </el-form-item>
          <el-form-item label="设为默认">
            <el-switch v-model="specForm.is_default" />
          </el-form-item>
        </el-form>
      </LoadingSection>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="specDialogVisible = false">关闭</el-button>
          <el-button type="primary" :loading="specSaving" @click="submitSpec">
            {{ editingSpec ? '保存规范' : '创建规范' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Star } from '@element-plus/icons-vue'
import { AppHeader, EmptyState, LoadingSection } from '@/components/common'
import { brandAPI } from '@/api/brand'
import type { Brand, BrandSpec, CreateBrandRequest } from '@/types/brand'

const loading = ref(false)
const saving = ref(false)
const brands = ref<Brand[]>([])
const dialogVisible = ref(false)
const editingBrand = ref<Brand | null>(null)

const formRef = ref<FormInstance>()
const form = reactive<CreateBrandRequest>({
  name: '',
  display_name: '',
  description: '',
  logo_url: '',
  logo_dark_url: '',
  logo_light_url: '',
  is_active: true
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入品牌名称', trigger: 'blur' }]
}

const isEdit = computed(() => !!editingBrand.value)

const loadBrands = async () => {
  if (loading.value) return
  loading.value = true
  try {
    brands.value = await brandAPI.list({ include_inactive: true, with_specs: true })
  } catch (error: any) {
    ElMessage.error(error?.message || '加载品牌失败')
    brands.value = []
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.name = ''
  form.display_name = ''
  form.description = ''
  form.logo_url = ''
  form.logo_dark_url = ''
  form.logo_light_url = ''
  form.is_active = true
}

const openCreateDialog = () => {
  editingBrand.value = null
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (brand: Brand) => {
  editingBrand.value = brand
  form.name = brand.name
  form.display_name = brand.display_name || ''
  form.description = brand.description || ''
  form.logo_url = brand.logo_url || ''
  form.logo_dark_url = brand.logo_dark_url || ''
  form.logo_light_url = brand.logo_light_url || ''
  form.is_active = brand.is_active !== false
  dialogVisible.value = true
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editingBrand.value) {
        await brandAPI.update(editingBrand.value.id, { ...form })
        ElMessage.success('更新成功')
      } else {
        await brandAPI.create(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await loadBrands()
    } catch (error: any) {
      ElMessage.error(error?.message || '保存失败')
    } finally {
      saving.value = false
    }
  })
}

const removeBrand = async (brand: Brand) => {
  try {
    await brandAPI.delete(brand.id)
    ElMessage.success('删除成功')
    await loadBrands()
  } catch (error: any) {
    ElMessage.error(error?.message || '删除失败')
  }
}

// Spec management
const specDialogVisible = ref(false)
const specLoading = ref(false)
const specSaving = ref(false)
const specs = ref<BrandSpec[]>([])
const currentBrand = ref<Brand | null>(null)
const editingSpec = ref<BrandSpec | null>(null)
const specFormRef = ref<FormInstance>()

const specForm = reactive({
  name: '',
  description: '',
  allowed_sizes_text: '',
  aspect_ratios_text: '',
  safe_area_json: '',
  logo_rules_json: '',
  text_rules_json: '',
  sort_order: 0,
  is_default: false,
  is_active: true
})

const specRules: FormRules = {
  name: [{ required: true, message: '请输入规范名称', trigger: 'blur' }]
}

const resetSpecForm = () => {
  specForm.name = ''
  specForm.description = ''
  specForm.allowed_sizes_text = ''
  specForm.aspect_ratios_text = ''
  specForm.safe_area_json = ''
  specForm.logo_rules_json = ''
  specForm.text_rules_json = ''
  specForm.sort_order = 0
  specForm.is_default = false
  specForm.is_active = true
}

const loadSpecs = async (brandId: number) => {
  if (specLoading.value) return
  specLoading.value = true
  try {
    specs.value = await brandAPI.listSpecs(brandId, true)
  } catch (error: any) {
    ElMessage.error(error?.message || '加载规范失败')
    specs.value = []
  } finally {
    specLoading.value = false
  }
}

const openSpecDialog = async (brand: Brand) => {
  currentBrand.value = brand
  editingSpec.value = null
  resetSpecForm()
  specDialogVisible.value = true
  await loadSpecs(brand.id)
}

const editSpec = (spec: BrandSpec) => {
  editingSpec.value = spec
  specForm.name = spec.name
  specForm.description = spec.description || ''
  specForm.allowed_sizes_text = (spec.allowed_sizes || []).join(', ')
  specForm.aspect_ratios_text = (spec.aspect_ratios || []).join(', ')
  specForm.safe_area_json = spec.safe_area ? JSON.stringify(spec.safe_area) : ''
  specForm.logo_rules_json = spec.logo_rules ? JSON.stringify(spec.logo_rules) : ''
  specForm.text_rules_json = spec.text_rules ? JSON.stringify(spec.text_rules) : ''
  specForm.sort_order = spec.sort_order || 0
  specForm.is_default = spec.is_default === true
  specForm.is_active = spec.is_active !== false
}

const parseJSON = (value: string) => {
  if (!value.trim()) return undefined
  return JSON.parse(value)
}

const parseList = (value: string) => {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

const submitSpec = async () => {
  if (!specFormRef.value || !currentBrand.value) return
  await specFormRef.value.validate(async (valid) => {
    if (!valid) return
    let safeArea
    let logoRules
    let textRules
    try {
      safeArea = parseJSON(specForm.safe_area_json)
      logoRules = parseJSON(specForm.logo_rules_json)
      textRules = parseJSON(specForm.text_rules_json)
    } catch (error: any) {
      ElMessage.error('JSON格式有误，请检查规则字段')
      return
    }

    specSaving.value = true
    try {
      const payload = {
        name: specForm.name,
        description: specForm.description || undefined,
        allowed_sizes: parseList(specForm.allowed_sizes_text),
        aspect_ratios: parseList(specForm.aspect_ratios_text),
        safe_area: safeArea,
        logo_rules: logoRules,
        text_rules: textRules,
        sort_order: specForm.sort_order,
        is_default: specForm.is_default,
        is_active: specForm.is_active
      }

      if (editingSpec.value) {
        await brandAPI.updateSpec(currentBrand.value.id, editingSpec.value.id, payload)
        ElMessage.success('更新规范成功')
      } else {
        await brandAPI.createSpec(currentBrand.value.id, payload)
        ElMessage.success('创建规范成功')
      }

      editingSpec.value = null
      resetSpecForm()
      await loadSpecs(currentBrand.value.id)
      await loadBrands()
    } catch (error: any) {
      ElMessage.error(error?.message || '保存规范失败')
    } finally {
      specSaving.value = false
    }
  })
}

const removeSpec = async (spec: BrandSpec) => {
  if (!currentBrand.value) return
  try {
    await brandAPI.deleteSpec(currentBrand.value.id, spec.id)
    ElMessage.success('删除规范成功')
    await loadSpecs(currentBrand.value.id)
    await loadBrands()
  } catch (error: any) {
    ElMessage.error(error?.message || '删除规范失败')
  }
}

loadBrands()
</script>

<style scoped>
.brand-table {
  margin-top: 16px;
}

.spec-form-title {
  margin: 20px 0 8px;
  font-weight: 600;
  color: var(--text-primary, #222);
}
</style>
