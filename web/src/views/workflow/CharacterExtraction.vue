<template>
  <div class="character-extraction-container">
    <el-page-header @back="goBack" :title="$t('character.backToProject')">
      <template #content>
        <h2>{{ $t('character.title') }}</h2>
      </template>
    </el-page-header>

    <el-card shadow="never" class="main-card">
      <template #header>
        <div class="card-header">
          <h3>{{ $t('character.list') }}</h3>
          <div class="header-actions">
            <el-button @click="addCharacter">
              <el-icon><Plus /></el-icon>
              {{ $t('character.add') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-empty v-if="characters.length === 0" :description="$t('character.empty')" />

      <el-row :gutter="20" v-else>
        <el-col :span="8" v-for="character in characters" :key="character.id">
          <el-card shadow="hover" class="character-card">
            <template #header>
              <div class="character-header">
                <el-avatar :size="60">{{ character.name[0] }}</el-avatar>
                <div class="character-info">
                  <h4>{{ character.name }}</h4>
                  <el-tag size="small">{{ character.role }}</el-tag>
                </div>
              </div>
            </template>

            <div class="character-details">
              <p><strong>{{ $t('character.personality') }}：</strong>{{ character.personality }}</p>
              <p><strong>{{ $t('character.appearance') }}：</strong>{{ character.appearance }}</p>
              <p><strong>{{ $t('character.description') }}：</strong>{{ character.description }}</p>
            </div>

            <template #footer>
              <el-button-group style="width: 100%">
                <el-button size="small" @click="editCharacter(character)">{{ $t('common.edit') }}</el-button>
                <el-button size="small" type="primary" @click="generateCharacterImage(character)">
                  {{ $t('character.generateImage') }}
                </el-button>
              </el-button-group>
            </template>
          </el-card>
        </el-col>
      </el-row>

      <div class="actions" v-if="characters.length > 0">
        <el-button type="success" size="large" @click="goToNextStep">
          {{ $t('character.nextStep') }}
        </el-button>
      </div>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog v-model="editDialogVisible" :title="dialogTitle" width="600px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item :label="$t('character.name')">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item :label="$t('character.role')">
          <el-input v-model="editForm.role" />
        </el-form-item>
        <el-form-item :label="$t('character.personality')">
          <el-input v-model="editForm.personality" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="$t('character.appearance')">
          <el-input v-model="editForm.appearance" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="$t('character.description')">
          <el-input v-model="editForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="dialogSaving" @click="saveCharacter">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { characterLibraryAPI } from '@/api/character-library'
import type { Character } from '@/types/drama'

const route = useRoute()
const router = useRouter()
const { t: $t } = useI18n()
const dramaId = route.params.id as string

const characters = ref<Character[]>([])
const dialogSaving = ref(false)
const editDialogVisible = ref(false)
const editingCharacterId = ref<number | null>(null)
const editForm = reactive({
  name: '',
  role: '',
  personality: '',
  appearance: '',
  description: ''
})

const dialogTitle = computed(() => {
  return editingCharacterId.value ? $t('character.edit') : $t('character.create')
})

const goBack = () => {
  router.push(`/dramas/${dramaId}`)
}

const addCharacter = () => {
  editingCharacterId.value = null
  Object.assign(editForm, {
    name: '',
    role: '',
    personality: '',
    appearance: '',
    description: ''
  })
  editDialogVisible.value = true
}

const loadCharacters = async () => {
  try {
    characters.value = await dramaAPI.getCharacters(dramaId)
  } catch (error: any) {
    ElMessage.error(error.message || '加载角色失败')
  }
}

const editCharacter = (character: Character) => {
  editingCharacterId.value = character.id
  Object.assign(editForm, {
    name: character.name || '',
    role: character.role || '',
    personality: character.personality || '',
    appearance: character.appearance || '',
    description: character.description || ''
  })
  editDialogVisible.value = true
}

const saveCharacter = async () => {
  const name = editForm.name.trim()
  if (!name) {
    ElMessage.warning('请输入角色名称')
    return
  }

  dialogSaving.value = true
  try {
    if (editingCharacterId.value) {
      await characterLibraryAPI.updateCharacter(editingCharacterId.value, {
        name,
        role: editForm.role,
        appearance: editForm.appearance || '',
        personality: editForm.personality || '',
        description: editForm.description || ''
      })
    } else {
      await dramaAPI.saveCharacters(dramaId, [
        {
          name,
          role: editForm.role,
          appearance: editForm.appearance || '',
          personality: editForm.personality || '',
          description: editForm.description || ''
        }
      ])
    }

    await loadCharacters()
    editDialogVisible.value = false
    ElMessage.success('保存成功')
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    dialogSaving.value = false
  }
}

const generateCharacterImage = (character: Character) => {
  router.push(`/dramas/${dramaId}/images/characters?character=${character.id}`)
}

const goToNextStep = () => {
  router.push(`/dramas/${dramaId}/images/characters`)
}

onMounted(() => {
  loadCharacters()
})
</script>

<style scoped>
.character-extraction-container {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.main-card {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
}

.character-card {
  margin-bottom: 20px;
}

.character-header {
  display: flex;
  gap: 16px;
  align-items: center;
}

.character-info h4 {
  margin: 0 0 8px 0;
}

.character-details p {
  margin: 8px 0;
  font-size: 14px;
  color: #606266;
}

.actions {
  margin-top: 30px;
  text-align: center;
}
</style>
