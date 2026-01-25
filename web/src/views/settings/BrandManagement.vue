<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false" :show-logo="false">
        <template #left>
          <div class="page-title">
            <h1>{{ $t('brandManagement.title') }}</h1>
            <span class="subtitle">{{ $t('brandManagement.subtitle') }}</span>
          </div>
        </template>
        <template #right>
          <el-button type="primary" class="header-btn primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            <span class="btn-text">{{ $t('brandManagement.create') }}</span>
          </el-button>
        </template>
      </AppHeader>

      <LoadingSection :loading="loading">
        <EmptyState
          v-if="!loading && brands.length === 0"
          :title="$t('brandManagement.empty.title')"
          :description="$t('brandManagement.empty.description')"
          :icon="Star"
        >
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            {{ $t('brandManagement.create') }}
          </el-button>
        </EmptyState>

        <el-table
          v-else
          :data="brands"
          border
          stripe
          class="brand-table"
        >
          <el-table-column prop="name" :label="$t('brandManagement.table.name')" min-width="200" />
          <el-table-column prop="display_name" :label="$t('brandManagement.table.displayName')" min-width="180" />
          <el-table-column :label="$t('brandManagement.table.status')" width="120">
            <template #default="{ row }">
              <el-tag :type="row.is_active === false ? 'warning' : 'success'">
                {{ row.is_active === false ? $t('brandManagement.status.inactive') : $t('brandManagement.status.active') }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="$t('brandManagement.table.specCount')" width="100">
            <template #default="{ row }">
              {{ row.specs?.length || 0 }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('brandManagement.table.actions')" width="240" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="openSpecDialog(row)">{{ $t('brandManagement.actions.manageSpecs') }}</el-button>
              <el-button size="small" type="primary" link @click="openEditDialog(row)">{{ $t('common.edit') }}</el-button>
              <el-popconfirm
                :title="$t('brandManagement.messages.deleteConfirm')"
                :confirm-button-text="$t('common.delete')"
                :cancel-button-text="$t('common.cancel')"
                @confirm="removeBrand(row)"
              >
                <template #reference>
                  <el-button size="small" type="danger" link>{{ $t('common.delete') }}</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </LoadingSection>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('brandManagement.dialog.editTitle') : $t('brandManagement.dialog.createTitle')"
      width="720px"
      :close-on-click-modal="false"
      destroy-on-close
      class="brand-dialog"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item :label="$t('brandManagement.form.name')" prop="name">
          <el-input v-model="form.name" :placeholder="$t('brandManagement.form.namePlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('brandManagement.form.displayName')">
          <el-input v-model="form.display_name" :placeholder="$t('brandManagement.form.displayNamePlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('brandManagement.form.description')">
          <el-input v-model="form.description" type="textarea" :rows="3" :placeholder="$t('brandManagement.form.descriptionPlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('brandManagement.form.logoUrl')">
          <el-input v-model="form.logo_url" :placeholder="$t('brandManagement.form.logoUrlPlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('brandManagement.form.logoDarkUrl')">
          <el-input v-model="form.logo_dark_url" :placeholder="$t('brandManagement.form.logoDarkUrlPlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('brandManagement.form.logoLightUrl')">
          <el-input v-model="form.logo_light_url" :placeholder="$t('brandManagement.form.logoLightUrlPlaceholder')" />
        </el-form-item>

        <el-form-item :label="$t('brandManagement.form.active')">
          <el-switch v-model="form.is_active" />
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

    <el-dialog
      v-model="specDialogVisible"
      :title="currentBrand
        ? $t('brandManagement.specDialog.titleWithName', { name: currentBrand.display_name || currentBrand.name })
        : $t('brandManagement.specDialog.title')"
      width="860px"
      :close-on-click-modal="false"
      destroy-on-close
      class="spec-dialog"
    >
      <LoadingSection :loading="specLoading">
        <el-table :data="specs" border stripe class="spec-table">
          <el-table-column prop="name" :label="$t('brandManagement.specDialog.table.name')" min-width="200" />
          <el-table-column :label="$t('brandManagement.specDialog.table.default')" width="80">
            <template #default="{ row }">
              <el-tag v-if="row.is_default" type="success" size="small">{{ $t('brandManagement.specDialog.default') }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="$t('brandManagement.specDialog.table.status')" width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_active === false ? 'warning' : 'info'" size="small">
                {{ row.is_active === false ? $t('brandManagement.status.inactive') : $t('brandManagement.status.active') }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="$t('brandManagement.specDialog.table.actions')" width="200" fixed="right">
            <template #default="{ row }">
              <el-button size="small" link type="primary" @click="editSpec(row)">{{ $t('common.edit') }}</el-button>
              <el-popconfirm
                :title="$t('brandManagement.specDialog.messages.deleteConfirm')"
                :confirm-button-text="$t('common.delete')"
                :cancel-button-text="$t('common.cancel')"
                @confirm="removeSpec(row)"
              >
                <template #reference>
                  <el-button size="small" link type="danger">{{ $t('common.delete') }}</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>

        <div class="spec-form-title">{{ $t('brandManagement.specDialog.formTitle') }}</div>
        <el-form ref="specFormRef" :model="specForm" :rules="specRules" label-width="110px">
          <el-form-item :label="$t('brandManagement.specDialog.form.name')" prop="name">
            <el-input v-model="specForm.name" :placeholder="$t('brandManagement.specDialog.form.namePlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('brandManagement.specDialog.form.description')">
            <el-input v-model="specForm.description" type="textarea" :rows="2" :placeholder="$t('brandManagement.specDialog.form.descriptionPlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('brandManagement.specDialog.form.allowedSizes')">
            <el-input v-model="specForm.allowed_sizes_text" :placeholder="$t('brandManagement.specDialog.form.allowedSizesPlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('brandManagement.specDialog.form.aspectRatios')">
            <el-input v-model="specForm.aspect_ratios_text" :placeholder="$t('brandManagement.specDialog.form.aspectRatiosPlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('brandManagement.specDialog.form.safeArea')">
            <el-input v-model="specForm.safe_area_json" type="textarea" :rows="2" :placeholder="$t('brandManagement.specDialog.form.safeAreaPlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('brandManagement.specDialog.form.logoRules')">
            <el-input v-model="specForm.logo_rules_json" type="textarea" :rows="2" :placeholder="$t('brandManagement.specDialog.form.logoRulesPlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('brandManagement.specDialog.form.textRules')">
            <el-input v-model="specForm.text_rules_json" type="textarea" :rows="2" :placeholder="$t('brandManagement.specDialog.form.textRulesPlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('brandManagement.specDialog.form.sortOrder')">
            <el-input-number v-model="specForm.sort_order" :min="0" :max="9999" />
          </el-form-item>
          <el-form-item :label="$t('brandManagement.specDialog.form.active')">
            <el-switch v-model="specForm.is_active" />
          </el-form-item>
          <el-form-item :label="$t('brandManagement.specDialog.form.isDefault')">
            <el-switch v-model="specForm.is_default" />
          </el-form-item>
        </el-form>
      </LoadingSection>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="specDialogVisible = false">{{ $t('common.close') }}</el-button>
          <el-button type="primary" :loading="specSaving" @click="submitSpec">
            {{ editingSpec ? $t('brandManagement.specDialog.actions.save') : $t('brandManagement.specDialog.actions.create') }}
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
import { useI18n } from 'vue-i18n'

const loading = ref(false)
const saving = ref(false)
const brands = ref<Brand[]>([])
const dialogVisible = ref(false)
const editingBrand = ref<Brand | null>(null)
const { t } = useI18n()

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
  name: [{ required: true, message: t('brandManagement.validation.nameRequired'), trigger: 'blur' }]
}

const isEdit = computed(() => !!editingBrand.value)

const loadBrands = async () => {
  if (loading.value) return
  loading.value = true
  try {
    brands.value = await brandAPI.list({ include_inactive: true, with_specs: true })
  } catch (error: any) {
    ElMessage.error(error?.message || t('brandManagement.messages.loadFailed'))
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
        ElMessage.success(t('message.updateSuccess'))
      } else {
        await brandAPI.create(form)
        ElMessage.success(t('message.createSuccess'))
      }
      dialogVisible.value = false
      await loadBrands()
    } catch (error: any) {
      ElMessage.error(error?.message || t('common.saveFailed'))
    } finally {
      saving.value = false
    }
  })
}

const removeBrand = async (brand: Brand) => {
  try {
    await brandAPI.delete(brand.id)
    ElMessage.success(t('common.deleteSuccess'))
    await loadBrands()
  } catch (error: any) {
    ElMessage.error(error?.message || t('common.deleteFailed'))
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
  name: [{ required: true, message: t('brandManagement.specDialog.validation.nameRequired'), trigger: 'blur' }]
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
    ElMessage.error(error?.message || t('brandManagement.specDialog.messages.loadFailed'))
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
      ElMessage.error(t('brandManagement.specDialog.messages.jsonInvalid'))
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
        ElMessage.success(t('brandManagement.specDialog.messages.updateSuccess'))
      } else {
        await brandAPI.createSpec(currentBrand.value.id, payload)
        ElMessage.success(t('brandManagement.specDialog.messages.createSuccess'))
      }

      editingSpec.value = null
      resetSpecForm()
      await loadSpecs(currentBrand.value.id)
      await loadBrands()
    } catch (error: any) {
      ElMessage.error(error?.message || t('brandManagement.specDialog.messages.saveFailed'))
    } finally {
      specSaving.value = false
    }
  })
}

const removeSpec = async (spec: BrandSpec) => {
  if (!currentBrand.value) return
  try {
    await brandAPI.deleteSpec(currentBrand.value.id, spec.id)
    ElMessage.success(t('brandManagement.specDialog.messages.deleteSuccess'))
    await loadSpecs(currentBrand.value.id)
    await loadBrands()
  } catch (error: any) {
    ElMessage.error(error?.message || t('brandManagement.specDialog.messages.deleteFailed'))
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
