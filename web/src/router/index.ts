import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    redirect: '/documents'
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
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
