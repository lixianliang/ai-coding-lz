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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Loading, VideoPlay, VideoPause } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useDocumentStore } from '@/stores/document'

const route = useRoute()
const router = useRouter()
const store = useDocumentStore()

const loading = ref(false)
const playingSceneId = ref<string | null>(null)
const audioRefs = ref<Map<string, HTMLAudioElement>>(new Map())

// 播放/暂停音频
const toggleAudio = async (sceneId: string, voiceUrl: string) => {
  if (playingSceneId.value === sceneId) {
    // 暂停当前播放的音频
    const audio = audioRefs.value.get(sceneId)
    if (audio) {
      audio.pause()
      playingSceneId.value = null
    }
    return
  }

  try {
    // 如果正在播放其他音频，先停止
    if (playingSceneId.value) {
      const prevAudio = audioRefs.value.get(playingSceneId.value)
      if (prevAudio) {
        prevAudio.pause()
        prevAudio.currentTime = 0
      }
    }

    // 创建新的音频对象
    const audio = new Audio(voiceUrl)
    
    // 监听播放结束
    audio.addEventListener('ended', () => {
      playingSceneId.value = null
      audioRefs.value.delete(sceneId)
    })

    // 监听错误
    audio.addEventListener('error', (e) => {
      console.error('音频播放失败:', e)
      ElMessage.error('音频播放失败')
      playingSceneId.value = null
      audioRefs.value.delete(sceneId)
    })

    // 保存音频引用并播放
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

// 组件卸载时清理音频资源
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

onUnmounted(() => {
  // 停止所有音频并清理
  audioRefs.value.forEach((audio) => {
    audio.pause()
    audio.src = ''
  })
  audioRefs.value.clear()
  playingSceneId.value = null
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

    .scene-audio {
      margin-bottom: 16px;
      
      .el-button {
        display: inline-flex;
        align-items: center;
        gap: 6px;
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
