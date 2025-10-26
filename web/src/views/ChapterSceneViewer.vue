<template>
  <div class="chapter-scene-viewer-page">
    <el-container>
      <el-header height="60px">
        <div class="header-content">
          <el-button @click="router.back()">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
          <h2>{{ store.currentChapter?.title || `章节 ${store.currentChapter?.index}` }} - 场景</h2>
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
                <el-image 
                  :src="scene.image_url" 
                  :preview-src-list="[scene.image_url]"
                  fit="cover"
                  class="image-preview"
                >
                  <template #error>
                    <div class="image-error">
                      <el-icon><Picture /></el-icon>
                      <span>加载失败</span>
                    </div>
                  </template>
                </el-image>
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
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Loading, VideoPlay, VideoPause, Picture } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useDocumentStore } from '@/stores/document'

const route = useRoute()
const router = useRouter()
const store = useDocumentStore()

const loading = ref(false)
const playingSceneId = ref<string | null>(null)
const audioRefs = ref<Map<string, HTMLAudioElement>>(new Map())

// 轮询定时器
let pollInterval: NodeJS.Timeout | null = null

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
  const chapterId = route.params.chapterId as string
  
  if (pollInterval) {
    clearInterval(pollInterval)
  }
  
  pollInterval = setInterval(async () => {
    await store.fetchChapterScenes(chapterId)
    
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

onMounted(async () => {
  const chapterId = route.params.chapterId as string
  loading.value = true
  try {
    await Promise.all([
      store.fetchChapter(chapterId),
      store.fetchChapterScenes(chapterId)
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
.chapter-scene-viewer-page {
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
        font-size: 15px;
      }
    }
    
    .scene-audio {
      margin-bottom: 16px;
      
      .el-button {
        display: inline-flex;
        align-items: center;
        gap: 6px;
      }
    }
    
    .scene-image {
      margin-top: 16px;
      
      .image-preview {
        width: 100%;
        border-radius: 8px;
        cursor: pointer;
        transition: transform 0.3s;
        
        &:hover {
          transform: scale(1.02);
        }
      }
      
      .image-error {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 60px;
        color: #999;
        
        .el-icon {
          font-size: 48px;
          margin-bottom: 8px;
        }
      }
    }
    
    .scene-loading {
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 60px;
      color: #999;
      
      .el-icon {
        font-size: 32px;
        margin-right: 12px;
      }
    }
  }
}
</style>

