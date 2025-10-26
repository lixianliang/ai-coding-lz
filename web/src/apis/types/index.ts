// API 响应基础结构
export interface BaseResponse<T = any> {
  code: number
  message?: string
  reqid: string
  data: T
}

// 文档
export interface Document {
  id: string
  name: string
  file_id: string
  status: 'chapterReady' | 'roleReady' | 'sceneReady' | 'imgReady'
  created_at: string
  updated_at: string
}

// 章节
export interface Chapter {
  id: string
  document_id: string
  index: number
  title: string
  content: string
  scene_ids: string[]
  created_at: string
  updated_at: string
}

// 角色
export interface Role {
  id: string
  document_id: string
  name: string
  gender: string
  character: string
  appearance: string
  created_at: string
  updated_at: string
}

// 场景
export interface Scene {
  id: string
  chapter_id: string
  document_id: string
  index: number
  content: string
  image_url: string
  voice_url: string
  created_at: string
  updated_at: string
}

// 创建文档请求
export interface CreateDocumentRequest {
  name: string
  file: File
}
