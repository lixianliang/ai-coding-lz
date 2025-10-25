# 前端详细设计文档

## 一、技术架构

### 1.1 技术栈
- Vue 3 + TypeScript + Vite
- Element Plus (UI 组件库)
- Pinia (状态管理)
- Vue Router (路由)
- Axios (HTTP 请求)
- SCSS (样式预处理)

### 1.2 项目结构
```
web/src/
├── apis/              # API 接口层
│   ├── types/         # 类型定义
│   ├── modules/       # 接口模块
│   └── request.ts     # 请求封装
├── components/        # 组件
│   ├── base/          # 基础组件
│   ├── DocumentCard.vue
│   ├── SceneCard.vue
│   └── RoleCard.vue
├── views/             # 页面
│   ├── DocumentList.vue
│   ├── DocumentDetail.vue
│   └── SceneView.vue
├── router/            # 路由
├── stores/            # 状态管理
├── utils/             # 工具函数
├── styles/            # 样式
├── App.vue
└── main.ts
```

## 二、页面设计

### 2.1 文档列表页 (DocumentList.vue)
- 文档卡片列表（grid 布局）
- 上传文档按钮 + 弹窗
- 状态标签展示（chapterReady/sceneReady/imgReady）
- 查看详情、删除操作

### 2.2 文档详情页 (DocumentDetail.vue)
- 文档基本信息
- 进度条展示（当前处理阶段）
- 角色列表（卡片）
- 章节列表（可展开查看场景）
- 自动轮询状态（5秒间隔）

### 2.3 场景查看页 (SceneView.vue)
- 场景图片瀑布流
- 图片点击放大预览
- 场景描述文字

## 三、核心组件

### 3.1 DocumentCard.vue
展示文档信息的卡片组件

**Props:**
- document: Document

**功能:**
- 显示文档名称、状态、时间
- 提供查看、删除操作

### 3.2 StatusTag.vue
状态标签组件

**状态颜色:**
- chapterReady: 蓝色
- sceneReady: 橙色
- imgReady: 绿色

### 3.3 RoleCard.vue
角色信息卡片

**Props:**
- role: Role

**显示:**
- 名字、性别、性格、外貌

### 3.4 SceneCard.vue
场景卡片

**Props:**
- scene: Scene

**功能:**
- 图片展示（懒加载）
- 点击放大预览

## 四、状态管理 (Pinia)

### 4.1 documentStore

```typescript
export const useDocumentStore = defineStore('document', () => {
  // 状态
  const documents = ref<Document[]>([])
  const currentDocument = ref<Document | null>(null)
  const currentChapters = ref<Chapter[]>([])
  const currentScenes = ref<Scene[]>([])
  const currentRoles = ref<Role[]>([])
  
  // 方法
  const fetchDocuments = async () => { }
  const fetchDocumentDetail = async (id: string) => { }
  const uploadDocument = async (name: string, file: File) => { }
  const deleteDocument = async (id: string) => { }
  
  // 轮询
  const startPolling = (docId: string) => {
    // 每 5 秒轮询一次
    // 如果状态为 imgReady 则停止
  }
  const stopPolling = () => { }
  
  return { documents, fetchDocuments, ... }
})
```

### 4.2 轮询机制
- 间隔：5秒
- 触发：文档详情页加载且状态非 imgReady
- 停止：状态变为 imgReady 或页面卸载

## 五、API 层设计

### 5.1 类型定义 (types/document.ts)

```typescript
export interface Document {
  id: string
  name: string
  status: 'chapterReady' | 'sceneReady' | 'imgReady'
  createdAt: string
  updatedAt: string
}

export interface Chapter {
  id: string
  index: number
  documentId: string
  content: string
  sceneIds?: string[]
}

export interface Scene {
  id: string
  chapterId: string
  documentId: string
  index: number
  content: string
  imageUrl: string
  voiceUrl: string
}

export interface Role {
  id: string
  documentId: string
  name: string
  gender: string
  character: string
  appearance: string
}
```

### 5.2 请求封装 (request.ts)

```typescript
import axios from 'axios'
import { ElMessage } from 'element-plus'

const service = axios.create({
  baseURL: '/v1',
  timeout: 30000
})

// 响应拦截器
service.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code === 200) {
      return res.data
    }
    ElMessage.error(res.msg || '请求失败')
    return Promise.reject(new Error(res.msg))
  },
  (error) => {
    ElMessage.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

export default service
```

### 5.3 文档 API (modules/document.ts)

```typescript
import request from '../request'
import type { Document, Chapter, Scene, Role } from '../types'

export const documentApi = {
  list(): Promise<Document[]> {
    return request.get('/documents').then(res => res.documents)
  },
  
  get(id: string): Promise<Document> {
    return request.get(`/documents/${id}`)
  },
  
  upload(name: string, file: File): Promise<Document> {
    const formData = new FormData()
    formData.append('name', name)
    formData.append('file', file)
    return request.post('/documents', formData)
  },
  
  delete(id: string): Promise<void> {
    return request.delete(`/documents/${id}`)
  },
  
  listChapters(documentId: string): Promise<Chapter[]> {
    return request.get(`/documents/${documentId}/chapters`)
      .then(res => res.Chapters)
  },
  
  listRoles(documentId: string): Promise<Role[]> {
    return request.get(`/documents/${documentId}/roles`)
      .then(res => res.roles)
  },
  
  listScenes(documentId: string): Promise<Scene[]> {
    return request.get(`/documents/${documentId}/scenes`)
      .then(res => res.scenes)
  }
}
```

## 六、样式系统

### 6.1 设计令牌 (variables.scss)

```scss
// 颜色
$color-primary: #409eff;
$color-success: #67c23a;
$color-warning: #e6a23c;
$color-danger: #f56c6c;

// 状态颜色
$color-chapter-ready: #409eff;
$color-scene-ready: #e6a23c;
$color-img-ready: #67c23a;

// 间距
$spacing-xs: 4px;
$spacing-sm: 8px;
$spacing-md: 16px;
$spacing-lg: 24px;

// 字体
$font-size-sm: 12px;
$font-size-base: 14px;
$font-size-lg: 16px;
```

### 6.2 组件样式规范

使用 BEM 命名：

```scss
.document-card {
  &__header { }
  &__title { }
  &__status { }
  &__actions { }
}
```

## 七、路由配置

```typescript
const routes = [
  {
    path: '/',
    name: 'DocumentList',
    component: () => import('@/views/DocumentList.vue')
  },
  {
    path: '/documents/:id',
    name: 'DocumentDetail',
    component: () => import('@/views/DocumentDetail.vue')
  },
  {
    path: '/documents/:documentId/scenes',
    name: 'SceneView',
    component: () => import('@/views/SceneView.vue')
  }
]
```

## 八、配置文件

### vite.config.ts
```typescript
export default defineConfig({
  server: {
    proxy: {
      '/v1': 'http://localhost:8000'
    }
  },
  resolve: {
    alias: {
      '@': '/src'
    }
  }
})
```

## 九、开发规范

- 使用 Composition API + `<script setup>`
- 全面使用 TypeScript
- 组件命名：PascalCase
- 文件命名：camelCase
- 样式使用 scoped + BEM
