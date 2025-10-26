<template>
  <div class="scene-viewer-page">
    <el-container>
      <el-header height="60px">
        <div class="header-content">
          <el-button @click="router.back()">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
          <h2>{{ store.currentDocument?.name }} - 场景</h2>
        </div>
      </el-header>
      
      <el-main>
        <el-empty v-if="!loading && store.scenes.length === 0" description="暂无场景" />
        
        <div v-else class="scene-list" v-loading="loading">
          <div 
            v-for="scene in store.scenes" 
            :key="scene.id" 
            class="scene-item"
          >
            <div class="scene-header">
              <span class="scene-index">场景 {{ scene.index }}</span>
              <el-tag size="small" v-if="scene.image_url">已完成</el-tag>
              <el-tag size="small" type="info" v-else>处理中</el-tag>
            </div>
            
            <div class="scene-content">
              <div class="scene-text">
                <p>{{ scene.content }}</p>
              </div>
              
              <div class="scene-image" v-if="scene.image_url">
                <img :src="scene.image_url" alt="场景图片" />
              </div>
              
              <div class="scene-loading" v-else>
                <el-icon class="is-loading"><Loading /></el-icon>
                <span>图片生成中...</span>
              </div>
            </div>
          </div>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Loading } from '@element-plus/icons-vue'
import { useDocumentStore } from '@/stores/document'

const route = useRoute()
const router = useRouter()
const store = useDocumentStore()

const loading = ref(false)

onMounted(async () => {
  const id = route.params.id as string
  loading.value = true
  try {
    await Promise.all([
      store.fetchDocument(id),
      store.fetchDocumentScenes(id)
    ])
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.scene-viewer-page {
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
  
  .scene-list {
    max-width: 1200px;
    margin: 0 auto;
  }
  
  .scene-item {
    margin-bottom: 32px;
    padding: 24px;
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
  
  .scene-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    
    .scene-index {
      font-weight: 600;
      font-size: 16px;
    }
  }
  
  .scene-content {
    .scene-text {
      margin-bottom: 16px;
      
      p {
        margin: 0;
        line-height: 1.6;
        color: #666;
      }
    }
    
    .scene-image {
      img {
        width: 100%;
        max-width: 800px;
        border-radius: 8px;
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
      }
    }
    
    .scene-loading {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 40px;
      text-align: center;
      color: #999;
      
      .el-icon {
        font-size: 24px;
      }
    }
  }
}
</style>
