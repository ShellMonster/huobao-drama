export interface StyleOption {
  id: number
  key: string
  name: string
  prompt_zh?: string
  prompt_en?: string
  preview_url?: string
  sort_order?: number
  is_active?: boolean
  is_default?: boolean
  is_system?: boolean
}

export interface StyleCreateRequest {
  key?: string
  name: string
  prompt_zh?: string
  prompt_en?: string
  preview_url?: string
  sort_order?: number
  is_active?: boolean
  is_default?: boolean
}

export interface StyleUpdateRequest {
  name?: string
  prompt_zh?: string
  prompt_en?: string
  preview_url?: string
  sort_order?: number
  is_active?: boolean
  is_default?: boolean
}
