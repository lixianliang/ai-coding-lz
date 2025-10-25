import request from '../request'
import type { BaseResponse, Document, Chapter, CreateDocumentRequest } from '../types'

// 文档列表
export function getDocuments() {
  return request.get('/documents') as Promise<BaseResponse<Document[]>>
}

// 创建文档
export function createDocument(data: CreateDocumentRequest) {
  const formData = new FormData()
  formData.append('name', data.name)
  formData.append('file', data.file)
  
  return request.post('/documents', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  }) as Promise<BaseResponse<Document>>
}

// 文档详情
export function getDocument(id: string) {
  return request.get(`/documents/${id}`) as Promise<BaseResponse<Document>>
}

// 删除文档
export function deleteDocument(id: string) {
  return request.delete(`/documents/${id}`) as Promise<BaseResponse<void>>
}

// 章节列表
export function getChapters(documentId: string) {
  return request.get(`/documents/${documentId}/chapters`) as Promise<BaseResponse<Chapter[]>>
}

// 章节详情
export function getChapter(id: string) {
  return request.get(`/chapters/${id}`) as Promise<BaseResponse<Chapter>>
}

// 更新章节
export function updateChapter(id: string, data: Partial<Chapter>) {
  return request.put(`/chapters/${id}`, data) as Promise<BaseResponse<Chapter>>
}

// 删除章节
export function deleteChapter(id: string) {
  return request.delete(`/chapters/${id}`) as Promise<BaseResponse<void>>
}

// 角色列表
export function getRoles(documentId: string) {
  return request.get(`/documents/${documentId}/roles`) as Promise<BaseResponse<any[]>>
}

// 文档场景列表
export function getDocumentScenes(documentId: string) {
  return request.get(`/documents/${documentId}/scenes`) as Promise<BaseResponse<any[]>>
}

// 章节场景列表
export function getChapterScenes(chapterId: string) {
  return request.get(`/chapters/${chapterId}/scenes`) as Promise<BaseResponse<any[]>>
}
