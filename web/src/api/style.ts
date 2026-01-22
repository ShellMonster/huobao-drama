import type { StyleCreateRequest, StyleOption, StyleUpdateRequest } from '@/types/style'
import request from '@/utils/request'

export const styleAPI = {
  list(params?: { include_inactive?: boolean }) {
    return request.get<StyleOption[]>('/styles', { params })
  },
  create(data: StyleCreateRequest) {
    return request.post<StyleOption>('/styles', data)
  },
  update(id: number, data: StyleUpdateRequest) {
    return request.put<StyleOption>(`/styles/${id}`, data)
  },
  remove(id: number) {
    return request.delete<{ message: string }>(`/styles/${id}`)
  }
}
