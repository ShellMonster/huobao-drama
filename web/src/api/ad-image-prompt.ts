import request from '@/utils/request'

export interface GenerateAdTextPromptRequest {
  drama_id: string
  brand_id?: number
  spec_id?: number
  prompt: string
  count?: number
}

export interface GenerateAdImagePromptRequest {
  drama_id: string
  brand_id?: number
  spec_id?: number
  image_url: string
  count?: number
}

export interface GetLatestAdPromptRequest {
  drama_id: string
  brand_id?: number
  spec_id?: number
}

export interface AdPromptItem {
  id: number
  prompt: string
  sort_order?: number
}

export const adPromptAPI = {
  getLatest(params: GetLatestAdPromptRequest) {
    return request.get<{ text_prompts: AdPromptItem[]; image_prompts: AdPromptItem[] }>(
      '/ad-image-prompts/latest',
      { params }
    )
  },
  generateText(data: GenerateAdTextPromptRequest) {
    return request.post<{ prompts: AdPromptItem[] }>('/ad-image-prompts/text', data)
  },
  generateFromImage(data: GenerateAdImagePromptRequest) {
    return request.post<{ prompts: AdPromptItem[] }>('/ad-image-prompts/image', data)
  },
  deleteItem(id: number) {
    return request.delete(`/ad-image-prompts/items/${id}`)
  }
}
