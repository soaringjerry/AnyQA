<!-- src/components/QuestionForm.vue -->
<template>
  <div class="container animate__animated animate__fadeIn">
    <h1 class="animate__animated animate__bounceIn">✨ Questions here ✨</h1>
    <div class="question-form animate__animated animate__fadeInUp">
      <v-textarea
        v-model="question"
        label="type here..."
        variant="outlined"
        rows="5"
        auto-grow
      ></v-textarea>
      <v-btn
        :disabled="!configLoaded || submitting"
        color="pink"
        class="ma-2"
        @click="submitQuestion"
      >
        {{ submitting ? '提交中...' : 'Send Question' }}
      </v-btn>
    </div>
    <div id="status" class="status" :class="statusClass">{{ statusMessage }}</div>
    <div class="mascot"></div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted } from 'vue'
import { appConfig } from '../store/configStore'

const configLoaded = inject('configLoaded', false)

const question = ref('')
const statusMessage = ref('')
const statusClass = ref('')
const submitting = ref(false)

function showStatus(message, type) {
  statusMessage.value = message
  statusClass.value = type
  if (type !== 'loading') {
    setTimeout(() => {
      statusMessage.value = ''
      statusClass.value = ''
    }, 3000)
  }
}

async function submitQuestion() {
  if (!question.value.trim()) return
  if (!appConfig.value?.api?.endpoint) {
    showStatus('系统配置错误', 'error')
    return
  }

  submitting.value = true
  showStatus('提交中...', 'loading')

  try {
    const response = await fetch(`${appConfig.value.api.endpoint}/question`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        sessionId: appConfig.value.session.id,
        content: question.value
      })
    })

    if (response.ok) {
      showStatus('发送成功！', 'success')
      question.value = ''
      const mascot = document.querySelector('.mascot')
      if (mascot) {
        mascot.style.transform = 'translateY(-20px)'
        setTimeout(() => {
          mascot.style.transform = 'translateY(0)'
        }, 500)
      }
    } else {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
  } catch (error) {
    console.error('提交错误:', error)
    showStatus('提交失败，请稍后重试', 'error')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  if (!configLoaded) {
    showStatus('系统初始化失败，请刷新页面重试', 'error')
  }
})

document.addEventListener('keydown', (e) => {
  if (e.ctrlKey && e.key === 'Enter') {
    submitQuestion()
  }
})
</script>

<style scoped>
:root {
  --primary-color: #FF69B4;
  --secondary-color: #FFB7C5;
  --accent-color: #FF1493;
  --text-color: #4A4A4A;
  --bg-color: #FFF0F5;
}

.container {
  max-width: 800px;
  width: 95%;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 20px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(255, 105, 180, 0.15);
  backdrop-filter: blur(5px);
}

h1 {
  color: var(--primary-color);
  text-align: center;
  font-size: 2.5em;
  margin-bottom: 30px;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
}

.mascot {
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 150px;
  height: 150px;
  background-image: url('/api/placeholder/150/150');
  background-size: contain;
  background-repeat: no-repeat;
  z-index: 1000;
  transition: transform 0.3s ease;
}

@media (max-width: 768px) {
  .mascot {
    width: 100px;
    height: 100px;
    bottom: 10px;
    right: 10px;
  }
}

.status {
  margin-top: 20px;
  padding: 15px;
  border-radius: 15px;
  text-align: center;
  font-weight: bold;
  opacity: 1;
  transition: all 0.3s ease;
}

.status.success {
  background: rgba(212, 237, 218, 0.9);
  color: #155724;
  animation: fadeInUp 0.5s ease forwards;
}

.status.error {
  background: rgba(248, 215, 218, 0.9);
  color: #721c24;
  animation: fadeInUp 0.5s ease forwards;
}

.status.loading {
  background: rgba(255, 255, 255, 0.9);
  color: #4A4A4A;
  display: flex;
  justify-content: center;
  align-items: center;
  animation: fadeInUp 0.5s ease forwards;
  position: relative;
}

.status.loading::after {
  content: '';
  display: inline-block;
  width: 20px;
  height: 20px;
  margin-left: 10px;
  border: 2px solid var(--primary-color);
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 600px) {
  .container {
    padding: 20px;
    width: 100%;
  }

  h1 {
    font-size: 2em;
  }
}
</style>
