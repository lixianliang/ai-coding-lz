<template>
  <div class="home-page">
    <!-- 动画背景 -->
    <div class="anime-background">
      <div class="particles">
        <div v-for="i in 20" :key="i" class="particle" :style="getParticleStyle(i)"></div>
      </div>
      <div class="decoration-stars">
        <div v-for="i in 30" :key="i" class="star" :style="getStarStyle(i)"></div>
      </div>
    </div>
    
    <!-- 主内容 -->
    <div class="home-content">
      <!-- 标题区域 -->
      <div class="hero-section slide-down">
        <h1 class="anime-title">动漫作品创作平台</h1>
        <p class="subtitle">将小说转化为精彩的连环漫画</p>
      </div>
      
      <!-- 统计信息 -->
      <div class="stats-section">
        <div class="stat-card slide-up delay-1">
          <div class="stat-icon">📚</div>
          <div class="stat-number">{{ totalWorks }}</div>
          <div class="stat-label">总作品数</div>
        </div>
        <div class="stat-card slide-up delay-2">
          <div class="stat-icon">✨</div>
          <div class="stat-number">{{ completedWorks }}</div>
          <div class="stat-label">已完成</div>
        </div>
        <div class="stat-card slide-up delay-3">
          <div class="stat-icon">🎨</div>
          <div class="stat-number">{{ inProgressWorks }}</div>
          <div class="stat-label">进行中</div>
        </div>
      </div>
      
      <!-- 最近作品 -->
      <div class="recent-works slide-up delay-4">
        <h2>最近作品</h2>
        <el-empty v-if="recentWorks.length === 0" description="还没有作品，开始创作吧！" />
        <div v-else class="works-grid">
          <div 
            v-for="work in recentWorks" 
            :key="work.id" 
            class="work-card"
            @click="goToScenes(work.id)"
          >
            <div class="work-cover">
              <el-icon class="cover-icon"><Document /></el-icon>
            </div>
            <div class="work-info">
              <h3 class="work-title">{{ work.name }}</h3>
              <el-tag 
                :type="getStatusType(work.status)" 
                size="small"
              >
                {{ getStatusText(work.status) }}
              </el-tag>
              <div class="work-time">{{ formatTime(work.created_at) }}</div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 进入作品管理按钮 -->
      <el-button 
        type="primary" 
        size="large"
        class="manage-btn slide-up delay-5"
        @click="goToManage"
      >
        <el-icon><Grid /></el-icon>
        进入作品管理
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Document, Grid } from '@element-plus/icons-vue'
import { useDocumentStore } from '@/stores/document'

const router = useRouter()
const store = useDocumentStore()

const loading = ref(false)

// 计算统计数据
const totalWorks = computed(() => store.documents.length)
const completedWorks = computed(() => 
  store.documents.filter(doc => doc.status === 'imgReady').length
)
const inProgressWorks = computed(() => 
  store.documents.filter(doc => doc.status !== 'imgReady').length
)

// 最近 5 个作品
const recentWorks = computed(() => 
  store.documents.slice(0, 5)
)

// 获取粒子样式
const getParticleStyle = (index: number) => {
  const delay = Math.random() * 10
  const duration = 15 + Math.random() * 10
  const left = Math.random() * 100
  return {
    left: `${left}%`,
    animationDelay: `${delay}s`,
    animationDuration: `${duration}s`
  }
}

// 获取星星样式
const getStarStyle = (index: number) => {
  const top = Math.random() * 100
  const left = Math.random() * 100
  const delay = Math.random() * 3
  const duration = 2 + Math.random() * 2
  const size = 2 + Math.random() * 3
  return {
    top: `${top}%`,
    left: `${left}%`,
    width: `${size}px`,
    height: `${size}px`,
    animationDelay: `${delay}s`,
    animationDuration: `${duration}s`
  }
}

// 获取状态类型
const getStatusType = (status: string) => {
  const typeMap: Record<string, any> = {
    chapterReady: 'info',
    roleReady: '',
    sceneReady: 'warning',
    imgReady: 'success'
  }
  return typeMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    chapterReady: '章节就绪',
    roleReady: '角色提取完成',
    sceneReady: '场景生成完成',
    imgReady: '已完成'
  }
  return textMap[status] || status
}

// 格式化时间
const formatTime = (time: string) => {
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days}天前`
  return date.toLocaleDateString('zh-CN')
}

// 跳转到作品场景
const goToScenes = (id: string) => {
  router.push(`/documents/${id}/scenes`)
}

// 跳转到作品管理
const goToManage = () => {
  router.push('/documents')
}

// 页面加载
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
@use '../styles/variables.scss' as *;

.home-page {
  min-height: 100vh;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%);
}

// 动画背景
.anime-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  pointer-events: none;
  
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

// 主内容
.home-content {
  position: relative;
  z-index: 1;
  max-width: 1200px;
  margin: 0 auto;
  padding: 60px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 48px;
}

// 标题区域
.hero-section {
  text-align: center;
  color: white;
  
  .anime-title {
    font-size: 56px;
    font-weight: 800;
    margin: 0 0 16px 0;
    text-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    letter-spacing: 2px;
  }
  
  .subtitle {
    font-size: 20px;
    opacity: 0.95;
    margin: 0;
    text-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  }
}

// 统计信息
.stats-section {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
  justify-content: center;
  width: 100%;
  
  .stat-card {
    flex: 1;
    min-width: 200px;
    max-width: 280px;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border-radius: $border-radius-xl;
    padding: 32px;
    text-align: center;
    box-shadow: $shadow-lg;
    transition: all $transition-normal;
    
    &:hover {
      transform: translateY(-8px);
      box-shadow: $shadow-xl, 0 0 30px rgba(255, 255, 255, 0.3);
    }
    
    .stat-icon {
      font-size: 48px;
      margin-bottom: 16px;
    }
    
    .stat-number {
      font-size: 42px;
      font-weight: 700;
      color: $primary-color;
      margin-bottom: 8px;
    }
    
    .stat-label {
      font-size: 16px;
      color: #666;
    }
  }
}

// 最近作品
.recent-works {
  width: 100%;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: $border-radius-xl;
  padding: 32px;
  box-shadow: $shadow-lg;
  
  h2 {
    font-size: 28px;
    margin: 0 0 24px 0;
    color: #333;
  }
  
  .works-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 20px;
    
    .work-card {
      background: white;
      border-radius: $border-radius-lg;
      padding: 20px;
      cursor: pointer;
      transition: all $transition-normal;
      border: 2px solid transparent;
      
      &:hover {
        transform: translateY(-4px);
        box-shadow: $shadow-md;
        border-color: $primary-light;
      }
      
      .work-cover {
        width: 100%;
        aspect-ratio: 3/2;
        background: $gradient-purple;
        border-radius: $border-radius-md;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-bottom: 16px;
        
        .cover-icon {
          font-size: 48px;
          color: white;
        }
      }
      
      .work-info {
        .work-title {
          font-size: 16px;
          font-weight: 600;
          margin: 0 0 8px 0;
          color: #333;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        
        .work-time {
          font-size: 12px;
          color: #999;
          margin-top: 8px;
        }
      }
    }
  }
}

// 管理按钮
.manage-btn {
  font-size: 18px;
  padding: 20px 60px;
  border-radius: 50px;
  background: $gradient-pink;
  border: none;
  box-shadow: $shadow-lg;
  transition: all $transition-normal;
  animation: pulse 2s ease-in-out infinite;
  
  &:hover {
    transform: scale(1.05);
    box-shadow: $shadow-xl, 0 0 40px rgba(255, 107, 157, 0.5);
    animation: none;
  }
  
  &:active {
    transform: scale(0.98);
  }
}

// 响应式
@media (max-width: 768px) {
  .hero-section {
    .anime-title {
      font-size: 36px;
    }
    
    .subtitle {
      font-size: 16px;
    }
  }
  
  .stats-section {
    .stat-card {
      min-width: 150px;
    }
  }
  
  .recent-works {
    .works-grid {
      grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    }
  }
}
</style>

