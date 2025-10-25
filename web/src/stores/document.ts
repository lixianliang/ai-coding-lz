import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { Document, Chapter, Role, Scene } from '@/apis/types'
import * as api from '@/apis/modules/document'

export const useDocumentStore = defineStore('document', () => {
  const documents = ref<Document[]>([])
  const currentDocument = ref<Document | null>(null)
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

  return {
    documents,
    currentDocument,
    chapters,
    roles,
    scenes,
    isLoading,
    fetchDocuments,
    fetchDocument,
    fetchChapters,
    fetchRoles,
    fetchDocumentScenes,
    createDocument,
    deleteDocument,
    pollDocumentStatus
  }
})
