<template>
  <div class="document-list-page">
    <el-container>
      <el-header height="60px">
        <div class="header-content">
          <div class="header-left">
            <el-breadcrumb separator="/">
              <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
              <el-breadcrumb-item>作品管理</el-breadcrumb-item>
            </el-breadcrumb>
            <h1 class="page-title">作品管理</h1>
          </div>
          <el-button type="primary" class="upload-btn" @click="showUploadDialog = true">
            <el-icon><Plus /></el-icon>
            上传作品
          </el-button>
        </div>
      </el-header>
      
      <el-main>
        <el-table :data="store.documents" v-loading="loading" stripe>
          <el-table-column prop="name" label="作品名称" width="300" />
          <el-table-column prop="status" label="状态" width="180">
            <template #default="{ row }">
              <el-tag v-if="row.status === 'chapterReady'" type="info">章节就绪</el-tag>
              <el-tag v-else-if="row.status === 'roleReady'" type="success">角色提取完成</el-tag>
              <el-tag v-else-if="row.status === 'sceneReady'" type="warning">场景生成完成</el-tag>
              <el-tag v-else-if="row.status === 'imgReady'" type="success">图片生成完成</el-tag>
              <el-tag v-else type="info">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="200" />
          <el-table-column label="操作" width="250" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">
                查看
              </el-button>
              <el-button type="success" size="small" @click="handleViewScenes(row)">
                查看场景
              </el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
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
import { Plus } from '@element-plus/icons-vue'
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
  min-height: 100vh;
  background: linear-gradient(to bottom, #f8f9fa 0%, #ffffff 100%);
  
  .header-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 100%;
    padding: 0 16px;
    
    .header-left {
      display: flex;
      flex-direction: column;
      gap: 8px;
      
      .page-title {
        margin: 0;
        font-size: 24px;
        font-weight: 600;
        color: #333;
        animation: slideDown 0.5s ease-out;
      }
    }
    
    .upload-btn {
      animation: pulse 2s ease-in-out infinite;
      
      &:hover {
        animation: none;
        transform: scale(1.05);
      }
    }
  }
  
  :deep(.el-table) {
    border-radius: $border-radius-md;
    overflow: hidden;
    box-shadow: $shadow-sm;
    
    .el-table__row {
      transition: all $transition-fast;
      
      &:hover {
        background-color: rgba(255, 107, 157, 0.05) !important;
        transform: scale(1.01);
      }
    }
  }
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}
</style>
