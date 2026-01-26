import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'DramaList',
    component: () => import('../views/drama/DramaList.vue'),
    meta: { sideNav: true }
  },
  {
    path: '/dramas/create',
    name: 'DramaCreate',
    component: () => import('../views/drama/DramaCreate.vue'),
    meta: { sideNav: true }
  },
  {
    path: '/dramas/:id/workspace',
    name: 'DramaWorkspace',
    redirect: (to) => ({
      name: 'DramaManagement',
      params: { id: to.params.id },
      query: to.query,
      hash: to.hash
    })
  },
  {
    path: '/dramas/:id',
    name: 'DramaManagement',
    redirect: (to) => {
      const section = typeof to.query.section === 'string' ? to.query.section : ''
      const tab = typeof to.query.tab === 'string' ? to.query.tab : ''
      const query = { ...to.query }
      delete (query as { section?: string }).section
      delete (query as { tab?: string }).tab

      if (section === 'images') {
        const mode = typeof to.query.mode === 'string' && ['text', 'image'].includes(to.query.mode)
          ? to.query.mode
          : 'text'
        delete (query as { mode?: string }).mode
        return { name: 'DramaManagementImagesMode', params: { id: to.params.id, mode }, query, hash: to.hash }
      }
      if (section === 'videos') {
        return { name: 'DramaManagementVideos', params: { id: to.params.id }, query, hash: to.hash }
      }

      const normalizedTab = ['overview', 'episodes', 'characters', 'scenes'].includes(tab)
        ? tab
        : 'overview'
      return {
        name: 'DramaManagementAdvancedTab',
        params: { id: to.params.id, tab: normalizedTab },
        query,
        hash: to.hash
      }
    }
  },
  {
    path: '/dramas/:id/images',
    name: 'DramaManagementImages',
    redirect: (to) => ({
      name: 'DramaManagementImagesMode',
      params: { id: to.params.id, mode: 'text' },
      query: to.query,
      hash: to.hash
    })
  },
  {
    path: '/dramas/:id/images/:mode(text|image)',
    name: 'DramaManagementImagesMode',
    component: () => import('../views/drama/DramaManagement.vue'),
    meta: { sideNav: true, section: 'images' }
  },
  {
    path: '/dramas/:id/videos',
    name: 'DramaManagementVideos',
    component: () => import('../views/drama/DramaManagement.vue'),
    meta: { sideNav: true, section: 'videos' }
  },
  {
    path: '/dramas/:id/advanced',
    name: 'DramaManagementAdvanced',
    redirect: (to) => {
      const tab = typeof to.query.tab === 'string' ? to.query.tab : ''
      const query = { ...to.query }
      delete (query as { section?: string }).section
      delete (query as { tab?: string }).tab
      const normalizedTab = ['overview', 'episodes', 'characters', 'scenes'].includes(tab)
        ? tab
        : 'overview'
      return {
        name: 'DramaManagementAdvancedTab',
        params: { id: to.params.id, tab: normalizedTab },
        query,
        hash: to.hash
      }
    }
  },
  {
    path: '/dramas/:id/advanced/:tab',
    name: 'DramaManagementAdvancedTab',
    component: () => import('../views/drama/DramaManagement.vue'),
    meta: { sideNav: true, section: 'advanced' }
  },
  {
    path: '/dramas/:id/episode/:episodeNumber',
    name: 'EpisodeWorkflowNew',
    component: () => import('../views/drama/EpisodeWorkflow.vue')
  },
  {
    path: '/dramas/:id/characters',
    name: 'CharacterExtraction',
    component: () => import('../views/workflow/CharacterExtraction.vue')
  },
  {
    path: '/dramas/:id/images/characters',
    name: 'CharacterImages',
    component: () => import('../views/workflow/CharacterImages.vue')
  },
  {
    path: '/dramas/:id/settings',
    name: 'DramaSettings',
    component: () => import('../views/workflow/DramaSettings.vue')
  },
  {
    path: '/episodes/:id/edit',
    name: 'ScriptEdit',
    component: () => import('../views/script/ScriptEdit.vue')
  },
  {
    path: '/episodes/:id/storyboard',
    name: 'StoryboardEdit',
    component: () => import('../views/storyboard/StoryboardEdit.vue')
  },
  {
    path: '/episodes/:id/generate',
    name: 'Generation',
    component: () => import('../views/generation/ImageGeneration.vue')
  },
  {
    path: '/timeline/:id',
    name: 'TimelineEditor',
    component: () => import('../views/editor/TimelineEditor.vue')
  },
  {
    path: '/dramas/:dramaId/episode/:episodeNumber/professional',
    name: 'ProfessionalEditor',
    component: () => import('../views/drama/ProfessionalEditor.vue')
  },
  {
    path: '/settings/ai-config',
    name: 'AIConfig',
    component: () => import('../views/settings/AIConfig.vue')
  },
  {
    path: '/settings/styles',
    name: 'StyleManagement',
    component: () => import('../views/settings/StyleManagement.vue'),
    meta: { sideNav: true }
  },
  {
    path: '/settings/brands',
    name: 'BrandManagement',
    component: () => import('../views/settings/BrandManagement.vue'),
    meta: { sideNav: true }
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// 开源版本 - 无需认证

export default router
