import type { StyleOption } from '@/types/style'
import request from '@/utils/request'

export const styleAPI = {
  list() {
    return request.get<StyleOption[]>('/styles')
  }
}
