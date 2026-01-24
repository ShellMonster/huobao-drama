export interface BrandSpec {
  id: number
  brand_id: number
  name: string
  description?: string
  allowed_sizes?: string[]
  aspect_ratios?: string[]
  safe_area?: Record<string, any>
  logo_rules?: Record<string, any>
  text_rules?: Record<string, any>
  sort_order?: number
  is_default?: boolean
  is_active?: boolean
  created_at?: string
  updated_at?: string
}

export interface Brand {
  id: number
  name: string
  display_name?: string
  logo_url?: string
  logo_dark_url?: string
  logo_light_url?: string
  description?: string
  tags?: string[]
  is_active?: boolean
  specs?: BrandSpec[]
  created_at?: string
  updated_at?: string
}

export interface BrandListQuery {
  include_inactive?: boolean
  with_specs?: boolean
}

export interface CreateBrandRequest {
  name: string
  display_name?: string
  logo_url?: string
  logo_dark_url?: string
  logo_light_url?: string
  description?: string
  tags?: string[]
  is_active?: boolean
}

export interface UpdateBrandRequest {
  name?: string
  display_name?: string
  logo_url?: string
  logo_dark_url?: string
  logo_light_url?: string
  description?: string
  tags?: string[]
  is_active?: boolean
}

export interface CreateBrandSpecRequest {
  name: string
  description?: string
  allowed_sizes?: string[]
  aspect_ratios?: string[]
  safe_area?: Record<string, any>
  logo_rules?: Record<string, any>
  text_rules?: Record<string, any>
  sort_order?: number
  is_default?: boolean
  is_active?: boolean
}

export interface UpdateBrandSpecRequest {
  name?: string
  description?: string
  allowed_sizes?: string[]
  aspect_ratios?: string[]
  safe_area?: Record<string, any>
  logo_rules?: Record<string, any>
  text_rules?: Record<string, any>
  sort_order?: number
  is_default?: boolean
  is_active?: boolean
}
