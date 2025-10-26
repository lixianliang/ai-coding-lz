<template>
  <div class="document-list-page">
    <!-- 动画背景 -->
    <div class="anime-background">
      <div class="particles">
        <div v-for="i in 20" :key="i" class="particle" :style="getParticleStyle(i)"></div>
      </div>
      <div class="decoration-stars">
        <div v-for="i in 30" :key="i" class="star" :style="getStarStyle(i)"></div>
      </div>
    </div>

    <el-container>
      <el-header height="80px">
        <div class="header-content">
          <div class="header-left">
            <el-button type="text" class="back-btn" @click="goBack">
              <el-icon><ArrowLeft /></el-icon>
              <span>返回首页</span>
            </el-button>
            <h1 class="page-title">作品管理</h1>
          </div>
          <el-button type="primary" class="upload-btn pulse-animation" @click="showUploadDialog = true">
            <el-icon><Plus /></el-icon>
            上传作品
          </el-button>
        </div>
      </el-header>
      
      <el-main>
        <div v-loading="loading" class="content-area">
          <el-empty v-if="!loading && store.documents.length === 0" description="还没有作品，开始上传吧！" class="empty-state" />
          <div v-else class="works-grid slide-up">
            <div v-for="work in store.documents" :key="work.id" class="work-card">
              <div class="work-cover" @click="handleView(work)">
                <el-icon class="cover-icon"><Document /></el-icon>
                <div class="status-badge">
                  <el-tag :type="getStatusType(work.status)" size="small">
                    {{ getStatusText(work.status) }}
                  </el-tag>
                </div>
              </div>
              <div class="work-info">
                <h3 class="work-title" @click="handleView(work)">{{ work.name }}</h3>
                <div class="work-meta">
                  <div class="work-time">
                    <el-icon><Clock /></el-icon>
                    {{ formatTime(work.created_at) }}
                  </div>
                </div>
                <div class="work-actions">
                  <el-button type="primary" size="small" @click="handleView(work)">章节</el-button>
                  <el-button type="success" size="small" @click="handleViewScenes(work)">场景</el-button>
                  <el-button type="danger" size="small" @click="handleDelete(work)">删除</el-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-main>
    </el-container>

    <!-- 上传对话框 -->
    <el-dialog v-model="showUploadDialog" title="上传作品" width="500px">
      <el-form :model="uploadForm" label-width="80px">
        <el-form-item label="作品名称">
          <el-input v-model="uploadForm.name" placeholder="请输入作品名称" />
        </el-form-item>
        <el-form-item label="选择文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
          >
            <el-button type="primary">选择文件</el-button>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpload" :loading="uploading">
          上传
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, ArrowLeft, Document, Clock } from '@element-plus/icons-vue'
import { useDocumentStore } from '@/stores/document'

const router = useRouter()
const store = useDocumentStore()

const loading = ref(false)
const showUploadDialog = ref(false)
const uploadForm = ref({
  name: '',
  file: null as File | null
})
const uploading = ref(false)
const uploadRef = ref()

// 轮询定时器
let pollInterval: NodeJS.Timeout | null = null

// 返回首页
const goBack = () => {
  router.push('/')
}

// 格式化状态文字
const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    chapterReady: '章节就绪',
    roleReady: '角色提取完成',
    sceneReady: '场景生成完成',
    imgReady: '漫画生成完成'
  }
  return statusMap[status] || status
}

// 获取状态类型
const getStatusType = (status: string): 'info' | 'success' | 'warning' => {
  const typeMap: Record<string, 'info' | 'success' | 'warning'> = {
    chapterReady: 'info',
    roleReady: 'success',
    sceneReady: 'warning',
    imgReady: 'success'
  }
  return typeMap[status] || 'info'
}

// 格式化时间
const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  return date.toLocaleDateString('zh-CN')
}

// 动画背景相关
const getParticleStyle = (index: number) => {
  return {
    left: `${(index * 37) % 100}%`,
    animationDelay: `${index * 0.3}s`,
    animationDuration: `${10 + (index % 10)}s`
  }
}

const getStarStyle = (index: number) => {
  return {
    left: `${(index * 47) % 100}%`,
    top: `${(index * 31) % 100}%`,
    width: `${4 + (index % 4)}px`,
    height: `${4 + (index % 4)}px`,
    animationDelay: `${index * 0.2}s`,
    animationDuration: `${2 + (index % 3)}s`
  }
}

const handleFileChange = (file: any) => {
  uploadForm.value.file = file.raw
  
  // 自动填充作品名称（去掉文件后缀）
  if (file.name) {
    const fileName = file.name
    const lastDotIndex = fileName.lastIndexOf('.')
    const nameWithoutExt = lastDotIndex > 0 
      ? fileName.substring(0, lastDotIndex) 
      : fileName
    uploadForm.value.name = nameWithoutExt
  }
}

const handleUpload = async () => {
  if (!uploadForm.value.name) {
    ElMessage.warning('请输入作品名称')
    return
  }
  if (!uploadForm.value.file) {
    ElMessage.warning('请选择文件')
    return
  }

  uploading.value = true
  try {
    const doc = await store.createDocument(uploadForm.value.name, uploadForm.value.file!)
    ElMessage.success('上传成功')
    showUploadDialog.value = false
    uploadForm.value = { name: '', file: null }
    uploadRef.value?.clearFiles()
    
    // 开始轮询作品状态
    store.pollDocumentStatus(doc.id)
  } catch (error) {
    ElMessage.error('上传失败')
  } finally {
    uploading.value = false
  }
}

const handleView = (doc: any) => {
  router.push(`/documents/${doc.id}`)
}

const handleViewScenes = (doc: any) => {
  router.push(`/documents/${doc.id}/scenes`)
}

const handleDelete = async (doc: any) => {
  try {
    await ElMessageBox.confirm('确定要删除这个作品吗？', '提示', {
      type: 'warning'
    })
    await store.deleteDocument(doc.id)
    ElMessage.success('删除成功')
  } catch (error) {
    // 用户取消
  }
}

// 开始轮询作品状态
const startPolling = () => {
  // 清除旧定时器
  if (pollInterval) {
    clearInterval(pollInterval)
  }
  
  // 每5秒轮询一次
  pollInterval = setInterval(async () => {
    await store.fetchDocuments()
  }, 5000)
}

// 停止轮询
const stopPolling = () => {
  if (pollInterval) {
    clearInterval(pollInterval)
    pollInterval = null
  }
}

onMounted(async () => {
  loading.value = true
  try {
    await store.fetchDocuments()
    // 开始轮询
    startPolling()
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  // 组件卸载时停止轮询
  stopPolling()
})
</script>

<style scoped lang="scss">
@use '../styles/variables.scss' as *;

.document-list-page {
  position: relative;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 25%, #f093fb 50%, #4facfe 75%, #00f2fe 100%);
  background-size: 400% 400%;
  animation: gradientShift 15s ease infinite;
  overflow-y: auto;
  
  // 动画背景
  .anime-background {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    overflow: hidden;
    pointer-events: none;
    z-index: 0;
    
    .particles {
      position: absolute;
      width: 100%;
      height: 100%;
      
      .particle {
        position: absolute;
        bottom: -10px;
        width: 10px;
        height: 10px;
        background: rgba(255, 255, 255, 0.6);
        border-radius: 50%;
        animation: particleFloat linear infinite;
      }
    }
    
    .decoration-stars {
      position: absolute;
      width: 100%;
      height: 100%;
      
      .star {
        position: absolute;
        background: white;
        border-radius: 50%;
        animation: twinkle ease-in-out infinite;
      }
    }
  }

  // Header
  :deep(.el-header) {
    position: relative;
    z-index: 1;
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(20px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.2);
    
    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
      height: 100%;
      padding: 0 24px;
      
      .header-left {
        display: flex;
        align-items: center;
        gap: 16px;
        
        .back-btn {
          color: white;
          font-size: 14px;
          
          &:hover {
            background: rgba(255, 255, 255, 0.1);
          }
        }
        
        .page-title {
          margin: 0;
          font-size: 28px;
          font-weight: 700;
          color: white;
          text-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
        }
      }
      
      .upload-btn {
        background: rgba(255, 255, 255, 0.95);
        border: none;
        color: #667eea;
        font-weight: 600;
        padding: 12px 24px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        
        &:hover {
          background: white;
          transform: translateY(-2px);
          box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
        }
      }
      
      .pulse-animation {
        animation: pulse 2s ease-in-out infinite;
        
        &:hover {
          animation: none;
        }
      }
    }
  }

  // Main
  :deep(.el-main) {
    position: relative;
    z-index: 1;
    padding: 40px 24px;
    
    .content-area {
      min-height: calc(100vh - 120px);
      
      .empty-state {
        margin-top: 100px;
        
        :deep(.el-empty__description) {
          color: white;
        }
      }
      
      .works-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
        gap: 24px;
        max-width: 1400px;
        margin: 0 auto;
        
        .work-card {
          background: rgba(255, 255, 255, 0.95);
          backdrop-filter: blur(10px);
          border-radius: $border-radius-lg;
          overflow: hidden;
          transition: all $transition-normal;
          box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
          
          &:hover {
            transform: translateY(-8px);
            box-shadow: 0 8px 30px rgba(0, 0, 0, 0.15);
          }
          
          .work-cover {
            position: relative;
            width: 100%;
            aspect-ratio: 16/9;
            background: $gradient-purple;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            overflow: hidden;
            
            .cover-icon {
              font-size: 64px;
              color: white;
              opacity: 0.9;
              transition: all $transition-normal;
            }
            
            .status-badge {
              position: absolute;
              top: 12px;
              right: 12px;
            }
            
            &:hover .cover-icon {
              transform: scale(1.1);
              opacity: 1;
            }
          }
          
          .work-info {
            padding: 20px;
            
            .work-title {
              font-size: 18px;
              font-weight: 600;
              color: #333;
              margin: 0 0 12px 0;
              cursor: pointer;
              transition: color $transition-fast;
              
              &:hover {
                color: #667eea;
              }
            }
            
            .work-meta {
              margin-bottom: 16px;
              
              .work-time {
                display: flex;
                align-items: center;
                gap: 6px;
                font-size: 13px;
                color: #999;
              }
            }
            
            .work-actions {
              display: flex;
              gap: 8px;
              
              .el-button {
                flex: 1;
              }
            }
          }
        }
      }
    }
  }
}

// 渐变动画
@keyframes gradientShift {
  0% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0% 50%;
  }
}

// 响应式
@media (max-width: 768px) {
  .document-list-page {
    :deep(.el-header) {
      .header-content {
        padding: 0 16px;
        
        .header-left {
          .page-title {
            font-size: 20px;
          }
        }
        
        .upload-btn {
          padding: 8px 16px;
          font-size: 14px;
        }
      }
    }
    
    :deep(.el-main) {
      padding: 20px 16px;
      
      .content-area {
        .works-grid {
          grid-template-columns: 1fr;
          gap: 16px;
        }
      }
    }
  }
}
</style>
