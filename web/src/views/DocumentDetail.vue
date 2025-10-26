<template>
  <div class="document-detail-page">
    <el-container>
      <el-header height="60px">
        <div class="header-content">
          <el-button type="primary" :icon="ArrowLeft" @click="router.back()">返回</el-button>
          <h2>{{ store.currentDocument?.name }}</h2>
        </div>
      </el-header>
      
      <el-main>
        <!-- 状态提示 -->
        <el-alert 
          v-if="store.currentDocument?.status === 'chapterReady'"
          title="正在提取角色信息..."
          type="info"
          :closable="false"
          show-icon
          class="status-alert"
        />
        <el-alert 
          v-else-if="store.currentDocument?.status === 'roleReady'"
          title="角色提取完成，正在生成场景..."
          type="success"
          :closable="false"
          show-icon
          class="status-alert"
        />
        <el-alert 
          v-else-if="store.currentDocument?.status === 'sceneReady'"
          title="场景生成完成，正在生成图片..."
          type="warning"
          :closable="false"
          show-icon
          class="status-alert"
        />

        <el-tabs v-model="activeTab" class="detail-tabs">
          <el-tab-pane label="章节列表" name="chapters">
            <div class="chapters-container">
              <el-table :data="store.chapters" v-loading="loading" stripe class="chapters-table">
                <el-table-column prop="index" label="序号" width="80" />
                <el-table-column prop="content" label="内容" show-overflow-tooltip />
                <el-table-column label="场景数" width="100">
                  <template #default="{ row }">
                    <el-tag size="small">{{ row.scene_ids?.length || 0 }}</el-tag>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-tab-pane>
          
          <el-tab-pane label="角色列表" name="roles" :disabled="!showRoles">
            <div class="roles-container">
              <el-table :data="store.roles" v-loading="loading" stripe class="roles-table">
              <el-table-column prop="name" label="名字" width="120" />
              <el-table-column prop="gender" label="性别" width="100" />
              <el-table-column prop="character" label="性格" show-overflow-tooltip />
              <el-table-column prop="appearance" label="外貌" show-overflow-tooltip />
              <el-table-column label="操作" width="100" fixed="right">
                <template #default="{ row }">
                  <el-button type="primary" size="small" :icon="Edit" @click="handleEditRole(row)">
                    编辑
                  </el-button>
                </template>
              </el-table-column>
              </el-table>
            </div>
          </el-tab-pane>
        </el-tabs>
      </el-main>
    </el-container>

    <!-- 编辑角色对话框 -->
    <el-dialog v-model="showEditRoleDialog" title="编辑角色" width="600px">
      <el-form :model="roleForm" label-width="80px" :rules="roleRules" ref="roleFormRef">
        <el-form-item label="名字" prop="name">
          <el-input v-model="roleForm.name" placeholder="请输入角色名字" />
        </el-form-item>
        <el-form-item label="性别" prop="gender">
          <el-select v-model="roleForm.gender" placeholder="请选择性别" style="width: 100%">
            <el-option label="男" value="男" />
            <el-option label="女" value="女" />
            <el-option label="未知" value="未知" />
          </el-select>
        </el-form-item>
        <el-form-item label="性格" prop="character">
          <el-input 
            v-model="roleForm.character" 
            type="textarea" 
            :rows="4" 
            placeholder="请输入角色性格特点"
          />
        </el-form-item>
        <el-form-item label="外貌" prop="appearance">
          <el-input 
            v-model="roleForm.appearance" 
            type="textarea" 
            :rows="4" 
            placeholder="请输入角色外貌描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditRoleDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveRole" :loading="submitting">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Edit } from '@element-plus/icons-vue'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { useDocumentStore } from '@/stores/document'
import { Role, UpdateRoleRequest } from '@/apis/types'

const route = useRoute()
const router = useRouter()
const store = useDocumentStore()

const loading = ref(false)
const activeTab = ref('chapters')
const showEditRoleDialog = ref(false)
const submitting = ref(false)
const roleFormRef = ref<FormInstance>()
const currentEditingRoleId = ref('')

const roleForm = ref<UpdateRoleRequest>({
  name: '',
  gender: '',
  character: '',
  appearance: ''
})

const roleRules: FormRules = {
  name: [{ required: true, message: '请输入角色名字', trigger: 'blur' }],
  gender: [{ required: true, message: '请选择性别', trigger: 'change' }],
  character: [{ required: true, message: '请输入角色性格特点', trigger: 'blur' }],
  appearance: [{ required: true, message: '请输入角色外貌描述', trigger: 'blur' }]
}



// 计算属性：是否显示角色
const showRoles = computed(() => {
  const status = store.currentDocument?.status
  return status === 'roleReady' || status === 'sceneReady' || status === 'imgReady'
})

// 编辑角色
const handleEditRole = (role: Role) => {
  currentEditingRoleId.value = role.id
  roleForm.value = {
    name: role.name,
    gender: role.gender,
    character: role.character,
    appearance: role.appearance
  }
  showEditRoleDialog.value = true
}

// 保存角色
const handleSaveRole = async () => {
  if (!roleFormRef.value) return
  
  await roleFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      await store.updateRole(currentEditingRoleId.value, roleForm.value)
      ElMessage.success('角色更新成功')
      showEditRoleDialog.value = false
    } catch (error) {
      console.error('更新角色失败:', error)
      ElMessage.error('角色更新失败')
    } finally {
      submitting.value = false
    }
  })
}

onMounted(async () => {
  const id = route.params.id as string
  loading.value = true
  try {
    await Promise.all([
      store.fetchDocument(id),
      store.fetchChapters(id),
      store.fetchRoles(id)
    ])
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.document-detail-page {
  height: 100vh;
  
  :deep(.el-header) {
    background: #fff;
    border-bottom: 1px solid #e4e7ed;
    
    .header-content {
      display: flex;
      align-items: center;
      gap: 16px;
      height: 100%;
      padding: 0 20px;
      
      h2 {
        margin: 0;
        font-size: 20px;
        font-weight: 500;
      }
    }
  }

  :deep(.el-main) {
    padding: 20px;

    .status-alert {
      margin-bottom: 20px;
    }

    .detail-tabs {
      :deep(.el-tabs__header) {
        margin-bottom: 20px;
      }
    }
  }
}
</style>
