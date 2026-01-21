import request from '../utils/request'
import type { Scene } from '../types/drama'

export const sceneAPI = {
  create(data: {
    drama_id: number | string
    episode_id?: string
    location: string
    time: string
    prompt?: string
  }) {
    return request.post<Scene>('/scenes', data)
  },

  updateDetails(sceneId: string, data: {
    location?: string
    time?: string
    prompt?: string
  }) {
    return request.put(`/scenes/${sceneId}/details`, data)
  }
}
