import request from '../request'
import type { BaseResponse, Document, Chapter, CreateDocumentRequest, Role, Scene, UpdateRoleRequest, UpdateSceneRequest } from '../types'

// 文档列表响应结构
interface DocumentsResponse {
  documents: Document[]
}

interface ChaptersResponse {
  chapters: Chapter[]
}

interface RolesResponse {
  roles: Role[]
}

interface ScenesResponse {
  scenes: Scene[]
}

// 文档列表
export function getDocuments() {
  return request.get('/documents') as Promise<BaseResponse<DocumentsResponse>>
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
  return request.get(`/documents/${documentId}/chapters`) as Promise<BaseResponse<ChaptersResponse>>
}

// 章节详情
export function getChapter(id: string) {
  return request.get(`/chapters/${id}`) as Promise<BaseResponse<Chapter>>
}

// 更新章节
export function updateChapter(documentId: string, chapterId: string, data: { content: string }) {
  return request.put(`/documents/${documentId}/chapters/${chapterId}`, data) as Promise<BaseResponse<Chapter>>
}

// 删除章节
export function deleteChapter(documentId: string, chapterId: string) {
  return request.delete(`/documents/${documentId}/chapters/${chapterId}`) as Promise<BaseResponse<void>>
}

// 角色列表
export function getRoles(documentId: string) {
  return request.get(`/documents/${documentId}/roles`) as Promise<BaseResponse<RolesResponse>>
}

// 文档场景列表
export function getDocumentScenes(documentId: string) {
  return request.get(`/documents/${documentId}/scenes`) as Promise<BaseResponse<ScenesResponse>>
}

// 章节场景列表
export function getChapterScenes(chapterId: string) {
  return request.get(`/chapters/${chapterId}/scenes`) as Promise<BaseResponse<ScenesResponse>>
}

// 更新角色
export function updateRole(roleId: string, data: UpdateRoleRequest) {
  return request.put(`/roles/${roleId}`, data) as Promise<BaseResponse<Role>>
}

// 更新场景
export function updateScene(sceneId: string, data: UpdateSceneRequest) {
  return request.put(`/scenes/${sceneId}`, data) as Promise<BaseResponse<Scene>>
}

// 删除场景
export function deleteScene(sceneId: string) {
  return request.delete(`/scenes/${sceneId}`) as Promise<BaseResponse<void>>
}
