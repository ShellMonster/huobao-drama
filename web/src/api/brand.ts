import request from '@/utils/request'
import type {
  Brand,
  BrandListQuery,
  CreateBrandRequest,
  UpdateBrandRequest,
  BrandSpec,
  CreateBrandSpecRequest,
  UpdateBrandSpecRequest
} from '@/types/brand'

export const brandAPI = {
  list(params?: BrandListQuery) {
    return request.get<Brand[]>('/brands', { params })
  },
  get(id: number | string) {
    return request.get<Brand>(`/brands/${id}`)
  },
  create(data: CreateBrandRequest) {
    return request.post<Brand>('/brands', data)
  },
  update(id: number | string, data: UpdateBrandRequest) {
    return request.put<Brand>(`/brands/${id}`, data)
  },
  delete(id: number | string) {
    return request.delete(`/brands/${id}`)
  },
  listSpecs(brandId: number | string, includeInactive = false) {
    return request.get<BrandSpec[]>(`/brands/${brandId}/specs`, {
      params: { include_inactive: includeInactive }
    })
  },
  createSpec(brandId: number | string, data: CreateBrandSpecRequest) {
    return request.post<BrandSpec>(`/brands/${brandId}/specs`, data)
  },
  updateSpec(brandId: number | string, specId: number | string, data: UpdateBrandSpecRequest) {
    return request.put<BrandSpec>(`/brands/${brandId}/specs/${specId}`, data)
  },
  deleteSpec(brandId: number | string, specId: number | string) {
    return request.delete(`/brands/${brandId}/specs/${specId}`)
  }
}
