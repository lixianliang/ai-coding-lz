import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue')
  },
  {
    path: '/documents',
    name: 'DocumentList',
    component: () => import('../views/DocumentList.vue')
  },
  {
    path: '/documents/:id',
    name: 'DocumentDetail',
    component: () => import('../views/DocumentDetail.vue')
  },
  {
    path: '/documents/:id/scenes',
    name: 'DocumentScenes',
    component: () => import('../views/SceneViewer.vue')
  },
  {
    path: '/chapters/:chapterId/scenes',
    name: 'ChapterScenes',
    component: () => import('../views/ChapterSceneViewer.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
