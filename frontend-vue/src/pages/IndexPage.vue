<!-- src/pages/QuestionPage.vue -->
<template>
  <div class="page-root">
    <!-- 樱花背景层 -->
    <div class="sakura-background">
      <div
        v-for="sakura in sakuras"
        :key="sakura.id"
        class="sakura"
        :style="sakura.style"
      >
        🌸
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="page-wrapper">
      <div class="content-container">
        <div class="container animate__animated animate__fadeIn">
          <h1 class="text-center animate__animated animate__bounceIn">
            ✨ Questions here ✨
          </h1>
          
          <div class="question-form">
            <v-textarea
              v-model="question"
              placeholder="type here..."
              variant="outlined"
              :disabled="isSubmitting"
              rows="5"
              class="animate__animated animate__fadeInUp question-textarea"
              hide-details
              @keydown.ctrl.enter="submitQuestion"
            ></v-textarea>

            <v-btn
              :loading="isSubmitting"
              :disabled="isSubmitting"
              color="primary"
              size="large"
              class="submit-btn animate__animated animate__fadeInUp"
              @click="submitQuestion"
            >
              Send Question
            </v-btn>
          </div>

          <!-- Status Message -->
          <transition name="fade">
            <div
              v-if="statusMessage"
              class="status animate__animated animate__fadeInUp"
              :class="[statusType]"
            >
              {{ statusMessage }}
            </div>
          </transition>
        </div>

        <!-- Mascot -->
        <div 
          ref="mascot"
          class="mascot"
        >
          <v-img
            src="/api/placeholder/150/150"
            width="150"
            height="150"
          ></v-img>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useConfig } from '../composables/useConfig'
import 'animate.css'

const question = ref('')
const statusMessage = ref('')
const statusType = ref('')
const isSubmitting = ref(false)
const mascot = ref(null)
const sakuras = ref([])
let sakuraId = 0
let sakuraInterval = null

const { config, configError } = useConfig()

// 状态展示逻辑
function showStatus(message, type) {
  statusMessage.value = message
  statusType.value = type
  
  if (type !== 'loading') {
    setTimeout(() => {
      statusMessage.value = ''
      statusType.value = ''
    }, 3000)
  }
}

// 提交问题
async function submitQuestion() {
  const content = question.value.trim()
  if (!content) return

  if (configError.value) {
    showStatus('System configuration error', 'error')
    return
  }

  isSubmitting.value = true
  showStatus('提交中...', 'loading')

  try {
    const response = await fetch(`${config.value.api.endpoint}/question`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        sessionId: config.value.session.id,
        content
      })
    })

    if (response.ok) {
      showStatus('发送成功！', 'success')
      question.value = ''
      // 吉祥物动画
      if (mascot.value) {
        mascot.value.$el.style.transform = 'translateY(-20px)'
        setTimeout(() => {
          if (mascot.value) {
            mascot.value.$el.style.transform = 'translateY(0)'
          }
        }, 500)
      }
    } else {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
  } catch (error) {
    console.error('提交错误:', error)
    showStatus('提交失败，请稍后重试', 'error')
  } finally {
    isSubmitting.value = false
  }
}

// 樱花效果
function createSakura() {
  const id = sakuraId++
  const startPositionLeft = Math.random() * window.innerWidth
  const endPositionLeft = startPositionLeft + (Math.random() * 200 - 100) // 随机左右飘动
  const duration = Math.random() * 3 + 2  // 2-5秒的随机时长
  
  const style = {
    left: `${startPositionLeft}px`,
    animation: `falling ${duration}s linear forwards`,
    transform: 'translateY(-20px)', // 初始位置在屏幕上方
  }

  sakuras.value.push({ id, style })
  
  // 确保动画结束后移除樱花
  setTimeout(() => {
    const index = sakuras.value.findIndex(s => s.id === id)
    if (index !== -1) {
      sakuras.value.splice(index, 1)
    }
  }, duration * 1000)
}

onMounted(() => {
  sakuraInterval = setInterval(createSakura, 500)
})

onBeforeUnmount(() => {
  if (sakuraInterval) {
    clearInterval(sakuraInterval)
  }
})
</script>

<style scoped>
.page-root {
  min-height: 100vh;
  width: 100%;
  position: relative;
  background: linear-gradient(140deg, #fdfbfb 0%, #ebedee 100%);
  overflow: hidden;
}

.sakura-background {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100vh;
  pointer-events: none;
  z-index: 1;
}

.sakura {
  position: absolute;
  z-index: 1;
  pointer-events: none;
  user-select: none;
  will-change: transform;
}

.page-wrapper {
  position: relative;
  z-index: 2;
  min-height: 100vh;
}

.content-container {
  position: relative;
  padding: 20px;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.container {
  max-width: 800px;
  width: 95%;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 20px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(255, 105, 180, 0.15);
  backdrop-filter: blur(5px);
  margin: 20px auto;
  position: relative;
  z-index: 2;
}

h1 {
  color: #FF69B4;
  font-size: 2.5em;
  margin-bottom: 30px;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
}

.question-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.question-textarea {
  background: rgba(255, 255, 255, 0.9) !important;
}

.question-textarea :deep(.v-field__outline) {
  border-color: #FFB7C5 !important;
}

.question-textarea :deep(.v-field--focused .v-field__outline) {
  border-color: #FF69B4 !important;
}

.submit-btn {
  align-self: stretch;
  height: 50px !important;
  border-radius: 25px !important;
  background: #FF69B4 !important;
  transition: all 0.3s ease !important;
}

.submit-btn:hover {
  background: #FF1493 !important;
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(255, 105, 180, 0.3) !important;
}

.status {
  margin-top: 20px;
  padding: 15px;
  border-radius: 15px;
  text-align: center;
  font-weight: bold;
}

.success {
  background: rgba(212, 237, 218, 0.9);
  color: #155724;
}

.error {
  background: rgba(248, 215, 218, 0.9);
  color: #721c24;
}

.loading {
  background: rgba(255, 255, 255, 0.9);
  color: #4A4A4A;
  display: flex;
  justify-content: center;
  align-items: center;
}

.loading::after {
  content: '';
  display: inline-block;
  width: 20px;
  height: 20px;
  margin-left: 10px;
  border: 2px solid #FF69B4;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.mascot {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 1000;
  transition: transform 0.3s ease;
}

@keyframes falling {
  0% {
    transform: translateY(-20px) rotate(0deg);
  }
  25% {
    transform: translateY(25vh) rotate(90deg);
  }
  50% {
    transform: translateY(50vh) rotate(180deg);
  }
  75% {
    transform: translateY(75vh) rotate(270deg);
  }
  100% {
    transform: translateY(105vh) rotate(360deg);
  }
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .container {
    padding: 20px;
    width: 100%;
  }

  h1 {
    font-size: 2em;
  }

  .mascot {
    width: 100px !important;
    height: 100px !important;
    bottom: 10px;
    right: 10px;
  }
}

/* 过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>