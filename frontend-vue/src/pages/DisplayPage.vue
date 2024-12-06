<template>
  <div>
    <div class="cyber-line"></div>
    <div class="container">
      <header class="header">
        <h1>Question Display</h1>
      </header>
      
      <div 
        class="current-question" 
        :class="{ loading: !currentQuestion }"
      >
        {{ currentQuestion ? currentQuestion.content : 'Waiting...' }}
      </div>

      <div class="question-list">
        <div 
          v-for="(q, index) in pendingQuestions" 
          :key="index"
          class="question-card animate__animated animate__fadeInUp"
        >
          {{ q.content }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import config from '../config/index.js' // 假设已在此文件加载YAML配置
// config.api.endpoint, config.ws.endpoint, config.session.id

const questions = ref([])

// 计算当前展示问题
const currentQuestion = computed(() => {
  return questions.value.find(q => q.status === 'showing')
})

// 计算待定问题列表
const pendingQuestions = computed(() => {
  return questions.value.filter(q => q.status === 'pending')
})

let ws = null
let intervalId = null

async function loadQuestions() {
  try {
    const response = await fetch(`${config.api.endpoint}/questions/${config.session.id}`)
    if (!response.ok) {
      console.error('加载问题失败:', response.status)
      return
    }
    const data = await response.json()
    questions.value = data
  } catch (error) {
    console.error('加载失败:', error)
  }
}

function initializeApp() {
  // 建立 WebSocket 连接
  ws = new WebSocket(config.ws.endpoint)
  ws.onmessage = () => {
    // 收到消息后刷新列表
    loadQuestions()
  }

  // 首次加载
  loadQuestions()

  // 定期刷新
  intervalId = setInterval(loadQuestions, 5000)
}

onMounted(() => {
  initializeApp()
})

onUnmounted(() => {
  // 清理资源
  if (ws) {
    ws.close()
    ws = null
  }
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
  }
})
</script>

<style scoped>
:root {
    --primary-bg: #1a1c2e;
    --secondary-bg: #2a2d4a;
    --accent-color: #6e7dff;
    --text-primary: #ffffff;
    --text-secondary: #b3b9ff;
    --card-bg: rgba(42, 45, 74, 0.8);
}

* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    background: var(--primary-bg);
    color: var(--text-primary);
    font-family: 'Arial', sans-serif;
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    background-image: 
        radial-gradient(circle at 10% 20%, rgba(110, 125, 255, 0.1) 0%, transparent 50%),
        radial-gradient(circle at 90% 80%, rgba(110, 125, 255, 0.1) 0%, transparent 50%);
    overflow: hidden;
}

.container {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 2rem;
    gap: 2rem;
}

.header {
    text-align: center;
    padding: 1rem;
    background: var(--secondary-bg);
    border-radius: 1rem;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
    animation: slideInDown 0.5s ease-out;
}

.header h1 {
    font-size: 1.5rem;
    color: var(--accent-color);
    text-transform: uppercase;
    letter-spacing: 2px;
}

.current-question {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: clamp(2rem, 5vw, 4rem);
    text-align: center;
    padding: 2rem;
    background: var(--card-bg);
    border-radius: 1.5rem;
    border: 1px solid rgba(110, 125, 255, 0.2);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
    margin: 1rem 0;
    animation: fadeIn 0.5s ease-out;
    transition: all 0.3s ease;
}

.current-question.loading {
    position: relative;
}

.current-question.loading::after {
    content: '...';
    animation: dots 1.5s steps(4, end) infinite;
    display: inline-block;
    width: 0;
    overflow: hidden;
    vertical-align: bottom;
}

.question-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 1rem;
    padding: 1rem;
    background: rgba(26, 28, 46, 0.9);
    border-radius: 1rem;
    max-height: 30vh;
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color: var(--accent-color) var(--primary-bg);
}

.question-list::-webkit-scrollbar {
    width: 8px;
}

.question-list::-webkit-scrollbar-track {
    background: var(--primary-bg);
}

.question-list::-webkit-scrollbar-thumb {
    background: var(--accent-color);
    border-radius: 4px;
}

.question-card {
    background: var(--card-bg);
    padding: 1rem;
    border-radius: 0.8rem;
    border: 1px solid rgba(110, 125, 255, 0.1);
    transition: transform 0.3s ease, box-shadow 0.3s ease;
    font-size: 1.1rem;
    line-height: 1.5;
}

.question-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 20px rgba(110, 125, 255, 0.2);
}

@media (max-width: 768px) {
    .container {
        padding: 1rem;
    }

    .current-question {
        font-size: clamp(1.5rem, 4vw, 2.5rem);
        padding: 1rem;
    }

    .question-list {
        grid-template-columns: 1fr;
        max-height: 40vh;
    }
}

@keyframes dots {
    to {
        width: 1.25em;
    }
}

.cyber-line {
    position: fixed;
    height: 2px;
    background: linear-gradient(90deg, transparent, var(--accent-color), transparent);
    width: 100%;
    animation: scan 8s linear infinite;
    opacity: 0.5;
}

@keyframes scan {
    0% {
        top: -10%;
    }
    100% {
        top: 110%;
    }
}

/* 您可以根据需要定义fadeIn、slideInDown的关键帧动画 */
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes fadeInUp {
  from { 
    opacity: 0;
    transform: translateY(20px);
  }
  to { 
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slideInDown {
  from {
    transform: translateY(-100%);
    opacity: 0;
  }
  to {
    transform: translateY(0%);
    opacity: 1;
  }
}
</style>
