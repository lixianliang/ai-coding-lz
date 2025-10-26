import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { Document, Chapter, Role, Scene, UpdateRoleRequest, UpdateSceneRequest } from '@/apis/types'
import * as api from '@/apis/modules/document'

export const useDocumentStore = defineStore('document', () => {
  const documents = ref<Document[]>([])
  const currentDocument = ref<Document | null>(null)
  const currentChapter = ref<Chapter | null>(null)
  const chapters = ref<Chapter[]>([])
  const roles = ref<Role[]>([])
  const scenes = ref<Scene[]>([])
  
  // 计算属性
  const isLoading = computed(() => {
    return currentDocument.value?.status !== 'imgReady'
  })

  // 获取文档列表
  async function fetchDocuments() {
    const res = await api.getDocuments()
    documents.value = res.data.documents || []
  }

  // 获取文档详情
  async function fetchDocument(id: string) {
    const res = await api.getDocument(id)
    currentDocument.value = res.data
  }

  // 获取章节列表
  async function fetchChapters(documentId: string) {
    const res = await api.getChapters(documentId)
    chapters.value = res.data.chapters || []
  }

  // 获取角色列表
  async function fetchRoles(documentId: string) {
    const res = await api.getRoles(documentId)
    roles.value = res.data.roles || []
  }

  // 获取文档场景列表
  async function fetchDocumentScenes(documentId: string) {
    const res = await api.getDocumentScenes(documentId)
    scenes.value = res.data.scenes || []
  }

  // 获取章节详情
  async function fetchChapter(chapterId: string) {
    const res = await api.getChapter(chapterId)
    currentChapter.value = res.data
  }

  // 获取章节场景列表
  async function fetchChapterScenes(chapterId: string) {
    const res = await api.getChapterScenes(chapterId)
    scenes.value = res.data.scenes || []
  }

  // 创建文档
  async function createDocument(name: string, file: File) {
    const res = await api.createDocument({ name, file })
    await fetchDocuments()
    return res.data
  }

  // 删除文档
  async function deleteDocument(id: string) {
    await api.deleteDocument(id)
    await fetchDocuments()
  }

  // 轮询文档状态（用于等待图片生成完成）
  function pollDocumentStatus(documentId: string, timeout = 30 * 60 * 1000) {
    const startTime = Date.now()
    const interval = setInterval(async () => {
      try {
        await fetchDocument(documentId)
        
        if (currentDocument.value?.status === 'imgReady') {
          clearInterval(interval)
          return
        }
        
        if (Date.now() - startTime > timeout) {
          clearInterval(interval)
          console.error('轮询超时')
        }
      } catch (error) {
        console.error('轮询错误:', error)
        clearInterval(interval)
      }
    }, 5000) // 5秒轮询一次
  }

  // 更新角色
  async function updateRole(roleId: string, data: UpdateRoleRequest) {
    const res = await api.updateRole(roleId, data)
    // 更新成功后刷新角色列表
    if (currentDocument.value) {
      await fetchRoles(currentDocument.value.id)
    }
    return res.data
  }

  // 更新场景
  async function updateScene(sceneId: string, data: UpdateSceneRequest) {
    const res = await api.updateScene(sceneId, data)
    // 更新成功后刷新场景列表
    if (currentDocument.value) {
      await fetchDocumentScenes(currentDocument.value.id)
    }
    return res.data
  }

  return {
    documents,
    currentDocument,
    currentChapter,
    chapters,
    roles,
    scenes,
    isLoading,
    fetchDocuments,
    fetchDocument,
    fetchChapter,
    fetchChapters,
    fetchRoles,
    fetchDocumentScenes,
    fetchChapterScenes,
    createDocument,
    deleteDocument,
    pollDocumentStatus,
    updateRole,
    updateScene
  }
})
