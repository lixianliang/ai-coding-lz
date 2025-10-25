<template>
  <div class="document-list-page">
    <el-container>
      <el-header height="60px">
        <div class="header-content">
          <h1>文档管理</h1>
          <el-button type="primary" @click="showUploadDialog = true">
            <el-icon><Plus /></el-icon>
            上传文档
          </el-button>
        </div>
      </el-header>
      
      <el-main>
        <el-table :data="store.documents" v-loading="loading" stripe>
          <el-table-column prop="name" label="文档名称" width="300" />
          <el-table-column prop="status" label="状态" width="150">
            <template #default="{ row }">
              <el-tag v-if="row.status === 'chapterReady'" type="info">章节就绪</el-tag>
              <el-tag v-else-if="row.status === 'sceneReady'" type="warning">场景就绪</el-tag>
              <el-tag v-else-if="row.status === 'imgReady'" type="success">图片就绪</el-tag>
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
    <el-dialog v-model="showUploadDialog" title="上传文档" width="500px">
      <el-form :model="uploadForm" label-width="80px">
        <el-form-item label="文档名称">
          <el-input v-model="uploadForm.name" placeholder="请输入文档名称" />
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
import { ref, onMounted } from 'vue'
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

const handleFileChange = (file: any) => {
  uploadForm.value.file = file.raw
}

const handleUpload = async () => {
  if (!uploadForm.value.name) {
    ElMessage.warning('请输入文档名称')
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
    
    // 开始轮询文档状态
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
    await ElMessageBox.confirm('确定要删除这个文档吗？', '提示', {
      type: 'warning'
    })
    await store.deleteDocument(doc.id)
    ElMessage.success('删除成功')
  } catch (error) {
    // 用户取消
  }
}

onMounted(async () => {
  loading.value = true
  try {
    await store.fetchDocuments()
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.document-list-page {
  height: 100vh;
  
  .header-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 100%;
    
    h1 {
      margin: 0;
    }
  }
}
</style>
