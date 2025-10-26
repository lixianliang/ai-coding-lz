<template>
  <div class="scene-viewer-page">
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
          <el-button type="text" class="back-btn" @click="router.back()">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
          <h2 class="page-title">{{ store.currentDocument?.name }} - 场景</h2>
        </div>
      </el-header>
      
      <el-main>
        <el-empty v-if="!loading && store.scenes.length === 0" description="暂无场景" />
        
        <div v-else class="scene-list" v-loading="loading">
          <div 
            v-for="(scene, idx) in store.scenes" 
            :key="scene.id" 
            class="scene-item slide-up"
            :style="{ animationDelay: `${idx * 0.1}s` }"
          >
            <div class="scene-header">
              <div class="scene-badge">
                <span class="scene-icon">🎬</span>
                <span class="scene-index">场景 {{ scene.index }}</span>
              </div>
              <div class="scene-actions">
                <el-tag size="small" v-if="scene.image_url" effect="dark">已完成</el-tag>
                <el-tag size="small" type="info" effect="dark" v-else>处理中</el-tag>
                <el-button type="primary" size="small" :icon="Edit" @click="handleEditScene(scene)">
                  编辑
                </el-button>
              </div>
            </div>
            
            <div class="scene-content">
              <div class="scene-text">
                <p>{{ scene.content }}</p>
              </div>
              
              <!-- 音频播放按钮 -->
              <div class="scene-audio" v-if="scene.voice_url">
                <el-button 
                  :type="playingSceneId === scene.id ? 'danger' : 'primary'"
                  :icon="playingSceneId === scene.id ? VideoPause : VideoPlay"
                  size="small"
                  @click="toggleAudio(scene.id, scene.voice_url)"
                >
                  {{ playingSceneId === scene.id ? '暂停' : '播放' }}
                </el-button>
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

    <!-- 编辑场景对话框 -->
    <el-dialog v-model="showEditSceneDialog" title="编辑场景" width="700px">
      <el-form :model="sceneForm" label-width="80px" :rules="sceneRules" ref="sceneFormRef">
        <el-form-item label="场景内容" prop="content">
          <el-input 
            v-model="sceneForm.content" 
            type="textarea" 
            :rows="8" 
            placeholder="请输入场景描述"
          />
        </el-form-item>
        <el-alert
          title="提示：修改场景内容后，将立即重新生成图片和语音（可能需要几秒钟）"
          type="info"
          :closable="false"
          show-icon
        />
      </el-form>
      <template #footer>
        <el-button @click="showEditSceneDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveScene" :loading="submitting">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Loading, VideoPlay, VideoPause, Edit } from '@element-plus/icons-vue'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { useDocumentStore } from '@/stores/document'
import { Scene, UpdateSceneRequest } from '@/apis/types'

const route = useRoute()
const router = useRouter()
const store = useDocumentStore()

const loading = ref(false)
const playingSceneId = ref<string | null>(null)
const audioRefs = ref<Map<string, HTMLAudioElement>>(new Map())
const showEditSceneDialog = ref(false)
const submitting = ref(false)
const sceneFormRef = ref<FormInstance>()
const currentEditingSceneId = ref('')

const sceneForm = ref<UpdateSceneRequest>({
  content: ''
})

const sceneRules: FormRules = {
  content: [{ required: true, message: '请输入场景内容', trigger: 'blur' }]
}

// 轮询定时器
let pollInterval: NodeJS.Timeout | null = null

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

// 播放/暂停音频
const toggleAudio = async (sceneId: string, voiceUrl: string) => {
  if (playingSceneId.value === sceneId) {
    const audio = audioRefs.value.get(sceneId)
    if (audio) {
      audio.pause()
      playingSceneId.value = null
    }
    return
  }

  try {
    if (playingSceneId.value) {
      const prevAudio = audioRefs.value.get(playingSceneId.value)
      if (prevAudio) {
        prevAudio.pause()
        prevAudio.currentTime = 0
      }
    }

    const audio = new Audio(voiceUrl)
    
    audio.addEventListener('ended', () => {
      playingSceneId.value = null
      audioRefs.value.delete(sceneId)
    })

    audio.addEventListener('error', (e) => {
      console.error('音频播放失败:', e)
      ElMessage.error('音频播放失败')
      playingSceneId.value = null
      audioRefs.value.delete(sceneId)
    })

    audioRefs.value.set(sceneId, audio)
    playingSceneId.value = sceneId
    await audio.play()
  } catch (error) {
    console.error('播放音频失败:', error)
    ElMessage.error('播放音频失败')
    playingSceneId.value = null
    audioRefs.value.delete(sceneId)
  }
}

// 开始轮询场景状态
const startPolling = () => {
  const id = route.params.id as string
  
  if (pollInterval) {
    clearInterval(pollInterval)
  }
  
  pollInterval = setInterval(async () => {
    await Promise.all([
      store.fetchDocument(id),
      store.fetchDocumentScenes(id)
    ])
    
    const allImagesReady = store.scenes.every(scene => scene.image_url)
    if (allImagesReady && pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
  }, 5000)
}

// 停止轮询
const stopPolling = () => {
  if (pollInterval) {
    clearInterval(pollInterval)
    pollInterval = null
  }
}

// 编辑场景
const handleEditScene = (scene: Scene) => {
  currentEditingSceneId.value = scene.id
  sceneForm.value = {
    content: scene.content
  }
  showEditSceneDialog.value = true
}

// 保存场景
const handleSaveScene = async () => {
  if (!sceneFormRef.value) return
  
  await sceneFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      await store.updateScene(currentEditingSceneId.value, sceneForm.value)
      ElMessage.success('场景更新成功，图片和语音已重新生成')
      showEditSceneDialog.value = false
      
      // 刷新场景列表以获取最新的图片和语音 URL
      const docId = store.currentDocument?.id
      if (docId) {
        await store.fetchDocumentScenes(docId)
      }
    } catch (error) {
      console.error('更新场景失败:', error)
      ElMessage.error('场景更新失败')
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
      store.fetchDocumentScenes(id)
    ])
    
    const hasUnfinished = store.scenes.some(scene => !scene.image_url)
    if (hasUnfinished) {
      startPolling()
    }
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  stopPolling()
  
  audioRefs.value.forEach((audio) => {
    audio.pause()
    audio.src = ''
  })
  audioRefs.value.clear()
  playingSceneId.value = null
})
</script>

<style scoped lang="scss">
@use '../styles/variables.scss' as *;

.scene-viewer-page {
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
  
  :deep(.el-header) {
    position: relative;
    z-index: 1;
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(20px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.2);
    
    .header-content {
      display: flex;
      align-items: center;
      gap: 16px;
      height: 100%;
      padding: 0 24px;
      
      .back-btn {
        color: white;
        font-size: 14px;
        
        &:hover {
          background: rgba(255, 255, 255, 0.1);
        }
      }
      
      .page-title {
        margin: 0;
        font-size: 24px;
        font-weight: 700;
        color: white;
        text-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
      }
    }
  }
  
  :deep(.el-main) {
    position: relative;
    z-index: 1;
    padding: 40px 24px;
    
    .scene-list {
      max-width: 1200px;
      margin: 0 auto;
    }
    
    .scene-item {
      margin-bottom: 32px;
      padding: 32px;
      background: rgba(255, 255, 255, 0.95);
      backdrop-filter: blur(10px);
      border-radius: $border-radius-lg;
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
      transition: all $transition-normal;
      
      &:hover {
        transform: translateY(-4px);
        box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
      }
    }
    
    .scene-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 24px;
      padding-bottom: 16px;
      border-bottom: 2px solid rgba(102, 126, 234, 0.2);
      
      .scene-badge {
        display: flex;
        align-items: center;
        gap: 12px;
        
        .scene-icon {
          font-size: 24px;
          animation: pulse 2s ease-in-out infinite;
        }
        
        .scene-index {
          font-weight: 700;
          font-size: 20px;
          color: #333;
        }
      }
      
      .scene-actions {
        display: flex;
        align-items: center;
        gap: 12px;
      }
    }
  
    .scene-content {
      .scene-text {
        margin-bottom: 20px;
        padding: 16px;
        background: rgba(102, 126, 234, 0.05);
        border-left: 4px solid #667eea;
        border-radius: 8px;
        
        p {
          margin: 0;
          line-height: 1.8;
          color: #555;
          font-size: 15px;
        }
      }

      .scene-audio {
        margin-bottom: 20px;
        
        .el-button {
          display: inline-flex;
          align-items: center;
          gap: 6px;
          padding: 10px 20px;
          font-weight: 600;
          border-radius: 25px;
          transition: all $transition-fast;
          
          &:hover {
            transform: scale(1.05);
          }
        }
      }
      
      .scene-image {
        text-align: center;
        margin-top: 24px;
        
        img {
          width: 100%;
          max-width: 900px;
          border-radius: 12px;
          box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
          transition: all $transition-normal;
          
          &:hover {
            transform: scale(1.02);
            box-shadow: 0 12px 48px rgba(0, 0, 0, 0.3);
          }
        }
      }
      
      .scene-loading {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 12px;
        padding: 60px;
        text-align: center;
        background: rgba(255, 255, 255, 0.6);
        border-radius: 12px;
        
        .el-icon {
          font-size: 32px;
          color: #667eea;
        }
        
        span {
          font-size: 16px;
          color: #667eea;
          font-weight: 600;
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
  .scene-viewer-page {
    :deep(.el-header) {
      .header-content {
        padding: 0 16px;
        
        .page-title {
          font-size: 18px;
        }
      }
    }
    
    :deep(.el-main) {
      padding: 20px 16px;
      
      .scene-item {
        padding: 20px;
        
        .scene-header {
          .scene-badge {
            .scene-icon {
              font-size: 20px;
            }
            
            .scene-index {
              font-size: 16px;
            }
          }
        }
      }
    }
  }
}
</style>
