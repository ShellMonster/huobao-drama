<template>
  <el-dialog
    v-model="visible"
    :title="$t('video.title')"
    width="700px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <LoadingSection :loading="dataLoading" :text="$t('common.loading')">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item :label="$t('video.dialog.form.drama')" prop="drama_id">
          <el-select
            v-model="form.drama_id"
            :placeholder="$t('video.dialog.form.dramaPlaceholder')"
            :disabled="isDramaLocked"
            @change="onDramaChange"
          >
            <el-option
              v-for="drama in dramas"
              :key="drama.id"
              :label="drama.title"
              :value="drama.id"
            />
          </el-select>
        </el-form-item>

      <el-form-item :label="$t('video.dialog.form.image')" prop="image_gen_id">
        <el-select
          v-model="form.image_gen_id"
          :placeholder="$t('video.dialog.form.imagePlaceholder')"
          clearable
          @change="onImageChange"
        >
          <el-option
            v-for="image in images"
            :key="image.id"
            :label="truncateText(image.prompt, 50)"
            :value="image.id"
          >
            <div class="image-option">
              <img v-if="image.image_url" :src="image.image_url" class="image-thumb" />
              <span>{{ truncateText(image.prompt, 40) }}</span>
            </div>
          </el-option>
        </el-select>
        <div class="form-tip">{{ $t('video.dialog.form.imageUrlTip') }}</div>
      </el-form-item>

      <el-form-item :label="$t('video.dialog.form.imageUrl')" prop="image_url">
        <el-input
          v-model="form.image_url"
          :placeholder="$t('video.dialog.form.imageUrlPlaceholder')"
          :disabled="!!form.image_gen_id"
        />
      </el-form-item>

      <el-form-item :label="$t('video.dialog.form.prompt')" prop="prompt">
        <el-input
          v-model="form.prompt"
          type="textarea"
          :rows="5"
          :placeholder="$t('video.dialog.form.promptPlaceholder')"
          maxlength="2000"
          show-word-limit
        />
      </el-form-item>

      <el-form-item :label="$t('video.dialog.form.service')">
        <el-select v-model="form.provider" :placeholder="$t('video.dialog.form.servicePlaceholder')">
          <el-option :label="$t('video.dialog.providers.doubao')" value="doubao" />
          <el-option :label="$t('video.dialog.providers.runway')" value="runway" />
          <el-option :label="$t('video.dialog.providers.pika')" value="pika" />
        </el-select>
      </el-form-item>

      <el-form-item :label="$t('video.dialog.form.duration')">
        <el-slider
          v-model="form.duration"
          :min="3"
          :max="10"
          :marks="durationMarks"
          show-stops
        />
        <span class="slider-value">{{ $t('common.seconds', { value: form.duration }) }}</span>
      </el-form-item>

      <el-form-item :label="$t('video.dialog.form.aspectRatio')">
        <el-radio-group v-model="form.aspect_ratio">
          <el-radio label="16:9">{{ $t('video.dialog.aspectRatioOptions.wide') }}</el-radio>
          <el-radio label="9:16">{{ $t('video.dialog.aspectRatioOptions.tall') }}</el-radio>
          <el-radio label="1:1">{{ $t('video.dialog.aspectRatioOptions.square') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-collapse>
        <el-collapse-item :title="$t('video.dialog.form.advanced')" name="advanced">
          <el-form-item :label="$t('video.dialog.form.motionIntensity')">
            <el-slider
              v-model="form.motion_level"
              :min="0"
              :max="100"
              :marks="motionMarks"
            />
            <span class="slider-value">{{ form.motion_level }}</span>
          </el-form-item>

          <el-form-item :label="$t('video.dialog.form.cameraMotion')">
            <el-select v-model="form.camera_motion" :placeholder="$t('video.dialog.form.cameraMotionPlaceholder')" clearable>
              <el-option :label="$t('video.dialog.cameraMotionOptions.static')" value="static" />
              <el-option :label="$t('video.dialog.cameraMotionOptions.zoomIn')" value="zoom_in" />
              <el-option :label="$t('video.dialog.cameraMotionOptions.zoomOut')" value="zoom_out" />
              <el-option :label="$t('video.dialog.cameraMotionOptions.panLeft')" value="pan_left" />
              <el-option :label="$t('video.dialog.cameraMotionOptions.panRight')" value="pan_right" />
              <el-option :label="$t('video.dialog.cameraMotionOptions.tiltUp')" value="tilt_up" />
              <el-option :label="$t('video.dialog.cameraMotionOptions.tiltDown')" value="tilt_down" />
              <el-option :label="$t('video.dialog.cameraMotionOptions.orbit')" value="orbit" />
            </el-select>
          </el-form-item>

          <el-form-item :label="$t('video.dialog.form.style')" v-if="form.provider === 'doubao'">
            <el-input v-model="form.style" :placeholder="$t('video.dialog.form.stylePlaceholder')" />
          </el-form-item>

          <el-form-item :label="$t('video.dialog.form.seed')">
            <el-input-number v-model="form.seed" :min="-1" :placeholder="$t('video.dialog.form.seedPlaceholder')" />
            <span class="form-tip">{{ $t('video.dialog.form.seedTip') }}</span>
          </el-form-item>
        </el-collapse-item>
        </el-collapse>
      </el-form>
    </LoadingSection>

    <template #footer>
      <el-button @click="handleClose">{{ $t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="generating" @click="handleGenerate">
        {{ $t('video.generate') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { videoAPI } from '@/api/video'
import { imageAPI } from '@/api/image'
import { dramaAPI } from '@/api/drama'
import type { Drama } from '@/types/drama'
import type { ImageGeneration } from '@/types/image'
import type { GenerateVideoRequest } from '@/types/video'
import { LoadingSection } from '@/components/common'

interface Props {
  modelValue: boolean
  dramaId?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: []
}>()

const { t } = useI18n()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const isDramaLocked = computed(() => Boolean(props.dramaId))

const formRef = ref<FormInstance>()
const generating = ref(false)
const dramas = ref<Drama[]>([])
const images = ref<ImageGeneration[]>([])
const dataLoadingCount = ref(0)
const dataLoading = computed(() => dataLoadingCount.value > 0)

const form = reactive<GenerateVideoRequest & { image_gen_id?: number }>({
  drama_id: props.dramaId || '',
  image_gen_id: undefined,
  image_url: '',
  prompt: '',
  provider: 'doubao',
  duration: 5,
  aspect_ratio: '16:9',
  motion_level: 50,
  camera_motion: undefined,
  style: undefined,
  seed: undefined
})

const rules = computed<FormRules>(() => ({
  drama_id: [
    { required: true, message: t('video.dialog.validation.dramaRequired'), trigger: 'change' }
  ],
  prompt: [
    { required: true, message: t('video.dialog.validation.promptRequired'), trigger: 'blur' },
    { min: 5, message: t('video.dialog.validation.promptMin'), trigger: 'blur' }
  ]
}))

const durationMarks = {
  3: '3s',
  5: '5s',
  7: '7s',
  10: '10s'
}

const motionMarks = computed(() => ({
  0: t('video.dialog.motionMarks.static'),
  50: t('video.dialog.motionMarks.medium'),
  100: t('video.dialog.motionMarks.intense')
}))

watch(() => props.modelValue, (val) => {
  if (val) {
    loadDramas()
    if (props.dramaId) {
      form.drama_id = props.dramaId
      loadImages(props.dramaId)
    }
  }
})

const loadDramas = async () => {
  dataLoadingCount.value += 1
  try {
    const result = await dramaAPI.list({ page: 1, page_size: 100 })
    dramas.value = result.items
  } catch (error: any) {
    console.error('Failed to load dramas:', error)
  } finally {
    dataLoadingCount.value = Math.max(0, dataLoadingCount.value - 1)
  }
}

const loadImages = async (dramaId: string) => {
  dataLoadingCount.value += 1
  try {
    const result = await imageAPI.listImages({
      drama_id: dramaId,
      status: 'completed',
      page: 1,
      page_size: 100
    })
    images.value = result.items
  } catch (error: any) {
    console.error('Failed to load images:', error)
  } finally {
    dataLoadingCount.value = Math.max(0, dataLoadingCount.value - 1)
  }
}

const onDramaChange = (dramaId: string) => {
  form.image_gen_id = undefined
  form.image_url = ''
  images.value = []
  if (dramaId) {
    loadImages(dramaId)
  }
}

const onImageChange = (imageGenId: number | undefined) => {
  if (!imageGenId) {
    form.image_url = ''
    return
  }
  
  const image = images.value.find(img => img.id === imageGenId)
  if (image && image.image_url) {
    form.image_url = image.image_url
    form.prompt = image.prompt
  }
}

const truncateText = (text: string, length: number) => {
  if (text.length <= length) return text
  return text.substring(0, length) + '...'
}

const handleGenerate = async () => {
  console.log('handleGenerate called')
  
  if (!formRef.value) {
    console.error('formRef is null')
    ElMessage.error(t('video.dialog.messages.formInitFailed'))
    return
  }

  try {
    const valid = await formRef.value.validate()
    console.log('Form validation result:', valid)
    
    if (!valid) {
      console.log('Form validation failed')
      return
    }

    generating.value = true
    console.log('Starting video generation...', form)
    
    try {
      if (form.image_gen_id) {
        console.log('Generating from image:', form.image_gen_id)
        await videoAPI.generateFromImage(form.image_gen_id)
      } else {
        const params: GenerateVideoRequest = {
          drama_id: form.drama_id,
          prompt: form.prompt,
          provider: form.provider
        }

        // 判断参考图模式
        if (form.image_url && form.image_url.trim()) {
          params.image_url = form.image_url
          params.reference_mode = 'single'
        } else {
          // 纯文本生成，无参考图
          params.reference_mode = 'none'
        }

        if (form.duration) params.duration = form.duration
        if (form.aspect_ratio) params.aspect_ratio = form.aspect_ratio
        if (form.motion_level !== undefined) params.motion_level = form.motion_level
        if (form.camera_motion) params.camera_motion = form.camera_motion
        if (form.style) params.style = form.style
        if (form.seed && form.seed > 0) params.seed = form.seed

        console.log('Generating video with params:', params)
        await videoAPI.generateVideo(params)
      }
      
      ElMessage.success(t('video.dialog.messages.generateSubmitted'))
      emit('success')
      handleClose()
    } catch (error: any) {
      console.error('Video generation failed:', error)
      ElMessage.error(error.response?.data?.message || error.message || t('common.generateFailed'))
    } finally {
      generating.value = false
    }
  } catch (error: any) {
    console.error('Form validation error:', error)
    ElMessage.warning(t('video.dialog.messages.formIncomplete'))
  }
}

const handleClose = () => {
  visible.value = false
  formRef.value?.resetFields()
}
</script>

<style scoped>
.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: #999;
}

.slider-value {
  margin-left: 12px;
  font-size: 14px;
  font-weight: 500;
  color: #409eff;
}

.image-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.image-thumb {
  width: 40px;
  height: 40px;
  object-fit: cover;
  border-radius: 4px;
}
</style>
