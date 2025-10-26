<template>
  <div class="document-detail-page">
    <el-container>
      <el-header height="60px">
        <div class="header-content">
          <el-button @click="router.back()">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
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

        <el-tabs v-model="activeTab">
          <el-tab-pane label="章节列表" name="chapters">
            <el-table :data="store.chapters" v-loading="loading" stripe>
              <el-table-column prop="index" label="序号" width="80" />
              <el-table-column prop="content" label="内容" show-overflow-tooltip />
              <el-table-column label="场景数" width="100">
                <template #default="{ row }">
                  {{ row.scene_ids?.length || 0 }}
                </template>
              </el-table-column>
            </el-table>
          </el-tab-pane>
          
          <el-tab-pane label="角色列表" name="roles" :disabled="!showRoles">
            <el-table :data="store.roles" v-loading="loading" stripe>
              <el-table-column prop="name" label="名字" width="120" />
              <el-table-column prop="gender" label="性别" width="100" />
              <el-table-column prop="character" label="性格" show-overflow-tooltip />
              <el-table-column prop="appearance" label="外貌" show-overflow-tooltip />
            </el-table>
          </el-tab-pane>
        </el-tabs>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { useDocumentStore } from '@/stores/document'

const route = useRoute()
const router = useRouter()
const store = useDocumentStore()

const loading = ref(false)
const activeTab = ref('chapters')

// 计算属性：是否显示角色
const showRoles = computed(() => {
  const status = store.currentDocument?.status
  return status === 'roleReady' || status === 'sceneReady' || status === 'imgReady'
})

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
  
  .header-content {
    display: flex;
    align-items: center;
    gap: 16px;
    height: 100%;
    
    h2 {
      margin: 0;
    }
  }

  .status-alert {
    margin-bottom: 20px;
  }
}
</style>
